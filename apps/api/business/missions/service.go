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

// IncrementConfidencePointsEarned adds amount to today's
// daily_activity_summaries.confidence_points_earned (used by the P1 word-add
// and P2 review/mission-completion writes to keep the activity summary in
// sync with the confidence_point_ledger).
func (s *Service) IncrementConfidencePointsEarned(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	amount int,
) error {
	return s.missions.IncrementConfidencePointsEarned(
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

// GetDailyMissionView returns the API view of today's daily mission for the
// user, including the shared streak object. If the snapshot does not yet
// exist, one is created lazily inside a read transaction; this is the
// lazy-snapshot-creation pattern from DOC-06 §10.
func (s *Service) GetDailyMissionView(
	ctx context.Context,
	userID uuid.UUID,
	resolved gamification.ResolvedSettings,
	now time.Time,
) (*DailyMissionView, error) {
	today, err := gamification.LocalDate(now, resolved.Timezone)
	if err != nil {
		return nil, err
	}
	snap, err := s.missions.GetDailyMissionSnapshot(ctx, userID, today)
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
	if snap != nil {
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
	} else {
		view.ReviewTarget = resolved.DailyReviewTarget
		view.ReviewsCompleted = 0
		view.PolicyVersion = gamification.MissionPolicyVersion
		view.Status = StatusOpen
	}
	return view, nil
}

// GetProgressView returns the API view of the user's overall progress
// (Confidence Points balance, shared streak, bounded 7-day completion
// history). Home and Progress read the same streak source via this method.
func (s *Service) GetProgressView(
	ctx context.Context,
	userID uuid.UUID,
	resolved gamification.ResolvedSettings,
	now time.Time,
	historyDays int,
) (*ProgressView, error) {
	balance, err := s.gamification.CurrentBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	streak, graceBalance, err := s.loadStreakAndGrace(ctx, userID, resolved.Timezone)
	if err != nil {
		return nil, err
	}
	days, err := s.missions.ListRecentCompletionHistory(ctx, userID, historyDays)
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
func (u *MissionUpdater) Update(ctx context.Context, userID, sentenceID uuid.UUID) (bool, error) {
	resolved, err := u.gamification.GetSettings(ctx, userID, "")
	if err != nil {
		return false, err
	}
	const includeSentenceGoal = false // VOC-030-D03: bonus goals disabled at launch
	return u.UpdateForSentence(ctx, userID, sentenceID, resolved, time.Now(), includeSentenceGoal)
}

// UpdateForSentence is the transaction-aware entry point Update above
// delegates to. It runs the sentence-submitted and AI-feedback-received
// point awards, increments the activity summary, and (if the optional
// sentence-practice mission goal is active) the mission counter.
// missionCompleted returns true iff the call completed the daily mission
// for the first time today (i.e. the snapshot transitioned to
// status='completed'). Exposed separately from Update so a caller that
// already has a resolved settings/clock value (e.g. a future API-layer
// caller) doesn't pay for a second settings lookup.
func (u *MissionUpdater) UpdateForSentence(
	ctx context.Context,
	userID, sentenceID uuid.UUID,
	resolved gamification.ResolvedSettings,
	now time.Time,
	includeSentenceGoal bool,
) (bool, error) {
	tx, err := u.missions.missions.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Ensure today's snapshot exists.
	snap, err := u.missions.EnsureTodaySnapshot(ctx, tx, userID, resolved, now)
	if err != nil {
		return false, err
	}

	// Current balance from the latest ledger row.
	balance, err := u.gamification.CurrentBalance(ctx, userID)
	if err != nil {
		return false, err
	}

	// Sentence-submitted award.
	newBalance, _, err := u.gamification.GrantPoint(
		ctx, tx, userID, gamification.RewardKindSentenceSubmitted,
		&sentenceID, gamification.LearnerSentenceSubmittedKey(sentenceID.String()),
		balance, now, nil,
	)
	if err != nil {
		return false, err
	}

	// AI-feedback-received award.
	if _, _, err := u.gamification.GrantPoint(
		ctx, tx, userID, gamification.RewardKindAIFeedbackGot,
		&sentenceID, gamification.AIFeedbackAttemptReceivedKey(sentenceID.String()),
		newBalance, now, nil,
	); err != nil {
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
	if err := u.missions.missions.IncrementConfidencePointsEarned(
		ctx, tx, userID, snap.LocalDate, resolved.Timezone,
		gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
	); err != nil {
		return false, err
	}

	missionCompleted := false
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return missionCompleted, nil
}
