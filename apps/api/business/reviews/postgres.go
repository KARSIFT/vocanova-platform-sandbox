package reviews

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/revieweligibility"
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
 AND `+revieweligibility.Due("uw", "CURRENT_TIMESTAMP")+`
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
 AND `+revieweligibility.Due("uw", "CURRENT_TIMESTAMP"), req.UserID).Scan(&totalCount)
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

	// Every review path uses the same user -> session -> word -> mission lock order.
	result, err := tx.ExecContext(ctx, "SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE", req.UserID)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, ErrUserWordNotFound
	}
	attempt, err := r.SubmitReviewTx(ctx, tx, req)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return attempt, nil
}

// SubmitDailyReviewTx binds trusted server grading to the owned session's frozen
// exact saved meaning through the shared membership-checked daily review path.
func (r *PostgreSQLRepository) SubmitDailyReviewTx(ctx context.Context, tx *sql.Tx, req SubmitReviewRequest, sessionID uuid.UUID) (*ReviewAttempt, error) {
	return r.submitDailyReviewTx(ctx, tx, req, sessionID, time.Time{})
}

// SubmitDailyReviewAtTx keeps an owned daily action on its single server-captured
// timestamp. operationNow comes from the locked daily service, never from HTTP
// request fields. Membership and review eligibility are still checked here.
func (r *PostgreSQLRepository) SubmitDailyReviewAtTx(ctx context.Context, tx *sql.Tx, req SubmitReviewRequest, sessionID uuid.UUID, operationNow time.Time) (*ReviewAttempt, error) {
	if operationNow.IsZero() {
		return nil, errors.New("server operation timestamp required")
	}
	return r.submitDailyReviewTx(ctx, tx, req, sessionID, operationNow.UTC())
}

func (r *PostgreSQLRepository) submitDailyReviewTx(ctx context.Context, tx *sql.Tx, req SubmitReviewRequest, sessionID uuid.UUID, operationNow time.Time) (*ReviewAttempt, error) {
	if tx == nil || sessionID == uuid.Nil {
		return nil, ErrUserWordNotFound
	}
	var member bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM daily_lesson_sessions s CROSS JOIN LATERAL jsonb_array_elements(s.state->'Words') w
 WHERE s.id=$1 AND s.user_id=$2 AND s.status='in_progress' AND w->>'MeaningID'=$3 AND w->>'UserWordID'=$4)`, sessionID, req.UserID, req.MeaningID.String(), req.UserWordID.String()).Scan(&member)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrUserWordNotFound
	}
	req.verifiedSessionID = sessionID
	return r.submitReviewTx(ctx, tx, req, operationNow)
}

// SubmitReviewTx applies the exact same scheduling, receipt and reward transaction
// to a server-graded daily lesson. The caller must hold the active user lock and
// commit this transaction together with its own durable session/action state.
func (r *PostgreSQLRepository) SubmitReviewTx(ctx context.Context, tx *sql.Tx, req SubmitReviewRequest) (*ReviewAttempt, error) {
	return r.submitReviewTx(ctx, tx, req, time.Time{})
}

func (r *PostgreSQLRepository) submitReviewTx(ctx context.Context, tx *sql.Tx, req SubmitReviewRequest, operationNow time.Time) (*ReviewAttempt, error) {
	var err error
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	var priorStep, totalCount, correctCount, consecutiveCorrect, consecutiveIncorrect int
	var meaningID uuid.UUID
	var status string
	var nextDue sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT review_step, meaning_id, total_review_count, correct_review_count,
		        consecutive_correct_count, consecutive_incorrect_count, status, next_review_at
		 FROM user_words
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 FOR UPDATE`,
		req.UserWordID, req.UserID,
	).Scan(&priorStep, &meaningID, &totalCount, &correctCount, &consecutiveCorrect, &consecutiveIncorrect, &status, &nextDue)
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
	now := operationNow
	if now.IsZero() {
		now = r.clock.Now().UTC()
	}
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
			return existing, nil
		}
		return nil, ErrIdempotencyConflict
	}
	if activeKey {
		return nil, ErrReviewAttemptNotFound
	}

	if req.ExpectedReviewCount != nil && *req.ExpectedReviewCount != totalCount {
		return nil, ErrNotDue
	}
	if !isDueStatus(status) || (nextDue.Valid && nextDue.Time.After(now)) {
		return nil, ErrNotDue
	}
	// A verified answer resolves this meaning for the learner's local day even
	// if a legacy restore reset its schedule. Exact replays returned above;
	// the private daily path retains its own receipt dedupe.
	var unavailable bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_word_knowledge WHERE user_id=$1 AND meaning_id=$2 AND self_reported_known) OR `+revieweligibility.ReviewedToday("$1", "(SELECT word_id FROM word_meanings WHERE id=$2)", "$4")+` OR ($3::boolean AND EXISTS(SELECT 1 FROM daily_review_successes d WHERE d.user_id=$1 AND d.meaning_id=$2 AND d.local_date=($4 AT TIME ZONE COALESCE((SELECT timezone FROM user_settings WHERE user_id=$1),'UTC'))::date))`, req.UserID, req.MeaningID, req.verifiedSessionID == uuid.Nil, now).Scan(&unavailable); err != nil {
		return nil, err
	}
	if unavailable {
		return nil, ErrNotDue
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
	if !operationNow.IsZero() {
		scheduledAt = operationNow
	} else if scheduledAt.IsZero() {
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
	// A skipped review has no scheduling rating. The schema permits NULL, not
	// an empty string, so keep the absent value nullable in both persisted
	// review state and immutable attempt history.
	lastRating := sql.NullString{}
	if sched.LastRating != "" {
		lastRating = sql.NullString{String: sched.LastRating, Valid: true}
	}
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
	resolved, err := r.gamification.GetSettingsTx(ctx, tx, req.UserID, "")
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
	credited := false
	if correct && !req.WasHintUsed && req.verifiedSessionID != uuid.Nil {
		result, e := tx.ExecContext(ctx, `INSERT INTO daily_review_successes(user_id,meaning_id,local_date,review_attempt_id,created_at,daily_session_id) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(user_id,meaning_id,local_date) DO NOTHING`, req.UserID, req.MeaningID, snap.LocalDate, attemptID, now, req.verifiedSessionID)
		if e != nil {
			return e
		}
		n, e := result.RowsAffected()
		if e != nil {
			return e
		}
		credited = n == 1
	}
	newReviewsCompleted, err := r.missions.IncrementReviewsCompleted(
		ctx, tx, req.UserID, snap.LocalDate, resolved.Timezone,
		resolved.DailyReviewTarget, correct, skipped, credited,
	)
	if err != nil {
		return fmt.Errorf("increment reviews completed: %w", err)
	}

	// Skipped reviews do not award a rating-tiered point (DOC-06 §11
	// prices only Again/Hard/Good/Easy, all of which require a rating)
	// but the activity counter above already recorded the attempt.
	// A server-graded daily lesson awards rating XP only for a newly credited
	// correct meaning. Repeated wrong retries are learning attempts, not XP.
	// Standalone legacy review keeps its original rating policy.
	awardRating := !skipped && (req.verifiedSessionID == uuid.Nil || (correct && credited))
	if awardRating {
		balance, err := r.gamification.CurrentBalanceTx(ctx, tx, req.UserID)
		if err != nil {
			return fmt.Errorf("get latest point balance: %w", err)
		}
		newBalance, _, _, err := r.gamification.GrantPoint(
			ctx, tx, req.UserID,
			rewardKind, &attemptID,
			gamification.ReviewAttemptRatedKey(attemptID.String()),
			balance, now, nil,
		)
		if err != nil {
			return fmt.Errorf("grant review point: %w", err)
		}
		if err := r.missions.RecordConfidencePointChange(
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
			balance, err := r.gamification.CurrentBalanceTx(ctx, tx, req.UserID)
			if err != nil {
				return fmt.Errorf("get latest point balance: %w", err)
			}
			localDateKey := snap.LocalDate.Format("2006-01-02")
			if _, _, _, err := r.gamification.GrantPoint(
				ctx, tx, req.UserID,
				gamification.RewardKindDailyMissionDone, nil,
				gamification.DailyMissionCompletedKey(req.UserID.String(), localDateKey),
				balance, now, nil,
			); err != nil {
				return fmt.Errorf("grant daily mission point: %w", err)
			}
			if err := r.missions.RecordConfidencePointChange(
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
	reconciliation, err := r.gamification.ReconcileAndAdvance(
		ctx, tx, req.UserID, now, resolved.Timezone,
		snaps, missionCompletedNow,
	)
	if err != nil {
		return fmt.Errorf("reconcile streak: %w", err)
	}
	if reconciliation.GraceDayUsed != nil && reconciliation.GraceDayUsedID != nil && reconciliation.YesterdayProtectedLocalDate != nil {
		protected, err := r.missions.MarkSnapshotProtected(
			ctx, tx, req.UserID, *reconciliation.YesterdayProtectedLocalDate, *reconciliation.GraceDayUsedID, resolved.Timezone,
			snap.ReviewTarget,
		)
		if err != nil {
			return fmt.Errorf("mark grace-protected snapshot: %w", err)
		}
		if !protected {
			return errors.New("mark grace-protected snapshot: missed snapshot was not updated")
		}
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
	// PostgreSQL timestamptz persists microsecond precision while the API
	// accepts RFC3339 nanoseconds. Compare at the database's precision so an
	// exact retry of a persisted request remains a replay.
	if !postgresTimestampEqual(a.AnsweredAt, req.AnsweredAt) {
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

// postgresTimestampEqual reproduces PostgreSQL's microsecond timestamp
// parsing and rounding. PostgreSQL parses the fractional seconds through a
// float before rounding half-to-even, so ideal integer nanosecond rounding and
// time.Time.Round do not match every accepted RFC3339 value.
func postgresTimestampEqual(a, b time.Time) bool {
	return postgresTimestamp(a).Equal(postgresTimestamp(b))
}

func postgresTimestamp(t time.Time) time.Time {
	t = t.UTC()
	microseconds := math.RoundToEven(float64(t.Nanosecond()) / float64(time.Second) * 1_000_000)
	return t.Truncate(time.Second).Add(time.Duration(microseconds) * time.Microsecond)
}
