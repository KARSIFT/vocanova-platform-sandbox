package users

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestLearningPreferencesFallbackRetryAndConflictingIntent(t *testing.T) {
	r := NewMemoryRepository()
	svc := NewLearningPreferencesService(r)
	uid := uuid.New()
	ctx := t.Context()
	require.NoError(t, r.SetOnboardingStatus(ctx, uid, OnboardingStatusCompleted, time.Now()))
	p, err := svc.GetLearningPreferences(ctx, uid)
	require.NoError(t, err)
	require.Nil(t, p.LearningGoal)
	require.Nil(t, p.MainUseCase)
	require.Zero(t, p.Revision)
	first := LearningPreferencesUpdate{LearningGoal: "travel", MainUseCase: "travel"}
	p, err = svc.UpdateLearningPreferences(ctx, uid, first)
	require.NoError(t, err)
	require.Equal(t, int64(1), p.Revision)
	retry, err := svc.UpdateLearningPreferences(ctx, uid, first)
	require.NoError(t, err)
	require.Equal(t, p, retry)
	_, err = svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: "work", MainUseCase: "work"})
	require.ErrorIs(t, err, ErrLearningPreferencesConflict)
	newer, err := svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: "work", MainUseCase: "work", ExpectedRevision: 1})
	require.NoError(t, err)
	require.Equal(t, int64(2), newer.Revision)
	_, err = svc.UpdateLearningPreferences(ctx, uid, first)
	require.ErrorIs(t, err, ErrLearningPreferencesConflict)
	copy, err := svc.GetLearningPreferences(ctx, uid)
	require.NoError(t, err)
	*copy.MainUseCase = "social"
	actual, err := svc.GetLearningPreferences(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, "work", *actual.MainUseCase)
	_, err = svc.GetLearningPreferences(ctx, uuid.New())
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestLearningPreferencesPreservesOriginalOnboardingAndSettings(t *testing.T) {
	r := NewMemoryRepository()
	uid := uuid.New()
	ctx := t.Context()
	answers := OnboardingAnswers{EnglishLevel: "a2", NativeLanguage: "fa", LearningGoal: "general", MainUseCase: "daily_life", DailyReviewTarget: 37}
	original, settings, err := r.CompleteOnboarding(ctx, uid, answers, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	svc := NewLearningPreferencesService(r)
	p, err := svc.GetLearningPreferences(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, "general", *p.LearningGoal)
	require.Zero(t, p.Revision)
	p, err = svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: "general", MainUseCase: "daily_life"})
	require.NoError(t, err)
	require.Zero(t, p.Revision)
	_, err = svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: "conversation", MainUseCase: "social"})
	require.NoError(t, err)
	after, err := r.GetOnboarding(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, original, after)
	afterSettings, err := r.GetStoredUserSettings(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, settings, afterSettings)
	_, _, err = r.CompleteOnboarding(ctx, uid, answers, time.Now())
	require.NoError(t, err, "original same-intent onboarding remains replayable")
}

func TestLearningPreferencesValidationAndConcurrentCAS(t *testing.T) {
	r := NewMemoryRepository()
	svc := NewLearningPreferencesService(r)
	uid := uuid.New()
	ctx := t.Context()
	require.NoError(t, r.SetOnboardingStatus(ctx, uid, OnboardingStatusInProgress, time.Now()))
	_, err := svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: "work", MainUseCase: "work"})
	require.ErrorIs(t, err, ErrLearningPreferencesOnboarding)
	require.NoError(t, r.SetOnboardingStatus(ctx, uid, OnboardingStatusCompleted, time.Now()))
	for _, bad := range []LearningPreferencesUpdate{{LearningGoal: "invalid", MainUseCase: "work"}, {LearningGoal: "work", MainUseCase: "invalid"}, {LearningGoal: "work", MainUseCase: "work", ExpectedRevision: -1}} {
		_, err := svc.UpdateLearningPreferences(ctx, uid, bad)
		require.ErrorIs(t, err, ErrInvalidLearningPreferences)
	}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, focus := range []string{"work", "travel"} {
		wg.Add(1)
		go func(focus string) {
			defer wg.Done()
			_, err := svc.UpdateLearningPreferences(ctx, uid, LearningPreferencesUpdate{LearningGoal: focus, MainUseCase: focus})
			results <- err
		}(focus)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrLearningPreferencesConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	require.Equal(t, 1, success)
	require.Equal(t, 1, conflict)
}
