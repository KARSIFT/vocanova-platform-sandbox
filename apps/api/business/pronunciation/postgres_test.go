package pronunciation

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func audioDB(t *testing.T) *sql.DB {
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
	schema := "audio_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn+" search_path="+schema+" application_name="+schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
		admin.Close()
	})
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(raw))
		require.NoError(t, err, path)
	}
	return db
}

func TestPronunciationPostgreSQLConcurrentMidnightClaimDoesNotOverwriteLease(t *testing.T) {
	db := audioDB(t)
	ctx := t.Context()
	repo := &postgresRepository{db}
	var appName string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT current_setting('application_name')`).Scan(&appName))
	barrierToken := uuid.New()
	barrierID := int64(binary.BigEndian.Uint64(barrierToken[:8]))
	// The insert trigger holds the first claimant just before publication. On
	// the broken implementation both dates reach it; with key serialization the
	// other caller waits on the shared cache key instead. Neither can complete
	// until the test releases its explicit barrier.
	_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE FUNCTION pronunciation_midnight_gate() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_advisory_xact_lock_shared(%d::bigint); RETURN NEW; END $$;
 CREATE TRIGGER pronunciation_midnight_gate BEFORE INSERT ON pronunciation_audio_cache FOR EACH ROW EXECUTE FUNCTION pronunciation_midnight_gate();`, barrierID))
	require.NoError(t, err)
	barrier, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer barrier.Rollback()
	_, err = barrier.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1::bigint)`, barrierID)
	require.NoError(t, err)
	key := cacheKey("midnight-shared-key")
	before := time.Date(2026, 10, 18, 23, 59, 59, 0, time.UTC)
	type result struct {
		reservation reservation
		err         error
	}
	results := make(chan result, 2)
	for _, now := range []time.Time{before, before.Add(2 * time.Second)} {
		go func(now time.Time) { r, e := repo.Reserve(ctx, key, 17, now); results <- result{r, e} }(now)
	}
	require.Eventually(t, func() bool {
		var waiting int
		err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_locks l JOIN pg_stat_activity a ON a.pid=l.pid WHERE l.locktype='advisory' AND NOT l.granted AND a.application_name=$1`, appName).Scan(&waiting)
		return err == nil && waiting == 2
	}, 3*time.Second, 5*time.Millisecond, "both first requests must be blocked at the deterministic claim barrier")
	require.NoError(t, barrier.Commit())
	var winner uuid.UUID
	success, busy := 0, 0
	for i := 0; i < 2; i++ {
		got := <-results
		if got.err == nil {
			success++
			winner = got.reservation.generation
		} else {
			require.ErrorIs(t, got.err, ErrBusy)
			busy++
		}
	}
	require.Equal(t, 1, success, "one cache key permits only one paid request across the UTC boundary")
	require.Equal(t, 1, busy)
	var stored uuid.UUID
	var spent, chargedDays int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT generation_id FROM pronunciation_audio_cache WHERE cache_key=$1`, key).Scan(&stored))
	require.Equal(t, winner, stored)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COALESCE(sum(reserved_characters),0),count(*) FILTER(WHERE reserved_characters>0) FROM pronunciation_audio_budget`).Scan(&spent, &chargedDays))
	require.Equal(t, 17, spent)
	require.Equal(t, 1, chargedDays, "the losing day cannot retain a second charge")
	require.NoError(t, repo.Complete(ctx, key, winner, testMP3, before.Add(3*time.Second)), "the winning finalizer still owns its lease")
}

func TestPronunciationPostgreSQLSourceAuthorizationAndGlobalBudget(t *testing.T) {
	db := audioDB(t)
	ctx := t.Context()
	repo := &postgresRepository{db}
	u, w, m, otherW, otherM := uuid.New(), uuid.New(), uuid.MustParse("ac93d068-7c3e-55c9-9d85-3fd3518592dc"), uuid.New(), uuid.New()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,'synthetic-audio@synthetic.invalid','active',$2,$2)`, u, now)
	for _, entry := range []struct {
		w, m uuid.UUID
		text string
	}{{w, m, "greeting"}, {otherW, otherM, "synthetic"}} {
		exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,$2,$2,'en','active',$3,$3)`, entry.w, entry.text, now)
		exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun','An authored synthetic definition.',1,'active',$3,$3)`, entry.m, entry.w, now)
		exec(`INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at)VALUES(gen_random_uuid(),$1,$2,1,'active',$3,$3)`, entry.m, "A "+entry.text+" starts this example.", now)
	}
	allowed, err := repo.AuthorizeText(ctx, u, "greeting")
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, err = repo.AuthorizeText(ctx, u, "A greeting starts this example.")
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, err = repo.AuthorizeText(ctx, u, "A synthetic starts this example.")
	require.NoError(t, err)
	require.False(t, allowed, "unreviewed reference examples are not an arbitrary TTS allowlist")
	allowed, err = repo.AuthorizeText(ctx, u, "My private learner sentence.")
	require.NoError(t, err)
	require.False(t, allowed)
	allowed, err = repo.AuthorizeText(ctx, uuid.New(), "greeting")
	require.NoError(t, err)
	require.False(t, allowed)
	exec(`UPDATE users SET status='disabled' WHERE id=$1`, u)
	allowed, err = repo.AuthorizeText(ctx, u, "greeting")
	require.NoError(t, err)
	require.False(t, allowed)

	first, err := repo.Reserve(ctx, cacheKey("first"), 500, now)
	require.NoError(t, err)
	_, err = repo.Reserve(ctx, cacheKey("second"), 500, now)
	require.NoError(t, err)
	_, err = repo.Reserve(ctx, cacheKey("third"), 1, now)
	require.ErrorIs(t, err, ErrRateLimited)
	require.NoError(t, repo.Complete(ctx, cacheKey("first"), first.generation, testMP3, now))
	hit, err := repo.Reserve(ctx, cacheKey("first"), 500, now)
	require.NoError(t, err)
	require.Equal(t, testMP3, hit.audio, "cached replay remains available at generation budget ceiling")
	var spent int
	require.NoError(t, db.QueryRow(`SELECT reserved_characters FROM pronunciation_audio_budget WHERE utc_day=$1`, now.Format(time.DateOnly)).Scan(&spent))
	require.Equal(t, 1000, spent)
	// A new UTC day has its own hard budget. Concurrent API replicas reserve
	// only one paid request for the same text; failed/ambiguous work stays charged.
	day2 := now.Add(24 * time.Hour)
	var wg sync.WaitGroup
	results := make(chan reservation, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := repo.Reserve(ctx, cacheKey("concurrent"), 10, day2)
			results <- r
			errs <- e
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	var reserved reservation
	success, busy := 0, 0
	for r := range results {
		if r.generation != uuid.Nil {
			reserved = r
		}
	}
	for e := range errs {
		if e == nil {
			success++
		} else {
			require.ErrorIs(t, e, ErrBusy)
			busy++
		}
	}
	require.Equal(t, 1, success)
	require.Equal(t, 7, busy)
	require.NoError(t, repo.Complete(ctx, cacheKey("concurrent"), reserved.generation, nil, day2))
	_, err = repo.Reserve(ctx, cacheKey("concurrent"), 10, day2)
	require.ErrorIs(t, err, ErrUnavailable)
	require.NoError(t, db.QueryRow(`SELECT reserved_characters FROM pronunciation_audio_budget WHERE utc_day=$1`, day2.Format(time.DateOnly)).Scan(&spent))
	require.Equal(t, 10, spent)
	next, err := repo.Reserve(ctx, cacheKey("concurrent"), 10, day2.Add(24*time.Hour))
	require.NoError(t, err)
	require.NotEqual(t, reserved.generation, next.generation)
	err = repo.Complete(ctx, cacheKey("concurrent"), reserved.generation, testMP3, day2.Add(24*time.Hour))
	require.ErrorIs(t, err, ErrBusy, "an older finalizer cannot overwrite a new reserved request")
	require.NoError(t, repo.Complete(ctx, cacheKey("concurrent"), next.generation, testMP3, day2.Add(24*time.Hour)))
}
