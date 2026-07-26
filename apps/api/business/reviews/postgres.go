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
type PostgreSQLRepository struct {
	db             *sql.DB
	clock          clock.Clock
	gamification   *gamification.Service
	missionsRepo   *missions.Repository
	missionsModule *missions.Service
}

// NewPostgreSQLRepository creates a repository backed by db.
func NewPostgreSQLRepository(db *sql.DB, c clock.Clock) *PostgreSQLRepository {
	if c == nil {
		c = clock.Real{}
	}
	return &PostgreSQLRepository{db: db, clock: c}
}

// WithGameification wires in gamification and missions services for P4 reward/streak tracking.
func (r *PostgreSQLRepository) WithGameification(
	gam *gamification.Service,
	missionsRepo *missions.Repository,
	missionsModule *missions.Service,
) *PostgreSQLRepository {
	r.gamification = gam
	r.missionsRepo = missionsRepo
	r.missionsModule = missionsModule
	return r
}

func (r *PostgreSQLRepository) ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
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
		`SELECT
		   uw.id, uw.meaning_id, cw.id, cw.text, cw.normalized_text,
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
		   AND ($2::timestamptz IS NULL OR
		        COALESCE(uw.next_review_at, '-infinity'::timestamptz) > $2 OR
		        (COALESCE(uw.next_review_at, '-infinity'::timestamptz) = $2 AND uw.id > $3))
		 ORDER BY COALESCE(uw.next_review_at, '-infinity'::timestamptz) ASC, uw.id ASC
		 LIMIT $4`,
		req.UserID, cursorNextReviewAt, cursorID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list due words: %w", err)
	}
	defer rows.Close()

	var items []DueWord
	var totalCount int
	var last DueWord
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
		last = d
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due words rows: %w", err)
	}

	resp := &ListDueWordsResponse{Items: items, TotalCount: totalCount}
	if len(items) == limit {
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

	prior := ReviewState{
		ReviewStep: priorStep,
		Counters: ReviewCounters{
			TotalReviewCount:          totalCount,
			CorrectReviewCount:        correctCount,
			ConsecutiveCorrectCount:   consecutiveCorrect,
			ConsecutiveIncorrectCount: consecutiveIncorrect,
		},
	}
	sched, err := ApplyReview(prior, ApplyReviewRequest{Result: req.Result, Rating: req.Rating}, req.AnsweredAt)
	if err != nil {
		return nil, err
	}

	attemptID := uuid.New()
	now := r.clock.Now().UTC()
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

	if r.gamification != nil && r.missionsRepo != nil {
		if err := r.wireGameification(ctx, tx, req, now, sched); err != nil {
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

func (r *PostgreSQLRepository) wireGameification(
	ctx context.Context,
	tx *sql.Tx,
	req SubmitReviewRequest,
	now time.Time,
	sched ApplyReviewResult,
) error {
	resolved, err := r.gamification.GetSettings(ctx, req.UserID, "")
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}

	today, err := gamification.LocalDate(now, resolved.Timezone)
	if err != nil {
		return fmt.Errorf("compute local date: %w", err)
	}

	snap, err := r.missionsRepo.CreateDailyMissionSnapshot(
		ctx, tx, req.UserID, today, resolved.Timezone,
		resolved.DailyReviewTarget, gamification.MissionPolicyVersion,
	)
	if err != nil {
		return fmt.Errorf("create mission snapshot: %w", err)
	}

	isCorrect := req.Result == ResultCorrect && req.Rating != RatingAgain
	isSkipped := req.Result == ResultSkipped

	reviewsCompleted, err := r.missionsRepo.IncrementReviewsCompleted(
		ctx, tx, req.UserID, today, resolved.Timezone,
		resolved.DailyReviewTarget, isCorrect, isSkipped,
	)
	if err != nil {
		return fmt.Errorf("increment reviews completed: %w", err)
	}

	rewardKind := ratingToRewardKind(req.Rating)
	currentBalance, err := r.gamification.CurrentBalance(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("get current balance: %w", err)
	}

	newBalance, _, err := r.gamification.GrantPoint(
		ctx, tx, req.UserID, rewardKind,
		&req.UserWordID, gamification.ReviewAttemptRatedKey(req.UserWordID.String()),
		currentBalance, now, nil,
	)
	if err != nil {
		return fmt.Errorf("grant review point: %w", err)
	}

	pointsReward := rewardKindToAmount(rewardKind)
	if err := r.missionsRepo.IncrementConfidencePointsEarned(
		ctx, tx, req.UserID, today, resolved.Timezone, pointsReward,
	); err != nil {
		return fmt.Errorf("increment points earned: %w", err)
	}

	graceBalance, err := r.gamification.CurrentGraceBalance(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("get grace balance: %w", err)
	}

	recentSnapshots, err := r.missionsRepo.ListRecentSnapshots(ctx, req.UserID, 14)
	if err != nil {
		return fmt.Errorf("list recent snapshots: %w", err)
	}

	streakSnapshots := make([]gamification.StreakSnapshot, 0, len(recentSnapshots))
	for _, s := range recentSnapshots {
		streakSnapshots = append(streakSnapshots, gamification.StreakSnapshot{
			LocalDate:    s.LocalDate,
			Status:       s.Status,
			CompletedAt:  s.CompletedAt,
			GraceApplied: s.GraceApplied,
			GraceDayID:   s.GraceDayID,
		})
	}

	if reviewsCompleted >= snap.ReviewTarget {
		completed, err := r.missionsRepo.MarkSnapshotCompleted(ctx, tx, req.UserID, today, now)
		if err != nil {
			return fmt.Errorf("mark snapshot completed: %w", err)
		}
		if completed {
			_, _, err := r.gamification.GrantPoint(
				ctx, tx, req.UserID, gamification.RewardKindDailyMissionDone,
				nil, gamification.DailyMissionCompletedKey(req.UserID.String(), today.Format("2006-01-02")),
				newBalance, now, nil,
			)
			if err != nil {
				return fmt.Errorf("grant mission completion point: %w", err)
			}

			if err := r.missionsRepo.IncrementConfidencePointsEarned(
				ctx, tx, req.UserID, today, resolved.Timezone, gamification.RewardDailyMissionDone,
			); err != nil {
				return fmt.Errorf("increment mission completion points: %w", err)
			}
		}
	}

	rec, err := r.gamification.ReconcileAndAdvance(
		ctx, tx, req.UserID, now, resolved.Timezone, streakSnapshots, graceBalance, reviewsCompleted >= snap.ReviewTarget,
	)
	if err != nil {
		return fmt.Errorf("reconcile streak: %w", err)
	}

	if rec.YesterdayProtectedLocalDate != nil && rec.YesterdaySnapshotID != nil && rec.GraceDayUsed != nil && rec.GraceDayUsed.SourceID != nil {
		_, err := r.missionsRepo.MarkSnapshotProtected(ctx, tx, uuid.MustParse(*rec.YesterdaySnapshotID), *rec.GraceDayUsed.SourceID)
		if err != nil {
			return fmt.Errorf("mark snapshot protected: %w", err)
		}
	}

	return nil
}

func ratingToRewardKind(rating string) gamification.RewardKind {
	switch rating {
	case RatingAgain:
		return gamification.RewardKindReviewAgain
	case RatingHard:
		return gamification.RewardKindReviewHard
	case RatingGood:
		return gamification.RewardKindReviewGood
	case RatingEasy:
		return gamification.RewardKindReviewEasy
	default:
		return gamification.RewardKindReviewGood
	}
}

func rewardKindToAmount(kind gamification.RewardKind) int {
	switch kind {
	case gamification.RewardKindReviewAgain:
		return gamification.RewardReviewAgain
	case gamification.RewardKindReviewHard:
		return gamification.RewardReviewHard
	case gamification.RewardKindReviewGood:
		return gamification.RewardReviewGood
	case gamification.RewardKindReviewEasy:
		return gamification.RewardReviewEasy
	default:
		return 0
	}
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
