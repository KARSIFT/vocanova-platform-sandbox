package migrations_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeCatalogRejectsMeaningFromAnotherCanonicalWord(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	w1, w2, m := uuid.New(), uuid.New(), uuid.New()
	for _, w := range []uuid.UUID{w1, w2} {
		_, err = tx.ExecContext(t.Context(), `INSERT INTO canonical_words(id,text,normalized_text,status,created_at,updated_at)VALUES($1,$2,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, w, "catalog-owner-"+w.String())
		require.NoError(t, err)
	}
	_, err = tx.ExecContext(t.Context(), `INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','Synthetic ownership test',1,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, m, w1)
	require.NoError(t, err)
	// This deletion is private to this rollback-only transaction, even when seeded.
	_, err = tx.ExecContext(t.Context(), `DELETE FROM knowledge_catalog_entries WHERE rank=1`)
	require.NoError(t, err)
	_, err = tx.ExecContext(t.Context(), `INSERT INTO knowledge_catalog_entries(rank,word_id,meaning_id,source_rank,source_count,source_sense_key,source_name,catalog_version) VALUES(1,$1,$2,1,1,'synthetic','synthetic','english-10000-v1')`, w2, m)
	var pe *pq.Error
	require.ErrorAs(t, err, &pe)
	require.Equal(t, pq.ErrorCode("23503"), pe.Code)
	require.Equal(t, "knowledge_catalog_meaning_owner", pe.Constraint)
}
