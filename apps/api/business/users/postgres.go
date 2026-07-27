package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgreSQLRepository implements Repository against the T00
// user_onboarding_profiles migration. It owns no state of its own;
// the user_settings read it needs for the D04 decision is delegated
// to the UserSettingsReader the Service is constructed with.
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a Repository backed by db.
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

// GetOnboarding returns the requester's profile and onboarding status.
// It uses a LEFT JOIN to keep the read in one round-trip and to
// preserve the always-defined onboarding_status field on the response.
func (r *PostgreSQLRepository) GetOnboarding(ctx context.Context, userID uuid.UUID) (*OnboardingProfile, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT u.onboarding_status,
		        COALESCE(p.english_level, ''),
		        COALESCE(p.native_language, ''),
		        COALESCE(p.learning_goal, ''),
		        COALESCE(p.main_use_case, ''),
		        COALESCE(p.daily_review_target, 0),
		        p.completed_at,
		        COALESCE(p.created_at, u.created_at),
		        COALESCE(p.updated_at, u.updated_at)
		 FROM users u
		 LEFT JOIN user_onboarding_profiles p ON p.user_id = u.id
		 WHERE u.id = $1 AND u.deleted_at IS NULL`,
		userID,
	)
	var (
		status            string
		englishLevel      string
		nativeLanguage    string
		learningGoal      string
		mainUseCase       string
		dailyReviewTarget int
		completedAt       sql.NullTime
		createdAt         time.Time
		updatedAt         time.Time
	)
	err := row.Scan(&status, &englishLevel, &nativeLanguage, &learningGoal, &mainUseCase,
		&dailyReviewTarget, &completedAt, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetch onboarding: %w", err)
	}
	if status == "" {
		status = OnboardingStatusNotStarted
	}
	p := &OnboardingProfile{
		UserID:    userID,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	if completedAt.Valid {
		t := completedAt.Time
		p.CompletedAt = &t
		p.EnglishLevel = englishLevel
		p.NativeLanguage = nativeLanguage
		p.LearningGoal = learningGoal
		p.MainUseCase = mainUseCase
		p.DailyReviewTarget = dailyReviewTarget
		return p, nil
	}
	return p, ErrOnboardingNotFound
}

// CompleteOnboarding performs, inside a single transaction:
//  1. UPDATE users SET onboarding_status='completed' WHERE id=$1 AND deleted_at IS NULL.
//  2. INSERT INTO user_onboarding_profiles ON CONFLICT (user_id) DO NOTHING.
//  3. Apply the D04 rule to user_settings via the same
//     on-conflict upsert the gamification module already uses for its
//     lazy-creation path, with a CASE that preserves any non-default
//     daily_review_target (the spec's "never overwrite a customized
//     row" rule).
//
// The function returns the resulting user_settings state in the shape
// Service.go's UserSettingsReader can re-verify.
func (r *PostgreSQLRepository) CompleteOnboarding(ctx context.Context, userID uuid.UUID, answers OnboardingAnswers, now time.Time) (*OnboardingProfile, StoredUserSettings, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, StoredUserSettings{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE users SET onboarding_status = 'completed', updated_at = $2
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID, now,
	)
	if err != nil {
		return nil, StoredUserSettings{}, fmt.Errorf("update onboarding status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, StoredUserSettings{}, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return nil, StoredUserSettings{}, ErrUserNotFound
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_onboarding_profiles (
			id, user_id, english_level, native_language, learning_goal,
			main_use_case, daily_review_target, completed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $8, $8
		)
		ON CONFLICT (user_id) DO NOTHING`,
		uuid.New(), userID,
		answers.EnglishLevel, answers.NativeLanguage, answers.LearningGoal,
		answers.MainUseCase, answers.DailyReviewTarget, now,
	)
	if err != nil {
		return nil, StoredUserSettings{}, fmt.Errorf("insert onboarding profile: %w", err)
	}

	// Apply the D04 seed rule. The user_settings row may or may not
	// exist yet; either way we use the same on-conflict upsert
	// gamification uses, with the daily_review_target overwrite
	// guarded by the CASE so any non-default existing value is
	// preserved (the spec's "never overwrite a customized row" rule).
	row := tx.QueryRowContext(ctx,
		`INSERT INTO user_settings (id, user_id, timezone, daily_review_target)
		 VALUES ($1, $2, 'UTC', $3)
		 ON CONFLICT (user_id) DO UPDATE
		   SET daily_review_target = CASE WHEN user_settings.daily_review_target <> $4
		                                  THEN user_settings.daily_review_target
		                                  ELSE EXCLUDED.daily_review_target
		                              END,
		       updated_at = NOW()
		 RETURNING user_id, daily_review_target`,
		uuid.New(), userID, answers.DailyReviewTarget, SchemaDailyReviewTargetDefault,
	)
	var (
		storedID  uuid.UUID
		storedTgt int
	)
	if err := row.Scan(&storedID, &storedTgt); err != nil {
		return nil, StoredUserSettings{}, fmt.Errorf("upsert user settings: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, StoredUserSettings{}, fmt.Errorf("commit: %w", err)
	}

	// Re-read the profile for the response. The transaction is
	// already committed; the LEFT JOIN keeps the response shape
	// consistent with GetOnboarding.
	profile, err := r.GetOnboarding(ctx, userID)
	if err != nil && !errors.Is(err, ErrOnboardingNotFound) {
		return nil, StoredUserSettings{}, err
	}
	if profile == nil {
		profile = &OnboardingProfile{UserID: userID, Status: OnboardingStatusCompleted}
	}
	// Determine whether a row existed before the upsert; the simpler
	// interpretation here is "the caller has the row now, so Stored=true
	// is correct", but the seed decision was based on the prior state.
	// Service.go re-reads through UserSettingsReader for the exact value
	// to honor D04, so we conservatively report Stored=true and
	// DailyReviewTarget from the upsert output here.
	stored := StoredUserSettings{Stored: true, DailyReviewTarget: storedTgt}
	return profile, stored, nil
}

// SetOnboardingStatus updates users.onboarding_status. T01's request
// path does not use it; the grandfather migration uses raw SQL,
// and T01's tests use it as a deterministic seam.
func (r *PostgreSQLRepository) SetOnboardingStatus(ctx context.Context, userID uuid.UUID, status string, now time.Time) error {
	if !validOnboardingStatus(status) {
		return fmt.Errorf("invalid onboarding status %q", status)
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET onboarding_status = $2, updated_at = $3
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID, status, now,
	)
	if err != nil {
		return fmt.Errorf("set onboarding status: %w", err)
	}
	return nil
}

func validOnboardingStatus(s string) bool {
	switch s {
	case OnboardingStatusNotStarted, OnboardingStatusInProgress, OnboardingStatusCompleted:
		return true
	}
	return false
}
