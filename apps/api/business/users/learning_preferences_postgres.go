package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const learningPreferencesSelect = `SELECT u.onboarding_status,
 COALESCE(p.learning_goal, o.learning_goal), COALESCE(p.main_use_case, o.main_use_case), COALESCE(p.revision, 0)
 FROM users u LEFT JOIN user_learning_preferences p ON p.user_id=u.id
 LEFT JOIN user_onboarding_profiles o ON o.user_id=u.id
 WHERE u.id=$1 AND u.status='active' AND u.deleted_at IS NULL`

func scanLearningPreferences(row interface{ Scan(...any) error }) (LearningPreferences, error) {
	var p LearningPreferences
	var status string
	err := row.Scan(&status, &p.LearningGoal, &p.MainUseCase, &p.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrUserNotFound
	}
	if err != nil {
		return p, fmt.Errorf("read learning preferences: %w", err)
	}
	if status != OnboardingStatusCompleted {
		return LearningPreferences{}, ErrLearningPreferencesOnboarding
	}
	return p, nil
}

func (r *PostgreSQLRepository) GetLearningPreferences(ctx context.Context, userID uuid.UUID) (LearningPreferences, error) {
	return scanLearningPreferences(r.db.QueryRowContext(ctx, learningPreferencesSelect, userID))
}

func (r *PostgreSQLRepository) UpdateLearningPreferences(ctx context.Context, userID uuid.UUID, update LearningPreferencesUpdate) (LearningPreferences, error) {
	if err := update.Validate(); err != nil {
		return LearningPreferences{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return LearningPreferences{}, err
	}
	defer tx.Rollback()
	// The owner row serializes absence/creation, competing changes and account
	// deletion. Never resurrect preferences after a soft deletion or purge.
	var status string
	err = tx.QueryRowContext(ctx, `SELECT onboarding_status FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE`, userID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return LearningPreferences{}, ErrUserNotFound
	}
	if err != nil {
		return LearningPreferences{}, err
	}
	if status != OnboardingStatusCompleted {
		return LearningPreferences{}, ErrLearningPreferencesOnboarding
	}
	current, err := scanLearningPreferences(tx.QueryRowContext(ctx, learningPreferencesSelect, userID))
	if err != nil {
		return LearningPreferences{}, err
	}
	// A retry of already-applied intent is a read-only success, even with the
	// original revision. A different stale intent can never overwrite it.
	if current.matches(update) {
		return current, tx.Commit()
	}
	if current.Revision != update.ExpectedRevision {
		return LearningPreferences{}, ErrLearningPreferencesConflict
	}
	var result LearningPreferences
	err = tx.QueryRowContext(ctx, `INSERT INTO user_learning_preferences(id,user_id,learning_goal,main_use_case,revision,created_at,updated_at)
 VALUES($1,$2,$3,$4,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
 ON CONFLICT(user_id) DO UPDATE SET learning_goal=EXCLUDED.learning_goal, main_use_case=EXCLUDED.main_use_case,
 revision=user_learning_preferences.revision+1, updated_at=CURRENT_TIMESTAMP
 WHERE user_learning_preferences.revision=$5
 RETURNING learning_goal,main_use_case,revision`, uuid.New(), userID, update.LearningGoal, update.MainUseCase, update.ExpectedRevision).Scan(&result.LearningGoal, &result.MainUseCase, &result.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return LearningPreferences{}, ErrLearningPreferencesConflict
	}
	if err != nil {
		return LearningPreferences{}, err
	}
	if err := tx.Commit(); err != nil {
		return LearningPreferences{}, err
	}
	return result, nil
}
