package lessons

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type seedData struct {
	Situations []struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	} `json:"journey_situations"`
	Words []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"canonical_words"`
	Meanings []struct {
		ID         string `json:"id"`
		WordID     string `json:"word_id"`
		Definition string `json:"short_definition"`
		POS        string `json:"part_of_speech"`
	} `json:"word_meanings"`
	Examples []struct {
		MeaningID string `json:"meaning_id"`
		Text      string `json:"example_text"`
	} `json:"word_examples"`
	Links []struct {
		SituationID string `json:"journey_situation_id"`
		MeaningID   string `json:"meaning_id"`
	} `json:"journey_words"`
}

func seedWords(t *testing.T, d Definition) []Word {
	t.Helper()
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	var seed seedData
	require.NoError(t, json.Unmarshal(raw, &seed))
	var situationID string
	for _, s := range seed.Situations {
		if s.Slug == d.SituationSlug {
			situationID = s.ID
		}
	}
	require.NotEmpty(t, situationID)
	words := []Word{}
	for _, ref := range d.Words {
		require.NotEqual(t, uuid.Nil, uuid.MustParse(ref.MeaningID))
		w := Word{MeaningID: ref.MeaningID}
		for _, m := range seed.Meanings {
			if m.ID == ref.MeaningID {
				w.Definition = m.Definition
				w.PartOfSpeech = m.POS
				for _, cw := range seed.Words {
					if cw.ID == m.WordID {
						w.WordText = cw.Text
					}
				}
			}
		}
		require.Equal(t, ref.WordText, w.WordText)
		w.WordSlug = strings.ReplaceAll(strings.ToLower(w.WordText), " ", "-")
		for _, ex := range seed.Examples {
			if ex.MeaningID == ref.MeaningID {
				w.Example = ex.Text
				break
			}
		}
		linked := false
		for _, link := range seed.Links {
			if link.SituationID == situationID && link.MeaningID == ref.MeaningID {
				linked = true
			}
		}
		require.True(t, linked, "lesson target belongs to its actual canonical situation")
		require.NotEmpty(t, w.Definition)
		require.NotEmpty(t, w.Example)
		words = append(words, w)
	}
	return words
}
func TestCatalogCoversThirtyLessonsAcrossCanonicalSituations(t *testing.T) {
	require.Len(t, catalog, 30)
	keys := map[string]bool{}
	for _, d := range catalog {
		t.Run(d.Key, func(t *testing.T) {
			require.False(t, keys[d.Key])
			keys[d.Key] = true
			require.Len(t, d.Words, 3)
			snap, err := buildSnapshot(d, seedWords(t, d), "fixed-session")
			require.NoError(t, err)
			require.Len(t, snap.Steps, 9)
			for i, step := range snap.Steps {
				require.NotEmpty(t, step.Step.Prompt)
				if i < 3 {
					require.Equal(t, "teach", step.Step.Kind)
					require.Empty(t, step.Step.Choices)
					continue
				}
				require.Len(t, step.Step.Choices, 3)
				seen := map[string]bool{}
				correct := 0
				for _, option := range step.Step.Choices {
					require.False(t, seen[option.Text], "distinct answer text")
					seen[option.Text] = true
					require.NotEmpty(t, step.Explanations[option.ID])
					if option.ID == step.CorrectChoiceID {
						correct++
					}
				}
				require.Equal(t, 1, correct)
				if i >= 6 {
					require.NotEmpty(t, step.Step.Context)
				}
			}
		})
	}
	for _, key := range []string{"airport", "restaurant", "hotel-check-in", "job-interview", "daily-conversation", "work-meeting", "university-class"} {
		require.True(t, keys[key])
	}
}
func TestLessonSequencePersistsHonestFirstAnswersAndRequiresCorrection(t *testing.T) {
	d := catalog[0]
	snap, err := buildSnapshot(d, seedWords(t, d), "session")
	require.NoError(t, err)
	st := State{ID: uuid.New(), UserID: uuid.New(), Snapshot: snap}
	now := time.Now().UTC()
	for st.Index < 9 {
		step := snap.Steps[st.Index]
		action := Action{StepID: step.Step.ID, ExpectedRevision: st.Revision, Action: "continue"}
		if step.Step.Kind != "teach" {
			require.ErrorIs(t, apply(&st, action, now), ErrConflict)
			if st.Index == 3 {
				wrong := ""
				for _, c := range step.Step.Choices {
					if c.ID != step.CorrectChoiceID {
						wrong = c.ID
						break
					}
				}
				require.NoError(t, apply(&st, Action{StepID: step.Step.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: wrong}, now))
				require.False(t, st.Feedback.Correct)
				require.Equal(t, 3, st.Index)
				require.False(t, project(st).CanContinue)
			}
			require.NoError(t, apply(&st, Action{StepID: step.Step.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: step.CorrectChoiceID}, now))
			require.True(t, project(st).CanContinue)
			require.Nil(t, st.CompletedAt, "answer alone must not complete even the final step")
			action.ExpectedRevision = st.Revision
		}
		require.NoError(t, apply(&st, action, now))
	}
	p := project(st)
	require.Equal(t, "completed", p.Status)
	require.Equal(t, 6, p.QuestionsAnswered)
	require.Equal(t, 5, p.FirstAnswersCorrect)
	require.Equal(t, 9, p.CompletedSteps)
	require.Nil(t, p.CurrentStep)
	require.False(t, p.CanContinue)
	require.Equal(t, now, *p.CompletedAt)
	require.ErrorIs(t, apply(&st, Action{ExpectedRevision: st.Revision, Action: "continue"}, now), ErrConflict)
}
func TestLessonRejectsStaleOutOfOrderAndUnknownChoicesWithoutMutation(t *testing.T) {
	snap, err := buildSnapshot(catalog[0], seedWords(t, catalog[0]), "one")
	require.NoError(t, err)
	for _, a := range []Action{
		{StepID: "teach-1", ExpectedRevision: 1, Action: "continue"},
		{StepID: "teach-2", ExpectedRevision: 0, Action: "continue"},
		{StepID: "teach-1", ExpectedRevision: 0, Action: "answer", ChoiceID: "forged"},
		{StepID: "teach-1", ExpectedRevision: 0, Action: "continue", ChoiceID: "extra"},
	} {
		st := State{Snapshot: snap}
		before, _ := json.Marshal(st)
		require.Error(t, apply(&st, a, time.Now()))
		after, _ := json.Marshal(st)
		require.Equal(t, string(before), string(after))
	}
	st := State{Snapshot: snap, Index: 3}
	require.ErrorIs(t, apply(&st, Action{StepID: "recall-1", Action: "answer", ChoiceID: "unknown"}, time.Now()), ErrInvalid)
	require.Zero(t, st.Revision)
	raw, err := json.Marshal(project(st))
	require.NoError(t, err)
	require.NotContains(t, string(raw), "correctChoiceId")
	require.NotContains(t, string(raw), "explanations")
}
func TestLessonChoiceOrderStableWithinSessionVariesBetweenSessions(t *testing.T) {
	d := catalog[0]
	words := seedWords(t, d)
	first, err := buildSnapshot(d, words, "first")
	require.NoError(t, err)
	again, err := buildSnapshot(d, words, "first")
	require.NoError(t, err)
	require.Equal(t, first, again)
	positions := map[int]bool{}
	for i := 0; i < 32; i++ {
		snap, err := buildSnapshot(d, words, string(rune('a'+i)))
		require.NoError(t, err)
		for j, c := range snap.Steps[3].Step.Choices {
			if c.ID == snap.Steps[3].CorrectChoiceID {
				positions[j] = true
			}
		}
	}
	require.Len(t, positions, 3)
	words[0].WordText = "another sense"
	_, err = buildSnapshot(d, words, "test")
	require.ErrorIs(t, err, ErrContentUnavailable)
}
