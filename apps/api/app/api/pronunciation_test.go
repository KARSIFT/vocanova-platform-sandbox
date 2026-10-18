package api

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/pronunciation"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type pronunciationAPIProvider struct{ calls int }

func (p *pronunciationAPIProvider) Generate(context.Context, string) ([]byte, error) {
	p.calls++
	return nil, pronunciation.ErrUnavailable
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
	require.Equal(t, 422, send(`{"text":"`+strings.Repeat("x", 501)+`"}`, u, true).Code)
	mock.ExpectQuery("SELECT EXISTS").WithArgs(u, "greeting").WillReturnRows(sqlmock.NewRows([]string{"active", "word"}).AddRow(true, true))
	mock.ExpectQuery("SELECT audio FROM pronunciation_audio_cache").WillReturnRows(sqlmock.NewRows([]string{"audio"}).AddRow([]byte{'I', 'D', '3', 4, 0, 0, 0, 0, 0, 1, 0}))
	w := send(`{"text":"greeting"}`, u, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"cacheHit":true`)
	require.Contains(t, w.Body.String(), `"voice":"luna"`)
	require.Zero(t, p.calls)
	require.NoError(t, mock.ExpectationsWereMet())
}
