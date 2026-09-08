package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const curatedContentOrderingMigration = "20260908250000_voc1431_curated_content_ordering.sql"

func TestCuratedContentOrderingMigrationCarriesAllInvariants(t *testing.T) {
	body, err := os.ReadFile(curatedContentOrderingMigration)
	require.NoError(t, err)
	text := string(body)
	for _, invariant := range []string{
		"word_meanings_meaning_order_positive",
		"CHECK (meaning_order > 0) NOT VALID",
		"word_examples_example_order_positive",
		"CHECK (example_order > 0) NOT VALID",
		"usage_notes_note_order_positive",
		"CHECK (note_order > 0) NOT VALID",
		"journey_situations_display_order_positive",
		"CHECK (display_order > 0) NOT VALID",
		"journey_words_display_order_positive",
		"CHECK (display_order IS NULL OR display_order > 0) NOT VALID",
	} {
		assert.Contains(t, text, invariant)
	}

	schemaChecks := map[string][]string{
		"../ent/schema/wordmeaning.go":      {`"meaning_order_positive": "meaning_order > 0"`},
		"../ent/schema/wordexample.go":      {`"example_order_positive": "example_order > 0"`},
		"../ent/schema/usagenote.go":        {`"note_order_positive": "note_order > 0"`},
		"../ent/schema/journeysituation.go": {`"display_order_positive": "display_order > 0"`},
		"../ent/schema/journeyword.go": {
			`"display_order_positive": "display_order IS NULL OR display_order > 0"`,
			`field.Int("display_order").Positive().Optional().Nillable()`,
		},
	}
	for filename, checks := range schemaChecks {
		body, readErr := os.ReadFile(filename)
		require.NoError(t, readErr)
		for _, check := range checks {
			assert.Contains(t, string(body), check)
		}
	}
}

func TestCuratedContentOrderingAgainstPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL migration test unavailable")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	schema := "vocanova_curated_ordering_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	for _, filename := range []string{
		"20260724210000_identity_foundation.sql",
		"20260725100000_voc026_p1_content_tables.sql",
	} {
		migration, readErr := os.ReadFile(filename)
		require.NoError(t, readErr)
		_, err = db.ExecContext(ctx, string(migration))
		require.NoError(t, err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	wordID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words
		(id, text, normalized_text, word_type, language_code, status, created_at, updated_at)
		VALUES ($1, 'order', 'order', 'word', 'en', 'draft', $2, $2)`, wordID, now)
	require.NoError(t, err)

	legacyMeaningID := uuid.New()
	legacyExampleID := uuid.New()
	legacyNoteID := uuid.New()
	legacySituationID := uuid.New()
	legacyJourneyWordID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings
		(id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at)
		VALUES ($1, $2, 'noun', 'position', -1, 'draft', $3, $3)`, legacyMeaningID, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_examples
		(id, meaning_id, example_text, example_order, status, created_at, updated_at)
		VALUES ($1, $2, 'An example.', 0, 'draft', $3, $3)`, legacyExampleID, legacyMeaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO usage_notes
		(id, meaning_id, note_type, note_text, note_order, status, created_at, updated_at)
		VALUES ($1, $2, 'grammar', 'A note.', -1, 'draft', $3, $3)`, legacyNoteID, legacyMeaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO journey_situations
		(id, slug, title, short_description, category, status, display_order, created_at, updated_at)
		VALUES ($1, 'legacy-order', 'Legacy order', 'Legacy position', 'study', 'draft', -1, $2, $2)`, legacySituationID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO journey_words
		(id, journey_situation_id, meaning_id, relevance_score, display_order, is_core, created_at, updated_at)
		VALUES ($1, $2, $3, 50, 0, false, $4, $4)`, legacyJourneyWordID, legacySituationID, legacyMeaningID, now)
	require.NoError(t, err)

	migration, err := os.ReadFile(curatedContentOrderingMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID constraints must preserve legacy ordering rows")

	validMeaningID := uuid.New()
	validExampleID := uuid.New()
	validNoteID := uuid.New()
	validSituationID := uuid.New()
	validJourneyWordID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings
		(id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at)
		VALUES ($1, $2, 'verb', 'arrange', 1, 'active', $3, $3)`, validMeaningID, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_examples
		(id, meaning_id, example_text, example_order, status, created_at, updated_at)
		VALUES ($1, $2, 'Put these in order.', 1, 'active', $3, $3)`, validExampleID, validMeaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO usage_notes
		(id, meaning_id, note_type, note_text, note_order, status, created_at, updated_at)
		VALUES ($1, $2, 'grammar', 'Used as a verb.', 1, 'active', $3, $3)`, validNoteID, validMeaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO journey_situations
		(id, slug, title, short_description, category, status, display_order, created_at, updated_at)
		VALUES ($1, 'valid-order', 'Valid order', 'One-based position', 'study', 'active', 1, $2, $2)`, validSituationID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO journey_words
		(id, journey_situation_id, meaning_id, relevance_score, display_order, is_core, created_at, updated_at)
		VALUES ($1, $2, $3, 50, NULL, false, $4, $4)`, validJourneyWordID, validSituationID, validMeaningID, now)
	require.NoError(t, err, "nullable journey-word position remains valid")
	_, err = db.ExecContext(ctx, `UPDATE journey_words SET display_order = 1 WHERE id = $1`, validJourneyWordID)
	require.NoError(t, err, "one-based journey-word position remains valid when present")

	invalidInserts := []struct {
		name  string
		query string
		args  []any
	}{
		{name: "meaning zero", query: `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at) VALUES ($1, $2, 'noun', 'invalid', 0, 'draft', $3, $3)`, args: []any{uuid.New(), wordID, now}},
		{name: "example negative", query: `INSERT INTO word_examples (id, meaning_id, example_text, example_order, status, created_at, updated_at) VALUES ($1, $2, 'invalid', -1, 'draft', $3, $3)`, args: []any{uuid.New(), validMeaningID, now}},
		{name: "note zero", query: `INSERT INTO usage_notes (id, meaning_id, note_type, note_text, note_order, status, created_at, updated_at) VALUES ($1, $2, 'grammar', 'invalid', 0, 'draft', $3, $3)`, args: []any{uuid.New(), validMeaningID, now}},
		{name: "situation negative", query: `INSERT INTO journey_situations (id, slug, title, short_description, category, status, display_order, created_at, updated_at) VALUES ($1, $2, 'Invalid', 'Invalid position', 'study', 'draft', -1, $3, $3)`, args: []any{uuid.New(), "invalid-" + uuid.NewString(), now}},
		{name: "journey word zero", query: `INSERT INTO journey_words (id, journey_situation_id, meaning_id, relevance_score, display_order, is_core, created_at, updated_at) VALUES ($1, $2, $3, 50, 0, false, $4, $4)`, args: []any{uuid.New(), validSituationID, legacyMeaningID, now}},
	}
	for _, tc := range invalidInserts {
		t.Run("reject insert "+tc.name, func(t *testing.T) {
			_, insertErr := db.ExecContext(ctx, tc.query, tc.args...)
			requireCuratedOrderingCheckViolation(t, insertErr)
		})
	}

	invalidUpdates := []struct {
		name  string
		query string
		id    uuid.UUID
	}{
		{name: "meaning", query: `UPDATE word_meanings SET meaning_order = 0 WHERE id = $1`, id: validMeaningID},
		{name: "example", query: `UPDATE word_examples SET example_order = -1 WHERE id = $1`, id: validExampleID},
		{name: "note", query: `UPDATE usage_notes SET note_order = 0 WHERE id = $1`, id: validNoteID},
		{name: "situation", query: `UPDATE journey_situations SET display_order = -1 WHERE id = $1`, id: validSituationID},
		{name: "journey word", query: `UPDATE journey_words SET display_order = 0 WHERE id = $1`, id: validJourneyWordID},
	}
	for _, tc := range invalidUpdates {
		t.Run("reject update "+tc.name, func(t *testing.T) {
			_, updateErr := db.ExecContext(ctx, tc.query, tc.id)
			requireCuratedOrderingCheckViolation(t, updateErr)
		})
	}

	legacyRows := []struct {
		table string
		id    uuid.UUID
	}{
		{table: "word_meanings", id: legacyMeaningID},
		{table: "word_examples", id: legacyExampleID},
		{table: "usage_notes", id: legacyNoteID},
		{table: "journey_situations", id: legacySituationID},
		{table: "journey_words", id: legacyJourneyWordID},
	}
	for _, legacy := range legacyRows {
		t.Run("legacy row protected on update "+legacy.table, func(t *testing.T) {
			_, updateErr := db.ExecContext(ctx,
				"UPDATE "+pq.QuoteIdentifier(legacy.table)+" SET updated_at = updated_at + interval '1 second' WHERE id = $1",
				legacy.id)
			requireCuratedOrderingCheckViolation(t, updateErr)
		})
	}

	for _, name := range []string{
		"word_meanings_meaning_order_positive",
		"word_examples_example_order_positive",
		"usage_notes_note_order_positive",
		"journey_situations_display_order_positive",
		"journey_words_display_order_positive",
	} {
		var validated bool
		var definition string
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT convalidated, pg_get_constraintdef(oid)
			FROM pg_constraint
			WHERE connamespace = current_schema()::regnamespace AND conname = $1`, name).Scan(&validated, &definition))
		assert.False(t, validated, name)
		assert.Contains(t, definition, "> 0", name)
	}
}

func requireCuratedOrderingCheckViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
