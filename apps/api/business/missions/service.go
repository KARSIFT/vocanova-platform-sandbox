package missions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
)

// Service ties the missions module's repository to the gamification
// transaction-scoped helpers. It exposes the two read DTOs
// (DailyMissionView, ProgressView) that the API layer maps to Huma
// responses, and a MissionUpdater implementation that the aifeedback
// module wires in.
type Service struct {
	missions     *Repository
	gamification *gamification.Service
}

// NewService creates a missions service backed by the given repository pair.
func NewService(missions *Repository, gam *gamification.Service) *Service {
	return &Service{missions: missions, gamification: gam}
}

// EnsureTodaySnapshot lazily creates today's daily_mission_snapshot for the
// user using the resolved settings (timezone + target). It is idempotent via
// the unique (user_id, local_date) index. Returns the (existing or new)
// snapshot.
func (s *Service) EnsureTodaySnapshot(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	resolved gamification.ResolvedSettings,
	now time.Time,
) (*DailyMissionSnapshot, error) {
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	today, err := gamification.LocalDate(now, resolved.Timezone)
	if err != nil {
		return nil, err
	}
	return s.missions.CreateDailyMissionSnapshot(
		ctx, tx, userID, today, resolved.Timezone,
		resolved.DailyReviewTarget, gamification.MissionPolicyVersion,
	)
}

// IncrementReviewsCompleted is a thin service-layer wrapper over the
// repository method, used by transaction-scoped callers (P1/P2/P3
// transactions) that need to record one review attempt against today's
// daily mission and the user's activity summary. The reviews_completed
// counter is capped at review_target by the SQL; the returned value is
// the post-increment counter.
func (s *Service) IncrementReviewsCompleted(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	reviewTarget int,
	correct bool,
	skipped bool,
) (int, error) {
	return s.missions.IncrementReviewsCompleted(
		ctx, tx, userID, localDate, timezone, reviewTarget, correct, skipped,
	)
}

// IncrementWordsAdded records a successful word addition in the daily
// activity summary. New-word mission goals remain optional, so callers pass
// whether the policy version in use has enabled that bonus.
func (s *Service) IncrementWordsAdded(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	includeNewWordGoal bool,
) error {
	return s.missions.IncrementWordsAdded(
		ctx, tx, userID, localDate, timezone, includeNewWordGoal,
	)
}

// RecordConfidencePointChange writes a signed ledger change to today's earned
// or spent daily activity counter.
func (s *Service) RecordConfidencePointChange(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	amount int,
) error {
	return s.missions.RecordConfidencePointChange(
		ctx, tx, userID, localDate, timezone, amount,
	)
}

// MarkSnapshotCompleted transitions today's daily_mission_snapshots row to
// status='completed' inside tx. The repository's WHERE status='open' guard
// makes this idempotent — a retried or replayed transaction can never
// double-complete the mission or double-award the +10 reward. The returned
// bool is true iff the row was actually transitioned (i.e. the call
// completed the mission for the first time today).
func (s *Service) MarkSnapshotCompleted(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	now time.Time,
) (bool, error) {
	return s.missions.MarkSnapshotCompleted(ctx, tx, userID, localDate, now)
}

// MarkSnapshotProtected links a grace-day ledger debit to the exact missed
// local-day snapshot that it protects. It is conditional on status='missed',
// so a retry cannot protect the day twice.
func (s *Service) MarkSnapshotProtected(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	graceDayID uuid.UUID,
) (bool, error) {
	return s.missions.MarkSnapshotProtected(ctx, tx, userID, localDate, graceDayID)
}

// GetDailyMissionView returns the API view of today's daily mission for the
// user, including the shared streak object. The clientTimezone is the
// optional request-time IANA timezone from the caller (validated by
// gamification.GetSettings, per VOC-030-D01). If today's snapshot does not
// yet exist, one is created lazily inside a short read transaction
// (DOC-06 §10 lazy-snapshot-creation pattern). The lazy creation runs
// streak reconciliation so the returned StreakView reflects a fresh
// at-risk/broken state rather than a stale "active" state on a multi-day
// gap (VOC-030-R06).
func (s *Service) GetDailyMissionView(
	ctx context.Context,
	userID uuid.UUID,
	clientTimezone string,
	now time.Time,
) (*DailyMissionView, error) {
	resolved, err := s.gamification.GetSettings(ctx, userID, clientTimezone)
	if err != nil {
		return nil, err
	}
	today, err := gamification.LocalDate(now, resolved.Timezone)
	if err != nil {
		return nil, err
	}
	snap, err := s.ensureTodaySnapshotAndReconcile(ctx, userID, resolved, today, now)
	if err != nil {
		return nil, err
	}
	streak, graceBalance, err := s.loadStreakAndGrace(ctx, userID, resolved.Timezone)
	if err != nil {
		return nil, err
	}
	view := &DailyMissionView{
		LocalDate:       today,
		Timezone:        resolved.Timezone,
		Streak:          streak,
		GraceDayBalance: graceBalance,
	}
	view.ReviewTarget = snap.ReviewTarget
	view.ReviewsCompleted = snap.ReviewsCompleted
	view.NewWordTarget = snap.NewWordTarget
	view.NewWordsCompleted = snap.NewWordsCompleted
	view.SentencePracticeTarget = snap.SentencePracticeTarget
	view.SentencePracticesCompleted = snap.SentencePracticesCompleted
	view.PolicyVersion = snap.PolicyVersion
	view.Status = snap.Status
	view.CompletedAt = snap.CompletedAt
	view.GraceApplied = snap.GraceApplied
	return view, nil
}

// ensureTodaySnapshotAndReconcile establishes today's stable snapshot and
// reconciles a stale streak together when a learner first reads either Home's
// mission or Progress on a new local day. Keeping the two writes in one
// transaction prevents either read surface from exposing a stale streak while
// preserving the established lazy, idempotent snapshot behavior.
func (s *Service) ensureTodaySnapshotAndReconcile(
	ctx context.Context,
	userID uuid.UUID,
	resolved gamification.ResolvedSettings,
	today time.Time,
	now time.Time,
) (*DailyMissionSnapshot, error) {
	tx, err := s.missions.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	snap, err := s.missions.GetDailyMissionSnapshot(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	if snap == nil {
		snap, err = s.missions.CreateDailyMissionSnapshot(
			ctx, tx, userID, today, resolved.Timezone,
			resolved.DailyReviewTarget, gamification.MissionPolicyVersion,
		)
		if err != nil {
			return nil, err
		}
		// Streak reconciliation on first read of a new local day so the
		// returned StreakView reflects the user's actual streak state
		// (a multi-day gap, e.g. day 1 completed and now reading on day 5,
		// must surface as "broken"/0, not a stale "active"). The
		// ReconcileAndAdvance helper is itself a no-op when there is
		// nothing to transition, so this is cheap for the common case.
		snaps, err := s.missions.ListRecentSnapshots(ctx, userID, 14)
		if err != nil {
			return nil, fmt.Errorf("list recent snapshots: %w", err)
		}
		streakSnaps := make([]gamification.StreakSnapshot, 0, len(snaps))
		for _, s := range snaps {
			streakSnaps = append(streakSnaps, gamification.StreakSnapshot{
				LocalDate:    s.LocalDate,
				Status:       s.Status,
				CompletedAt:  s.CompletedAt,
				GraceApplied: s.GraceApplied,
				GraceDayID:   s.GraceDayID,
			})
		}
		if _, err := s.gamification.ReconcileAndAdvance(
			ctx, tx, userID, now, resolved.Timezone,
			streakSnaps, false,
		); err != nil {
			return nil, fmt.Errorf("reconcile streak: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return snap, nil
}

// GetProgressView returns the API view of the user's overall progress
// (Confidence Points balance, shared streak, bounded 7-day completion
// history). Home and Progress read the same streak source via this method.
// The clientTimezone is the optional request-time IANA timezone from the
// caller (validated by gamification.GetSettings, per VOC-030-D01).
func (s *Service) GetProgressView(
	ctx context.Context,
	userID uuid.UUID,
	clientTimezone string,
	now time.Time,
	historyDays int,
) (*ProgressView, error) {
	resolved, err := s.gamification.GetSettings(ctx, userID, clientTimezone)
	if err != nil {
		return nil, err
	}
	today, err := gamification.LocalDate(now, resolved.Timezone)
	if err != nil {
		return nil, err
	}
	if _, err := s.ensureTodaySnapshotAndReconcile(ctx, userID, resolved, today, now); err != nil {
		return nil, err
	}
	balance, err := s.gamification.CurrentBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	streak, graceBalance, err := s.loadStreakAndGrace(ctx, userID, resolved.Timezone)
	if err != nil {
		return nil, err
	}
	startDate := today.AddDate(0, 0, 1-historyDays)
	days, err := s.missions.ListRecentCompletionHistory(ctx, userID, startDate, today)
	if err != nil {
		return nil, err
	}
	view := &ProgressView{
		ConfidencePointsBalance: balance,
		Streak:                  streak,
		GraceDayBalance:         graceBalance,
		CompletionHistory:       make([]CompletionHistoryEntry, 0, historyDays),
	}
	for _, d := range days {
		view.CompletionHistory = append(view.CompletionHistory, CompletionHistoryEntry{
			LocalDate: d.LocalDate,
			Completed: d.Status == StatusCompleted || d.Status == StatusProtected,
		})
	}
	return view, nil
}

// DailyMissionView is the public API DTO for GET /api/v1/daily-mission.
type DailyMissionView struct {
	LocalDate                  time.Time
	Timezone                   string
	ReviewTarget               int
	ReviewsCompleted           int
	NewWordTarget              *int
	NewWordsCompleted          *int
	SentencePracticeTarget     *int
	SentencePracticesCompleted *int
	PolicyVersion              string
	Status                     string
	CompletedAt                *time.Time
	GraceApplied               bool
	Streak                     StreakView
	GraceDayBalance            int
}

// ProgressView is the public API DTO for GET /api/v1/progress.
type ProgressView struct {
	ConfidencePointsBalance int
	Streak                  StreakView
	GraceDayBalance         int
	CompletionHistory       []CompletionHistoryEntry
}

// CompletionHistoryEntry is one day in the bounded 7-day history.
type CompletionHistoryEntry struct {
	LocalDate time.Time
	Completed bool
}

// StreakView is the shared streak object that backs both
// GET /api/v1/daily-mission and GET /api/v1/progress.
type StreakView struct {
	CurrentStreakCount int
	LongestStreakCount int
	Status             string
	GraceDayBalance    int
}

func (s *Service) loadStreakAndGrace(ctx context.Context, userID uuid.UUID, timezone string) (StreakView, int, error) {
	// Lazy default if no streak row exists yet.
	view := StreakView{
		CurrentStreakCount: 0,
		LongestStreakCount: 0,
		Status:             gamification.StreakStatusActive,
	}
	streak, err := s.gamification.GetStreakStateForRead(ctx, userID)
	if err != nil {
		return view, 0, err
	}
	if streak != nil {
		view.CurrentStreakCount = streak.CurrentStreakCount
		view.LongestStreakCount = streak.LongestStreakCount
		view.Status = streak.Status
		if streak.Timezone != "" {
			timezone = streak.Timezone
		}
	}
	grace, err := s.gamification.CurrentGraceBalance(ctx, userID)
	if err != nil {
		return view, 0, err
	}
	view.GraceDayBalance = grace
	_ = timezone
	return view, grace, nil
}

// MissionUpdater is the real implementation of the aifeedback.MissionUpdater
// interface. It is wired in apps/api/business/aifeedback/service.go in
// T03; this file lives in the missions package so the transaction logic
// stays co-located with the rest of the missions domain.
type MissionUpdater struct {
	missions     *Service
	gamification *gamification.Service
}

// NewMissionUpdater returns a real MissionUpdater.
func NewMissionUpdater(m *Service, g *gamification.Service) *MissionUpdater {
	return &MissionUpdater{missions: m, gamification: g}
}

// Update implements the aifeedback.MissionUpdater interface (see that
// package's mission.go). It resolves the caller's settings from stored
// user_settings only - no request-time client timezone is available at
// this seam, unlike an HTTP-request-scoped read - falling back through the
// same D01 chain's remaining UTC/default step, and applies the VOC-030-D03
// policy decision (bonus sentence-practice mission goal disabled at
// launch) before delegating to UpdateForSentence.
func (u *MissionUpdater) Update(ctx context.Context, userID, sentenceID, attemptID uuid.UUID) (bool, error) {
	return u.UpdateInTransaction(ctx, nil, userID, sentenceID, attemptID)
}

// UpdateInTransaction applies sentence feedback accounting inside a caller's
// transaction. sentenceID owns the one submission reward; attemptID owns the
// per-successful-feedback reward. A nil transaction preserves the standalone
// Update entry point.
func (u *MissionUpdater) UpdateInTransaction(ctx context.Context, tx *sql.Tx, userID, sentenceID, attemptID uuid.UUID) (bool, error) {
	resolved, err := u.gamification.GetSettings(ctx, userID, "")
	if err != nil {
		return false, err
	}
	const includeSentenceGoal = false // VOC-030-D03: bonus goals disabled at launch
	return u.updateForSentence(ctx, tx, userID, sentenceID, attemptID, resolved, time.Now(), includeSentenceGoal)
}

// UpdateForSentence is the transaction-aware entry point Update above
// delegates to. It runs the sentence-submitted and AI-feedback-received
// point awards, increments the activity summary, and (if the optional
// sentence-practice mission goal is active) the mission counter. It also
// runs streak reconciliation so the read APIs' shared StreakView stays
// consistent with backend state (e.g. at_risk when today isn't yet
// completed). missionCompleted returns true iff the call completed the
// daily mission for the first time today (i.e. the snapshot transitioned
// to status='completed'). On the P3 path, reviews_completed is never
// incremented (only the P2 review path does that), so the structural
// return value is always false here; the explicit assignment replaces
// the earlier always-false hardcode so the function is honest about the
// path invariant. Exposed separately from Update so a caller that
// already has a resolved settings/clock value (e.g. a future API-layer
// caller) doesn't pay for a second settings lookup.
func (u *MissionUpdater) UpdateForSentence(
	ctx context.Context,
	userID, sentenceID, attemptID uuid.UUID,
	resolved gamification.ResolvedSettings,
	now time.Time,
	includeSentenceGoal bool,
) (bool, error) {
	return u.updateForSentence(ctx, nil, userID, sentenceID, attemptID, resolved, now, includeSentenceGoal)
}

func (u *MissionUpdater) updateForSentence(ctx context.Context, tx *sql.Tx, userID, sentenceID, attemptID uuid.UUID, resolved gamification.ResolvedSettings, now time.Time, includeSentenceGoal bool) (bool, error) {
	ownedTx := tx == nil
	var err error
	if ownedTx {
		tx, err = u.missions.missions.db.BeginTx(ctx, nil)
		if err != nil {
			return false, fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()
	}

	// Ensure today's snapshot exists.
	snap, err := u.missions.EnsureTodaySnapshot(ctx, tx, userID, resolved, now)
	if err != nil {
		return false, err
	}

	// Lock and read the current balance inside this award transaction.
	balance, err := u.gamification.CurrentBalanceTx(ctx, tx, userID)
	if err != nil {
		return false, err
	}

	// Sentence-submitted award.
	newBalance, _, sentenceAwarded, err := u.gamification.GrantPoint(
		ctx, tx, userID, gamification.RewardKindSentenceSubmitted,
		&sentenceID, gamification.LearnerSentenceSubmittedKey(sentenceID.String()),
		balance, now, nil,
	)
	if err != nil {
		return false, err
	}

	// AI-feedback-received award.
	_, _, feedbackAwarded, err := u.gamification.GrantPoint(
		ctx, tx, userID, gamification.RewardKindAIFeedbackGot,
		&attemptID, gamification.AIFeedbackAttemptReceivedKey(attemptID.String()),
		newBalance, now, nil,
	)
	if err != nil {
		return false, err
	}

	// Activity counters.
	if err := u.missions.missions.IncrementSentenceSubmitted(
		ctx, tx, userID, snap.LocalDate, resolved.Timezone, includeSentenceGoal,
	); err != nil {
		return false, err
	}
	if err := u.missions.missions.IncrementAIFeedbackReceived(
		ctx, tx, userID, snap.LocalDate, resolved.Timezone,
	); err != nil {
		return false, err
	}
	earned := 0
	if sentenceAwarded {
		earned += gamification.RewardSentenceSubmitted
	}
	if feedbackAwarded {
		earned += gamification.RewardAIFeedbackGot
	}
	if earned != 0 {
		if err := u.missions.missions.RecordConfidencePointChange(
			ctx, tx, userID, snap.LocalDate, resolved.Timezone, earned,
		); err != nil {
			return false, err
		}
	}

	// Streak reconciliation: lazy, no queue/cron (DOC-06 §15), computed
	// from the recent snapshot history. P3 never increments
	// reviews_completed and never transitions the snapshot, so the
	// mission/streak transition never fires here; the reconciliation
	// still runs so the streak status reflects "at_risk" honestly if
	// today isn't yet completed. This keeps Home/Progress reads
	// consistent with backend state.
	snaps, err := u.missions.missions.ListRecentSnapshots(ctx, userID, 14)
	if err != nil {
		return false, fmt.Errorf("list recent snapshots: %w", err)
	}
	streakSnaps := make([]gamification.StreakSnapshot, 0, len(snaps))
	for _, s := range snaps {
		streakSnaps = append(streakSnaps, gamification.StreakSnapshot{
			LocalDate:    s.LocalDate,
			Status:       s.Status,
			CompletedAt:  s.CompletedAt,
			GraceApplied: s.GraceApplied,
			GraceDayID:   s.GraceDayID,
		})
	}
	if _, err := u.gamification.ReconcileAndAdvance(
		ctx, tx, userID, now, resolved.Timezone,
		streakSnaps, false, // currentCompletion: P3 never completes the mission
	); err != nil {
		return false, fmt.Errorf("reconcile streak: %w", err)
	}

	// P3 invariant: only the P2 review path increments reviews_completed
	// and marks the snapshot completed, so the structural return value
	// here is always false. The explicit assignment replaces the prior
	// always-false hardcode with the honest path-invariant value.
	missionCompleted := false
	if ownedTx {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit: %w", err)
		}
	}
	return missionCompleted, nil
}
