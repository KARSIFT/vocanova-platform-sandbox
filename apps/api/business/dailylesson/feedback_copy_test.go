package dailylesson

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func gapWord(t *testing.T, meaningID, text string) *word {
	t.Helper()
	cloze, ok := mixedCloze(gapSets[meaningID].Example, text)
	require.True(t, ok)
	return &word{MeaningID: meaningID, WordText: text, Public: Step{Kind: "gap_choice", Context: cloze}}
}

func TestFeedbackExplanationRestoresOnlyTheFrozenVettedSentence(t *testing.T) {
	gate := gapWord(t, "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "gate")
	sentence, ok := vettedSentence(gate)
	require.True(t, ok)
	require.Equal(t, "Our flight leaves from gate twelve.", sentence)
	for _, outcome := range []struct{ correct, assisted bool }{{true, false}, {false, false}, {false, true}} {
		text := feedbackExplanation(gate, outcome.correct, outcome.assisted)
		require.Contains(t, text, "“Our flight leaves from gate twelve.”")
		require.Equal(t, text, feedbackExplanation(gate, outcome.correct, outcome.assisted), "copy is deterministic for reloads")
	}

	// A different frozen context, a non-gap kind, or an unknown meaning never
	// borrows the authored sentence.
	changed := *gate
	changed.Public.Context = "Our train leaves from _____ three."
	_, ok = vettedSentence(&changed)
	require.False(t, ok)
	require.NotContains(t, feedbackExplanation(&changed, false, false), "twelve")
	meaning := *gate
	meaning.Public.Kind = "meaning_choice"
	_, ok = vettedSentence(&meaning)
	require.False(t, ok)
	legacy := *gate
	legacy.WordText = ""
	_, ok = vettedSentence(&legacy)
	require.False(t, ok, "historical snapshots without a frozen headword stay generic")
}

func TestFeedbackExplanationIsPlainAndHonestPerOutcome(t *testing.T) {
	for _, kind := range []string{"meaning_choice", "picture_choice", "gap_choice"} {
		w := &word{MeaningID: "unknown", WordText: "harbor", Public: Step{Kind: kind}}
		correct := feedbackExplanation(w, true, false)
		require.True(t, strings.HasPrefix(correct, "Yes"), kind)
		require.Contains(t, correct, "reviewed this word for today", kind)
		miss := feedbackExplanation(w, false, false)
		require.Contains(t, miss, "will return on another day", kind)
		require.NotContains(t, miss, "Yes", kind)
		assisted := feedbackExplanation(w, false, true)
		require.Contains(t, assisted, "not a remembered answer", kind)
		require.Contains(t, assisted, "try it on your own", kind)
		// Assisted copy never sounds like a correct grade, even if misused.
		require.Equal(t, assisted, feedbackExplanation(w, true, true), kind)
		for _, text := range []string{correct, miss, assisted} {
			require.NotContains(t, text, "harbor", "copy never repeats the hidden headword")
		}
	}
}
