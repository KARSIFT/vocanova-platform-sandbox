package learning

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// MemoryIdempotencyStore is a deterministic, in-process idempotency store for
// service and handler tests. It is not safe for concurrent use across processes
// and does not survive restarts.
type MemoryIdempotencyStore struct {
	mu   sync.Mutex
	keys map[string]memoryIdempotencyEntry
}

type memoryIdempotencyEntry struct {
	userID      uuid.UUID
	operation   string
	key         string
	fingerprint string
}

// NewMemoryIdempotencyStore creates an empty in-memory idempotency store.
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{keys: make(map[string]memoryIdempotencyEntry)}
}

func idempotencyCacheKey(userID uuid.UUID, operation, key string) string {
	return fmt.Sprintf("%s:%s:%s", userID.String(), operation, key)
}

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

func (s *MemoryIdempotencyStore) Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[idempotencyCacheKey(userID, operation, key)] = memoryIdempotencyEntry{
		userID:      userID,
		operation:   operation,
		key:         key,
		fingerprint: fingerprint,
	}
	return nil
}
