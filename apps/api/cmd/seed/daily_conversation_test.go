package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const dailySituationID = "d9f0dfc7-bc05-5595-bbf2-035c4a61ca9c"

func TestDailyConversationCurriculum(t *testing.T) {
	seed, err := loadSeed()
	require.NoError(t, err)
	words := map[string]canonicalWord{}
	meanings := map[string]wordMeaning{}
	examples := map[string][]wordExample{}
	notes := map[string][]usageNote{}
	for _, w := range seed.CanonicalWords {
		words[w.ID] = w
	}
	for _, m := range seed.WordMeanings {
		meanings[m.ID] = m
	}
	for _, e := range seed.WordExamples {
		examples[e.MeaningID] = append(examples[e.MeaningID], e)
	}
	for _, n := range seed.UsageNotes {
		notes[n.MeaningID] = append(notes[n.MeaningID], n)
	}
	var journey []journeyWord
	for _, s := range seed.JourneySituations {
		if s.ID == dailySituationID {
			require.NotNil(t, s.LevelBand)
			require.Equal(t, "a2_b1", *s.LevelBand)
		}
	}
	for _, j := range seed.JourneyWords {
		if j.JourneySituationID == dailySituationID {
			require.NotNil(t, j.DisplayOrder)
			journey = append(journey, j)
		}
	}
	require.Len(t, journey, 18)
	sort.Slice(journey, func(i, j int) bool {
		if journey[i].IsCore != journey[j].IsCore {
			return journey[i].IsCore
		}
		return *journey[i].DisplayOrder < *journey[j].DisplayOrder
	})
	expected := []string{"greeting", "small talk", "casual", "weekend plans", "available", "invite", "join", "suggest", "sounds good", "arrange", "meet up", "confirm", "on time", "reschedule", "cancel", "catch up", "keep in touch", "farewell"}
	for i, j := range journey {
		require.Equal(t, i+1, *j.DisplayOrder)
		m, ok := meanings[j.MeaningID]
		require.True(t, ok)
		w, ok := words[m.WordID]
		require.True(t, ok)
		require.Equal(t, expected[i], w.Text)
		t.Run(w.Text, func(t *testing.T) {
			require.NotNil(t, m.LearnerDefinition)
			require.NotEmpty(t, strings.TrimSpace(*m.LearnerDefinition))
			require.NotEqual(t, strings.TrimSpace(m.ShortDefinition), strings.TrimSpace(*m.LearnerDefinition))
			require.NotNil(t, m.DifficultyLevel)
			require.Contains(t, []string{"a2", "b1"}, *m.DifficultyLevel)
			require.Equal(t, m.DifficultyLevel, w.DifficultyLevel)
			require.Len(t, examples[m.ID], 2)
			orders := []int{}
			seenText := map[string]bool{}
			for _, e := range examples[m.ID] {
				require.False(t, seenText[e.ExampleText], "duplicate example")
				seenText[e.ExampleText] = true
				require.GreaterOrEqual(t, len(strings.Fields(e.ExampleText)), 3)
				require.LessOrEqual(t, len([]rune(e.ExampleText)), 300)
				orders = append(orders, e.ExampleOrder)
			}
			require.ElementsMatch(t, []int{1, 2}, orders)
			require.Len(t, notes[m.ID], 3)
			types := []string{}
			for _, n := range notes[m.ID] {
				require.NotEmpty(t, n.NoteText)
				types = append(types, n.NoteType)
			}
			require.ElementsMatch(t, []string{"collocation", "register", "common_mistake"}, types)
		})
	}
}

func TestCanonicalSeedIdentityAndReferenceIntegrity(t *testing.T) {
	seed, err := loadSeed()
	require.NoError(t, err)
	seen := map[string]bool{}
	addID := func(id string) {
		_, err := uuid.Parse(id)
		require.NoError(t, err)
		require.False(t, seen[id], "duplicate ID %s", id)
		seen[id] = true
	}
	words := map[string]bool{}
	meanings := map[string]bool{}
	situations := map[string]bool{}
	unique := map[string]bool{}
	addKey := func(key string) { require.False(t, unique[key], "duplicate database key %s", key); unique[key] = true }
	for _, s := range seed.JourneySituations {
		addID(s.ID)
		situations[s.ID] = true
		addKey("situation/" + s.Slug)
	}
	for _, w := range seed.CanonicalWords {
		addID(w.ID)
		words[w.ID] = true
		addKey("word/" + w.LanguageCode + "/" + w.NormalizedText)
		require.Equal(t, strings.ToLower(strings.TrimSpace(w.Text)), w.NormalizedText)
	}
	for _, m := range seed.WordMeanings {
		addID(m.ID)
		meanings[m.ID] = true
		require.True(t, words[m.WordID])
		addKey(fmt.Sprintf("meaning/%s/%d", m.WordID, m.MeaningOrder))
	}
	for _, e := range seed.WordExamples {
		addID(e.ID)
		require.True(t, meanings[e.MeaningID])
		addKey(fmt.Sprintf("example/%s/%d", e.MeaningID, e.ExampleOrder))
	}
	for _, n := range seed.UsageNotes {
		addID(n.ID)
		require.True(t, meanings[n.MeaningID])
		addKey(fmt.Sprintf("note/%s/%d", n.MeaningID, n.NoteOrder))
	}
	for _, j := range seed.JourneyWords {
		addID(j.ID)
		require.True(t, meanings[j.MeaningID])
		require.True(t, situations[j.JourneySituationID])
		addKey("journey/" + j.JourneySituationID + "/" + j.MeaningID)
	}
}

func TestDailyConversationStableAdditionIDs(t *testing.T) {
	seed, err := loadSeed()
	require.NoError(t, err)
	// Frozen semantic keys: wording and display-order edits must not replace identities.
	senses := map[string]string{
		"invite": "social-invitation", "join": "participate-with-others", "available": "free-at-a-time", "suggest": "propose-an-activity", "arrange": "plan-a-social-event", "confirm": "make-a-plan-definite", "reschedule": "change-event-time", "cancel": "stop-a-planned-event", "on-time": "punctual", "sounds-good": "accept-a-suggestion", "meet-up": "meet-socially", "keep-in-touch": "continue-communicating",
	}
	id := func(key string) string {
		return uuid.NewSHA1(uuid.MustParse(dailySituationID), []byte("daily-conversation-plans-v1/"+key)).String()
	}
	words := map[string]canonicalWord{}
	meanings := map[string]wordMeaning{}
	examples := map[string]wordExample{}
	notes := map[string]usageNote{}
	journey := map[string]journeyWord{}
	for _, w := range seed.CanonicalWords {
		words[w.ID] = w
	}
	for _, m := range seed.WordMeanings {
		meanings[m.ID] = m
	}
	for _, e := range seed.WordExamples {
		examples[e.ID] = e
	}
	for _, n := range seed.UsageNotes {
		notes[n.ID] = n
	}
	for _, j := range seed.JourneyWords {
		journey[j.ID] = j
	}
	for slug, sense := range senses {
		t.Run(slug, func(t *testing.T) {
			wid, mid := id("word/"+slug), id("meaning/"+slug+"/"+sense)
			require.Equal(t, strings.ReplaceAll(slug, "-", " "), words[wid].NormalizedText)
			require.Equal(t, wid, meanings[mid].WordID)
			for i := 1; i <= 2; i++ {
				require.Equal(t, mid, examples[id(fmt.Sprintf("example/%s/%s/%d", slug, sense, i))].MeaningID)
			}
			for _, typ := range []string{"collocation", "register", "common-mistake"} {
				require.Equal(t, mid, notes[id("note/"+slug+"/"+sense+"/"+typ)].MeaningID)
			}
			require.Equal(t, mid, journey[id("journey/"+slug+"/"+sense)].MeaningID)
		})
	}
	retained := map[string]string{
		"small talk": "3265c02c-5751-5465-b94b-e767aa871d8d", "catch up": "3d64c3c9-ede0-5ffd-b1ef-278f6b70e486", "weekend plans": "c99075bf-94cb-5f51-90d5-f244f66ea942", "greeting": "ac93d068-7c3e-55c9-9d85-3fd3518592dc", "farewell": "c954cdb6-c0bc-5d61-9045-18db360f4b84", "casual": "0d267f75-f227-5478-962d-b4b944e0e93e",
	}
	for text, mid := range retained {
		require.Equal(t, text, words[meanings[mid].WordID].Text)
		require.Equal(t, mid, examples[id("example/retained/"+mid+"/2")].MeaningID)
	}
}
