package learning

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeSummaryPostgreSQLFullSetIsolationAndReviewStates(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(t.Context(), `CREATE TEMP TABLE canonical_words(id uuid PRIMARY KEY,status text);
	CREATE TEMP TABLE word_meanings(id uuid PRIMARY KEY,word_id uuid REFERENCES canonical_words(id),status text);
	CREATE TEMP TABLE user_words(id uuid PRIMARY KEY,user_id uuid NOT NULL,meaning_id uuid REFERENCES word_meanings(id),status text NOT NULL,total_review_count integer NOT NULL DEFAULT 0,next_review_at timestamptz,deleted_at timestamptz);
	CREATE UNIQUE INDEX ON user_words(user_id,meaning_id) WHERE deleted_at IS NULL;`)
	require.NoError(t, err)
	owner, other, wordID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	_, err = db.ExecContext(t.Context(), `INSERT INTO canonical_words VALUES($1,'archived')`, wordID)
	require.NoError(t, err)
	// 55 unique saved meanings prove the aggregate is independent of the 50-row
	// list cap. Archived canonical content remains in existing saved/due reads.
	for i := 0; i < 55; i++ {
		meaningID := uuid.New()
		_, err = db.ExecContext(t.Context(), `INSERT INTO word_meanings VALUES($1,$2,'archived')`, meaningID, wordID)
		require.NoError(t, err)
		_, err = db.ExecContext(t.Context(), `INSERT INTO user_words(id,user_id,meaning_id,status) VALUES($1,$2,$3,'new')`, uuid.New(), owner, meaningID)
		require.NoError(t, err)
	}
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	for _, row := range []struct {
		user         uuid.UUID
		status       string
		reviews      int
		due, deleted *time.Time
	}{
		{owner, "new", 1, &future, nil}, {owner, "learning", 1, &now, nil}, {owner, "reviewing", 2, &past, nil}, {owner, "reviewing", 2, &future, nil}, {owner, "mastered", 5, &past, nil}, {owner, "ignored", 0, nil, nil}, {owner, "archived", 0, nil, nil}, {owner, "learning", 1, &past, &now}, {other, "new", 0, nil, nil},
	} {
		meaningID := uuid.New()
		_, err = db.ExecContext(t.Context(), `INSERT INTO word_meanings VALUES($1,$2,'archived')`, meaningID, wordID)
		require.NoError(t, err)
		_, err = db.ExecContext(t.Context(), `INSERT INTO user_words(id,user_id,meaning_id,status,total_review_count,next_review_at,deleted_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, uuid.New(), row.user, meaningID, row.status, row.reviews, row.due, row.deleted)
		require.NoError(t, err)
	}
	repo := NewPostgreSQLRepository(db)
	summary, err := repo.GetKnowledgeSummary(t.Context(), owner)
	require.NoError(t, err)
	require.Equal(t, &KnowledgeSummary{Saved: 62, New: 55, Learning: 2, Reviewing: 2, Mastered: 1, Ignored: 1, Archived: 1, Due: 57}, summary)
	foreign, err := repo.GetKnowledgeSummary(t.Context(), other)
	require.NoError(t, err)
	require.Equal(t, &KnowledgeSummary{Saved: 1, New: 1, Due: 1}, foreign)
	empty, err := repo.GetKnowledgeSummary(t.Context(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, &KnowledgeSummary{}, empty)
	_, err = db.ExecContext(t.Context(), `UPDATE user_words SET deleted_at=$2 WHERE user_id=$1 AND deleted_at IS NULL`, owner, now)
	require.NoError(t, err)
	removed, err := repo.GetKnowledgeSummary(t.Context(), owner)
	require.NoError(t, err)
	require.Equal(t, &KnowledgeSummary{}, removed)
}
