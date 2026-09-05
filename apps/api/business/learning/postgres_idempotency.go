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
