package learning

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSavedCollectionFiltersPostgreSQLFullSetSnapshotAndIsolation(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(t.Context(), `CREATE TEMP TABLE canonical_words(id uuid PRIMARY KEY,text text NOT NULL,normalized_text text NOT NULL,status text NOT NULL);
 CREATE TEMP TABLE word_meanings(id uuid PRIMARY KEY,word_id uuid NOT NULL REFERENCES canonical_words(id),part_of_speech text NOT NULL,short_definition text NOT NULL,status text NOT NULL);
 CREATE TEMP TABLE user_words(id uuid PRIMARY KEY,user_id uuid NOT NULL,meaning_id uuid NOT NULL REFERENCES word_meanings(id),status text NOT NULL,source text NOT NULL,added_at timestamptz NOT NULL,total_review_count integer NOT NULL DEFAULT 0,next_review_at timestamptz,deleted_at timestamptz);
 CREATE UNIQUE INDEX ON user_words(user_id,meaning_id) WHERE deleted_at IS NULL;`)
	require.NoError(t, err)
	data, owner, other := savedFilterFixture()
	for _, word := range data.Words {
		_, err = db.ExecContext(t.Context(), `INSERT INTO canonical_words VALUES($1,$2,$3,$4)`, word.ID, word.Text, word.NormalizedText, word.Status)
		require.NoError(t, err)
	}
	for _, meaning := range data.Meanings {
		_, err = db.ExecContext(t.Context(), `INSERT INTO word_meanings VALUES($1,$2,$3,$4,$5)`, meaning.ID, meaning.WordID, meaning.PartOfSpeech, meaning.ShortDefinition, meaning.Status)
		require.NoError(t, err)
	}
	for _, word := range data.UserWords {
		_, err = db.ExecContext(t.Context(), `INSERT INTO user_words VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, word.ID, word.UserID, word.MeaningID, word.Status, word.Source, word.AddedAt, word.TotalReviewCount, word.NextReviewAt, word.DeletedAt)
		require.NoError(t, err)
	}
	memory, repo := NewMemoryRepository(data), NewPostgreSQLRepository(db)
	for _, request := range []ListSavedWordsRequest{
		{UserID: owner, Limit: 50}, {UserID: owner, Query: "refund", Limit: 50},
		{UserID: owner, Stage: "new"}, {UserID: owner, Stage: "learning"}, {UserID: owner, Stage: "reviewing", DueOnly: true},
		{UserID: owner, Stage: "mastered", DueOnly: true}, {UserID: owner, Stage: "ignored"}, {UserID: owner, Stage: "archived"},
		{UserID: owner, DueOnly: true, Limit: 1}, {UserID: owner, Query: " RETURNED ", Limit: 1},
		{UserID: owner, Query: "%"}, {UserID: owner, Query: "_"}, {UserID: owner, Query: "_100%"},
		{UserID: owner, Query: "missing"}, {UserID: other, Query: "refund"}, {UserID: uuid.New()},
	} {
		expected, err := memory.ListSavedWords(t.Context(), request)
		require.NoError(t, err)
		actual, err := repo.ListSavedWords(t.Context(), request)
		require.NoError(t, err)
		require.Equal(t, expected, actual, "request=%+v", request)
	}
	request := ListSavedWordsRequest{UserID: owner, Query: "refund", Limit: 50}
	first, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, first.Items, 50)
	require.Equal(t, 62, first.TotalCount)
	request.AfterCursor = first.NextCursor
	second, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, second.Items, 12)
	require.Equal(t, 62, second.TotalCount)
	seen := map[uuid.UUID]bool{}
	for _, item := range append(first.Items, second.Items...) {
		require.False(t, seen[item.UserWordID])
		seen[item.UserWordID] = true
	}
	request.AfterCursor = nextSavedCursor(ListSavedWordsRequest{UserID: owner, Query: "refund"}, second.Items[len(second.Items)-1])
	exhausted, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	require.Empty(t, exhausted.Items)
	require.Equal(t, 62, exhausted.TotalCount)
	require.Empty(t, exhausted.NextCursor)

	// Removing the first boundary does not restart or skip the next saved record.
	boundaryRequest := ListSavedWordsRequest{UserID: owner, Query: "refund", Limit: 1}
	boundary, err := repo.ListSavedWords(t.Context(), boundaryRequest)
	require.NoError(t, err)
	boundaryRequest.AfterCursor = boundary.NextCursor
	before, err := repo.ListSavedWords(t.Context(), boundaryRequest)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), `UPDATE user_words SET deleted_at=$2 WHERE id=$1`, boundary.Items[0].UserWordID, time.Now())
	require.NoError(t, err)
	after, err := repo.ListSavedWords(t.Context(), boundaryRequest)
	require.NoError(t, err)
	require.Equal(t, before.Items[0].UserWordID, after.Items[0].UserWordID)
	require.Equal(t, 61, after.TotalCount)
	for _, changed := range []ListSavedWordsRequest{
		{UserID: other, Query: "refund"}, {UserID: owner, Query: "returned"}, {UserID: owner, Query: "refund", Stage: "new"}, {UserID: owner, Query: "refund", DueOnly: true},
	} {
		changed.AfterCursor = first.NextCursor
		_, err = repo.ListSavedWords(t.Context(), changed)
		require.ErrorIs(t, err, ErrInvalidCursor)
	}
}

func TestSavedCollectionFiltersPostgreSQLMigratedCanonicalSeed(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	owner, other := uuid.New(), uuid.New()
	// These are fresh isolated fixture identities, never an existing learner.
	defer func() {
		_, cleanupErr := db.ExecContext(context.Background(), "DELETE FROM user_words WHERE user_id IN ($1,$2)", owner, other)
		require.NoError(t, cleanupErr)
		_, cleanupErr = db.ExecContext(context.Background(), "DELETE FROM users WHERE id IN ($1,$2)", owner, other)
		require.NoError(t, cleanupErr)
	}()
	_, err = db.ExecContext(t.Context(), "INSERT INTO users(id,status,onboarding_status,created_at,updated_at) VALUES($1,'active','completed',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),($2,'active','completed',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)", owner, other)
	require.NoError(t, err)
	type target struct {
		id               uuid.UUID
		text, definition string
	}
	rows, err := db.QueryContext(t.Context(), "SELECT m.id,cw.text,m.short_definition FROM word_meanings m JOIN canonical_words cw ON cw.id=m.word_id WHERE m.status='active' AND cw.status='active' ORDER BY m.id")
	require.NoError(t, err)
	var targets []target
	for rows.Next() {
		var item target
		require.NoError(t, rows.Scan(&item.id, &item.text, &item.definition))
		targets = append(targets, item)
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	require.Greater(t, len(targets), 50, "this integration requires the migrated, seeded starter catalog")
	for i, item := range targets {
		_, err = db.ExecContext(t.Context(), "INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at) VALUES($1,$2,$3,'new','manual',CURRENT_TIMESTAMP-($4::int*interval '1 minute'),CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)", uuid.New(), owner, item.id, i)
		require.NoError(t, err)
	}
	_, err = db.ExecContext(t.Context(), "INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at) VALUES($1,$2,$3,'new','manual',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)", uuid.New(), other, targets[0].id)
	require.NoError(t, err)
	repo := NewPostgreSQLRepository(db)
	request := ListSavedWordsRequest{UserID: owner, Stage: "new", DueOnly: true, Limit: 50}
	first, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, first.Items, 50)
	require.Equal(t, len(targets), first.TotalCount)
	request.AfterCursor = first.NextCursor
	second, err := repo.ListSavedWords(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, second.Items, len(targets)-50)
	require.Equal(t, len(targets), second.TotalCount)
	query := strings.ToLower(targets[0].text)
	expected := 0
	for _, item := range targets {
		if strings.Contains(strings.ToLower(item.text), query) || strings.Contains(strings.ToLower(item.definition), query) {
			expected++
		}
	}
	filtered, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: owner, Query: query, Limit: 50})
	require.NoError(t, err)
	require.Equal(t, expected, filtered.TotalCount)
	foreign, err := repo.ListSavedWords(t.Context(), ListSavedWordsRequest{UserID: other})
	require.NoError(t, err)
	require.Equal(t, 1, foreign.TotalCount)
	for _, item := range append(first.Items, second.Items...) {
		require.Equal(t, "new", item.ReviewState)
		require.True(t, item.Due)
	}
}
