package practice

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func listeningChoiceState(t *testing.T) State {
	t.Helper()
	snapshot, err := build(StartRequest{Mode: "listening_choice", LessonKey: "airport"}, seedWords(t), nil, uuid.NewString())
	require.NoError(t, err)
	return State{ID: uuid.New(), UserID: uuid.New(), Snapshot: snapshot}
}

func confirmedChoiceJSON(t *testing.T, session Session) string {
	t.Helper()
	raw, err := json.Marshal(session)
	require.NoError(t, err)
	return string(raw)
}

func TestCorrectListeningChoiceRestoresFromHistoricalStoredFeedback(t *testing.T) {
	st := listeningChoiceState(t)
	step := st.Snapshot.Steps[0]
	st.Snapshot.ContentVersion = "starter-21-v1"
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	_, err := apply(&st, Action{StepID: step.Public.ID, Action: "answer", ChoiceID: step.CorrectChoice}, now)
	require.NoError(t, err)
	require.True(t, st.Feedback.Correct)
	require.NotEqual(t, step.Word.MeaningID, step.CorrectChoice)
	require.NotEqual(t, step.Word.WordText, step.CorrectChoice)
	before := *st.Feedback
	answered := project(st)
	require.Equal(t, step.CorrectChoice, answered.Feedback.CorrectChoiceID)
	require.Contains(t, confirmedChoiceJSON(t, answered), `"correctChoiceId":"`+step.CorrectChoice+`"`)
	require.Equal(t, before, *st.Feedback, "projection must not enrich persisted feedback or a shared receipt")
	require.NotSame(t, st.Feedback, answered.Feedback)

	// Simulate an existing persisted session: its feedback JSON lacks the new
	// optional field, so restoring must use its original opaque snapshot ID.
	snapshotJSON, err := json.Marshal(st.Snapshot)
	require.NoError(t, err)
	feedbackJSON, err := json.Marshal(st.Feedback)
	require.NoError(t, err)
	require.NotContains(t, string(feedbackJSON), "correctChoiceId")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectQuery("SELECT.*FROM practice_sessions").WithArgs(st.ID, st.UserID).WillReturnRows(sqlmock.NewRows([]string{
		"id", "user_id", "snapshot", "current_step", "revision", "feedback", "first_answers_correct", "questions_answered", "created_at", "updated_at", "completed_at",
	}).AddRow(st.ID, st.UserID, snapshotJSON, 0, st.Revision, feedbackJSON, st.FirstAnswersCorrect, st.QuestionsAnswered, now, now, nil))
	reloaded, err := NewService(NewPostgreSQLRepository(db), clock.Fixed{T: now}).Get(t.Context(), st.UserID, st.ID)
	require.NoError(t, err)
	require.Equal(t, step.Public.ID, reloaded.CurrentStep.ID)
	require.Equal(t, step.Public.Choices, reloaded.CurrentStep.Choices)
	require.Equal(t, step.CorrectChoice, reloaded.Feedback.CorrectChoiceID)
	require.Contains(t, confirmedChoiceJSON(t, *reloaded), `"correctChoiceId":"`+step.CorrectChoice+`"`)
	require.False(t, reloaded.Feedback.Assisted)
	require.Equal(t, 1, reloaded.FirstAnswersCorrect)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWrongOrRevealedListeningChoiceDoesNotConfirmUntilCorrectRetry(t *testing.T) {
	for _, initialAction := range []string{"wrong", "reveal"} {
		t.Run(initialAction, func(t *testing.T) {
			st := listeningChoiceState(t)
			step := st.Snapshot.Steps[0]
			require.NotContains(t, confirmedChoiceJSON(t, project(st)), "correctChoiceId", "ungraded steps must not identify the answer")
			action := Action{StepID: step.Public.ID, Action: "reveal"}
			if initialAction == "wrong" {
				action.Action = "answer"
				for _, choice := range step.Public.Choices {
					if choice.ID != step.CorrectChoice {
						action.ChoiceID = choice.ID
						break
					}
				}
			}
			_, err := apply(&st, action, time.Now())
			require.NoError(t, err)
			feedback := project(st).Feedback
			require.False(t, feedback.Correct)
			require.Empty(t, feedback.CorrectChoiceID)
			require.NotContains(t, confirmedChoiceJSON(t, project(st)), "correctChoiceId")
			_, err = apply(&st, Action{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: step.CorrectChoice}, time.Now())
			require.NoError(t, err)
			retried := project(st)
			require.True(t, retried.Feedback.Correct)
			require.True(t, retried.Feedback.Assisted)
			require.Equal(t, step.CorrectChoice, retried.Feedback.CorrectChoiceID)
			require.Zero(t, retried.FirstAnswersCorrect, "restored assisted choices must not grant independent credit")
			require.Equal(t, 1, retried.QuestionsAnswered)
		})
	}
}

func TestConfirmedListeningChoiceRejectsStaleOrUntrustedStoredIdentity(t *testing.T) {
	for _, scenario := range []string{"stale-step", "wrong", "reveal", "typed", "unknown-snapshot-choice"} {
		t.Run(scenario, func(t *testing.T) {
			st := listeningChoiceState(t)
			step := &st.Snapshot.Steps[0]
			st.Feedback = &Feedback{StepID: step.Public.ID, Correct: true, CorrectChoiceID: "untrusted-stored-choice"}
			switch scenario {
			case "stale-step":
				st.Feedback.StepID = st.Snapshot.Steps[1].Public.ID
			case "wrong":
				st.Feedback.Correct = false
			case "reveal":
				st.Feedback.Correct = false
				st.Feedback.Assisted = true
			case "typed":
				step.Public.Kind = "typed_recall"
			case "unknown-snapshot-choice":
				step.CorrectChoice = "not-in-public-choices"
			}
			before := *st.Feedback
			response := project(st)
			require.Empty(t, response.Feedback.CorrectChoiceID)
			require.NotContains(t, confirmedChoiceJSON(t, response), "correctChoiceId")
			require.Equal(t, before, *st.Feedback)
		})
	}
}
