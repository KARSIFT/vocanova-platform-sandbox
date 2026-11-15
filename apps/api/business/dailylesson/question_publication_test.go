package dailylesson

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/stretchr/testify/require"
)

func TestPictureProjectionDoesNotPublishTargetIdentityOutsideChoices(t *testing.T) {
	const target = "0b593408-6679-5fb8-a21c-b73a525c76e5"
	const definition = "An object or group of objects wrapped for sending or delivery."
	revision := meaningidentity.Revision(target, "parcel", "noun", definition)
	st := &state{Words: []word{{MeaningID: target, Definition: definition, Public: Step{Kind: "picture_choice", WordText: "parcel", WordSlug: "parcel", PartOfSpeech: "noun", PronunciationMeaningID: target, PronunciationRevision: revision, Choices: []Choice{{ID: "opaque-a", Text: definition, Image: &meaningmedia.Image{MeaningID: target, MeaningRevision: revision}}}}}}, Queue: []int{0}}
	out := project(st)
	require.Equal(t, "parcel", out.CurrentStep.WordText)
	require.Equal(t, "parcel", out.CurrentStep.WordSlug)
	raw, err := json.Marshal(out)
	require.NoError(t, err)
	var public map[string]any
	require.NoError(t, json.Unmarshal(raw, &public))
	step := public["currentStep"].(map[string]any)
	choices := step["choices"].([]any)
	require.Equal(t, target, choices[0].(map[string]any)["image"].(map[string]any)["meaningId"])
	delete(step, "choices") // Every choice has equivalent image metadata; no target marker may appear elsewhere.
	raw, err = json.Marshal(public)
	require.NoError(t, err)
	require.NotContains(t, string(raw), target)
	require.NotContains(t, string(raw), revision)
	require.NotContains(t, string(raw), "pronunciationMeaningId")
	require.NotContains(t, string(raw), "pronunciationRevision")
}

func TestChoicePlacementSupportsEveryAnswerStartingIndexAndDestination(t *testing.T) {
	// Fixed step IDs exercise each hash bucket. This is an exhaustive permutation
	// property for four starting indices, not a randomized distribution assertion.
	seeds := []string{"slot-9", "slot-0", "slot-1", "slot-11"}
	destinations := []int{0, 1, 2, 3}
	for initial := 0; initial < 4; initial++ {
		for n, seed := range seeds {
			t.Run(fmt.Sprintf("from_%d_%s", initial, seed), func(t *testing.T) {
				choices := []Choice{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
				answer := choices[initial].ID
				positionChoices(choices, answer, seed)
				require.Equal(t, answer, choices[destinations[n]].ID)
				replay := []Choice{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
				positionChoices(replay, answer, seed)
				require.Equal(t, choices, replay)
				require.ElementsMatch(t, []string{"a", "b", "c", "d"}, []string{choices[0].ID, choices[1].ID, choices[2].ID, choices[3].ID})
			})
		}
	}
}
