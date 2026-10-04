package lessons

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestAllMixedLessonSnapshotsGradeAndRetainOriginalBuilder(t *testing.T) {
	for _, d := range catalog {
		t.Run(d.Key, func(t *testing.T) {
			words := seedWords(t, d)
			snap, e := buildVariedSnapshot(d, words, "stable")
			require.NoError(t, e)
			again, e := buildVariedSnapshot(d, words, "stable")
			require.NoError(t, e)
			require.Equal(t, snap, again)
			require.Equal(t, ExerciseVersion, snap.ExerciseVersion)
			legacy, e := buildSnapshot(d, words, "stable")
			require.NoError(t, e)
			require.Empty(t, legacy.ExerciseVersion)
			require.Equal(t, "recall", legacy.Steps[4].Step.Kind)
			require.Equal(t, "recall", legacy.Steps[5].Step.Kind)
			kinds := []string{}
			for _, step := range snap.Steps {
				kinds = append(kinds, step.Step.Kind)
			}
			require.Equal(t, []string{"teach", "teach", "teach", "recall", "typed_recall", "listening_choice", "context", "context", "context"}, kinds)
			require.Empty(t, snap.Steps[4].Step.Choices)
			require.Equal(t, []string{words[1].WordText}, snap.Steps[4].Accepted)
			require.Equal(t, d.Words[1].Context, snap.Steps[4].Step.Context)
			require.Len(t, snap.Steps[5].Step.Choices, 3)
			require.Equal(t, words[2].WordText, snap.Steps[5].Step.SpeechText)
			require.Equal(t, "en-US", snap.Steps[5].Step.SpeechLanguage)
			st := State{ID: uuid.New(), UserID: uuid.New(), Snapshot: snap}
			now := time.Now().UTC()
			for st.Index < 9 {
				step := snap.Steps[st.Index]
				if step.Step.Kind != "teach" {
					a := Action{StepID: step.Step.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: step.CorrectChoiceID}
					if step.Step.Kind == "typed_recall" {
						a.TypedAnswer = "wrong"
						require.NoError(t, apply(&st, a, now))
						require.False(t, st.Feedback.Correct)
						require.Equal(t, words[1].WordText, st.Feedback.Answer)
						require.ErrorIs(t, apply(&st, Action{StepID: step.Step.ID, ExpectedRevision: st.Revision, Action: "continue"}, now), ErrConflict)
						a.ExpectedRevision = st.Revision
						a.TypedAnswer = " " + strings.ToUpper(words[1].WordText) + " "
					}
					require.NoError(t, apply(&st, a, now))
					require.True(t, st.Feedback.Correct)
				}
				require.NoError(t, apply(&st, Action{StepID: step.Step.ID, ExpectedRevision: st.Revision, Action: "continue"}, now))
			}
			p := project(st)
			require.Equal(t, "completed", p.Status)
			require.Equal(t, ExerciseVersion, p.ExerciseVersion)
			require.Equal(t, 5, p.FirstAnswersCorrect)
			require.Equal(t, 6, p.QuestionsAnswered)
			raw, e := json.Marshal(p)
			require.NoError(t, e)
			require.NotContains(t, string(raw), "accepted")
		})
	}
}
func TestTypedLessonInvalidActionsNeverMutateAndNormalizeNFC(t *testing.T) {
	d := catalog[0]
	snap, e := buildVariedSnapshot(d, seedWords(t, d), "test")
	require.NoError(t, e)
	snap.Steps[4].Accepted = []string{"café time"}
	snap.Steps[4].Step.Word.WordText = "café time"
	st := State{Snapshot: snap, Index: 4, Revision: 4}
	for _, a := range []Action{{StepID: "typed_recall-2", ExpectedRevision: 4, Action: "answer", TypedAnswer: "", ChoiceID: "forged"}, {StepID: "typed_recall-2", ExpectedRevision: 4, Action: "answer", TypedAnswer: "   "}, {StepID: "typed_recall-2", ExpectedRevision: 4, Action: "answer", TypedAnswer: "café time", ChoiceID: "forged"}, {StepID: "typed_recall-2", ExpectedRevision: 4, Action: "answer", TypedAnswer: "bad\x00answer"}, {StepID: "typed_recall-2", ExpectedRevision: 4, Action: "answer", TypedAnswer: strings.Repeat("界", 67)}, {StepID: "typed_recall-2", ExpectedRevision: 4, Action: "continue", TypedAnswer: "café time"}, {StepID: "typed_recall-2", ExpectedRevision: 4, Action: "answer", TypedAnswer: string([]byte{0xff})}} {
		before, _ := json.Marshal(st)
		require.ErrorIs(t, apply(&st, a, time.Now()), ErrInvalid)
		after, _ := json.Marshal(st)
		require.Equal(t, before, after)
	}
	require.NoError(t, apply(&st, Action{StepID: "typed_recall-2", ExpectedRevision: 4, Action: "answer", TypedAnswer: " CAFE\u0301   TIME "}, time.Now()))
	require.True(t, st.Feedback.Correct)
	require.Equal(t, 1, st.FirstAnswersCorrect)
}
func TestLegacyLessonActionFingerprintAndSnapshotOmissions(t *testing.T) {
	legacy := struct {
		StepID           string `json:"stepId"`
		ExpectedRevision int    `json:"expectedRevision"`
		ClientActionID   string `json:"clientActionId"`
		Action           string `json:"action"`
		ChoiceID         string `json:"choiceId,omitempty"`
	}{"recall-1", 3, "answer-one", "answer", "canonical"}
	require.Equal(t, fingerprint(legacy), fingerprint(Action{StepID: "recall-1", ExpectedRevision: 3, ClientActionID: "answer-one", Action: "answer", ChoiceID: "canonical"}))
	snap, e := buildSnapshot(catalog[0], seedWords(t, catalog[0]), "legacy")
	require.NoError(t, e)
	raw, e := json.Marshal(snap)
	require.NoError(t, e)
	require.NotContains(t, string(raw), "exerciseVersion")
	require.NotContains(t, string(raw), "accepted")
	require.NotContains(t, string(raw), "speechText")
}
