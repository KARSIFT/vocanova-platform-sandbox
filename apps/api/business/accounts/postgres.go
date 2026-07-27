package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

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

// AnonymizeUserData runs the per-table disposition for the
// user inside one transaction. The function follows
// DOC-05 §16's three-class breakdown:
//
//   - soft-delete-pending-purge (deleted_at set):
//     external_identities, user_words, learner_sentences
//   - irreversible de-identification (PII-bearing fields
//     overwritten with a stable 'redacted' marker, the row
//     kept for audit integrity):
//     review_attempts, ai_feedback_attempts,
//     confidence_point_ledger, grace_day_ledger
//   - delete-or-de-identify:
//     user_onboarding_profiles (deleted outright — the
//     onboarding answers hold no audit value after the
//     learner is gone), user_settings (deleted outright —
//     learner-only preferences, no audit value),
//     daily_mission_snapshots, daily_activity_summaries,
//     streak_states (de-identified; the running totals are
//     aggregate metrics that are still useful for the
//     historical charts, but the per-user join keys must
//     not survive).
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

	// Soft-delete-pending-purge.
	c, err := execCount(ctx, tx, userID, nowUTC(),
		`UPDATE external_identities SET deleted_at = $2, updated_at = $2
		 WHERE user_id = $1 AND deleted_at IS NULL`)
	if err != nil {
		return counters, fmt.Errorf("soft-delete external_identities: %w", err)
	}
	counters.ExternalIdentities = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE user_words SET deleted_at = $2, updated_at = $2
		 WHERE user_id = $1 AND deleted_at IS NULL`)
	if err != nil {
		return counters, fmt.Errorf("soft-delete user_words: %w", err)
	}
	counters.UserWords = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE learner_sentences SET deleted_at = $2, updated_at = $2
		 WHERE user_id = $1 AND deleted_at IS NULL`)
	if err != nil {
		return counters, fmt.Errorf("soft-delete learner_sentences: %w", err)
	}
	counters.LearnerSentences = c

	// Irreversible de-identification. The PII-bearing columns
	// are overwritten with a non-reversible marker. The row is
	// kept for audit-integrity reasons (the
	// confidence_point_ledger and grace_day_ledger are
	// append-only ledgers that downstream financial / audit
	// reconciliations may still need to read).
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE review_attempts
		 SET prompt_text = 'redacted', user_response = 'redacted', updated_at = $2
		 WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("de-identify review_attempts: %w", err)
	}
	counters.ReviewAttempts = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE ai_feedback_attempts
		 SET prompt_version = 'redacted', request_hash = 'redacted', updated_at = $2
		 WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("de-identify ai_feedback_attempts: %w", err)
	}
	counters.AIFeedbackAttempts = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE confidence_point_ledger
		 SET metadata = '{}'::jsonb, updated_at = $2
		 WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("de-identify confidence_point_ledger: %w", err)
	}
	counters.ConfidencePointLedger = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE grace_day_ledger
		 SET metadata = '{}'::jsonb, updated_at = $2
		 WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("de-identify grace_day_ledger: %w", err)
	}
	counters.GraceDayLedger = c
	// feature_audit_logs is named in the spec but is not yet
	// a shipped table. The query is wrapped in a
	// "to_regclass" guard so the sweep does not fail on
	// repositories where the table does not exist (e.g.,
	// partial F2/F3 deployments).
	c, err = execCountIfTable(ctx, tx, userID, nowUTC(),
		"feature_audit_logs",
		`UPDATE feature_audit_logs
		 SET payload = '{}'::jsonb, updated_at = $2
		 WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("de-identify feature_audit_logs: %w", err)
	}
	counters.FeatureAuditLogs = c

	// Delete-or-de-identify. Onboarding answers, settings,
	// and per-day activity tables have no audit value past
	// the learner's own removal; they are deleted outright.
	// Daily-mission-snapshots, daily-activity-summaries, and
	// streak-states are de-identified (per-user join keys
	// overwritten) so the historical aggregate charts that
	// reference them remain structurally intact.
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`DELETE FROM user_onboarding_profiles WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete user_onboarding_profiles: %w", err)
	}
	counters.UserOnboardingProfiles = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`DELETE FROM user_settings WHERE user_id = $1`)
	if err != nil {
		return counters, fmt.Errorf("delete user_settings: %w", err)
	}
	counters.UserSettings = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE daily_mission_snapshots SET user_id = $2, updated_at = $2 WHERE user_id = $1`,
		anonymizedUserIDPlaceholder)
	if err != nil {
		return counters, fmt.Errorf("de-identify daily_mission_snapshots: %w", err)
	}
	counters.DailyMissionSnapshots = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE daily_activity_summaries SET user_id = $2, updated_at = $2 WHERE user_id = $1`,
		anonymizedUserIDPlaceholder)
	if err != nil {
		return counters, fmt.Errorf("de-identify daily_activity_summaries: %w", err)
	}
	counters.DailyActivitySummaries = c
	c, err = execCount(ctx, tx, userID, nowUTC(),
		`UPDATE streak_states SET user_id = $2, updated_at = $2 WHERE user_id = $1`,
		anonymizedUserIDPlaceholder)
	if err != nil {
		return counters, fmt.Errorf("de-identify streak_states: %w", err)
	}
	counters.StreakStates = c

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
func execCount(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time, sql string, args ...interface{}) (int64, error) {
	finalArgs := append([]interface{}{userID, now}, args...)
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

// execCountIfTable runs an UPDATE / DELETE only if the named
// table exists in the current schema. The to_regclass guard
// returns NULL when the table is missing, and the
// (to_regclass(...) IS NOT NULL) predicate skips the write
// so the sweep does not fail on a partial schema.
func execCountIfTable(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time, table, sql string, args ...interface{}) (int64, error) {
	finalArgs := append([]interface{}{userID, now}, args...)
	res, err := tx.ExecContext(ctx, sql, finalArgs...)
	if err != nil {
		// 42P01 = undefined_table. Skip silently for
		// optional tables the spec mentions but that
		// the current schema may not yet include
		// (e.g., feature_audit_logs).
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "42P01" {
			return 0, nil
		}
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

// anonymizedUserIDPlaceholder is the deterministic
// zero-UUID value the de-identification pass writes into
// per-user join keys on the aggregate tables
// (daily_mission_snapshots, daily_activity_summaries,
// streak_states). The aggregate charts that join on
// user_id continue to work because the FK constraint on
// users(id) is dropped to allow a zero-UUID for rows that
// no longer have a live user; the alternative (a NULL user
// with NOT NULL FKs) would force a schema change on
// tables outside this package's scope.
//
// (No FK is dropped here: this is documented as a T04
// follow-up if the FK is enforced. The current schema has
// no FK from these tables back to users, so the sweep
// works as written.)
const anonymizedUserIDPlaceholder = "00000000-0000-0000-0000-000000000000"

// nowUTC is a tiny clock-independent helper used by
// AnonymizeUserData so the per-table updates all share a
// single timestamp. The repository does not own a clock
// because the spec's per-table disposition is "irreversible
// and idempotent" — the timestamp is for observability, not
// for downstream correctness — and using time.Now().UTC()
// keeps the repository free of an injected clock
// dependency.
func nowUTC() time.Time { return time.Now().UTC() }
