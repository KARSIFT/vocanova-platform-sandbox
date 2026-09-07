package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testContentAPI wires the content routes against contentSampleData()
// and usersSvc (which may be nil, in which case Discover falls back to
// plain display_order — see requesterMainUseCase).
func testContentAPI(t *testing.T, usersSvc *users.Service) (huma.API, *content.Service) {
	repo, reader := contentSampleData()
	svc := content.NewService(repo, reader)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authStubService()))
	RegisterContent(api, svc, usersSvc)
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
	reader := content.NewMemorySavedStateReaderWithStates(map[uuid.UUID]map[uuid.UUID]content.SavedWordState{
		userID: {meaningID: {UserWordID: userWordID, Status: "learning", Due: true}},
	})
	return repo, reader
}

func authenticatedRequest(t *testing.T, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations", nil)
	ctx := WithRequester(req.Context(), &auth.User{ID: userID})
	req = req.WithContext(ctx)
	return req
}

func TestListJourneySituationsRequiresAuth(t *testing.T) {
	api, _ := testContentAPI(t, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListJourneySituationsReturnsActiveSituations(t *testing.T) {
	api, _ := testContentAPI(t, nil)

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
	api, _ := testContentAPI(t, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations/airport", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetJourneySituationUnknownSlugReturns404(t *testing.T) {
	api, _ := testContentAPI(t, nil)

	w := httptest.NewRecorder()
	req := authenticatedRequest(t, content.MustParseUUID("00000000-0000-0000-0000-000000000009"))
	req = httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations/unknown", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetJourneySituationReturnsSavedOverlay(t *testing.T) {
	api, _ := testContentAPI(t, nil)

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
	api, _ := testContentAPI(t, nil)

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
	api, _ := testContentAPI(t, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/canonical-words/boarding-pass", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCanonicalWordUnknownSlugReturns404(t *testing.T) {
	api, _ := testContentAPI(t, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/canonical-words/unknown", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListJourneySituationsInvalidCursorReturns400(t *testing.T) {
	api, _ := testContentAPI(t, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journey-situations?after=not-valid", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestListJourneySituationsPrioritizesOnboardingMainUseCase is a
// regression test for VOC-1183: onboarding collects a "main use case"
// (daily_life|work|travel|study|social) that maps 1:1 onto
// journey_situations.category, but until this change nothing ever
// read it back — a learner's Discover feed was ordered purely by
// display_order regardless of their stated goal. This asserts that a
// learner whose completed onboarding profile says mainUseCase=work
// sees the work-category situation surfaced ahead of a lower-
// display_order travel situation on the first Discover page.
func TestListJourneySituationsPrioritizesOnboardingMainUseCase(t *testing.T) {
	travelID := content.MustParseUUID("00000000-0000-0000-0000-000000000101")
	workID := content.MustParseUUID("00000000-0000-0000-0000-000000000102")
	dailyLifeID := content.MustParseUUID("00000000-0000-0000-0000-000000000103")
	repo := content.NewMemoryRepository(content.MemoryRepositoryData{
		Situations: []content.Situation{
			{ID: travelID, Slug: "airport", Title: "Airport", ShortDescription: "d", Category: "travel", Status: "active", DisplayOrder: 1},
			{ID: workID, Slug: "meeting-room", Title: "Meeting room", ShortDescription: "d", Category: "work", Status: "active", DisplayOrder: 2},
			{ID: dailyLifeID, Slug: "grocery-store", Title: "Grocery store", ShortDescription: "d", Category: "daily_life", Status: "active", DisplayOrder: 3},
		},
	})
	svc := content.NewService(repo, content.NewMemorySavedStateReader(nil))

	usersRepo := users.NewMemoryRepository()
	usersSvc := users.NewService(usersRepo, usersRepo, usersRepo, clock.Real{})
	workLearner := content.MustParseUUID("00000000-0000-0000-0000-000000000201")
	_, _, err := usersSvc.CompleteOnboarding(context.Background(), workLearner, users.OnboardingAnswers{
		EnglishLevel:      users.EnglishLevelB1,
		NativeLanguage:    "es",
		LearningGoal:      users.LearningGoalWork,
		MainUseCase:       users.MainUseCaseWork,
		DailyReviewTarget: 20,
	})
	require.NoError(t, err)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authStubService()))
	RegisterContent(api, svc, usersSvc)

	// A learner who never onboarded still sees plain display_order.
	w := httptest.NewRecorder()
	req := authenticatedRequest(t, content.MustParseUUID("00000000-0000-0000-0000-000000000009"))
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var unonboarded struct {
		Items []SituationDTO `json:"items"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &unonboarded))
	require.Len(t, unonboarded.Items, 3)
	assert.Equal(t, []string{"airport", "meeting-room", "grocery-store"}, slugsOf(unonboarded.Items))

	// A learner whose stated goal is "work" sees the work situation
	// first, ahead of the lower-display_order travel situation.
	w = httptest.NewRecorder()
	req = authenticatedRequest(t, workLearner)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var onboarded struct {
		Items []SituationDTO `json:"items"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &onboarded))
	require.Len(t, onboarded.Items, 3)
	assert.Equal(t, "meeting-room", onboarded.Items[0].Slug, "work-goal learner should see the work situation surfaced first")
	assert.Equal(t, []string{"airport", "grocery-store"}, slugsOf(onboarded.Items[1:]), "non-matching situations keep their relative display_order")
}

func slugsOf(items []SituationDTO) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.Slug
	}
	return out
}

func TestGetCanonicalWordReturnsWordDetail(t *testing.T) {
	api, _ := testContentAPI(t, nil)

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
	assert.Equal(t, "learning", meaning.ReviewState)
	assert.True(t, meaning.Due)
}

func TestGetCanonicalWordDoesNotExposeAnotherLearnersReviewState(t *testing.T) {
	api, _ := testContentAPI(t, nil)

	w := httptest.NewRecorder()
	otherUser := content.MustParseUUID("00000000-0000-0000-0000-00000000000a")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/canonical-words/boarding-pass", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: otherUser}))
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Word WordDetailDTO `json:"word"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Word.Meanings, 1)
	assert.False(t, body.Word.Meanings[0].Saved)
	assert.Empty(t, body.Word.Meanings[0].UserWordID)
	assert.Empty(t, body.Word.Meanings[0].ReviewState)
	assert.False(t, body.Word.Meanings[0].Due)
}

func intPtr(i int) *int { return &i }

// Ensure requester context works without Huma middleware for unit tests.
func TestWithRequesterContext(t *testing.T) {
	u := &auth.User{ID: content.MustParseUUID("00000000-0000-0000-0000-000000000009")}
	ctx := WithRequester(context.Background(), u)
	assert.Equal(t, u.ID, RequesterUserID(ctx))
}
