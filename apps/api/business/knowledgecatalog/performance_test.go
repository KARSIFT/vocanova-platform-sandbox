package knowledgecatalog

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Mirrors the SQL comparison's mixed active/removed histories while measuring
// the complete service read, including item revisions, decoding and allocations.
func BenchmarkKnowledgeMapReadMatureHistory(b *testing.B) {
	db := isolatedCatalogDB(b)
	preloadAuthoredCore(b, db)
	tx, err := db.BeginTx(b.Context(), nil)
	require.NoError(b, err)
	require.NoError(b, Seed(b.Context(), tx))
	require.NoError(b, tx.Commit())
	svc := NewService(NewPostgreSQLRepository(db), nil)
	for _, history := range []int{100, 1000, 5000} {
		user := uuid.New()
		_, err = db.Exec(`INSERT INTO users(id,email,status,created_at,updated_at) VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, user, user.String()+"@mature-benchmark.invalid")
		require.NoError(b, err)
		_, err = db.Exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at,deleted_at)
 SELECT gen_random_uuid(),$1,kc.meaning_id,
 CASE WHEN kc.rank%11=0 THEN 'ignored' WHEN kc.rank%13=0 THEN 'archived' ELSE 'learning' END,
 'manual','2026-01-01','2026-01-01','2026-10-01',
 CASE WHEN h=0 THEN NULL WHEN kc.rank%3=0 THEN '2026-10-03'::timestamptz ELSE '2026-10-02'::timestamptz+h*interval '1 day' END
 FROM knowledge_catalog_entries kc CROSS JOIN generate_series(0,2) h
 WHERE kc.rank<=$2 AND (h<>0 OR kc.rank%7<>0)`, user, history)
		require.NoError(b, err)
		_, err = db.Exec(`INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at)
 SELECT $1,meaning_id,true,'','2026-10-01' FROM knowledge_catalog_entries WHERE rank<=$2 AND rank%5=0`, user, history)
		require.NoError(b, err)
		_, err = db.Exec(`ANALYZE user_words; ANALYZE user_word_knowledge; ANALYZE knowledge_catalog_entries`)
		require.NoError(b, err)
		b.Run(fmt.Sprintf("Meanings%d", history), func(b *testing.B) {
			_, e := svc.Read(b.Context(), user, 1)
			require.NoError(b, e) // Untimed warm-up, matching the SQL experiment.
			times := make([]time.Duration, 0, b.N)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				start := time.Now()
				_, e := svc.Read(b.Context(), user, 1)
				require.NoError(b, e)
				times = append(times, time.Since(start))
			}
			b.StopTimer()
			sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
			if len(times) > 0 {
				b.ReportMetric(float64(times[(len(times)*95+99)/100-1].Nanoseconds())/1e6, "p95-ms")
			}
		})
	}
}

// Run explicitly against the disposable migrated PostgreSQL test instance.
// These are local service/DB timings, not browser or production measurements.
func BenchmarkKnowledgeMapRead(b *testing.B) {
	db := isolatedCatalogDB(b)
	preloadAuthoredCore(b, db)
	tx, err := db.BeginTx(b.Context(), nil)
	require.NoError(b, err)
	require.NoError(b, Seed(b.Context(), tx))
	require.NoError(b, tx.Commit())
	user := uuid.New()
	_, err = db.Exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,$2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, user, user.String()+"@catalog-benchmark.invalid")
	require.NoError(b, err)
	svc := NewService(NewPostgreSQLRepository(db), nil)
	measure := func(b *testing.B) {
		times := make([]time.Duration, 0, b.N)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			start := time.Now()
			_, e := svc.Read(b.Context(), user, 1)
			require.NoError(b, e)
			times = append(times, time.Since(start))
		}
		b.StopTimer()
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
		if len(times) > 0 {
			b.ReportMetric(float64(times[(len(times)-1)*95/100].Nanoseconds())/1e6, "p95-ms")
		}
	}
	b.Run("Fresh", measure)
	state, err := svc.Read(b.Context(), user, 1)
	require.NoError(b, err)
	for _, item := range state.CurrentBand.Items[:5] {
		_, err = svc.Decide(b.Context(), DecisionRequest{UserID: user, MeaningID: item.MeaningID, Decision: "learn", ExpectedRevision: item.Revision, IdempotencyKey: uuid.NewString()})
		require.NoError(b, err)
	}
	b.Run("FiveSaved", measure)
}
