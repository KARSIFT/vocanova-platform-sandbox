package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var starterSenses = map[string]string{
	"size": "clothing-measurement", "try-on": "test-clothing", "fit": "right-clothing-size", "receipt": "proof-of-payment", "refund": "returned-money", "exchange": "replace-purchase",
	"rent": "housing-payment", "landlord": "property-provider", "repair": "fixing-work", "bus-stop": "bus-passenger-place", "timetable": "transport-times", "platform": "train-passenger-place",
	"single-ticket": "one-direction-travel", "return-ticket": "there-and-back-travel", "change": "transfer-transport", "parcel": "wrapped-delivery", "address": "location-details", "delivery": "bringing-goods",
	"cash": "coins-and-notes", "bank-card": "payment-card", "payment": "money-paid", "form": "information-document", "signature": "signed-name", "queue": "waiting-line",
	"appointment": "arranged-visit", "symptom": "health-sign", "pharmacy": "medicine-service", "signal": "mobile-connection", "voicemail": "recorded-phone-message", "call-back": "return-phone-call",
	"attachment": "message-file", "link": "online-resource-reference", "reply": "answer-message", "shift": "work-period", "break": "rest-from-work", "colleague": "fellow-worker", "update": "progress-news", "presentation": "audience-talk",
}

func starterID(key string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("vocanova/starter-30-v1/"+key)).String()
}

func TestStarterCurriculumStableIdentitiesAndHistoricalInventory(t *testing.T) {
	seed, err := loadSeed()
	require.NoError(t, err)
	var tables map[string][]struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(seedJSON, &tables))
	newIDs := map[string]bool{}
	words := map[string]canonicalWord{}
	meanings := map[string]wordMeaning{}
	examples := map[string]wordExample{}
	notes := map[string]usageNote{}
	for _, v := range seed.CanonicalWords {
		words[v.ID] = v
	}
	for _, v := range seed.WordMeanings {
		meanings[v.ID] = v
	}
	for _, v := range seed.WordExamples {
		examples[v.ID] = v
	}
	for _, v := range seed.UsageNotes {
		notes[v.ID] = v
	}
	for _, slug := range []string{"shopping", "home-and-renting", "public-transport", "deliveries", "everyday-payments", "everyday-services", "health-appointments", "phone-calls", "email-and-online-tasks", "everyday-work"} {
		newIDs[starterID("situation/"+slug)] = true
	}
	for slug, sense := range starterSenses {
		wid, mid := starterID("word/"+slug), starterID("meaning/"+slug+"/"+sense)
		newIDs[wid] = true
		newIDs[mid] = true
		require.Equal(t, strings.ReplaceAll(slug, "-", " "), words[wid].Text)
		require.Equal(t, wid, meanings[mid].WordID)
		require.NotEmpty(t, meanings[mid].ShortDefinition)
		require.Nil(t, words[wid].FrequencyRank, "no unsupported frequency claim")
		for i := 1; i <= 2; i++ {
			eid := starterID(fmt.Sprintf("example/%s/%s/%d", slug, sense, i))
			newIDs[eid] = true
			require.Equal(t, mid, examples[eid].MeaningID)
			require.GreaterOrEqual(t, len(strings.Fields(examples[eid].ExampleText)), 3)
		}
		nid := starterID("note/" + slug + "/" + sense + "/usage")
		newIDs[nid] = true
		require.Equal(t, mid, notes[nid].MeaningID)
		links := 0
		for _, j := range seed.JourneyWords {
			if j.MeaningID == mid {
				links++
				for _, s := range seed.JourneySituations {
					if s.ID == j.JourneySituationID {
						require.Equal(t, starterID("journey/"+s.Slug+"/"+slug+"/"+sense), j.ID)
						newIDs[j.ID] = true
					}
				}
			}
		}
		require.Equal(t, 1, links, "one new canonical situation link for %s", slug)
	}
	// Frozen hashes of all pre-expansion IDs; no historical identity is discarded,
	// including meanings deliberately left outside the course. Order is irrelevant.
	baseline := map[string]struct {
		count int
		hash  string
	}{
		"journey_situations": {7, "b635135d151e74b2dfd9d1527d72189b577bea596a31dde1964e16785c221c9c"},
		"canonical_words":    {51, "e3d73f05f733441bf3b8607ea16cbc9bf8c728d4de8a51eecfa72c93693a11fa"},
		"word_meanings":      {54, "5fe3547efbe120e78e321afb63f8cfdfde6575082254fea964d3e72c8524c12c"},
		"word_examples":      {72, "beaff971a5afd9ee169fb9a8350e6facb900a47d8a50df64f43a490d32a59f19"},
		"usage_notes":        {162, "3e40d72cbbe6f4d67ab3985b6c2f7da18b755cbc554e588419e9792ba217f9f5"},
		"journey_words":      {54, "9d31423b931a74240c12c6c02f66fb5ff04f4d3fa1e58db88b4568792f1c3b55"},
	}
	foundNew := 0
	for table, rows := range tables {
		retained := []string{}
		for _, row := range rows {
			if newIDs[row.ID] {
				foundNew++
			} else {
				retained = append(retained, row.ID)
			}
		}
		sort.Strings(retained)
		require.Len(t, retained, baseline[table].count, table)
		require.Equal(t, baseline[table].hash, fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(retained, "\n")))), table)
	}
	require.Len(t, newIDs, 238)
	require.Equal(t, len(newIDs), foundNew, "every planned new deterministic ID is in the seed")
}

func TestStarterCurriculumSeedUsesAllowedMetadata(t *testing.T) {
	seed, err := loadSeed()
	require.NoError(t, err)
	for _, n := range seed.UsageNotes {
		require.Contains(t, []string{"collocation", "register", "common_mistake", "grammar", "pronunciation", "other"}, n.NoteType)
	}
	for _, s := range seed.JourneySituations {
		require.Contains(t, []string{"daily_life", "travel", "work", "study", "social"}, s.Category)
	}
}
