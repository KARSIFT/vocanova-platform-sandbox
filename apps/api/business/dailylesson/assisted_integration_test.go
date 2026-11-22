package dailylesson

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// "I'm not sure" shows the meaning with help. It uses the miss schedule, stays
// replay-safe and never earns a correct grade, success credit or goal effort.
func TestPostgreSQLDailyLessonAssistedTeachingKeepsScheduleAndEarnsNoCredit(t *testing.T) {
	f := setup(t, 2)
	ctx := t.Context()
	s, e := f.svc.Start(ctx, f.u, "start")
	require.NoError(t, e)
	require.Equal(t, 2, s.TotalWords)
	st := f.state(t, s)

	// An answer still needs a choice unless help was requested.
	_, e = f.svc.Act(ctx, f.u, st.ID, Action{StepID: s.CurrentStep.ID, ExpectedRevision: 0, ClientActionID: "no-choice", Action: "answer"}, "no-choice")
	require.ErrorIs(t, e, ErrInvalid)

	assistedWord := st.Words[st.Queue[0]]
	unsure := Action{StepID: s.CurrentStep.ID, ExpectedRevision: 0, ClientActionID: "not-sure", Action: "answer", WasHintUsed: true}
	out, e := f.svc.Act(ctx, f.u, st.ID, unsure, "not-sure")
	require.NoError(t, e)
	fb := out.Feedback
	require.NotNil(t, fb)
	require.False(t, fb.Correct)
	require.False(t, fb.Credited)
	require.False(t, fb.Skipped)
	require.Equal(t, AssistedReason, fb.Reason)
	require.Empty(t, fb.SelectedChoiceID)
	require.Equal(t, assistedWord.AnswerID, fb.CorrectChoiceID, "the meaning is shown")
	require.Equal(t, assistedWord.Definition, fb.Answer)
	require.Contains(t, fb.Explanation, "practice with help")
	require.NotNil(t, fb.NextReviewAt)
	require.NotNil(t, fb.ReviewStep)
	require.Equal(t, 1, out.QuestionsAnswered, "an assisted question counts as answered, not correct")
	require.Zero(t, out.FirstAnswersCorrect)
	require.Zero(t, out.CreditedToday)

	// Reload and retry return the same immutable result and write nothing new.
	reloaded, e := f.svc.Get(ctx, f.u, st.ID)
	require.NoError(t, e)
	require.Equal(t, fb, reloaded.Feedback)
	replayed, e := f.svc.Act(ctx, f.u, st.ID, unsure, "not-sure")
	require.NoError(t, e)
	require.Equal(t, fb, replayed.Feedback)
	require.Equal(t, out.Revision, replayed.Revision)
	var attempts, effort, successes int
	var hint bool
	var result string
	require.NoError(t, f.db.QueryRow("SELECT count(*),bool_and(was_hint_used),min(result) FROM review_attempts WHERE user_id=$1", f.u).Scan(&attempts, &hint, &result))
	require.Equal(t, 1, attempts)
	require.True(t, hint)
	require.Equal(t, "incorrect", result)
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_review_effort WHERE user_id=$1", f.u).Scan(&effort))
	require.Zero(t, effort, "assisted work is not objective review effort")
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM daily_review_successes WHERE user_id=$1", f.u).Scan(&successes))
	require.Zero(t, successes)
	var points int
	require.NoError(t, f.db.QueryRow("SELECT count(*) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&points))
	require.Zero(t, points, "no rating points for assisted work")
	require.Zero(t, out.Goal.Reviewed)

	s = f.next(t, out)
	s = f.answer(t, s, false)
	require.Empty(t, s.Feedback.Reason, "an unassisted miss is not labelled assisted")
	require.Contains(t, s.Feedback.Explanation, "different")
	require.Equal(t, 1, s.Goal.Reviewed, "only the unassisted miss counts as effort")

	// Both words follow exactly the same miss schedule.
	var assistedStep, missStep int
	var assistedDue, missDue time.Time
	require.NoError(t, f.db.QueryRow("SELECT review_step,next_review_at FROM user_words WHERE id=$1", assistedWord.UserWordID).Scan(&assistedStep, &assistedDue))
	missWord := f.state(t, s).Words[f.state(t, s).Queue[0]]
	require.NoError(t, f.db.QueryRow("SELECT review_step,next_review_at FROM user_words WHERE id=$1", missWord.UserWordID).Scan(&missStep, &missDue))
	require.Equal(t, missStep, assistedStep)
	require.True(t, missDue.Equal(assistedDue))
	require.True(t, assistedDue.Equal(*fb.NextReviewAt))

	s = f.next(t, s)
	require.Equal(t, "completed", s.Status)
	require.Equal(t, 2, s.QuestionsAnswered)
	require.Zero(t, s.FirstAnswersCorrect)

	// The assisted word is done for today and does not come back as due.
	current, e := f.svc.Current(ctx, f.u)
	require.NoError(t, e)
	require.Zero(t, current.DueCount)
	_, e = f.svc.Act(ctx, f.u, uuid.MustParse(s.ID), unsure, "not-sure-again")
	require.ErrorIs(t, e, ErrConflict)
}

// A learner may still change the date after assisted teaching, using the same
// confirmed schedule check as any answered word.
func TestPostgreSQLDailyLessonAssistedAllowsSameReschedule(t *testing.T) {
	f := setup(t, 1)
	s, e := f.svc.Start(t.Context(), f.u, "start")
	require.NoError(t, e)
	st := f.state(t, s)
	out, e := f.svc.Act(t.Context(), f.u, st.ID, Action{StepID: s.CurrentStep.ID, ExpectedRevision: 0, ClientActionID: "unsure", Action: "answer", WasHintUsed: true}, "unsure")
	require.NoError(t, e)
	done, e := f.svc.Act(t.Context(), f.u, st.ID, Action{StepID: s.CurrentStep.ID, ExpectedRevision: out.Revision, ClientActionID: "later", Action: "continue", NextReview: "3d"}, "later")
	require.NoError(t, e)
	require.Equal(t, "completed", done.Status)
	var due time.Time
	require.NoError(t, f.db.QueryRow("SELECT next_review_at FROM user_words WHERE id=$1", st.Words[0].UserWordID).Scan(&due))
	require.True(t, due.Equal(f.now.Add(72*time.Hour)))
	require.Zero(t, done.CreditedToday)
}
