package api

import (
	"context"
	"net/http/httptest"
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
	send := func(body string, user uuid.UUID, csrf bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/api/v1/pronunciation", strings.NewReader(body))
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
	mock.ExpectQuery("SELECT audio FROM pronunciation_audio_cache").WillReturnRows(sqlmock.NewRows([]string{"audio"}).AddRow([]byte{'I', 'D', '3', 4, 0, 0, 0, 0, 0, 1, 0}))
	w := send(`{"text":"greeting","meaningId":"`+meaning+`","meaningRevision":"`+revision+`"}`, u, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"cacheHit":true`)
	require.Contains(t, w.Body.String(), `"voice":"marin"`)
	require.Contains(t, w.Body.String(), `"meaningId":"`+meaning+`"`)
	require.Contains(t, w.Body.String(), `"partOfSpeech":"noun"`)
	require.Contains(t, w.Body.String(), `"meaningRevision":"`+revision+`"`)
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
