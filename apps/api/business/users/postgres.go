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
		`INSERT INTO user_settings (id, user_id, timezone, daily_review_target, created_at, updated_at)
		 VALUES ($1, $2, 'UTC', $3, $5, $5)
		 ON CONFLICT (user_id) DO UPDATE
		   SET daily_review_target = CASE WHEN user_settings.daily_review_target <> $4
		                                  THEN user_settings.daily_review_target
		                                  ELSE EXCLUDED.daily_review_target
		                              END,
		       updated_at = NOW()
		 RETURNING user_id, daily_review_target`,
		uuid.New(), userID, answers.DailyReviewTarget, SchemaDailyReviewTargetDefault, now,
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

// GetStoredUserSettings implements UserSettingsReader against the
// user_settings schema. The D04 seed-eligibility decision the
// users.Service makes in CompleteOnboarding re-reads the stored
// daily_review_target to decide between SeedCreateRow,
// SeedOverwriteDefault, and SeedPreserveExisting (VOC-031-D04).
// Mirrors the MemoryRepository implementation: when no row exists
// yet, returns Stored=false so the service picks SeedCreateRow.
func (r *PostgreSQLRepository) GetStoredUserSettings(ctx context.Context, userID uuid.UUID) (StoredUserSettings, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT daily_review_target FROM user_settings WHERE user_id = $1`,
		userID,
	)
	var target int
	err := row.Scan(&target)
	if errors.Is(err, sql.ErrNoRows) {
		return StoredUserSettings{Stored: false}, nil
	}
	if err != nil {
		return StoredUserSettings{}, fmt.Errorf("fetch stored user settings: %w", err)
	}
	return StoredUserSettings{Stored: true, DailyReviewTarget: target}, nil
}

// GetSettings returns the requester's Settings projection. The
// user_settings row may not exist yet; in that case every field
// is filled from the user_settings schema defaults (the values
// the schema would have produced for a brand-new row). The
// users.display_name read always comes from the users table, even
// when no user_settings row exists, so a learner with no row yet
// still sees a stable response.
func (r *PostgreSQLRepository) GetSettings(ctx context.Context, userID uuid.UUID) (Settings, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT u.display_name,
		        s.daily_review_target,
		        s.review_interval_preset,
		        s.notifications_enabled,
		        s.marketing_emails_enabled,
		        s.app_language
		 FROM users u
		 LEFT JOIN user_settings s ON s.user_id = u.id
		 WHERE u.id = $1 AND u.deleted_at IS NULL`,
		userID,
	)
	var (
		displayName      sql.NullString
		dailyReview      sql.NullInt32
		intervalPreset   sql.NullString
		notifsEnabled    sql.NullBool
		marketingEnabled sql.NullBool
		appLanguage      sql.NullString
	)
	err := row.Scan(&displayName, &dailyReview, &intervalPreset,
		&notifsEnabled, &marketingEnabled, &appLanguage)
	if errors.Is(err, sql.ErrNoRows) {
		return Settings{}, ErrSettingsNotFound
	}
	if err != nil {
		return Settings{}, fmt.Errorf("fetch settings: %w", err)
	}
	settings := Settings{
		DailyReviewTarget:      schemaDailyReviewTargetDefaultInt,
		ReviewIntervalPreset:   schemaReviewIntervalPresetDefault,
		AppLanguage:            schemaAppLanguageDefault,
		NotificationsEnabled:   schemaNotificationsEnabledDefault,
		MarketingEmailsEnabled: schemaMarketingEmailsEnabledDefault,
		DisplayName:            displayName.String,
	}
	if dailyReview.Valid {
		settings.DailyReviewTarget = int(dailyReview.Int32)
	}
	if intervalPreset.Valid {
		settings.ReviewIntervalPreset = intervalPreset.String
	}
	if notifsEnabled.Valid {
		settings.NotificationsEnabled = notifsEnabled.Bool
	}
	if marketingEnabled.Valid {
		settings.MarketingEmailsEnabled = marketingEnabled.Bool
	}
	if appLanguage.Valid {
		settings.AppLanguage = appLanguage.String
	}
	return settings, nil
}

// UpdateSettings atomically applies a partial Settings update to
// the user_settings row and (when the caller supplies a new
// display_name) users.display_name. The implementation reads the
// existing user_settings row inside the transaction, merges the
// update in Go, then writes the merged result with an ON CONFLICT
// upsert. This handles the first-ever-write case (no existing
// row) without a unique-constraint race against the gamification
// module's lazy-create path (VOC-031-R05): the ON CONFLICT
// (user_id) DO UPDATE is the same pattern gamification uses, and
// the transactional read-modify-write is atomic.
func (r *PostgreSQLRepository) UpdateSettings(ctx context.Context, userID uuid.UUID, update SettingsUpdate, now time.Time) (Settings, error) {
	if err := update.Validate(); err != nil {
		return Settings{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Settings{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	merged, err := readMergedSettingsForUpdate(ctx, tx, userID, update)
	if err != nil {
		return Settings{}, err
	}

	row := tx.QueryRowContext(ctx,
		`INSERT INTO user_settings (id, user_id, timezone, daily_review_target,
		                           review_interval_preset, notifications_enabled,
		                           marketing_emails_enabled, app_language)
		 VALUES ($1, $2, 'UTC', $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id) DO UPDATE
		   SET daily_review_target = EXCLUDED.daily_review_target,
		       review_interval_preset = EXCLUDED.review_interval_preset,
		       notifications_enabled = EXCLUDED.notifications_enabled,
		       marketing_emails_enabled = EXCLUDED.marketing_emails_enabled,
		       app_language = EXCLUDED.app_language,
		       updated_at = NOW()
		 RETURNING daily_review_target, review_interval_preset,
		           notifications_enabled, marketing_emails_enabled, app_language`,
		uuid.New(), userID,
		merged.DailyReviewTarget, merged.ReviewIntervalPreset,
		merged.NotificationsEnabled, merged.MarketingEmailsEnabled,
		merged.AppLanguage,
	)
	if err := row.Scan(
		&merged.DailyReviewTarget, &merged.ReviewIntervalPreset,
		&merged.NotificationsEnabled, &merged.MarketingEmailsEnabled,
		&merged.AppLanguage,
	); err != nil {
		return Settings{}, fmt.Errorf("upsert user settings: %w", err)
	}

	if update.DisplayName != nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET display_name = $2, updated_at = $3
			 WHERE id = $1 AND deleted_at IS NULL`,
			userID, *update.DisplayName, now,
		); err != nil {
			return Settings{}, fmt.Errorf("update display name: %w", err)
		}
		merged.DisplayName = *update.DisplayName
	} else {
		// Re-read the display_name so the response always
		// reflects the persisted value, even when the
		// caller did not include it in the PATCH. This is
		// the symmetric counterpart of GetSettings' LEFT
		// JOIN, kept in a single round-trip via a SELECT
		// on the same tx.
		var name sql.NullString
		if err := tx.QueryRowContext(ctx,
			`SELECT display_name FROM users WHERE id = $1`, userID,
		).Scan(&name); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return Settings{}, fmt.Errorf("re-read display name: %w", err)
		}
		merged.DisplayName = name.String
	}

	if err := tx.Commit(); err != nil {
		return Settings{}, fmt.Errorf("commit: %w", err)
	}
	return merged, nil
}

// readMergedSettingsForUpdate reads the existing user_settings
// row inside tx (returning schema defaults when no row exists) and
// applies the partial update in Go. This read-modify-write is
// what lets the conflict-updating upsert always pass a complete,
// correct row to EXCLUDED: the conflict path replaces every
// field with EXCLUDED.*, and only the EXCLUDED fields we just
// computed can win.
func readMergedSettingsForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, update SettingsUpdate) (Settings, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT daily_review_target, review_interval_preset,
		        notifications_enabled, marketing_emails_enabled, app_language
		 FROM user_settings WHERE user_id = $1`,
		userID,
	)
	merged := Settings{
		DailyReviewTarget:      schemaDailyReviewTargetDefaultInt,
		ReviewIntervalPreset:   schemaReviewIntervalPresetDefault,
		AppLanguage:            schemaAppLanguageDefault,
		NotificationsEnabled:   schemaNotificationsEnabledDefault,
		MarketingEmailsEnabled: schemaMarketingEmailsEnabledDefault,
	}
	var (
		dailyReview    sql.NullInt32
		intervalPreset sql.NullString
		notifsEnabled  sql.NullBool
		marketingOn    sql.NullBool
		appLanguage    sql.NullString
	)
	err := row.Scan(&dailyReview, &intervalPreset, &notifsEnabled, &marketingOn, &appLanguage)
	switch {
	case err == nil:
		if dailyReview.Valid {
			merged.DailyReviewTarget = int(dailyReview.Int32)
		}
		if intervalPreset.Valid {
			merged.ReviewIntervalPreset = intervalPreset.String
		}
		if notifsEnabled.Valid {
			merged.NotificationsEnabled = notifsEnabled.Bool
		}
		if marketingOn.Valid {
			merged.MarketingEmailsEnabled = marketingOn.Bool
		}
		if appLanguage.Valid {
			merged.AppLanguage = appLanguage.String
		}
	case errors.Is(err, sql.ErrNoRows):
		// No existing row — schema defaults are already in
		// merged. Caller is performing a first-ever write;
		// the upsert will create the row.
	default:
		return Settings{}, fmt.Errorf("read existing settings: %w", err)
	}

	if update.DailyReviewTarget != nil {
		merged.DailyReviewTarget = *update.DailyReviewTarget
	}
	if update.ReviewIntervalPreset != nil {
		merged.ReviewIntervalPreset = *update.ReviewIntervalPreset
	}
	if update.AppLanguage != nil {
		merged.AppLanguage = *update.AppLanguage
	}
	if update.NotificationsEnabled != nil {
		merged.NotificationsEnabled = *update.NotificationsEnabled
	}
	if update.MarketingEmailsEnabled != nil {
		merged.MarketingEmailsEnabled = *update.MarketingEmailsEnabled
	}
	return merged, nil
}

// Schema defaults for user_settings. These match the NOT NULL
// DEFAULT clauses in apps/api/migrations/20260725130000_voc030_p4_user_settings.sql
// and the user_settings ent schema's Default(...) calls. They are
// duplicated here (rather than imported from gamification) so
// users.Settings stays a self-contained module: importing
// gamification just for the constants would create a cycle the
// current package layout cannot resolve.
const (
	schemaDailyReviewTargetDefaultInt   = 20
	schemaReviewIntervalPresetDefault   = "vocanova_default"
	schemaAppLanguageDefault            = "en"
	schemaNotificationsEnabledDefault   = true
	schemaMarketingEmailsEnabledDefault = false
)
