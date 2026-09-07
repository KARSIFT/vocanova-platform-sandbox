//go:build integration

package gamification_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// ledgerWriterDB runs all committed forward migrations in an isolated schema.
// Its cleanup closes the pooled database before dropping that schema.
func ledgerWriterDB(t *testing.T) *sql.DB {
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
	t.Cleanup(func() { _ = admin.Close() })
	schema := "ledger_writers_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); require.NoError(t, err) })
	db, err := sql.Open("postgres", dsn+" search_path="+schema+" application_name="+schema)
	require.NoError(t, err)
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(12)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
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
		require.NoError(t, err, "migration %s", path)
	}
	return db
}

func seedWriterMeaning(t *testing.T, db *sql.DB, now time.Time) uuid.UUID {
	t.Helper()
	wordID, meaningID := uuid.New(), uuid.New()
	_, err := db.Exec(`INSERT INTO canonical_words (id,text,normalized_text,status,created_at,updated_at) VALUES ($1,$2,$2,'active',$3,$3)`, wordID, wordID.String(), now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO word_meanings (id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES ($1,$2,'noun','fixture',1,'active',$3,$3)`, meaningID, wordID, now)
	require.NoError(t, err)
	return meaningID
}

func waitForLedgerLock(t *testing.T, ctx context.Context, db *sql.DB, applicationName string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity
			WHERE pid <> pg_backend_pid() AND application_name=$1 AND state='active'
			  AND wait_event_type='Lock' AND query LIKE '%pg_advisory_xact_lock%'`, applicationName).Scan(&n)
		require.NoError(t, err)
		if n > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("second production writer did not wait on pg_advisory_xact_lock")
}

func TestProductionPointWritersSerializeAndRemainReplaySafePostgreSQL(t *testing.T) {
	db := ledgerWriterDB(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	userID := uuid.New()
	// Use three UTC local dates. The finalizer's production MissionUpdater owns
	// its wall clock, so its date is today; the save and review writers use the
	// preceding two days through their injected clocks.
	dayThree := time.Now().UTC().Truncate(time.Microsecond)
	dayOne, dayTwo := dayThree.AddDate(0, 0, -2), dayThree.AddDate(0, 0, -1)
	applicationName := ""
	require.NoError(t, db.QueryRowContext(ctx, "SELECT current_setting('application_name')").Scan(&applicationName))
	_, err := db.ExecContext(ctx, `INSERT INTO users (id,email,status,created_at,updated_at) VALUES ($1,$2,'active',$3,$3)`, userID, userID.String()+"@example.test", dayOne)
	require.NoError(t, err)
	saveMeaningID, reviewMeaningID := seedWriterMeaning(t, db, dayOne), seedWriterMeaning(t, db, dayOne)
	reviewWordID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO user_words (id,user_id,meaning_id,source,added_at,created_at,updated_at) VALUES ($1,$2,$3,'manual',$4,$4,$4)`, reviewWordID, userID, reviewMeaningID, dayOne)
	require.NoError(t, err)
	// A real review completes this pre-existing mission, so it emits both the
	// review and daily-mission source entries from its production transaction.
	_, err = db.ExecContext(ctx, `INSERT INTO daily_mission_snapshots (id,user_id,local_date,timezone,review_target,reviews_completed,policy_version,status,created_at,updated_at) VALUES ($1,$2,$3,'UTC',5,4,$4,'open',$5,$5)`, uuid.New(), userID, dayTwo.Format("2006-01-02"), gamification.MissionPolicyVersion, dayTwo)
	require.NoError(t, err)

	gam := gamification.NewService(gamification.NewRepository(db))
	missionService := missions.NewService(missions.NewRepository(db), gam)
	save := learning.NewService(learning.NewPostgreSQLRepository(db, gam), nil, &clock.Fixed{T: dayOne})
	reviewClock := &clock.Fixed{T: dayTwo}
	review := reviews.NewService(reviews.NewPostgreSQLRepository(db, reviewClock, reviews.WithGamificationService(gam), reviews.WithMissionsService(missionService)), nil, reviewClock)

	// This trigger holds a real first writer after it acquired the advisory
	// lock. pg_stat_activity then verifies the second writer is actually parked
	// on that lock instead of relying on a sleep-based race.
	_, err = db.ExecContext(ctx, `CREATE FUNCTION hold_first_ledger_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(0.5); RETURN NEW; END $$; CREATE TRIGGER hold_first_ledger_insert BEFORE INSERT ON confidence_point_ledger FOR EACH ROW EXECUTE FUNCTION hold_first_ledger_insert()`)
	require.NoError(t, err)
	saved := make(chan error, 1)
	go func() {
		_, err := save.SaveUserWord(ctx, learning.SaveUserWordRequest{UserID: userID, MeaningID: saveMeaningID, Source: "manual", IdempotencyKey: "save-day-one"})
		saved <- err
	}()
	firstWriterSleeping := false
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity
			WHERE pid <> pg_backend_pid() AND application_name=$1 AND state='active'
			  AND wait_event='PgSleep' AND query LIKE '%INSERT INTO confidence_point_ledger%'`, applicationName).Scan(&n)
		require.NoError(t, err)
		if n > 0 {
			firstWriterSleeping = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.True(t, firstWriterSleeping, "first production writer did not reach the controlled ledger insert")
	reviewed := make(chan error, 1)
	go func() {
		_, err := review.SubmitReview(ctx, reviews.SubmitReviewRequest{UserID: userID, UserWordID: reviewWordID, MeaningID: reviewMeaningID, AttemptType: reviews.AttemptTypeReview, PromptType: reviews.PromptTypeSelfCheck, Result: reviews.ResultCorrect, Rating: reviews.RatingGood, AnsweredAt: dayTwo, Source: reviews.SourceReview, ClientAttemptID: "review-day-two", IdempotencyKey: "review-day-two"})
		reviewed <- err
	}()
	waitForLedgerLock(t, ctx, db, applicationName)
	require.NoError(t, <-saved)
	require.NoError(t, <-reviewed)
	_, err = db.ExecContext(ctx, "DROP TRIGGER hold_first_ledger_insert ON confidence_point_ledger; DROP FUNCTION hold_first_ledger_insert()")
	require.NoError(t, err)

	// Exercise the real mission/feedback writer on a third local day. Its first
	// attempt rolls back; retrying identical source IDs must publish exactly two
	// same-clock entries, and a settled replay must add none.
	updater, sentenceID, attemptID := missions.NewMissionUpdater(missionService, gam), uuid.New(), uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO learner_sentences (id,user_id,sentence_text,normalized_sentence_text,source,status,submitted_at,created_at,updated_at) VALUES ($1,$2,'I work.','i work.','free_practice','submitted',$3,$3,$3)`, sentenceID, userID, dayThree)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO ai_feedback_attempts (id,learner_sentence_id,status,provider,model,prompt_version,request_hash,created_at,updated_at) VALUES ($1,$2,'pending','test','test','v1',$3,$4,$4)`, attemptID, sentenceID, attemptID.String(), dayThree)
	require.NoError(t, err)
	pending := aifeedback.PendingAttempt{SentenceID: sentenceID, AttemptID: attemptID}
	feedback := &aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect, Explanation: "fixture", RawJSON: map[string]any{"status": aifeedback.LearningStatusCorrect}}
	feedbackRepo := aifeedback.NewPostgreSQLRepository(db, nil)
	complete := func(ctx context.Context, tx *sql.Tx) (bool, error) {
		return updater.UpdateInTransaction(ctx, tx, userID, sentenceID)
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE confidence_point_ledger ADD CONSTRAINT reject_fixture_sentence CHECK (user_id <> '%s') NOT VALID", userID))
	require.NoError(t, err)
	_, err = feedbackRepo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, dayThree, complete)
	require.Error(t, err)
	var count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM confidence_point_ledger WHERE user_id=$1`, userID).Scan(&count))
	require.Equal(t, 3, count, "rollback must not publish either mission-writer award")
	_, err = db.ExecContext(ctx, "ALTER TABLE confidence_point_ledger DROP CONSTRAINT reject_fixture_sentence")
	require.NoError(t, err)
	_, err = feedbackRepo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, dayThree, complete)
	require.NoError(t, err)
	_, err = feedbackRepo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, dayThree, complete)
	require.NoError(t, err)

	rows, err := db.QueryContext(ctx, `SELECT amount,balance_after,occurred_at FROM confidence_point_ledger WHERE user_id=$1 ORDER BY balance_after`, userID)
	require.NoError(t, err)
	defer rows.Close()
	var amounts, balances []int
	var occurred []time.Time
	for rows.Next() {
		var amount, balance int
		var at time.Time
		require.NoError(t, rows.Scan(&amount, &balance, &at))
		amounts, balances, occurred = append(amounts, amount), append(balances, balance), append(occurred, at)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []int{gamification.RewardAddWord, gamification.RewardReviewGood, gamification.RewardDailyMissionDone, gamification.RewardSentenceSubmitted, gamification.RewardAIFeedbackGot}, amounts)
	require.Equal(t, []int{2, 7, 17, 20, 22}, balances)
	require.Equal(t, dayOne.Format("2006-01-02"), occurred[0].UTC().Format("2006-01-02"))
	require.Equal(t, dayTwo.Format("2006-01-02"), occurred[1].UTC().Format("2006-01-02"))
	require.NotEqual(t, occurred[0].UTC().Format("2006-01-02"), occurred[1].UTC().Format("2006-01-02"), "production writers span distinct local dates")
	require.True(t, occurred[1].Equal(occurred[2]), "review and mission awards use one clock")
	require.True(t, occurred[3].Equal(occurred[4]), "sentence and feedback awards use one clock")
	var sum int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount),0) FROM confidence_point_ledger WHERE user_id=$1`, userID).Scan(&sum))
	displayed, err := gam.CurrentBalance(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, sum, displayed)
	require.Equal(t, 22, displayed)
	var earned, reviewedCount, submitted, feedbackCount int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(confidence_points_earned),0), COALESCE(SUM(reviews_attempted),0),
		COALESCE(SUM(sentences_submitted),0), COALESCE(SUM(ai_feedback_received),0)
		FROM daily_activity_summaries WHERE user_id=$1`, userID).Scan(&earned, &reviewedCount, &submitted, &feedbackCount))
	// Word-add activity remains intentionally disabled with the optional
	// new-word mission goal; the other production writers must account for
	// every point delta they publish (5 + 10 + 3 + 2).
	require.Equal(t, gamification.RewardReviewGood+gamification.RewardDailyMissionDone+gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot, earned)
	require.Equal(t, 1, reviewedCount)
	require.Equal(t, 1, submitted, "settled feedback replay must not repeat sentence activity")
	require.Equal(t, 1, feedbackCount, "settled feedback replay must not repeat feedback activity")
}
