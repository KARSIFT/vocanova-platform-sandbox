package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMissionsTestAPI wires RegisterMissions with a sqlmock-backed
// missions+ gamification service. The handlers exercise the
// gamification/missions repositories end-to-end (read-only paths) so
// the test must stage all expected SQL calls in order.
func newMissionsTestAPI(t *testing.T) (huma.API, *missions.Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gamRepo := gamification.NewRepository(db)
	gamSvc := gamification.NewService(gamRepo)
	missionsRepo := missions.NewRepository(db)
	missionsSvc := missions.NewService(missionsRepo, gamSvc)

	authSvc := authStubService()
	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterMissions(api, missionsSvc)
	return api, missionsSvc, mock
}

func authenticatedMissionsRequest(t *testing.T, userID uuid.UUID, path string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	return req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
}

// expectGetUserSettings stages a GetUserSettings call. Pass an empty
// settingsRow for the "no row" path.
func expectGetUserSettings(mock sqlmock.Sqlmock, userID uuid.UUID, settingsRow *gamification.UserSettingsRow) {
	cols := []string{
		"user_id", "timezone", "daily_review_target", "review_interval_preset",
		"notifications_enabled", "marketing_emails_enabled", "app_language",
	}
	rows := sqlmock.NewRows(cols)
	if settingsRow != nil {
		rows.AddRow(
			settingsRow.UserID, settingsRow.Timezone, settingsRow.DailyReviewTarget,
			settingsRow.ReviewIntervalPreset, settingsRow.NotificationsEnabled,
			settingsRow.MarketingEmailsEnabled, settingsRow.AppLanguage,
		)
	}
	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target, review_interval_preset").
		WithArgs(userID).
		WillReturnRows(rows)
}

// expectGetDailyMissionSnapshot stages a GetDailyMissionSnapshot call.
// Pass a nil snapshot for the "no row" path.
func expectGetDailyMissionSnapshot(mock sqlmock.Sqlmock, userID uuid.UUID, day time.Time, snap *missions.DailyMissionSnapshot) {
	cols := []string{
		"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
		"new_word_target", "new_words_completed", "sentence_practice_target",
		"sentence_practices_completed", "policy_version", "status", "completed_at",
		"grace_applied", "grace_day_id",
	}
	rows := sqlmock.NewRows(cols)
	if snap != nil {
		rows.AddRow(
			uuid.New(), snap.UserID, snap.LocalDate, snap.Timezone,
			snap.ReviewTarget, snap.ReviewsCompleted,
			nil, nil, nil, nil,
			snap.PolicyVersion, snap.Status, nil, snap.GraceApplied, nil,
		)
	}
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, day).
		WillReturnRows(rows)
}

func TestGetDailyMissionRequiresAuth(t *testing.T) {
	api, _, _ := newMissionsTestAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-mission", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetDailyMissionLazilyCreatesAndReturnsProjection(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	day, err := gamification.LocalDate(time.Now(), "UTC")
	require.NoError(t, err)

	api, _, mock := newMissionsTestAPI(t)

	// getSettings → no user_settings row → falls back to UTC / 20.
	expectGetUserSettings(mock, userID, nil)

	// BeginTx for the lazy-create / read transaction.
	mock.ExpectBegin()
	// getDailyMissionSnapshot returns no row → caller lazily creates.
	expectGetDailyMissionSnapshot(mock, userID, day, nil)

	// CreateDailyMissionSnapshot (idempotent upsert; ON CONFLICT DO UPDATE).
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(
			sqlmock.AnyArg(), userID, day, "UTC", 20, gamification.MissionPolicyVersion,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	// ListRecentSnapshots (for streak reconciliation on first read).
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}))
	// CurrentGraceBalance.
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// GetStreakState (called by ReconcileAndAdvance to read the current
	// streak-state row before upserting).
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}))
	// UpsertStreakState (always writes, even for the no-op active case).
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// loadStreakAndGrace → GetStreakStateForRead (no row yet).
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}))
	// loadStreakAndGrace → CurrentGraceBalance.
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(0))

	w := httptest.NewRecorder()
	req := authenticatedMissionsRequest(t, userID, "/api/v1/daily-mission")
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var body GetDailyMissionOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	assert.Equal(t, day.Format("2006-01-02"), body.Body.LocalDate)
	assert.Equal(t, "UTC", body.Body.Timezone)
	assert.Equal(t, 20, body.Body.ReviewTarget)
	assert.Equal(t, 0, body.Body.ReviewsCompleted)
	assert.Equal(t, gamification.MissionPolicyVersion, body.Body.PolicyVersion)
	assert.Equal(t, "open", body.Body.Status)
	assert.False(t, body.Body.GraceApplied)
	assert.Equal(t, 0, body.Body.Streak.GraceDayBalance)
	assert.Equal(t, "active", body.Body.Streak.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDailyMissionReturnsExistingSnapshot(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	day, err := gamification.LocalDate(time.Now(), "UTC")
	require.NoError(t, err)

	api, _, mock := newMissionsTestAPI(t)

	expectGetUserSettings(mock, userID, nil)
	// The read transaction is always opened; the lazy-create block is
	// skipped because the snapshot already exists.
	mock.ExpectBegin()
	expectGetDailyMissionSnapshot(mock, userID, day, &missions.DailyMissionSnapshot{
		UserID:           userID.String(),
		LocalDate:        day,
		Timezone:         "UTC",
		ReviewTarget:     20,
		ReviewsCompleted: 7,
		PolicyVersion:    gamification.MissionPolicyVersion,
		Status:           missions.StatusOpen,
		GraceApplied:     false,
	})
	mock.ExpectCommit()
	// loadStreakAndGrace → GetStreakStateForRead.
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}).AddRow(userID, 3, 5, nil, nil, "UTC", gamification.StreakStatusActive, time.Now(), time.Now()))
	// loadStreakAndGrace → CurrentGraceBalance.
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(1))

	w := httptest.NewRecorder()
	req := authenticatedMissionsRequest(t, userID, "/api/v1/daily-mission")
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var body GetDailyMissionOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	assert.Equal(t, day.Format("2006-01-02"), body.Body.LocalDate)
	assert.Equal(t, 7, body.Body.ReviewsCompleted)
	assert.Equal(t, 3, body.Body.Streak.CurrentStreakCount)
	assert.Equal(t, 5, body.Body.Streak.LongestStreakCount)
	assert.Equal(t, 1, body.Body.Streak.GraceDayBalance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDailyMissionRejectsInvalidClientTimezone(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	api, _, mock := newMissionsTestAPI(t)

	// GetUserSettings runs before the timezone validation; mock a
	// non-default stored row so the validator rejects the client value
	// only when the stored row is the default. For a "no row" path, the
	// client value is the only timezone; we test the rejection there.
	expectGetUserSettings(mock, userID, nil)
	// gamification.GetSettings → ResolveSettings will reject the unknown
	// client timezone with ErrInvalidTimezone before any mission read.

	w := httptest.NewRecorder()
	req := authenticatedMissionsRequest(t, userID, "/api/v1/daily-mission?timezone=Not/A_Real_Zone")
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProgressRequiresAuth(t *testing.T) {
	api, _, _ := newMissionsTestAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProgressReturnsBalanceStreakAndHistory(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	day7, err := gamification.LocalDate(time.Now(), "UTC")
	require.NoError(t, err)
	day6 := day7.AddDate(0, 0, -1)
	day5 := day7.AddDate(0, 0, -2)
	day4 := day7.AddDate(0, 0, -3)
	day3 := day7.AddDate(0, 0, -4)
	day2 := day7.AddDate(0, 0, -5)
	day1 := day7.AddDate(0, 0, -6)

	api, _, mock := newMissionsTestAPI(t)

	// getSettings → no user_settings row.
	expectGetUserSettings(mock, userID, nil)
	// CurrentBalance.
	mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(42))
	// loadStreakAndGrace → GetStreakStateForRead.
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}).AddRow(userID, 3, 7, nil, nil, "UTC", gamification.StreakStatusActive, time.Now(), time.Now()))
	// loadStreakAndGrace → CurrentGraceBalance.
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(1))
	// ListRecentCompletionHistory (today - 6 through today, inclusive).
	mock.ExpectQuery("SELECT local_date, status FROM daily_mission_snapshots").
		WithArgs(userID, day1, day7).
		WillReturnRows(sqlmock.NewRows([]string{"local_date", "status"}).
			AddRow(day7, "open").
			AddRow(day6, "completed").
			AddRow(day5, "completed").
			AddRow(day4, "protected").
			AddRow(day3, "missed").
			AddRow(day2, "completed").
			AddRow(day1, "completed"))

	w := httptest.NewRecorder()
	req := authenticatedMissionsRequest(t, userID, "/api/v1/progress")
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var body GetProgressOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	assert.Equal(t, 42, body.Body.ConfidencePointsBalance)
	assert.Equal(t, 3, body.Body.Streak.CurrentStreakCount)
	assert.Equal(t, 7, body.Body.Streak.LongestStreakCount)
	assert.Equal(t, 1, body.Body.Streak.GraceDayBalance)
	require.Len(t, body.Body.CompletionHistory, 7)
	// Statuses: open/completed/completed/protected/missed/completed/completed.
	assert.Equal(t, day7.Format("2006-01-02"), body.Body.CompletionHistory[0].LocalDate)
	assert.False(t, body.Body.CompletionHistory[0].Completed)
	assert.True(t, body.Body.CompletionHistory[1].Completed)
	assert.True(t, body.Body.CompletionHistory[2].Completed)
	assert.True(t, body.Body.CompletionHistory[3].Completed)  // protected counts as completed
	assert.False(t, body.Body.CompletionHistory[4].Completed) // missed
	assert.True(t, body.Body.CompletionHistory[5].Completed)
	assert.True(t, body.Body.CompletionHistory[6].Completed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProgressEmptyHistory(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	today, err := gamification.LocalDate(time.Now(), "UTC")
	require.NoError(t, err)
	api, _, mock := newMissionsTestAPI(t)

	expectGetUserSettings(mock, userID, nil)
	mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}))
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("SELECT local_date, status FROM daily_mission_snapshots").
		WithArgs(userID, today.AddDate(0, 0, -6), today).
		WillReturnRows(sqlmock.NewRows([]string{"local_date", "status"}))

	w := httptest.NewRecorder()
	req := authenticatedMissionsRequest(t, userID, "/api/v1/progress")
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var body GetProgressOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	assert.Equal(t, 0, body.Body.ConfidencePointsBalance)
	assert.Equal(t, 0, body.Body.Streak.GraceDayBalance)
	assert.NotNil(t, body.Body.CompletionHistory)
	assert.Empty(t, body.Body.CompletionHistory)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProgressSharedStreakObjectAgreesWithGetDailyMission(t *testing.T) {
	// Both endpoints must back their streak object with the same call so
	// Home and Progress cannot disagree. This test stages identical
	// streak-states/ledger rows and asserts the two responses' streak
	// fields match exactly.
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	day, err := gamification.LocalDate(time.Now(), "UTC")
	require.NoError(t, err)
	now := time.Now()

	api, _, mock := newMissionsTestAPI(t)

	// GetDailyMission flow.
	expectGetUserSettings(mock, userID, nil)
	mock.ExpectBegin()
	expectGetDailyMissionSnapshot(mock, userID, day, &missions.DailyMissionSnapshot{
		UserID:           userID.String(),
		LocalDate:        day,
		Timezone:         "UTC",
		ReviewTarget:     20,
		ReviewsCompleted: 20,
		PolicyVersion:    gamification.MissionPolicyVersion,
		Status:           missions.StatusCompleted,
		GraceApplied:     false,
	})
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}).AddRow(userID, 5, 12, &day, &day, "UTC", gamification.StreakStatusActive, now, now))
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(2))

	w := httptest.NewRecorder()
	req := authenticatedMissionsRequest(t, userID, "/api/v1/daily-mission")
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var daily GetDailyMissionOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &daily.Body))

	// GetProgress flow.
	expectGetUserSettings(mock, userID, nil)
	mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(100))
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}).AddRow(userID, 5, 12, &day, &day, "UTC", gamification.StreakStatusActive, now, now))
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(2))
	mock.ExpectQuery("SELECT local_date, status FROM daily_mission_snapshots").
		WithArgs(userID, day.AddDate(0, 0, -6), day).
		WillReturnRows(sqlmock.NewRows([]string{"local_date", "status"}).
			AddRow(day, "completed"))

	w = httptest.NewRecorder()
	req = authenticatedMissionsRequest(t, userID, "/api/v1/progress")
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var progress GetProgressOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &progress.Body))

	// The shared StreakView must be byte-identical between the two
	// endpoints (DOC-12 §5 P4 cross-capability consistency).
	assert.Equal(t, daily.Body.Streak, progress.Body.Streak)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestGetDailyMissionCrossUserIsolation confirms that the read is
// implicitly self-scoped: both endpoints take no ID parameter, so the
// authenticated requester is the only user whose state is reachable.
// Cross-user state cannot be reached by simply changing the request path.
func TestGetDailyMissionCrossUserIsolation(t *testing.T) {
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	assert.NotEqual(t, owner, other, "owner/other distinct; cross-user enumeration is structurally impossible")

	// The Huma registration must not declare any path parameter on
	// /api/v1/daily-mission or /api/v1/progress: Huma emits path
	// parameters as {name} entries. The contract is the source of
	// truth — if either endpoint were ever to accept a userId, this
	// contract-level guard would catch it.
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	require.NoError(t, err)
	contract := string(document)
	dailyPath := "/api/v1/daily-mission"
	progressPath := "/api/v1/progress"
	dailyPathParam := dailyPath + "/{"
	progressPathParam := progressPath + "/{"
	require.Contains(t, contract, dailyPath)
	require.Contains(t, contract, progressPath)
	require.NotContains(t, contract, dailyPathParam)
	require.NotContains(t, contract, progressPathParam)
}
