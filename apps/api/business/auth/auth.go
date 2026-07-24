// Package auth implements the A1 authentication service and identity/session
// lifecycle. It depends on repository and external-provider interfaces only.
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User is the service-layer identity projection. It intentionally omits tokens
// and provider subjects.
type User struct {
	ID              uuid.UUID
	Email           string
	DisplayName     string
	AvatarURL       string
	Status          string
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Active reports whether the user may authenticate.
func (u User) Active() bool { return u.Status == "active" }

// Session is the service-layer session projection.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// Valid reports whether the session has not expired or been revoked.
func (s Session) Valid(now time.Time) bool {
	return now.Before(s.ExpiresAt) && s.RevokedAt == nil
}

// MagicLink is the service-layer magic-link projection.
type MagicLink struct {
	ID          uuid.UUID
	UserID      *uuid.UUID
	Email       string
	Environment string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ConsumedAt  *time.Time
	RevokedAt   *time.Time
}

// Valid reports whether the link has not expired, been consumed, or been revoked.
func (m MagicLink) Valid(now time.Time) bool {
	return now.Before(m.ExpiresAt) && m.ConsumedAt == nil && m.RevokedAt == nil
}

// Repository is the persistence boundary for the auth service. Implementations
// must not expose bearer tokens or raw secrets in errors or logs.
type Repository interface {
	// Users
	CreateUser(ctx context.Context, email string, verifiedAt *time.Time) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUserLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error

	// Sessions
	CreateSession(ctx context.Context, userID uuid.UUID, tokenHash []byte, createdAt, expiresAt time.Time) (*Session, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (*Session, error)
	RevokeSession(ctx context.Context, id uuid.UUID, revokedAt time.Time) error

	// Magic links
	CreateMagicLink(ctx context.Context, email string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*MagicLink, error)
	GetMagicLinkByTokenHash(ctx context.Context, tokenHash []byte) (*MagicLink, error)
	ConsumeMagicLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error
	RevokeMagicLink(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	AttachMagicLinkUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Cleanup
	CleanupExpiredSessions(ctx context.Context, before time.Time) (int64, error)
	CleanupExpiredMagicLinks(ctx context.Context, before time.Time) (int64, error)
}

// ExternalIdentity is the service-layer provider-identity projection.
type ExternalIdentity struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	Provider              string
	ProviderSubject       string
	ProviderEmail         string
	ProviderEmailVerified bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// OAuthProvider is the verified OAuth provider boundary.
type OAuthProvider interface {
	Verify(ctx context.Context, code, state, redirectURI string) (*OAuthIdentity, error)
}

// OAuthIdentity is the verified identity returned by an OAuth provider.
type OAuthIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
}
