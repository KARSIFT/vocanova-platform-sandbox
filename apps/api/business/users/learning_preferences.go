package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// LearningPreferences is current learning direction, independent of the original
// onboarding submission. Revision zero means there is no explicit override.
type LearningPreferences struct {
	LearningGoal *string `json:"learningGoal"`
	MainUseCase  *string `json:"mainUseCase"`
	Revision     int64   `json:"revision"`
}

type LearningPreferencesUpdate struct {
	LearningGoal     string
	MainUseCase      string
	ExpectedRevision int64
}

var (
	ErrInvalidLearningPreferences    = errors.New("invalid learning preferences")
	ErrLearningPreferencesConflict   = errors.New("learning preferences changed")
	ErrLearningPreferencesOnboarding = errors.New("complete onboarding first")
)

func (p LearningPreferencesUpdate) Validate() error {
	if p.ExpectedRevision < 0 {
		return ErrInvalidLearningPreferences
	}
	switch p.LearningGoal {
	case LearningGoalGeneral, LearningGoalWork, LearningGoalTravel, LearningGoalStudy, LearningGoalConversation, LearningGoalExam:
	default:
		return ErrInvalidLearningPreferences
	}
	switch p.MainUseCase {
	case MainUseCaseDailyLife, MainUseCaseWork, MainUseCaseTravel, MainUseCaseStudy, MainUseCaseSocial:
	default:
		return ErrInvalidLearningPreferences
	}
	return nil
}

func (p LearningPreferences) matches(update LearningPreferencesUpdate) bool {
	return p.LearningGoal != nil && p.MainUseCase != nil && *p.LearningGoal == update.LearningGoal && *p.MainUseCase == update.MainUseCase
}

type LearningPreferencesRepository interface {
	GetLearningPreferences(context.Context, uuid.UUID) (LearningPreferences, error)
	UpdateLearningPreferences(context.Context, uuid.UUID, LearningPreferencesUpdate) (LearningPreferences, error)
}

type LearningPreferencesService struct{ repo LearningPreferencesRepository }

func NewLearningPreferencesService(repo LearningPreferencesRepository) *LearningPreferencesService {
	return &LearningPreferencesService{repo: repo}
}

func (s *LearningPreferencesService) GetLearningPreferences(ctx context.Context, userID uuid.UUID) (LearningPreferences, error) {
	if userID == uuid.Nil {
		return LearningPreferences{}, ErrUserNotFound
	}
	return s.repo.GetLearningPreferences(ctx, userID)
}

func (s *LearningPreferencesService) UpdateLearningPreferences(ctx context.Context, userID uuid.UUID, update LearningPreferencesUpdate) (LearningPreferences, error) {
	if userID == uuid.Nil {
		return LearningPreferences{}, ErrUserNotFound
	}
	if err := update.Validate(); err != nil {
		return LearningPreferences{}, err
	}
	return s.repo.UpdateLearningPreferences(ctx, userID, update)
}
