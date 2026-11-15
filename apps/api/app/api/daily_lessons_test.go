package api

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/dailylesson"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
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
	"time"
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
	gam := gamification.NewService(gamification.NewRepository(db))
	mis := missions.NewService(missions.NewRepository(db), gam)
	review := reviews.NewPostgreSQLRepository(db, nil, reviews.WithGamificationService(gam), reviews.WithMissionsService(mis))
	RegisterDailyLessons(api, dailylesson.NewService(db, review, nil), a)
	mock.ExpectBegin()

	mock.ExpectQuery("SELECT id FROM users").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(u))
	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").WithArgs(u).WillReturnRows(sqlmock.NewRows([]string{"user_id", "timezone", "daily_review_target", "review_interval_preset", "notifications_enabled", "marketing_emails_enabled", "app_language"}).AddRow(u, "UTC", 15, "vocanova_default", false, false, "en"))
	day := time.Now().UTC().Truncate(24 * time.Hour)
	snapshotRows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed", "new_word_target", "new_words_completed", "sentence_practice_target", "sentence_practices_completed", "policy_version", "status", "completed_at", "grace_applied", "grace_day_id"}).AddRow(uuid.New(), u, day, "UTC", 15, 0, nil, nil, nil, nil, gamification.MissionPolicyVersion, "open", nil, false, nil)
	}
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").WillReturnRows(snapshotRows())
	mock.ExpectExec("INSERT INTO daily_review_effort").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("UPDATE daily_mission_snapshots").WillReturnRows(sqlmock.NewRows([]string{"reviews_completed"}).AddRow(0))
	mock.ExpectQuery("SELECT local_date, status, completed_at, grace_applied, grace_day_id FROM daily_mission_snapshots").WillReturnRows(sqlmock.NewRows([]string{"local_date", "status", "completed_at", "grace_applied", "grace_day_id"}))
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT COALESCE").WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(0))
	mock.ExpectQuery("SELECT user_id, current_streak_count").WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_streak_count", "longest_streak_count", "last_completed_local_date", "last_activity_local_date", "timezone", "status", "created_at", "updated_at"}))
	mock.ExpectExec("INSERT INTO streak_states").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id, user_id, local_date").WillReturnRows(snapshotRows())
	mock.ExpectQuery("SELECT count.* FROM daily_review_effort").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT min").WillReturnRows(sqlmock.NewRows([]string{"next_review_at"}).AddRow(nil))
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
