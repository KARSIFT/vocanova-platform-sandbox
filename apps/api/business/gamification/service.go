package gamification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Service wires the pure gamification domain to the repository. It is
// transaction-scoped: every method that writes takes the caller's *sql.Tx
// from the existing P1/P2/P3 transaction (DOC-06 §3).
type Service struct {
	repo *Repository
}

// NewService creates a gamification service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetSettings returns the per-user resolved settings (timezone + target) by
// reading user_settings or falling back to schema defaults. The optional
// clientTimezone is the request-time IANA timezone; if non-empty it is
// validated and used as a fallback. Per D01, an unrecognized value is
// rejected via ErrInvalidTimezone.
func (s *Service) GetSettings(ctx context.Context, userID uuid.UUID, clientTimezone string) (ResolvedSettings, error) {
	row, err := s.repo.GetUserSettings(ctx, userID)
	if err != nil {
		return ResolvedSettings{}, err
	}
	source := UserSettingsSource{Stored: false}
	if row != nil {
		source = UserSettingsSource{
			Stored:            true,
			Timezone:          row.Timezone,
			DailyReviewTarget: row.DailyReviewTarget,
		}
	}
	return ResolveSettings(source, clientTimezone)
}

// EnsureUserSettings lazily creates the user_settings row inside tx with
// schema defaults if one does not exist; otherwise returns the existing row.
func (s *Service) EnsureUserSettings(ctx context.Context, tx *sql.Tx, userID uuid.UUID, timezone string, dailyReviewTarget int) (*UserSettingsRow, error) {
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	if timezone == "" {
		timezone = DefaultTimezone
	}
	if dailyReviewTarget <= 0 {
		dailyReviewTarget = DefaultDailyReviewTarget
	}
	return s.repo.UpsertUserSettings(ctx, tx, userID, timezone, dailyReviewTarget)
}

// GrantPoint writes one confidence_point_ledger row inside tx and returns the
// new running balance. If the idempotency key already exists for this user,
// the insert is a no-op (the existing row's id and amount are returned). The
// caller must have already computed the new balance (current + amount).
func (s *Service) GrantPoint(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	kind RewardKind,
	sourceID *uuid.UUID,
	idempotencyKey PointIdempotencyKey,
	currentBalance int,
	now time.Time,
	metadata json.RawMessage,
) (int, uuid.UUID, error) {
	if tx == nil {
		return 0, uuid.Nil, errors.New("transaction required")
	}
	outcome, err := RewardFor(kind)
	if err != nil {
		return 0, uuid.Nil, err
	}
	newBalance := currentBalance + outcome.Amount
	rowID, err := s.repo.InsertPointLedger(
		ctx, tx, userID, outcome.Amount, newBalance,
		outcome.Reason, outcome.SourceType, sourceID,
		idempotencyKey, metadata, now,
	)
	if err != nil {
		return 0, uuid.Nil, err
	}
	return newBalance, rowID, nil
}

// GrantGraceDay writes one grace_day_ledger row inside tx and returns the
// new running balance. The same idempotency rule as confidence_point_ledger
// applies. amount is the signed change (+1 for earn, -1 for use). The caller
// supplies the pre-computed balanceAfter (so chained calls within a single
// streak reconciliation see the right running balance).
func (s *Service) GrantGraceDay(
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
	rowID, err := s.repo.InsertGraceLedger(
		ctx, tx, userID, amount, balanceAfter,
		reason, sourceType, sourceID,
		appliedToLocalDate, timezone, idempotencyKey,
	)
	if err != nil {
		return uuid.Nil, err
	}
	return rowID, nil
}

// ReconcileAndAdvance runs the pure streak-reconciliation algorithm against
// the caller's tx, then writes back the new streak_states row, any earned or
// used grace-day ledger entries, and the yesterday-protected snapshot status
// transition. The caller supplies:
//
//   - now: the current time (UTC) at which the reconciliation is being done
//   - userID: the requester
//   - timezone: the resolved effective timezone (from the missions module)
//   - snapshots: the recent (last 14 local days) daily_mission_snapshots for
//     the user, in any order; reconciliation sorts them internally
//   - graceBalance: the user's current grace-day balance
//   - currentCompletion: true if the caller just completed today's mission
//     in the same call
//
// The returned *StreakReconciliation describes what was written so the caller
// can update its own UI-facing state. The state write happens inside the
// caller's tx; the helper never opens its own.
func (s *Service) ReconcileAndAdvance(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	now time.Time,
	timezone string,
	snapshots []StreakSnapshot,
	graceBalance int,
	currentCompletion bool,
) (*StreakReconciliation, error) {
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	state, err := s.repo.GetStreakState(ctx, userID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		state = &StreakStateRow{
			UserID:             userID,
			CurrentStreakCount: 0,
			LongestStreakCount: 0,
			Timezone:           timezone,
			Status:             StreakStatusActive,
		}
	}
	domainState := StreakState{
		CurrentStreakCount:     state.CurrentStreakCount,
		LongestStreakCount:     state.LongestStreakCount,
		LastCompletedLocalDate: state.LastCompletedLocalDate,
		LastActivityLocalDate:  state.LastActivityLocalDate,
		// The resolved timezone is the learner's current timezone. Historical
		// mission snapshots retain their own timezone, but streak reconciliation
		// must calculate today and yesterday in the current one.
		Timezone: timezone,
		Status:   state.Status,
	}
	rec, err := ReconcileStreak(userID.String(), now, domainState, GraceBalance{Balance: graceBalance}, snapshots, currentCompletion)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpsertStreakState(
		ctx, tx, userID,
		rec.NewState.CurrentStreakCount,
		rec.NewState.LongestStreakCount,
		rec.NewState.LastCompletedLocalDate,
		rec.NewState.LastActivityLocalDate,
		rec.NewState.Timezone,
		rec.NewState.Status,
	); err != nil {
		return nil, err
	}
	if rec.GraceDayUsed != nil {
		entry := rec.GraceDayUsed
		if _, err := s.GrantGraceDay(
			ctx, tx, userID,
			entry.Amount, entry.BalanceAfter,
			entry.Reason, entry.SourceType,
			entry.SourceID, entry.AppliedToLocalDate, entry.Timezone,
			entry.IdempotencyKey,
		); err != nil {
			return nil, err
		}
	}
	if rec.GraceDayEarned != nil {
		entry := rec.GraceDayEarned
		if _, err := s.GrantGraceDay(
			ctx, tx, userID,
			entry.Amount, entry.BalanceAfter,
			entry.Reason, entry.SourceType,
			entry.SourceID, entry.AppliedToLocalDate, entry.Timezone,
			entry.IdempotencyKey,
		); err != nil {
			return nil, err
		}
	}
	return &rec, nil
}

// CurrentBalance returns the user's current Confidence Points balance. The
// daily_activity_summaries.earned/spent columns stay in sync via the
// mission/activity writes; the ledger is the record of truth.
func (s *Service) CurrentBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetLatestPointBalance(ctx, userID)
}

// CurrentGraceBalance returns the user's current grace-day balance.
func (s *Service) CurrentGraceBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetLatestGraceBalance(ctx, userID)
}

// GetStreakStateForRead returns the current streak_states row for the user.
// It is read-only (no tx required) and is used by the missions read APIs to
// build the shared StreakView.
func (s *Service) GetStreakStateForRead(ctx context.Context, userID uuid.UUID) (*StreakStateRow, error) {
	return s.repo.GetStreakState(ctx, userID)
}
