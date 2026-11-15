package dailylesson

import (
	"bytes"
	"crypto/sha256"
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
	// Exhaust every starting index and private destination without relying on
	// statistical randomness or any public question identifier.
	destinations := []int{0, 1, 2, 3}
	for initial := 0; initial < 4; initial++ {
		for _, slot := range destinations {
			t.Run(fmt.Sprintf("from_%d_to_%d", initial, slot), func(t *testing.T) {
				choices := []Choice{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
				answer := choices[initial].ID
				positionChoices(choices, answer, slot)
				require.Equal(t, answer, choices[slot].ID)
				replay := []Choice{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
				positionChoices(replay, answer, slot)
				require.Equal(t, choices, replay)
				require.ElementsMatch(t, []string{"a", "b", "c", "d"}, []string{choices[0].ID, choices[1].ID, choices[2].ID, choices[3].ID})
			})
		}
	}
}

func TestChoicePlacementUsesPrivateEntropyForEveryKind(t *testing.T) {
	const publicID = "same-public-step-id"
	publicHash := sha256.Sum256([]byte(publicID))
	hashSlot := int(publicHash[0]) % 4
	for _, kind := range []string{"meaning_choice", "picture_choice", "gap_choice"} {
		for slot := 0; slot < 4; slot++ {
			t.Run(fmt.Sprintf("%s_private_%d", kind, slot), func(t *testing.T) {
				w := word{AnswerID: "b", Public: Step{ID: publicID, Kind: kind, Choices: []Choice{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}}}
				require.NoError(t, freezeChoiceOrder(&w, bytes.NewReader([]byte{byte(slot)})))
				require.Equal(t, "b", w.Public.Choices[slot].ID)
				if slot != hashSlot {
					require.NotEqual(t, "b", w.Public.Choices[hashSlot].ID, "public step hash cannot determine the answer")
				}
				raw, err := json.Marshal(state{Words: []word{w}, Queue: []int{0}})
				require.NoError(t, err)
				var reloaded state
				require.NoError(t, json.Unmarshal(raw, &reloaded))
				require.Equal(t, w.Public.Choices, project(&reloaded).CurrentStep.Choices, "reload uses persisted order without drawing again")
			})
		}
	}
}

func TestChoicePlacementEntropyFailureDoesNotPublishFallbackOrder(t *testing.T) {
	w := word{AnswerID: "b", Public: Step{ID: "public", Choices: []Choice{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}}}
	before := append([]Choice{}, w.Public.Choices...)
	require.Error(t, freezeChoiceOrder(&w, bytes.NewReader(nil)))
	require.Equal(t, before, w.Public.Choices)
}
