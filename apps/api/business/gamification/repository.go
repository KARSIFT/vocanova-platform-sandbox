package gamification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PointLedgerRow is one confidence_point_ledger row read back from the DB.
type PointLedgerRow struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Amount         int
	BalanceAfter   int
	Reason         string
	SourceType     string
	SourceID       *uuid.UUID
	IdempotencyKey *string
	Metadata       json.RawMessage
	OccurredAt     time.Time
	CreatedAt      time.Time
}

// StreakStateRow is one streak_states row read back from the DB.
type StreakStateRow struct {
	UserID                 uuid.UUID
	CurrentStreakCount     int
	LongestStreakCount     int
	LastCompletedLocalDate *time.Time
	LastActivityLocalDate  *time.Time
	Timezone               string
	Status                 string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// GraceDayRow is one grace_day_ledger row read back from the DB.
type GraceDayRow struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Amount             int
	BalanceAfter       int
	Reason             string
	SourceType         string
	SourceID           *uuid.UUID
	AppliedToLocalDate time.Time
	Timezone           string
	IdempotencyKey     *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// UserSettingsRow is one user_settings row read back from the DB.
type UserSettingsRow struct {
	UserID                 uuid.UUID
	Timezone               string
	DailyReviewTarget      int
	ReviewIntervalPreset   string
	NotificationsEnabled   bool
	MarketingEmailsEnabled bool
	AppLanguage            string
}

// Repository persists gamification state. The transaction-scoped helpers
// accept an existing *sql.Tx from the caller's existing P1/P2/P3 transaction
// (DOC-06 §3 — missions/gamification never open their own transaction).
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by db.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetStreakState reads the current streak_states row for userID. Returns
// (nil, nil) if no row exists yet.
func (r *Repository) GetStreakState(ctx context.Context, userID uuid.UUID) (*StreakStateRow, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT user_id, current_streak_count, longest_streak_count,
		        last_completed_local_date, last_activity_local_date,
		        timezone, status, created_at, updated_at
		 FROM streak_states
		 WHERE user_id = $1`,
		userID,
	)
	var s StreakStateRow
	var lastCompleted, lastActivity sql.NullTime
	if err := row.Scan(&s.UserID, &s.CurrentStreakCount, &s.LongestStreakCount,
		&lastCompleted, &lastActivity, &s.Timezone, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("fetch streak state: %w", err)
	}
	if lastCompleted.Valid {
		t := lastCompleted.Time
		s.LastCompletedLocalDate = &t
	}
	if lastActivity.Valid {
		t := lastActivity.Time
		s.LastActivityLocalDate = &t
	}
	return &s, nil
}

// GetLatestGraceBalance returns the user's current available grace-day
// balance from the latest grace_day_ledger row. 0 if no rows exist.
func (r *Repository) GetLatestGraceBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT balance_after FROM grace_day_ledger
		 WHERE user_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT 1`,
		userID,
	)
	var balance int
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("fetch latest grace balance: %w", err)
	}
	return balance, nil
}

// GetUserSettings reads the user_settings row for userID. Returns
// (nil, nil) if no row exists yet.
func (r *Repository) GetUserSettings(ctx context.Context, userID uuid.UUID) (*UserSettingsRow, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT user_id, timezone, daily_review_target, review_interval_preset,
		        notifications_enabled, marketing_emails_enabled, app_language
		 FROM user_settings
		 WHERE user_id = $1`,
		userID,
	)
	var s UserSettingsRow
	if err := row.Scan(&s.UserID, &s.Timezone, &s.DailyReviewTarget, &s.ReviewIntervalPreset,
		&s.NotificationsEnabled, &s.MarketingEmailsEnabled, &s.AppLanguage); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("fetch user settings: %w", err)
	}
	return &s, nil
}

// UpsertUserSettings lazily creates the user_settings row on first use
// (DOC-06 §10-style lazy creation), then updates the timezone / target if
// either differs from the supplied values. Returns the effective stored row.
func (r *Repository) UpsertUserSettings(ctx context.Context, tx *sql.Tx, userID uuid.UUID, timezone string, dailyReviewTarget int) (*UserSettingsRow, error) {
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	if timezone == "" {
		timezone = DefaultTimezone
	}
	if dailyReviewTarget < MinDailyReviewTarget || dailyReviewTarget > MaxDailyReviewTarget {
		return nil, fmt.Errorf("daily review target %d out of range [%d,%d]", dailyReviewTarget, MinDailyReviewTarget, MaxDailyReviewTarget)
	}
	row := tx.QueryRowContext(ctx,
		`INSERT INTO user_settings (
			id, user_id, timezone, daily_review_target, created_at, updated_at
		)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (user_id) DO UPDATE
		   SET timezone = COALESCE(NULLIF(user_settings.timezone, 'UTC'), EXCLUDED.timezone),
		       daily_review_target = CASE WHEN user_settings.daily_review_target <> 20
		                                  THEN user_settings.daily_review_target
		                                  ELSE EXCLUDED.daily_review_target
		                              END,
		       updated_at = NOW()
		 RETURNING user_id, timezone, daily_review_target, review_interval_preset,
		           notifications_enabled, marketing_emails_enabled, app_language`,
		uuid.New(), userID, timezone, dailyReviewTarget,
	)
	var s UserSettingsRow
	if err := row.Scan(&s.UserID, &s.Timezone, &s.DailyReviewTarget, &s.ReviewIntervalPreset,
		&s.NotificationsEnabled, &s.MarketingEmailsEnabled, &s.AppLanguage); err != nil {
		return nil, fmt.Errorf("upsert user settings: %w", err)
	}
	return &s, nil
}

// InsertPointLedger writes one confidence_point_ledger row inside tx.
// idempotencyKey may be empty. Returns the inserted row id. If a row with the
// same (user_id, idempotency_key) already exists, the insert is a no-op
// (returning the existing id) — the partial unique index turns a duplicate
// into a defensive guard against the primary idempotency mechanism being
// bypassed.
func (r *Repository) InsertPointLedger(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	amount int,
	balanceAfter int,
	reason, sourceType string,
	sourceID *uuid.UUID,
	idempotencyKey PointIdempotencyKey,
	metadata json.RawMessage,
	occurredAt time.Time,
) (uuid.UUID, error) {
	if tx == nil {
		return uuid.Nil, errors.New("transaction required")
	}
	var key sql.NullString
	if idempotencyKey != "" {
		key = sql.NullString{String: idempotencyKey.String(), Valid: true}
	}
	var meta any
	if len(metadata) > 0 {
		meta = []byte(metadata)
	}
	var id uuid.UUID
	err := tx.QueryRowContext(ctx,
		`INSERT INTO confidence_point_ledger (
			id, user_id, amount, balance_after, reason, source_type,
			source_id, idempotency_key, metadata, occurred_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, NOW(), NOW()
		)
		ON CONFLICT (user_id, idempotency_key) WHERE idempotency_key IS NOT NULL
		DO UPDATE SET amount = confidence_point_ledger.amount
		RETURNING id`,
		uuid.New(), userID, amount, balanceAfter, reason, sourceType,
		sourceID, key, meta, occurredAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert point ledger: %w", err)
	}
	return id, nil
}

// InsertGraceLedger writes one grace_day_ledger row inside tx. idempotencyKey
// may be empty. The same ON CONFLICT rule as confidence_point_ledger applies.
func (r *Repository) InsertGraceLedger(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	amount int,
	balanceAfter int,
	reason, sourceType string,
	sourceID *uuid.UUID,
	appliedToLocalDate time.Time,
	timezone string,
	idempotencyKey GraceIdempotencyKey,
) (uuid.UUID, error) {
	if tx == nil {
		return uuid.Nil, errors.New("transaction required")
	}
	var key sql.NullString
	if idempotencyKey != "" {
		key = sql.NullString{String: idempotencyKey.String(), Valid: true}
	}
	var src any
	if sourceID != nil {
		src = *sourceID
	}
	var id uuid.UUID
	err := tx.QueryRowContext(ctx,
		`INSERT INTO grace_day_ledger (
			id, user_id, amount, balance_after, reason, source_type,
			source_id, applied_to_local_date, timezone, idempotency_key, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, NOW(), NOW()
		)
		ON CONFLICT (user_id, idempotency_key) WHERE idempotency_key IS NOT NULL
		DO UPDATE SET amount = grace_day_ledger.amount
		RETURNING id`,
		uuid.New(), userID, amount, balanceAfter, reason, sourceType,
		src, appliedToLocalDate, timezone, key,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert grace ledger: %w", err)
	}
	return id, nil
}

// UpsertStreakState creates or updates the streak_states row for userID
// inside tx.
func (r *Repository) UpsertStreakState(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	currentStreakCount, longestStreakCount int,
	lastCompletedLocalDate, lastActivityLocalDate *time.Time,
	timezone, status string,
) error {
	if tx == nil {
		return errors.New("transaction required")
	}
	var lastCompleted, lastActivity sql.NullTime
	if lastCompletedLocalDate != nil {
		lastCompleted = sql.NullTime{Time: *lastCompletedLocalDate, Valid: true}
	}
	if lastActivityLocalDate != nil {
		lastActivity = sql.NullTime{Time: *lastActivityLocalDate, Valid: true}
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO streak_states (
			id, user_id, current_streak_count, longest_streak_count,
			last_completed_local_date, last_activity_local_date,
			timezone, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8, NOW(), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE
		  SET current_streak_count = EXCLUDED.current_streak_count,
		      longest_streak_count = EXCLUDED.longest_streak_count,
		      last_completed_local_date = EXCLUDED.last_completed_local_date,
		      last_activity_local_date = EXCLUDED.last_activity_local_date,
		      timezone = EXCLUDED.timezone,
		      status = EXCLUDED.status,
		      updated_at = NOW()`,
		uuid.New(), userID, currentStreakCount, longestStreakCount,
		lastCompleted, lastActivity, timezone, status,
	)
	if err != nil {
		return fmt.Errorf("upsert streak state: %w", err)
	}
	return nil
}

// GetLatestPointBalance returns the user's current Confidence Points balance
// from the latest confidence_point_ledger row. 0 if no rows exist.
func (r *Repository) GetLatestPointBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT balance_after FROM confidence_point_ledger
		 WHERE user_id = $1
		 ORDER BY occurred_at DESC, id DESC
		 LIMIT 1`,
		userID,
	)
	var balance int
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("fetch latest point balance: %w", err)
	}
	return balance, nil
}
