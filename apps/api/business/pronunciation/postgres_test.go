package pronunciation

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
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
	authorize := func(user uuid.UUID, text string) (bool, error) {
		_, err := repo.ResolveUtterance(ctx, user, text, "")
		if err == ErrNotFound {
			return false, nil
		}
		return err == nil, err
	}
	allowed, err := authorize(u, "greeting")
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, err = authorize(u, "A greeting starts this example.")
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, err = authorize(u, "A synthetic starts this example.")
	require.NoError(t, err)
	require.False(t, allowed, "unreviewed reference examples are not an arbitrary TTS allowlist")
	allowed, err = authorize(u, "My private learner sentence.")
	require.NoError(t, err)
	require.False(t, allowed)
	allowed, err = authorize(uuid.New(), "greeting")
	require.NoError(t, err)
	require.False(t, allowed)
	exec(`UPDATE users SET status='disabled' WHERE id=$1`, u)
	allowed, err = authorize(u, "greeting")
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

func TestPronunciationPostgreSQLMeaningResolutionSeparatesPOSAndRejectsMismatches(t *testing.T) {
	db := audioDB(t)
	repo := &postgresRepository{db}
	ctx := t.Context()
	user, word := uuid.New(), uuid.New()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,'meaning-audio@synthetic.invalid','active',$2,$2)`, user, now)
	exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,'close','close','en','active',$2,$2)`, word, now)
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	positions := []string{"noun", "verb", "adjective", "adverb"}
	keys := map[string]bool{}
	for i, pos := range positions {
		definition := "A synthetic " + pos + " meaning."
		exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,$3,$4,$5,'active',$6,$6)`, ids[i], word, pos, definition, i+1, now)
		u, err := repo.ResolveUtterance(ctx, user, "close", ids[i].String())
		require.NoError(t, err)
		require.Equal(t, ids[i].String(), u.MeaningID)
		require.Equal(t, pos, u.PartOfSpeech)
		require.Equal(t, definition, u.Definition)
		require.Equal(t, "en-US", u.Locale)
		require.Equal(t, "close", u.WordText)
		require.Equal(t, "word", u.Kind)
		key := utteranceCacheKey(NewOpenAIProvider(OpenAIConfig{}).Identity(), u)
		require.False(t, keys[key])
		keys[key] = true
		claim, err := repo.Reserve(ctx, key, 5, now)
		require.NoError(t, err)
		tag := make([]byte, 128)
		copy(tag, []byte("TAG"))
		tag[3] = byte(i)
		clip := append(append([]byte{}, testMP3...), tag...)
		require.NoError(t, repo.Complete(ctx, key, claim.generation, clip, now))
		cached, err := repo.Reserve(ctx, key, 5, now)
		require.NoError(t, err)
		require.Equal(t, clip, cached.audio, "different senses retain separate persisted clips")
	}
	for _, request := range []struct{ text, id string }{
		{"close", ""},               // ambiguous legacy spelling
		{"record", ids[1].String()}, // exact sense does not permit unrelated text
		{"close", uuid.NewString()},
		{"My private draft.", ids[0].String()},
	} {
		_, err := repo.ResolveUtterance(ctx, user, request.text, request.id)
		require.ErrorIs(t, err, ErrNotFound)
	}
	exec(`UPDATE word_meanings SET status='archived' WHERE id=$1`, ids[0])
	_, err := repo.ResolveUtterance(ctx, user, "close", ids[0].String())
	require.ErrorIs(t, err, ErrNotFound)
	exec(`UPDATE canonical_words SET status='archived' WHERE id=$1`, word)
	_, err = repo.ResolveUtterance(ctx, user, "close", ids[1].String())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPronunciationPostgreSQLDailyGenerationCapChargesFailuresAndReplaysCache(t *testing.T) {
	db := audioDB(t)
	repo := &postgresRepository{db}
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	var cachedKey string
	for i := 0; i < DailyGenerationLimit; i++ {
		key := cacheKey(fmt.Sprintf("bounded-generation-%d", i))
		claim, err := repo.Reserve(t.Context(), key, 1, now)
		require.NoError(t, err)
		var audio []byte
		if i == 0 {
			audio = testMP3
			cachedKey = key
		}
		require.NoError(t, repo.Complete(t.Context(), key, claim.generation, audio, now))
		if i > 0 {
			_, err = repo.Reserve(t.Context(), key, 1, now)
			require.ErrorIs(t, err, ErrUnavailable, "failure is not retried or refunded")
		}
	}
	_, err := repo.Reserve(t.Context(), cacheKey("twenty-first"), 1, now)
	require.ErrorIs(t, err, ErrRateLimited, "small words cannot bypass the request count ceiling")
	cached, err := repo.Reserve(t.Context(), cachedKey, 1, now)
	require.NoError(t, err)
	require.Equal(t, testMP3, cached.audio)
	var characters, generations int
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT reserved_characters,reserved_generations FROM pronunciation_audio_budget WHERE utc_day=$1`, now.Format(time.DateOnly)).Scan(&characters, &generations))
	require.Equal(t, 20, characters)
	require.Equal(t, 20, generations)
	_, err = repo.Reserve(t.Context(), cacheKey("next-day"), 1, now.Add(24*time.Hour))
	require.NoError(t, err)
}

func TestPronunciationPostgreSQLReviewedSupplementalContextIsExact(t *testing.T) {
	db := audioDB(t)
	repo := &postgresRepository{db}
	ctx := t.Context()
	user := uuid.New()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,'curated-audio@synthetic.invalid','active',$2,$2)`, user, now)
	var entry meaninguses.Entry
	for _, candidate := range meaninguses.Entries() {
		if candidate.Word == "record" && candidate.PartOfSpeech == "noun" {
			entry = candidate
			break
		}
	}
	require.NotEqual(t, uuid.Nil, entry.MeaningID)
	exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,$2,$2,'en','active',$3,$3)`, entry.WordID, entry.Word, now)
	exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,$3,$4,1,'active',$5,$5)`, entry.MeaningID, entry.WordID, entry.PartOfSpeech, entry.Definition, now)
	for i, example := range []string{"An arbitrary database example must not replace reviewed context.", entry.Example} {
		exec(`INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at)VALUES(gen_random_uuid(),$1,$2,$3,'active',$4,$4)`, entry.MeaningID, example, i+1, now)
	}
	word, err := repo.ResolveUtterance(ctx, user, entry.Word, entry.MeaningID.String())
	require.NoError(t, err)
	require.Equal(t, entry.Example, word.Context, "curated context overrides the first database example")
	example, err := repo.ResolveUtterance(ctx, user, entry.Example, entry.MeaningID.String())
	require.NoError(t, err)
	require.Equal(t, "example", example.Kind)
	require.Equal(t, entry.Word, example.WordText)
	require.Equal(t, entry.Example, example.Context)
	_, err = repo.ResolveUtterance(ctx, user, "An arbitrary database example must not replace reviewed context.", entry.MeaningID.String())
	require.ErrorIs(t, err, ErrNotFound)
	exec(`UPDATE word_meanings SET short_definition='A modified unreviewed definition.' WHERE id=$1`, entry.MeaningID)
	_, err = repo.ResolveUtterance(ctx, user, entry.Example, entry.MeaningID.String())
	require.ErrorIs(t, err, ErrNotFound, "known meaning ID alone does not bless a modified definition/example")
}

func TestPronunciationPostgreSQLResumeAliasKeepsCanonicalRevision(t *testing.T) {
	db := audioDB(t)
	repo := &postgresRepository{db}
	ctx := t.Context()
	user, word, noun, verb := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,'alias-audio@synthetic.invalid','active',$2,$2)`, user, now)
	exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,'resume','resume','en','active',$2,$2)`, word, now)
	exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$3,'noun','A summary of work experience.',1,'active',$4,$4),($2,$3,'verb','To begin again after a pause.',2,'active',$4,$4)`, noun, verb, word, now)
	canonical, err := repo.ResolveUtterance(ctx, user, "resume", noun.String())
	require.NoError(t, err)
	alias, err := repo.ResolveUtterance(ctx, user, "résumé", noun.String())
	require.NoError(t, err)
	require.Equal(t, "resume", alias.WordText)
	require.Equal(t, "résumé", alias.Text)
	require.Equal(t, "résumé", canonical.Text)
	identity := NewOpenAIProvider(OpenAIConfig{}).Identity()
	canonicalKey := utteranceCacheKey(identity, canonical)
	require.Equal(t, canonicalKey, utteranceCacheKey(identity, alias), "noun aliases share one paid cache identity")
	reserved, err := repo.Reserve(ctx, canonicalKey, len(canonical.Text), now)
	require.NoError(t, err)
	require.NoError(t, repo.Complete(ctx, canonicalKey, reserved.generation, testMP3, now))
	replay, err := repo.Reserve(ctx, utteranceCacheKey(identity, alias), len(alias.Text), now)
	require.NoError(t, err)
	require.Equal(t, testMP3, replay.audio)
	var generations int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT reserved_generations FROM pronunciation_audio_budget WHERE utc_day=$1`, now.Format(time.DateOnly)).Scan(&generations))
	require.Equal(t, 1, generations, "replaying the alias must not reserve another generation")
	require.Equal(t, meaningidentity.Revision(canonical.MeaningID, canonical.WordText, canonical.PartOfSpeech, canonical.Definition), meaningidentity.Revision(alias.MeaningID, alias.WordText, alias.PartOfSpeech, alias.Definition))
	_, err = repo.ResolveUtterance(ctx, user, "résumé", verb.String())
	require.ErrorIs(t, err, ErrNotFound, "accented noun spelling cannot address the verb sense")
}

func TestPronunciationPostgreSQLExistingReviewedLessonExamplesStayMeaningBound(t *testing.T) {
	db := audioDB(t)
	repo := &postgresRepository{db}
	ctx := t.Context()
	user := uuid.New()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,'existing-lesson-audio@synthetic.invalid','active',$2,$2)`, user, now)
	for _, entry := range []struct{ id, wordID, text, definition, example string }{
		{"446b7b9e-8190-5987-9401-0378e95ff493", "19e74673-e746-540e-8a33-20724c758443", "follow-up", "A later message or action.", "The manager scheduled a follow-up for Monday."},
		{"f582a137-c77d-58df-942e-884fe336cd64", "fb5affef-ca90-5f39-b5b1-83a8569d548a", "deadline", "When work must be submitted.", "The essay deadline is at noon."},
	} {
		wrongUse := uuid.New()
		exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,$2,$2,'en','active',$3,$3)`, entry.wordID, entry.text, now)
		exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'noun',$3,2,'active',$4,$4),($5,$2,'noun','An unreviewed distinct meaning.',3,'active',$4,$4)`, entry.id, entry.wordID, entry.definition, now, wrongUse)
		unreviewed := "An unreviewed " + entry.text + " example from the database."
		for i, row := range []struct{ id, text string }{{entry.id, entry.example}, {entry.id, unreviewed}, {wrongUse.String(), entry.example}} {
			exec(`INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at)VALUES(gen_random_uuid(),$1,$2,$4,'active',$3,$3)`, row.id, row.text, now, i+1)
		}
		actual, err := repo.ResolveUtterance(ctx, user, entry.example, entry.id)
		require.NoError(t, err)
		require.Equal(t, "example", actual.Kind)
		require.Equal(t, entry.id, actual.MeaningID)
		require.Equal(t, entry.text, actual.WordText)
		require.Equal(t, "noun", actual.PartOfSpeech)
		require.Equal(t, entry.example, actual.Text)
		require.Equal(t, entry.example, actual.Context)
		_, err = repo.ResolveUtterance(ctx, user, entry.example, wrongUse.String())
		require.ErrorIs(t, err, ErrNotFound, "reviewed text cannot authorize the wrong use, even if a database example repeats it")
		_, err = repo.ResolveUtterance(ctx, user, unreviewed, entry.id)
		require.ErrorIs(t, err, ErrNotFound, "an exact reviewed meaning does not authorize arbitrary database examples")
		exec(`UPDATE word_meanings SET short_definition='A changed unreviewed definition.' WHERE id=$1`, entry.id)
		_, err = repo.ResolveUtterance(ctx, user, entry.example, entry.id)
		require.ErrorIs(t, err, ErrNotFound, "reviewed context is bound to its exact definition")
	}
}

func TestPronunciationPostgreSQLSilentPaidGenerationStaysChargedAndDoesNotRetry(t *testing.T) {
	db := audioDB(t)
	ctx := t.Context()
	user, word, meaning := uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	for _, statement := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO users(id,email,status,created_at,updated_at)VALUES($1,'silent-provider@synthetic.invalid','active',$2,$2)`, []any{user, now}},
		{`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at)VALUES($1,'synthetic','synthetic','en','active',$2,$2)`, []any{word, now}},
		{`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at)VALUES($1,$2,'adjective','An original synthetic meaning.',1,'active',$3,$3)`, []any{meaning, word, now}},
	} {
		_, err := db.ExecContext(ctx, statement.sql, statement.args...)
		require.NoError(t, err)
	}

	provider := &fakeProvider{fn: func(context.Context, Utterance) ([]byte, error) { return silentMP3, nil }}
	svc := NewService(db, provider, true, clock.Fixed{T: now})
	for i := 0; i < 2; i++ {
		_, err := svc.SpeakMeaning(ctx, user, "synthetic", meaning.String())
		require.ErrorIs(t, err, ErrUnavailable)
	}
	require.Equal(t, 1, provider.calls, "silent successful HTTP output must not be paid for twice")
	var status string
	var data []byte
	var characters, generations int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT status,audio FROM pronunciation_audio_cache`).Scan(&status, &data))
	require.Equal(t, "failed", status)
	require.Empty(t, data)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT reserved_characters,reserved_generations FROM pronunciation_audio_budget`).Scan(&characters, &generations))
	require.Equal(t, len("synthetic"), characters)
	require.Equal(t, 1, generations)
}
