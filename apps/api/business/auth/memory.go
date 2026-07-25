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
	mu                 sync.Mutex
	users              map[uuid.UUID]*User
	usersByEmail       map[string]*User
	sessions           map[uuid.UUID]*Session
	sessionsByHash     map[string]*Session
	magicLinks         map[uuid.UUID]*MagicLink
	magicLinksByHash   map[string]*MagicLink
	oauthStates        map[uuid.UUID]*OAuthState
	oauthStatesByHash  map[string]*OAuthState
	externalIdentities map[uuid.UUID]*ExternalIdentity
	externalByProvider map[string]*ExternalIdentity
}

// NewMemoryRepository initializes an empty in-memory repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:              make(map[uuid.UUID]*User),
		usersByEmail:       make(map[string]*User),
		sessions:           make(map[uuid.UUID]*Session),
		sessionsByHash:     make(map[string]*Session),
		magicLinks:         make(map[uuid.UUID]*MagicLink),
		magicLinksByHash:   make(map[string]*MagicLink),
		oauthStates:        make(map[uuid.UUID]*OAuthState),
		oauthStatesByHash:  make(map[string]*OAuthState),
		externalIdentities: make(map[uuid.UUID]*ExternalIdentity),
		externalByProvider: make(map[string]*ExternalIdentity),
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

// SetUserStatus is a test helper to change a user's status without expanding the
// production Repository interface.
func (r *MemoryRepository) SetUserStatus(id uuid.UUID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return errors.New("user not found")
	}
	u.Status = status
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

func (r *MemoryRepository) CreateOAuthState(ctx context.Context, tokenHash []byte, environment, provider, appReturnURL string, createdAt, expiresAt time.Time) (*OAuthState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o := &OAuthState{
		ID:           uuid.New(),
		Environment:  environment,
		Provider:     provider,
		AppReturnURL: appReturnURL,
		CreatedAt:    createdAt.UTC(),
		ExpiresAt:    expiresAt.UTC(),
	}
	r.oauthStates[o.ID] = o
	r.oauthStatesByHash[string(tokenHash)] = o
	return o, nil
}

func (r *MemoryRepository) GetOAuthStateByTokenHash(ctx context.Context, tokenHash []byte) (*OAuthState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.oauthStatesByHash[string(tokenHash)]
	if !ok {
		return nil, errors.New("oauth state not found")
	}
	copy := *o
	return &copy, nil
}

func (r *MemoryRepository) ConsumeOAuthState(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.oauthStates[id]
	if !ok {
		return errors.New("oauth state not found")
	}
	o.ConsumedAt = &consumedAt
	return nil
}

func (r *MemoryRepository) GetExternalIdentity(ctx context.Context, provider, providerSubject string) (*ExternalIdentity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := provider + ":" + providerSubject
	ext, ok := r.externalByProvider[key]
	if !ok {
		return nil, errors.New("external identity not found")
	}
	copy := *ext
	return &copy, nil
}

func (r *MemoryRepository) CreateExternalIdentity(ctx context.Context, userID uuid.UUID, provider, providerSubject, providerEmail string, providerEmailVerified bool) (*ExternalIdentity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	ext := &ExternalIdentity{
		ID:                    uuid.New(),
		UserID:                userID,
		Provider:              provider,
		ProviderSubject:       providerSubject,
		ProviderEmail:         providerEmail,
		ProviderEmailVerified: providerEmailVerified,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	r.externalIdentities[ext.ID] = ext
	r.externalByProvider[provider+":"+providerSubject] = ext
	return ext, nil
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

func (r *MemoryRepository) CleanupExpiredOAuthStates(ctx context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, o := range r.oauthStates {
		if !o.ExpiresAt.After(before) || o.ConsumedAt != nil {
			delete(r.oauthStates, id)
			for h, oo := range r.oauthStatesByHash {
				if oo.ID == id {
					delete(r.oauthStatesByHash, h)
					break
				}
			}
			count++
		}
	}
	return count, nil
}
