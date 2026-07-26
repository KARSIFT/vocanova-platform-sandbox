package missions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists mission and activity state. The transaction-scoped
// helpers accept an existing *sql.Tx from the caller's existing P1/P2/P3
// transaction (DOC-06 §3 — missions/gamification never open their own
// transaction).
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by db.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetDailyMissionSnapshot fetches the daily_mission_snapshots row for
// (userID, localDate). Returns (nil, nil) if no row exists.
func (r *Repository) GetDailyMissionSnapshot(ctx context.Context, userID uuid.UUID, localDate time.Time) (*DailyMissionSnapshot, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, local_date, timezone, review_target, reviews_completed,
		        new_word_target, new_words_completed, sentence_practice_target,
		        sentence_practices_completed, policy_version, status, completed_at,
		        grace_applied, grace_day_id
		 FROM daily_mission_snapshots
		 WHERE user_id = $1 AND local_date = $2`,
		userID, localDate,
	)
	return scanSnapshot(row)
}

// GetDailyActivitySummary fetches the daily_activity_summaries row for
// (userID, localDate). Returns (nil, nil) if no row exists.
func (r *Repository) GetDailyActivitySummary(ctx context.Context, userID uuid.UUID, localDate time.Time) (*DailyActivitySummary, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, local_date, timezone, reviews_attempted, reviews_correct,
		        reviews_skipped, words_discovered, words_added, sentences_submitted,
		        ai_feedback_received, confidence_points_earned, confidence_points_spent
		 FROM daily_activity_summaries
		 WHERE user_id = $1 AND local_date = $2`,
		userID, localDate,
	)
	return scanActivity(row)
}

// ListRecentSnapshots fetches the last `days` daily_mission_snapshots for
// userID ending at today. Used by streak reconciliation.
func (r *Repository) ListRecentSnapshots(ctx context.Context, userID uuid.UUID, days int) ([]DailyMissionSnapshot, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, local_date, timezone, review_target, reviews_completed,
		        new_word_target, new_words_completed, sentence_practice_target,
		        sentence_practices_completed, policy_version, status, completed_at,
		        grace_applied, grace_day_id
		 FROM daily_mission_snapshots
		 WHERE user_id = $1
		 ORDER BY local_date DESC
		 LIMIT $2`,
		userID, days,
	)
	if err != nil {
		return nil, fmt.Errorf("list recent snapshots: %w", err)
	}
	defer rows.Close()
	var out []DailyMissionSnapshot
	for rows.Next() {
		s, err := scanSnapshotRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list recent snapshots rows: %w", err)
	}
	return out, nil
}

// ListRecentCompletionHistory fetches the last `days` local-date + status
// pairs for userID, used by GET /api/v1/progress's bounded 7-day history.
func (r *Repository) ListRecentCompletionHistory(ctx context.Context, userID uuid.UUID, days int) ([]CompletionDay, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT local_date, status
		 FROM daily_mission_snapshots
		 WHERE user_id = $1
		 ORDER BY local_date DESC
		 LIMIT $2`,
		userID, days,
	)
	if err != nil {
		return nil, fmt.Errorf("list recent completion history: %w", err)
	}
	defer rows.Close()
	var out []CompletionDay
	for rows.Next() {
		var d CompletionDay
		if err := rows.Scan(&d.LocalDate, &d.Status); err != nil {
			return nil, fmt.Errorf("scan completion day: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list recent completion history rows: %w", err)
	}
	return out, nil
}

// CompletionDay is a (localDate, status) pair from daily_mission_snapshots.
type CompletionDay struct {
	LocalDate time.Time
	Status    string
}

// CreateDailyMissionSnapshot inserts one daily_mission_snapshots row inside
// tx. If a row already exists for (userID, localDate), this is a no-op and
// the existing row is returned. The unique index handles dedup.
func (r *Repository) CreateDailyMissionSnapshot(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	reviewTarget int,
	policyVersion string,
) (*DailyMissionSnapshot, error) {
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	if reviewTarget < 5 || reviewTarget > 100 {
		return nil, fmt.Errorf("review target %d out of range", reviewTarget)
	}
	if timezone == "" {
		return nil, errors.New("timezone required")
	}
	if policyVersion == "" {
		return nil, errors.New("policy version required")
	}
	row := tx.QueryRowContext(ctx,
		`INSERT INTO daily_mission_snapshots (
			id, user_id, local_date, timezone, review_target, reviews_completed,
			policy_version, status, grace_applied
		) VALUES (
			$1, $2, $3, $4, $5, 0,
			$6, 'open', false
		)
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET timezone = EXCLUDED.timezone,
		      review_target = EXCLUDED.review_target,
		      updated_at = NOW()
		RETURNING id, user_id, local_date, timezone, review_target, reviews_completed,
		          new_word_target, new_words_completed, sentence_practice_target,
		          sentence_practices_completed, policy_version, status, completed_at,
		          grace_applied, grace_day_id`,
		uuid.New(), userID, localDate, timezone, reviewTarget, policyVersion,
	)
	return scanSnapshot(row)
}

// IncrementReviewsCompleted increments daily_mission_snapshots.reviews_completed
// and daily_activity_summaries.reviews_attempted (and _correct/_skipped if
// the result is correct/skipped) atomically inside tx. The reviews_completed
// increment is capped at review_target by the DB check; the caller passes
// the new value, and the SQL handles the cap via LEAST. Returns the new
// reviews_completed value.
func (r *Repository) IncrementReviewsCompleted(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	reviewTarget int,
	correct bool,
	skipped bool,
) (int, error) {
	if tx == nil {
		return 0, errors.New("transaction required")
	}
	var newCount int
	err := tx.QueryRowContext(ctx,
		`UPDATE daily_mission_snapshots
		 SET reviews_completed = LEAST(reviews_completed + 1, review_target),
		     updated_at = NOW()
		 WHERE user_id = $1 AND local_date = $2
		 RETURNING reviews_completed`,
		userID, localDate,
	).Scan(&newCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Snapshot not yet created; the caller should have
			// ensured it before invoking this. Surface a typed
			// error so the caller can recover.
			return 0, ErrSnapshotNotFound
		}
		return 0, fmt.Errorf("increment reviews completed: %w", err)
	}
	// Activity summary update.
	reviewsAttempted := 1
	reviewsCorrect := 0
	reviewsSkipped := 0
	if correct {
		reviewsCorrect = 1
	}
	if skipped {
		reviewsSkipped = 1
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO daily_activity_summaries (
			id, user_id, local_date, timezone,
			reviews_attempted, reviews_correct, reviews_skipped
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET reviews_attempted = daily_activity_summaries.reviews_attempted + EXCLUDED.reviews_attempted,
		      reviews_correct = daily_activity_summaries.reviews_correct + EXCLUDED.reviews_correct,
		      reviews_skipped = daily_activity_summaries.reviews_skipped + EXCLUDED.reviews_skipped,
		      updated_at = NOW()`,
		uuid.New(), userID, localDate, timezone,
		reviewsAttempted, reviewsCorrect, reviewsSkipped,
	); err != nil {
		return 0, fmt.Errorf("upsert activity summary reviews: %w", err)
	}
	_ = reviewTarget
	return newCount, nil
}

// IncrementWordsAdded increments daily_activity_summaries.words_added by 1
// and, if the optional new-word mission goal is active, also increments
// daily_mission_snapshots.new_words_completed. The caller controls which
// behavior is invoked by passing includeNewWordGoal.
func (r *Repository) IncrementWordsAdded(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	includeNewWordGoal bool,
) error {
	if tx == nil {
		return errors.New("transaction required")
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO daily_activity_summaries (
			id, user_id, local_date, timezone, words_added
		) VALUES (
			$1, $2, $3, $4, 1
		)
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET words_added = daily_activity_summaries.words_added + 1,
		      updated_at = NOW()`,
		uuid.New(), userID, localDate, timezone,
	); err != nil {
		return fmt.Errorf("upsert activity summary words_added: %w", err)
	}
	if includeNewWordGoal {
		if _, err := tx.ExecContext(ctx,
			`UPDATE daily_mission_snapshots
			 SET new_words_completed = LEAST(COALESCE(new_words_completed, 0) + 1,
			                                COALESCE(new_word_target, 1)),
			     updated_at = NOW()
			 WHERE user_id = $1 AND local_date = $2`,
			userID, localDate,
		); err != nil {
			return fmt.Errorf("increment new_words_completed: %w", err)
		}
	}
	return nil
}

// IncrementSentenceSubmitted increments daily_activity_summaries
// .sentences_submitted and, if the optional sentence-practice goal is
// active, daily_mission_snapshots.sentence_practices_completed.
func (r *Repository) IncrementSentenceSubmitted(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	includeSentenceGoal bool,
) error {
	if tx == nil {
		return errors.New("transaction required")
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO daily_activity_summaries (
			id, user_id, local_date, timezone, sentences_submitted
		) VALUES (
			$1, $2, $3, $4, 1
		)
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET sentences_submitted = daily_activity_summaries.sentences_submitted + 1,
		      updated_at = NOW()`,
		uuid.New(), userID, localDate, timezone,
	); err != nil {
		return fmt.Errorf("upsert activity summary sentences_submitted: %w", err)
	}
	if includeSentenceGoal {
		if _, err := tx.ExecContext(ctx,
			`UPDATE daily_mission_snapshots
			 SET sentence_practices_completed = LEAST(COALESCE(sentence_practices_completed, 0) + 1,
			                                         COALESCE(sentence_practice_target, 1)),
			     updated_at = NOW()
			 WHERE user_id = $1 AND local_date = $2`,
			userID, localDate,
		); err != nil {
			return fmt.Errorf("increment sentence_practices_completed: %w", err)
		}
	}
	return nil
}

// IncrementAIFeedbackReceived increments daily_activity_summaries
// .ai_feedback_received.
func (r *Repository) IncrementAIFeedbackReceived(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
) error {
	if tx == nil {
		return errors.New("transaction required")
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO daily_activity_summaries (
			id, user_id, local_date, timezone, ai_feedback_received
		) VALUES (
			$1, $2, $3, $4, 1
		)
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET ai_feedback_received = daily_activity_summaries.ai_feedback_received + 1,
		      updated_at = NOW()`,
		uuid.New(), userID, localDate, timezone,
	); err != nil {
		return fmt.Errorf("upsert activity summary ai_feedback_received: %w", err)
	}
	return nil
}

// IncrementConfidencePointsEarned increments daily_activity_summaries
// .confidence_points_earned by amount. amount may be negative for spend
// events.
func (r *Repository) IncrementConfidencePointsEarned(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	amount int,
) error {
	if tx == nil {
		return errors.New("transaction required")
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO daily_activity_summaries (
			id, user_id, local_date, timezone, confidence_points_earned
		) VALUES (
			$1, $2, $3, $4, $5
		)
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET confidence_points_earned = daily_activity_summaries.confidence_points_earned + EXCLUDED.confidence_points_earned,
		      updated_at = NOW()`,
		uuid.New(), userID, localDate, timezone, amount,
	); err != nil {
		return fmt.Errorf("upsert activity summary points earned: %w", err)
	}
	return nil
}

// MarkSnapshotCompleted transitions the snapshot to status='completed' with
// completed_at = now, and returns whether the transition actually happened.
// The conditional WHERE status='open' makes this idempotent: a retried call
// will not re-set status or completed_at.
func (r *Repository) MarkSnapshotCompleted(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	now time.Time,
) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction required")
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE daily_mission_snapshots
		 SET status = 'completed', completed_at = $3, updated_at = NOW()
		 WHERE user_id = $1 AND local_date = $2 AND status = 'open'`,
		userID, localDate, now,
	)
	if err != nil {
		return false, fmt.Errorf("mark snapshot completed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return rows > 0, nil
}

// MarkSnapshotProtected transitions yesterday's snapshot to status='protected'
// with grace_day_id, and marks grace_applied=true. Conditional on status being
// 'missed' so it can't double-apply a grace day.
func (r *Repository) MarkSnapshotProtected(
	ctx context.Context,
	tx *sql.Tx,
	snapshotID uuid.UUID,
	graceDayID uuid.UUID,
) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction required")
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE daily_mission_snapshots
		 SET status = 'protected', grace_applied = true, grace_day_id = $2, updated_at = NOW()
		 WHERE id = $1 AND status = 'missed'`,
		snapshotID, graceDayID,
	)
	if err != nil {
		return false, fmt.Errorf("mark snapshot protected: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return rows > 0, nil
}

// MarkSnapshotMissed is called lazily when streak reconciliation detects a
// missed day during a read or write. status='missed', grace_applied remains
// false until a grace day is later applied.
func (r *Repository) MarkSnapshotMissed(
	ctx context.Context,
	tx *sql.Tx,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
) error {
	if tx == nil {
		return errors.New("transaction required")
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO daily_mission_snapshots (
			id, user_id, local_date, timezone, review_target, policy_version, status
		) VALUES (
			$1, $2, $3, $4, 20, 'p4-mission-policy-v1', 'missed'
		)
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET status = CASE WHEN daily_mission_snapshots.status = 'open' THEN 'missed' ELSE daily_mission_snapshots.status END,
		      updated_at = NOW()`,
		uuid.New(), userID, localDate, timezone,
	); err != nil {
		return fmt.Errorf("mark snapshot missed: %w", err)
	}
	return nil
}

// ErrSnapshotNotFound is returned by mutation helpers when a snapshot for
// (userID, localDate) does not exist. The caller is expected to lazily
// create the snapshot before re-issuing the mutation.
var ErrSnapshotNotFound = errors.New("daily mission snapshot not found")

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSnapshot(row rowScanner) (*DailyMissionSnapshot, error) {
	var s DailyMissionSnapshot
	var id, userID uuid.UUID
	var newWordTarget, newWordsCompleted, sentenceTarget, sentenceCompleted sql.NullInt32
	var completedAt sql.NullTime
	var graceDayID *uuid.UUID
	if err := row.Scan(
		&id, &userID, &s.LocalDate, &s.Timezone, &s.ReviewTarget, &s.ReviewsCompleted,
		&newWordTarget, &newWordsCompleted, &sentenceTarget, &sentenceCompleted,
		&s.PolicyVersion, &s.Status, &completedAt, &s.GraceApplied, &graceDayID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan snapshot: %w", err)
	}
	s.ID = id.String()
	s.UserID = userID.String()
	if newWordTarget.Valid {
		v := int(newWordTarget.Int32)
		s.NewWordTarget = &v
	}
	if newWordsCompleted.Valid {
		v := int(newWordsCompleted.Int32)
		s.NewWordsCompleted = &v
	}
	if sentenceTarget.Valid {
		v := int(sentenceTarget.Int32)
		s.SentencePracticeTarget = &v
	}
	if sentenceCompleted.Valid {
		v := int(sentenceCompleted.Int32)
		s.SentencePracticesCompleted = &v
	}
	if completedAt.Valid {
		t := completedAt.Time
		s.CompletedAt = &t
	}
	if graceDayID != nil {
		v := graceDayID.String()
		s.GraceDayID = &v
	}
	return &s, nil
}

func scanSnapshotRows(rows *sql.Rows) (*DailyMissionSnapshot, error) {
	return scanSnapshot(rows)
}

func scanActivity(row rowScanner) (*DailyActivitySummary, error) {
	var a DailyActivitySummary
	var id, userID uuid.UUID
	if err := row.Scan(
		&id, &userID, &a.LocalDate, &a.Timezone, &a.ReviewsAttempted, &a.ReviewsCorrect,
		&a.ReviewsSkipped, &a.WordsDiscovered, &a.WordsAdded, &a.SentencesSubmitted,
		&a.AIFeedbackReceived, &a.ConfidencePointsEarned, &a.ConfidencePointsSpent,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan activity: %w", err)
	}
	a.ID = id.String()
	a.UserID = userID.String()
	return &a, nil
}
