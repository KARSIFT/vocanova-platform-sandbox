package api

import (
	"context"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/pronunciation"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type pronunciationAPIProvider struct{ calls int }

func (p *pronunciationAPIProvider) Generate(context.Context, pronunciation.Utterance) ([]byte, error) {
	p.calls++
	return nil, pronunciation.ErrUnavailable
}

func (p *pronunciationAPIProvider) Identity() pronunciation.ProviderIdentity {
	return pronunciation.ProviderIdentity{Provider: "openai", Model: pronunciation.OpenAIModel, Voice: pronunciation.OpenAIVoice, InstructionVersion: pronunciation.OpenAIInstructionVersion}
}

func TestPronunciationAPIAuthCSRFSourceBoundaryAndCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	p := &pronunciationAPIProvider{}
	svc := pronunciation.NewService(db, p, true, nil)
	authSvc := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("pronunciation", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterPronunciation(api, svc, authSvc)
	u := uuid.New()
	send := func(body string, user uuid.UUID, csrf bool, query ...string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/api/v1/pronunciation"+strings.Join(query, ""), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if user != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
		}
		if csrf {
			addCSRF(req, authSvc)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	require.Equal(t, 401, send(`{"text":"greeting"}`, uuid.Nil, false).Code)
	require.Equal(t, 403, send(`{"text":"greeting"}`, u, false).Code)
	require.Equal(t, 422, send(`{"text":"greeting","voice":"unbounded"}`, u, true).Code)
	require.Equal(t, 422, send(`{"text":"greeting","instructions":"Speak my private draft"}`, u, true).Code)
	require.Equal(t, 422, send(`{"text":"greeting","meaningId":"not-a-uuid"}`, u, true).Code)
	require.Equal(t, 422, send(`{"text":"`+strings.Repeat("x", 501)+`"}`, u, true).Code)
	require.Equal(t, 422, send(`{"text":"greeting","meaningRevision":"bad"}`, u, true).Code)
	require.Equal(t, 400, send(`{"text":"greeting","meaningRevision":"`+strings.Repeat("a", 64)+`"}`, u, true).Code)
	meaning := uuid.NewString()
	revision := meaningidentity.Revision(meaning, "greeting", "noun", "A greeting.")
	mock.ExpectQuery("SELECT EXISTS").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"active"}).AddRow(true))
	mock.ExpectQuery("SELECT wm.id").WithArgs("greeting", meaning).WillReturnRows(sqlmock.NewRows([]string{"id", "text", "pos", "definition", "example"}).AddRow(meaning, "greeting", "noun", "A greeting.", ""))
	tone, err := os.ReadFile("../../business/pronunciation/testdata/tone.mp3")
	require.NoError(t, err)
	mock.ExpectQuery("SELECT audio FROM pronunciation_audio_cache").WillReturnRows(sqlmock.NewRows([]string{"audio"}).AddRow(tone))
	w := send(`{"text":"greeting","meaningId":"`+meaning+`","meaningRevision":"`+revision+`"}`, u, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"cacheHit":true`)
	require.Contains(t, w.Body.String(), `"voice":"marin"`)
	require.Contains(t, w.Body.String(), `"meaningId":"`+meaning+`"`)
	require.Contains(t, w.Body.String(), `"partOfSpeech":"noun"`)
	require.Contains(t, w.Body.String(), `"meaningRevision":"`+revision+`"`)
	require.Zero(t, p.calls)
	// The published noun and verb recordings share spelling but have independent identities.
	const recordID = "009949f4-3605-530f-84f4-301a6beeb809"
	const recordRevision = "ac85897ffcd435c298931efb2cecda765a0ac46eeafbb95cb5c5d9bb8f43127f"
	const recordSHA = "277db426ed8ac44abf350ff75ae788bec1ae967cdceeb4697f56c84df3ac889f"
	const recordKey = "ea10ba2f066d9cfff498629929e48a9d2c7351bb00226795c6c46aefc16bdc6b"
	const recordBody = `{"text":"record","meaningId":"` + recordID + `","meaningRevision":"` + recordRevision + `"}`
	expectRecord := func(definition string) {
		mock.ExpectQuery("SELECT EXISTS").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"active"}).AddRow(true))
		mock.ExpectQuery("SELECT wm.id").WithArgs("record", recordID).WillReturnRows(sqlmock.NewRows([]string{"id", "text", "pos", "definition", "example"}).
			AddRow(recordID, "record", "noun", definition, "Keep a record of every payment."))
	}
	const recordDefinition = "Information kept so that it can be checked later."
	expectRecord(recordDefinition)
	w = send(recordBody, u, true, "?preferCdn=true")
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"audioUrl":"https://audio.vocanova.site/meanings/`+recordID+"/"+recordRevision+"/"+recordKey+"/"+recordSHA+`.mp3"`)
	require.Contains(t, w.Body.String(), `"audioBase64":""`)
	require.Contains(t, w.Body.String(), `"meaningRevision":"`+recordRevision+`"`)

	// Omitting opt-in, including a browser retry after CDN failure, keeps the byte response.
	expectRecord(recordDefinition)
	mock.ExpectQuery("SELECT audio FROM pronunciation_audio_cache").WithArgs(recordKey).WillReturnRows(sqlmock.NewRows([]string{"audio"}).AddRow(tone))
	w = send(recordBody, u, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.NotContains(t, w.Body.String(), `"audioUrl"`)
	require.NotContains(t, w.Body.String(), `"audioBase64":""`)

	// A current meaning with a changed definition must use a different recording.
	expectRecord("Saved information.")
	require.Equal(t, 404, send(recordBody, u, true, "?preferCdn=true").Code)
	expectRecord("Saved information.")
	mock.ExpectQuery("SELECT audio FROM pronunciation_audio_cache").WillReturnRows(sqlmock.NewRows([]string{"audio"}).AddRow(tone))
	w = send(`{"text":"record","meaningId":"`+recordID+`"}`, u, true, "?preferCdn=true")
	require.Equal(t, 200, w.Code, w.Body.String())
	require.NotContains(t, w.Body.String(), `"audioUrl"`)
	require.Zero(t, p.calls)
	// A valid public spelling cannot be paired with an unrelated selected meaning.
	mock.ExpectQuery("SELECT EXISTS").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"active"}).AddRow(true))
	mock.ExpectQuery("SELECT wm.id").WithArgs("unrelated", meaning).WillReturnRows(sqlmock.NewRows([]string{"id", "text", "pos", "definition", "example"}))
	require.Equal(t, 404, send(`{"text":"unrelated","meaningId":"`+meaning+`"}`, u, true).Code)
	require.Zero(t, p.calls)
	// Same ID with changed live POS must not play audio for a frozen noun card.
	mock.ExpectQuery("SELECT EXISTS").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"active"}).AddRow(true))
	mock.ExpectQuery("SELECT wm.id").WithArgs("greeting", meaning).WillReturnRows(sqlmock.NewRows([]string{"id", "text", "pos", "definition", "example"}).AddRow(meaning, "greeting", "verb", "A greeting.", ""))
	require.Equal(t, 404, send(`{"text":"greeting","meaningId":"`+meaning+`","meaningRevision":"`+revision+`"}`, u, true).Code)
	require.Zero(t, p.calls)
	require.NoError(t, mock.ExpectationsWereMet())
}
