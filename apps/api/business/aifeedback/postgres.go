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

// LoadTargetForReplay is deliberately narrower than LoadTarget: it resolves a
// target that the authenticated learner formerly owned so the service can find
// an exact immutable result. It must never be used to authorize a generation.
func (r *PostgreSQLRepository) LoadTargetForReplay(ctx context.Context, req LoadTargetRequest) (*Target, error) {
	switch req.Source {
	case SourceWordDetail:
		return r.loadTargetFromUserWordForReplay(ctx, req.UserID, req.AttemptID)
	case SourceReview:
		return r.loadTargetFromReviewAttemptForReplay(ctx, req.UserID, req.AttemptID)
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

func (r *PostgreSQLRepository) loadTargetFromUserWordForReplay(ctx context.Context, userID, userWordID uuid.UUID) (*Target, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT cw.id, cw.text, cw.normalized_text, cw.word_type, cw.difficulty_level,
		        wm.id, wm.part_of_speech, wm.short_definition, uw.id
		 FROM user_words uw
		 JOIN word_meanings wm ON wm.id = uw.meaning_id
		 JOIN canonical_words cw ON cw.id = wm.word_id
		 WHERE uw.id = $1 AND uw.user_id = $2`,
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
		 WHERE ra.id = $1 AND ra.user_id = $2 AND uw.deleted_at IS NULL`,
		reviewAttemptID, userID,
	)
	return r.scanTarget(row, uuid.Nil, &reviewAttemptID)
}

func (r *PostgreSQLRepository) loadTargetFromReviewAttemptForReplay(ctx context.Context, userID, reviewAttemptID uuid.UUID) (*Target, error) {
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
		`SELECT attempt.id, attempt.learner_sentence_id, attempt.status, attempt.provider,
		        attempt.model, attempt.prompt_version, attempt.request_hash,
		        attempt.feedback_json, attempt.feedback_text, attempt.error_code, attempt.error_message,
		        EXISTS (SELECT 1 FROM ai_feedback_quality_review_reports r WHERE r.ai_feedback_attempt_id = attempt.id),
		        attempt.created_at, sentence.submitted_at
		 FROM ai_feedback_attempts attempt
		 JOIN learner_sentences sentence ON sentence.id = attempt.learner_sentence_id
		 WHERE attempt.request_hash = $1
		 ORDER BY CASE attempt.status WHEN 'succeeded' THEN 0 WHEN 'pending' THEN 1 ELSE 2 END,
		          attempt.created_at DESC`,
		requestHash,
	)
	return r.scanStoredAttempt(row)
}

func (r *PostgreSQLRepository) scanStoredAttempt(row *sql.Row) (*StoredFeedbackAttempt, error) {
	return scanStoredAttempt(row)
}

// ListLearnerSentences returns the authenticated learner's retained sentence
// history. The lateral join selects the newest generation without exposing
// provider metadata or internal error text.
func (r *PostgreSQLRepository) ListLearnerSentences(ctx context.Context, req ListLearnerSentencesRequest) (*ListLearnerSentencesResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	args := []any{req.UserID}
	cursorClause := ""
	if req.AfterCursor != "" {
		cursor, err := decodeLearnerSentenceCursor(req.AfterCursor)
		if err != nil {
			return nil, err
		}
		cursorClause = " AND (ls.submitted_at, ls.id) < ($2, $3)"
		args = append(args, cursor.SubmittedAt, cursor.ID)
	}
	args = append(args, limit+1)
	limitPlaceholder := fmt.Sprintf("$%d", len(args))

	rows, err := r.db.QueryContext(ctx, `
		SELECT ls.id, COALESCE(afa.id::text, ''), COALESCE(cw.id::text, ''),
		       ls.status, COALESCE(afa.status, ''), ls.sentence_text,
		       COALESCE(afa.feedback_json, '{}'::jsonb),
		       COALESCE(afa.feedback_text, ''),
		       CASE WHEN afa.id IS NULL THEN false ELSE EXISTS (
		         SELECT 1 FROM ai_feedback_quality_review_reports report
		         WHERE report.ai_feedback_attempt_id = afa.id
		       ) END,
		       ls.submitted_at
		FROM learner_sentences ls
		LEFT JOIN word_meanings wm ON wm.id = ls.meaning_id
		LEFT JOIN canonical_words cw ON cw.id = wm.word_id
		LEFT JOIN LATERAL (
		  SELECT attempt.id, attempt.status, attempt.feedback_json, attempt.feedback_text
		  FROM ai_feedback_attempts attempt
		  WHERE attempt.learner_sentence_id = ls.id
		  ORDER BY attempt.created_at DESC, attempt.id DESC
		  LIMIT 1
		) afa ON true
		WHERE ls.user_id = $1 AND ls.deleted_at IS NULL`+cursorClause+`
		ORDER BY ls.submitted_at DESC, ls.id DESC
		LIMIT `+limitPlaceholder, args...)
	if err != nil {
		return nil, fmt.Errorf("list learner sentences: %w", err)
	}
	defer rows.Close()

	items := make([]LearnerSentence, 0, limit+1)
	for rows.Next() {
		item, err := scanLearnerSentence(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate learner sentences: %w", err)
	}

	response := &ListLearnerSentencesResponse{Items: items}
	if len(items) > limit {
		response.Items = items[:limit]
		last := response.Items[len(response.Items)-1]
		response.NextCursor = encodeLearnerSentenceCursor(learnerSentenceCursor{SubmittedAt: last.CreatedAt, ID: last.ID})
	}
	return response, nil
}

// GetLearnerSentence returns one retained sentence only when it belongs to the
// authenticated learner.
func (r *PostgreSQLRepository) GetLearnerSentence(ctx context.Context, userID, sentenceID uuid.UUID) (*LearnerSentence, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT ls.id, COALESCE(afa.id::text, ''), COALESCE(cw.id::text, ''),
		       ls.status, COALESCE(afa.status, ''), ls.sentence_text,
		       COALESCE(afa.feedback_json, '{}'::jsonb),
		       COALESCE(afa.feedback_text, ''),
		       CASE WHEN afa.id IS NULL THEN false ELSE EXISTS (
		         SELECT 1 FROM ai_feedback_quality_review_reports report
		         WHERE report.ai_feedback_attempt_id = afa.id
		       ) END,
		       ls.submitted_at
		FROM learner_sentences ls
		LEFT JOIN word_meanings wm ON wm.id = ls.meaning_id
		LEFT JOIN canonical_words cw ON cw.id = wm.word_id
		LEFT JOIN LATERAL (
		  SELECT attempt.id, attempt.status, attempt.feedback_json, attempt.feedback_text
		  FROM ai_feedback_attempts attempt
		  WHERE attempt.learner_sentence_id = ls.id
		  ORDER BY attempt.created_at DESC, attempt.id DESC
		  LIMIT 1
		) afa ON true
		WHERE ls.id = $1 AND ls.user_id = $2 AND ls.deleted_at IS NULL`, sentenceID, userID)
	item, err := scanLearnerSentence(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

type learnerSentenceScanner interface {
	Scan(dest ...any) error
}

func scanLearnerSentence(row learnerSentenceScanner) (LearnerSentence, error) {
	var sentenceID uuid.UUID
	var feedbackIDText, targetWordIDText string
	var sentenceStatus, attemptStatus, original string
	var feedbackJSON []byte
	var feedbackText string
	var reported bool
	var submittedAt time.Time
	if err := row.Scan(
		&sentenceID, &feedbackIDText, &targetWordIDText, &sentenceStatus,
		&attemptStatus, &original, &feedbackJSON, &feedbackText, &reported, &submittedAt,
	); err != nil {
		return LearnerSentence{}, err
	}
	feedback := make(map[string]any)
	if len(feedbackJSON) > 0 {
		if err := json.Unmarshal(feedbackJSON, &feedback); err != nil {
			return LearnerSentence{}, fmt.Errorf("unmarshal learner sentence feedback: %w", err)
		}
	}
	feedbackID, _ := uuid.Parse(feedbackIDText)
	targetWordID, _ := uuid.Parse(targetWordIDText)
	return learnerSentenceFromStored(
		sentenceID, feedbackID, targetWordID, sentenceStatus, attemptStatus,
		original, feedback, feedbackText, reported, submittedAt,
	), nil
}

type storedAttemptScanner interface {
	Scan(dest ...any) error
}

func scanStoredAttempt(row storedAttemptScanner) (*StoredFeedbackAttempt, error) {
	var a StoredFeedbackAttempt
	var feedbackJSON []byte
	var feedbackText, errorCode, errorMessage sql.NullString

	err := row.Scan(
		&a.ID, &a.LearnerSentenceID, &a.Status, &a.Provider, &a.Model, &a.PromptVersion, &a.RequestHash,
		&feedbackJSON, &feedbackText, &errorCode, &errorMessage, &a.Reported, &a.CreatedAt, &a.SubmittedAt,
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

	// LoadTarget is only a preflight. Claim the active saved word in this
	// transaction immediately before writing so an Unsave cannot commit between
	// eligibility validation and the new feedback generation.
	var activeUserWordID uuid.UUID
	if err := tx.QueryRowContext(ctx, `SELECT id FROM user_words
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL FOR UPDATE`, target.UserWordID, req.UserID).Scan(&activeUserWordID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTargetNotFound
		}
		return nil, fmt.Errorf("claim active user word: %w", err)
	}

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

	return &PendingAttempt{SentenceID: sentenceID, AttemptID: attemptID, SubmittedAt: now}, nil
}

// CreateRetryAttempt appends an immutable retry generation for the same
// learner sentence after a failed provider attempt. The partial unique index on
// request_hash prevents more than one pending or successful generation.
func (r *PostgreSQLRepository) CreateRetryAttempt(ctx context.Context, failed *StoredFeedbackAttempt, provider string, model string, now time.Time) (*RetryAttempt, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin retry tx: %w", err)
	}
	defer tx.Rollback()

	attemptID := uuid.New()
	for {
		result, err := tx.ExecContext(ctx,
			`INSERT INTO ai_feedback_attempts (
			id, learner_sentence_id, status, provider, model, prompt_version, request_hash,
			feedback_json, feedback_text, error_code, error_message,
			started_at, completed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, NULL, NULL, NULL, NULL, NULL, $8, $8)
		 ON CONFLICT (request_hash) WHERE status IN ('pending', 'succeeded') DO NOTHING`,
			attemptID, failed.LearnerSentenceID, AttemptStatusPending, provider, model,
			PromptVersionSentenceFeedbackV1, failed.RequestHash, now,
		)
		if err != nil {
			return nil, fmt.Errorf("insert retry attempt: %w", err)
		}
		created, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("check retry insert: %w", err)
		}
		if created == 1 {
			break
		}

		row := tx.QueryRowContext(ctx,
			`SELECT attempt.id, attempt.learner_sentence_id, attempt.status, attempt.provider,
			        attempt.model, attempt.prompt_version, attempt.request_hash,
			        attempt.feedback_json, attempt.feedback_text, attempt.error_code, attempt.error_message,
			        EXISTS (SELECT 1 FROM ai_feedback_quality_review_reports r WHERE r.ai_feedback_attempt_id = attempt.id),
			        attempt.created_at, sentence.submitted_at
			 FROM ai_feedback_attempts attempt
			 JOIN learner_sentences sentence ON sentence.id = attempt.learner_sentence_id
			 WHERE attempt.request_hash = $1 AND attempt.status IN ('pending', 'succeeded')
			 ORDER BY CASE attempt.status WHEN 'succeeded' THEN 0 ELSE 1 END, attempt.created_at DESC
			 LIMIT 1`, failed.RequestHash)
		existing, err := scanStoredAttempt(row)
		if err == nil && existing != nil {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("commit existing retry attempt: %w", err)
			}
			return &RetryAttempt{Existing: existing}, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("load active retry attempt: %w", err)
		}

		// A conflicting pending generation may have failed after the INSERT
		// observed it but before this statement's fresh read. A new key is an
		// explicit retry under DOC-09, so retry the insert rather than turning
		// that legitimate next generation into an internal error.
		attemptID = uuid.New()
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
	return &RetryAttempt{Pending: &PendingAttempt{
		SentenceID:  failed.LearnerSentenceID,
		AttemptID:   attemptID,
		SubmittedAt: failed.SubmittedAt,
	}}, nil
}

func (r *PostgreSQLRepository) CompleteFeedbackAttempt(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, failureCode, failureMessage string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if feedback != nil {
		rawJSON, err := json.Marshal(feedback.StructuredJSON())
		if err != nil {
			return fmt.Errorf("marshal feedback json: %w", err)
		}
		result, err := tx.ExecContext(ctx,
			`UPDATE ai_feedback_attempts
			 SET status = $1, feedback_json = $2, feedback_text = $3, completed_at = $4, updated_at = $5
			 WHERE id = $6 AND status = $7`,
			AttemptStatusSucceeded, rawJSON, feedback.Explanation, now, now, pending.AttemptID, AttemptStatusPending,
		)
		if err != nil {
			return fmt.Errorf("update attempt succeeded: %w", err)
		}
		updated, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check succeeded attempt update: %w", err)
		}
		if updated == 0 {
			return tx.Commit()
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
		result, err := tx.ExecContext(ctx,
			`UPDATE ai_feedback_attempts
			 SET status = $1, error_code = $2, error_message = $3, completed_at = $4, updated_at = $5
			 WHERE id = $6 AND status = $7`,
			AttemptStatusFailed, code, failureMessage, now, now, pending.AttemptID, AttemptStatusPending,
		)
		if err != nil {
			return fmt.Errorf("update attempt failed: %w", err)
		}
		updated, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check failed attempt update: %w", err)
		}
		if updated == 0 {
			return tx.Commit()
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

// CompleteSuccessfulFeedbackAttempt atomically persists a successful provider
// result and the mission/reward/activity accounting it triggers. The pending
// status predicate fences concurrent or ambiguous replays: only the caller
// that transitions pending may run accounting.
func (r *PostgreSQLRepository) CompleteSuccessfulFeedbackAttempt(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, now time.Time, completion SuccessfulFeedbackCompletion) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	rawJSON, err := json.Marshal(feedback.StructuredJSON())
	if err != nil {
		return false, fmt.Errorf("marshal feedback json: %w", err)
	}
	result, err := tx.ExecContext(ctx, `UPDATE ai_feedback_attempts
		SET status = $1, feedback_json = $2, feedback_text = $3, completed_at = $4, updated_at = $5
		WHERE id = $6 AND status = $7`, AttemptStatusSucceeded, rawJSON, feedback.Explanation, now, now, pending.AttemptID, AttemptStatusPending)
	if err != nil {
		return false, fmt.Errorf("update attempt succeeded: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check succeeded attempt update: %w", err)
	}
	if updated == 0 {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit settled attempt: %w", err)
		}
		return false, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE learner_sentences SET status = $1, updated_at = $2 WHERE id = $3`, SentenceStatusFeedbackReady, now, pending.SentenceID); err != nil {
		return false, fmt.Errorf("update sentence feedback ready: %w", err)
	}
	completed, err := completion(ctx, tx)
	if err != nil {
		return false, fmt.Errorf("complete mission accounting: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit successful feedback and accounting: %w", err)
	}
	return completed, nil
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

// CreateQualityReviewReport claims the operation key and persists the report in
// one transaction. The attempt uniqueness constraint independently prevents a
// second report submitted with a different key.
func (r *PostgreSQLRepository) CreateQualityReviewReport(ctx context.Context, report QualityReviewReport, idempotencyKey string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	fingerprint := fmt.Sprintf("%s|%s", report.AttemptID, report.Reason)
	var claimed string
	err = tx.QueryRowContext(ctx, `INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
		VALUES ($1, $2, 'report_sentence_feedback', $3, $4, $5)
		ON CONFLICT (user_id, operation, key) DO UPDATE
		SET fingerprint = EXCLUDED.fingerprint, created_at = EXCLUDED.created_at
		WHERE idempotency_keys.created_at <= $6
		RETURNING fingerprint`, uuid.New(), report.UserID, idempotencyKey, fingerprint, report.CreatedAt, report.CreatedAt.Add(-24*time.Hour)).Scan(&claimed)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT fingerprint FROM idempotency_keys
			WHERE user_id = $1 AND operation = 'report_sentence_feedback' AND key = $2`, report.UserID, idempotencyKey).Scan(&claimed)
	}
	if err != nil {
		return false, fmt.Errorf("claim report idempotency: %w", err)
	}
	if claimed != fingerprint {
		return false, ErrReportIdempotencyConflict
	}
	result, err := tx.ExecContext(ctx,
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
		if err := tx.QueryRowContext(ctx, `SELECT reason FROM ai_feedback_quality_review_reports WHERE ai_feedback_attempt_id = $1 AND user_id = $2`, report.AttemptID, report.UserID).Scan(&reason); err != nil {
			return false, err
		}
		if reason != report.Reason {
			return false, ErrReportIdempotencyConflict
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return rows == 1, nil
}

// RequestHash computes the deduplication key for a sentence-feedback request.
func RequestHash(userID uuid.UUID, attemptID uuid.UUID, targetWord, normalizedSentence, promptVersion string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s|%s", userID.String(), attemptID.String(), targetWord, normalizedSentence, promptVersion)
	return hex.EncodeToString(h.Sum(nil))
}
