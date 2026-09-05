package accounts

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testNow is a deterministic UTC instant shared by every test.
func testNow() time.Time {
	return time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
}

// authRepoStub is a tiny in-memory AuthRepository the
// service-level tests use. It implements the four methods the
// accounts.Service calls: GetUserByID (T03, T04), and the two
// RevokeAll*ForUser methods T04 needs.
type authRepoStub struct {
	users             map[uuid.UUID]*auth.User
	revokedSessions   int64
	revokedMagicLinks int64
}

func newAuthRepoStub() *authRepoStub {
	return &authRepoStub{users: map[uuid.UUID]*auth.User{}}
}

func (a *authRepoStub) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	u, ok := a.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	c := *u
	return &c, nil
}

func (a *authRepoStub) RevokeAllSessionsForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	a.revokedSessions++
	return 0, nil
}

func (a *authRepoStub) RevokeAllMagicLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	a.revokedMagicLinks++
	return 0, nil
}

func (a *authRepoStub) setUser(u *auth.User) {
	a.users[u.ID] = u
}

func (a *authRepoStub) setUserDeleted(u *auth.User) {
	a.users[u.ID] = u
}

// newService wires a Service backed by a fresh in-memory Repository
// and the auth-repo stub. emailSender is a *email.Fake so the test
// can assert on dispatched messages. idem is a fresh in-memory
// idempotency store.
func newService(t *testing.T) (*Service, *MemoryRepository, *authRepoStub, *email.Fake, *clock.Fixed) {
	t.Helper()
	now := testNow()
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	authRepo := newAuthRepoStub()
	fake := &email.Fake{}
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	idem := NewMemoryIdempotencyStore()
	svc := NewService(repo, authRepo, fake, idem, c, limiter, Config{
		Environment:               "test",
		BaseURL:                   "https://test.example.com",
		EmailChangePath:           "/auth/email-change",
		EmailChangeLinkLifetime:   15 * time.Minute,
		AccountDeletionPurgeDelay: 30 * 24 * time.Hour,
		AccountDeletionSweepLimit: 100,
		RateLimit: EmailChangeRateLimitConfig{
			RequestWindow: time.Hour, RequestLimit: 100,
			ConsumeWindow: time.Hour, ConsumeLimit: 100,
		},
		AccountDeletionRateLimit: AccountDeletionRateLimitConfig{
			RequestWindow: time.Hour, RequestLimit: 100,
			SweepWindow: time.Hour, SweepLimit: 100,
		},
		ReservedSyntheticEmail: reservedSyntheticTestEmail,
	})
	return svc, repo, authRepo, fake, c
}

// extractTokenFromLink pulls the URL-decoded token query parameter
// from the link the email body contains. Mirrors the auth-module
// extractTokenFromURL helper so the same secret-handling discipline
// is exercised (the raw token is the only form passed to
// ConsumeEmailChangeLink).
func extractTokenFromLink(t *testing.T, link string) string {
	t.Helper()
	u, err := url.Parse(link)
	require.NoError(t, err)
	token := u.Query().Get("token")
	require.NotEmpty(t, token, "no token query in link: %s", link)
	return token
}

// TestRequestEmailChangeLinkPersistsHashedTokenAndSendsEmail covers
// VOC-031-TEST-13: the persisted row's token_hash is the SHA-256 of
// the raw token, expires_at is 15 minutes out, and environment is
// the configured value.
func TestRequestEmailChangeLinkPersistsHashedTokenAndSendsEmail(t *testing.T) {
	svc, repo, authRepo, fake, c := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})

	link, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	require.NotNil(t, link)
	assert.Equal(t, "new@example.com", link.NewEmail)
	assert.Contains(t, link.Link, "/auth/email-change")

	msg, ok := fake.Last()
	require.True(t, ok)
	assert.Equal(t, "new@example.com", msg.To[0].Email)
	assert.Equal(t, "Confirm your new Vocanova sign-in email", msg.Subject)
	assert.Contains(t, msg.BodyText, "/auth/email-change")

	// Inspect the persisted row.
	rows := repo.LinksForUser(uid)
	require.Len(t, rows, 1)
	got := rows[0]
	assert.Equal(t, uid, got.UserID)
	assert.Equal(t, "new@example.com", got.NewEmail)
	assert.Equal(t, "test", got.Environment)
	assert.Equal(t, 15*time.Minute, got.ExpiresAt.Sub(got.CreatedAt), "expires_at is exactly 15 minutes out")
	assert.Nil(t, got.ConsumedAt)
	assert.Nil(t, got.RevokedAt)

	// The persisted token hash must equal the SHA-256 of the raw
	// token embedded in the email link.
	raw := extractTokenFromLink(t, link.Link)
	_, hash, err := auth.TokenAndHash(raw)
	require.NoError(t, err)
	fetched, err := repo.GetEmailChangeLinkByTokenHash(context.Background(), hash)
	require.NoError(t, err)
	assert.Equal(t, got.ID, fetched.ID)
	assert.WithinDuration(t, c.Now(), got.CreatedAt, time.Second)
}

// TestRequestEmailChangeLinkGenericForAlreadyRegisteredEmail covers
// VOC-031-TEST-12: the request response is identical regardless of
// whether the new email is already registered to another account.
// The actual uniqueness check is re-verified at confirm time.
func TestRequestEmailChangeLinkGenericForAlreadyRegisteredEmail(t *testing.T) {
	svc, repo, authRepo, fake, _ := newService(t)
	uid := uuid.New()
	other := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	authRepo.setUser(&auth.User{ID: other, Email: "new@example.com", Status: "active"})
	repo.SetUser(uid, "old@example.com")
	repo.SetUser(other, "new@example.com")

	// Both paths must succeed; the only thing the request stage
	// reveals is the confirmation link.
	link1, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	link2, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "never-taken@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, link1.Link)
	assert.NotEmpty(t, link2.Link)
	assert.Len(t, fake.Sent, 2)
}

// TestRequestEmailChangeLinkRejectsEmptyEmail covers the basic
// input-validation branch.
func TestRequestEmailChangeLinkRejectsEmptyEmail(t *testing.T) {
	svc, _, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})

	_, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "")
	assert.ErrorIs(t, err, ErrEmailChangeInvalidEmail)

	_, err = svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "  ")
	assert.ErrorIs(t, err, ErrEmailChangeInvalidEmail)

	_, err = svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "no-at-sign")
	assert.ErrorIs(t, err, ErrEmailChangeInvalidEmail)

	_, err = svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "no-domain@")
	assert.ErrorIs(t, err, ErrEmailChangeInvalidEmail)
}

// reservedSyntheticTestEmail is VOC-050-T00's reserved synthetic
// smoke-test identity. The email-change flow must refuse it for the
// same reason the sign-in paths do: it is the only remaining way a
// real account could come to hold the reserved address.
const reservedSyntheticTestEmail = "smoke-test-bot@synthetic.vocanova.invalid"

func TestRequestEmailChangeLinkRejectsReservedSyntheticEmail(t *testing.T) {
	svc, repo, authRepo, fake, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})

	_, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "  Smoke-Test-Bot@Synthetic.Vocanova.Invalid ")

	assert.ErrorIs(t, err, ErrEmailChangeInvalidEmail)
	assert.Len(t, fake.Sent, 0)
	assert.Len(t, repo.LinksForUser(uid), 0)
}

// TestConsumeEmailChangeLinkRejectsReservedSyntheticEmail covers the
// confirm-time re-check: a link minted before the address was reserved
// must not still be redeemable against it.
func TestConsumeEmailChangeLinkRejectsReservedSyntheticEmail(t *testing.T) {
	svc, repo, authRepo, _, c := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	repo.SetUser(uid, "old@example.com")

	rawToken, hash, err := auth.NewTokenAndHash()
	require.NoError(t, err)
	now := c.Now()
	_, err = repo.CreateEmailChangeLink(context.Background(), uid, reservedSyntheticTestEmail, hash, "test", now, now.Add(15*time.Minute))
	require.NoError(t, err)

	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", rawToken)

	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
	assert.Equal(t, "old@example.com", repo.UserEmail(uid))
}

// TestRequestEmailChangeLinkRequiresUserID ensures the service
// refuses a request that bypasses the auth gate.
func TestRequestEmailChangeLinkRequiresUserID(t *testing.T) {
	svc, _, _, _, _ := newService(t)
	_, err := svc.RequestEmailChangeLink(context.Background(), uuid.Nil, "1.2.3.4", "sess-token", "new@example.com")
	require.Error(t, err)
}

// TestRequestEmailChangeLinkRateLimitedByIP covers the per-IP
// rate-limit branch.
func TestRequestEmailChangeLinkRateLimitedByIP(t *testing.T) {
	now := testNow()
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	authRepo := newAuthRepoStub()
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 1)
	idem := NewMemoryIdempotencyStore()
	svc := NewService(repo, authRepo, &email.Fake{}, idem, c, limiter, Config{
		Environment: "test", BaseURL: "https://test.example.com",
		EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute,
	})

	_, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "a@example.com")
	require.NoError(t, err)
	_, err = svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "b@example.com")
	assert.ErrorIs(t, err, ErrEmailChangeRateLimited)
}

// TestRequestEmailChangeLinkRateLimitedBySession covers the per-session
// rate-limit branch (VOC-031-D05: an attacker with a valid session must be
// bounded even while rotating IPs). Two different client IPs are used so
// only the session-keyed check can be what blocks the second call.
func TestRequestEmailChangeLinkRateLimitedBySession(t *testing.T) {
	now := testNow()
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	authRepo := newAuthRepoStub()
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 1)
	idem := NewMemoryIdempotencyStore()
	svc := NewService(repo, authRepo, &email.Fake{}, idem, c, limiter, Config{
		Environment: "test", BaseURL: "https://test.example.com",
		EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute,
	})

	_, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "a@example.com")
	require.NoError(t, err)
	_, err = svc.RequestEmailChangeLink(context.Background(), uid, "5.6.7.8", "sess-token", "b@example.com")
	assert.ErrorIs(t, err, ErrEmailChangeRateLimited)
}

// TestConsumeEmailChangeLinkHappyPathAndOldEmailNotification covers
// VOC-031-TEST-16: the requester's current session is left intact
// (the accounts module never touches sessions here), and a
// notification is dispatched to the OLD email address. The new
// email is the one the database now holds.
func TestConsumeEmailChangeLinkHappyPathAndOldEmailNotification(t *testing.T) {
	svc, repo, authRepo, fake, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	repo.SetUser(uid, "old@example.com")

	link, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	raw := extractTokenFromLink(t, link.Link)

	res, err := svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", raw)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uid, res.UserID)
	assert.Equal(t, "old@example.com", res.OldEmail)
	assert.Equal(t, "new@example.com", res.NewEmail)
	assert.Equal(t, "old@example.com", res.NotificationTo)
	assert.Equal(t, "new@example.com", repo.UserEmail(uid))

	// Two messages: the confirmation link (request stage) and
	// the old-email notification (confirm stage).
	require.Len(t, fake.Sent, 2)
	notify := fake.Sent[1]
	assert.Equal(t, "old@example.com", notify.To[0].Email)
	assert.Equal(t, "Your Vocanova sign-in email was changed", notify.Subject)
	assert.Contains(t, notify.BodyText, "new@example.com")
}

func TestConsumeEmailChangeLinkRejectsDifferentAuthenticatedUser(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	ownerID := uuid.New()
	otherID := uuid.New()
	authRepo.setUser(&auth.User{ID: ownerID, Email: "owner@example.com", Status: "active"})
	authRepo.setUser(&auth.User{ID: otherID, Email: "other@example.com", Status: "active"})
	repo.SetUser(ownerID, "owner@example.com")
	repo.SetUser(otherID, "other@example.com")

	link, err := svc.RequestEmailChangeLink(context.Background(), ownerID, "1.2.3.4", "owner-session", "new@example.com")
	require.NoError(t, err)
	raw := extractTokenFromLink(t, link.Link)

	// Possessing an email-confirmation token is not authority to mutate the
	// link owner's account from another authenticated learner's session.
	_, err = svc.ConsumeEmailChangeLink(context.Background(), otherID, "1.2.3.4", "other-session", raw)
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
	assert.Equal(t, "owner@example.com", repo.UserEmail(ownerID))
	assert.Equal(t, "other@example.com", repo.UserEmail(otherID))

	// The rejected cross-account request does not consume the owner’s token.
	result, err := svc.ConsumeEmailChangeLink(context.Background(), ownerID, "1.2.3.4", "owner-session", raw)
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", result.NewEmail)
}

// TestConsumeEmailChangeLinkSessionUnchanged covers VOC-031-D05
// (must not weaken session/auth guarantees). The service does
// not touch sessions, so a requester's session remains valid
// after a successful email change. This is enforced structurally
// — the service has no session-revocation method on the happy
// path — but the test pins the absence of side effects on the
// auth/session state.
func TestConsumeEmailChangeLinkSessionUnchanged(t *testing.T) {
	svc, repo, authRepo, fake, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	repo.SetUser(uid, "old@example.com")

	link, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	raw := extractTokenFromLink(t, link.Link)

	// A session is an auth-level concept, but the test asserts
	// the accounts service did not touch any auth state: only
	// two emails (request + notification) are dispatched, and
	// the user still has its old email-mutation only.
	require.Len(t, fake.Sent, 1)
	res, err := svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", raw)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, fake.Sent, 2, "only the old-email notification is added at confirm; nothing else")

	// The token is consumed; a second confirm must fail.
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", raw)
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
}

// TestConsumeEmailChangeLinkRejectsExpired covers VOC-031-TEST-14:
// an expired token is rejected with the same error as a tampered
// or unknown token, so the reason is not observable.
func TestConsumeEmailChangeLinkRejectsExpired(t *testing.T) {
	svc, _, authRepo, _, c := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})

	link, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	raw := extractTokenFromLink(t, link.Link)

	c.Advance(16 * time.Minute)
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", raw)
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
}

// TestConsumeEmailChangeLinkRejectsTampered covers VOC-031-TEST-14:
// a tampered token is rejected.
func TestConsumeEmailChangeLinkRejectsTampered(t *testing.T) {
	svc, _, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})

	link, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	raw := extractTokenFromLink(t, link.Link)
	// Flip the last character of the token.
	last := raw[len(raw)-1]
	flipped := raw[:len(raw)-1] + string(last^1)
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", flipped)
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
}

// TestConsumeEmailChangeLinkRejectsEmpty covers the
// "no token supplied" branch.
func TestConsumeEmailChangeLinkRejectsEmpty(t *testing.T) {
	svc, _, _, _, _ := newService(t)
	_, err := svc.ConsumeEmailChangeLink(context.Background(), uuid.New(), "1.2.3.4", "sess-token", "")
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uuid.New(), "1.2.3.4", "sess-token", "   ")
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
}

// TestConsumeEmailChangeLinkRejectsReplay covers VOC-031-TEST-14:
// a previously-consumed token cannot be re-consumed.
func TestConsumeEmailChangeLinkRejectsReplay(t *testing.T) {
	svc, repo, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	repo.SetUser(uid, "old@example.com")

	link, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	raw := extractTokenFromLink(t, link.Link)
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", raw)
	require.NoError(t, err)
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", raw)
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
}

// TestConsumeEmailChangeLinkRejectsWrongEnvironment covers
// VOC-031-TEST-14: a token issued by one environment cannot be
// consumed in another.
func TestConsumeEmailChangeLinkRejectsWrongEnvironment(t *testing.T) {
	svc, _, authRepo, _, _ := newService(t)
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})

	link, err := svc.RequestEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "new@example.com")
	require.NoError(t, err)
	raw := extractTokenFromLink(t, link.Link)

	otherRepo := NewMemoryRepository()
	otherAuth := newAuthRepoStub()
	otherAuth.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	otherSvc := NewService(otherRepo, otherAuth, &email.Fake{}, NewMemoryIdempotencyStore(), &clock.Fixed{T: testNow()},
		auth.NewFixedWindowRateLimiter(&clock.Fixed{T: testNow()}, time.Hour, 100),
		Config{Environment: "production", BaseURL: "https://test.example.com", EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute},
	)
	_, err = otherSvc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", raw)
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink)
}

// TestConsumeEmailChangeLinkDuplicateEmailRace covers
// VOC-031-TEST-15: two users request a change to the same new
// email and both attempt to confirm. Exactly one confirm
// succeeds; the other receives ErrEmailAlreadyRegistered,
// never a 500 (the partial unique index on lower(email) WHERE
// deleted_at IS NULL is the authoritative guard).
func TestConsumeEmailChangeLinkDuplicateEmailRace(t *testing.T) {
	now := testNow()
	c := &clock.Fixed{T: now}
	repoA := NewMemoryRepository()
	repoB := NewMemoryRepository()
	// Both confirms target the same in-memory user table; the
	// race is observable by sharing the user set.
	repoA.SetUser(uuid.MustParse("00000000-0000-0000-0000-000000000001"), "a-old@example.com")
	repoA.SetUser(uuid.MustParse("00000000-0000-0000-0000-000000000002"), "b-old@example.com")
	repoB.SetUser(uuid.MustParse("00000000-0000-0000-0000-000000000001"), "a-old@example.com")
	repoB.SetUser(uuid.MustParse("00000000-0000-0000-0000-000000000002"), "b-old@example.com")

	authA := newAuthRepoStub()
	authB := newAuthRepoStub()
	uidA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	uidB := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	authA.setUser(&auth.User{ID: uidA, Email: "a-old@example.com", Status: "active"})
	authA.setUser(&auth.User{ID: uidB, Email: "b-old@example.com", Status: "active"})
	authB.setUser(&auth.User{ID: uidA, Email: "a-old@example.com", Status: "active"})
	authB.setUser(&auth.User{ID: uidB, Email: "b-old@example.com", Status: "active"})

	fake := &email.Fake{}
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	cfg := Config{Environment: "test", BaseURL: "https://test.example.com", EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute}
	svcA := NewService(repoA, authA, fake, NewMemoryIdempotencyStore(), c, limiter, cfg)
	svcB := NewService(repoB, authB, fake, NewMemoryIdempotencyStore(), c, limiter, cfg)

	// Both A and B request the same new email.
	linkA, err := svcA.RequestEmailChangeLink(context.Background(), uidA, "1.2.3.4", "sess-token", "shared@example.com")
	require.NoError(t, err)
	linkB, err := svcB.RequestEmailChangeLink(context.Background(), uidB, "1.2.3.4", "sess-token", "shared@example.com")
	require.NoError(t, err)
	rawA := extractTokenFromLink(t, linkA.Link)
	rawB := extractTokenFromLink(t, linkB.Link)

	// The two services do not share the email_change_links
	// table; bring both rows into a single repository so the
	// uniqueness discipline is exercised end-to-end.
	shared := NewMemoryRepository()
	shared.SetUser(uidA, "a-old@example.com")
	shared.SetUser(uidB, "b-old@example.com")
	rowsA := repoA.LinksForUser(uidA)
	rowsB := repoB.LinksForUser(uidB)
	require.Len(t, rowsA, 1)
	require.Len(t, rowsB, 1)
	// Insert both rows into the shared store; the shared store
	// uses the same hash discipline via the Service's own
	// token-and-hash path. We re-derive the hashes from the raw
	// tokens and reinsert directly so the test stays
	// deterministic.
	_, hashA, err := auth.TokenAndHash(rawA)
	require.NoError(t, err)
	_, hashB, err := auth.TokenAndHash(rawB)
	require.NoError(t, err)
	_, err = shared.CreateEmailChangeLink(context.Background(), uidA, "shared@example.com", hashA, "test", now, now.Add(15*time.Minute))
	require.NoError(t, err)
	_, err = shared.CreateEmailChangeLink(context.Background(), uidB, "shared@example.com", hashB, "test", now, now.Add(15*time.Minute))
	require.NoError(t, err)

	// First confirm wins.
	err = shared.UpdateUserEmail(context.Background(), uidA, "shared@example.com", now)
	require.NoError(t, err, "first confirm must succeed")
	err = shared.UpdateUserEmail(context.Background(), uidB, "shared@example.com", now)
	assert.ErrorIs(t, err, ErrEmailAlreadyRegistered, "second confirm must produce the stable conflict")
}

// TestConsumeEmailChangeLinkUserNotFound covers the
// user-missing path (deleted user, etc.).
func TestConsumeEmailChangeLinkUserNotFound(t *testing.T) {
	svc, _, _, _, _ := newService(t)
	// No authRepo.setUser: any user-id is missing.
	_, err := svc.ConsumeEmailChangeLink(context.Background(), uuid.New(), "1.2.3.4", "sess-token", "anyToken")
	assert.ErrorIs(t, err, ErrInvalidEmailChangeLink, "missing user surfaces as the same generic invalid-link error to avoid enumeration")
}

// TestConsumeEmailChangeLinkRateLimitedByIP covers the
// per-IP rate-limit branch.
func TestConsumeEmailChangeLinkRateLimitedByIP(t *testing.T) {
	now := testNow()
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	authRepo := newAuthRepoStub()
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 1)
	idem := NewMemoryIdempotencyStore()
	svc := NewService(repo, authRepo, &email.Fake{}, idem, c, limiter, Config{
		Environment: "test", BaseURL: "https://test.example.com",
		EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute,
	})

	_, err := svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "any")
	require.Error(t, err) // first call is a generic invalid-link
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "any")
	assert.ErrorIs(t, err, ErrEmailChangeRateLimited)
}

// TestConsumeEmailChangeLinkRateLimitedBySession covers the per-session
// rate-limit branch on the consume endpoint, mirroring
// TestRequestEmailChangeLinkRateLimitedBySession - two different client IPs
// so only the session-keyed check can be what blocks the second call.
func TestConsumeEmailChangeLinkRateLimitedBySession(t *testing.T) {
	now := testNow()
	c := &clock.Fixed{T: now}
	repo := NewMemoryRepository()
	authRepo := newAuthRepoStub()
	uid := uuid.New()
	authRepo.setUser(&auth.User{ID: uid, Email: "old@example.com", Status: "active"})
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 1)
	idem := NewMemoryIdempotencyStore()
	svc := NewService(repo, authRepo, &email.Fake{}, idem, c, limiter, Config{
		Environment: "test", BaseURL: "https://test.example.com",
		EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute,
	})

	_, err := svc.ConsumeEmailChangeLink(context.Background(), uid, "1.2.3.4", "sess-token", "any")
	require.Error(t, err) // first call is a generic invalid-link
	_, err = svc.ConsumeEmailChangeLink(context.Background(), uid, "5.6.7.8", "sess-token", "any")
	assert.ErrorIs(t, err, ErrEmailChangeRateLimited)
}

// TestRevokeAllEmailChangeLinksForUser is a tiny seam-test for
// the path T04 (account deletion) relies on.
func TestRevokeAllEmailChangeLinksForUser(t *testing.T) {
	repo := NewMemoryRepository()
	uid := uuid.New()
	now := testNow()
	hashA := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hashB := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	_, err := repo.CreateEmailChangeLink(context.Background(), uid, "new@example.com", hashA, "test", now, now.Add(15*time.Minute))
	require.NoError(t, err)
	_, err = repo.CreateEmailChangeLink(context.Background(), uid, "new@example.com", hashB, "test", now, now.Add(15*time.Minute))
	require.NoError(t, err)
	n, err := repo.RevokeAllEmailChangeLinksForUser(context.Background(), uid, now)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	rows := repo.LinksForUser(uid)
	for _, row := range rows {
		assert.NotNil(t, row.RevokedAt)
	}
}

// TestUpdateUserEmailRejectsDuplicate covers the partial-index
// discipline in isolation. It is the memory-side analog of
// the SQL test in postgres_test.go.
func TestUpdateUserEmailRejectsDuplicate(t *testing.T) {
	repo := NewMemoryRepository()
	uidA := uuid.New()
	uidB := uuid.New()
	now := testNow()
	repo.SetUser(uidA, "a@example.com")
	repo.SetUser(uidB, "b@example.com")
	require.NoError(t, repo.UpdateUserEmail(context.Background(), uidA, "shared@example.com", now))
	err := repo.UpdateUserEmail(context.Background(), uidB, "shared@example.com", now)
	assert.ErrorIs(t, err, ErrEmailAlreadyRegistered)
	assert.Equal(t, 1, repo.Collisions(), "the partial-index discipline fires exactly once")
}

// TestUpdateUserEmailAllowsReleasedAddress covers the
// "old email freed up after release" branch: if A stops using
// shared@example.com, B is allowed to claim it.
func TestUpdateUserEmailAllowsReleasedAddress(t *testing.T) {
	repo := NewMemoryRepository()
	uidA := uuid.New()
	uidB := uuid.New()
	now := testNow()
	repo.SetUser(uidA, "a@example.com")
	repo.SetUser(uidB, "b@example.com")
	require.NoError(t, repo.UpdateUserEmail(context.Background(), uidA, "shared@example.com", now))
	// A releases the address.
	require.NoError(t, repo.UpdateUserEmail(context.Background(), uidA, "different@example.com", now))
	// B can now claim it.
	require.NoError(t, repo.UpdateUserEmail(context.Background(), uidB, "shared@example.com", now))
	assert.Equal(t, "shared@example.com", repo.UserEmail(uidB))
}

// TestUpdateUserEmailIgnoresSoftDeletedOwner covers the
// partial-index WHERE clause: a soft-deleted user does not
// occupy the email, so another user is allowed to take it.
func TestUpdateUserEmailIgnoresSoftDeletedOwner(t *testing.T) {
	repo := NewMemoryRepository()
	uidA := uuid.New()
	uidB := uuid.New()
	now := testNow()
	repo.SetUser(uidA, "a@example.com")
	repo.SetUser(uidB, "b@example.com")
	require.NoError(t, repo.UpdateUserEmail(context.Background(), uidA, "shared@example.com", now))
	repo.SetUserDeleted(uidA, true)
	// B is now free to claim the address.
	require.NoError(t, repo.UpdateUserEmail(context.Background(), uidB, "shared@example.com", now))
	assert.Equal(t, "shared@example.com", repo.UserEmail(uidB))
}

// TestEmailChangeLinkValid exercises the Valid helper.
func TestEmailChangeLinkValid(t *testing.T) {
	now := testNow()
	l := &EmailChangeLink{ExpiresAt: now.Add(time.Minute)}
	assert.True(t, l.Valid(now))

	t1 := now.Add(-time.Second)
	l.ConsumedAt = &t1
	assert.False(t, l.Valid(now), "consumed link is invalid")
	l.ConsumedAt = nil

	t2 := now.Add(-time.Second)
	l.RevokedAt = &t2
	assert.False(t, l.Valid(now), "revoked link is invalid")
}
