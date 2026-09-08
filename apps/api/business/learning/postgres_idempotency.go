package learning

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgreSQLIdempotencyStore implements IdempotencyStore against the
// idempotency_keys table.
type PostgreSQLIdempotencyStore struct {
	db  *sql.DB
	now func() time.Time
}

// NewPostgreSQLIdempotencyStore creates an idempotency store backed by db.
func NewPostgreSQLIdempotencyStore(db *sql.DB) *PostgreSQLIdempotencyStore {
	return &PostgreSQLIdempotencyStore{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *PostgreSQLIdempotencyStore) Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error) {
	var stored string
	err := s.db.QueryRowContext(ctx,
		`SELECT fingerprint FROM idempotency_keys
		 WHERE user_id = $1 AND operation = $2 AND key = $3 AND created_at > $4`,
		userID, operation, key, s.now().UTC().Add(-idempotencyRetention),
	).Scan(&stored)
	if errors.Is(err, sql.ErrNoRows) {
		return IdempotencyAbsent, nil
	}
	if err != nil {
		return IdempotencyAbsent, fmt.Errorf("check idempotency: %w", err)
	}
	if stored == fingerprint {
		return IdempotencyMatch, nil
	}
	return IdempotencyConflict, nil
}

func (s *PostgreSQLIdempotencyStore) Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error {
	now := s.now().UTC()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id, operation, key) DO UPDATE
		 SET fingerprint = EXCLUDED.fingerprint, created_at = EXCLUDED.created_at
		 WHERE idempotency_keys.created_at <= $7`,
		uuid.New(), userID, operation, key, fingerprint, now, now.Add(-idempotencyRetention),
	)
	if err != nil {
		return fmt.Errorf("record idempotency: %w", err)
	}
	return nil
}

// CleanupExpired deletes at most limit records whose 24-hour replay window
// has elapsed. The bounded, skip-locked claim keeps cleanup safe when several
// API replicas run the same maintenance loop.
func (s *PostgreSQLIdempotencyStore) CleanupExpired(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		return 0, errors.New("cleanup limit must be positive")
	}
	result, err := s.db.ExecContext(ctx,
		`WITH expired AS (
		   SELECT id FROM idempotency_keys
		   WHERE created_at <= $1
		   -- The retention migration indexes created_at. Keeping the claim in
		   -- that order lets PostgreSQL stop after the batch instead of sorting
		   -- every expired row before it can apply LIMIT.
		   ORDER BY created_at
		   LIMIT $2
		   FOR UPDATE SKIP LOCKED
		 )
		 DELETE FROM idempotency_keys AS keys
		 USING expired
		 WHERE keys.id = expired.id`,
		s.now().UTC().Add(-idempotencyRetention), limit,
	)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired idempotency keys: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count cleaned idempotency keys: %w", err)
	}
	return int(deleted), nil
}
