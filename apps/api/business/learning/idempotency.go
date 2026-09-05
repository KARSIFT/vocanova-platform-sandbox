package learning

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// idempotencyRetention is the DOC-07 lifetime for a client-provided
// Idempotency-Key. A key is expired at the exact 24-hour boundary.
const idempotencyRetention = 24 * time.Hour

// MemoryIdempotencyStore is a deterministic, in-process idempotency store for
// service and handler tests. It is not safe for concurrent use across processes
// and does not survive restarts.
type MemoryIdempotencyStore struct {
	mu   sync.Mutex
	keys map[string]memoryIdempotencyEntry
	now  func() time.Time
}

type memoryIdempotencyEntry struct {
	userID      uuid.UUID
	operation   string
	key         string
	fingerprint string
	createdAt   time.Time
}

// NewMemoryIdempotencyStore creates an empty in-memory idempotency store.
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{
		keys: make(map[string]memoryIdempotencyEntry),
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func idempotencyCacheKey(userID uuid.UUID, operation, key string) string {
	return fmt.Sprintf("%s:%s:%s", userID.String(), operation, key)
}

func (s *MemoryIdempotencyStore) Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cacheKey := idempotencyCacheKey(userID, operation, key)
	entry, ok := s.keys[cacheKey]
	if !ok {
		return IdempotencyAbsent, nil
	}
	if !entry.createdAt.After(s.now().UTC().Add(-idempotencyRetention)) {
		delete(s.keys, cacheKey)
		return IdempotencyAbsent, nil
	}
	if entry.fingerprint == fingerprint {
		return IdempotencyMatch, nil
	}
	return IdempotencyConflict, nil
}

func (s *MemoryIdempotencyStore) Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cacheKey := idempotencyCacheKey(userID, operation, key)
	now := s.now().UTC()
	// Match PostgreSQL's insert-or-replace-expired behavior: replaying an
	// active key must not renew its lifetime or replace its fingerprint.
	if entry, ok := s.keys[cacheKey]; ok && entry.createdAt.After(now.Add(-idempotencyRetention)) {
		return nil
	}
	s.keys[cacheKey] = memoryIdempotencyEntry{
		userID:      userID,
		operation:   operation,
		key:         key,
		fingerprint: fingerprint,
		createdAt:   now,
	}
	return nil
}
