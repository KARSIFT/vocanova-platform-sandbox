package knowledgecatalog

import (
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
