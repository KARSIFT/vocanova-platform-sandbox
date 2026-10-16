package knowledgecatalog

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func isolatedCatalogDB(t testing.TB) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := "catalog_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, e := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); require.NoError(t, e); admin.Close() })
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	db.SetMaxOpenConns(4)
	t.Cleanup(func() { db.Close() })
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	for _, p := range paths {
		if strings.HasSuffix(p, ".down.sql") {
			continue
		}
		raw, e := os.ReadFile(p)
		require.NoError(t, e)
		_, e = db.Exec(string(raw))
		require.NoError(t, e, p)
	}
	return db
}
func preloadAuthoredCore(t testing.TB, db *sql.DB) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	for _, q := range []string{
		`INSERT INTO canonical_words(id,text,normalized_text,word_type,language_code,status,difficulty_level,frequency_rank,created_at,updated_at)
 SELECT (x->>'id')::uuid,x->>'text',x->>'normalized_text',x->>'word_type',x->>'language_code',x->>'status',x->>'difficulty_level',(x->>'frequency_rank')::integer,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'canonical_words') x`,
		`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,learner_definition,meaning_order,status,difficulty_level,created_at,updated_at)
 SELECT (x->>'id')::uuid,(x->>'word_id')::uuid,x->>'part_of_speech',x->>'short_definition',x->>'learner_definition',(x->>'meaning_order')::integer,x->>'status',x->>'difficulty_level',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'word_meanings') x`,
		`INSERT INTO word_examples(id,meaning_id,example_text,example_order,difficulty_level,situation_label,status,created_at,updated_at)
 SELECT (x->>'id')::uuid,(x->>'meaning_id')::uuid,x->>'example_text',(x->>'example_order')::integer,x->>'difficulty_level',x->>'situation_label',x->>'status',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'word_examples') x`,
	} {
		_, err = db.Exec(q, string(raw))
		require.NoError(t, err)
	}
	return raw
}

// Exported only in test builds so cross-domain consistency checks can share the
// isolated schema and immutable authored fixtures without a production test API.
func IsolatedCatalogDBForTest(t testing.TB) *sql.DB       { return isolatedCatalogDB(t) }
func PreloadAuthoredCoreForTest(t testing.TB, db *sql.DB) { preloadAuthoredCore(t, db) }
