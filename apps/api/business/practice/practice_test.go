package practice

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"
)

func seedWords(t *testing.T) []Word {
	t.Helper()
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	var seed struct {
		Words    []struct{ ID, Text string } `json:"canonical_words"`
		Meanings []struct {
			ID         string
			WordID     string `json:"word_id"`
			Definition string `json:"short_definition"`
		} `json:"word_meanings"`
	}
	require.NoError(t, json.Unmarshal(raw, &seed))
	words := []Word{}
	for _, ref := range catalog {
		for _, m := range seed.Meanings {
			if m.ID == ref.MeaningID {
				for _, w := range seed.Words {
					if w.ID == m.WordID {
						require.Equal(t, ref.WordText, w.Text)
						words = append(words, Word{MeaningID: m.ID, WordText: w.Text, WordSlug: strings.ReplaceAll(w.Text, " ", "-"), Definition: m.Definition, LessonKey: ref.LessonKey})
					}
				}
			}
		}
	}
	require.Len(t, words, 90)
	return words
}
func TestPracticeAllContentProjectionAndListening(t *testing.T) {
	words := seedWords(t)
	seen := map[string]bool{}
	for _, ref := range catalog {
		if seen[ref.LessonKey] {
			continue
		}
		seen[ref.LessonKey] = true
		for _, mode := range []string{"typed_recall", "listening_choice"} {
			snap, err := build(StartRequest{Mode: mode, LessonKey: ref.LessonKey}, words, nil, "seed")
			require.NoError(t, err)
			require.Len(t, snap.Steps, 3)
			for i, step := range snap.Steps {
				st := State{ID: uuid.New(), Snapshot: snap, Index: i}
				p := project(st)
				raw, err := json.Marshal(p)
				require.NoError(t, err)
				for _, secret := range []string{"Accepted", "CorrectChoice", "WordSlug", "MeaningID", "Source"} {
					require.NotContains(t, string(raw), secret)
				}
				require.NotContains(t, string(raw), step.Word.MeaningID)
				require.Nil(t, p.Feedback)
				if mode == "typed_recall" {
					require.Empty(t, p.CurrentStep.Choices)
					require.Empty(t, p.CurrentStep.SpeechText)
					require.NotContains(t, string(raw), `"wordText"`)
				} else {
					require.Len(t, p.CurrentStep.Choices, 3)
					require.NotEmpty(t, p.CurrentStep.SpeechText)
					if step.Word.WordText == "resume" {
						require.Equal(t, "résumé", p.CurrentStep.SpeechText)
					}
					choices := map[string]bool{}
					for _, c := range p.CurrentStep.Choices {
						require.NotEqual(t, step.Word.MeaningID, c.ID)
						require.False(t, choices[c.Text])
						choices[c.Text] = true
					}
				}
			}
		}
	}
}
func TestPracticeTypedAliasesAssistanceAndResolution(t *testing.T) {
	words := seedWords(t)
	for _, w := range words {
		snap, err := build(StartRequest{Mode: "typed_recall", LessonKey: w.LessonKey}, words, nil, "seed")
		require.NoError(t, err)
		var step privateStep
		for _, s := range snap.Steps {
			if s.Word.MeaningID == w.MeaningID {
				step = s
			}
		}
		step.Source = &Source{Kind: "lesson", ID: uuid.NewString()}
		for _, answer := range acceptedAnswers(w.WordText) {
			st := State{Snapshot: Snapshot{Steps: []privateStep{step}}}
			resolved, err := apply(&st, Action{StepID: step.Public.ID, Action: "answer", TypedAnswer: "  " + strings.ToUpper(answer) + "  "}, time.Now())
			require.NoError(t, err)
			require.True(t, st.Feedback.Correct, answer)
			require.True(t, resolved)
			require.Equal(t, 1, st.FirstAnswersCorrect)
		}
	}
	snap, err := build(StartRequest{Mode: "typed_recall", LessonKey: "airport"}, words, nil, "seed")
	require.NoError(t, err)
	step := snap.Steps[0]
	step.Source = &Source{Kind: "review", ID: uuid.NewString()}
	st := State{Snapshot: Snapshot{Steps: []privateStep{step}}}
	a := Action{StepID: step.Public.ID, Action: "answer", TypedAnswer: "valid but different synonym"}
	resolved, err := apply(&st, a, time.Now())
	require.NoError(t, err)
	require.False(t, resolved)
	require.False(t, st.Feedback.Correct)
	require.Contains(t, st.Feedback.Explanation, "The word from this lesson is")
	a.ExpectedRevision = 1
	a.TypedAnswer = step.Word.WordText
	resolved, err = apply(&st, a, time.Now())
	require.NoError(t, err)
	require.False(t, resolved)
	require.True(t, st.Feedback.Assisted)
	require.Zero(t, st.FirstAnswersCorrect)
	require.Equal(t, 1, st.QuestionsAnswered)
	_, err = apply(&st, Action{StepID: step.Public.ID, Action: "continue", ExpectedRevision: 2}, time.Now())
	require.NoError(t, err)
	require.NotNil(t, st.CompletedAt)
	st = State{Snapshot: Snapshot{Steps: []privateStep{step}}}
	resolved, err = apply(&st, Action{StepID: step.Public.ID, Action: "reveal"}, time.Now())
	require.NoError(t, err)
	require.False(t, resolved)
	require.True(t, st.Feedback.Assisted)
	require.False(t, st.Feedback.Correct)
	require.True(t, project(st).CanContinue)
}
func TestPracticeRejectsForgedAndStaleAnswers(t *testing.T) {
	snap, err := build(StartRequest{Mode: "listening_choice", LessonKey: "airport"}, seedWords(t), nil, "seed")
	require.NoError(t, err)
	step := snap.Steps[0]
	for _, a := range []Action{{StepID: step.Public.ID, Action: "answer", ChoiceID: step.Word.MeaningID}, {StepID: step.Public.ID, Action: "answer", ChoiceID: step.CorrectChoice, TypedAnswer: "also text"}, {StepID: step.Public.ID, Action: "continue"}, {StepID: "wrong-step", Action: "answer", ChoiceID: step.CorrectChoice}, {StepID: step.Public.ID, Action: "answer", ChoiceID: step.CorrectChoice, ExpectedRevision: 1}} {
		st := State{Snapshot: snap}
		_, err = apply(&st, a, time.Now())
		require.Error(t, err)
		require.Zero(t, st.Revision)
	}
	st := State{Snapshot: snap}
	_, err = apply(&st, Action{StepID: step.Public.ID, Action: "answer", ChoiceID: step.CorrectChoice}, time.Now())
	require.NoError(t, err)
	require.True(t, st.Feedback.Correct)
	require.Equal(t, 1, st.FirstAnswersCorrect)
}

func TestPracticeTypedCanonicalUnicodeEquivalence(t *testing.T) {
	words := seedWords(t)
	snap, err := build(StartRequest{Mode: "typed_recall", LessonKey: "job-interview"}, words, nil, "unicode")
	require.NoError(t, err)
	var step privateStep
	for _, candidate := range snap.Steps {
		if candidate.Word.WordText == "resume" {
			step = candidate
		}
	}
	require.NotEmpty(t, step.Public.ID)
	// Independent input vectors: mixed compositions must match the same noun.
	for _, answer := range []string{"re\u0301sumé", "résume\u0301", "RE\u0301SUMÉ", "  RÉSUME\u0301  ", "resume\u0301"} {
		st := State{Snapshot: Snapshot{Steps: []privateStep{step}}}
		_, err = apply(&st, Action{StepID: step.Public.ID, Action: "answer", TypedAnswer: answer}, time.Now())
		require.NoError(t, err)
		require.True(t, st.Feedback.Correct, answer)
	}
	for _, answer := range []string{"resumes", "résumés", "resume!", "curriculum vitae"} {
		st := State{Snapshot: Snapshot{Steps: []privateStep{step}}}
		_, err = apply(&st, Action{StepID: step.Public.ID, Action: "answer", TypedAnswer: answer}, time.Now())
		require.NoError(t, err)
		require.False(t, st.Feedback.Correct, answer)
	}
}
