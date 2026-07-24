package auth

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() Config {
	return Config{
		Environment:       "test",
		BaseURL:           "https://test.example.com",
		MagicLinkPath:     "/auth/magic",
		SessionLifetime:   30 * 24 * time.Hour,
		MagicLinkLifetime: 15 * time.Minute,
		RateLimit: RateLimitConfig{
			MagicRequestWindow: time.Hour,
			MagicRequestLimit:  5,
			MagicConsumeWindow: time.Hour,
			MagicConsumeLimit:  5,
			LogoutWindow:       time.Hour,
			LogoutLimit:        5,
		},
	}
}

func testService(t *testing.T) (*Service, *MemoryRepository, *email.Fake, *clock.Fixed) {
	t.Helper()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	fake := &email.Fake{}
	limiter := NewFixedWindowRateLimiter(c, time.Hour, 100)
	svc := NewService(repo, fake, c, limiter, testConfig())
	return svc, repo, fake, c
}

func TestRequestMagicLinkCreatesHashedLinkAndSendsEmail(t *testing.T) {
	svc, repo, fake, _ := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)

	msg, ok := fake.Last()
	require.True(t, ok)
	assert.Equal(t, "user@example.com", msg.To[0].Email)
	assert.Equal(t, "Sign in to Vocanova", msg.Subject)
	assert.Contains(t, msg.BodyText, "Use this single-use link to sign in")
	assert.Contains(t, msg.BodyText, "test.example.com/auth/magic")

	// The repository should store a hashed link, not the raw token.
	require.Len(t, repo.magicLinks, 1)
	for _, m := range repo.magicLinks {
		assert.Equal(t, "user@example.com", m.Email)
		assert.Equal(t, "test", m.Environment)
		assert.Nil(t, m.UserID)
		assert.Nil(t, m.ConsumedAt)
		assert.Nil(t, m.RevokedAt)
	}
}

func TestRequestMagicLinkEmptyEmailReturnsNoErrorAndNoEmail(t *testing.T) {
	svc, _, fake, _ := testService(t)
	err := svc.RequestMagicLink(context.Background(), "1.2.3.4", "")
	require.NoError(t, err)
	assert.Len(t, fake.Sent, 0)
}

func TestConsumeMagicLinkCreatesUserAndSession(t *testing.T) {
	svc, repo, fake, c := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	user, session, token, err := svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)
	assert.NotNil(t, user.EmailVerifiedAt)
	assert.Equal(t, "active", user.Status)
	assert.Equal(t, session.UserID, user.ID)
	assert.NotEmpty(t, token)
	assert.WithinDuration(t, c.Now().Add(30*24*time.Hour), session.ExpiresAt, time.Second)

	// Link is consumed.
	for _, m := range repo.magicLinks {
		assert.NotNil(t, m.ConsumedAt)
	}

	// Session can be validated.
	validated, err := svc.ValidateSession(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validated.ID)
}

func TestConsumeMagicLinkRejectsReplay(t *testing.T) {
	svc, _, fake, _ := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	_, _, _, err = svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	require.NoError(t, err)

	_, _, _, err = svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	assert.ErrorIs(t, err, ErrInvalidMagicLink)
}

func TestConsumeMagicLinkRejectsExpiredLink(t *testing.T) {
	svc, _, fake, c := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	c.Advance(16 * time.Minute)
	_, _, _, err = svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	assert.ErrorIs(t, err, ErrInvalidMagicLink)
}

func TestConsumeMagicLinkRejectsWrongEnvironment(t *testing.T) {
	svc, _, fake, _ := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	svc2 := NewService(NewMemoryRepository(), &email.Fake{}, clock.Real{}, NewFixedWindowRateLimiter(clock.Real{}, time.Hour, 100), Config{Environment: "other"})
	_, _, _, err = svc2.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	assert.ErrorIs(t, err, ErrInvalidMagicLink)
}

func TestConsumeMagicLinkRejectsWrongEmail(t *testing.T) {
	svc, _, fake, _ := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	_, _, _, err = svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "attacker@example.com")
	assert.ErrorIs(t, err, ErrInvalidMagicLink)
}

func TestConsumeMagicLinkRejectsDisabledUser(t *testing.T) {
	svc, repo, fake, _ := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	_, _, _, err = svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	require.NoError(t, err)
	u, err := repo.GetUserByEmail(ctx, "user@example.com")
	require.NoError(t, err)
	u.Status = "disabled"
	repo.users[u.ID] = u
	repo.usersByEmail[u.Email] = u

	err = svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ = fake.Last()
	rawToken = extractTokenFromURL(t, msg.BodyText)

	_, _, _, err = svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	assert.ErrorIs(t, err, ErrUserDisabled)
}

func TestLogoutRevokesSession(t *testing.T) {
	svc, repo, fake, _ := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	_, _, sessionToken, err := svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	require.NoError(t, err)

	err = svc.Logout(ctx, sessionToken)
	require.NoError(t, err)

	require.Len(t, repo.sessions, 1)
	for _, s := range repo.sessions {
		assert.NotNil(t, s.RevokedAt)
	}

	_, err = svc.ValidateSession(ctx, sessionToken)
	assert.ErrorIs(t, err, ErrAuthenticationRequired)
}

func TestValidateSessionRejectsExpired(t *testing.T) {
	svc, _, fake, c := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	_, _, sessionToken, err := svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	require.NoError(t, err)

	c.Advance(31 * 24 * time.Hour)
	_, err = svc.ValidateSession(ctx, sessionToken)
	assert.ErrorIs(t, err, ErrAuthenticationRequired)
}

func TestRequestMagicLinkRateLimited(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	fake := &email.Fake{}
	cfg := testConfig()
	cfg.RateLimit.MagicRequestLimit = 2
	limiter := NewFixedWindowRateLimiter(c, time.Hour, cfg.RateLimit.MagicRequestLimit)
	svc := NewService(repo, fake, c, limiter, cfg)
	ctx := context.Background()

	require.NoError(t, svc.RequestMagicLink(ctx, "1.2.3.4", "a@example.com"))
	require.NoError(t, svc.RequestMagicLink(ctx, "1.2.3.4", "b@example.com"))
	assert.ErrorIs(t, svc.RequestMagicLink(ctx, "1.2.3.4", "c@example.com"), ErrRateLimited)

	// A different IP is not blocked.
	require.NoError(t, svc.RequestMagicLink(ctx, "5.6.7.8", "d@example.com"))
}

func TestCleanupRemovesExpired(t *testing.T) {
	svc, repo, fake, c := testService(t)
	ctx := context.Background()

	err := svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com")
	require.NoError(t, err)
	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	_, _, sessionToken, err := svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
	require.NoError(t, err)

	err = svc.Logout(ctx, sessionToken)
	require.NoError(t, err)

	c.Advance(31 * 24 * time.Hour)
	err = svc.RequestMagicLink(ctx, "1.2.3.4", "other@example.com")
	require.NoError(t, err)
	c.Advance(16 * time.Minute)

	err = svc.Cleanup(ctx)
	require.NoError(t, err)

	assert.Len(t, repo.sessions, 0)
	assert.Len(t, repo.magicLinks, 0)
}

func extractTokenFromURL(t *testing.T, body string) string {
	t.Helper()
	start := ""
	for _, prefix := range []string{"\n\n", "href=\"", ">"} {
		if i := strings.Index(body, prefix); i >= 0 {
			candidate := body[i+len(prefix):]
			if j := strings.Index(candidate, "\n"); j >= 0 {
				candidate = candidate[:j]
			}
			if j := strings.Index(candidate, "\""); j >= 0 {
				candidate = candidate[:j]
			}
			if strings.Contains(candidate, "token=") {
				start = candidate
				break
			}
		}
	}
	require.NotEmpty(t, start, "no token URL found in body: %s", body)
	u, err := url.Parse(start)
	require.NoError(t, err)
	return u.Query().Get("token")
}
