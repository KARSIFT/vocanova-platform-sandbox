package pronunciation

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type failedRecoverer interface {
	RecoverFailed(context.Context, uuid.UUID, uuid.UUID, CatalogBudget) (CatalogReport, error)
}

func failedSignalFixture(t *testing.T) (*sql.DB, *CatalogOperator, uuid.UUID, uuid.UUID, string) {
	t.Helper()
	db, u := seedCatalogFixture(t)
	p := &fakeProvider{identity: NewOpenAIProvider(OpenAIConfig{}).Identity(), fn: func(context.Context, Utterance) ([]byte, error) { return silentMP3, nil }}
	o := NewCatalogOperator(db, p)
	job := uuid.New()
	_, err := o.Generate(t.Context(), job, CatalogBudget{1, 100, 60})
	require.Error(t, err)
	r, err := o.JobReport(t.Context(), job)
	require.NoError(t, err)
	require.Len(t, r.Receipts, 1)
	require.Equal(t, "failed", r.Receipts[0].Status)
	require.Positive(t, r.Receipts[0].AudioSeconds)
	_, err = db.Exec(`UPDATE pronunciation_audio_cache SET lease_until=now()-interval '1 second' WHERE cache_key=$1`, r.Receipts[0].CacheKey)
	require.NoError(t, err)
	return db, o, job, uuid.MustParse(r.Receipts[0].ID), utteranceCacheKey(o.identity(), u)
}
func TestCatalogPostgreSQLExplicitSignalRecovery(t *testing.T) {
	db, o, originalJob, original, key := failedSignalFixture(t)
	p := &fakeProvider{identity: o.identity()}
	o.provider = p
	ordinary, err := o.Generate(t.Context(), uuid.New(), CatalogBudget{1, 100, 60})
	require.NoError(t, err)
	require.Zero(t, ordinary.Attempted)
	require.Zero(t, p.calls)
	recovery, ok := any(o).(failedRecoverer)
	require.True(t, ok, "explicit operator signal recovery must be available")
	var before string
	require.NoError(t, db.QueryRow(`SELECT row_to_json(a)::text FROM pronunciation_catalog_attempts a WHERE id=$1`, original).Scan(&before))
	job := uuid.New()
	_, err = recovery.RecoverFailed(t.Context(), job, original, CatalogBudget{1, 100, 60})
	require.NoError(t, err)
	require.Equal(t, 1, p.calls)
	report, err := o.JobReport(t.Context(), job)
	require.NoError(t, err)
	require.Equal(t, 1, report.Attempted)
	require.Equal(t, "ready", report.Receipts[0].Status)
	var after string
	require.NoError(t, db.QueryRow(`SELECT row_to_json(a)::text FROM pronunciation_catalog_attempts a WHERE id=$1`, original).Scan(&after))
	require.Equal(t, before, after)
	var ready bool
	require.NoError(t, db.QueryRow(`SELECT status='ready' AND audio IS NOT NULL FROM pronunciation_audio_cache WHERE cache_key=$1`, key).Scan(&ready))
	require.True(t, ready)
	_, err = recovery.RecoverFailed(t.Context(), job, original, CatalogBudget{1, 100, 60})
	require.Error(t, err)
	_, err = recovery.RecoverFailed(t.Context(), uuid.New(), original, CatalogBudget{1, 100, 60})
	require.Error(t, err)
	_, err = o.Generate(t.Context(), job, CatalogBudget{1, 100, 60})
	require.Error(t, err)
	_, err = recovery.RecoverFailed(t.Context(), originalJob, original, CatalogBudget{1, 100, 60})
	require.Error(t, err)
	require.Equal(t, 1, p.calls)
	var public int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_audio_budget`).Scan(&public))
	require.Zero(t, public)
}

func TestCatalogPostgreSQLRecoveryRejectsIneligibleOriginal(t *testing.T) {
	cases := map[string]string{
		"zero decoded seconds":  `UPDATE pronunciation_catalog_attempts SET audio_seconds=0 WHERE id=$1`,
		"unfinished":            `UPDATE pronunciation_catalog_attempts SET finished_at=NULL WHERE id=$1`,
		"not failed":            `UPDATE pronunciation_catalog_attempts SET status='changed' WHERE id=$1`,
		"wrong provider":        `UPDATE pronunciation_catalog_jobs SET provider_identity='{}' WHERE id=(SELECT job_id FROM pronunciation_catalog_attempts WHERE id=$1)`,
		"recovery chain":        `UPDATE pronunciation_catalog_jobs SET recovery_of_attempt=$1 WHERE id=(SELECT job_id FROM pronunciation_catalog_attempts WHERE id=$1)`,
		"missing cache":         `DELETE FROM pronunciation_audio_cache WHERE generation_id=$1`,
		"pending cache":         `UPDATE pronunciation_audio_cache SET status='pending' WHERE generation_id=$1`,
		"live lease":            `UPDATE pronunciation_audio_cache SET lease_until=now()+interval '1 hour' WHERE generation_id=$1`,
		"different attempt day": `UPDATE pronunciation_audio_cache SET attempted_day=attempted_day-1 WHERE generation_id=$1`,
		"different generation":  `UPDATE pronunciation_audio_cache SET generation_id='00000000-0000-0000-0000-000000000001' WHERE generation_id=$1`,
		"changed definition":    `UPDATE word_meanings SET short_definition='New definition.' WHERE id=(SELECT meaning_id FROM pronunciation_catalog_attempts WHERE id=$1)`,
		"archived meaning":      `UPDATE word_meanings SET status='archived' WHERE id=(SELECT meaning_id FROM pronunciation_catalog_attempts WHERE id=$1)`,
	}
	for name, query := range cases {
		t.Run(name, func(t *testing.T) {
			db, o, _, original, _ := failedSignalFixture(t)
			_, err := db.Exec(query, original)
			require.NoError(t, err)
			p := &fakeProvider{identity: o.identity()}
			o.provider = p
			job := uuid.New()
			_, err = o.RecoverFailed(t.Context(), job, original, CatalogBudget{1, 100, 60})
			require.Error(t, err)
			require.Zero(t, p.calls)
			var n int
			require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_catalog_jobs WHERE id=$1`, job).Scan(&n))
			require.Zero(t, n)
		})
	}
	t.Run("ready cannot be replaced", func(t *testing.T) {
		db, o, _, original, key := failedSignalFixture(t)
		_, err := db.Exec(`UPDATE pronunciation_audio_cache SET status='ready',audio=$2 WHERE cache_key=$1`, key, testMP3)
		require.NoError(t, err)
		p := &fakeProvider{identity: o.identity()}
		o.provider = p
		_, err = o.RecoverFailed(t.Context(), uuid.New(), original, CatalogBudget{1, 100, 60})
		require.Error(t, err)
		require.Zero(t, p.calls)
	})
}

func TestCatalogPostgreSQLRecoveryConsumesFailureAndCancellation(t *testing.T) {
	for _, mode := range []string{"signal", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			db, o, _, original, key := failedSignalFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			job := uuid.New()
			p := &fakeProvider{identity: o.identity(), fn: func(context.Context, Utterance) ([]byte, error) {
				var linked bool
				require.NoError(t, db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pronunciation_catalog_jobs j JOIN pronunciation_catalog_attempts a ON a.job_id=j.id WHERE j.id=$1 AND j.recovery_of_attempt=$2 AND a.status='reserved')`, job, original).Scan(&linked))
				require.True(t, linked, "durable linkage/reservation must precede provider")
				if mode == "cancel" {
					cancel()
					return nil, context.Canceled
				}
				return silentMP3, nil
			}}
			o.provider = p
			_, err := o.RecoverFailed(ctx, job, original, CatalogBudget{1, 100, 60})
			require.Error(t, err)
			require.Equal(t, 1, p.calls)
			r, err := o.JobReport(t.Context(), job)
			require.NoError(t, err)
			require.Equal(t, original.String(), r.RecoveryOfAttempt)
			require.Equal(t, 1, r.Attempted)
			require.Equal(t, "failed", r.Receipts[0].Status)
			for _, budget := range []CatalogBudget{{1, 100, 60}, {1, 101, 60}, {1, 100, 61}} {
				_, err = o.RecoverFailed(t.Context(), job, original, budget)
				require.Error(t, err)
			}
			_, err = o.RecoverFailed(t.Context(), uuid.New(), uuid.MustParse(r.Receipts[0].ID), CatalogBudget{1, 100, 60})
			require.Error(t, err)
			// Even restoring the old cache generation cannot evade unique original linkage.
			_, err = db.Exec(`UPDATE pronunciation_audio_cache SET generation_id=$2,lease_until=now()-interval '1 second' WHERE cache_key=$1`, key, original)
			require.NoError(t, err)
			_, err = o.RecoverFailed(t.Context(), uuid.New(), original, CatalogBudget{1, 100, 60})
			require.Error(t, err)
			require.Equal(t, 1, p.calls)
		})
	}
}

func TestCatalogPostgreSQLRecoverySerializesPublicAndOtherOperator(t *testing.T) {
	db, o, _, original, key := failedSignalFixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	p := &fakeProvider{identity: o.identity(), fn: func(context.Context, Utterance) ([]byte, error) { close(started); <-release; return testMP3, nil }}
	o.provider = p
	done := make(chan error, 1)
	go func() {
		_, err := o.RecoverFailed(t.Context(), uuid.New(), original, CatalogBudget{1, 100, 60})
		done <- err
	}()
	<-started
	second := NewCatalogOperator(db, &fakeProvider{identity: o.identity()})
	_, err := second.RecoverFailed(t.Context(), uuid.New(), original, CatalogBudget{1, 100, 60})
	require.ErrorIs(t, err, ErrBusy)
	repo := &postgresRepository{db}
	_, err = repo.Reserve(t.Context(), key, 7, time.Now())
	require.ErrorIs(t, err, ErrBusy)
	var calls int
	require.NoError(t, db.QueryRow(`SELECT COALESCE(sum(reserved_generations),0) FROM pronunciation_audio_budget`).Scan(&calls))
	require.Zero(t, calls)
	close(release)
	require.NoError(t, <-done)
	require.Equal(t, 1, p.calls)
}

func TestCatalogRecoveryRejectsInvalidBoundsBeforeDatabase(t *testing.T) {
	o := NewCatalogOperator(nil, &fakeProvider{})
	job, original := uuid.New(), uuid.New()
	for _, b := range []CatalogBudget{{2, 100, 60}, {0, 100, 60}, {1, 0, 60}, {1, 100, 0}} {
		_, err := o.RecoverFailed(t.Context(), job, original, b)
		require.Error(t, err)
	}
	for _, ids := range [][2]uuid.UUID{{uuid.Nil, original}, {job, uuid.Nil}, {job, job}} {
		_, err := o.RecoverFailed(t.Context(), ids[0], ids[1], CatalogBudget{1, 100, 60})
		require.Error(t, err)
	}
}

func TestCatalogPostgreSQLRecoveryRejectsPriorDay(t *testing.T) {
	db, o, _, original, key := failedSignalFixture(t)
	_, err := db.Exec(`UPDATE pronunciation_catalog_attempts SET created_at=created_at-interval '1 day' WHERE id=$1`, original)
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE pronunciation_audio_cache SET attempted_day=attempted_day-1 WHERE cache_key=$1`, key)
	require.NoError(t, err)
	p := &fakeProvider{identity: o.identity()}
	o.provider = p
	_, err = o.RecoverFailed(t.Context(), uuid.New(), original, CatalogBudget{1, 100, 60})
	require.Error(t, err)
	require.Zero(t, p.calls)
}

func TestCatalogPostgreSQLRecoveryHonorsCanonicalCompletionAndCharacterBound(t *testing.T) {
	db, o, _, original, key := failedSignalFixture(t)
	p := &fakeProvider{identity: o.identity()}
	o.provider = p
	_, err := o.RecoverFailed(t.Context(), uuid.New(), original, CatalogBudget{1, 1, 60})
	require.Error(t, err)
	require.Zero(t, p.calls)
	p.fn = func(context.Context, Utterance) ([]byte, error) {
		_, err := db.Exec(`UPDATE word_meanings SET short_definition='Changed while recovering.' WHERE id=(SELECT meaning_id FROM pronunciation_catalog_attempts WHERE id=$1)`, original)
		require.NoError(t, err)
		return testMP3, nil
	}
	job := uuid.New()
	_, err = o.RecoverFailed(t.Context(), job, original, CatalogBudget{1, 100, 60})
	require.NoError(t, err)
	report, err := o.JobReport(t.Context(), job)
	require.NoError(t, err)
	require.Equal(t, "changed", report.Receipts[0].Status)
	require.Equal(t, 1, report.Attempted)
	var rejected bool
	require.NoError(t, db.QueryRow(`SELECT status='failed' AND audio IS NULL FROM pronunciation_audio_cache WHERE cache_key=$1`, key).Scan(&rejected))
	require.True(t, rejected)
}

func TestCatalogPostgreSQLRecoveryLeaseStartsAfterKeyWait(t *testing.T) {
	db, o, _, original, key := failedSignalFixture(t)
	barrier, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer barrier.Rollback()
	_, err = barrier.Exec(`SELECT pg_advisory_xact_lock(hashtextextended('pronunciation:' || $1,0))`, key)
	require.NoError(t, err)
	remaining := make(chan float64, 1)
	o.provider = &fakeProvider{identity: o.identity(), fn: func(context.Context, Utterance) ([]byte, error) {
		var seconds float64
		err := db.QueryRow(`SELECT extract(epoch FROM lease_until-clock_timestamp()) FROM pronunciation_audio_cache WHERE cache_key=$1`, key).Scan(&seconds)
		if err != nil {
			return nil, err
		}
		remaining <- seconds
		return testMP3, nil
	}}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := o.RecoverFailed(ctx, uuid.New(), original, CatalogBudget{1, 100, 60}); done <- err }()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE wait_event='advisory' AND query LIKE '%pg_advisory_xact_lock%')`).Scan(&waiting)
		return err == nil && waiting
	}, 3*time.Second, 10*time.Millisecond)
	time.Sleep(2 * time.Second)
	require.NoError(t, barrier.Commit())
	require.NoError(t, <-done)
	require.Greater(t, <-remaining, 19.0, "lease must start at reservation, not before per-key lock wait")
}
