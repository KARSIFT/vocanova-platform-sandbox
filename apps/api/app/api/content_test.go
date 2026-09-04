package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContentAPI(t *testing.T) (huma.API, *content.Service) {
	repo, reader := contentSampleData()
	svc := content.NewService(repo, reader)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authStubService()))
	RegisterContent(api, svc)
	return api, svc
}

func authStubService() *auth.Service {
	// AuthMiddleware only needs the session cookie name; the service is not used
	// for content route tests because the requester is injected directly.
	repo := auth.NewMemoryRepository()
	return auth.NewService(repo, nil, nil, nil, nil, auth.Config{
		Cookie: auth.CookieConfig{Name: "vocanova_session", CSRName: "vocanova_csrf"},
	})
}

func contentSampleData() (*content.MemoryRepository, *content.MemorySavedStateReader) {
	sitID := content.MustParseUUID("00000000-0000-0000-0000-000000000001")
	wordID := content.MustParseUUID("00000000-0000-0000-0000-000000000002")
	meaningID := content.MustParseUUID("00000000-0000-0000-0000-000000000003")
	userID := content.MustParseUUID("00000000-0000-0000-0000-000000000009")

	repo := content.NewMemoryRepository(content.MemoryRepositoryData{
		Situations: []content.Situation{
			{
				ID:               sitID,
				Slug:             "airport",
				Title:            "Airport",
				ShortDescription: "Airport words.",
				LevelBand:        "a1_a2",
				Category:         "travel",
				Status:           "active",
				DisplayOrder:     1,
			},
		},
		Words: []content.SeedWord{
			{
				ID:              wordID,
				Text:            "boarding pass",
				NormalizedText:  "boarding pass",
				WordType:        "phrase",
				LanguageCode:    "en",
				Status:          "active",
				DifficultyLevel: "a1_a2",
			},
		},
		Meanings: []content.SeedMeaning{
			{
				ID:                meaningID,
				WordID:            wordID,
				PartOfSpeech:      "noun",
				ShortDefinition:   "A document that lets you get on your flight.",
				LearnerDefinition: "The ticket-like paper or app screen you show before getting on a plane.",
				MeaningOrder:      1,
				Status:            "active",
			},
		},
		Examples: []content.SeedExample{
			{
				ID:             content.MustParseUUID("00000000-0000-0000-0000-000000000004"),
				MeaningID:      meaningID,
				ExampleText:    "Please have your boarding pass ready at the gate.",
				ExampleOrder:   1,
				Status:         "active",
				SituationLabel: "Airport",
			},
		},
		Notes: []content.SeedNote{
			{
				ID:        content.MustParseUUID("00000000-0000-0000-0000-000000000005"),
				MeaningID: meaningID,
				NoteType:  "formality",
				NoteText:  "Neutral; used in both spoken and written travel contexts.",
				NoteOrder: 1,
				Status:    "active",
			},
		},
		JourneyWords: []content.SeedJourneyWord{
			{
				ID:                 content.MustParseUUID("00000000-0000-0000-0000-000000000008"),
				JourneySituationID: sitID,
				MeaningID:          meaningID,
				RelevanceScore:     80,
				DisplayOrder:       intPtr(1),
				IsCore:             true,
			},
		},
	})
	userWordID := content.MustParseUUID("00000000-0000-0000-0000-000000000010")
	reader := content.NewMemorySavedStateReaderWithIDs(
		map[uuid.UUID]map[uuid.UUID]bool{
			userID: {meaningID: true},
		},
		map[uuid.UUID]map[uuid.UUID]uuid.UUID{
			userID: {meaningID: userWordID},
		},
	)
	return repo, reader
}

func authenticatedRequest(t *testing.T, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations", nil)
	ctx := WithRequester(req.Context(), &auth.User{ID: userID})
	req = req.WithContext(ctx)
	return req
}

func TestListJourneySituationsRequiresAuth(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListJourneySituationsReturnsActiveSituations(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	req := authenticatedRequest(t, content.MustParseUUID("00000000-0000-0000-0000-000000000009"))
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items      []SituationDTO `json:"items"`
		NextCursor string         `json:"nextCursor"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Items, 1)
	assert.Equal(t, "airport", body.Items[0].Slug)
	assert.Empty(t, body.NextCursor)
}

func TestGetJourneySituationRequiresAuth(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations/airport", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetJourneySituationUnknownSlugReturns404(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	req := authenticatedRequest(t, content.MustParseUUID("00000000-0000-0000-0000-000000000009"))
	req = httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations/unknown", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetJourneySituationReturnsSavedOverlay(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	userID := content.MustParseUUID("00000000-0000-0000-0000-000000000009")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations/airport", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Situation SituationDTO          `json:"situation"`
		Meanings  []SituationMeaningDTO `json:"meanings"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "airport", body.Situation.Slug)
	require.Len(t, body.Meanings, 1)
	assert.Equal(t, "boarding-pass", body.Meanings[0].WordSlug)
	assert.True(t, body.Meanings[0].Saved)
}

func TestGetJourneySituationCrossUserSavedState(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	otherUser := content.MustParseUUID("00000000-0000-0000-0000-00000000000a")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations/airport", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: otherUser}))
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Meanings []SituationMeaningDTO `json:"meanings"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Meanings, 1)
	assert.False(t, body.Meanings[0].Saved)
}

func TestGetCanonicalWordRequiresAuth(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/canonical-words/boarding-pass", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCanonicalWordUnknownSlugReturns404(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/canonical-words/unknown", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListJourneySituationsInvalidCursorReturns400(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations?after=not-valid", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCanonicalWordReturnsWordDetail(t *testing.T) {
	api, _ := testContentAPI(t)

	w := httptest.NewRecorder()
	userID := content.MustParseUUID("00000000-0000-0000-0000-000000000009")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/canonical-words/boarding-pass", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Word WordDetailDTO `json:"word"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "boarding pass", body.Word.Text)
	assert.Equal(t, "boarding-pass", body.Word.Slug)
	assert.Equal(t, "phrase", body.Word.WordType)
	assert.Equal(t, "a1_a2", body.Word.DifficultyLevel)
	require.Len(t, body.Word.Meanings, 1)

	// PRD §2 MVP completion criteria requires word detail to surface meaning,
	// part of speech, examples, and usage notes. Assert each field explicitly so a
	// future change that silently drops one from the DTO or its mapping fails
	// this test instead of shipping unnoticed.
	meaning := body.Word.Meanings[0]
	assert.Equal(t, "noun", meaning.PartOfSpeech)
	assert.Equal(t, "A document that lets you get on your flight.", meaning.ShortDefinition)
	assert.Equal(t, "The ticket-like paper or app screen you show before getting on a plane.", meaning.LearnerDefinition)
	require.Len(t, meaning.Examples, 1)
	assert.Equal(t, "Please have your boarding pass ready at the gate.", meaning.Examples[0].ExampleText)
	require.Len(t, meaning.UsageNotes, 1)
	assert.Equal(t, "formality", meaning.UsageNotes[0].NoteType)
	assert.Equal(t, "Neutral; used in both spoken and written travel contexts.", meaning.UsageNotes[0].NoteText)

	assert.True(t, meaning.Saved)
	assert.Equal(t, "00000000-0000-0000-0000-000000000010", meaning.UserWordID)
}

func intPtr(i int) *int { return &i }

// Ensure requester context works without Huma middleware for unit tests.
func TestWithRequesterContext(t *testing.T) {
	u := &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}
	ctx := WithRequester(context.Background(), u)
	assert.Equal(t, u.ID, RequesterUserID(ctx))
}
