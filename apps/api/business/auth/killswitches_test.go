package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestKillSwitchesDefaultToEnabled covers the "no switch
// installed" path: when the service has a nil *KillSwitches
// (the default for any service constructed via NewService
// without a follow-up SetKillSwitches call, e.g. the existing
// NewContractAPI / test wiring), every gated path remains open
// exactly as it was before T00. The kill switches are additive,
// not breaking.
func TestKillSwitchesDefaultToEnabled(t *testing.T) {
	svc, _, _, _ := testService(t)
	assert.Nil(t, svc.KillSwitches())

	email := "user@example.com"
	err := svc.RequestMagicLink(context.Background(), "127.0.0.1", email)
	assert.NoError(t, err, "magic link should be open with no kill switches")
}

// TestKillSwitches_MagicLinkDisabled covers the
// EMAIL_MAGIC_LINK_ENABLED=false path: RequestMagicLink must
// short-circuit with ErrMagicLinkDisabled before any rate-limit
// counter is consumed, so a disabled path can never be re-enabled
// mid-attack by simply flipping the switch on. The implementation
// is in service.go's RequestMagicLink and ConsumeMagicLink.
func TestKillSwitches_MagicLinkDisabled(t *testing.T) {
	svc, _, _, _ := testService(t)
	svc.SetKillSwitches(&KillSwitches{MagicLinkEnabled: false})

	err := svc.RequestMagicLink(context.Background(), "127.0.0.1", "user@example.com")
	assert.ErrorIs(t, err, ErrMagicLinkDisabled)
}

// TestKillSwitches_OAuthDisabled covers the
// GOOGLE_OAUTH_ENABLED=false path: OAuthStart must return
// ErrOAuthDisabled (not ErrOAuthNotConfigured) so the HTTP layer
// can distinguish "founder turned this off" from "no provider
// wired in this environment".
func TestKillSwitches_OAuthDisabled(t *testing.T) {
	svc, _, _, _ := testServiceWithOAuth(t, NewFakeOAuthProvider(&OAuthIdentity{
		Subject: "s", Email: "e@example.com", EmailVerified: true,
	}))
	svc.SetKillSwitches(&KillSwitches{OAuthEnabled: false})

	_, _, err := svc.OAuthStart(context.Background(), "127.0.0.1", "https://test.example.com/app")
	assert.ErrorIs(t, err, ErrOAuthDisabled)
}

// TestKillSwitches_NewSignupsDisabled covers the
// NEW_USER_SIGNUP_ENABLED=false path on the magic-link consume
// side: an existing user can still sign in (their row already
// exists), but a never-before-seen email must be rejected with
// ErrSignupsDisabled. This is the "founder wants to allow
// existing users in but freeze new account creation" posture
// DOC-11 §3 explicitly calls out. Magic-link and OAuth are kept
// on explicitly so the test isolates the new-signup gate.
func TestKillSwitches_NewSignupsDisabled(t *testing.T) {
	svc, _, fake, _ := testService(t)
	svc.SetKillSwitches(&KillSwitches{
		MagicLinkEnabled:  true,
		OAuthEnabled:      true,
		NewSignupsEnabled: false,
	})

	// Request a magic link, then attempt to consume it with a
	// never-before-seen email. The consume path is where the
	// new-user create happens, so this is the right place to
	// assert.
	if err := svc.RequestMagicLink(context.Background(), "127.0.0.1", "newcomer@example.com"); err != nil {
		t.Fatalf("RequestMagicLink: %v", err)
	}
	msg, ok := fake.Last()
	if !ok {
		t.Fatalf("expected a sent email after RequestMagicLink")
	}
	rawToken := extractTokenFromURL(t, msg.BodyText)
	_, _, _, err := svc.ConsumeMagicLink(context.Background(), "127.0.0.1", rawToken, "newcomer@example.com")
	assert.ErrorIs(t, err, ErrSignupsDisabled)
}

// TestKillSwitches_SignupAllowlistAdmitsListedEmail covers the
// VOC-038-D00/D01 path: NEW_USER_SIGNUP_ENABLED=false still blocks
// new signups in general, but a never-before-seen email present in
// SignupAllowlist must be admitted anyway (case-insensitively).
func TestKillSwitches_SignupAllowlistAdmitsListedEmail(t *testing.T) {
	svc, _, fake, _ := testService(t)
	svc.SetKillSwitches(&KillSwitches{
		MagicLinkEnabled:  true,
		OAuthEnabled:      true,
		NewSignupsEnabled: false,
		SignupAllowlist:   map[string]struct{}{"founder@example.com": {}},
	})

	if err := svc.RequestMagicLink(context.Background(), "127.0.0.1", "Founder@Example.com"); err != nil {
		t.Fatalf("RequestMagicLink: %v", err)
	}
	msg, ok := fake.Last()
	if !ok {
		t.Fatalf("expected a sent email after RequestMagicLink")
	}
	rawToken := extractTokenFromURL(t, msg.BodyText)
	_, _, _, err := svc.ConsumeMagicLink(context.Background(), "127.0.0.1", rawToken, "Founder@Example.com")
	assert.NoError(t, err, "allowlisted email must be admitted even while signups are otherwise disabled")
}

// TestKillSwitches_SignupAllowlistIgnoresUnlistedEmail confirms the
// allowlist is exclusionary: a non-allowlisted email still gets
// ErrSignupsDisabled while NEW_USER_SIGNUP_ENABLED=false, so the
// allowlist can never widen access beyond the named cohort.
func TestKillSwitches_SignupAllowlistIgnoresUnlistedEmail(t *testing.T) {
	svc, _, fake, _ := testService(t)
	svc.SetKillSwitches(&KillSwitches{
		MagicLinkEnabled:  true,
		OAuthEnabled:      true,
		NewSignupsEnabled: false,
		SignupAllowlist:   map[string]struct{}{"founder@example.com": {}},
	})

	if err := svc.RequestMagicLink(context.Background(), "127.0.0.1", "outsider@example.com"); err != nil {
		t.Fatalf("RequestMagicLink: %v", err)
	}
	msg, ok := fake.Last()
	if !ok {
		t.Fatalf("expected a sent email after RequestMagicLink")
	}
	rawToken := extractTokenFromURL(t, msg.BodyText)
	_, _, _, err := svc.ConsumeMagicLink(context.Background(), "127.0.0.1", rawToken, "outsider@example.com")
	assert.ErrorIs(t, err, ErrSignupsDisabled)
}

// TestKillSwitches_SignupAllowlistIgnoredWhenSignupsEnabled ensures
// the allowlist has no effect once the blanket switch is open: it
// must not accidentally start rejecting non-listed emails.
func TestKillSwitches_SignupAllowlistIgnoredWhenSignupsEnabled(t *testing.T) {
	svc, _, fake, _ := testService(t)
	svc.SetKillSwitches(&KillSwitches{
		MagicLinkEnabled:  true,
		OAuthEnabled:      true,
		NewSignupsEnabled: true,
		SignupAllowlist:   map[string]struct{}{"founder@example.com": {}},
	})

	if err := svc.RequestMagicLink(context.Background(), "127.0.0.1", "anyone@example.com"); err != nil {
		t.Fatalf("RequestMagicLink: %v", err)
	}
	msg, ok := fake.Last()
	if !ok {
		t.Fatalf("expected a sent email after RequestMagicLink")
	}
	rawToken := extractTokenFromURL(t, msg.BodyText)
	_, _, _, err := svc.ConsumeMagicLink(context.Background(), "127.0.0.1", rawToken, "anyone@example.com")
	assert.NoError(t, err, "allowlist must not narrow access when signups are broadly enabled")
}

// TestKillSwitches_AllEnabledWhenExplicit guards against an
// accidental future regression where a service is constructed
// with a non-nil *KillSwitches that someone forgot to populate:
// the production wiring must pass all three booleans explicitly
// (true or false), and the service must honor whatever it
// receives. Setting every flag to true must keep every path
// open.
func TestKillSwitches_AllEnabledWhenExplicit(t *testing.T) {
	svc, _, _, _ := testService(t)
	svc.SetKillSwitches(&KillSwitches{
		MagicLinkEnabled:  true,
		OAuthEnabled:      true,
		NewSignupsEnabled: true,
	})

	err := svc.RequestMagicLink(context.Background(), "127.0.0.1", "user@example.com")
	assert.NoError(t, err)
}

// TestSetKillSwitches_NilClears confirms SetKillSwitches(nil)
// restores the "all enabled" default. A production rollback that
// wants to re-enable a path does not need to construct a fresh
// KillSwitches struct with all-true booleans - passing nil
// suffices and matches the pre-T00 default.
func TestSetKillSwitches_NilClears(t *testing.T) {
	svc, _, _, _ := testService(t)
	svc.SetKillSwitches(&KillSwitches{MagicLinkEnabled: false})
	assert.NotNil(t, svc.KillSwitches())

	svc.SetKillSwitches(nil)
	assert.Nil(t, svc.KillSwitches())
}

// TestKillSwitches_DisabledErrorIsDistinct ensures each disabled
// path surfaces its own sentinel (rather than reusing
// ErrOAuthNotConfigured, for example) so the HTTP layer can map
// each to a distinct 503/disabled response without conflating
// "founder disabled this" with "this isn't wired in this env".
func TestKillSwitches_DisabledErrorIsDistinct(t *testing.T) {
	for _, tc := range []struct {
		got, want error
	}{
		{ErrMagicLinkDisabled, ErrMagicLinkDisabled},
		{ErrOAuthDisabled, ErrOAuthDisabled},
		{ErrSignupsDisabled, ErrSignupsDisabled},
	} {
		assert.True(t, errors.Is(tc.got, tc.want), "%v should match itself", tc.want)
		assert.False(t, errors.Is(tc.got, ErrOAuthNotConfigured), "%v must not collide with ErrOAuthNotConfigured", tc.want)
	}
}
