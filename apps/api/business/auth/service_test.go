package auth

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() Config {
	return Config{
		Environment:            "test",
		BaseURL:                "https://test.example.com",
		MagicLinkPath:          "/auth/magic",
		OAuthRedirectURI:       "https://test.example.com/auth/oauth/google/callback",
		OAuthRedirectAllowlist: []string{"https://test.example.com/app"},
		SessionLifetime:        30 * 24 * time.Hour,
		MagicLinkLifetime:      15 * time.Minute,
		OAuthStateLifetime:     10 * time.Minute,
		RateLimit: RateLimitConfig{
			MagicRequestWindow:  time.Hour,
			MagicRequestLimit:   5,
			MagicConsumeWindow:  time.Hour,
			MagicConsumeLimit:   5,
			OAuthStartWindow:    time.Hour,
			OAuthStartLimit:     10,
			OAuthCallbackWindow: time.Hour,
			OAuthCallbackLimit:  10,
			LogoutWindow:        time.Hour,
			LogoutLimit:         5,
		},
	}
}

func testService(t *testing.T) (*Service, *MemoryRepository, *email.Fake, *clock.Fixed) {
	t.Helper()
	return testServiceWithOAuth(t, nil)
}

func testServiceWithOAuth(t *testing.T, oauth OAuthProvider) (*Service, *MemoryRepository, *email.Fake, *clock.Fixed) {
	t.Helper()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	fake := &email.Fake{}
	limiter := NewFixedWindowRateLimiter(c, time.Hour, 100)
	cfg := testConfig()
	svc := NewService(repo, fake, oauth, c, limiter, cfg)
	return svc, repo, fake, c
}

// losingOAuthStateRepository simulates the callback that loses the atomic
// single-use claim after it has read an otherwise valid OAuth state.
type losingOAuthStateRepository struct{ *MemoryRepository }

func (r losingOAuthStateRepository) ConsumeOAuthState(context.Context, uuid.UUID, time.Time) (bool, error) {
	return false, nil
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

func TestRequestMagicLinkWithReturnToIncludesOnlySafeAppRelativeDestination(t *testing.T) {
	ctx := context.Background()

	t.Run("preserves an app-relative destination", func(t *testing.T) {
		svc, _, fake, _ := testService(t)
		require.NoError(t, svc.RequestMagicLinkWithReturnTo(ctx, "1.2.3.4", "user@example.com", "/reviews?mode=due"))

		message, ok := fake.Last()
		require.True(t, ok)
		linkText := strings.TrimSpace(strings.Split(message.BodyText, "\n\n")[1])
		link, err := url.Parse(linkText)
		require.NoError(t, err)
		assert.Equal(t, "/reviews?mode=due", link.Query().Get("returnTo"))
	})

	t.Run("drops an external destination", func(t *testing.T) {
		svc, _, fake, _ := testService(t)
		require.NoError(t, svc.RequestMagicLinkWithReturnTo(ctx, "1.2.3.4", "user@example.com", "https://evil.example/"))

		message, ok := fake.Last()
		require.True(t, ok)
		assert.NotContains(t, message.BodyText, "returnTo=")
	})
}

func TestRequestMagicLinkEmptyEmailReturnsNoErrorAndNoEmail(t *testing.T) {
	svc, _, fake, _ := testService(t)
	err := svc.RequestMagicLink(context.Background(), "1.2.3.4", "")
	require.NoError(t, err)
	assert.Len(t, fake.Sent, 0)
}

// reservedSyntheticTestEmail is the address VOC-050-T00 reserves for
// the deploy-seeded synthetic smoke-test account in these tests.
const reservedSyntheticTestEmail = "smoke-test-bot@synthetic.vocanova.invalid"

// reservedSyntheticKillSwitches leaves every real path enabled so the
// tests below prove the refusal comes from the reserved-identity guard
// alone, not from an incidentally-closed kill switch.
func reservedSyntheticKillSwitches() *KillSwitches {
	return &KillSwitches{
		MagicLinkEnabled:       true,
		OAuthEnabled:           true,
		NewSignupsEnabled:      true,
		ReservedSyntheticEmail: reservedSyntheticTestEmail,
	}
}

func TestRequestMagicLinkReservedSyntheticEmailSendsNothing(t *testing.T) {
	svc, repo, fake, _ := testService(t)
	svc.SetKillSwitches(reservedSyntheticKillSwitches())

	err := svc.RequestMagicLink(context.Background(), "1.2.3.4", "  Smoke-Test-Bot@Synthetic.Vocanova.Invalid ")

	require.NoError(t, err)
	assert.Len(t, fake.Sent, 0, "no sign-in link may be dispatched for the reserved synthetic identity")
	assert.Len(t, repo.magicLinks, 0, "no magic link may be persisted for the reserved synthetic identity")
}

func TestConsumeMagicLinkRejectsReservedSyntheticEmail(t *testing.T) {
	svc, _, _, _ := testService(t)
	svc.SetKillSwitches(reservedSyntheticKillSwitches())

	_, _, _, err := svc.ConsumeMagicLink(context.Background(), "1.2.3.4", "not-a-real-token", reservedSyntheticTestEmail)

	assert.ErrorIs(t, err, ErrInvalidMagicLink)
}

func TestSignupAllowedRefusesReservedSyntheticEmailEvenWhenAllowlisted(t *testing.T) {
	switches := reservedSyntheticKillSwitches()
	switches.NewSignupsEnabled = false
	switches.SignupAllowlist = map[string]struct{}{reservedSyntheticTestEmail: {}}

	assert.False(t, switches.signupAllowed(reservedSyntheticTestEmail))
	assert.False(t, switches.IsReservedSyntheticEmail("someone@example.com"))
}

func TestMintSyntheticSmokeTestSessionRequiresReservedIdentity(t *testing.T) {
	svc, _, _, _ := testService(t)

	_, _, err := svc.MintSyntheticSmokeTestSession(context.Background())

	assert.ErrorIs(t, err, ErrSyntheticSessionMintDisabled)
}

func TestMintSyntheticSmokeTestSessionRequiresSeededUser(t *testing.T) {
	svc, _, _, _ := testService(t)
	svc.SetKillSwitches(reservedSyntheticKillSwitches())

	_, _, err := svc.MintSyntheticSmokeTestSession(context.Background())

	assert.ErrorIs(t, err, ErrSyntheticUserNotSeeded)
}

func TestMintSyntheticSmokeTestSessionIssuesRealSessionForReservedIdentityOnly(t *testing.T) {
	svc, repo, _, _ := testService(t)
	svc.SetKillSwitches(reservedSyntheticKillSwitches())
	ctx := context.Background()

	_, err := repo.CreateUser(ctx, reservedSyntheticTestEmail, nil)
	require.NoError(t, err)

	session, token, err := svc.MintSyntheticSmokeTestSession(ctx)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, token)
	assert.Equal(t, "active", mustGetUser(t, repo, reservedSyntheticTestEmail).Status)

	validated, err := svc.ValidateSession(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, reservedSyntheticTestEmail, validated.Email)
}

func mustGetUser(t *testing.T, repo *MemoryRepository, email string) *User {
	t.Helper()
	user, err := repo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	return user
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

	ext, err := repo.GetExternalIdentity(ctx, "email", "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, ext.UserID)
	assert.Equal(t, "user@example.com", ext.ProviderEmail)
	assert.True(t, ext.ProviderEmailVerified)
}

func TestConsumeMagicLinkReusesExistingEmailIdentity(t *testing.T) {
	svc, repo, fake, _ := testService(t)
	ctx := context.Background()

	for range 2 {
		require.NoError(t, svc.RequestMagicLink(ctx, "1.2.3.4", "user@example.com"))
		msg, ok := fake.Last()
		require.True(t, ok)
		rawToken := extractTokenFromURL(t, msg.BodyText)
		_, _, _, err := svc.ConsumeMagicLink(ctx, "1.2.3.4", rawToken, "user@example.com")
		require.NoError(t, err)
	}

	assert.Len(t, repo.externalIdentities, 1)
	ext, err := repo.GetExternalIdentity(ctx, "email", "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "email", ext.Provider)
	assert.Equal(t, "user@example.com", ext.ProviderSubject)
}

func TestEnsureEmailIdentityRejectsAnotherUsersIdentity(t *testing.T) {
	svc, repo, _, _ := testService(t)
	ctx := context.Background()
	now := time.Now().UTC()
	owner, err := repo.CreateUser(ctx, "owner@example.com", &now)
	require.NoError(t, err)
	other, err := repo.CreateUser(ctx, "other@example.com", &now)
	require.NoError(t, err)
	_, err = repo.CreateExternalIdentity(ctx, owner.ID, "email", "owner@example.com", "owner@example.com", true)
	require.NoError(t, err)

	err = svc.ensureEmailIdentity(ctx, other.ID, "owner@example.com")
	require.ErrorIs(t, err, ErrInvalidMagicLink)
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

	svc2 := NewService(NewMemoryRepository(), &email.Fake{}, nil, clock.Real{}, NewFixedWindowRateLimiter(clock.Real{}, time.Hour, 100), Config{Environment: "other"})
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
	svc := NewService(repo, fake, nil, c, limiter, cfg)
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

func TestOAuthStartCreatesStateAndReturnsURL(t *testing.T) {
	oauth := NewFakeOAuthProvider(&OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true, DisplayName: "User", AvatarURL: "https://example.com/avatar.png"})
	svc, repo, _, _ := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	url, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	assert.Contains(t, url, "https://fake-oauth.example.com/auth")
	assert.Contains(t, url, stateToken)
	assert.NotEmpty(t, stateToken)

	// State is stored as a hash, not the raw token.
	require.Len(t, repo.oauthStates, 1)
	for _, o := range repo.oauthStates {
		assert.Equal(t, "test", o.Environment)
		assert.Equal(t, "google", o.Provider)
		assert.Nil(t, o.ConsumedAt)
	}
}

func TestOAuthCallbackCreatesUserAndSession(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true, DisplayName: "User", AvatarURL: "https://example.com/avatar.png"}
	oauth := NewFakeOAuthProvider(identity)
	svc, repo, _, _ := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	_, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)

	user, session, token, _, err := svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)
	assert.NotNil(t, user.EmailVerifiedAt)
	assert.Equal(t, "active", user.Status)
	assert.Equal(t, session.UserID, user.ID)
	assert.NotEmpty(t, token)

	// External identity is recorded.
	require.Len(t, repo.externalIdentities, 1)
	for _, ext := range repo.externalIdentities {
		assert.Equal(t, user.ID, ext.UserID)
		assert.Equal(t, "google", ext.Provider)
		assert.Equal(t, "sub-123", ext.ProviderSubject)
		assert.True(t, ext.ProviderEmailVerified)
	}

	// State is consumed.
	for _, o := range repo.oauthStates {
		assert.NotNil(t, o.ConsumedAt)
	}

	// Reusing the same code/state fails.
	_, _, _, _, err = svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)
	assert.ErrorIs(t, err, ErrInvalidOAuthState)

	// Validated session works.
	validated, err := svc.ValidateSession(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validated.ID)
}

func TestOAuthCallbackRejectsStateWhenAnotherCallbackClaimsIt(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	_, repo, fake, c := testServiceWithOAuth(t, oauth)
	svc := NewService(losingOAuthStateRepository{repo}, fake, oauth, c, NewFixedWindowRateLimiter(c, time.Hour, 100), testConfig())

	_, stateToken, err := svc.OAuthStart(context.Background(), "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)

	_, _, _, _, err = svc.OAuthCallback(context.Background(), "1.2.3.4", "auth-code", stateToken, stateToken)
	assert.ErrorIs(t, err, ErrInvalidOAuthState)
	assert.Empty(t, repo.sessions, "a callback that loses the state claim must not mint a session")
}

func TestOAuthCallbackLinksExistingUserByEmail(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	svc, repo, _, _ := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	// Create an existing user directly so an email-linked identity can be added.
	_, err := repo.CreateUser(ctx, "user@example.com", nil)
	require.NoError(t, err)

	_, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)

	user, _, _, _, err := svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)
	require.Len(t, repo.externalIdentities, 1)
	for _, ext := range repo.externalIdentities {
		assert.Equal(t, user.ID, ext.UserID)
	}
}

func TestOAuthCallbackRejectsMismatchedCookieState(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	svc, _, _, _ := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	_, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)

	_, _, _, _, err = svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, "other-state")
	assert.ErrorIs(t, err, ErrInvalidOAuthState)
}

func TestOAuthCallbackRejectsExpiredState(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	svc, _, _, c := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	url, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	_ = url

	c.Advance(11 * time.Minute)
	_, _, _, _, err = svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)
	assert.ErrorIs(t, err, ErrInvalidOAuthState)
}

func TestOAuthCallbackRejectsUnverifiedEmail(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: false}
	oauth := NewFakeOAuthProvider(identity)
	svc, _, _, _ := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	url, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	_ = url

	_, _, _, _, err = svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)
	assert.ErrorIs(t, err, ErrOAuthProviderFailed)
}

func TestOAuthCallbackRejectsDisabledUser(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	svc, repo, _, _ := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	u, err := repo.CreateUser(ctx, "user@example.com", nil)
	require.NoError(t, err)
	u.Status = "disabled"
	repo.users[u.ID] = u
	repo.usersByEmail[u.Email] = u

	url, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	_ = url

	_, _, _, _, err = svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)
	assert.ErrorIs(t, err, ErrUserDisabled)
}

func TestOAuthCallbackRejectsUnknownRedirectURI(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	svc, _, _, _ := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	_, _, err := svc.OAuthStart(ctx, "1.2.3.4", "https://evil.example.com/callback")
	assert.ErrorIs(t, err, ErrOAuthProviderFailed)
}

func TestOAuthCallbackRejectsReservedSyntheticEmail(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub-123", Email: reservedSyntheticTestEmail, EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	svc, repo, _, _ := testServiceWithOAuth(t, oauth)
	svc.SetKillSwitches(reservedSyntheticKillSwitches())
	ctx := context.Background()

	_, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	_, _, _, _, err = svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)

	assert.ErrorIs(t, err, ErrOAuthProviderFailed)
	assert.Len(t, repo.sessions, 0, "no session may be issued for the reserved synthetic identity")
}

func TestOAuthStartRateLimited(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	fake := &email.Fake{}
	cfg := testConfig()
	cfg.RateLimit.OAuthStartLimit = 2
	limiter := NewFixedWindowRateLimiter(c, time.Hour, cfg.RateLimit.OAuthStartLimit)
	oauth := NewFakeOAuthProvider(&OAuthIdentity{Subject: "sub", Email: "a@example.com", EmailVerified: true})
	svc := NewService(repo, fake, oauth, c, limiter, cfg)
	ctx := context.Background()

	_, _, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	_, _, err = svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	_, _, err = svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	assert.ErrorIs(t, err, ErrRateLimited)

	// A different IP is not blocked.
	_, _, err = svc.OAuthStart(ctx, "5.6.7.8", "https://test.example.com/app")
	require.NoError(t, err)
}

func TestCleanupRemovesExpiredOAuthStates(t *testing.T) {
	identity := &OAuthIdentity{Subject: "sub", Email: "a@example.com", EmailVerified: true}
	oauth := NewFakeOAuthProvider(identity)
	svc, repo, _, c := testServiceWithOAuth(t, oauth)
	ctx := context.Background()

	_, stateToken, err := svc.OAuthStart(ctx, "1.2.3.4", "https://test.example.com/app")
	require.NoError(t, err)
	_, _, _, _, err = svc.OAuthCallback(ctx, "1.2.3.4", "auth-code", stateToken, stateToken)
	require.NoError(t, err)

	c.Advance(11 * time.Minute)
	err = svc.Cleanup(ctx)
	require.NoError(t, err)
	assert.Len(t, repo.oauthStates, 0)
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
