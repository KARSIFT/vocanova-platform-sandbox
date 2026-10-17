package practice

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"
)

func TestJourneyBroadCatalogHasAvailableMeaningChoices(t *testing.T) {
	raw, err := os.ReadFile("../knowledgecatalog/catalog.json.gz")
	require.NoError(t, err)
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	require.NoError(t, err)
	defer reader.Close()
	var catalog struct {
		Items []struct {
			Word, Definition string
			MeaningID        string `json:"meaning_id"`
		}
	}
	require.NoError(t, json.NewDecoder(reader).Decode(&catalog))
	require.Len(t, catalog.Items, 10000)
	pool := seedWords(t)
	for _, entry := range catalog.Items {
		w := Word{MeaningID: entry.MeaningID, WordText: entry.Word, Definition: entry.Definition}
		step, err := journeyChoice(w, pool, "catalog-availability", "meaning_choice")
		require.NoError(t, err, entry.Word)
		require.GreaterOrEqual(t, len(step.Public.Choices), 3, entry.Word)
		seen := map[string]bool{}
		correct := 0
		for _, choice := range step.Public.Choices {
			require.False(t, seen[normalize(choice.Text)], entry.Word)
			seen[normalize(choice.Text)] = true
			if choice.ID == step.CorrectChoice {
				correct++
			}
		}
		require.Equal(t, 1, correct, entry.Word)
	}
}

func TestPersonalJourneyMixedPracticeUsesOnlyAuthoredContextAndSeparatesProgress(t *testing.T) {
	pool := seedWords(t)
	pool[0].Example = "A greeting can start a conversation."
	pool[1].Example = "We made small talk before the meeting."
	known := []Word{pool[0], pool[1]}
	familiar := map[string]bool{pool[0].MeaningID: true, pool[1].MeaningID: true}
	snap, err := buildJourney(known, pool, familiar, 1, "journey-test")
	require.NoError(t, err)
	require.Equal(t, 6, len(snap.Steps))
	require.Zero(t, snap.NewWords)
	require.Equal(t, []string{"listening_choice", "gap_choice", "sentence_build", "listening_choice", "gap_choice", "sentence_build"}, []string{snap.Steps[0].Public.Kind, snap.Steps[1].Public.Kind, snap.Steps[2].Public.Kind, snap.Steps[3].Public.Kind, snap.Steps[4].Public.Kind, snap.Steps[5].Public.Kind})
	require.NotContains(t, snap.Steps[1].Public.Context, "greeting")
	require.Empty(t, snap.Steps[2].Public.Context)
	st := State{ID: uuid.New(), Snapshot: snap, Index: 2}
	initial := project(st)
	require.Nil(t, initial.Feedback, "initial projection cannot leak token ordering")
	step := snap.Steps[2]
	bad := append([]string{}, step.CorrectTokens...)
	bad[0], bad[1] = bad[1], bad[0]
	a := Action{StepID: step.Public.ID, Action: "answer", TokenIDs: bad}
	_, err = apply(&st, a, time.Now())
	require.NoError(t, err)
	require.False(t, st.Feedback.Correct)
	require.Equal(t, step.CorrectTokens, st.Feedback.CorrectTokenIDs)
	require.Equal(t, bad, st.Feedback.SelectedTokenIDs)
	require.False(t, project(st).CanContinue)
	_, err = apply(&st, Action{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "continue"}, time.Now())
	require.ErrorIs(t, err, ErrConflict)
	a.ExpectedRevision = st.Revision
	a.TokenIDs = step.CorrectTokens
	_, err = apply(&st, a, time.Now())
	require.NoError(t, err)
	require.True(t, st.Feedback.Correct)
	require.True(t, st.Feedback.Assisted)
	require.Zero(t, st.FirstAnswersCorrect)
	_, err = apply(&st, Action{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "continue"}, time.Now())
	require.NoError(t, err)
	require.Equal(t, 3, st.Index)

	// A broad catalog meaning has no fabricated usage sentence: choose a safe
	// meaning/listening task instead, preserving access beyond starter content.
	broad := Word{MeaningID: uuid.NewString(), WordText: "unfamiliar-synthetic", Definition: "A synthetic distinct reference definition."}
	fallback, err := buildJourney([]Word{broad}, pool, map[string]bool{broad.MeaningID: true}, 2, "fallback")
	require.NoError(t, err)
	require.Equal(t, 2, len(fallback.Steps))
	require.Equal(t, "meaning_choice", fallback.Steps[1].Public.Kind)
	for _, s := range fallback.Steps {
		require.Empty(t, s.Public.Context)
	}

	third, err := buildJourney(known, pool, familiar, 3, "third")
	require.NoError(t, err)
	require.Equal(t, 1, third.NewWords)
	var teaches int
	for _, s := range third.Steps {
		if s.Public.Kind == "teach" {
			teaches++
			require.False(t, familiar[s.Word.MeaningID])
			require.True(t, project(State{Snapshot: Snapshot{Mode: "journey", Steps: []privateStep{s}}}).CanContinue)
		}
	}
	require.Equal(t, 1, teaches)
}

func TestJourneyTeachingAndUnavailableAudioRemainHonest(t *testing.T) {
	pool := seedWords(t)
	snap, err := buildJourney(nil, pool, map[string]bool{}, 1, "newcomer")
	require.NoError(t, err)
	require.Equal(t, 2, snap.NewWords)
	require.Equal(t, "teach", snap.Steps[0].Public.Kind)
	st := State{Snapshot: snap}
	require.Equal(t, snap.Steps[0].Word.MeaningID, project(st).CurrentStep.MeaningID)
	for i, saved := range snap.Steps {
		if saved.Public.Kind != "teach" {
			projection := project(State{Snapshot: snap, Index: i})
			require.Empty(t, projection.CurrentStep.MeaningID, "graded questions must not expose a picture hint")
		}
	}
	_, err = apply(&st, Action{StepID: snap.Steps[0].Public.ID, Action: "answer", ChoiceID: "pretend"}, time.Now())
	require.ErrorIs(t, err, ErrInvalid)
	_, err = apply(&st, Action{StepID: snap.Steps[0].Public.ID, Action: "continue"}, time.Now())
	require.NoError(t, err)
	require.Zero(t, st.QuestionsAnswered)
	step := st.Snapshot.Steps[st.Index]
	var wrong string
	for _, choice := range step.Public.Choices {
		if choice.ID != step.CorrectChoice {
			wrong = choice.ID
			break
		}
	}
	_, err = apply(&st, Action{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: wrong}, time.Now())
	require.NoError(t, err)
	require.False(t, st.Feedback.Correct)
	require.False(t, project(st).CanContinue)
	_, err = apply(&st, Action{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "reveal"}, time.Now())
	require.NoError(t, err)
	require.True(t, st.Feedback.Revealed)
	require.True(t, st.Feedback.Assisted)
	require.False(t, st.Feedback.Correct)
	require.Zero(t, st.FirstAnswersCorrect)
	require.True(t, project(st).CanContinue)
	_, err = apply(&st, Action{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "continue"}, time.Now())
	require.NoError(t, err)
}

func TestJourneyAllKnownDoesNotInventNewAndTokenIdentityIsValidated(t *testing.T) {
	pool := seedWords(t)
	familiar := map[string]bool{}
	for _, w := range pool {
		familiar[w.MeaningID] = true
	}
	snap, err := buildJourney(pool[:2], pool, familiar, 3, "all-known")
	require.NoError(t, err)
	require.Zero(t, snap.NewWords)
	for _, s := range snap.Steps {
		require.NotEqual(t, "teach", s.Public.Kind)
	}
	// Equal visible words can use either token identity at the same position;
	// tokens with different identities must each be used exactly once.
	step := privateStep{Word: Word{Example: "We can can."}, Public: Step{ID: "tokens", Kind: "sentence_build", Choices: []Choice{{ID: "a", Text: "We"}, {ID: "b", Text: "can"}, {ID: "c", Text: "can"}}}, CorrectTokens: []string{"a", "b", "c"}}
	st := State{Snapshot: Snapshot{Mode: "journey", Steps: []privateStep{step}}}
	_, err = apply(&st, Action{StepID: "tokens", Action: "answer", TokenIDs: []string{"a", "a", "c"}}, time.Now())
	require.ErrorIs(t, err, ErrInvalid)
	_, err = apply(&st, Action{StepID: "tokens", Action: "answer", TokenIDs: []string{"a", "b", "foreign"}}, time.Now())
	require.ErrorIs(t, err, ErrInvalid)
	_, err = apply(&st, Action{StepID: "tokens", Action: "answer", TokenIDs: []string{"a", "c", "b"}}, time.Now())
	require.NoError(t, err)
	require.True(t, st.Feedback.Correct)
}

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
