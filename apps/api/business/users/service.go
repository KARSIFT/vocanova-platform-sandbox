package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// Onboarding status values from the users.onboarding_status enum. The
// 'in_progress' value is the schema-defined middle state; the adopted
// VOC-031-D02 design treats the flow as single-submit/non-resumable, so
// 'in_progress' is never written by this package, but it is preserved as a
// valid value here and never validated away to keep the contract future-proof
// and consistent with the existing schema.
const (
	OnboardingStatusNotStarted = "not_started"
	OnboardingStatusInProgress = "in_progress"
	OnboardingStatusCompleted  = "completed"
)

// English level enum values (DOC-05 §6, user_onboarding_profiles check
// constraint).
const (
	EnglishLevelA1      = "a1"
	EnglishLevelA2      = "a2"
	EnglishLevelB1      = "b1"
	EnglishLevelB2      = "b2"
	EnglishLevelUnknown = "unknown"
)

// Learning goal enum values.
const (
	LearningGoalGeneral      = "general"
	LearningGoalWork         = "work"
	LearningGoalTravel       = "travel"
	LearningGoalStudy        = "study"
	LearningGoalConversation = "conversation"
	LearningGoalExam         = "exam"
)

// Main use case enum values.
const (
	MainUseCaseDailyLife = "daily_life"
	MainUseCaseWork      = "work"
	MainUseCaseTravel    = "travel"
	MainUseCaseStudy     = "study"
	MainUseCaseSocial    = "social"
)

// Daily review target bounds (mirrors DOC-05 §6's 5..100 check on both
// user_onboarding_profiles and user_settings).
const (
	MinDailyReviewTarget = 5
	MaxDailyReviewTarget = 100
)

// OnboardingProfile is the persisted, requester-owned onboarding submission
// record (DOC-05 §6). It is the public projection for the onboarding API and
// the domain input to the D04 seed function.
type OnboardingProfile struct {
	UserID            uuid.UUID
	Status            string
	EnglishLevel      string
	NativeLanguage    string
	LearningGoal      string
	MainUseCase       string
	DailyReviewTarget int
	CompletedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// OnboardingAnswers is the inbound payload for CompleteOnboarding. It
// deliberately omits the UserID (always taken from the authenticated
// requester) and the Status/CompletedAt (always derived server-side).
type OnboardingAnswers struct {
	EnglishLevel      string
	NativeLanguage    string
	LearningGoal      string
	MainUseCase       string
	DailyReviewTarget int
}

// Validate enforces the DOC-05 §6 enum and range checks at the service
// boundary. The DB check constraints provide a second layer; this
// function returns a stable 400 (mapped by the API layer) so a caller
// receives a clear "invalid" signal before the DB rejects.
func (a OnboardingAnswers) Validate() error {
	switch a.EnglishLevel {
	case EnglishLevelA1, EnglishLevelA2, EnglishLevelB1, EnglishLevelB2, EnglishLevelUnknown:
	default:
		return fmt.Errorf("%w: invalid english level %q", ErrInvalidOnboarding, a.EnglishLevel)
	}
	if a.NativeLanguage == "" {
		return fmt.Errorf("%w: native language is required", ErrInvalidOnboarding)
	}
	switch a.LearningGoal {
	case LearningGoalGeneral, LearningGoalWork, LearningGoalTravel,
		LearningGoalStudy, LearningGoalConversation, LearningGoalExam:
	default:
		return fmt.Errorf("%w: invalid learning goal %q", ErrInvalidOnboarding, a.LearningGoal)
	}
	switch a.MainUseCase {
	case MainUseCaseDailyLife, MainUseCaseWork, MainUseCaseTravel,
		MainUseCaseStudy, MainUseCaseSocial:
	default:
		return fmt.Errorf("%w: invalid main use case %q", ErrInvalidOnboarding, a.MainUseCase)
	}
	if a.DailyReviewTarget < MinDailyReviewTarget || a.DailyReviewTarget > MaxDailyReviewTarget {
		return fmt.Errorf("%w: daily review target %d out of range [%d,%d]", ErrInvalidOnboarding, a.DailyReviewTarget, MinDailyReviewTarget, MaxDailyReviewTarget)
	}
	return nil
}

// UserSettingsReader is the minimum view the D04 seed-eligibility decision
// needs about the learner's user_settings row. The users module's
// caller is responsible for fetching this from user_settings; the
// users.Service treats it as a read-only interface so it does not own
// or modify user_settings. This matches the pattern documented in
// specification.md's D04: T00's seed function is the source of truth
// for the decision, but it operates on the StoredUserSettings view.
type UserSettingsReader interface {
	GetStoredUserSettings(ctx context.Context, userID uuid.UUID) (StoredUserSettings, error)
}

// Repository is the persistence boundary for onboarding reads/writes. It is
// the contract between the users.Service and the SQL/memory backends. The
// Complete method is expected to be atomic and to set users.onboarding_status
// to 'completed' inside the same transaction that persists the profile.
type Repository interface {
	// GetOnboarding returns the requester's onboarding profile plus the
	// current users.onboarding_status. Returns ErrOnboardingNotFound when
	// no row exists for the user.
	GetOnboarding(ctx context.Context, userID uuid.UUID) (*OnboardingProfile, error)
	// CompleteOnboarding atomically writes the user_onboarding_profiles
	// row and sets users.onboarding_status to 'completed'. When no
	// user_settings row exists, it creates one with daily_review_target
	// set to answers.DailyReviewTarget. When one already exists, it
	// applies the D04 rule (overwrite iff at the schema default;
	// preserve any other value). The function returns the persisted
	// profile and the resulting user_settings state.
	CompleteOnboarding(ctx context.Context, userID uuid.UUID, answers OnboardingAnswers, now time.Time) (*OnboardingProfile, StoredUserSettings, error)
	// SetOnboardingStatus updates users.onboarding_status. This is used
	// by tests and (in a future package) by the grandfather backfill
	// reconciliation. T01 does not use it on the request path.
	SetOnboardingStatus(ctx context.Context, userID uuid.UUID, status string, now time.Time) error
}

// Public service errors. The API layer maps these to stable 4xx
// responses (and 404 for the privacy-preserving "no such resource"
// case).
var (
	ErrInvalidOnboarding  = errors.New("invalid onboarding answers")
	ErrOnboardingNotFound = errors.New("onboarding profile not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrOnboardingConflict = errors.New("onboarding profile already exists")
)

// Service is the requester-scoped onboarding + settings read/write boundary.
type Service struct {
	repo         Repository
	settings     UserSettingsReader
	settingsRepo SettingsRepository
	clock        clock.Clock
}

// NewService creates a users.Service. settingsRepo may be nil for
// callers that only need onboarding; passing a non-nil value is
// required to exercise the GET/PATCH /api/v1/settings endpoints.
func NewService(repo Repository, settings UserSettingsReader, settingsRepo SettingsRepository, c clock.Clock) *Service {
	if c == nil {
		c = clock.Real{}
	}
	return &Service{repo: repo, settings: settings, settingsRepo: settingsRepo, clock: c}
}

// GetOnboarding returns the requester's onboarding profile and the
// current onboarding_status. Returns ErrOnboardingNotFound when the
// user has never submitted onboarding; the API layer maps that to a
// 200 with status='not_started' and a null profile body, so a
// frontend can branch on status alone without a separate "no profile"
// error.
func (s *Service) GetOnboarding(ctx context.Context, userID uuid.UUID) (*OnboardingProfile, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	return s.repo.GetOnboarding(ctx, userID)
}

// CompleteOnboarding persists the onboarding answers, sets
// users.onboarding_status='completed', and applies the D04 seed rule
// to user_settings. The whole flow is wrapped in a single repository
// transaction (see postgres.go).
//
// The caller is expected to have authenticated the requester; the
// userID here is the requester-scoped identifier. There is no
// "complete someone else's onboarding" path; any owner mismatch is
// caught at the API layer (the requester never supplies a userID).
func (s *Service) CompleteOnboarding(ctx context.Context, userID uuid.UUID, answers OnboardingAnswers) (*OnboardingProfile, StoredUserSettings, error) {
	if userID == uuid.Nil {
		return nil, StoredUserSettings{}, errors.New("user id required")
	}
	if err := answers.Validate(); err != nil {
		return nil, StoredUserSettings{}, err
	}
	now := s.clock.Now().UTC()
	profile, stored, err := s.repo.CompleteOnboarding(ctx, userID, answers, now)
	if err != nil {
		return nil, StoredUserSettings{}, err
	}
	// Re-read the stored user_settings through the UserSettingsReader so
	// the seed decision can be re-verified by callers (and future
	// tests) without coupling to the repository's return value.
	if s.settings != nil {
		stored, err = s.settings.GetStoredUserSettings(ctx, userID)
		if err != nil {
			return nil, StoredUserSettings{}, fmt.Errorf("read settings: %w", err)
		}
	}
	return profile, stored, nil
}

// GetSettings returns the requester's Settings projection. The
// requester is always taken from the session; the userID is never
// trusted from the request body. The call returns
// ErrSettingsNotFound when the user does not exist or has been
// soft-deleted (the API layer maps that to 404).
func (s *Service) GetSettings(ctx context.Context, userID uuid.UUID) (Settings, error) {
	if userID == uuid.Nil {
		return Settings{}, errors.New("user id required")
	}
	if s.settingsRepo == nil {
		return Settings{}, errors.New("settings repository not configured")
	}
	return s.settingsRepo.GetSettings(ctx, userID)
}

// UpdateSettings applies a partial update to the requester's
// Settings. Empty updates (no fields set) are accepted and return
// the current state unchanged, matching DOC-07 §3's "no-op PATCH
// is a well-formed read" rule. Validation runs before the
// repository call so the API layer can map errors to 400
// consistently.
func (s *Service) UpdateSettings(ctx context.Context, userID uuid.UUID, update SettingsUpdate) (Settings, error) {
	if userID == uuid.Nil {
		return Settings{}, errors.New("user id required")
	}
	if s.settingsRepo == nil {
		return Settings{}, errors.New("settings repository not configured")
	}
	if err := update.Validate(); err != nil {
		return Settings{}, err
	}
	now := s.clock.Now().UTC()
	return s.settingsRepo.UpdateSettings(ctx, userID, update, now)
}
