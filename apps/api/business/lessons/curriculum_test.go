package lessons

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStarterCourseOrderAndDistinctLessonSituationKeys(t *testing.T) {
	expected := []string{"conversation-start", "conversation-find-time", "conversation-suggest-activity", "daily-conversation", "conversation-plan-changes", "conversation-stay-connected", "restaurant", "restaurant-other-plans", "shopping-clothes", "shopping-returns", "home-rental", "transport-station", "transport-tickets", "airport", "airport-checks", "hotel-check-in", "hotel-services", "delivery-parcel", "money-payment", "services-form", "health-services", "phone-calls", "digital-email", "work-day", "job-interview", "interview-application", "work-meeting", "meeting-collaboration", "university-class", "study-collaboration"}
	keys := []string{}
	meanings := map[string]bool{}
	situations := map[string]int{}
	for _, d := range catalog {
		keys = append(keys, d.Key)
		situations[d.SituationSlug]++
		for _, r := range d.Words {
			require.False(t, meanings[r.MeaningID], "no duplicate target meaning across course")
			meanings[r.MeaningID] = true
		}
	}
	require.Equal(t, expected, keys)
	require.Len(t, meanings, 90)
	require.Len(t, situations, 17)
	require.Equal(t, 6, situations["daily-conversation"])
	require.Equal(t, 2, situations["shopping"])
	for _, key := range []string{"conversation-start", "conversation-find-time", "conversation-suggest-activity", "conversation-plan-changes", "conversation-stay-connected"} {
		d, ok := definition(key)
		require.True(t, ok)
		require.Equal(t, "daily-conversation", d.SituationSlug)
		require.NotEqual(t, d.Key, d.SituationSlug)
	}
}

func TestOriginalSevenLessonDefinitionsRemainFrozen(t *testing.T) {
	// Serialized original definitions, captured before course expansion. This
	// protects their keys, version, selected meanings and authored contexts;
	// intentional future editorial changes must explicitly version the lesson.
	expected := map[string]string{
		"airport":            "b4fbd17cf0a22f1f7379c04938293bfadcb7b2f87250600457264937d9ef13ad",
		"restaurant":         "3468b874d04deef3bbd54f7e0b503b7912489db621e597f40676b99c58cfcc58",
		"hotel-check-in":     "86b5920f2f323b0777c6e64ee7f975610c4c74f3a62787e63541aff8f6b45742",
		"job-interview":      "a8564d35475637776def18250a32bb9c9176edd664582fe51315b6d64a9a857c",
		"daily-conversation": "864bbac0b2f384db38382de893b591e34dd006fd0ccfc776a82da0b1c8ced4e1",
		"work-meeting":       "6648074c209eacff19cc624d40e4fa6a2d88a5dde2389e5e955db3941f808836",
		"university-class":   "058305d51992f5c1759d4bb020c76e669a711484f449f9c266fb5b6990ef1900",
	}
	for key, hash := range expected {
		d, ok := definition(key)
		require.True(t, ok)
		raw, err := json.Marshal(d)
		require.NoError(t, err)
		require.Equal(t, hash, fmt.Sprintf("%x", sha256.Sum256(raw)), key)
	}
}
