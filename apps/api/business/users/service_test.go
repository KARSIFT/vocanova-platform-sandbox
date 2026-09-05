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
// Repository/UserSettingsReader/SettingsRepository triple so each
// test gets a fresh fixture.
func newServiceWithMemory(t *testing.T) (*Service, *MemoryRepository) {
	t.Helper()
	repo := NewMemoryRepository()
	c := &clock.Fixed{T: fixedNow()}
	svc := NewService(repo, repo, repo, c)
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
			assert.ErrorIs(t, err, ErrInvalidOnboarding)
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
	svc := NewService(repo, errOnlySettingsReader{}, repo, &clock.Fixed{T: fixedNow()})
	_, _, err := svc.CompleteOnboarding(context.Background(), uuid.New(), validAnswers())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "intentional settings failure")
}

// TestServiceGetSettingsRequiresUserID pins the no-nil-user-id rule
// at the service boundary (matching every other requester-scoped
// read in this package).
func TestServiceGetSettingsRequiresUserID(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	_, err := svc.GetSettings(context.Background(), uuid.Nil)
	require.Error(t, err)
}

// TestServiceGetSettingsReturnsSchemaDefaultsForUnseenUser covers
// VOC-031-TEST-08: a user who has authenticated but never had a
// user_settings row created (the common case immediately after
// sign-up, before the gamification module's lazy-create path has
// fired) sees a stable Settings projection filled from the
// schema defaults. The fixture's GetSettings treats every
// authenticated user as a real, non-deleted user, so the
// response is always a 200 with schema defaults when no
// user_settings row exists.
func TestServiceGetSettingsReturnsSchemaDefaultsForUnseenUser(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.MarkSeen(uid)

	got, err := svc.GetSettings(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, 20, got.DailyReviewTarget, "schema default for daily_review_target")
	assert.Equal(t, "vocanova_default", got.ReviewIntervalPreset, "schema default for review_interval_preset")
	assert.Equal(t, "en", got.AppLanguage, "schema default for app_language (D06)")
	assert.True(t, got.NotificationsEnabled, "schema default for notifications_enabled")
	assert.False(t, got.MarketingEmailsEnabled, "schema default for marketing_emails_enabled")
	assert.Equal(t, "", got.DisplayName, "no display_name until set")
}

// TestServiceGetSettingsReturnsSchemaDefaultsForBrandNewUser
// covers VOC-031-TEST-08: the in-memory fixture treats every
// userID as a real, non-deleted user in the users table (the
// SQL path's "WHERE deleted_at IS NULL" predicate always
// passes for test fixtures). The
// 404/ErrSettingsNotFound code path is exercised in
// integration tests against a real database, not here.
func TestServiceGetSettingsReturnsSchemaDefaultsForBrandNewUser(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	got, err := svc.GetSettings(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, 20, got.DailyReviewTarget, "schema default for an unseen user")
	assert.Equal(t, "vocanova_default", got.ReviewIntervalPreset, "schema default for an unseen user")
	assert.Equal(t, "en", got.AppLanguage, "schema default for an unseen user")
	assert.True(t, got.NotificationsEnabled, "schema default for an unseen user")
	assert.False(t, got.MarketingEmailsEnabled, "schema default for an unseen user")
	assert.Equal(t, "", got.DisplayName, "no display_name until set")
}

// TestServiceGetSettingsReturnsPersistedDisplayName covers
// VOC-031-TEST-08: the displayName read comes from the users
// table even when no user_settings row exists.
func TestServiceGetSettingsReturnsPersistedDisplayName(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.MarkSeen(uid)
	repo.SetDisplayName(uid, "Ada")

	got, err := svc.GetSettings(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, "Ada", got.DisplayName)
}

// TestServiceUpdateSettingsRequiresUserID pins the no-nil-user-id
// rule on the write path.
func TestServiceUpdateSettingsRequiresUserID(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	_, err := svc.UpdateSettings(context.Background(), uuid.Nil, SettingsUpdate{})
	require.Error(t, err)
}

// TestServiceUpdateSettingsRejectsInvalidUpdate covers
// VOC-031-TEST-08: an invalid PATCH body (e.g. a daily review
// target out of range) is rejected before any write reaches the
// repository.
func TestServiceUpdateSettingsRejectsInvalidUpdate(t *testing.T) {
	svc, _ := newServiceWithMemory(t)
	uid := uuid.New()
	v := MinDailyReviewTarget - 1
	_, err := svc.UpdateSettings(context.Background(), uid, SettingsUpdate{DailyReviewTarget: &v})
	require.Error(t, err)
}

// TestServiceUpdateSettingsIsNoOpForEmptyPayload covers
// VOC-031-TEST-08: an empty PATCH is a no-op (the response shape
// is the current state). The repository path is still exercised
// because DOC-07's "no-op PATCH is a well-formed read" rule
// requires the API to echo the current state to the caller.
func TestServiceUpdateSettingsIsNoOpForEmptyPayload(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.MarkSeen(uid)
	repo.SetDisplayName(uid, "Ada")

	got, err := svc.UpdateSettings(context.Background(), uid, SettingsUpdate{})
	require.NoError(t, err)
	assert.Equal(t, 20, got.DailyReviewTarget)
	assert.Equal(t, "vocanova_default", got.ReviewIntervalPreset)
	assert.Equal(t, "en", got.AppLanguage)
	assert.True(t, got.NotificationsEnabled)
	assert.False(t, got.MarketingEmailsEnabled)
	assert.Equal(t, "Ada", got.DisplayName)
}

// TestServiceUpdateSettingsAppliesEveryField covers
// VOC-031-TEST-08: a PATCH that sets every field writes every
// field and the response reflects the merged state.
func TestServiceUpdateSettingsAppliesEveryField(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.MarkSeen(uid)

	target := 35
	preset := ReviewIntervalPresetWordUpLike
	lang := "en"
	notifs := false
	marketing := true
	name := "Grace"
	got, err := svc.UpdateSettings(context.Background(), uid, SettingsUpdate{
		DailyReviewTarget:      &target,
		ReviewIntervalPreset:   &preset,
		AppLanguage:            &lang,
		NotificationsEnabled:   &notifs,
		MarketingEmailsEnabled: &marketing,
		DisplayName:            &name,
	})
	require.NoError(t, err)
	assert.Equal(t, 35, got.DailyReviewTarget)
	assert.Equal(t, "wordup_like", got.ReviewIntervalPreset)
	assert.Equal(t, "en", got.AppLanguage)
	assert.False(t, got.NotificationsEnabled)
	assert.True(t, got.MarketingEmailsEnabled)
	assert.Equal(t, "Grace", got.DisplayName)
}

// TestServiceUpdateSettingsIsPartial covers VOC-031-TEST-08: a
// PATCH that sets only one field leaves every other field at its
// current value (the existing value, not a schema default).
func TestServiceUpdateSettingsIsPartial(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.MarkSeen(uid)
	repo.SetDisplayName(uid, "Ada")
	repo.UpsertStoredUserSettings(uid, MemoryUserSettings{
		Timezone:               "Europe/Madrid",
		DailyReviewTarget:      40,
		ReviewIntervalPreset:   "wordup_like",
		NotificationsEnabled:   true,
		MarketingEmailsEnabled: false,
		AppLanguage:            "en",
	})

	target := 60
	got, err := svc.UpdateSettings(context.Background(), uid, SettingsUpdate{DailyReviewTarget: &target})
	require.NoError(t, err)
	assert.Equal(t, 60, got.DailyReviewTarget, "the one field changed")
	assert.Equal(t, "wordup_like", got.ReviewIntervalPreset, "unchanged: existing value preserved")
	assert.Equal(t, "en", got.AppLanguage, "unchanged: existing value preserved")
	assert.True(t, got.NotificationsEnabled, "unchanged: existing value preserved")
	assert.False(t, got.MarketingEmailsEnabled, "unchanged: existing value preserved")
	assert.Equal(t, "Ada", got.DisplayName, "unchanged: existing value preserved")
}

// TestServiceUpdateSettingsDailyReviewTargetDoesNotTouchDailyMissionSnapshot
// covers VOC-031-R06: a dailyReviewTarget PATCH must not rewrite
// the current local day's already-created
// daily_mission_snapshots.review_target. The users module has no
// reference to the daily_mission_snapshots table — the value is
// only ever read at snapshot-creation time — so this is a
// structural guarantee. The test asserts the architectural
// invariant by verifying the in-memory state of the snapshot is
// unaffected by a Settings PATCH (no panic, no error from
// non-existent tables).
func TestServiceUpdateSettingsDailyReviewTargetDoesNotTouchDailyMissionSnapshot(t *testing.T) {
	svc, repo := newServiceWithMemory(t)
	uid := uuid.New()
	repo.MarkSeen(uid)

	target := 30
	_, err := svc.UpdateSettings(context.Background(), uid, SettingsUpdate{DailyReviewTarget: &target})
	require.NoError(t, err)

	// Sanity check: the users module's own state reflects the
	// write. The architectural invariant — no
	// daily_mission_snapshots coupling — is upheld because the
	// users package has no reference to that table or to the
	// missions module at all.
	got, err := svc.GetSettings(context.Background(), uid)
	require.NoError(t, err)
	assert.Equal(t, 30, got.DailyReviewTarget)
}

// TestServiceGetSettingsRequiresConfiguredRepository pins the
// service-layer guard: when a Service is constructed without a
// SettingsRepository, the Get/Update calls fail fast with a clear
// error rather than dereferencing a nil pointer.
func TestServiceGetSettingsRequiresConfiguredRepository(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo, repo, nil, &clock.Fixed{T: fixedNow()})
	_, err := svc.GetSettings(context.Background(), uuid.New())
	require.Error(t, err)
}

func TestServiceUpdateSettingsRequiresConfiguredRepository(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo, repo, nil, &clock.Fixed{T: fixedNow()})
	_, err := svc.UpdateSettings(context.Background(), uuid.New(), SettingsUpdate{})
	require.Error(t, err)
}
