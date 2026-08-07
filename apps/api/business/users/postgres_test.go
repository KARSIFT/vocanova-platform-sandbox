package users

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// userSettingsInsertColumnsPattern matches the user_settings INSERT column
// list only: the regression this guards is the omission of
// created_at/updated_at, not the placeholder numbering or timestamp source the
// fix chooses for them. sqlmock collapses each whitespace run in the actual
// statement to a single space, so the pattern tolerates optional spaces around
// the parentheses rather than assuming how the SQL literal is wrapped.
const userSettingsInsertColumnsPattern = `INSERT INTO user_settings \(\s*id,\s*user_id,\s*timezone,\s*daily_review_target,\s*created_at,\s*updated_at\s*\)`
const updateSettingsInsertColumnsPattern = `INSERT INTO user_settings \(\s*id,\s*user_id,\s*timezone,\s*daily_review_target,\s*review_interval_preset,\s*notifications_enabled,\s*marketing_emails_enabled,\s*app_language,\s*created_at,\s*updated_at\s*\)`

func TestPostgreSQLRepositoryCompleteOnboardingFreshUserSettingsInsertSuppliesTimestamps(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000111")
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	answers := OnboardingAnswers{
		EnglishLevel:      "b1",
		NativeLanguage:    "es",
		LearningGoal:      "general",
		MainUseCase:       "daily_life",
		DailyReviewTarget: 30,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET onboarding_status = 'completed'").
		WithArgs(userID, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO user_onboarding_profiles").
		WithArgs(
			sqlmock.AnyArg(), userID, answers.EnglishLevel, answers.NativeLanguage,
			answers.LearningGoal, answers.MainUseCase, answers.DailyReviewTarget, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(userSettingsInsertColumnsPattern).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "daily_review_target"}).AddRow(userID, answers.DailyReviewTarget))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT u.onboarding_status,").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"onboarding_status",
			"coalesce",
			"coalesce",
			"coalesce",
			"coalesce",
			"coalesce",
			"completed_at",
			"coalesce",
			"coalesce",
		}).AddRow(
			OnboardingStatusCompleted,
			answers.EnglishLevel,
			answers.NativeLanguage,
			answers.LearningGoal,
			answers.MainUseCase,
			answers.DailyReviewTarget,
			now,
			now,
			now,
		))

	profile, stored, err := repo.CompleteOnboarding(t.Context(), userID, answers, now)
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, OnboardingStatusCompleted, profile.Status)
	assert.True(t, stored.Stored)
	assert.Equal(t, answers.DailyReviewTarget, stored.DailyReviewTarget)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryUpdateSettingsFreshInsertSuppliesTimestamps(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000222")
	now := time.Date(2026, 8, 7, 11, 0, 0, 0, time.UTC)
	target := 25

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT daily_review_target, review_interval_preset,").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"daily_review_target",
			"review_interval_preset",
			"notifications_enabled",
			"marketing_emails_enabled",
			"app_language",
		}))
	mock.ExpectQuery(updateSettingsInsertColumnsPattern).
		WithArgs(
			sqlmock.AnyArg(), userID, target, schemaReviewIntervalPresetDefault,
			schemaNotificationsEnabledDefault, schemaMarketingEmailsEnabledDefault,
			schemaAppLanguageDefault, now,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"daily_review_target",
			"review_interval_preset",
			"notifications_enabled",
			"marketing_emails_enabled",
			"app_language",
		}).AddRow(
			target,
			schemaReviewIntervalPresetDefault,
			schemaNotificationsEnabledDefault,
			schemaMarketingEmailsEnabledDefault,
			schemaAppLanguageDefault,
		))
	mock.ExpectExec("UPDATE users SET display_name =").
		WithArgs(userID, "Ana", now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	displayName := "Ana"
	updated, err := repo.UpdateSettings(t.Context(), userID, SettingsUpdate{
		DailyReviewTarget: &target,
		DisplayName:       &displayName,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, target, updated.DailyReviewTarget)
	assert.Equal(t, displayName, updated.DisplayName)
	require.NoError(t, mock.ExpectationsWereMet())
}
