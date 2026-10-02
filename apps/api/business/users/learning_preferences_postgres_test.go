package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"time"
)

func TestLearningPreferencesPostgreSQLPreservationConcurrencyAndPrivacy(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	exec := func(q string, args ...any) {
		t.Helper()
		_, err := db.ExecContext(ctx, q, args...)
		require.NoError(t, err)
	}
	uid, legacy, other, incomplete, rollback := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	ids := []uuid.UUID{uid, legacy, other, incomplete, rollback}
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	for _, id := range ids {
		exec(`INSERT INTO users(id,email,status,onboarding_status,created_at,updated_at)VALUES($1,$2,'active','completed',$3,$3)`, id, id.String()+"@preferences.invalid", now)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			for _, table := range []string{"user_learning_preferences", "review_attempts", "user_words", "daily_mission_snapshots", "user_onboarding_profiles", "user_settings", "users"} {
				column := "user_id"
				if table == "users" {
					column = "id"
				}
				_, _ = db.Exec("DELETE FROM "+table+" WHERE "+column+"=$1", id)
			}
		}
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewLearningPreferencesService(repo)
	_, _, err = repo.CompleteOnboarding(ctx, uid, OnboardingAnswers{EnglishLevel: "b1", NativeLanguage: "fa", LearningGoal: "general", MainUseCase: "daily_life", DailyReviewTarget: 37, Timezone: "Asia/Tehran"}, now.Add(-24*time.Hour))
	require.NoError(t, err)
	var meaning uuid.UUID
	require.NoError(t, db.QueryRowContext(ctx, `SELECT id FROM word_meanings ORDER BY id LIMIT 1`).Scan(&meaning))
	word := uuid.New()
	exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at)VALUES($1,$2,$3,'learning','manual',$4,$4,$4)`, word, uid, meaning, now)
	exec(`INSERT INTO review_attempts(id,user_id,user_word_id,meaning_id,attempt_type,prompt_type,result,rating,review_step_before,review_step_after,answered_at,source,created_at,updated_at)VALUES($1,$2,$3,$4,'review','self_check','incorrect','again',0,0,$5,'review',$5,$5)`, uuid.New(), uid, word, meaning, now)
	exec(`INSERT INTO daily_mission_snapshots(id,user_id,local_date,timezone,review_target,reviews_completed,policy_version,status,created_at,updated_at)VALUES($1,$2,'2026-10-02','Asia/Tehran',5,1,'v1','open',$3,$3)`, uuid.New(), uid, now)
	accountRepo := accounts.NewPostgreSQLRepository(db)
	export := func(id uuid.UUID) map[string]any {
		t.Helper()
		raw, e := accountRepo.ExportPersonalData(ctx, id)
		require.NoError(t, e)
		var p map[string]any
		require.NoError(t, json.Unmarshal(raw, &p))
		delete(p, "exportedAt")
		return p
	}
	before := export(uid)
	require.Nil(t, before["learningPreferences"])
	p, err := svc.GetLearningPreferences(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, "general", *p.LearningGoal)
	require.Zero(t, p.Revision)
	p, err = svc.GetLearningPreferences(ctx, legacy)
	require.NoError(t, err)
	require.Nil(t, p.LearningGoal)
	require.Nil(t, p.MainUseCase)
	first := LearningPreferencesUpdate{LearningGoal: "travel", MainUseCase: "travel"}
	saved, err := svc.UpdateLearningPreferences(ctx, uid, first)
	require.NoError(t, err)
	require.Equal(t, int64(1), saved.Revision)
	replayed, err := svc.UpdateLearningPreferences(ctx, uid, first)
	require.NoError(t, err)
	require.Equal(t, saved, replayed)
	after := export(uid)
	require.Equal(t, "1.4", after["schemaVersion"])
	require.Equal(t, "travel", after["learningPreferences"].(map[string]any)["learningGoal"])
	delete(before, "learningPreferences")
	delete(after, "learningPreferences")
	require.Equal(t, before, after, "every original exported learning/profile/settings field is unchanged")
	_, err = svc.UpdateLearningPreferences(ctx, legacy, LearningPreferencesUpdate{LearningGoal: "conversation", MainUseCase: "social"})
	require.NoError(t, err)
	var profileCount int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM user_onboarding_profiles WHERE user_id=$1`, legacy).Scan(&profileCount))
	require.Zero(t, profileCount, "do not fabricate grandfathered onboarding")
	otherBefore := export(other)
	require.Nil(t, otherBefore["learningPreferences"])
	results := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, focus := range []string{"work", "study"} {
		wg.Add(1)
		go func(focus string) {
			defer wg.Done()
			<-start
			_, e := svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: focus, MainUseCase: focus, ExpectedRevision: 1})
			results <- e
		}(focus)
	}
	close(start)
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for e := range results {
		if e == nil {
			wins++
		} else if errors.Is(e, ErrLearningPreferencesConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	require.Equal(t, 1, wins)
	require.Equal(t, 1, conflicts)
	_, err = svc.UpdateLearningPreferences(ctx, uid, first)
	require.ErrorIs(t, err, ErrLearningPreferencesConflict)
	current, err := svc.GetLearningPreferences(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, int64(2), current.Revision)
	require.NotEqual(t, "travel", *current.LearningGoal)
	require.Equal(t, otherBefore, export(other), "foreign account remains unchanged")
	exec(`UPDATE users SET onboarding_status='in_progress' WHERE id=$1`, incomplete)
	_, err = svc.UpdateLearningPreferences(ctx, incomplete, first)
	require.ErrorIs(t, err, ErrLearningPreferencesOnboarding)
	_, err = svc.GetLearningPreferences(ctx, incomplete)
	require.ErrorIs(t, err, ErrLearningPreferencesOnboarding)
	// A failing constraint rolls back the row entirely; retry is still revision zero.
	constraint := "preferences_test_" + rollback.String()[:8]
	exec(`ALTER TABLE user_learning_preferences ADD CONSTRAINT ` + constraint + ` CHECK(user_id <> '` + rollback.String() + `')`)
	_, err = svc.UpdateLearningPreferences(ctx, rollback, first)
	require.Error(t, err)
	exec(`ALTER TABLE user_learning_preferences DROP CONSTRAINT ` + constraint)
	p, err = svc.GetLearningPreferences(ctx, rollback)
	require.NoError(t, err)
	require.Zero(t, p.Revision)
	require.Nil(t, p.LearningGoal)
	_, err = svc.UpdateLearningPreferences(ctx, rollback, first)
	require.NoError(t, err)
	// A writer waiting behind deactivation must observe the committed deleted row.
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `UPDATE users SET status='deleted',deleted_at=$2 WHERE id=$1`, uid, now)
	require.NoError(t, err)
	blocked := make(chan error, 1)
	go func() {
		_, e := svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: "exam", MainUseCase: "study", ExpectedRevision: 2})
		blocked <- e
	}()
	select {
	case e := <-blocked:
		t.Fatalf("write escaped account lock: %v", e)
	case <-time.After(40 * time.Millisecond):
	}
	require.NoError(t, tx.Commit())
	require.ErrorIs(t, <-blocked, ErrUserNotFound)
	counters, err := accountRepo.AnonymizeUserData(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, int64(1), counters.LearningPreferences)
	var remaining int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM user_learning_preferences WHERE user_id=$1`, uid).Scan(&remaining))
	require.Zero(t, remaining)
	_, err = svc.UpdateLearningPreferences(ctx, uid, first)
	require.ErrorIs(t, err, ErrUserNotFound)
	p, err = svc.GetLearningPreferences(ctx, legacy)
	require.NoError(t, err)
	require.Equal(t, "social", *p.MainUseCase, "purge is requester-scoped")
}
