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
	mu         sync.Mutex
	profiles   map[uuid.UUID]*MemoryOnboardingProfile
	settings   map[uuid.UUID]*MemoryUserSettings
	onboarding map[uuid.UUID]string // userID -> onboarding_status
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
// be seeded via UpsertStoredUserSettings and SetOnboardingStatus.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		profiles:   make(map[uuid.UUID]*MemoryOnboardingProfile),
		settings:   make(map[uuid.UUID]*MemoryUserSettings),
		onboarding: make(map[uuid.UUID]string),
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

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
