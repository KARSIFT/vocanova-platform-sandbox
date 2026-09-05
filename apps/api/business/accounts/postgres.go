package accounts

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ExportPersonalData builds the complete learner-visible export in PostgreSQL
// so every collection is scoped by the same requester id. It intentionally
// does not select session/token tables, raw provider request material, or
// internal feedback reports/classifications.
func (r *PostgreSQLRepository) ExportPersonalData(ctx context.Context, userID uuid.UUID) (json.RawMessage, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	var payload []byte
	err := r.db.QueryRowContext(ctx, `
SELECT jsonb_build_object(
 'schemaVersion', '1.0', 'exportedAt', now(),
 'profile', (SELECT jsonb_build_object('id', id, 'email', email, 'displayName', display_name, 'avatarUrl', avatar_url, 'onboardingStatus', onboarding_status, 'emailVerifiedAt', email_verified_at, 'createdAt', created_at, 'updatedAt', updated_at) FROM users WHERE id = $1),
 'settings', COALESCE((SELECT jsonb_build_object('timezone', timezone, 'dailyReviewTarget', daily_review_target, 'reviewIntervalPreset', review_interval_preset, 'notificationsEnabled', notifications_enabled, 'marketingEmailsEnabled', marketing_emails_enabled, 'appLanguage', app_language, 'createdAt', created_at, 'updatedAt', updated_at) FROM user_settings WHERE user_id = $1), jsonb_build_object('timezone', 'UTC', 'dailyReviewTarget', 20, 'reviewIntervalPreset', 'vocanova_default', 'notificationsEnabled', true, 'marketingEmailsEnabled', false, 'appLanguage', 'en', 'createdAt', NULL, 'updatedAt', NULL)),
 'onboardingProfile', (SELECT jsonb_build_object('englishLevel', english_level, 'nativeLanguage', native_language, 'learningGoal', learning_goal, 'mainUseCase', main_use_case, 'dailyReviewTarget', daily_review_target, 'completedAt', completed_at, 'createdAt', created_at, 'updatedAt', updated_at) FROM user_onboarding_profiles WHERE user_id = $1),
 'savedWords', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', uw.id, 'meaningId', uw.meaning_id, 'wordId', cw.id, 'wordText', cw.text, 'partOfSpeech', wm.part_of_speech, 'shortDefinition', wm.short_definition, 'status', uw.status, 'source', uw.source, 'reviewStep', uw.review_step, 'nextReviewAt', uw.next_review_at, 'lastReviewedAt', uw.last_reviewed_at, 'lastResult', uw.last_result, 'lastRating', uw.last_rating, 'consecutiveCorrectCount', uw.consecutive_correct_count, 'consecutiveIncorrectCount', uw.consecutive_incorrect_count, 'totalReviewCount', uw.total_review_count, 'correctReviewCount', uw.correct_review_count, 'addedAt', uw.added_at, 'masteredAt', uw.mastered_at, 'ignoredAt', uw.ignored_at, 'createdAt', uw.created_at, 'updatedAt', uw.updated_at) ORDER BY uw.added_at) FROM user_words uw JOIN word_meanings wm ON wm.id = uw.meaning_id JOIN canonical_words cw ON cw.id = wm.word_id WHERE uw.user_id = $1 AND uw.deleted_at IS NULL), '[]'::jsonb),
 'reviewHistory', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', id, 'userWordId', user_word_id, 'meaningId', meaning_id, 'attemptType', attempt_type, 'promptType', prompt_type, 'result', result, 'rating', rating, 'reviewStepBefore', review_step_before, 'reviewStepAfter', review_step_after, 'answeredAt', answered_at, 'responseTimeMs', response_time_ms, 'selectedOptionMeaningId', selected_option_meaning_id, 'typedAnswer', typed_answer, 'wasHintUsed', was_hint_used, 'source', source, 'createdAt', created_at, 'updatedAt', updated_at) ORDER BY answered_at) FROM review_attempts WHERE user_id = $1), '[]'::jsonb),
 'sentenceFeedbackHistory', COALESCE((SELECT jsonb_agg(jsonb_build_object('sentence', jsonb_build_object('id', s.id, 'meaningId', s.meaning_id, 'userWordId', s.user_word_id, 'sentenceText', s.sentence_text, 'source', s.source, 'status', s.status, 'submittedAt', s.submitted_at, 'createdAt', s.created_at, 'updatedAt', s.updated_at), 'feedbackAttempts', COALESCE((SELECT jsonb_agg(jsonb_build_object('status', f.status, 'feedback', CASE WHEN f.feedback_json IS NULL THEN NULL ELSE jsonb_strip_nulls(jsonb_build_object('status', f.feedback_json->'status', 'targetWordUsedCorrectly', f.feedback_json->'target_word_used_correctly', 'correctedSentence', f.feedback_json->'corrected_sentence', 'explanation', f.feedback_json->'explanation', 'improvementTip', f.feedback_json->'improvement_tip')) END, 'feedbackText', f.feedback_text, 'startedAt', f.started_at, 'completedAt', f.completed_at, 'createdAt', f.created_at, 'updatedAt', f.updated_at) ORDER BY f.started_at) FROM ai_feedback_attempts f WHERE f.learner_sentence_id = s.id), '[]'::jsonb)) ORDER BY s.submitted_at) FROM learner_sentences s WHERE s.user_id = $1 AND s.deleted_at IS NULL), '[]'::jsonb),
 'dailyMissions', COALESCE((SELECT jsonb_agg(jsonb_build_object('localDate', local_date, 'timezone', timezone, 'reviewTarget', review_target, 'reviewsCompleted', reviews_completed, 'newWordTarget', new_word_target, 'newWordsCompleted', new_words_completed, 'sentencePracticeTarget', sentence_practice_target, 'sentencePracticesCompleted', sentence_practices_completed, 'policyVersion', policy_version, 'status', status, 'completedAt', completed_at, 'graceApplied', grace_applied, 'createdAt', created_at, 'updatedAt', updated_at) ORDER BY local_date) FROM daily_mission_snapshots WHERE user_id = $1), '[]'::jsonb),
 'dailyActivity', COALESCE((SELECT jsonb_agg(jsonb_build_object('localDate', local_date, 'timezone', timezone, 'reviewsAttempted', reviews_attempted, 'reviewsCorrect', reviews_correct, 'reviewsSkipped', reviews_skipped, 'wordsDiscovered', words_discovered, 'wordsAdded', words_added, 'sentencesSubmitted', sentences_submitted, 'aiFeedbackReceived', ai_feedback_received, 'confidencePointsEarned', confidence_points_earned, 'confidencePointsSpent', confidence_points_spent, 'createdAt', created_at, 'updatedAt', updated_at) ORDER BY local_date) FROM daily_activity_summaries WHERE user_id = $1), '[]'::jsonb),
 'confidencePointLedger', COALESCE((SELECT jsonb_agg(jsonb_build_object('amount', amount, 'balanceAfter', balance_after, 'reason', reason, 'sourceType', source_type, 'sourceId', source_id, 'occurredAt', occurred_at, 'createdAt', created_at, 'updatedAt', updated_at) ORDER BY occurred_at) FROM confidence_point_ledger WHERE user_id = $1), '[]'::jsonb),
 'graceDayLedger', COALESCE((SELECT jsonb_agg(jsonb_build_object('amount', amount, 'balanceAfter', balance_after, 'reason', reason, 'sourceType', source_type, 'sourceId', source_id, 'appliedToLocalDate', applied_to_local_date, 'timezone', timezone, 'createdAt', created_at, 'updatedAt', updated_at) ORDER BY applied_to_local_date) FROM grace_day_ledger WHERE user_id = $1), '[]'::jsonb),
 'streakState', (SELECT jsonb_build_object('currentStreakCount', current_streak_count, 'longestStreakCount', longest_streak_count, 'lastCompletedLocalDate', last_completed_local_date, 'lastActivityLocalDate', last_activity_local_date, 'timezone', timezone, 'status', status, 'createdAt', created_at, 'updatedAt', updated_at) FROM streak_states WHERE user_id = $1)
)`, userID).Scan(&payload)
	if err != nil {
		return nil, fmt.Errorf("query personal data export: %w", err)
	}
	return json.RawMessage(payload), nil
}

// PostgreSQLRepository implements Repository against the T03
// email_change_links migration. It owns no state of its own; the
// user-lookup and session-revocation operations the email-change
// flow needs are delegated to the AuthRepository the service is
// constructed with, mirroring the seed/settings split in the
// users module.
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a Repository backed by db.
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

// CreateEmailChangeLink inserts one row. The token_hash is the
// SHA-256 of the raw token (auth.TokenAndHash); only the hash is
// persisted. user_id is required (NOT NULL) and new_email is
// non-empty (mirrors the magic_links discipline with the
// D05-specific differences).
func (r *PostgreSQLRepository) CreateEmailChangeLink(ctx context.Context, userID uuid.UUID, newEmail string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*EmailChangeLink, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if newEmail == "" {
		return nil, errors.New("new email required")
	}
	id := uuid.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO email_change_links (id, user_id, new_email, token_hash, environment, created_at, expires_at)
		 VALUES ($1, $2, lower($3), $4, $5, $6, $7)`,
		id, userID, newEmail, tokenHash, environment, createdAt, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create email change link: %w", err)
	}
	return &EmailChangeLink{
		ID:          id,
		UserID:      userID,
		NewEmail:    strings.ToLower(newEmail),
		Environment: environment,
		CreatedAt:   createdAt.UTC(),
		ExpiresAt:   expiresAt.UTC(),
	}, nil
}

// GetEmailChangeLinkByTokenHash returns the projection or a
// not-found error.
func (r *PostgreSQLRepository) GetEmailChangeLinkByTokenHash(ctx context.Context, tokenHash []byte) (*EmailChangeLink, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, new_email, environment, created_at, expires_at, consumed_at, revoked_at
		 FROM email_change_links WHERE token_hash = $1`, tokenHash)
	var l EmailChangeLink
	var consumedAt, revokedAt sql.NullTime
	err := row.Scan(&l.ID, &l.UserID, &l.NewEmail, &l.Environment, &l.CreatedAt, &l.ExpiresAt, &consumedAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("email change link not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan email change link: %w", err)
	}
	if consumedAt.Valid {
		t := consumedAt.Time
		l.ConsumedAt = &t
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		l.RevokedAt = &t
	}
	return &l, nil
}

// ConsumeEmailChangeLink marks the row consumed. The service is
// expected to have already verified Valid() so the row exists
// and is unconsumed.
func (r *PostgreSQLRepository) ConsumeEmailChangeLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE email_change_links SET consumed_at = $1 WHERE id = $2`, consumedAt, id)
	if err != nil {
		return fmt.Errorf("consume email change link: %w", err)
	}
	return nil
}

// RevokeAllEmailChangeLinksForUser revokes every unconsumed link
// for the user in one statement. Used by the account-deletion
// path (T04).
func (r *PostgreSQLRepository) RevokeAllEmailChangeLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE email_change_links SET revoked_at = $1
		 WHERE user_id = $2 AND consumed_at IS NULL AND revoked_at IS NULL`,
		revokedAt, userID)
	if err != nil {
		return 0, fmt.Errorf("revoke all email change links: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("revoke all email change links rows: %w", err)
	}
	return n, nil
}

// UpdateUserEmail sets users.email for the user, enforcing the
// partial unique index users_active_email_key. A unique_violation
// (SQLSTATE 23505) on the email index is translated to
// ErrEmailAlreadyRegistered so the API layer returns a stable
// 409-style conflict response, never an unhandled 500
// (VOC-031-R02). The error is detected via lib/pq's pq.Error
// type, matching the convention already used by the auth and
// users modules.
func (r *PostgreSQLRepository) UpdateUserEmail(ctx context.Context, userID uuid.UUID, newEmail string, now time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = lower($2), updated_at = $3
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID, newEmail, now)
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		// Unique violation: lower(email) WHERE deleted_at IS NULL
		// already has this email. Translate to the stable
		// public error; the rest of the error chain is preserved
		// in the wrapped value for logs.
		return fmt.Errorf("%w: %s", ErrEmailAlreadyRegistered, pqErr.Message)
	}
	return fmt.Errorf("update user email: %w", err)
}

// CreateAccountDeletionRequest performs the entire
// deactivation transaction in one call. Inside one
// transaction, in this order:
//
//  1. UPDATE users SET status='deleted', deleted_at=now() WHERE
//     id=$1 AND deleted_at IS NULL. A no-match (0 rows) means
//     the user is missing or already deleted; both are
//     surfaced as ErrUserNotFound so the API layer maps to a
//     404 (never a 500).
//  2. UPDATE sessions SET revoked_at=$2 WHERE user_id=$1 AND
//     revoked_at IS NULL. Every active session for the
//     account is invalidated so the requester cannot stay
//     signed in across the deletion.
//  3. UPDATE magic_links SET revoked_at=$2 WHERE user_id=$1
//     AND consumed_at IS NULL AND revoked_at IS NULL. No
//     in-flight sign-in link can be consumed after deletion.
//  4. UPDATE email_change_links SET revoked_at=$2 WHERE
//     user_id=$1 AND consumed_at IS NULL AND revoked_at IS
//     NULL. Same posture for email-change tokens.
//  5. INSERT INTO account_deletion_requests
//     (..., status='deactivated', purge_after=$3) where
//     purge_after = requested_at + 30 days.
//
// The (user_id) UNIQUE constraint on
// account_deletion_requests is the authoritative guard
// against double-deactivation: a second concurrent call for
// the same user receives SQLSTATE 23505, which is translated
// to ErrAccountDeletionAlreadyInFlight. The transaction
// ensures the unique-violation path never leaves the user in
// a half-deactivated state.
func (r *PostgreSQLRepository) CreateAccountDeletionRequest(ctx context.Context, userID uuid.UUID, idempotencyKey string, now time.Time, purgeDelay time.Duration) (*AccountDeletionRequest, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if idempotencyKey == "" {
		return nil, errors.New("idempotency key required")
	}
	purgeAfter := now.Add(purgeDelay)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE users SET status = 'deleted', deleted_at = $2, updated_at = $2
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID, now,
	)
	if err != nil {
		return nil, fmt.Errorf("deactivate user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("deactivate user rows: %w", err)
	}
	if rows == 0 {
		return nil, ErrUserNotFound
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = $2
		 WHERE user_id = $1 AND revoked_at IS NULL`,
		userID, now,
	); err != nil {
		return nil, fmt.Errorf("revoke sessions: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE magic_links SET revoked_at = $2
		 WHERE user_id = $1 AND consumed_at IS NULL AND revoked_at IS NULL`,
		userID, now,
	); err != nil {
		return nil, fmt.Errorf("revoke magic links: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE email_change_links SET revoked_at = $2
		 WHERE user_id = $1 AND consumed_at IS NULL AND revoked_at IS NULL`,
		userID, now,
	); err != nil {
		return nil, fmt.Errorf("revoke email change links: %w", err)
	}

	id := uuid.New()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO account_deletion_requests (
			id, user_id, status, requested_at, purge_after,
			idempotency_key, created_at, updated_at
		) VALUES (
			$1, $2, 'deactivated', $3, $4, $5, $3, $3
		)`,
		id, userID, now, purgeAfter, idempotencyKey,
	); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			// (user_id) UNIQUE on account_deletion_requests
			// means a deletion for this user is already
			// in flight. Translate to the stable 409.
			return nil, fmt.Errorf("%w: %s", ErrAccountDeletionAlreadyInFlight, pqErr.Message)
		}
		return nil, fmt.Errorf("insert account deletion request: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &AccountDeletionRequest{
		ID:             id,
		UserID:         userID,
		Status:         "deactivated",
		RequestedAt:    now.UTC(),
		PurgeAfter:     purgeAfter.UTC(),
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now.UTC(),
		UpdatedAt:      now.UTC(),
	}, nil
}

// GetAccountDeletionRequestByUserID returns the current row
// for the user. The (user_id) UNIQUE constraint means at most
// one row exists at any time; the sweep's claim step uses
// this read to get the row ID it then transitions to
// 'anonymizing'.
func (r *PostgreSQLRepository) GetAccountDeletionRequestByUserID(ctx context.Context, userID uuid.UUID) (*AccountDeletionRequest, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, status, requested_at, purge_after,
		        completed_at, idempotency_key, created_at, updated_at
		 FROM account_deletion_requests WHERE user_id = $1`, userID)
	return scanAccountDeletionRequest(row)
}

// ListDeactivatedRequestsDueForPurge returns up to limit
// rows whose status is 'deactivated' and whose purge_after is
// at or before now. The (status, purge_after) partial index
// makes this an index scan; the LIMIT caps the work one
// pass can do.
func (r *PostgreSQLRepository) ListDeactivatedRequestsDueForPurge(ctx context.Context, now time.Time, limit int) ([]AccountDeletionRequest, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, status, requested_at, purge_after,
		        completed_at, idempotency_key, created_at, updated_at
		 FROM account_deletion_requests
		 WHERE status = 'deactivated' AND purge_after <= $1
		 ORDER BY purge_after ASC
		 LIMIT $2`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list due deletion requests: %w", err)
	}
	defer rows.Close()
	var out []AccountDeletionRequest
	for rows.Next() {
		row, err := scanAccountDeletionRequestRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deletion requests: %w", err)
	}
	return out, nil
}

// ClaimAccountDeletionRequestForAnonymization atomically
// transitions a row from 'deactivated' to 'anonymizing'. The
// WHERE clause is the claim predicate: a losing claim (the
// row is already 'anonymizing' or 'completed', or the row
// has been removed) returns 0 rows, the call surfaces false,
// and the caller skips the row.
func (r *PostgreSQLRepository) ClaimAccountDeletionRequestForAnonymization(ctx context.Context, id uuid.UUID, now time.Time) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE account_deletion_requests
		 SET status = 'anonymizing', updated_at = $2
		 WHERE id = $1 AND status = 'deactivated'`,
		id, now,
	)
	if err != nil {
		return false, fmt.Errorf("claim deletion request: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("claim deletion request rows: %w", err)
	}
	return rows == 1, nil
}

// AnonymizeUserData runs the account-deletion disposition inside one
// transaction.  At this stage the account is already deactivated; this is
// the irreversible part of the staged process.
//
// The committed schema uses NOT NULL, ON DELETE RESTRICT user_id foreign
// keys throughout the learning tables.  Replacing those IDs with a sentinel
// would therefore fail, and retaining them would leave the data linkable.
// We delete learner-linked operational records instead.  This also removes
// feedback request hashes rather than attempting to redact them to a shared
// value that violates their unique index.  The retained users row is needed
// by account_deletion_requests, so only its identifying fields are erased.
//
// The function returns the per-table counts for the API /
// observability layer. SQL errors are wrapped in
// ErrAccountDeletionSweep so the service layer can recognize
// them without losing the underlying error chain.
func (r *PostgreSQLRepository) AnonymizeUserData(ctx context.Context, userID uuid.UUID) (AnonymizationCounters, error) {
	var counters AnonymizationCounters
	if userID == uuid.Nil {
		return counters, errors.New("user id required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return counters, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Purge quality reports before their parent feedback attempts and retain
	// an explicit affected-row count for the deletion audit.
	counters.AIQualityReviewReports, err = execCount(ctx, tx, userID,
		`DELETE FROM ai_feedback_quality_review_reports WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete quality review reports: %w", err)
	}
	// Delete dependent records before their parent learner rows. This order is
	// required by the committed ON DELETE RESTRICT constraints.
	c, err := execCount(ctx, tx, userID,
		`DELETE FROM ai_feedback_attempts AS attempt
		 USING learner_sentences AS sentence
		 WHERE attempt.learner_sentence_id = sentence.id
		   AND sentence.user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete ai_feedback_attempts: %w", err)
	}
	counters.AIFeedbackAttempts = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM review_attempts WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete review_attempts: %w", err)
	}
	counters.ReviewAttempts = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM learner_sentences WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete learner_sentences: %w", err)
	}
	counters.LearnerSentences = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM user_words WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete user_words: %w", err)
	}
	counters.UserWords = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM confidence_point_ledger WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete confidence_point_ledger: %w", err)
	}
	counters.ConfidencePointLedger = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM grace_day_ledger WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete grace_day_ledger: %w", err)
	}
	counters.GraceDayLedger = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM daily_mission_snapshots WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete daily_mission_snapshots: %w", err)
	}
	counters.DailyMissionSnapshots = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM daily_activity_summaries WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete daily_activity_summaries: %w", err)
	}
	counters.DailyActivitySummaries = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM streak_states WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete streak_states: %w", err)
	}
	counters.StreakStates = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM user_onboarding_profiles WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete user_onboarding_profiles: %w", err)
	}
	counters.UserOnboardingProfiles = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM user_settings WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete user_settings: %w", err)
	}
	counters.UserSettings = c
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM idempotency_keys WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete idempotency_keys: %w", err)
	}
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM email_change_links WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete email_change_links: %w", err)
	}
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM magic_links WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete magic_links: %w", err)
	}
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM sessions WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete sessions: %w", err)
	}
	c, err = execCount(ctx, tx, userID,
		`DELETE FROM external_identities WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete external_identities: %w", err)
	}
	counters.ExternalIdentities = c
	c, err = execCount(ctx, tx, userID,
		`UPDATE account_deletion_requests
		 SET idempotency_key = 'redacted:' || id::text, updated_at = NOW()
		 WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("redact deletion request idempotency key: %w", err)
	}
	c, err = execCount(ctx, tx, userID,
		`UPDATE users
		 SET email = NULL, display_name = NULL, avatar_url = NULL,
		     email_verified_at = NULL, last_login_at = NULL, updated_at = NOW()
		 WHERE id = $1`)
	if err != nil {
		return counters, fmt.Errorf("redact user identity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return counters, fmt.Errorf("commit: %w", err)
	}
	return counters, nil
}

// MarkAccountDeletionRequestCompleted transitions the row
// from 'anonymizing' to 'completed' and stamps completed_at.
// Idempotent: a second call on a 'completed' row is a no-op.
func (r *PostgreSQLRepository) MarkAccountDeletionRequestCompleted(ctx context.Context, id uuid.UUID, now time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE account_deletion_requests
		 SET status = 'completed', completed_at = $2, updated_at = $2
		 WHERE id = $1 AND status = 'anonymizing'`,
		id, now,
	)
	if err != nil {
		return fmt.Errorf("complete deletion request: %w", err)
	}
	return nil
}

func scanAccountDeletionRequest(row *sql.Row) (*AccountDeletionRequest, error) {
	var r AccountDeletionRequest
	var completedAt sql.NullTime
	err := row.Scan(&r.ID, &r.UserID, &r.Status, &r.RequestedAt, &r.PurgeAfter,
		&completedAt, &r.IdempotencyKey, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("account deletion request not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan account deletion request: %w", err)
	}
	if completedAt.Valid {
		t := completedAt.Time
		r.CompletedAt = &t
	}
	return &r, nil
}

func scanAccountDeletionRequestRows(rows *sql.Rows) (*AccountDeletionRequest, error) {
	var r AccountDeletionRequest
	var completedAt sql.NullTime
	err := rows.Scan(&r.ID, &r.UserID, &r.Status, &r.RequestedAt, &r.PurgeAfter,
		&completedAt, &r.IdempotencyKey, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan account deletion request row: %w", err)
	}
	if completedAt.Valid {
		t := completedAt.Time
		r.CompletedAt = &t
	}
	return &r, nil
}

// execCount runs an UPDATE / DELETE and returns the
// affected-row count. Centralized so the AnonymizeUserData
// code path can stay a flat list of named per-table
// statements.
func execCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, sql string, args ...interface{}) (int64, error) {
	finalArgs := append([]interface{}{userID}, args...)
	res, err := tx.ExecContext(ctx, sql, finalArgs...)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}
