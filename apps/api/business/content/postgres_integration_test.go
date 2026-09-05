package content

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestGetMeaningsBySituationPostgreSQLUsesDiscoveryPriority exercises the
// documented core/display-order/relevance ordering against PostgreSQL itself.
// It uses connection-local temporary tables and never touches persistent data.
func TestGetMeaningsBySituationPostgreSQLUsesDiscoveryPriority(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		CREATE TEMP TABLE canonical_words (id uuid PRIMARY KEY, text text NOT NULL, normalized_text text NOT NULL, status text NOT NULL);
		CREATE TEMP TABLE word_meanings (id uuid PRIMARY KEY, word_id uuid NOT NULL REFERENCES canonical_words(id), part_of_speech text NOT NULL, short_definition text NOT NULL, status text NOT NULL);
		CREATE TEMP TABLE journey_words (journey_situation_id uuid NOT NULL, meaning_id uuid NOT NULL REFERENCES word_meanings(id), display_order integer, relevance_score integer NOT NULL, is_core boolean NOT NULL);`)
	require.NoError(t, err)

	situationID, wordID := uuid.New(), uuid.New()
	firstID, coreID, relevantID := uuid.New(), uuid.New(), uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words (id, text, normalized_text, status) VALUES ($1, 'word', 'word', 'active')`, wordID)
	require.NoError(t, err)
	for _, meaningID := range []uuid.UUID{firstID, coreID, relevantID} {
		_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, status) VALUES ($1, $2, 'noun', 'definition', 'active')`, meaningID, wordID)
		require.NoError(t, err)
	}
	for _, row := range []struct {
		meaningID uuid.UUID
		relevance int
		core      bool
	}{
		{firstID, 20, false},
		{coreID, 1, true},
		{relevantID, 90, false},
	} {
		_, err = db.ExecContext(ctx, `INSERT INTO journey_words (journey_situation_id, meaning_id, display_order, relevance_score, is_core) VALUES ($1, $2, 1, $3, $4)`, situationID, row.meaningID, row.relevance, row.core)
		require.NoError(t, err)
	}

	meanings, err := NewPostgreSQLRepository(db).GetMeaningsBySituation(ctx, situationID)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{coreID, relevantID, firstID}, []uuid.UUID{meanings[0].MeaningID, meanings[1].MeaningID, meanings[2].MeaningID})
}
