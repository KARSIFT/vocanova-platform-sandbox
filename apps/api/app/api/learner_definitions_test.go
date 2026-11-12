package api

import (
	"context"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordlists"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLearnerExplanationsPreserveCanonicalFieldsAcrossProjections(t *testing.T) {
	id := uuid.MustParse("6d7b77f2-7914-5bcb-99f3-73ce1ea4619f")
	const canonical = "forsake, leave behind"
	const explanation = "To leave someone or something behind."
	search := searchMeaningToDTO(content.SearchMeaning{MeaningSummary: content.MeaningSummary{MeaningID: id, WordText: "abandon", PartOfSpeech: "verb", ShortDefinition: canonical}})
	require.Equal(t, canonical, search.ShortDefinition)
	require.Equal(t, explanation, search.LearnerDefinition)

	word := wordToDTO(&content.WordDetail{Text: "abandon", Meanings: []content.WordMeaning{{ID: id, PartOfSpeech: "verb", ShortDefinition: canonical}}})
	require.Equal(t, canonical, word.Meanings[0].ShortDefinition)
	require.Equal(t, explanation, word.Meanings[0].LearnerDefinition)
	require.Equal(t, meaningidentity.Revision(id.String(), "abandon", "verb", canonical), word.Meanings[0].MeaningRevision)
	saved := savedMeaningToDTO(learning.SavedMeaning{MeaningID: id, WordText: "abandon", PartOfSpeech: "verb", ShortDefinition: canonical})
	require.Equal(t, canonical, saved.ShortDefinition)
	require.Equal(t, explanation, saved.LearnerDefinition)
	due := dueWordToDTO(reviews.DueWord{MeaningID: id, WordText: "abandon", PartOfSpeech: "verb", ShortDefinition: canonical})
	require.Equal(t, canonical, due.ShortDefinition)
	require.Equal(t, explanation, due.LearnerDefinition)
	detail := wordListWithExplanations(&wordlists.WordListDetail{Members: []wordlists.WordListMember{{MeaningID: id.String(), WordText: "abandon", PartOfSpeech: "verb", ShortDefinition: canonical}}})
	require.Equal(t, canonical, detail.Members[0].ShortDefinition)
	require.Equal(t, explanation, detail.Members[0].LearnerDefinition)
	// Projection must not mutate a repository's shared result or authenticated state.
	source := &wordlists.WordListDetail{Members: []wordlists.WordListMember{{MeaningID: id.String(), WordText: "abandon", PartOfSpeech: "verb", ShortDefinition: canonical}}}
	wordListWithExplanations(source)
	require.Empty(t, source.Members[0].LearnerDefinition)
	authored := wordToDTO(&content.WordDetail{Text: "abandon", Meanings: []content.WordMeaning{{ID: id, PartOfSpeech: "verb", ShortDefinition: canonical, LearnerDefinition: "An existing teaching explanation."}}})
	require.Equal(t, "An existing teaching explanation.", authored.Meanings[0].LearnerDefinition)
}

func TestLearnerExplanationListProjectionKeepsEmptyArray(t *testing.T) {
	source := &wordlists.WordListDetail{Members: []wordlists.WordListMember{}}
	result := wordListWithExplanations(source)
	data, err := json.Marshal(result)
	require.NoError(t, err)
	require.Contains(t, string(data), `"members":[]`)
	require.NotNil(t, source.Members)
}

type learnerCatalogFixture struct{ catalogHTTPFixture }

func (r *learnerCatalogFixture) Read(_ context.Context, _ uuid.UUID, band int, _ time.Time) (*knowledgecatalog.Map, error) {
	return &knowledgecatalog.Map{CatalogVersion: knowledgecatalog.CatalogVersion, TotalWords: 10000, BandSize: 100, CurrentBand: knowledgecatalog.CurrentBand{Number: band, Items: []knowledgecatalog.Item{{MeaningID: uuid.MustParse("6d7b77f2-7914-5bcb-99f3-73ce1ea4619f"), WordText: "abandon", PartOfSpeech: "verb", ShortDefinition: "forsake, leave behind"}}}}, nil
}
func TestKnowledgeMapAddsExplanationWithoutChangingRevision(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	RegisterKnowledgeMap(api, knowledgecatalog.NewService(&learnerCatalogFixture{}, nil), authSvc)
	req := httptest.NewRequest("GET", "/api/v1/knowledge-map", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: uuid.New()}))
	response := httptest.NewRecorder()
	api.Adapter().ServeHTTP(response, req)
	require.Equal(t, 200, response.Code)
	var result knowledgecatalog.Map
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Len(t, result.CurrentBand.Items, 1)
	item := result.CurrentBand.Items[0]
	require.Equal(t, "forsake, leave behind", item.ShortDefinition)
	require.Equal(t, "To leave someone or something behind.", item.LearnerDefinition)
	require.Equal(t, meaningidentity.Revision(item.MeaningID.String(), item.WordText, item.PartOfSpeech, item.ShortDefinition), item.MeaningRevision)
}
