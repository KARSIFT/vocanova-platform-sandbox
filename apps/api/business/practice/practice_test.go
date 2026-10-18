package practice

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
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
				require.Empty(t, p.CurrentStep.MeaningID, "graded steps cannot expose teaching imagery")
				require.Nil(t, p.Feedback)
				if mode == "typed_recall" {
					require.NotContains(t, string(raw), step.Word.MeaningID)
					require.Empty(t, p.CurrentStep.PronunciationMeaningID)
					require.Empty(t, p.CurrentStep.Choices)
					require.Empty(t, p.CurrentStep.SpeechText)
					require.NotContains(t, string(raw), `"wordText"`)
				} else {
					require.Equal(t, step.Word.MeaningID, p.CurrentStep.PronunciationMeaningID)
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

func TestPracticeFrozenPronunciationUseAndFirstAnswerGrading(t *testing.T) {
	pool := seedWords(t)
	uses := []Word{
		{MeaningID: uuid.NewString(), WordText: "resume", Definition: "A document describing professional experience.", PartOfSpeech: "noun"},
		{MeaningID: uuid.NewString(), WordText: "resume", Definition: "To continue after a pause.", PartOfSpeech: "verb"},
		{MeaningID: uuid.NewString(), WordText: "record", Definition: "Stored information about an event.", PartOfSpeech: "noun"},
		{MeaningID: uuid.NewString(), WordText: "record", Definition: "To store sound for later listening.", PartOfSpeech: "verb"},
		{MeaningID: uuid.NewString(), WordText: "fast", Definition: "Able to move at high speed.", PartOfSpeech: "adjective"},
		{MeaningID: uuid.NewString(), WordText: "fast", Definition: "At high speed.", PartOfSpeech: "adverb"},
		{MeaningID: uuid.NewString(), WordText: "under", Definition: "In a lower position than something.", PartOfSpeech: "preposition"},
		{MeaningID: uuid.NewString(), WordText: "and", Definition: "Used to connect two things.", PartOfSpeech: "conjunction"},
	}
	for _, w := range uses {
		t.Run(w.WordText+"/"+w.PartOfSpeech, func(t *testing.T) {
			snap, err := buildJourney([]Word{w}, pool, map[string]bool{w.MeaningID: true}, 1, "frozen-use")
			require.NoError(t, err)
			st := State{ID: uuid.New(), Snapshot: snap}
			before := project(st)
			require.Equal(t, w.MeaningID, before.CurrentStep.PronunciationMeaningID)
			require.Equal(t, meaningidentity.Revision(w.MeaningID, w.WordText, w.PartOfSpeech, w.Definition), before.CurrentStep.PronunciationRevision)
			require.NotEmpty(t, before.CurrentStep.PronunciationRevision)
			require.Equal(t, w.PartOfSpeech, before.CurrentStep.PartOfSpeech)
			if w.WordText == "resume" {
				if w.PartOfSpeech == "noun" {
					require.Equal(t, "résumé", before.CurrentStep.SpeechText)
				} else {
					require.Equal(t, "resume", before.CurrentStep.SpeechText)
				}
			}
			require.Empty(t, before.CurrentStep.MeaningID)
			// Persistence and repeated projection cannot choose another sense.
			raw, err := json.Marshal(st)
			require.NoError(t, err)
			var resumed State
			require.NoError(t, json.Unmarshal(raw, &resumed))
			require.Equal(t, before, project(resumed))
			step := resumed.Snapshot.Steps[0]
			var wrong string
			for _, choice := range step.Public.Choices {
				if choice.ID == step.CorrectChoice {
					require.Equal(t, w.Definition, choice.Text)
				} else {
					wrong = choice.ID
				}
			}
			_, err = apply(&resumed, Action{StepID: step.Public.ID, Action: "answer", ChoiceID: wrong}, time.Now())
			require.NoError(t, err)
			require.False(t, resumed.Feedback.Correct)
			require.Equal(t, w.MeaningID, resumed.Feedback.MeaningID)
			require.Equal(t, w.PartOfSpeech, project(resumed).Feedback.PartOfSpeech)
			require.Equal(t, before.CurrentStep, project(resumed).CurrentStep)
			_, err = apply(&resumed, Action{StepID: step.Public.ID, ExpectedRevision: resumed.Revision, Action: "answer", ChoiceID: step.CorrectChoice}, time.Now())
			require.NoError(t, err)
			require.True(t, resumed.Feedback.Correct)
			require.True(t, resumed.Feedback.Assisted)
			require.Zero(t, resumed.FirstAnswersCorrect)
			require.Equal(t, 1, resumed.QuestionsAnswered)
			fresh := State{Snapshot: snap}
			_, err = apply(&fresh, Action{StepID: step.Public.ID, Action: "answer", ChoiceID: step.CorrectChoice}, time.Now())
			require.NoError(t, err)
			require.Equal(t, 1, fresh.FirstAnswersCorrect)
		})
	}
	// Old snapshots gain audio context without inventing missing POS metadata.
	legacy := State{Snapshot: Snapshot{Steps: []privateStep{{Word: Word{MeaningID: "old-meaning"}, Public: Step{Kind: "listening_choice", SpeechText: "record"}}}}}
	require.Empty(t, project(legacy).CurrentStep.PartOfSpeech)
	require.Empty(t, project(legacy).CurrentStep.PronunciationRevision)
	require.Equal(t, "old-meaning", project(legacy).CurrentStep.PronunciationMeaningID)
}

func TestPracticeSiblingUsesAreTaughtBeforePracticeAndSeededAcrossSessions(t *testing.T) {
	known := Word{MeaningID: uuid.NewString(), WordText: "present", PartOfSpeech: "noun", Definition: "A gift given to another person."}
	siblings := []Word{
		{MeaningID: uuid.NewString(), WordText: "present", PartOfSpeech: "verb", Definition: "To show something to people."},
		{MeaningID: uuid.NewString(), WordText: "present", PartOfSpeech: "adjective", Definition: "Being in this place now."},
	}
	pool := append(seedWords(t), siblings...)
	familiar := map[string]bool{known.MeaningID: true}
	seen := map[string]bool{}
	for i := 0; i < 32; i++ {
		snap, err := buildJourney([]Word{known}, pool, familiar, 3, string(rune('a'+i)))
		require.NoError(t, err)
		require.Equal(t, 1, snap.NewWords)
		teaching := map[string]bool{}
		for _, step := range snap.Steps {
			if step.Word.MeaningID == known.MeaningID {
				continue
			}
			require.Equal(t, "present", step.Word.WordText, "prefer a real new use at the normal teaching interval")
			require.False(t, familiar[step.Word.MeaningID])
			if step.Public.Kind == "teach" {
				teaching[step.Word.MeaningID] = true
				seen[step.Word.PartOfSpeech] = true
			} else {
				require.True(t, teaching[step.Word.MeaningID])
			}
		}
	}
	require.Equal(t, map[string]bool{"verb": true, "adjective": true}, seen)
	require.Len(t, familiar, 1, "building sessions cannot mark sibling uses as known")
}

func TestReviewedSupplementIdentityCopiesContextAndPractice(t *testing.T) {
	entries := meaninguses.Entries()
	require.Len(t, entries, 19)
	raw, err := os.ReadFile("../knowledgecatalog/catalog.json.gz")
	require.NoError(t, err)
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	require.NoError(t, err)
	defer reader.Close()
	var source struct {
		Items []struct {
			Word   string
			WordID string `json:"word_id"`
		}
	}
	require.NoError(t, json.NewDecoder(reader).Decode(&source))
	wordIDs := map[string]string{}
	for _, item := range source.Items {
		wordIDs[item.Word] = item.WordID
	}
	words, ids, examples := map[string]bool{}, map[string]bool{}, map[string]bool{}
	pool := seedWords(t)
	for _, entry := range entries {
		words[entry.Word] = true
		require.Equal(t, wordIDs[entry.Word], entry.WordID.String())
		require.False(t, ids[entry.MeaningID.String()])
		ids[entry.MeaningID.String()] = true
		require.False(t, examples[entry.ExampleID.String()])
		examples[entry.ExampleID.String()] = true
		key := strings.TrimPrefix(entry.SourceSenseKey, "vocanova:meaning-use-v1:"+entry.Word+":")
		require.Equal(t, uuid.NewSHA1(entry.WordID, []byte("vocanova:meaning-use-v1:"+key)), entry.MeaningID)
		require.Equal(t, uuid.NewSHA1(entry.MeaningID, []byte("vocanova:meaning-use-v1:example")), entry.ExampleID)
		require.True(t, SupportsMeaning(entry.MeaningID.String(), entry.Word, entry.Definition))
		require.False(t, meaninguses.SupportsMeaning(entry.MeaningID.String(), entry.Word, entry.Definition+" changed"))
		require.False(t, meaninguses.SupportsMeaning(entry.MeaningID.String(), "another-word", entry.Definition))
		context, ok := meaninguses.ReviewedContext(entry.MeaningID.String(), entry.Word, entry.Definition)
		require.True(t, ok)
		require.Equal(t, entry.Example, context)
		require.Contains(t, strings.ToLower(entry.Example), entry.Word)
		w := Word{MeaningID: entry.MeaningID.String(), WordText: entry.Word, PartOfSpeech: entry.PartOfSpeech, Definition: entry.Definition, Example: entry.Example}
		snap, err := buildJourney([]Word{w}, pool, map[string]bool{w.MeaningID: true}, 1, "reviewed-supplement")
		require.NoError(t, err, entry.SourceSenseKey)
		require.Len(t, snap.Steps, 3, entry.SourceSenseKey)
		require.Equal(t, "sentence_build", snap.Steps[2].Public.Kind)
		require.Equal(t, entry.Example, snap.Steps[2].Public.SpeechText)
	}
	require.Len(t, words, 12)
	entries[0].Definition = "Caller mutation must not change trusted data."
	require.NotEqual(t, entries[0].Definition, meaninguses.Entries()[0].Definition)
	context, ok := meaninguses.ReviewedContext("9e8813c5-81d2-584e-8197-882110c19402", "read", "To look at written words and understand what they mean.")
	require.True(t, ok)
	require.Equal(t, "I read a book every evening.", context)
	_, ok = meaninguses.ReviewedContext("9e8813c5-81d2-584e-8197-882110c19402", "read", "Changed definition")
	require.False(t, ok)
	require.ErrorContains(t, meaninguses.Seed(t.Context(), nil), "transaction required")
}

func TestPracticePronunciationRevisionUsesFrozenFullMeaning(t *testing.T) {
	word := Word{MeaningID: uuid.NewString(), WordText: "record", PartOfSpeech: "verb", Definition: strings.Repeat("A full meaning definition. ", 40)}
	st := State{Snapshot: Snapshot{Steps: []privateStep{{Word: word, Public: Step{ID: "frozen", Kind: "listening_choice", SpeechText: "record", PronunciationRevision: "untrusted-public-field"}}}}}
	before := project(st).CurrentStep.PronunciationRevision
	require.Equal(t, meaningidentity.Revision(word.MeaningID, word.WordText, word.PartOfSpeech, word.Definition), before)
	require.NotEmpty(t, before)
	require.NotEqual(t, meaningidentity.Revision(word.MeaningID, word.WordText, word.PartOfSpeech, word.Definition[:500]), before, "hash full definition, never a display/provider truncation")
	raw, err := json.Marshal(st)
	require.NoError(t, err)
	var resumed State
	require.NoError(t, json.Unmarshal(raw, &resumed))
	require.Equal(t, before, project(resumed).CurrentStep.PronunciationRevision)
	// The same spelling under a different sense or edited frozen metadata must
	// produce a distinct revision, without changing the original saved snapshot.
	changed := word
	changed.PartOfSpeech = "noun"
	require.NotEqual(t, before, meaningidentity.Revision(changed.MeaningID, changed.WordText, changed.PartOfSpeech, changed.Definition))
	changed = word
	changed.Definition += "An added distinction."
	require.NotEqual(t, before, meaningidentity.Revision(changed.MeaningID, changed.WordText, changed.PartOfSpeech, changed.Definition))
	require.Equal(t, before, project(st).CurrentStep.PronunciationRevision)
	st.Snapshot.Steps[0].Public.SpeechText = ""
	require.Empty(t, project(st).CurrentStep.PronunciationRevision, "do not expose pronunciation metadata on silent typed steps")
}
