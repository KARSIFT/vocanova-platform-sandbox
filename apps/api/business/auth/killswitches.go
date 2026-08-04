package auth

import "errors"

// KillSwitches holds the DOC-11 §3 emergency-disable flags the
// production wiring installs on the auth service. They are
// intentionally separate from auth.Config so the existing
// NewContractAPI / test wiring (which never installs them) is
// unaffected - a Service with a nil KillSwitches is treated as
// "all paths enabled", matching the pre-T00 contract.
//
// A nil *KillSwitches on a Service is a no-op: every path
// remains open. A non-nil *KillSwitches with one or more false
// values disables only those paths. This shape lets the
// production-wiring function (T00) read the env-var state once
// and pass a single struct in, instead of threading three
// independent booleans through the constructor (which would
// change the constructor signature and break the existing
// contract / test wiring that calls NewService).
type KillSwitches struct {
	// MagicLinkEnabled gates POST /api/v1/auth/magic (request)
	// and the corresponding consume / callback. When false, the
	// service refuses to issue or consume any magic link.
	MagicLinkEnabled bool

	// OAuthEnabled gates GET /api/v1/auth/oauth/:provider/start
	// and the corresponding callback. When false, OAuthStart
	// returns ErrOAuthNotConfigured regardless of whether a
	// real provider is wired, so the start endpoint is closed
	// at the service layer (not by removing the route from the
	// mux).
	OAuthEnabled bool

	// NewSignupsEnabled gates the implicit "create a new user
	// on first verified sign-in" path inside ConsumeMagicLink
	// and the OAuth callback. When false, an existing user can
	// still sign in, but a never-before-seen email address is
	// rejected with ErrSignupsDisabled, unless it appears in
	// SignupAllowlist. Magic-link / OAuth themselves are still
	// gated by their own flags.
	NewSignupsEnabled bool

	// SignupAllowlist holds normalized (lowercased, trimmed) email
	// addresses admitted to sign up even while NewSignupsEnabled is
	// false, per VOC-038-D00/D01 (L1 controlled-launch cohort). It
	// has no effect when NewSignupsEnabled is true - the blanket
	// switch already admits everyone. A nil or empty map allowlists
	// no one, matching pre-VOC-038 behavior.
	SignupAllowlist map[string]struct{}
}

// signupAllowed reports whether a never-before-seen normalizedEmail
// may be signed up right now: either the blanket NewSignupsEnabled
// switch is open, or it's closed but normalizedEmail is present in
// the founder-maintained SignupAllowlist. Callers must pass an
// already-normalized email (see normalizeEmail).
func (sw *KillSwitches) signupAllowed(normalizedEmail string) bool {
	if sw == nil || sw.NewSignupsEnabled {
		return true
	}
	_, ok := sw.SignupAllowlist[normalizedEmail]
	return ok
}

// SetKillSwitches installs a kill-switch struct on the service.
// A nil value clears any previously installed switches, restoring
// the "everything enabled" default that contract wiring relies on.
// This is intentionally a setter, not a constructor parameter:
// the existing auth.NewService signature must stay stable for the
// T00-protected "never modify" guarantee on auth's existing
// token/session/rate-limit primitives (the production-wiring
// caller in app/api installs these switches once after the
// service is constructed).
func (s *Service) SetKillSwitches(sw *KillSwitches) {
	s.killSwitches = sw
}

// KillSwitches returns the currently installed kill switches, or
// nil if none have been installed. Useful for tests and for the
// /healthz handler's startup probe to confirm the wiring layer
// did install them.
func (s *Service) KillSwitches() *KillSwitches {
	return s.killSwitches
}

// ErrMagicLinkDisabled is returned by RequestMagicLink and
// ConsumeMagicLink when EMAIL_MAGIC_LINK_ENABLED is off.
var ErrMagicLinkDisabled = errors.New("magic link sign-in is disabled")

// ErrOAuthDisabled is returned by OAuthStart and the OAuth
// callback when GOOGLE_OAUTH_ENABLED is off. We surface a
// dedicated sentinel (rather than reusing ErrOAuthNotConfigured)
// so the HTTP layer can map the disable case to a distinct
// status / message without conflating it with the
// "real provider not wired" case.
var ErrOAuthDisabled = errors.New("google oauth sign-in is disabled")

// ErrSignupsDisabled is returned by ConsumeMagicLink and the
// OAuth callback when NEW_USER_SIGNUP_ENABLED is off and the
// verified identity is for a never-before-seen user.
var ErrSignupsDisabled = errors.New("new sign-ups are disabled")
