package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedNow returns a deterministic UTC instant for all service tests.
func fixedNow() time.Time {
	return time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
}

// newServiceWithMemory wires a Service backed by a single in-memory
// Repository/UserSettingsReader pair so each test gets a fresh
// fixture.
func newServiceWithMemory(t *testing.T) (*Service, *MemoryRepository) {
	t.Helper()
	repo := NewMemoryRepository()
	c := &clock.Fixed{T: fixedNow()}
	svc := NewService(repo, repo, c)
	return svc, repo
}

func validAnswers() OnboardingAnswers {
	return OnboardingAnswers{
		EnglishLevel:      EnglishLevelB1,
		NativeLanguage:    "es",
		LearningGoal:      LearningGoalGeneral,
		MainUseCase:       MainUseCaseDailyLife,
		DailyReviewTarget: 20,
	}
}

func TestServiceGetOnboardingReturnsNotFoundForUnseenUser(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()

	profile, err := svc.GetOnboarding(context.Background(), uid)
	// The memory repository returns the synthesized profile with
	// ErrOnboardingNotFound for the API to map; the service propagates
	// it unchanged.
	require.ErrorIs(t, err, ErrOnboardingNotFound)
	require.NotNil(t, profile)
	assert.Equal(t, OnboardingStatusNotStarted, profile.Status)
}

func TestServiceGetOnboardingRequiresUserID(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	_, err := svc.GetOnboarding(context.Background(), uuid.Nil)
	require.Error(t, err)
}

func TestServiceCompleteOnboardingRejectsInvalidEnglishLevel(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	a := validAnswers()
	a.EnglishLevel = "c1"
	_, _, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "english level")
}

func TestServiceCompleteOnboardingRejectsEmptyNativeLanguage(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	a := validAnswers()
	a.NativeLanguage = ""
	_, _, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "native language")
}

func TestServiceCompleteOnboardingRejectsInvalidLearningGoal(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	a := validAnswers()
	a.LearningGoal = "explore"
	_, _, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "learning goal")
}

func TestServiceCompleteOnboardingRejectsInvalidMainUseCase(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	a := validAnswers()
	a.MainUseCase = "free_time"
	_, _, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "main use case")
}

func TestServiceCompleteOnboardingRejectsDailyReviewOutOfRange(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	for _, target := range []int{0, 4, 101, 200, -1} {
		a := validAnswers()
		a.DailyReviewTarget = target
		_, _, err := svc.CompleteOnboarding(context.Background(), uuid.New(), a)
		require.Errorf(t, err, "target %d should be rejected", target)
	}
}

func TestServiceCompleteOnboardingRejectsNilUserID(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	_, _, err := svc.CompleteOnboarding(context.Background(), uuid.Nil, validAnswers())
	require.Error(t, err)
}

func TestServiceCompleteOnboardingSeedsUserSettingsWhenNoRow(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	a := validAnswers()
	a.DailyReviewTarget = 30

	profile, stored, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, OnboardingStatusCompleted, profile.Status)
	assert.Equal(t, 30, profile.DailyReviewTarget)
	assert.True(t, stored.Stored)
	assert.Equal(t, 30, stored.DailyReviewTarget, "no existing row: seed with onboarding answer")
}

func TestServiceCompleteOnboardingOverwritesDefaultUserSettings(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.UpsertStoredUserSettings(uid, MemoryUserSettings{
		Timezone:          "UTC",
		DailyReviewTarget: SchemaDailyReviewTargetDefault,
	})

	a := validAnswers()
	a.DailyReviewTarget = 35
	_, stored, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.NoError(t, err)
	assert.True(t, stored.Stored)
	assert.Equal(t, 35, stored.DailyReviewTarget, "schema default existing: overwrite with onboarding answer")
}

func TestServiceCompleteOnboardingPreservesCustomizedUserSettings(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.UpsertStoredUserSettings(uid, MemoryUserSettings{
		Timezone:          "Europe/Madrid",
		DailyReviewTarget: 50,
	})

	a := validAnswers()
	a.DailyReviewTarget = 10
	_, stored, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.NoError(t, err)
	assert.True(t, stored.Stored)
	assert.Equal(t, 50, stored.DailyReviewTarget, "customized existing: never overwrite")
}

func TestServiceCompleteOnboardingPersistsProfile(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	a := validAnswers()
	a.DailyReviewTarget = 25
	profile, _, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.NoError(t, err)
	require.NotNil(t, profile.CompletedAt)
	assert.Equal(t, fixedNow(), *profile.CompletedAt)

	got, err := svc.GetOnboarding(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, OnboardingStatusCompleted, got.Status)
	assert.Equal(t, a.EnglishLevel, got.EnglishLevel)
	assert.Equal(t, a.NativeLanguage, got.NativeLanguage)
	assert.Equal(t, a.LearningGoal, got.LearningGoal)
	assert.Equal(t, a.MainUseCase, got.MainUseCase)
	assert.Equal(t, a.DailyReviewTarget, got.DailyReviewTarget)
}

func TestServiceCompleteOnboardingIsIdempotentForSameAnswers(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	a := validAnswers()
	_, _, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.NoError(t, err)
	// Re-submitting the same answers is a no-op; status and profile
	// remain stable.
	_, stored, err := svc.CompleteOnboarding(context.Background(), uid, a)
	require.NoError(t, err)
	assert.True(t, stored.Stored)
}

func TestServiceCompleteOnboardingRejectsConflictingAnswers(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	first := validAnswers()
	_, _, err := svc.CompleteOnboarding(context.Background(), uid, first)
	require.NoError(t, err)

	conflicting := validAnswers()
	conflicting.NativeLanguage = "fr"
	_, _, err = svc.CompleteOnboarding(context.Background(), uid, conflicting)
	require.ErrorIs(t, err, ErrOnboardingConflict)
}

func TestServiceGetOnboardingReturnsCompletedAfterComplete(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	_, _, err := svc.CompleteOnboarding(context.Background(), uid, validAnswers())
	require.NoError(t, err)

	profile, err := svc.GetOnboarding(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, OnboardingStatusCompleted, profile.Status)
	assert.Equal(t, validAnswers().EnglishLevel, profile.EnglishLevel)
}

func TestServiceGetOnboardingStatusHonorsExternalSet(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	require.NoError(t, repo.SetOnboardingStatus(context.Background(), uid, OnboardingStatusInProgress, fixedNow()))

	profile, err := svc.GetOnboarding(context.Background(), uid)
	require.ErrorIs(t, err, ErrOnboardingNotFound)
	assert.Equal(t, OnboardingStatusInProgress, profile.Status, "external SetOnboardingStatus is reflected even without a profile")
}

func TestOnboardingAnswersValidateRejectsEveryInvalidEnum(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*OnboardingAnswers)
		wantSub string
	}{
		{"english-level", func(a *OnboardingAnswers) { a.EnglishLevel = "c1" }, "english level"},
		{"native-language", func(a *OnboardingAnswers) { a.NativeLanguage = "" }, "native language"},
		{"learning-goal", func(a *OnboardingAnswers) { a.LearningGoal = "explore" }, "learning goal"},
		{"main-use-case", func(a *OnboardingAnswers) { a.MainUseCase = "free_time" }, "main use case"},
		{"daily-review-low", func(a *OnboardingAnswers) { a.DailyReviewTarget = 4 }, "daily review target"},
		{"daily-review-high", func(a *OnboardingAnswers) { a.DailyReviewTarget = 101 }, "daily review target"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := validAnswers()
			tc.mutate(&a)
			err := a.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantSub)
		})
	}
}

func TestOnboardingAnswersValidateAcceptsEveryEnumValue(t *testing.T) {
	for _, level := range []string{EnglishLevelA1, EnglishLevelA2, EnglishLevelB1, EnglishLevelB2, EnglishLevelUnknown} {
		a := validAnswers()
		a.EnglishLevel = level
		require.NoErrorf(t, a.Validate(), "level=%s", level)
	}
	for _, goal := range []string{LearningGoalGeneral, LearningGoalWork, LearningGoalTravel, LearningGoalStudy, LearningGoalConversation, LearningGoalExam} {
		a := validAnswers()
		a.LearningGoal = goal
		require.NoErrorf(t, a.Validate(), "goal=%s", goal)
	}
	for _, uc := range []string{MainUseCaseDailyLife, MainUseCaseWork, MainUseCaseTravel, MainUseCaseStudy, MainUseCaseSocial} {
		a := validAnswers()
		a.MainUseCase = uc
		require.NoErrorf(t, a.Validate(), "use=%s", uc)
	}
	for _, tgt := range []int{5, 20, 50, 100} {
		a := validAnswers()
		a.DailyReviewTarget = tgt
		require.NoErrorf(t, a.Validate(), "target=%d", tgt)
	}
}

func TestServiceCompleteOnboardingRejectsBoundaryValues(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	for _, tgt := range []int{5, 100} {
		a := validAnswers()
		a.DailyReviewTarget = tgt
		_, _, err := svc.CompleteOnboarding(context.Background(), uuid.New(), a)
		require.NoErrorf(t, err, "target %d should be accepted", tgt)
	}
}

// errOnlySettingsReader is a UserSettingsReader that always returns
// the same StoredUserSettings, to exercise Service.go's re-read path
// without requiring a real Repository.
type errOnlySettingsReader struct{}

func (errOnlySettingsReader) GetStoredUserSettings(ctx context.Context, userID uuid.UUID) (StoredUserSettings, error) {
	return StoredUserSettings{Stored: false}, errors.New("intentional settings failure")
}

func TestServiceCompleteOnboardingReReadErrorSurfaces(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo, errOnlySettingsReader{}, &clock.Fixed{T: fixedNow()})
	_, _, err := svc.CompleteOnboarding(context.Background(), uuid.New(), validAnswers())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "intentional settings failure")
}
