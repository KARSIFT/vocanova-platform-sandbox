package aifeedback

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// PostgreSQLRepository implements Repository against the VOC-028 P3 schema.
type PostgreSQLRepository struct {
	db    *sql.DB
	clock clock.Clock
}

// NewPostgreSQLRepository creates a repository backed by db.
func NewPostgreSQLRepository(db *sql.DB, c clock.Clock) *PostgreSQLRepository {
	if c == nil {
		c = clock.Real{}
	}
	return &PostgreSQLRepository{db: db, clock: c}
}

func (r *PostgreSQLRepository) LoadTarget(ctx context.Context, req LoadTargetRequest) (*Target, error) {
	switch req.Source {
	case SourceWordDetail:
		return r.loadTargetFromUserWord(ctx, req.UserID, req.AttemptID)
	case SourceReview:
		return r.loadTargetFromReviewAttempt(ctx, req.UserID, req.AttemptID)
	case SourceDailyMission, SourceFreePractice:
		return nil, ErrTargetNotFound
	default:
		return nil, ErrTargetNotFound
	}
}

func (r *PostgreSQLRepository) loadTargetFromUserWord(ctx context.Context, userID, userWordID uuid.UUID) (*Target, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT cw.id, cw.text, cw.normalized_text, cw.word_type, cw.difficulty_level,
		        wm.id, wm.part_of_speech, wm.short_definition, uw.id
		 FROM user_words uw
		 JOIN word_meanings wm ON wm.id = uw.meaning_id
		 JOIN canonical_words cw ON cw.id = wm.word_id
		 WHERE uw.id = $1 AND uw.user_id = $2 AND uw.deleted_at IS NULL`,
		userWordID, userID,
	)
	return r.scanTarget(row, userWordID, nil)
}

func (r *PostgreSQLRepository) loadTargetFromReviewAttempt(ctx context.Context, userID, reviewAttemptID uuid.UUID) (*Target, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT cw.id, cw.text, cw.normalized_text, cw.word_type, cw.difficulty_level,
		        wm.id, wm.part_of_speech, wm.short_definition, ra.user_word_id
		 FROM review_attempts ra
		 JOIN user_words uw ON uw.id = ra.user_word_id
		 JOIN word_meanings wm ON wm.id = ra.meaning_id
		 JOIN canonical_words cw ON cw.id = wm.word_id
		 WHERE ra.id = $1 AND ra.user_id = $2`,
		reviewAttemptID, userID,
	)
	return r.scanTarget(row, uuid.Nil, &reviewAttemptID)
}

func (r *PostgreSQLRepository) scanTarget(row *sql.Row, userWordID uuid.UUID, reviewAttemptID *uuid.UUID) (*Target, error) {
	var t Target
	var wordID, meaningID, loadedUserWordID uuid.UUID
	var wordText, normalizedText, wordType, partOfSpeech, shortDefinition string
	var difficultyLevel sql.NullString

	err := row.Scan(
		&wordID, &wordText, &normalizedText, &wordType, &difficultyLevel,
		&meaningID, &partOfSpeech, &shortDefinition, &loadedUserWordID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan target: %w", err)
	}

	t = Target{
		WordID:          wordID,
		MeaningID:       meaningID,
		UserWordID:      loadedUserWordID,
		ReviewAttemptID: reviewAttemptID,
		WordText:        wordText,
		NormalizedWord:  normalizedText,
		WordType:        wordType,
		PartOfSpeech:    partOfSpeech,
		ShortDefinition: shortDefinition,
		LearnerLevel:    learnerLevel(difficultyLevel.String),
		AcceptedForms:   BuildAcceptedForms(normalizedText, wordType, partOfSpeech),
	}
	if userWordID != uuid.Nil {
		t.UserWordID = userWordID
	}
	return &t, nil
}

func learnerLevel(level string) string {
	switch level {
	case "a1", "a2", "b1", "b2", "c1":
		return level
	}
	return "a2"
}

func (r *PostgreSQLRepository) GetFeedbackAttemptByRequestHash(ctx context.Context, requestHash string) (*StoredFeedbackAttempt, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, learner_sentence_id, status, provider, model, prompt_version, request_hash,
		        feedback_json, feedback_text, error_code, error_message,
		        EXISTS (SELECT 1 FROM ai_feedback_quality_review_reports r WHERE r.ai_feedback_attempt_id = ai_feedback_attempts.id)
		 FROM ai_feedback_attempts
		 WHERE request_hash = $1
		 ORDER BY CASE status WHEN 'succeeded' THEN 0 WHEN 'pending' THEN 1 ELSE 2 END,
		          created_at DESC`,
		requestHash,
	)
	return r.scanStoredAttempt(row)
}

func (r *PostgreSQLRepository) scanStoredAttempt(row *sql.Row) (*StoredFeedbackAttempt, error) {
	var a StoredFeedbackAttempt
	var feedbackJSON []byte
	var feedbackText, errorCode, errorMessage sql.NullString

	err := row.Scan(
		&a.ID, &a.LearnerSentenceID, &a.Status, &a.Provider, &a.Model, &a.PromptVersion, &a.RequestHash,
		&feedbackJSON, &feedbackText, &errorCode, &errorMessage, &a.Reported,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan stored attempt: %w", err)
	}

	if len(feedbackJSON) > 0 {
		a.FeedbackJSON = make(map[string]any)
		if err := json.Unmarshal(feedbackJSON, &a.FeedbackJSON); err != nil {
			return nil, fmt.Errorf("unmarshal feedback json: %w", err)
		}
	}
	if feedbackText.Valid {
		a.FeedbackText = feedbackText.String
	}
	if errorCode.Valid {
		a.ErrorCode = errorCode.String
	}
	if errorMessage.Valid {
		a.ErrorMessage = errorMessage.String
	}
	return &a, nil
}

func (r *PostgreSQLRepository) CreatePendingAttempt(ctx context.Context, req SubmitSentenceFeedbackRequest, target *Target, normalized string, requestHash string, provider string, model string, now time.Time) (*PendingAttempt, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	sentenceID := uuid.New()
	attemptID := uuid.New()

	var meaningID, userWordID interface{}
	if target.MeaningID != uuid.Nil {
		meaningID = target.MeaningID
	} else {
		meaningID = nil
	}
	if target.UserWordID != uuid.Nil {
		userWordID = target.UserWordID
	} else {
		userWordID = nil
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO learner_sentences (
			id, user_id, meaning_id, user_word_id, sentence_text, normalized_sentence_text,
			source, status, submitted_at, deleted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL, $10, $10)`,
		sentenceID, req.UserID, meaningID, userWordID, req.SentenceText, normalized,
		req.Source, SentenceStatusSubmitted, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert learner sentence: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO ai_feedback_attempts (
			id, learner_sentence_id, status, provider, model, prompt_version, request_hash,
			feedback_json, feedback_text, error_code, error_message,
			started_at, completed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, NULL, NULL, NULL, NULL, NULL, $8, $8)`,
		attemptID, sentenceID, AttemptStatusPending, provider, model, PromptVersionSentenceFeedbackV1, requestHash, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert ai feedback attempt: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pending attempt: %w", err)
	}

	return &PendingAttempt{SentenceID: sentenceID, AttemptID: attemptID}, nil
}

// CreateRetryAttempt appends an immutable retry generation for the same
// learner sentence after a failed provider attempt. The partial unique index on
// request_hash prevents more than one pending or successful generation.
func (r *PostgreSQLRepository) CreateRetryAttempt(ctx context.Context, failed *StoredFeedbackAttempt, provider string, model string, now time.Time) (*PendingAttempt, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin retry tx: %w", err)
	}
	defer tx.Rollback()

	attemptID := uuid.New()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO ai_feedback_attempts (
			id, learner_sentence_id, status, provider, model, prompt_version, request_hash,
			feedback_json, feedback_text, error_code, error_message,
			started_at, completed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, NULL, NULL, NULL, NULL, NULL, $8, $8)`,
		attemptID, failed.LearnerSentenceID, AttemptStatusPending, provider, model,
		PromptVersionSentenceFeedbackV1, failed.RequestHash, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert retry attempt: %w", err)
	}
	_, err = tx.ExecContext(ctx,
		`UPDATE learner_sentences SET status = $1, updated_at = $2 WHERE id = $3`,
		SentenceStatusSubmitted, now, failed.LearnerSentenceID,
	)
	if err != nil {
		return nil, fmt.Errorf("reset learner sentence for retry: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit retry attempt: %w", err)
	}
	return &PendingAttempt{SentenceID: failed.LearnerSentenceID, AttemptID: attemptID}, nil
}

func (r *PostgreSQLRepository) CompleteFeedbackAttempt(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, failureCode, failureMessage string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if feedback != nil {
		rawJSON, err := json.Marshal(feedback.RawJSON)
		if err != nil {
			return fmt.Errorf("marshal feedback json: %w", err)
		}
		_, err = tx.ExecContext(ctx,
			`UPDATE ai_feedback_attempts
			 SET status = $1, feedback_json = $2, feedback_text = $3, completed_at = $4, updated_at = $5
			 WHERE id = $6`,
			AttemptStatusSucceeded, rawJSON, feedback.Explanation, now, now, pending.AttemptID,
		)
		if err != nil {
			return fmt.Errorf("update attempt succeeded: %w", err)
		}
		_, err = tx.ExecContext(ctx,
			`UPDATE learner_sentences SET status = $1, updated_at = $2 WHERE id = $3`,
			SentenceStatusFeedbackReady, now, pending.SentenceID,
		)
		if err != nil {
			return fmt.Errorf("update sentence feedback ready: %w", err)
		}
	} else {
		code := failureCode
		if code == "" {
			code = ErrorCodeTemporaryFailure
		}
		_, err = tx.ExecContext(ctx,
			`UPDATE ai_feedback_attempts
			 SET status = $1, error_code = $2, error_message = $3, completed_at = $4, updated_at = $5
			 WHERE id = $6`,
			AttemptStatusFailed, code, failureMessage, now, now, pending.AttemptID,
		)
		if err != nil {
			return fmt.Errorf("update attempt failed: %w", err)
		}
		_, err = tx.ExecContext(ctx,
			`UPDATE learner_sentences SET status = $1, updated_at = $2 WHERE id = $3`,
			SentenceStatusFeedbackFailed, now, pending.SentenceID,
		)
		if err != nil {
			return fmt.Errorf("update sentence feedback failed: %w", err)
		}
	}

	return tx.Commit()
}

// GetFeedbackAttemptOwner implements Repository.
func (r *PostgreSQLRepository) GetFeedbackAttemptOwner(ctx context.Context, attemptID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT ls.user_id
		 FROM ai_feedback_attempts afa
		 JOIN learner_sentences ls ON ls.id = afa.learner_sentence_id
		 WHERE afa.id = $1`,
		attemptID,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrTargetNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("get feedback attempt owner: %w", err)
	}
	return userID, nil
}

// CreateQualityReviewReport implements Repository. The unique attempt key is
// the final idempotency guard, including across concurrent requests.
func (r *PostgreSQLRepository) CreateQualityReviewReport(ctx context.Context, report QualityReviewReport) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_feedback_quality_review_reports (
			id, ai_feedback_attempt_id, user_id, reason, state, classification, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NULL, $6, $6)
		ON CONFLICT (ai_feedback_attempt_id) DO NOTHING`,
		report.ID, report.AttemptID, report.UserID, report.Reason, report.State, report.CreatedAt,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows == 0 {
		var reason string
		if err := r.db.QueryRowContext(ctx, `SELECT reason FROM ai_feedback_quality_review_reports WHERE ai_feedback_attempt_id = $1 AND user_id = $2`, report.AttemptID, report.UserID).Scan(&reason); err != nil {
			return false, err
		}
		if reason != report.Reason {
			return false, ErrReportIdempotencyConflict
		}
	}
	return rows == 1, nil
}

// RequestHash computes the deduplication key for a sentence-feedback request.
func RequestHash(userID uuid.UUID, attemptID uuid.UUID, targetWord, normalizedSentence, promptVersion string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s|%s", userID.String(), attemptID.String(), targetWord, normalizedSentence, promptVersion)
	return hex.EncodeToString(h.Sum(nil))
}
