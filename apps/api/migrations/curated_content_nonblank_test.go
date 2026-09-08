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

const curatedContentNonblankMigration = "20260908220000_voc1426_curated_content_nonblank.sql"

func TestCuratedContentNonblankMigrationCarriesAllInvariants(t *testing.T) {
	body, err := os.ReadFile(curatedContentNonblankMigration)
	require.NoError(t, err)
	text := string(body)

	for _, invariant := range []string{
		"canonical_words_text_nonblank",
		"canonical_words_normalized_text_nonblank",
		"word_meanings_short_definition_nonblank",
		"word_examples_example_text_nonblank",
		"usage_notes_note_text_nonblank",
		"journey_situations_slug_nonblank",
		"journey_situations_title_nonblank",
		"journey_situations_short_description_nonblank",
		"~ '[^[:space:]]'",
		"NOT VALID",
	} {
		assert.Contains(t, text, invariant)
	}

	schemaChecks := map[string][]string{
		"../ent/schema/canonicalword.go": {
			`"text_nonblank":            "text ~ '[^[:space:]]'"`,
			`"normalized_text_nonblank": "normalized_text ~ '[^[:space:]]'"`,
		},
		"../ent/schema/wordmeaning.go": {
			`"short_definition_nonblank": "short_definition ~ '[^[:space:]]'"`,
		},
		"../ent/schema/wordexample.go": {
			`"example_text_nonblank": "example_text ~ '[^[:space:]]'"`,
		},
		"../ent/schema/usagenote.go": {
			`"note_text_nonblank": "note_text ~ '[^[:space:]]'"`,
		},
		"../ent/schema/journeysituation.go": {
			`"slug_nonblank":              "slug ~ '[^[:space:]]'"`,
			`"title_nonblank":             "title ~ '[^[:space:]]'"`,
			`"short_description_nonblank": "short_description ~ '[^[:space:]]'"`,
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

func TestCuratedContentNonblankAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_curated_nonblank_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
	meaningID := uuid.New()
	insertCuratedWord(t, ctx, db, wordID, "learn", "learn", now)
	insertCuratedMeaning(t, ctx, db, meaningID, wordID, "to gain knowledge", now)

	legacyRows := []struct {
		table string
		id    uuid.UUID
	}{
		{table: "canonical_words", id: uuid.New()},
		{table: "word_meanings", id: uuid.New()},
		{table: "word_examples", id: uuid.New()},
		{table: "usage_notes", id: uuid.New()},
		{table: "journey_situations", id: uuid.New()},
	}
	insertCuratedWord(t, ctx, db, legacyRows[0].id, " \t\n", " \t\n", now)
	insertCuratedMeaning(t, ctx, db, legacyRows[1].id, legacyRows[0].id, " \t\n", now)
	_, err = db.ExecContext(ctx, `INSERT INTO word_examples
		(id, meaning_id, example_text, example_order, status, created_at, updated_at)
		VALUES ($1, $2, E' \t\n', 1, 'draft', $3, $3)`, legacyRows[2].id, meaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO usage_notes
		(id, meaning_id, note_type, note_text, note_order, status, created_at, updated_at)
		VALUES ($1, $2, 'grammar', E' \t\n', 1, 'draft', $3, $3)`, legacyRows[3].id, meaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO journey_situations
		(id, slug, title, short_description, category, status, display_order, created_at, updated_at)
		VALUES ($1, E' \t\n', E' \t\n', E' \t\n', 'study', 'draft', 1, $2, $2)`, legacyRows[4].id, now)
	require.NoError(t, err)

	migration, err := os.ReadFile(curatedContentNonblankMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID constraints must preserve legacy blank content")

	validWordID := uuid.New()
	validMeaningID := uuid.New()
	validExampleID := uuid.New()
	validNoteID := uuid.New()
	validSituationID := uuid.New()
	insertCuratedWord(t, ctx, db, validWordID, "airport", "airport", now)
	insertCuratedMeaning(t, ctx, db, validMeaningID, validWordID, "a place where aircraft arrive", now)
	_, err = db.ExecContext(ctx, `INSERT INTO word_examples
		(id, meaning_id, example_text, example_order, status, created_at, updated_at)
		VALUES ($1, $2, 'We arrived at the airport.', 1, 'active', $3, $3)`, validExampleID, validMeaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO usage_notes
		(id, meaning_id, note_type, note_text, note_order, status, created_at, updated_at)
		VALUES ($1, $2, 'collocation', 'at the airport', 1, 'active', $3, $3)`, validNoteID, validMeaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO journey_situations
		(id, slug, title, short_description, category, status, display_order, created_at, updated_at)
		VALUES ($1, 'airport', 'Airport', 'Useful airport vocabulary', 'travel', 'active', 2, $2, $2)`, validSituationID, now)
	require.NoError(t, err)

	invalidInserts := []struct {
		name  string
		query string
		args  []any
	}{
		{name: "word text", query: `INSERT INTO canonical_words (id, text, normalized_text, word_type, language_code, status, created_at, updated_at) VALUES ($1, E' \t\n', $2, 'word', 'en', 'draft', $3, $3)`, args: []any{uuid.New(), "blank-word-text", now}},
		{name: "word normalized text", query: `INSERT INTO canonical_words (id, text, normalized_text, word_type, language_code, status, created_at, updated_at) VALUES ($1, 'blank normalized', E' \t\n', 'word', 'en', 'draft', $2, $2)`, args: []any{uuid.New(), now}},
		{name: "meaning definition", query: `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at) VALUES ($1, $2, 'noun', E' \t\n', 2, 'draft', $3, $3)`, args: []any{uuid.New(), wordID, now}},
		{name: "example text", query: `INSERT INTO word_examples (id, meaning_id, example_text, example_order, status, created_at, updated_at) VALUES ($1, $2, E' \t\n', 2, 'draft', $3, $3)`, args: []any{uuid.New(), meaningID, now}},
		{name: "note text", query: `INSERT INTO usage_notes (id, meaning_id, note_type, note_text, note_order, status, created_at, updated_at) VALUES ($1, $2, 'grammar', E' \t\n', 2, 'draft', $3, $3)`, args: []any{uuid.New(), meaningID, now}},
		{name: "situation slug", query: `INSERT INTO journey_situations (id, slug, title, short_description, category, status, display_order, created_at, updated_at) VALUES ($1, E' \t\n', 'Title', 'Description', 'travel', 'draft', 3, $2, $2)`, args: []any{uuid.New(), now}},
		{name: "situation title", query: `INSERT INTO journey_situations (id, slug, title, short_description, category, status, display_order, created_at, updated_at) VALUES ($1, $2, E' \t\n', 'Description', 'travel', 'draft', 4, $3, $3)`, args: []any{uuid.New(), "blank-title-" + uuid.NewString(), now}},
		{name: "situation description", query: `INSERT INTO journey_situations (id, slug, title, short_description, category, status, display_order, created_at, updated_at) VALUES ($1, $2, 'Title', E' \t\n', 'travel', 'draft', 5, $3, $3)`, args: []any{uuid.New(), "blank-description-" + uuid.NewString(), now}},
	}
	for _, tc := range invalidInserts {
		t.Run("reject insert "+tc.name, func(t *testing.T) {
			_, insertErr := db.ExecContext(ctx, tc.query, tc.args...)
			requireCuratedContentCheckViolation(t, insertErr)
		})
	}

	invalidUpdates := []struct {
		name  string
		query string
		id    uuid.UUID
	}{
		{name: "word text", query: `UPDATE canonical_words SET text = E' \t\n' WHERE id = $1`, id: validWordID},
		{name: "word normalized text", query: `UPDATE canonical_words SET normalized_text = E' \t\n' WHERE id = $1`, id: validWordID},
		{name: "meaning definition", query: `UPDATE word_meanings SET short_definition = E' \t\n' WHERE id = $1`, id: validMeaningID},
		{name: "example text", query: `UPDATE word_examples SET example_text = E' \t\n' WHERE id = $1`, id: validExampleID},
		{name: "note text", query: `UPDATE usage_notes SET note_text = E' \t\n' WHERE id = $1`, id: validNoteID},
		{name: "situation slug", query: `UPDATE journey_situations SET slug = E' \t\n' WHERE id = $1`, id: validSituationID},
		{name: "situation title", query: `UPDATE journey_situations SET title = E' \t\n' WHERE id = $1`, id: validSituationID},
		{name: "situation description", query: `UPDATE journey_situations SET short_description = E' \t\n' WHERE id = $1`, id: validSituationID},
	}
	for _, tc := range invalidUpdates {
		t.Run("reject update "+tc.name, func(t *testing.T) {
			_, updateErr := db.ExecContext(ctx, tc.query, tc.id)
			requireCuratedContentCheckViolation(t, updateErr)
		})
	}

	for _, legacy := range legacyRows {
		t.Run("legacy row protected on update "+legacy.table, func(t *testing.T) {
			_, updateErr := db.ExecContext(ctx,
				"UPDATE "+pq.QuoteIdentifier(legacy.table)+" SET updated_at = updated_at + interval '1 second' WHERE id = $1",
				legacy.id)
			requireCuratedContentCheckViolation(t, updateErr)
		})
	}

	constraintNames := []string{
		"canonical_words_text_nonblank",
		"canonical_words_normalized_text_nonblank",
		"word_meanings_short_definition_nonblank",
		"word_examples_example_text_nonblank",
		"usage_notes_note_text_nonblank",
		"journey_situations_slug_nonblank",
		"journey_situations_title_nonblank",
		"journey_situations_short_description_nonblank",
	}
	for _, name := range constraintNames {
		var validated bool
		var definition string
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT convalidated, pg_get_constraintdef(oid)
			FROM pg_constraint
			WHERE connamespace = current_schema()::regnamespace AND conname = $1`, name).Scan(&validated, &definition))
		assert.False(t, validated, name)
		assert.Contains(t, definition, "[^[:space:]]", name)
	}
}

func insertCuratedWord(t *testing.T, ctx context.Context, db *sql.DB, id uuid.UUID, text, normalized string, now time.Time) {
	t.Helper()
	_, err := db.ExecContext(ctx, `INSERT INTO canonical_words
		(id, text, normalized_text, word_type, language_code, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'word', 'en', 'draft', $4, $4)`, id, text, normalized, now)
	require.NoError(t, err)
}

func insertCuratedMeaning(t *testing.T, ctx context.Context, db *sql.DB, id, wordID uuid.UUID, definition string, now time.Time) {
	t.Helper()
	_, err := db.ExecContext(ctx, `INSERT INTO word_meanings
		(id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at)
		VALUES ($1, $2, 'verb', $3, 1, 'draft', $4, $4)`, id, wordID, definition, now)
	require.NoError(t, err)
}

func requireCuratedContentCheckViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
