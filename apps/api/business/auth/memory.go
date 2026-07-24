package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository is a deterministic in-memory repository for service tests.
// It is not concurrency-safe for the same email/user identity across racing
// requests; production uses PostgreSQLRepository.
type MemoryRepository struct {
	mu               sync.Mutex
	users            map[uuid.UUID]*User
	usersByEmail     map[string]*User
	sessions         map[uuid.UUID]*Session
	sessionsByHash   map[string]*Session
	magicLinks       map[uuid.UUID]*MagicLink
	magicLinksByHash map[string]*MagicLink
}

// NewMemoryRepository initializes an empty in-memory repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:            make(map[uuid.UUID]*User),
		usersByEmail:     make(map[string]*User),
		sessions:         make(map[uuid.UUID]*Session),
		sessionsByHash:   make(map[string]*Session),
		magicLinks:       make(map[uuid.UUID]*MagicLink),
		magicLinksByHash: make(map[string]*MagicLink),
	}
}

func (r *MemoryRepository) CreateUser(ctx context.Context, email string, verifiedAt *time.Time) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	u := &User{
		ID:              uuid.New(),
		Email:           email,
		Status:          "active",
		EmailVerifiedAt: verifiedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	r.users[u.ID] = u
	if email != "" {
		r.usersByEmail[email] = u
	}
	return u, nil
}

func (r *MemoryRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	copy := *u
	return &copy, nil
}

func (r *MemoryRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.usersByEmail[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	copy := *u
	return &copy, nil
}

func (r *MemoryRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return errors.New("user not found")
	}
	u.LastLoginAt = &t
	return nil
}

func (r *MemoryRepository) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash []byte, createdAt, expiresAt time.Time) (*Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := &Session{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: createdAt.UTC(),
		ExpiresAt: expiresAt.UTC(),
	}
	r.sessions[s.ID] = s
	r.sessionsByHash[string(tokenHash)] = s
	return s, nil
}

func (r *MemoryRepository) GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (*Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessionsByHash[string(tokenHash)]
	if !ok {
		return nil, errors.New("session not found")
	}
	copy := *s
	return &copy, nil
}

func (r *MemoryRepository) RevokeSession(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return errors.New("session not found")
	}
	s.RevokedAt = &revokedAt
	return nil
}

func (r *MemoryRepository) CreateMagicLink(ctx context.Context, email string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*MagicLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := &MagicLink{
		ID:          uuid.New(),
		Email:       email,
		Environment: environment,
		CreatedAt:   createdAt.UTC(),
		ExpiresAt:   expiresAt.UTC(),
	}
	r.magicLinks[m.ID] = m
	r.magicLinksByHash[string(tokenHash)] = m
	return m, nil
}

func (r *MemoryRepository) GetMagicLinkByTokenHash(ctx context.Context, tokenHash []byte) (*MagicLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.magicLinksByHash[string(tokenHash)]
	if !ok {
		return nil, errors.New("magic link not found")
	}
	copy := *m
	return &copy, nil
}

func (r *MemoryRepository) ConsumeMagicLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.magicLinks[id]
	if !ok {
		return errors.New("magic link not found")
	}
	m.ConsumedAt = &consumedAt
	return nil
}

func (r *MemoryRepository) RevokeMagicLink(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.magicLinks[id]
	if !ok {
		return errors.New("magic link not found")
	}
	m.RevokedAt = &revokedAt
	return nil
}

func (r *MemoryRepository) AttachMagicLinkUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.magicLinks[id]
	if !ok {
		return errors.New("magic link not found")
	}
	m.UserID = &userID
	return nil
}

func (r *MemoryRepository) CleanupExpiredSessions(ctx context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, s := range r.sessions {
		if !s.ExpiresAt.After(before) || (s.RevokedAt != nil && !s.RevokedAt.After(before)) {
			delete(r.sessions, id)
			for h, ss := range r.sessionsByHash {
				if ss.ID == id {
					delete(r.sessionsByHash, h)
					break
				}
			}
			count++
		}
	}
	return count, nil
}

func (r *MemoryRepository) CleanupExpiredMagicLinks(ctx context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, m := range r.magicLinks {
		if !m.ExpiresAt.After(before) || m.ConsumedAt != nil || (m.RevokedAt != nil && !m.RevokedAt.After(before)) {
			delete(r.magicLinks, id)
			for h, mm := range r.magicLinksByHash {
				if mm.ID == id {
					delete(r.magicLinksByHash, h)
					break
				}
			}
			count++
		}
	}
	return count, nil
}
