package api

import (
	"context"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"time"
)

type learningExportRepository struct {
	*accounts.MemoryRepository
	payload json.RawMessage
}

func (r *learningExportRepository) ExportPersonalData(context.Context, uuid.UUID) (json.RawMessage, error) {
	return r.payload, nil
}

func TestPersonalDataExportHandlerPreservesLearningProjection(t *testing.T) {
	c := &clock.Fixed{T: testNow()}
	ar := auth.NewMemoryRepository()
	uid := uuid.New()
	ar.UpsertUser(&auth.User{ID: uid, Email: "export@example.invalid", Status: "active"})
	as := auth.NewService(ar, nil, nil, c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), auth.Config{Environment: "test", Cookie: auth.CookieConfig{Name: "session", CSRName: "csrf"}})
	memory := accounts.NewMemoryRepository()
	memory.SetUser(uid, "export@example.invalid")
	base, e := memory.ExportPersonalData(t.Context(), uid)
	require.NoError(t, e)
	var projection map[string]any
	require.NoError(t, json.Unmarshal(base, &projection))
	var learning map[string]any
	require.NoError(t, json.Unmarshal([]byte(`{
 "wordLists":[{"id":"list","name":"Personal words","revision":2,"createdAt":"created","updatedAt":"updated","deletedAt":null,"members":[{"meaningId":"meaning","addedAt":"added"}]}],
 "storySessions":[{"id":"story","storyKey":"a-quiet-lunch","title":"Original saved title","contentVersion":"original-dialogues-v1","gradingVersion":"curated-choice-v1","status":"in_progress","completedSteps":4,"totalSteps":10,"firstAnswersCorrect":0,"questionsAnswered":1,"startedAt":"created","updatedAt":"updated","completedAt":null,"actions":[{"action":{"stepId":"step","action":"answer","choiceId":"choice-1"},"feedback":{"stepId":"step","correct":false,"answer":"Canonical correction","explanation":"Shown correction"},"createdAt":"created"}]}],
 "learningPreferences":{"learningGoal":"travel","mainUseCase":"travel","revision":2,"createdAt":"created","updatedAt":"updated"},
 "wordKnowledge":[{"meaningId":"meaning","selfReportedKnown":true,"note":"Personal note","updatedAt":"updated"}],
 "practiceSessions":[{"id":"practice","mode":"typed_recall","lessonKey":null,"listId":"list","listName":"Personal words","listRevision":2,"contentVersion":"starter-90-v1","gradingVersion":"curated-v1","status":"in_progress","completedSteps":0,"totalSteps":1,"firstAnswersCorrect":0,"questionsAnswered":1,"startedAt":"created","updatedAt":"updated","completedAt":null,"actions":[{"action":{"stepId":"step","action":"answer","typedAnswer":"My original typed wording"},"meaningId":"meaning","correct":false,"feedback":{"stepId":"step","correct":false,"assisted":false,"answer":"menu","explanation":"Shown correction","wordText":"menu","wordSlug":"menu","meaningId":"meaning"},"createdAt":"created"}]}],
 "guidedLessons":[{"id":"lesson","lessonKey":"airport","lessonVersion":"1","exerciseVersion":null,"title":"Legacy title","words":[{"meaningId":"meaning","wordText":"menu","wordSlug":"menu","partOfSpeech":"noun","definition":"Definition","example":"Example","usageNote":"Note"}],"status":"in_progress","completedSteps":4,"totalSteps":9,"firstAnswersCorrect":0,"questionsAnswered":1,"feedback":null,"startedAt":"created","updatedAt":"updated","completedAt":null,"actions":[{"action":{"stepId":"typed","action":"answer","typedAnswer":"Typed original"},"feedback":{"stepId":"typed","correct":false,"explanation":"Shown correction","correctChoiceId":"","answer":"menu"},"completedSteps":4,"createdAt":"created"}]}]
 }`), &learning))
	for k, v := range learning {
		projection[k] = v
	}
	payload, e := json.Marshal(projection)
	require.NoError(t, e)
	repo := &learningExportRepository{MemoryRepository: memory, payload: payload}
	svc := accounts.NewService(repo, ar, nil, accounts.NewMemoryIdempotencyStore(), c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), accounts.Config{})
	a := humachi.New(chi.NewMux(), huma.DefaultConfig("export", "test"))
	a.UseMiddleware(withHumaContext)
	a.UseMiddleware(AuthMiddleware(as))
	RegisterPersonalDataExports(a, svc, as)
	token, cookie := as.IssueCSRFCookie()
	req := exportRequest(t, uid, "projection", true)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", token)
	w := httptest.NewRecorder()
	a.Adapter().ServeHTTP(w, req)
	require.Equal(t, 200, w.Code, w.Body.String())
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	for k, v := range learning {
		require.Equal(t, v, got[k], "HTTP download must retain repository projection %s", k)
	}
	for _, key := range []string{"wordLists", "storySessions", "learningPreferences", "practiceSessions", "wordKnowledge", "guidedLessons"} {
		require.Contains(t, a.OpenAPI().Components.Schemas.Map()["PersonalDataExportDTO"].Properties, key)
	}

	// Older repository exports and exact retries may lack newly introduced
	// collections. Preserve their version and allow null rather than manufacture
	// learning data or advertise a schema newer than the document actually used.
	for _, field := range []string{"wordLists", "storySessions", "learningPreferences", "practiceSessions", "wordKnowledge", "guidedLessons"} {
		delete(projection, field)
	}
	projection["schemaVersion"] = "1.2"
	repo.payload, e = json.Marshal(projection)
	require.NoError(t, e)
	legacyRequest := exportRequest(t, uid, "projection", true)
	legacyRequest.AddCookie(cookie)
	legacyRequest.Header.Set("X-CSRF-Token", token)
	legacyResponse := httptest.NewRecorder()
	a.Adapter().ServeHTTP(legacyResponse, legacyRequest)
	require.Equal(t, 200, legacyResponse.Code, legacyResponse.Body.String())
	var legacy map[string]any
	require.NoError(t, json.Unmarshal(legacyResponse.Body.Bytes(), &legacy))
	require.Equal(t, "1.2", legacy["schemaVersion"])
	for key, value := range projection["profile"].(map[string]any) {
		require.Equal(t, value, legacy["profile"].(map[string]any)[key])
	}
	properties := a.OpenAPI().Components.Schemas.Map()["PersonalDataExportDTO"].Properties
	for _, field := range []string{"wordLists", "storySessions", "learningPreferences", "practiceSessions", "wordKnowledge", "guidedLessons"} {
		require.Nil(t, legacy[field], "absent historical %s must stay empty", field)
		schema := properties[field]
		if schema.Ref != "" {
			schema = a.OpenAPI().Components.Schemas.SchemaFromRef(schema.Ref)
		}
		require.True(t, schema.Nullable, "null historical %s must match its schema", field)
	}

}
