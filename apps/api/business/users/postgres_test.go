package users

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	mock.ExpectQuery("INSERT INTO user_settings \\(id, user_id, timezone, daily_review_target, created_at, updated_at\\)").
		WithArgs(sqlmock.AnyArg(), userID, answers.DailyReviewTarget, now, now, SchemaDailyReviewTargetDefault).
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
