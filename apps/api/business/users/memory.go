package users

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository is a deterministic in-memory Repository for
// service and route tests. It also implements UserSettingsReader
// against the same in-memory user_settings slice, so tests can wire a
// single instance to both Repository and UserSettingsReader.
type MemoryRepository struct {
	mu           sync.Mutex
	profiles     map[uuid.UUID]*MemoryOnboardingProfile
	settings     map[uuid.UUID]*MemoryUserSettings
	displayNames map[uuid.UUID]string // userID -> display_name
	seenUsers    map[uuid.UUID]struct{}
	onboarding   map[uuid.UUID]string // userID -> onboarding_status
}

// MemoryOnboardingProfile is the in-memory shape of a
// user_onboarding_profiles row.
type MemoryOnboardingProfile struct {
	UserID            uuid.UUID
	EnglishLevel      string
	NativeLanguage    string
	LearningGoal      string
	MainUseCase       string
	DailyReviewTarget int
	CompletedAt       time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// MemoryUserSettings is the in-memory shape of the user_settings
// fields the users module needs to read for the D04 seed decision
// and the API layer needs for the GET /api/v1/onboarding response.
type MemoryUserSettings struct {
	UserID                 uuid.UUID
	Timezone               string
	DailyReviewTarget      int
	ReviewIntervalPreset   string
	NotificationsEnabled   bool
	MarketingEmailsEnabled bool
	AppLanguage            string
}

// NewMemoryRepository creates an empty in-memory repository. It can
// be seeded via UpsertStoredUserSettings, SetDisplayName, and
// SetOnboardingStatus.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		profiles:     make(map[uuid.UUID]*MemoryOnboardingProfile),
		settings:     make(map[uuid.UUID]*MemoryUserSettings),
		displayNames: make(map[uuid.UUID]string),
		seenUsers:    make(map[uuid.UUID]struct{}),
		onboarding:   make(map[uuid.UUID]string),
	}
}

// GetOnboarding returns the requester's profile plus the current
// users.onboarding_status. When no row exists, returns
// (status, nil, ErrOnboardingNotFound).
func (r *MemoryRepository) GetOnboarding(ctx context.Context, userID uuid.UUID) (*OnboardingProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	status, ok := r.onboarding[userID]
	if !ok {
		status = OnboardingStatusNotStarted
	}
	p, ok := r.profiles[userID]
	if !ok {
		return &OnboardingProfile{
			UserID: userID,
			Status: status,
		}, ErrOnboardingNotFound
	}
	return &OnboardingProfile{
		UserID:            userID,
		Status:            status,
		EnglishLevel:      p.EnglishLevel,
		NativeLanguage:    p.NativeLanguage,
		LearningGoal:      p.LearningGoal,
		MainUseCase:       p.MainUseCase,
		DailyReviewTarget: p.DailyReviewTarget,
		CompletedAt:       timePtr(p.CompletedAt),
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}, nil
}

// CompleteOnboarding applies the D04 rule against the in-memory
// user_settings slice and writes the profile. It is atomic from the
// point of view of the MemoryRepository's own lock.
func (r *MemoryRepository) CompleteOnboarding(ctx context.Context, userID uuid.UUID, answers OnboardingAnswers, now time.Time) (*OnboardingProfile, StoredUserSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.profiles[userID]; ok {
		// Idempotent: the same answers are accepted as a no-op;
		// differing answers are rejected (mirrors DB ON CONFLICT
		// behavior). Onboarding is single-submit per D02.
		if existing.EnglishLevel != answers.EnglishLevel ||
			existing.NativeLanguage != answers.NativeLanguage ||
			existing.LearningGoal != answers.LearningGoal ||
			existing.MainUseCase != answers.MainUseCase ||
			existing.DailyReviewTarget != answers.DailyReviewTarget {
			return nil, StoredUserSettings{}, ErrOnboardingConflict
		}
	} else {
		r.profiles[userID] = &MemoryOnboardingProfile{
			UserID:            userID,
			EnglishLevel:      answers.EnglishLevel,
			NativeLanguage:    answers.NativeLanguage,
			LearningGoal:      answers.LearningGoal,
			MainUseCase:       answers.MainUseCase,
			DailyReviewTarget: answers.DailyReviewTarget,
			CompletedAt:       now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
	}

	r.onboarding[userID] = OnboardingStatusCompleted

	settings, ok := r.settings[userID]
	stored := StoredUserSettings{Stored: ok}
	if ok {
		stored.DailyReviewTarget = settings.DailyReviewTarget
	}
	decision := ResolveDailyReviewTargetSeed(stored, answers.DailyReviewTarget)
	switch decision.Action {
	case SeedCreateRow:
		r.settings[userID] = &MemoryUserSettings{
			UserID:                 userID,
			Timezone:               "UTC",
			DailyReviewTarget:      decision.Value,
			ReviewIntervalPreset:   "vocanova_default",
			NotificationsEnabled:   true,
			MarketingEmailsEnabled: false,
			AppLanguage:            "en",
		}
	case SeedOverwriteDefault:
		settings.DailyReviewTarget = decision.Value
	case SeedPreserveExisting:
		// no-op
	}

	final, ok := r.settings[userID]
	finalStored := StoredUserSettings{Stored: ok}
	if ok {
		finalStored.DailyReviewTarget = final.DailyReviewTarget
	}

	p := r.profiles[userID]
	return &OnboardingProfile{
		UserID:            userID,
		Status:            OnboardingStatusCompleted,
		EnglishLevel:      p.EnglishLevel,
		NativeLanguage:    p.NativeLanguage,
		LearningGoal:      p.LearningGoal,
		MainUseCase:       p.MainUseCase,
		DailyReviewTarget: p.DailyReviewTarget,
		CompletedAt:       timePtr(p.CompletedAt),
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}, finalStored, nil
}

// SetOnboardingStatus updates the in-memory users.onboarding_status.
func (r *MemoryRepository) SetOnboardingStatus(ctx context.Context, userID uuid.UUID, status string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onboarding[userID] = status
	return nil
}

// UpsertStoredUserSettings is a test helper to seed or modify the
// user_settings slice the users module reads. It is not part of the
// Repository contract.
func (r *MemoryRepository) UpsertStoredUserSettings(userID uuid.UUID, s MemoryUserSettings) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s.UserID = userID
	r.settings[userID] = &s
	r.seenUsers[userID] = struct{}{}
}

// GetStoredUserSettings implements UserSettingsReader against the
// in-memory user_settings slice.
func (r *MemoryRepository) GetStoredUserSettings(ctx context.Context, userID uuid.UUID) (StoredUserSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.settings[userID]
	if !ok {
		return StoredUserSettings{Stored: false}, nil
	}
	return StoredUserSettings{
		Stored:            true,
		DailyReviewTarget: s.DailyReviewTarget,
	}, nil
}

// ProfilesOrderedByCompletion returns a stable, test-friendly view of
// all in-memory profiles (sorted by CompletedAt then UserID). Used
// only by tests for assertions; never on the request path.
func (r *MemoryRepository) ProfilesOrderedByCompletion() []MemoryOnboardingProfile {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]MemoryOnboardingProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CompletedAt.Equal(out[j].CompletedAt) {
			return out[i].UserID.String() < out[j].UserID.String()
		}
		return out[i].CompletedAt.Before(out[j].CompletedAt)
	})
	return out
}

// markSeen records that a user exists in the in-memory fixture.
// GetSettings uses it to distinguish "user exists with no row yet"
// (returns schema defaults) from "user does not exist" (returns
// ErrSettingsNotFound). Marking on every Get/Update mirrors the
// SQL semantics where the LEFT JOIN on users is the source of
// truth for existence.
func (r *MemoryRepository) markSeen(userID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seenUsers[userID] = struct{}{}
}

// GetSettings returns the requester's Settings projection, filling
// any unset fields from the user_settings schema defaults so the
// response shape is stable. The in-memory fixture treats every
// userID as a real, non-deleted user in the users table (the SQL
// path's "WHERE deleted_at IS NULL" predicate always passes for
// test fixtures) so the response is always 200 with schema
// defaults when no user_settings row exists yet. The
// 404/ErrSettingsNotFound code path is exercised in integration
// tests against a real database, not in this fixture.
func (r *MemoryRepository) GetSettings(ctx context.Context, userID uuid.UUID) (Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seenUsers[userID] = struct{}{}
	s := r.settings[userID]
	settings := Settings{
		DailyReviewTarget:      schemaDailyReviewTargetDefaultInt,
		ReviewIntervalPreset:   schemaReviewIntervalPresetDefault,
		AppLanguage:            schemaAppLanguageDefault,
		NotificationsEnabled:   schemaNotificationsEnabledDefault,
		MarketingEmailsEnabled: schemaMarketingEmailsEnabledDefault,
		DisplayName:            r.displayNames[userID],
	}
	if s != nil {
		settings.DailyReviewTarget = s.DailyReviewTarget
		settings.ReviewIntervalPreset = s.ReviewIntervalPreset
		settings.AppLanguage = s.AppLanguage
		settings.NotificationsEnabled = s.NotificationsEnabled
		settings.MarketingEmailsEnabled = s.MarketingEmailsEnabled
	}
	return settings, nil
}

// UpdateSettings applies a partial update to the in-memory
// user_settings slice and (when supplied) users.display_name,
// then returns the merged Settings projection. The first-ever
// write creates a row with schema defaults filled in. Validation
// is the caller's responsibility (the service runs it before
// reaching this method), matching the SQL implementation's
// contract.
func (r *MemoryRepository) UpdateSettings(ctx context.Context, userID uuid.UUID, update SettingsUpdate, now time.Time) (Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seenUsers[userID] = struct{}{}
	s, ok := r.settings[userID]
	if !ok {
		s = &MemoryUserSettings{
			UserID:                 userID,
			Timezone:               "UTC",
			DailyReviewTarget:      schemaDailyReviewTargetDefaultInt,
			ReviewIntervalPreset:   schemaReviewIntervalPresetDefault,
			NotificationsEnabled:   schemaNotificationsEnabledDefault,
			MarketingEmailsEnabled: schemaMarketingEmailsEnabledDefault,
			AppLanguage:            schemaAppLanguageDefault,
		}
		r.settings[userID] = s
	}
	if update.DailyReviewTarget != nil {
		s.DailyReviewTarget = *update.DailyReviewTarget
	}
	if update.ReviewIntervalPreset != nil {
		s.ReviewIntervalPreset = *update.ReviewIntervalPreset
	}
	if update.AppLanguage != nil {
		s.AppLanguage = *update.AppLanguage
	}
	if update.NotificationsEnabled != nil {
		s.NotificationsEnabled = *update.NotificationsEnabled
	}
	if update.MarketingEmailsEnabled != nil {
		s.MarketingEmailsEnabled = *update.MarketingEmailsEnabled
	}
	if update.DisplayName != nil {
		r.displayNames[userID] = *update.DisplayName
	}
	return Settings{
		DailyReviewTarget:      s.DailyReviewTarget,
		ReviewIntervalPreset:   s.ReviewIntervalPreset,
		AppLanguage:            s.AppLanguage,
		NotificationsEnabled:   s.NotificationsEnabled,
		MarketingEmailsEnabled: s.MarketingEmailsEnabled,
		DisplayName:            r.displayNames[userID],
	}, nil
}

// SetDisplayName seeds users.display_name for a fixture user. It
// is a test helper, not part of the SettingsRepository contract.
func (r *MemoryRepository) SetDisplayName(userID uuid.UUID, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seenUsers[userID] = struct{}{}
	r.displayNames[userID] = name
}

// MarkSeen records a userID as existing without changing any
// stored state. Tests that want to exercise the no-row-yet path
// of GetSettings use it to make the user "seen" without writing
// a user_settings row. Not part of the contract.
func (r *MemoryRepository) MarkSeen(userID uuid.UUID) {
	r.markSeen(userID)
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
