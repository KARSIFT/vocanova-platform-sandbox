package accounts

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// memoryIdempotencyEntry is the per-key payload the in-memory
// idempotency store records.
type memoryIdempotencyEntry struct {
	fingerprint string
}

// MemoryIdempotencyStore is a deterministic in-process
// idempotency store for the accounts package. It mirrors the
// (user, operation, key) -> fingerprint shape the learning
// package's MemoryIdempotencyStore already implements; the
// interface is duplicated locally (rather than imported) so the
// accounts package has no upstream dependency on the learning
// module. The production wiring in openapi.go / cmd/api is
// expected to pass the existing learning.PostgreSQLIdempotencyStore
// via a tiny adapter that maps the learning.IdempotencyStatus
// enum to accounts.IdempotencyStatus.
type MemoryIdempotencyStore struct {
	mu   sync.Mutex
	keys map[string]memoryIdempotencyEntry
}

// NewMemoryIdempotencyStore creates an empty in-memory
// idempotency store. Exported so the API-layer and the
// accounts-level tests can use the same constructor.
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{keys: map[string]memoryIdempotencyEntry{}}
}

func idempotencyCacheKey(userID uuid.UUID, operation, key string) string {
	return userID.String() + "|" + operation + "|" + key
}

// Check returns IdempotencyMatch when the stored fingerprint
// equals the supplied one, IdempotencyConflict when the stored
// fingerprint differs, and IdempotencyAbsent when no row
// exists.
func (s *MemoryIdempotencyStore) Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.keys[idempotencyCacheKey(userID, operation, key)]
	if !ok {
		return IdempotencyAbsent, nil
	}
	if entry.fingerprint == fingerprint {
		return IdempotencyMatch, nil
	}
	return IdempotencyConflict, nil
}

// Record stores the (user, operation, key, fingerprint) tuple.
// A second Record with the same key is a no-op when the
// fingerprint matches, and a conflict (no overwrite) when it
// differs — matching the SQL ON CONFLICT (user_id, operation,
// key) DO NOTHING discipline the learning module uses.
func (s *MemoryIdempotencyStore) Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cacheKey := idempotencyCacheKey(userID, operation, key)
	if existing, ok := s.keys[cacheKey]; ok && existing.fingerprint != fingerprint {
		return nil
	}
	s.keys[cacheKey] = memoryIdempotencyEntry{fingerprint: fingerprint}
	return nil
}
