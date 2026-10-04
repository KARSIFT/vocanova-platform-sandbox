package stories

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"
)

func TestOriginalStoryCatalogEditorialAndMeaningAlignment(t *testing.T) {
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	var seed struct {
		Words    []struct{ ID, Text string } `json:"canonical_words"`
		Meanings []struct {
			ID     string
			WordID string `json:"word_id"`
		} `json:"word_meanings"`
	}
	require.NoError(t, json.Unmarshal(raw, &seed))
	words := map[string]string{}
	meanings := map[string]string{}
	for _, w := range seed.Words {
		words[w.ID] = w.Text
	}
	for _, m := range seed.Meanings {
		meanings[m.ID] = words[m.WordID]
	}
	require.Len(t, catalog, 6)
	situations := map[string]bool{}
	keys := map[string]bool{}
	for _, snap := range catalog {
		require.False(t, keys[snap.Story.Key])
		keys[snap.Story.Key] = true
		situations[snap.Story.Situation] = true
		require.Equal(t, 8, snap.Story.LineCount)
		require.Equal(t, 2, snap.Story.QuestionCount)
		require.Len(t, snap.Steps, 10)
		require.Len(t, snap.Story.Vocabulary, 3)
		for _, v := range snap.Story.Vocabulary {
			require.Equal(t, v.WordText, meanings[v.MeaningID])
			require.NotEmpty(t, v.Definition)
			require.NotEmpty(t, v.WordSlug)
		}
		ids := map[string]bool{}
		kinds := map[string]bool{}
		lineCount := 0
		for _, step := range snap.Steps {
			require.False(t, ids[step.Public.ID])
			ids[step.Public.ID] = true
			kinds[step.Public.Kind] = true
			if step.Public.Kind == "line" {
				lineCount++
				require.NotEmpty(t, step.Public.Line.Speaker)
				require.Greater(t, len(strings.Fields(step.Public.Line.Text)), 4)
				continue
			}
			require.GreaterOrEqual(t, lineCount, 4, "question must follow meaningful context")
			require.Len(t, step.Public.Choices, 3)
			require.NotEmpty(t, step.Explanation)
			correct := 0
			seen := map[string]bool{}
			for _, c := range step.Public.Choices {
				require.False(t, seen[c.Text])
				seen[c.Text] = true
				if c.ID == step.CorrectChoice {
					correct++
				}
			}
			require.Equal(t, 1, correct)
		}
		require.True(t, kinds["comprehension"])
		require.True(t, kinds["phrase_completion"])
	}
	require.Len(t, situations, 6)
}
func TestStoryProgressionWrongRetryResumeCompletionAndNoAnswerLeak(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	for _, catalogStory := range catalog {
		t.Run(catalogStory.Story.Key, func(t *testing.T) {
			snap, err := build(catalogStory.Story.Key)
			require.NoError(t, err)
			st := State{ID: uuid.New(), UserID: uuid.New(), Snapshot: snap, CreatedAt: now, UpdatedAt: now}
			for st.CompletedAt == nil {
				p := project(st)
				step := snap.Steps[st.Index]
				raw, err := json.Marshal(p)
				require.NoError(t, err)
				require.NotContains(t, string(raw), "CorrectChoice")
				require.NotContains(t, string(raw), "Explanation")
				if step.Public.Kind == "line" {
					require.True(t, p.CanContinue)
					require.ErrorIs(t, apply(&st, StoryAction{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: "choice-1"}, now), ErrInvalid)
				} else {
					require.False(t, p.CanContinue)
					require.ErrorIs(t, apply(&st, StoryAction{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "continue"}, now), ErrConflict)
					wrong := ""
					for _, c := range step.Public.Choices {
						if c.ID != step.CorrectChoice {
							wrong = c.ID
							break
						}
					}
					require.NoError(t, apply(&st, StoryAction{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: wrong}, now))
					require.False(t, project(st).CanContinue)
					require.ErrorIs(t, apply(&st, StoryAction{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "continue"}, now), ErrConflict)
					raw, err = json.Marshal(st)
					require.NoError(t, err)
					require.NoError(t, json.Unmarshal(raw, &st), "resume preserves wrong-answer first attempt")
					require.NoError(t, apply(&st, StoryAction{StepID: step.Public.ID, ExpectedRevision: st.Revision, Action: "answer", ChoiceID: step.CorrectChoice}, now))
					require.True(t, project(st).CanContinue)
					require.Zero(t, st.FirstAnswersCorrect)
				}
				revision := st.Revision
				require.NoError(t, apply(&st, StoryAction{StepID: step.Public.ID, ExpectedRevision: revision, Action: "continue"}, now))
				require.ErrorIs(t, apply(&st, StoryAction{StepID: step.Public.ID, ExpectedRevision: revision, Action: "continue"}, now), ErrConflict)
			}
			p := project(st)
			require.Equal(t, "completed", p.Status)
			require.Len(t, p.VisibleLines, 8)
			require.Nil(t, p.CurrentStep)
			require.Nil(t, p.Feedback)
			require.False(t, p.CanContinue)
			require.Equal(t, 2, p.QuestionsAnswered)
			require.Zero(t, p.FirstAnswersCorrect)
			require.Equal(t, 10, p.CompletedSteps)
		})
	}
}
func TestStorySnapshotPreservesHistoricalContent(t *testing.T) {
	snap, err := build("a-quiet-lunch")
	require.NoError(t, err)
	original := snap.Steps[0].Public.Line.Text
	snap.Steps[0].Public.Line.Text = "Old saved wording stays in its snapshot."
	fresh, err := build("a-quiet-lunch")
	require.NoError(t, err)
	require.Equal(t, original, fresh.Steps[0].Public.Line.Text)
	st := State{Snapshot: snap}
	require.Equal(t, "Old saved wording stays in its snapshot.", project(st).VisibleLines[0].Text)
	require.ErrorIs(t, apply(&st, StoryAction{StepID: "another-step", Action: "continue"}, time.Now()), ErrConflict)
	require.ErrorIs(t, apply(&st, StoryAction{StepID: snap.Steps[0].Public.ID, ExpectedRevision: 1, Action: "continue"}, time.Now()), ErrConflict)
}
