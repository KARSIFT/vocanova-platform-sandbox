package reviews

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// PostgreSQLRepository implements Repository against the VOC-027 review schema.
// Optional gamification and missions dependencies enable the P4 reward, mission,
// and streak wiring inside the existing review-submission transaction
// (DOC-06 §3). When both are nil, the pre-existing P2 behavior is preserved
// byte-for-byte (no extra SQL, no extra writes) — this is the contract the
// existing reviews package tests already rely on.
type PostgreSQLRepository struct {
	db           *sql.DB
	clock        clock.Clock
	gamification *gamification.Service
	missions     *missions.Service
}

// ReviewRepositoryOption configures a PostgreSQLRepository at construction
// time. Typed options keep the constructor type-safe while remaining
// backwards-compatible with existing two-argument call sites that only
// supply db + clock.
type ReviewRepositoryOption func(*PostgreSQLRepository)

// WithGamificationService wires in the gamification service for P4 reward
// and streak writes inside the review transaction.
func WithGamificationService(s *gamification.Service) ReviewRepositoryOption {
	return func(r *PostgreSQLRepository) { r.gamification = s }
}

// WithMissionsService wires in the missions service for daily-mission
// snapshot and activity-summary writes inside the review transaction.
func WithMissionsService(s *missions.Service) ReviewRepositoryOption {
	return func(r *PostgreSQLRepository) { r.missions = s }
}

// NewPostgreSQLRepository creates a repository backed by db. c may be nil
// (defaults to clock.Real). Optional gamification/missions dependencies
// enable the P4 reward wiring; both are independent and may be supplied
// together or left nil to keep the pre-P2 behavior unchanged.
func NewPostgreSQLRepository(db *sql.DB, c clock.Clock, opts ...ReviewRepositoryOption) *PostgreSQLRepository {
	if c == nil {
		c = clock.Real{}
	}
	repo := &PostgreSQLRepository{db: db, clock: c}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

// HasP4Wiring reports whether both P4 dependencies are wired so SubmitReview
// will run applyP4ReviewWiring. The production composition root must wire both.
func (r *PostgreSQLRepository) HasP4Wiring() bool {
	return r.gamification != nil && r.missions != nil
}

func (r *PostgreSQLRepository) ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	var cursorNextReviewAt sql.NullTime
	var cursorID uuid.UUID
	if req.AfterCursor != "" {
		c, err := decodeDueCursor(req.AfterCursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		cursorID = c.ID
		if !c.NextReviewAt.IsZero() {
			cursorNextReviewAt = sql.NullTime{Time: c.NextReviewAt, Valid: true}
		}
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT * FROM (SELECT
		   uw.id AS user_word_id, uw.meaning_id, cw.id AS word_id, cw.text, cw.normalized_text,
		   wm.part_of_speech, wm.short_definition, uw.status, uw.source, uw.review_step,
		   uw.next_review_at,
		   COUNT(*) OVER() AS total_count
		 FROM user_words uw
		 JOIN word_meanings wm ON wm.id = uw.meaning_id
		 JOIN canonical_words cw ON cw.id = wm.word_id
		 WHERE uw.user_id = $1
		   AND uw.deleted_at IS NULL
		   AND uw.status IN ('new', 'learning', 'reviewing')
		   AND (uw.next_review_at IS NULL OR uw.next_review_at <= NOW())
		 ) AS due
		 WHERE ($3::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR
		        (COALESCE(due.next_review_at, '-infinity'::timestamptz), due.user_word_id) >
		        (COALESCE($2::timestamptz, '-infinity'::timestamptz), $3::uuid))
		 ORDER BY COALESCE(due.next_review_at, '-infinity'::timestamptz) ASC, due.user_word_id ASC
		 LIMIT $4`,
		req.UserID, cursorNextReviewAt, cursorID, limit+1,
	)
	if err != nil {
		return nil, fmt.Errorf("list due words: %w", err)
	}
	defer rows.Close()

	var items []DueWord
	var totalCount int
	for rows.Next() {
		var d DueWord
		var normalizedText string
		var nextReviewAt sql.NullTime
		if err := rows.Scan(
			&d.UserWordID, &d.MeaningID, &d.WordID, &d.WordText, &normalizedText,
			&d.PartOfSpeech, &d.ShortDefinition, &d.Status, &d.Source, &d.ReviewStep,
			&nextReviewAt, &totalCount,
		); err != nil {
			return nil, fmt.Errorf("scan due word: %w", err)
		}
		d.WordSlug = wordSlug(normalizedText)
		if nextReviewAt.Valid {
			d.NextReviewAt = &nextReviewAt.Time
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due words rows: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close due words rows: %w", err)
	}
	// No row carries the window count when a cursor is exhausted.
	if len(items) == 0 && req.AfterCursor != "" {
		err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_words uw
		 JOIN word_meanings wm ON wm.id = uw.meaning_id
		 JOIN canonical_words cw ON cw.id = wm.word_id
		 WHERE uw.user_id = $1 AND uw.deleted_at IS NULL
		 AND uw.status IN ('new', 'learning', 'reviewing')
		 AND (uw.next_review_at IS NULL OR uw.next_review_at <= NOW())`, req.UserID).Scan(&totalCount)
		if err != nil {
			return nil, fmt.Errorf("count due words: %w", err)
		}
	}

	resp := &ListDueWordsResponse{Items: items, TotalCount: totalCount}
	if len(items) > limit {
		resp.Items = items[:limit]
		last := resp.Items[limit-1]
		cursor := dueCursor{ID: last.UserWordID}
		if last.NextReviewAt != nil {
			cursor.NextReviewAt = *last.NextReviewAt
		}
		resp.NextCursor = encodeDueCursor(cursor)
	}
	return resp, nil
}

func (r *PostgreSQLRepository) GetReviewAttemptByClientAttemptID(ctx context.Context, userID uuid.UUID, clientAttemptID string) (*ReviewAttempt, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT ra.id, ra.user_word_id, ra.meaning_id, ra.attempt_type, ra.prompt_type,
		        ra.result, ra.rating, ra.review_step_before, ra.review_step_after,
		        ra.answered_at, ra.response_time_ms, ra.selected_option_meaning_id,
		        ra.typed_answer, ra.was_hint_used, ra.source, ra.client_attempt_id,
		        ra.metadata, uw.next_review_at
		 FROM review_attempts ra
		 JOIN user_words uw ON uw.id = ra.user_word_id
		 WHERE ra.user_id = $1 AND ra.client_attempt_id = $2`,
		userID, clientAttemptID,
	)
	attempt, err := r.scanReviewAttempt(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReviewAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetch review attempt: %w", err)
	}
	return attempt, nil
}

func (r *PostgreSQLRepository) SubmitReview(ctx context.Context, req SubmitReviewRequest) (*ReviewAttempt, error) {
	if req.IdempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequired
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var priorStep, totalCount, correctCount, consecutiveCorrect, consecutiveIncorrect int
	var meaningID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`SELECT review_step, meaning_id, total_review_count, correct_review_count,
		        consecutive_correct_count, consecutive_incorrect_count
		 FROM user_words
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 FOR UPDATE`,
		req.UserWordID, req.UserID,
	).Scan(&priorStep, &meaningID, &totalCount, &correctCount, &consecutiveCorrect, &consecutiveIncorrect)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserWordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock user word: %w", err)
	}
	if meaningID != req.MeaningID {
		return nil, ErrUserWordNotFound
	}

	// Claim inside the existing transaction, after ownership validation and
	// before any effects. The unique index waits for concurrent claimants; an
	// active conflicting fingerprint must not reach scheduling or rewards.
	now := r.clock.Now().UTC()
	fingerprint := submitReviewFingerprint(req)
	var storedFingerprint string
	err = tx.QueryRowContext(ctx, `INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, operation, key) DO UPDATE
		SET fingerprint = EXCLUDED.fingerprint, created_at = EXCLUDED.created_at
		WHERE idempotency_keys.created_at <= $7
		RETURNING fingerprint`, uuid.New(), req.UserID, operationSubmitReview, req.IdempotencyKey,
		fingerprint, now, now.Add(-reviewIdempotencyRetention)).Scan(&storedFingerprint)
	activeKey := errors.Is(err, sql.ErrNoRows)
	if activeKey {
		err = tx.QueryRowContext(ctx, `SELECT fingerprint FROM idempotency_keys
			WHERE user_id = $1 AND operation = $2 AND key = $3 FOR UPDATE`,
			req.UserID, operationSubmitReview, req.IdempotencyKey).Scan(&storedFingerprint)
	}
	if err != nil {
		return nil, fmt.Errorf("claim review idempotency: %w", err)
	}
	if activeKey && storedFingerprint != fingerprint {
		return nil, ErrIdempotencyConflict
	}

	// Idempotency guard on (user_id, client_attempt_id).
	existing, err := r.fetchAttemptByClientAttemptID(ctx, tx, req.UserID, req.ClientAttemptID)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if existing != nil {
		if reviewAttemptEqualsRequest(existing, req) {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("commit idempotent: %w", err)
			}
			return existing, nil
		}
		return nil, ErrIdempotencyConflict
	}
	if activeKey {
		return nil, ErrReviewAttemptNotFound
	}

	prior := ReviewState{
		ReviewStep: priorStep,
		Counters: ReviewCounters{
			TotalReviewCount:          totalCount,
			CorrectReviewCount:        correctCount,
			ConsecutiveCorrectCount:   consecutiveCorrect,
			ConsecutiveIncorrectCount: consecutiveIncorrect,
		},
	}
	scheduledAt := req.scheduledAt
	if scheduledAt.IsZero() {
		scheduledAt = r.clock.Now().UTC()
	}
	sched, err := ApplyReview(prior, ApplyReviewRequest{Result: req.Result, Rating: req.Rating}, scheduledAt)
	if err != nil {
		return nil, err
	}

	attemptID := uuid.New()
	selectedOpt := sql.NullString{}
	if req.SelectedOptionMeaningID != nil {
		selectedOpt = sql.NullString{String: req.SelectedOptionMeaningID.String(), Valid: true}
	}
	typedAnswer := sql.NullString{}
	if req.TypedAnswer != nil {
		typedAnswer = sql.NullString{String: *req.TypedAnswer, Valid: true}
	}
	rating := sql.NullString{}
	if req.Rating != "" {
		rating = sql.NullString{String: req.Rating, Valid: true}
	}
	var metadata any
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}
		metadata = b
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO review_attempts (
			id, user_id, user_word_id, meaning_id, attempt_type, prompt_type, result, rating,
			review_step_before, review_step_after, answered_at, response_time_ms,
			selected_option_meaning_id, typed_answer, was_hint_used, source, client_attempt_id,
			metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19, $20
		)`,
		attemptID, req.UserID, req.UserWordID, req.MeaningID, req.AttemptType, req.PromptType, req.Result, rating,
		prior.ReviewStep, sched.ReviewStep, req.AnsweredAt, req.ResponseTimeMs,
		selectedOpt, typedAnswer, req.WasHintUsed, req.Source, req.ClientAttemptID,
		metadata, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert review attempt: %w", err)
	}

	lastResult := sql.NullString{String: sched.LastResult, Valid: true}
	lastRating := sql.NullString{String: sched.LastRating, Valid: true}
	_, err = tx.ExecContext(ctx,
		`UPDATE user_words
		 SET review_step = $1,
		     next_review_at = $2,
		     last_reviewed_at = $3,
		     last_result = $4,
		     last_rating = $5,
		     total_review_count = $6,
		     correct_review_count = $7,
		     consecutive_correct_count = $8,
		     consecutive_incorrect_count = $9,
		     updated_at = $10
		 WHERE id = $11 AND user_id = $12`,
		sched.ReviewStep, sched.NextReviewAt, sched.LastReviewedAt, lastResult, lastRating,
		sched.Counters.TotalReviewCount, sched.Counters.CorrectReviewCount,
		sched.Counters.ConsecutiveCorrectCount, sched.Counters.ConsecutiveIncorrectCount,
		now, req.UserWordID, req.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("update user word schedule: %w", err)
	}

	// P4 reward wiring: record the rating-tiered point award, update the daily
	// activity summary, advance the daily mission, and reconcile the streak —
	// all inside the same transaction. Both dependencies are required to wire
	// anything; either being nil preserves the pre-P4 P2 behavior. A skipped
	// review awards no rating-tiered point (DOC-06 §11 prices only
	// correct/incorrect ratings); the activity summary still gets the
	// reviews_skipped counter and a non-zero attempted event so the read APIs
	// can show "skipped today" honestly.
	if r.gamification != nil && r.missions != nil {
		if err := r.applyP4ReviewWiring(ctx, tx, req, attemptID, now); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit review submission: %w", err)
	}

	return &ReviewAttempt{
		ID:                      attemptID,
		UserID:                  req.UserID,
		UserWordID:              req.UserWordID,
		MeaningID:               req.MeaningID,
		AttemptType:             req.AttemptType,
		PromptType:              req.PromptType,
		Result:                  req.Result,
		Rating:                  req.Rating,
		ReviewStepBefore:        prior.ReviewStep,
		ReviewStepAfter:         sched.ReviewStep,
		AnsweredAt:              req.AnsweredAt,
		ResponseTimeMs:          req.ResponseTimeMs,
		SelectedOptionMeaningID: req.SelectedOptionMeaningID,
		TypedAnswer:             req.TypedAnswer,
		WasHintUsed:             req.WasHintUsed,
		Source:                  req.Source,
		ClientAttemptID:         req.ClientAttemptID,
		Metadata:                req.Metadata,
		NextReviewAt:            sched.NextReviewAt,
	}, nil
}

// applyP4ReviewWiring runs the P4 reward/mission/streak writes for one
// successful review submission. It is invoked strictly after the existing
// review_attempts INSERT and user_words UPDATE and strictly before the
// existing commit, so all writes either land together or roll back together.
// Idempotency is enforced by the confidence_point_ledger
// (user_id, idempotency_key) partial unique index and the daily_mission_snapshots
// status='open' guard in MarkSnapshotCompleted — even a retried/replayed
// transaction can never award a second reward or double-complete the mission.
func (r *PostgreSQLRepository) applyP4ReviewWiring(
	ctx context.Context,
	tx *sql.Tx,
	req SubmitReviewRequest,
	attemptID uuid.UUID,
	now time.Time,
) error {
	// Resolve per-user settings (timezone + review target) using the
	// user_settings row first, then the request-time client IANA timezone
	// (empty here — P2's SubmitReview is the caller's own transaction, not
	// an HTTP request), falling back to UTC/default. D01 governs.
	resolved, err := r.gamification.GetSettings(ctx, req.UserID, "")
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}

	// Ensure today's snapshot exists (lazy creation per DOC-06 §10).
	snap, err := r.missions.EnsureTodaySnapshot(ctx, tx, req.UserID, resolved, now)
	if err != nil {
		return fmt.Errorf("ensure today snapshot: %w", err)
	}

	// Map review result/rating to the P4 activity flags used by both the
	// daily_activity_summaries row and the rating-tiered reward grant.
	correct, skipped, rewardKind, ok := reviewRewardKind(req.Result, req.Rating)
	if !ok {
		// Unrecognized result/rating — should be unreachable because
		// Service.SubmitReview already validates, but be defensive.
		return fmt.Errorf("review reward kind: result=%q rating=%q", req.Result, req.Rating)
	}

	// Increment reviews_completed on the snapshot (capped at review_target
	// by LEAST) and upsert the activity summary's review counters.
	newReviewsCompleted, err := r.missions.IncrementReviewsCompleted(
		ctx, tx, req.UserID, snap.LocalDate, resolved.Timezone,
		resolved.DailyReviewTarget, correct, skipped,
	)
	if err != nil {
		return fmt.Errorf("increment reviews completed: %w", err)
	}

	// Skipped reviews do not award a rating-tiered point (DOC-06 §11
	// prices only Again/Hard/Good/Easy, all of which require a rating)
	// but the activity counter above already recorded the attempt.
	if !skipped {
		balance, err := getLatestPointBalanceTx(ctx, tx, req.UserID)
		if err != nil {
			return fmt.Errorf("get latest point balance: %w", err)
		}
		newBalance, _, err := r.gamification.GrantPoint(
			ctx, tx, req.UserID,
			rewardKind, &attemptID,
			gamification.ReviewAttemptRatedKey(attemptID.String()),
			balance, now, nil,
		)
		if err != nil {
			return fmt.Errorf("grant review point: %w", err)
		}
		if err := r.missions.IncrementConfidencePointsEarned(
			ctx, tx, req.UserID, snap.LocalDate, resolved.Timezone, newBalance-balance,
		); err != nil {
			return fmt.Errorf("increment points earned: %w", err)
		}
	}

	// If the review target is met for the first time today, transition
	// the snapshot to completed, record the +10 daily-mission-completion
	// award, and let the streak reconciliation know the day completed
	// (so it advances the streak and may earn a grace day).
	missionCompletedNow := false
	if newReviewsCompleted >= snap.ReviewTarget && snap.Status == missions.StatusOpen {
		completed, err := r.missions.MarkSnapshotCompleted(ctx, tx, req.UserID, snap.LocalDate, now)
		if err != nil {
			return fmt.Errorf("mark snapshot completed: %w", err)
		}
		if completed {
			missionCompletedNow = true
			balance, err := getLatestPointBalanceTx(ctx, tx, req.UserID)
			if err != nil {
				return fmt.Errorf("get latest point balance: %w", err)
			}
			localDateKey := snap.LocalDate.Format("2006-01-02")
			if _, _, err := r.gamification.GrantPoint(
				ctx, tx, req.UserID,
				gamification.RewardKindDailyMissionDone, nil,
				gamification.DailyMissionCompletedKey(req.UserID.String(), localDateKey),
				balance, now, nil,
			); err != nil {
				return fmt.Errorf("grant daily mission point: %w", err)
			}
			if err := r.missions.IncrementConfidencePointsEarned(
				ctx, tx, req.UserID, snap.LocalDate, resolved.Timezone, gamification.RewardDailyMissionDone,
			); err != nil {
				return fmt.Errorf("increment points earned: %w", err)
			}
		}
	}

	// Streak reconciliation: lazy, no queue/cron (DOC-06 §15), computed
	// from the recent snapshot history. Reads happen on the caller's tx
	// (consistent read) and any grace-day ledger writes happen here too.
	snaps, err := r.fetchStreakSnapshotsTx(ctx, tx, req.UserID)
	if err != nil {
		return fmt.Errorf("fetch streak snapshots: %w", err)
	}
	graceBalance, err := getLatestGraceBalanceTx(ctx, tx, req.UserID)
	if err != nil {
		return fmt.Errorf("get latest grace balance: %w", err)
	}
	if _, err := r.gamification.ReconcileAndAdvance(
		ctx, tx, req.UserID, now, resolved.Timezone,
		snaps, graceBalance, missionCompletedNow,
	); err != nil {
		return fmt.Errorf("reconcile streak: %w", err)
	}
	return nil
}

// reviewRewardKind maps a (result, rating) pair to the per-event reward
// kind and the activity summary flags. Skipped reviews return ok=false for
// the reward kind (no rating-tiered award applies) but still record the
// attempt in the activity summary. Unknown pairs are returned with ok=false
// so the caller can fail closed without silently awarding 0 points.
func reviewRewardKind(result, rating string) (correct, skipped bool, kind gamification.RewardKind, ok bool) {
	switch result {
	case ResultCorrect:
		correct = true
		switch rating {
		case RatingHard:
			return correct, false, gamification.RewardKindReviewHard, true
		case RatingGood:
			return correct, false, gamification.RewardKindReviewGood, true
		case RatingEasy:
			return correct, false, gamification.RewardKindReviewEasy, true
		}
		// Defensive: a correct attempt with empty rating is treated
		// as Good by ApplyReview, so mirror that mapping here.
		if rating == "" {
			return correct, false, gamification.RewardKindReviewGood, true
		}
		return false, false, "", false
	case ResultIncorrect:
		// Again is the only permitted rating for an incorrect attempt.
		if rating == RatingAgain {
			return false, false, gamification.RewardKindReviewAgain, true
		}
		return false, false, "", false
	case ResultSkipped:
		// Skipped attempts have no rating; no rating-tiered reward.
		return false, true, "", true
	}
	return false, false, "", false
}

// fetchStreakSnapshotsTx reads the recent daily_mission_snapshots history
// for the user (a 14-day window is sufficient for streak reconciliation)
// inside the caller's tx so the read is consistent with the writes just
// performed by the P4 wiring block.
func (r *PostgreSQLRepository) fetchStreakSnapshotsTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) ([]gamification.StreakSnapshot, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT local_date, status, completed_at, grace_applied, grace_day_id
		 FROM daily_mission_snapshots
		 WHERE user_id = $1
		 ORDER BY local_date DESC
		 LIMIT $2`,
		userID, 14,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch streak snapshots: %w", err)
	}
	defer rows.Close()
	var out []gamification.StreakSnapshot
	for rows.Next() {
		var s gamification.StreakSnapshot
		var completedAt sql.NullTime
		var graceDayID *uuid.UUID
		if err := rows.Scan(&s.LocalDate, &s.Status, &completedAt, &s.GraceApplied, &graceDayID); err != nil {
			return nil, fmt.Errorf("scan streak snapshot: %w", err)
		}
		if completedAt.Valid {
			t := completedAt.Time
			s.CompletedAt = &t
		}
		if graceDayID != nil {
			id := graceDayID.String()
			s.GraceDayID = &id
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan streak snapshots rows: %w", err)
	}
	return out, nil
}

// getLatestPointBalanceTx reads the user's current confidence-point balance
// inside the caller's tx so the rating-tiered reward grant and the
// daily-mission completion grant see the post-write value (re-using the
// same connection-scoped snapshot, avoiding the @@race window that the
// pre-existing GetLatestPointBalance context-based read would open).
func getLatestPointBalanceTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (int, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT COALESCE(balance_after, 0) FROM confidence_point_ledger
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

// getLatestGraceBalanceTx reads the user's current grace-day balance inside
// the caller's tx (see getLatestPointBalanceTx for the rationale).
func getLatestGraceBalanceTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (int, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT COALESCE(balance_after, 0) FROM grace_day_ledger
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

func (r *PostgreSQLRepository) fetchAttemptByClientAttemptID(ctx context.Context, tx *sql.Tx, userID uuid.UUID, clientAttemptID string) (*ReviewAttempt, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT ra.id, ra.user_word_id, ra.meaning_id, ra.attempt_type, ra.prompt_type,
		        ra.result, ra.rating, ra.review_step_before, ra.review_step_after,
		        ra.answered_at, ra.response_time_ms, ra.selected_option_meaning_id,
		        ra.typed_answer, ra.was_hint_used, ra.source, ra.client_attempt_id,
		        ra.metadata, uw.next_review_at
		 FROM review_attempts ra
		 JOIN user_words uw ON uw.id = ra.user_word_id
		 WHERE ra.user_id = $1 AND ra.client_attempt_id = $2`,
		userID, clientAttemptID,
	)
	attempt, err := r.scanReviewAttempt(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return attempt, nil
}

func (r *PostgreSQLRepository) scanReviewAttempt(row *sql.Row) (*ReviewAttempt, error) {
	var a ReviewAttempt
	var rating, selectedOption, typedAnswer sql.NullString
	var metadata []byte
	var nextReviewAt sql.NullTime
	err := row.Scan(
		&a.ID, &a.UserWordID, &a.MeaningID, &a.AttemptType, &a.PromptType,
		&a.Result, &rating, &a.ReviewStepBefore, &a.ReviewStepAfter,
		&a.AnsweredAt, &a.ResponseTimeMs, &selectedOption,
		&typedAnswer, &a.WasHintUsed, &a.Source, &a.ClientAttemptID,
		&metadata, &nextReviewAt,
	)
	if err != nil {
		return nil, err
	}
	if rating.Valid {
		a.Rating = rating.String
	}
	if selectedOption.Valid {
		id, err := uuid.Parse(selectedOption.String)
		if err != nil {
			return nil, fmt.Errorf("parse selected option meaning id: %w", err)
		}
		a.SelectedOptionMeaningID = &id
	}
	if typedAnswer.Valid {
		a.TypedAnswer = &typedAnswer.String
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &a.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}
	if nextReviewAt.Valid {
		a.NextReviewAt = nextReviewAt.Time
	}
	return &a, nil
}

func reviewAttemptEqualsRequest(a *ReviewAttempt, req SubmitReviewRequest) bool {
	if a.UserWordID != req.UserWordID || a.MeaningID != req.MeaningID {
		return false
	}
	if a.AttemptType != req.AttemptType || a.PromptType != req.PromptType || a.Result != req.Result || a.Rating != req.Rating {
		return false
	}
	if a.ResponseTimeMs != req.ResponseTimeMs || a.WasHintUsed != req.WasHintUsed || a.Source != req.Source {
		return false
	}
	if !a.AnsweredAt.Equal(req.AnsweredAt) {
		return false
	}
	if !ptrUUIDEqual(a.SelectedOptionMeaningID, req.SelectedOptionMeaningID) {
		return false
	}
	if !ptrStringEqual(a.TypedAnswer, req.TypedAnswer) {
		return false
	}
	return true
}
