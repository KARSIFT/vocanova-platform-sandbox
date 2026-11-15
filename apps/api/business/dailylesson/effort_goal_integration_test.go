package dailylesson

import (
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDailyGoalUsesTargetAndWrongAnswerEffort(t *testing.T) {
	f := setup(t, 18)
	_, err := f.db.Exec("UPDATE user_settings SET daily_review_target=15 WHERE user_id=$1", f.u)
	require.NoError(t, err)
	s, err := f.svc.Start(t.Context(), f.u, "goal15")
	require.NoError(t, err)
	require.Equal(t, 15, s.TotalWords, "new sessions use today's remaining goal")
	s = f.answer(t, s, false)
	require.False(t, s.Feedback.Correct)
	require.False(t, s.Feedback.Credited, "success credit must remain truthful")
	var completed, successes, points int
	require.NoError(t, f.db.QueryRow("SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1", f.u).Scan(&completed))
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_review_successes WHERE user_id=$1", f.u).Scan(&successes))
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&points))
	require.Equal(t, 1, completed)
	require.Zero(t, successes)
	require.Zero(t, points)
}
func TestDailyWrongAnswerEarnsEffortWithoutSuccess(t *testing.T) {
	f := setup(t, 1)
	s, err := f.svc.Start(t.Context(), f.u, "wrong-effort")
	require.NoError(t, err)
	s = f.answer(t, s, false)
	require.False(t, s.Feedback.Credited)
	var completed int
	require.NoError(t, f.db.QueryRow("SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1", f.u).Scan(&completed))
	require.Equal(t, 1, completed, "a committed wrong answer counts toward review effort")
}

func TestDailyGoalCompletionAdditionalAndFrozenTarget(t *testing.T) {
	f := setup(t, 9)
	s, err := f.svc.Start(t.Context(), f.u, "required")
	require.NoError(t, err)
	require.Equal(t, 5, s.Goal.Target)
	_, err = f.db.Exec("UPDATE user_settings SET daily_review_target=15 WHERE user_id=$1", f.u)
	require.NoError(t, err)
	for s.Status != "completed" {
		s = f.answer(t, s, false)
		s = f.next(t, s)
	}
	require.True(t, s.Goal.Complete)
	require.Equal(t, 5, s.Goal.Reviewed)
	require.Zero(t, s.CreditedToday)
	current, err := f.svc.Current(t.Context(), f.u)
	require.NoError(t, err)
	require.Equal(t, s.Goal, current.Goal)
	require.Equal(t, 4, current.DueCount)
	_, err = f.svc.Start(t.Context(), f.u, "no-silent-more")
	require.ErrorIs(t, err, ErrNoDue)
	extra, err := f.svc.StartAdditional(t.Context(), f.u, "deliberate")
	require.NoError(t, err)
	require.Equal(t, 4, extra.TotalWords)
	require.True(t, extra.Goal.Complete)
	extra = f.answer(t, extra, false)
	require.Equal(t, 6, extra.Goal.Reviewed)
	require.Zero(t, extra.Goal.Remaining)
	var rewards, points, streak int
	require.NoError(t, f.db.QueryRow("SELECT count(*),COALESCE(sum(amount),0) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&rewards, &points))
	require.Equal(t, 1, rewards)
	require.Equal(t, 10, points, "wrong answers earn goal completion reward only")
	require.NoError(t, f.db.QueryRow("SELECT current_streak_count FROM streak_states WHERE user_id=$1", f.u).Scan(&streak))
	require.Equal(t, 1, streak)
	_, err = f.svc.Start(t.Context(), f.u, "deliberate")
	require.ErrorIs(t, err, ErrConflict, "additional flag belongs to idempotency fingerprint")
}

func TestDailyGoalSkipHintAndActualNextReview(t *testing.T) {
	f := setup(t, 2)
	s, err := f.svc.Start(t.Context(), f.u, "skip")
	require.NoError(t, err)
	s, err = f.svc.Act(t.Context(), f.u, uuid.MustParse(s.ID), Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: "skip", Action: "skip"}, "skip")
	require.NoError(t, err)
	require.Zero(t, s.Goal.Reviewed)
	s = f.next(t, s)
	st := f.state(t, s)
	s, err = f.svc.Act(t.Context(), f.u, st.ID, Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: "hint", Action: "answer", ChoiceID: st.Words[st.Queue[0]].AnswerID, WasHintUsed: true}, "hint")
	require.NoError(t, err)
	require.Zero(t, s.Goal.Reviewed)
	s = f.next(t, s)
	current, err := f.svc.Current(t.Context(), f.u)
	require.NoError(t, err)
	require.Zero(t, current.DueCount)
	require.False(t, current.Goal.Complete)
	require.Equal(t, 5, current.Goal.Remaining)
	var next time.Time
	require.NoError(t, f.db.QueryRow("SELECT min(next_review_at) FROM user_words WHERE user_id=$1", f.u).Scan(&next))
	require.Equal(t, next, *current.Goal.NextReviewAt)
}

func TestDailyGoalCurrentDayCutoverIsAtomicAndIdempotent(t *testing.T) {
	f := setup(t, 5)
	s, err := f.svc.Start(t.Context(), f.u, "cutover")
	require.NoError(t, err)
	for s.Status != "completed" {
		s = f.answer(t, s, false)
		s = f.next(t, s)
	}
	// Simulate rollout after five legacy verified wrong-answer receipts. Preserve
	// existing mission reward to prove reconciliation cannot award it twice.
	_, err = f.db.Exec("DELETE FROM daily_review_effort WHERE user_id=$1", f.u)
	require.NoError(t, err)
	_, err = f.db.Exec("UPDATE daily_mission_snapshots SET reviews_completed=0,status='open',completed_at=NULL WHERE user_id=$1", f.u)
	require.NoError(t, err)
	_, err = f.db.Exec(`INSERT INTO daily_mission_snapshots(id,user_id,local_date,timezone,review_target,reviews_completed,policy_version,status,created_at,updated_at) SELECT gen_random_uuid(),user_id,local_date-1,timezone,review_target,2,policy_version,'open',created_at,updated_at FROM daily_mission_snapshots WHERE user_id=$1`, f.u)
	require.NoError(t, err)
	for range 3 {
		current, err := f.svc.Current(t.Context(), f.u)
		require.NoError(t, err)
		require.Equal(t, 5, current.Goal.Reviewed)
		require.True(t, current.Goal.Complete)
	}
	var efforts, rewards, past int
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_review_effort WHERE user_id=$1", f.u).Scan(&efforts))
	require.Equal(t, 5, efforts)
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&rewards))
	require.Equal(t, 1, rewards)
	require.NoError(t, f.db.QueryRow("SELECT reviews_completed FROM daily_mission_snapshots WHERE user_id=$1 ORDER BY local_date LIMIT 1", f.u).Scan(&past))
	require.Equal(t, 2, past)
	// A subsequent local day neither recredits old receipts nor rewrites yesterday.
	f.svc.clock = clock.Fixed{T: f.now.Add(24 * time.Hour)}
	fresh, err := f.svc.Current(t.Context(), f.u)
	require.NoError(t, err)
	require.Zero(t, fresh.Goal.Reviewed)
	require.Equal(t, 5, fresh.Goal.Target)
	historical, err := f.svc.Get(t.Context(), f.u, uuid.MustParse(s.ID))
	require.NoError(t, err)
	require.Zero(t, historical.Goal.Reviewed)
	require.Equal(t, 5, historical.TotalWords)
	replay, err := f.svc.Start(t.Context(), f.u, "cutover")
	require.NoError(t, err)
	require.Equal(t, s.Goal.LocalDate, replay.Goal.LocalDate, "start replay retains original response day")
}

func TestDailyGoalReconcilesLegacyAllowanceWithoutLosingFeedback(t *testing.T) {
	f := setup(t, 8)
	_, err := f.db.Exec("UPDATE user_settings SET daily_review_target=10 WHERE user_id=$1", f.u)
	require.NoError(t, err)
	s, err := f.svc.Start(t.Context(), f.u, "old-large-session")
	require.NoError(t, err)
	require.Equal(t, 8, s.TotalWords)
	s = f.answer(t, s, false)
	// Simulate an older oversized frozen session and today's established target.
	_, err = f.db.Exec("UPDATE daily_mission_snapshots SET review_target=5 WHERE user_id=$1", f.u)
	require.NoError(t, err)
	current, err := f.svc.Current(t.Context(), f.u)
	require.NoError(t, err)
	require.Equal(t, 8, current.Session.TotalWords)
	require.Equal(t, s.Feedback, current.Session.Feedback)
	require.Equal(t, 4, current.Goal.Remaining)
	s = f.next(t, current.Session)
	for s.Status != "completed" {
		s = f.answer(t, s, false)
		s = f.next(t, s)
	}
	require.Equal(t, 8, s.TotalWords)
	require.Equal(t, 5, s.QuestionsAnswered)
	require.True(t, s.Goal.Complete)
	current, err = f.svc.Current(t.Context(), f.u)
	require.NoError(t, err)
	require.Equal(t, 3, current.DueCount)
}

func TestDailyGoalCutoverVerifiesEscapedClientIdentity(t *testing.T) {
	f := setup(t, 1)
	s, err := f.svc.Start(t.Context(), f.u, "escaped-start")
	require.NoError(t, err)
	st := f.state(t, s)
	s, err = f.svc.Act(t.Context(), f.u, st.ID, Action{StepID: s.CurrentStep.ID, ExpectedRevision: s.Revision, ClientActionID: "a<&>\u2028\u2029\"\\\n", Action: "answer", ChoiceID: st.Words[0].AnswerID}, "escaped-key")
	require.NoError(t, err)
	_, err = f.db.Exec("DELETE FROM daily_review_effort WHERE user_id=$1", f.u)
	require.NoError(t, err)
	current, err := f.svc.Current(t.Context(), f.u)
	require.NoError(t, err)
	require.Equal(t, 1, current.Goal.Reviewed)
}
