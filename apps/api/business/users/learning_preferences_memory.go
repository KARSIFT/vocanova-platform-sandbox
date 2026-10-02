package users

import (
	"context"
	"github.com/google/uuid"
)

func copyLearningPreferences(p LearningPreferences) LearningPreferences {
	if p.LearningGoal != nil {
		v := *p.LearningGoal
		p.LearningGoal = &v
	}
	if p.MainUseCase != nil {
		v := *p.MainUseCase
		p.MainUseCase = &v
	}
	return p
}

func (r *MemoryRepository) learningPreferencesLocked(userID uuid.UUID) (LearningPreferences, error) {
	status, exists := r.onboarding[userID]
	if !exists {
		return LearningPreferences{}, ErrUserNotFound
	}
	if status != OnboardingStatusCompleted {
		return LearningPreferences{}, ErrLearningPreferencesOnboarding
	}
	if p, ok := r.learningPreferences[userID]; ok {
		return copyLearningPreferences(p), nil
	}
	if p := r.profiles[userID]; p != nil {
		goal, focus := p.LearningGoal, p.MainUseCase
		return LearningPreferences{LearningGoal: &goal, MainUseCase: &focus}, nil
	}
	return LearningPreferences{}, nil
}

func (r *MemoryRepository) GetLearningPreferences(_ context.Context, userID uuid.UUID) (LearningPreferences, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.learningPreferencesLocked(userID)
}

func (r *MemoryRepository) UpdateLearningPreferences(_ context.Context, userID uuid.UUID, update LearningPreferencesUpdate) (LearningPreferences, error) {
	if err := update.Validate(); err != nil {
		return LearningPreferences{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	p, err := r.learningPreferencesLocked(userID)
	if err != nil {
		return p, err
	}
	if p.matches(update) {
		return p, nil
	}
	if p.Revision != update.ExpectedRevision {
		return LearningPreferences{}, ErrLearningPreferencesConflict
	}
	p = LearningPreferences{LearningGoal: &update.LearningGoal, MainUseCase: &update.MainUseCase, Revision: p.Revision + 1}
	if r.learningPreferences == nil {
		r.learningPreferences = make(map[uuid.UUID]LearningPreferences)
	}
	r.learningPreferences[userID] = copyLearningPreferences(p)
	return p, nil
}
