package practice

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExpandedPracticeTripletsHaveUniqueMeaningIdentities(t *testing.T) {
	require.Len(t, catalog, 90)
	groups := map[string]int{}
	meanings := map[string]bool{}
	for _, r := range catalog {
		require.False(t, meanings[r.MeaningID])
		meanings[r.MeaningID] = true
		groups[r.LessonKey]++
	}
	require.Len(t, groups, 30)
	for key, count := range groups {
		require.Equal(t, 3, count, key)
	}
}

func TestExpandedPracticeExplicitNounSpellingsRemainLiteral(t *testing.T) {
	words := seedWords(t)
	for _, tc := range []struct {
		word               string
		accepted, rejected []string
	}{
		{"take-out", []string{"take-out", "takeout", " TAKEOUT "}, []string{"take out", "takeouts", "takeout!", "takeaway"}},
		{"check-out", []string{"check-out", "checkout"}, []string{"check out", "checkout!", "checkouts"}},
		{"follow-up", []string{"follow-up", "followup"}, []string{"follow up", "follow-ups", "followups!"}},
		{"voicemail", []string{"voicemail", "voice mail"}, []string{"voicemails", "voice-mail", "voicemail!"}},
	} {
		t.Run(tc.word, func(t *testing.T) {
			var word Word
			for _, w := range words {
				if w.WordText == tc.word {
					word = w
				}
			}
			snap, err := build(StartRequest{Mode: "typed_recall", LessonKey: word.LessonKey}, words, nil, "aliases")
			require.NoError(t, err)
			var step privateStep
			for _, s := range snap.Steps {
				if s.Word.WordText == tc.word {
					step = s
				}
			}
			require.NotEmpty(t, step.Public.ID)
			for _, group := range []struct {
				answers []string
				correct bool
			}{{tc.accepted, true}, {tc.rejected, false}} {
				for _, answer := range group.answers {
					st := State{Snapshot: Snapshot{Steps: []privateStep{step}}}
					_, err := apply(&st, Action{StepID: step.Public.ID, Action: "answer", TypedAnswer: answer}, time.Now())
					require.NoError(t, err)
					require.Equal(t, group.correct, st.Feedback.Correct, answer)
				}
			}
		})
	}
}

func TestPracticePersistedContentVersionIsNotRewrittenByExpansion(t *testing.T) {
	st := State{Snapshot: Snapshot{Mode: "typed_recall", ContentVersion: "starter-21-v1", GradingVersion: "exact-recall-v1", Steps: []privateStep{{Public: Step{ID: "old-step", Kind: "typed_recall", Prompt: "An older saved prompt", Choices: []Choice{}}}}}}
	projected := project(st)
	require.Equal(t, "starter-21-v1", projected.ContentVersion)
	require.Equal(t, "old-step", projected.CurrentStep.ID)
	require.Equal(t, "An older saved prompt", projected.CurrentStep.Prompt)
	require.Equal(t, "starter-90-v2", ContentVersion)
}
