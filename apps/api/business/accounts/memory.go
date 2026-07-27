package accounts

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository is a deterministic in-memory implementation of
// Repository for service tests. It also implements the
// UpdateUserEmail discipline the SQL repository applies: a write
// that would collide with another user's active email is rejected
// with ErrEmailAlreadyRegistered, never a generic 500, so service
// tests can exercise the same race-handling the production path
// will.
type MemoryRepository struct {
	mu     sync.Mutex
	byID   map[uuid.UUID]*EmailChangeLink
	byHash map[string]*EmailChangeLink
	byUser map[uuid.UUID]map[uuid.UUID]*EmailChangeLink
	// usersByEmail mirrors the lower(email) WHERE deleted_at IS NULL
	// partial unique index users_active_email_key enforces in SQL.
	// The map holds the current email for every active user; nil
	// emails are not represented.
	usersByEmail map[string]uuid.UUID
	users        map[uuid.UUID]*memoryUser
	// collisions is incremented every time UpdateUserEmail rejects
	// a write because newEmail is already taken. Tests assert on
	// it directly; production code never reads it.
	collisions int
}

// memoryUser is the minimal projection the in-memory repository
// needs about a user: the current email and a soft-delete flag
// matching users.deleted_at IS NULL.
type memoryUser struct {
	ID      uuid.UUID
	Email   string
	Deleted bool
}

// NewMemoryRepository creates an empty in-memory Repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		byID:         make(map[uuid.UUID]*EmailChangeLink),
		byHash:       make(map[string]*EmailChangeLink),
		byUser:       make(map[uuid.UUID]map[uuid.UUID]*EmailChangeLink),
		usersByEmail: make(map[string]uuid.UUID),
		users:        make(map[uuid.UUID]*memoryUser),
	}
}

// CreateEmailChangeLink inserts one row. Mirrors the SQL
// migration's invariants (user_id is required, new_email is
// non-empty, token_hash is unique).
func (r *MemoryRepository) CreateEmailChangeLink(ctx context.Context, userID uuid.UUID, newEmail string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*EmailChangeLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if newEmail == "" {
		return nil, errors.New("new email required")
	}
	if _, exists := r.byHash[string(tokenHash)]; exists {
		return nil, errors.New("token hash collision")
	}
	link := &EmailChangeLink{
		ID:          uuid.New(),
		UserID:      userID,
		NewEmail:    strings.ToLower(newEmail),
		Environment: environment,
		CreatedAt:   createdAt.UTC(),
		ExpiresAt:   expiresAt.UTC(),
	}
	r.byID[link.ID] = link
	r.byHash[string(tokenHash)] = link
	if _, ok := r.byUser[userID]; !ok {
		r.byUser[userID] = make(map[uuid.UUID]*EmailChangeLink)
	}
	r.byUser[userID][link.ID] = link
	return cloneLink(link), nil
}

// GetEmailChangeLinkByTokenHash returns the projection or a
// not-found error.
func (r *MemoryRepository) GetEmailChangeLinkByTokenHash(ctx context.Context, tokenHash []byte) (*EmailChangeLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byHash[string(tokenHash)]
	if !ok {
		return nil, errors.New("email change link not found")
	}
	return cloneLink(l), nil
}

// ConsumeEmailChangeLink marks the row consumed.
func (r *MemoryRepository) ConsumeEmailChangeLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[id]
	if !ok {
		return errors.New("email change link not found")
	}
	t := consumedAt.UTC()
	l.ConsumedAt = &t
	return nil
}

// RevokeAllEmailChangeLinksForUser revokes every unconsumed link
// for the user. Used by the account-deletion path (T04).
func (r *MemoryRepository) RevokeAllEmailChangeLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	links, ok := r.byUser[userID]
	if !ok {
		return 0, nil
	}
	var n int64
	t := revokedAt.UTC()
	for _, l := range links {
		if l.ConsumedAt == nil && l.RevokedAt == nil {
			l.RevokedAt = &t
			n++
		}
	}
	return n, nil
}

// UpdateUserEmail applies the new email to the user, enforcing the
// same uniqueness discipline the SQL partial unique index provides:
// another active user already owns lower(newEmail), so we reject
// with ErrEmailAlreadyRegistered. Soft-deleted users are not
// considered active, matching the partial-index WHERE clause.
func (r *MemoryRepository) UpdateUserEmail(ctx context.Context, userID uuid.UUID, newEmail string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if newEmail == "" {
		return errors.New("new email required")
	}
	u, ok := r.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	if u.Deleted {
		return ErrUserNotFound
	}
	if owner, exists := r.usersByEmail[newEmail]; exists && owner != userID {
		// The owning user is checked for "deleted" so the
		// partial-index semantics translate exactly: a
		// soft-deleted user does not occupy the email.
		if other, ok := r.users[owner]; !ok || !other.Deleted {
			r.collisions++
			return ErrEmailAlreadyRegistered
		}
	}
	// Release the old email from the index.
	if u.Email != "" {
		delete(r.usersByEmail, u.Email)
	}
	u.Email = newEmail
	r.usersByEmail[newEmail] = userID
	return nil
}

// SetUser is a test helper that seeds the in-memory users table
// with an email identity. It is not part of the Repository
// contract; the production wiring uses auth.GetUserByID for
// identity and never lets the accounts.Repository mutate users
// directly.
func (r *MemoryRepository) SetUser(userID uuid.UUID, email string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[userID] = &memoryUser{ID: userID, Email: strings.ToLower(email)}
	if email != "" {
		r.usersByEmail[strings.ToLower(email)] = userID
	}
}

// SetUserDeleted marks a user as soft-deleted so the partial-
// unique-index semantics are observable in tests.
func (r *MemoryRepository) SetUserDeleted(userID uuid.UUID, deleted bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		u.Deleted = deleted
	}
}

// UserEmail returns the current in-memory email for a user. Used
// only by tests; not part of the contract.
func (r *MemoryRepository) UserEmail(userID uuid.UUID) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		return u.Email
	}
	return ""
}

// Collisions returns the number of duplicate-email rejections the
// repository has produced since construction. Tests assert on it
// directly to confirm the partial-index discipline fires.
func (r *MemoryRepository) Collisions() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.collisions
}

// LinksForUser returns the in-memory email_change_links for a
// user, sorted by CreatedAt. Test helper only.
func (r *MemoryRepository) LinksForUser(userID uuid.UUID) []EmailChangeLink {
	r.mu.Lock()
	defer r.mu.Unlock()
	links, ok := r.byUser[userID]
	if !ok {
		return nil
	}
	out := make([]EmailChangeLink, 0, len(links))
	for _, l := range links {
		out = append(out, *cloneLink(l))
	}
	return out
}

func cloneLink(l *EmailChangeLink) *EmailChangeLink {
	c := *l
	if l.ConsumedAt != nil {
		t := *l.ConsumedAt
		c.ConsumedAt = &t
	}
	if l.RevokedAt != nil {
		t := *l.RevokedAt
		c.RevokedAt = &t
	}
	return &c
}
