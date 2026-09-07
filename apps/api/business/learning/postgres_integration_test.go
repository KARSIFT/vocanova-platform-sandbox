package learning

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// isolatedPostgres opens a database whose connections all use a fresh schema.
// Unlike temporary tables, the schema is visible to every pooled connection, so
// integration tests can exercise real overlapping transactions without touching
// the shared database's public schema.
func isolatedPostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "learning_integration_" + uuid.NewString()[0:8]
	if _, err := admin.Exec(`CREATE SCHEMA ` + schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
		admin.Close()
	})

	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	t.Cleanup(func() { db.Close() })
	return db
}

// TestSaveUserWordRejectsInactiveCanonicalWordPostgreSQL verifies the save
// eligibility predicate against a real PostgreSQL engine in an isolated schema.
func TestSaveUserWordRejectsInactiveCanonicalWordPostgreSQL(t *testing.T) {
	db := isolatedPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE canonical_words (id uuid PRIMARY KEY, status text NOT NULL);
		CREATE TABLE word_meanings (
			id uuid PRIMARY KEY,
			word_id uuid NOT NULL REFERENCES canonical_words(id),
			status text NOT NULL
		);
		CREATE TABLE user_words (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			meaning_id uuid NOT NULL,
			status text NOT NULL,
			source text NOT NULL,
			review_step integer NOT NULL,
			next_review_at timestamptz,
			last_reviewed_at timestamptz,
			last_result text,
			last_rating text,
			consecutive_correct_count integer NOT NULL,
			consecutive_incorrect_count integer NOT NULL,
			total_review_count integer NOT NULL,
			correct_review_count integer NOT NULL,
			added_at timestamptz NOT NULL,
			mastered_at timestamptz,
			ignored_at timestamptz,
			deleted_at timestamptz,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL
		);`); err != nil {
		t.Fatal(err)
	}

	wordID, meaningID := uuid.New(), uuid.New()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO canonical_words (id, status) VALUES ($1, 'archived')`, wordID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO word_meanings (id, word_id, status) VALUES ($1, $2, 'active')`, meaningID, wordID,
	); err != nil {
		t.Fatal(err)
	}

	_, err := NewPostgreSQLRepository(db).SaveUserWord(ctx, SaveUserWordRequest{
		UserID:    uuid.New(),
		MeaningID: meaningID,
		Source:    "journey",
	}, time.Now().UTC())
	if !errors.Is(err, ErrMeaningNotFound) {
		t.Fatalf("SaveUserWord() error = %v, want ErrMeaningNotFound", err)
	}

	var savedCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_words`).Scan(&savedCount); err != nil {
		t.Fatal(err)
	}
	if savedCount != 0 {
		t.Fatalf("saved rows = %d, want 0", savedCount)
	}
}

// TestSavedWordStatesPostgreSQLUsesPersistedStatesAndDatabaseTime proves the
// Word Detail projection against PostgreSQL itself. It deliberately uses the
// database clock instead of a Go clock so the due result cannot drift with a
// requester or application-server clock.
func TestSavedWordStatesPostgreSQLUsesPersistedStatesAndDatabaseTime(t *testing.T) {
	db := isolatedPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE user_words (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			meaning_id uuid NOT NULL,
			status text NOT NULL,
			total_review_count integer NOT NULL DEFAULT 0,
			next_review_at timestamptz,
			deleted_at timestamptz
		);`); err != nil {
		t.Fatal(err)
	}

	requester, otherUser := uuid.New(), uuid.New()
	type savedWord struct {
		meaningID uuid.UUID
		status    string
		reviews   int
		nextDue   interface{}
		deleted   bool
		userID    uuid.UUID
	}
	words := []savedWord{
		{uuid.New(), "new", 0, nil, false, requester},
		{uuid.New(), "new", 1, "CURRENT_TIMESTAMP + INTERVAL '1 minute'", false, requester},
		{uuid.New(), "learning", 0, "CURRENT_TIMESTAMP - INTERVAL '1 minute'", false, requester},
		{uuid.New(), "reviewing", 0, "CURRENT_TIMESTAMP + INTERVAL '1 minute'", false, requester},
		{uuid.New(), "mastered", 0, "CURRENT_TIMESTAMP - INTERVAL '1 minute'", false, requester},
		{uuid.New(), "ignored", 0, "CURRENT_TIMESTAMP - INTERVAL '1 minute'", false, requester},
		{uuid.New(), "archived", 0, "CURRENT_TIMESTAMP - INTERVAL '1 minute'", false, requester},
		{uuid.New(), "learning", 0, "CURRENT_TIMESTAMP - INTERVAL '1 minute'", true, requester},
		{uuid.New(), "learning", 0, "CURRENT_TIMESTAMP - INTERVAL '1 minute'", false, otherUser},
	}
	for _, word := range words {
		query := `INSERT INTO user_words (id, user_id, meaning_id, status, total_review_count, next_review_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, `
		if word.nextDue == nil {
			query += `NULL`
		} else {
			query += word.nextDue.(string)
		}
		if word.deleted {
			query += `, CURRENT_TIMESTAMP)`
		} else {
			query += `, NULL)`
		}
		if _, err := db.ExecContext(ctx, query, uuid.New(), word.userID, word.meaningID, word.status, word.reviews); err != nil {
			t.Fatal(err)
		}
	}

	meaningIDs := make([]uuid.UUID, len(words))
	for i, word := range words {
		meaningIDs[i] = word.meaningID
	}
	states, err := NewPostgreSQLRepository(db).SavedWordStates(ctx, requester, meaningIDs)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		index int
		due   bool
	}{
		{0, true},  // new / NULL
		{1, false}, // reviewed legacy new / future
		{2, true},  // learning / past
		{3, false}, // reviewing / future
		{4, false}, // mastered is never due
		{5, false}, // ignored is never due
		{6, false}, // archived is never due
	} {
		state, ok := states[words[tc.index].meaningID]
		if !ok {
			t.Fatalf("state for %s missing", words[tc.index].status)
		}
		wantStatus := words[tc.index].status
		if words[tc.index].status == "new" && words[tc.index].reviews > 0 {
			wantStatus = "learning"
		}
		if state.Status != wantStatus || state.Due != tc.due {
			t.Errorf("%s state = %#v, want status %q and due %t", words[tc.index].status, state, wantStatus, tc.due)
		}
	}
	for _, word := range words {
		if !word.deleted && word.userID == requester {
			continue
		}
		if _, ok := states[word.meaningID]; ok {
			t.Errorf("inaccessible saved word was returned: deleted=%t requester=%s", word.deleted, word.userID)
		}
	}
}

// TestSaveUserWordAtomicIdempotencyPostgreSQL exercises the production
// transaction path against PostgreSQL. Temporary tables and one connection keep
// it isolated from application data while still using the real unique-index and
// rollback semantics.
func TestSaveUserWordAtomicIdempotencyPostgreSQL(t *testing.T) {
	db := isolatedPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, `
		CREATE TABLE canonical_words (id uuid PRIMARY KEY, text text NOT NULL DEFAULT 'word', normalized_text text NOT NULL DEFAULT 'word', status text NOT NULL);
		CREATE TABLE word_meanings (id uuid PRIMARY KEY, word_id uuid NOT NULL REFERENCES canonical_words(id), part_of_speech text NOT NULL DEFAULT 'noun', short_definition text NOT NULL DEFAULT 'definition', status text NOT NULL);
		CREATE TABLE user_words (id uuid PRIMARY KEY, user_id uuid NOT NULL, meaning_id uuid NOT NULL, status text NOT NULL, source text NOT NULL, review_step integer NOT NULL, next_review_at timestamptz, last_reviewed_at timestamptz, last_result text, last_rating text, consecutive_correct_count integer NOT NULL, consecutive_incorrect_count integer NOT NULL, total_review_count integer NOT NULL, correct_review_count integer NOT NULL, added_at timestamptz NOT NULL, mastered_at timestamptz, ignored_at timestamptz, deleted_at timestamptz, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL);
		CREATE TABLE idempotency_keys (id uuid PRIMARY KEY, user_id uuid NOT NULL, operation text NOT NULL, key text NOT NULL, fingerprint text NOT NULL, created_at timestamptz NOT NULL, UNIQUE(user_id, operation, key));
		CREATE TABLE confidence_point_ledger (id uuid PRIMARY KEY, user_id uuid NOT NULL, amount integer NOT NULL CHECK (amount <> 0), balance_after integer NOT NULL, reason text NOT NULL, source_type text NOT NULL, source_id uuid, idempotency_key text, metadata jsonb, occurred_at timestamptz NOT NULL, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL, UNIQUE(user_id, idempotency_key));
		CREATE FUNCTION hold_idempotency_claim() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(0.15); RETURN NEW; END $$;
		CREATE TRIGGER hold_idempotency_claim BEFORE INSERT ON idempotency_keys FOR EACH ROW EXECUTE FUNCTION hold_idempotency_claim();`)
	if err != nil {
		t.Fatal(err)
	}
	wordID, firstMeaning, secondMeaning, failingMeaning, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.ExecContext(ctx, `INSERT INTO canonical_words (id, status) VALUES ($1, 'active')`, wordID); err != nil {
		t.Fatal(err)
	}
	for _, meaningID := range []uuid.UUID{firstMeaning, secondMeaning, failingMeaning} {
		if _, err := db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, status) VALUES ($1, $2, 'active')`, meaningID, wordID); err != nil {
			t.Fatal(err)
		}
	}
	fixedClock := &clock.Fixed{T: time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)}
	svc := NewService(NewPostgreSQLRepository(db, gamification.NewService(gamification.NewRepository(db))), nil, fixedClock)

	// The Go gate starts both requests together; the PostgreSQL trigger keeps
	// their claim statements overlapping on separate pooled connections.
	var wg sync.WaitGroup
	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	type result struct {
		meaningID uuid.UUID
		err       error
	}
	results := make(chan result, 2)
	for _, meaningID := range []uuid.UUID{firstMeaning, secondMeaning} {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			ready <- struct{}{}
			<-start
			_, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: id, Source: "journey", IdempotencyKey: "one"})
			results <- result{meaningID: id, err: err}
		}(meaningID)
	}
	<-ready
	<-ready
	close(start)
	wg.Wait()
	close(results)
	var success, conflict int
	var winner uuid.UUID
	for result := range results {
		if result.err == nil {
			success++
			winner = result.meaningID
		} else if errors.Is(result.err, ErrIdempotencyConflict) {
			conflict++
		} else {
			t.Fatalf("save: %v", result.err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success/conflict = %d/%d, want 1/1", success, conflict)
	}
	var words int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM user_words`).Scan(&words); err != nil || words != 1 {
		t.Fatalf("words = %d, err = %v; want 1", words, err)
	}
	if _, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: winner, Source: "journey", IdempotencyKey: "one"}); err != nil {
		t.Fatalf("matching replay: %v", err)
	}
	var ledger int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM confidence_point_ledger WHERE user_id = $1`, userID).Scan(&ledger); err != nil || ledger != 1 {
		t.Fatalf("replay ledger rows = %d, err = %v; want 1", ledger, err)
	}

	// The exact 24-hour boundary expires the claim, permitting a fresh meaning.
	fixedClock.T = fixedClock.T.Add(idempotencyRetention)
	if _, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: secondMeaning, Source: "journey", IdempotencyKey: "one"}); err != nil {
		t.Fatalf("expired key reuse: %v", err)
	}

	// This is an actual database reward failure after the user-word INSERT.
	// The transaction must remove the word and idempotency claim along with the
	// failed ledger write.
	failingUser := uuid.New()
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE confidence_point_ledger ADD CONSTRAINT reject_failing_user CHECK (user_id <> '%s'::uuid)`, failingUser)); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: failingUser, MeaningID: failingMeaning, Source: "journey", IdempotencyKey: "reward-fails"}); err == nil {
		t.Fatal("save with rejected reward succeeded")
	}
	for table, want := range map[string]int{"user_words": 0, "idempotency_keys": 0, "confidence_point_ledger": 0} {
		var count int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, failingUser).Scan(&count); err != nil || count != want {
			t.Fatalf("rollback %s rows = %d, err = %v; want %d", table, count, err, want)
		}
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE confidence_point_ledger DROP CONSTRAINT reject_failing_user`); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: failingUser, MeaningID: failingMeaning, Source: "journey", IdempotencyKey: "reward-fails"}); err != nil {
		t.Fatalf("retry after rolled-back reward: %v", err)
	}
}

// migratedWordActivityDB applies every committed forward migration in a fresh
// schema so this test exercises the same constraints and defaults as the live
// P1/P4 write path without sharing any application data.
func migratedWordActivityDB(t *testing.T) *sql.DB {
	t.Helper()
	db := isolatedPostgres(t)
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		if strings.HasSuffix(path, ".down.sql") {
			continue
		}
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoErrorf(t, err, "apply migration %s", filepath.Base(path))
	}
	return db
}

// TestSaveUserWordRecordsDailyActivityPostgreSQL proves the production P1
// first-save transaction keeps the exported daily aggregate in lockstep with
// its +2 immutable ledger award. It also covers the paths that must remain
// no-ops (duplicate, idempotent replay, and restore) and rollback on an
// activity-write failure.
func TestSaveUserWordRecordsDailyActivityPostgreSQL(t *testing.T) {
	db := migratedWordActivityDB(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	userID, meaningID, failingMeaningID := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, func() error {
		_, err := db.ExecContext(ctx,
			`INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
			 VALUES ($1, $2, 'active', 'completed', $3, $3)`, userID, userID.String()+"@example.test", now)
		return err
	}())
	for _, id := range []uuid.UUID{meaningID, failingMeaningID} {
		wordID := uuid.New()
		_, err := db.ExecContext(ctx,
			`INSERT INTO canonical_words (id, text, normalized_text, status, created_at, updated_at)
			 VALUES ($1, $2, $2, 'active', $3, $3)`, wordID, wordID.String(), now)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx,
			`INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at)
			 VALUES ($1, $2, 'noun', 'fixture', 1, 'active', $3, $3)`, id, wordID, now)
		require.NoError(t, err)
	}

	gam := gamification.NewService(gamification.NewRepository(db))
	missionSvc := missions.NewService(missions.NewRepository(db), gam)
	svc := NewService(NewPostgreSQLRepository(db, gam, missionSvc), nil, &clock.Fixed{T: now})
	save := SaveUserWordRequest{UserID: userID, MeaningID: meaningID, Source: "manual", IdempotencyKey: "first-save"}
	_, err := svc.SaveUserWord(ctx, save)
	require.NoError(t, err)

	localDate := now.Format("2006-01-02")
	var wordsAdded, pointsEarned, ledgerRows int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT words_added, confidence_points_earned FROM daily_activity_summaries WHERE user_id = $1 AND local_date = $2`, userID, localDate,
	).Scan(&wordsAdded, &pointsEarned))
	require.Equal(t, 1, wordsAdded)
	require.Equal(t, gamification.RewardAddWord, pointsEarned)
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM confidence_point_ledger WHERE user_id = $1`, userID,
	).Scan(&ledgerRows))
	require.Equal(t, 1, ledgerRows)

	// A new request key for an already active word, and an exact replay, are
	// both no-op saves rather than additional activity or rewards.
	_, err = svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: meaningID, Source: "manual", IdempotencyKey: "duplicate"})
	require.NoError(t, err)
	_, err = svc.SaveUserWord(ctx, save)
	require.NoError(t, err)
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT words_added, confidence_points_earned FROM daily_activity_summaries WHERE user_id = $1 AND local_date = $2`, userID, localDate,
	).Scan(&wordsAdded, &pointsEarned))
	require.Equal(t, 1, wordsAdded)
	require.Equal(t, gamification.RewardAddWord, pointsEarned)

	// Restoring an unsaved word retains the existing anti-farming policy.
	require.NoError(t, svc.UnsaveUserWord(ctx, userID, meaningID))
	_, err = svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: meaningID, Source: "manual", IdempotencyKey: "restore"})
	require.NoError(t, err)
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT words_added, confidence_points_earned FROM daily_activity_summaries WHERE user_id = $1 AND local_date = $2`, userID, localDate,
	).Scan(&wordsAdded, &pointsEarned))
	require.Equal(t, 1, wordsAdded)
	require.Equal(t, gamification.RewardAddWord, pointsEarned)

	// The activity write occurs after the ledger insert but before commit; a
	// database failure must leave none of the first-save state behind.
	failingUserID := uuid.New()
	require.NoError(t, func() error {
		_, err := db.ExecContext(ctx,
			`INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
			 VALUES ($1, $2, 'active', 'completed', $3, $3)`, failingUserID, failingUserID.String()+"@example.test", now)
		return err
	}())
	require.NoError(t, func() error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(
			`ALTER TABLE daily_activity_summaries ADD CONSTRAINT reject_word_activity CHECK (user_id <> '%s'::uuid)`, failingUserID))
		return err
	}())
	_, err = svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: failingUserID, MeaningID: failingMeaningID, Source: "manual", IdempotencyKey: "activity-fails"})
	require.Error(t, err)
	for _, table := range []string{"user_words", "idempotency_keys", "confidence_point_ledger", "daily_mission_snapshots", "daily_activity_summaries"} {
		var count int
		require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, failingUserID).Scan(&count))
		require.Zerof(t, count, "rollback left %s rows", table)
	}
}
