package lessons

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Run against a disposable fully migrated database with the canonical seed.
func TestPostgreSQLLessonRecommendationCoverageAndPrivacy(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx := t.Context()
	u, other := uuid.New(), uuid.New()
	exec := func(q string, args ...any) {
		t.Helper()
		_, e := db.ExecContext(ctx, q, args...)
		require.NoError(t, e)
	}
	for _, id := range []uuid.UUID{u, other} {
		exec(`INSERT INTO users(id,email,status,onboarding_status,created_at,updated_at) VALUES($1,$2,'active','completed',NOW(),NOW())`, id, id.String()+"@recommendation.invalid")
	}
	t.Cleanup(func() {
		for _, id := range []uuid.UUID{u, other} {
			for _, table := range []string{"lesson_actions", "lesson_sessions", "user_learning_preferences", "user_word_knowledge", "user_words", "user_onboarding_profiles", "user_settings"} {
				_, _ = db.ExecContext(context.Background(), "DELETE FROM "+table+" WHERE user_id=$1", id)
			}
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id=$1", id)
		}
	})
	repo := NewPostgreSQLRepository(db)
	svc := NewRecommendationService(repo)
	get := func() RecommendationResult { t.Helper(); r, e := svc.Get(ctx, u); require.NoError(t, e); return r }
	got := get()
	require.Equal(t, catalog[0].Key, got.Recommendation.Lesson.Key, "grandfathered absent preferences use stable catalog order")
	// Current focus overrides the original onboarding focus without changing it.
	_, _, err = users.NewPostgreSQLRepository(db).CompleteOnboarding(ctx, u, users.OnboardingAnswers{EnglishLevel: "a2", NativeLanguage: "en", LearningGoal: "travel", MainUseCase: "travel", DailyReviewTarget: 10, Timezone: "UTC"}, time.Now())
	require.NoError(t, err)
	data, err := repo.ReadRecommendation(ctx, u)
	require.NoError(t, err)
	require.Equal(t, "travel", data.Focus)
	exec(`INSERT INTO user_learning_preferences(id,user_id,learning_goal,main_use_case,revision,created_at,updated_at) VALUES($1,$2,'conversation','social',1,NOW(),NOW())`, uuid.New(), u)
	data, err = repo.ReadRecommendation(ctx, u)
	require.NoError(t, err)
	require.Equal(t, "social", data.Focus)
	require.Len(t, data.Available, len(catalog), "all current lessons can start against the canonical seed")
	exec(`DELETE FROM user_learning_preferences WHERE user_id=$1`, u)
	exec(`DELETE FROM user_onboarding_profiles WHERE user_id=$1`, u)
	// More than fifty persisted assessments: only the final catalog lesson is
	// useful, regardless of any public list page size or ordering.
	last := catalog[len(catalog)-1]
	for _, lesson := range catalog {
		for _, w := range lesson.Words {
			exec(`INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note) VALUES($1,$2,true,'private synthetic note') ON CONFLICT(user_id,meaning_id) DO UPDATE SET self_reported_known=true`, u, w.MeaningID)
		}
	}
	var count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM user_word_knowledge WHERE user_id=$1`, u).Scan(&count))
	require.Greater(t, count, 50)
	require.Equal(t, "no_useful_targets", get().Status)
	for i, w := range last.Words {
		exec(`UPDATE user_word_knowledge SET self_reported_known=false WHERE user_id=$1 AND meaning_id=$2`, u, w.MeaningID)
		exec(`INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known) VALUES($1,$2,true)`, other, w.MeaningID)
		exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at) VALUES($1,$2,$3,$4,'manual',NOW(),NOW(),NOW())`, uuid.New(), u, w.MeaningID, []string{"new", "learning", "reviewing"}[i])
	}
	got = get()
	require.Equal(t, last.Key, got.Recommendation.Lesson.Key)
	require.Equal(t, 3, got.Recommendation.UsefulTargetCount, "notes, foreign knowledge and merely saved stages do not exclude targets")
	exec(`UPDATE user_words SET status='mastered' WHERE user_id=$1 AND meaning_id=$2`, u, last.Words[0].MeaningID)
	require.Equal(t, 2, get().Recommendation.UsefulTargetCount)
	exec(`UPDATE user_words SET deleted_at=NOW() WHERE user_id=$1 AND meaning_id=$2`, u, last.Words[0].MeaningID)
	require.Equal(t, 3, get().Recommendation.UsefulTargetCount, "removed mastery is not an active saved stage")
	// Missing current canonical content must not become a false zero-coverage result.
	meaning := last.Words[0].MeaningID
	exec(`UPDATE word_meanings SET status='archived' WHERE id=$1`, meaning)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `UPDATE word_meanings SET status='active' WHERE id=$1`, meaning)
	})
	got = get()
	require.Equal(t, "content_unavailable", got.Status)
	require.Nil(t, got.Recommendation)
	exec(`UPDATE word_meanings SET status='active' WHERE id=$1`, meaning)
	// Resume the most recently touched owned snapshot, including known targets.
	lessonSvc := NewService(repo, nil)
	first, err := lessonSvc.Start(ctx, u, catalog[0].Key, "recommendation-first")
	require.NoError(t, err)
	second, err := lessonSvc.Start(ctx, u, last.Key, "recommendation-last")
	require.NoError(t, err)
	_, err = lessonSvc.Start(ctx, other, "airport", "foreign-session")
	require.NoError(t, err)
	exec(`UPDATE lesson_sessions SET updated_at=NOW()-INTERVAL '1 hour' WHERE id=$1`, second.ID)
	exec(`UPDATE lesson_sessions SET updated_at=NOW() WHERE id=$1`, first.ID)
	got = get()
	require.Equal(t, first.ID, got.Recommendation.Lesson.SessionID)
	require.Equal(t, "resume", got.Recommendation.Reason)
	require.Zero(t, got.Recommendation.UsefulTargetCount)
	// A completion is never fabricated by a recommendation. Once explicitly
	// persisted by this fixture, it is excluded and the other session resumes.
	exec(`UPDATE lesson_sessions SET current_step=total_steps,status='completed',completed_at=NOW() WHERE id=$1`, first.ID)
	require.Equal(t, second.ID, get().Recommendation.Lesson.SessionID)
	var before, after string
	const stateSQL = `SELECT jsonb_build_object('sessions',(SELECT jsonb_agg(to_jsonb(s) ORDER BY id) FROM lesson_sessions s WHERE user_id=$1),'words',(SELECT jsonb_agg(to_jsonb(w) ORDER BY id) FROM user_words w WHERE user_id=$1),'knowledge',(SELECT jsonb_agg(to_jsonb(k) ORDER BY id) FROM user_word_knowledge k WHERE user_id=$1),'actions',(SELECT count(*) FROM lesson_actions WHERE user_id=$1),'missions',(SELECT count(*) FROM daily_mission_snapshots WHERE user_id=$1),'points',(SELECT count(*) FROM confidence_point_ledger WHERE user_id=$1))::text`
	require.NoError(t, db.QueryRowContext(ctx, stateSQL, u).Scan(&before))
	_ = get()
	_ = get()
	require.NoError(t, db.QueryRowContext(ctx, stateSQL, u).Scan(&after))
	require.Equal(t, before, after, "recommendation GET cannot advance, award or change private state")
	exec(`UPDATE users SET status='disabled' WHERE id=$1`, u)
	_, err = svc.Get(ctx, u)
	require.ErrorIs(t, err, ErrNotFound)
	exec(`UPDATE users SET status='active',deleted_at=NOW() WHERE id=$1`, u)
	_, err = svc.Get(ctx, u)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = svc.Get(ctx, uuid.New())
	require.ErrorIs(t, err, ErrNotFound)
	foreign, err := svc.Get(ctx, other)
	require.NoError(t, err)
	require.Equal(t, "airport", foreign.Recommendation.Lesson.Key)
}
