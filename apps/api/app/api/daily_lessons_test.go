package api

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/dailylesson"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDailyLessonAPIProtectsAllRoutes(t *testing.T) {
	for _, r := range []struct{ method, path string }{{"GET", "/api/v1/daily-lessons/current"}, {"GET", "/api/v1/daily-lessons/00000000-0000-0000-0000-000000000001"}, {"POST", "/api/v1/daily-lessons"}, {"POST", "/api/v1/daily-lessons/00000000-0000-0000-0000-000000000001/actions"}} {
		t.Run(r.method+r.path, func(t *testing.T) {
			a := authStubService()
			api := humachi.New(chi.NewMux(), huma.DefaultConfig("Daily", "test"))
			api.UseMiddleware(withHumaContext)
			api.UseMiddleware(AuthMiddleware(a))
			RegisterDailyLessons(api, dailylesson.NewService(nil, nil, nil), a)
			w := httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, httptest.NewRequest(r.method, r.path, nil))
			require.Equal(t, http.StatusUnauthorized, w.Code)
			require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		})
	}
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("Daily", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	RegisterDailyLessons(api, dailylesson.NewService(nil, nil, nil), a)
	req := httptest.NewRequest("POST", "/api/v1/daily-lessons", nil)
	req.Header.Set("Idempotency-Key", "start")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: uuid.New()}))
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
}
func TestDailyLessonErrorsRemainDistinct(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{{dailylesson.ErrNotFound, 404}, {dailylesson.ErrInvalid, 400}, {dailylesson.ErrConflict, 409}, {dailylesson.ErrNoDue, 409}, {dailylesson.ErrContentUnavailable, 503}} {
		e, ok := mapDailyLessonError(tc.err).(huma.StatusError)
		require.True(t, ok)
		require.Equal(t, tc.status, e.GetStatus())
	}
}

func TestLegacyReviewAPIRejectsForgedDailyLessonProvenance(t *testing.T) {
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("Review provenance", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	svc := reviews.NewService(reviews.NewMemoryRepository(reviews.MemoryRepositoryData{}), nil, nil)
	RegisterReviews(api, svc, a)
	body := fmt.Sprintf(`{"userWordId":"%s","meaningId":"%s","promptType":"self_check","result":"correct","rating":"good","answeredAt":"2026-10-17T12:00:00Z","clientAttemptId":"forged","metadata":{"dailySessionId":"%s"}}`, uuid.New(), uuid.New(), uuid.New())
	req := httptest.NewRequest("POST", "/api/v1/reviews/submissions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "forged")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: uuid.New()}))
	addCSRF(req, a)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "reserved")
}

func TestDailyLessonCurrentSuccessfulResponseIsNeverCached(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	u := uuid.New()
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("Daily private", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	RegisterDailyLessons(api, dailylesson.NewService(db, nil, nil), a)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT state FROM daily_lesson_sessions").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"state"}))
	mock.ExpectQuery("SELECT count").WithArgs(u, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectCommit()
	req := httptest.NewRequest("GET", "/api/v1/daily-lessons/current", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: u}))
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.NoError(t, mock.ExpectationsWereMet())
}
