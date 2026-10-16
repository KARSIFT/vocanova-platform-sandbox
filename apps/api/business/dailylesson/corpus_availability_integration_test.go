package dailylesson

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/stretchr/testify/require"
)

// Exercise the production SQL pool and semantic guard for every real catalog
// target. This establishes question availability, not semantic gold correctness.
func TestPostgreSQLFullCatalogHasEnoughDailyLessonChoices(t *testing.T) {
	f := setup(t, 0) // Fresh isolated schema, never shared or production content.
	ctx := t.Context()
	// Remove synthetic curriculum links so only the actual authored pool qualifies.
	_, e := f.db.ExecContext(ctx, `DELETE FROM journey_words`)
	require.NoError(t, e)
	_, e = f.db.ExecContext(ctx, `DELETE FROM journey_situations`)
	require.NoError(t, e)
	raw, e := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, e)
	for _, q := range []string{
		`INSERT INTO canonical_words(id,text,normalized_text,word_type,language_code,status,difficulty_level,frequency_rank,created_at,updated_at) SELECT (x->>'id')::uuid,x->>'text',x->>'normalized_text',x->>'word_type',x->>'language_code',x->>'status',x->>'difficulty_level',(x->>'frequency_rank')::integer,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'canonical_words') x`,
		`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,learner_definition,meaning_order,status,difficulty_level,created_at,updated_at) SELECT (x->>'id')::uuid,(x->>'word_id')::uuid,x->>'part_of_speech',x->>'short_definition',x->>'learner_definition',(x->>'meaning_order')::integer,x->>'status',x->>'difficulty_level',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'word_meanings') x`,
		`INSERT INTO journey_situations(id,slug,title,short_description,level_band,category,status,display_order,created_at,updated_at) SELECT (x->>'id')::uuid,x->>'slug',x->>'title',x->>'short_description',x->>'level_band',x->>'category',x->>'status',(x->>'display_order')::integer,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'journey_situations') x`,
		`INSERT INTO journey_words(id,journey_situation_id,meaning_id,relevance_score,display_order,is_core,created_at,updated_at) SELECT (x->>'id')::uuid,(x->>'journey_situation_id')::uuid,(x->>'meaning_id')::uuid,(x->>'relevance_score')::numeric,(x->>'display_order')::integer,(x->>'is_core')::boolean,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'journey_words') x`,
	} {
		_, e = f.db.ExecContext(ctx, q, string(raw))
		require.NoError(t, e)
	}
	tx, e := f.db.BeginTx(ctx, nil)
	require.NoError(t, e)
	require.NoError(t, knowledgecatalog.Seed(ctx, tx))
	require.NoError(t, tx.Commit())
	// Seeded fixtures need current planner statistics before the exhaustive batch.
	for _, table := range []string{"canonical_words", "word_meanings", "knowledge_catalog_entries", "journey_words", "journey_situations"} {
		_, e = f.db.ExecContext(ctx, "ANALYZE "+table)
		require.NoError(t, e)
	}
	// One lateral batch uses the exact production query for all targets; avoid
	// 10,000 separate API requests or reimplementing its selection rules in Go.
	poolSQL := strings.NewReplacer("$1", "corpus_target.word_id", "$2", "corpus_target.definition", "$3", "corpus_target.word_text", "$4", "corpus_target.meaning_id").Replace(candidateDefinitionsSQL)
	query := `SELECT corpus_target.rank,corpus_target.word_text,candidate.short_definition,candidate.text FROM (SELECT kc.rank,kc.word_id,kc.meaning_id,cw.text AS word_text,wm.short_definition AS definition FROM knowledge_catalog_entries kc JOIN canonical_words cw ON cw.id=kc.word_id JOIN word_meanings wm ON wm.id=kc.meaning_id) corpus_target LEFT JOIN LATERAL (` + poolSQL + `) candidate ON true ORDER BY corpus_target.rank`
	rows, e := f.db.QueryContext(ctx, query)
	require.NoError(t, e)
	defer rows.Close()
	counts := map[int]int{}
	words := map[int]string{}
	for rows.Next() {
		var rank int
		var target string
		var definition, candidate *string
		require.NoError(t, rows.Scan(&rank, &target, &definition, &candidate))
		words[rank] = target
		if candidate == nil {
			continue
		}
		alias, e := knowledgecatalog.SemanticAliases(target, *candidate)
		require.NoError(t, e)
		if !alias {
			counts[rank]++
		}
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	require.Len(t, words, knowledgecatalog.TotalWords)
	minimum := 10000
	shortCounts := map[string]int{}
	minimumWords := []string{}
	failures := []string{}
	for rank := 1; rank <= knowledgecatalog.TotalWords; rank++ {
		if len(words[rank]) <= 2 {
			shortCounts[words[rank]] = counts[rank]
		}
		if counts[rank] < minimum {
			minimum = counts[rank]
			minimumWords = []string{words[rank]}
		} else if counts[rank] == minimum {
			minimumWords = append(minimumWords, words[rank])
		}
		if counts[rank] < 2 {
			failures = append(failures, fmt.Sprintf("%d:%s has %d safe distractors", rank, words[rank], counts[rank]))
		}
	}
	require.Empty(t, failures, "each target needs correct answer plus at least two alternatives")
	t.Logf("Checked all %d real targets; minimum safe distractors after production pool + alias guard: %d", len(words), minimum)
	t.Logf("Minimum words: %v; short target availability: %v", minimumWords, shortCounts)
}
