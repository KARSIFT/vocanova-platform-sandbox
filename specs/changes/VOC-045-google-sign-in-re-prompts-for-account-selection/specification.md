# VOC-045 — Google Sign-In Re-Prompts for Account Selection/Consent on Every Sign-In: Specification

## Objective and requirement source

Stop Google's account-selection/consent screen from reappearing on every
sign-in when the user already has (or should have) an active session, per
[issue #343](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/343).
The issue is founder-reported, not urgent, and explicitly requested to be
recorded for future work rather than fixed immediately - this specification
scopes that future work, not an emergency patch.

Acceptance shape, quoted from the issue: after a real sign-in, revisiting
the app (without signing out) should not require going through Google's
account picker again until either the app's own session expires or the user
explicitly signs out.

## Confirmed finding (made during drafting, not assumed)

`apps/api/business/auth/google_oauth.go`'s `AuthURL` method (lines 204-223)
unconditionally sets `prompt=consent` on every generated Google authorization
URL, for every call, regardless of whether the requesting browser already
has a valid session or has ever authorized this app before. The method's own
doc comment (lines 192-203) states this is intentional: it exists so Google
returns a fresh `id_token` whose `email_verified` claim reflects the user's
*current* verified status, at the cost of showing the consent/account-picker
screen every time.

`apps/api/business/auth/service.go`'s `OAuthStart` (lines 205-236) calls
`s.oauth.AuthURL(token, s.cfg.OAuthRedirectURI)` with no additional
conditioning - it does not check for an existing valid session, existing
prior authorization for this Google identity, or any other signal before
delegating to `AuthURL`. So the unconditional-`prompt=consent` behavior in
`AuthURL` is the entire mechanism, not one of several contributing factors
discovered so far.

This is a **directly confirmed** finding (the code was read during
drafting), not an inference from the issue text alone - unlike VOC-044's
root-cause finding, which relied on the issue's own reported reproduction
evidence.

## What remains unconfirmed

This drafting pass did not runtime-verify the issue's second named
hypothesis: whether the app's own session
(`SESSION_LIFETIME=720h`, `SESSION_COOKIE_SECURE=true`,
`SESSION_COOKIE_DOMAIN=.vocanova.site` per `apps/api/.env.example`) actually
persists correctly across page loads and browser restarts. It is plausible
that the app-level session is entirely correct and only Google's own prompt
is the visible symptom (in which case the fix is scoped entirely to
`AuthURL`), but this specification does not assert that without the
investigation task (`VOC-045-T00`) confirming it first.

## Scope and non-goals

In scope:
- Investigating (not assuming) whether the app-level session is itself
  working as intended, isolated from Google's own prompt behavior.
- Changing when `prompt=consent` (and/or `access_type`) is requested in
  `AuthURL`/`OAuthStart`, so that a user with a currently-valid app session,
  or a user who has already granted consent for this app before, is not
  forced through Google's account picker again.
- Preserving some workable way to obtain a reasonably fresh
  `email_verified` status without forcing the screen on literally every
  sign-in (see open question 2 for the exact mechanism, left open for the
  reviewing human and/or `VOC-045-T01`'s implementer to settle).

Out of scope (see `README.md`'s "What this package deliberately does NOT
do" for the complete list):
- Any change to `apps/web`'s client-side sign-in flow or cookie handling.
- Any change to OAuth scopes or `access_type=online`.
- Any change to non-Google authentication (magic link).
- Any production deployment or rollout decision - this package proposes a
  fix; it does not authorize releasing it.

## Risk and protected areas

Builder assessment: the confirmed root cause and its fix both live in
`apps/api/business/auth/google_oauth.go`'s `AuthURL` and
`apps/api/business/auth/service.go`'s `OAuthStart` - the exact code path
that starts every real user's Google sign-in. This is squarely inside the
authentication/authorization protected area
`docs/governance/change-risk-classification.md` names as an R3 floor.
Because the change reverses a previously deliberate, documented design
choice (not merely fixes an obviously accidental bug), the reviewing human
may reasonably decide it warrants more than routine R3 scrutiny - this
specification does not attempt to make that call itself (see open question
1).

## Decisions, contradictions, security, and privacy

No `VOC-045-D0x` decisions are defined here; none may be defined before
adoption. Recording the constraints a fix must satisfy, for the reviewing
human and/or implementer to turn into an actual decision at or after
adoption:

- The fix must not weaken the state-token CSRF protection already present
  in `OAuthStart`/`OAuthCallback` (the `state`/`cookieState` comparison at
  `service.go` lines 256-259). Nothing about changing the `prompt` parameter
  requires touching that mechanism, and no task in this package may do so.
- The fix must not silently grant a session to a browser that never
  actually completed a Google OAuth round-trip. "Don't re-prompt" must mean
  "don't force Google's UI to reappear," not "skip verifying the user's
  identity with Google."
- If `prompt=consent` is dropped or made conditional, Google's default
  behavior is to silently re-authenticate and re-authorize a previously
  consented user without any prompt (assuming no scope changes) - this is
  the desired outcome, not a security regression, but it does mean Google's
  own consent history for this app becomes a component of the security
  model that did not need consideration before. No task may assume this
  without stating it.
- No production secrets, credentials, or personal data are introduced by
  this specification. This is a URL-construction and/or session-check
  change; it does not add new data collection or storage.

## Data, migrations, analytics, and accessibility

- Data/migrations: None. This specification proposes no schema, table, or
  stored-data change. Confirmed by inspecting the affected files - neither
  `google_oauth.go` nor `service.go`'s `OAuthStart` reads or writes any
  database table beyond `CreateOAuthState`'s existing, unaffected call.
- Analytics: None currently proposed. If `VOC-045-T00`'s investigation finds
  a need to distinguish "re-prompted due to no prior consent" from
  "re-prompted due to a bug" in production going forward, that would be a
  new, separate proposal, not something this specification assumes.
- Accessibility: Not applicable. This is a backend URL-construction and
  session-validation concern; it does not change any rendered UI, markup,
  or interaction pattern in `apps/web`.

## Open questions

1. Does reversing `AuthURL`'s existing, documented `prompt=consent`
   decision (attributed to task "T15" in the code comment) require explicit
   founder/product sign-off beyond routine R3 review, since it changes a
   previously deliberate trade-off rather than fixing an unintentional
   defect? This specification does not resolve that; it is for the
   reviewing human to decide at adoption time.
2. What should the replacement condition for `prompt` actually be? Candidate
   options, none selected here:
   - Omit `prompt` entirely on every call and rely on Google's own default
     silent-reauthentication behavior for a user with an existing valid
     Google session and prior consent.
   - Only set `prompt=consent` (or `prompt=select_account`) the first time a
     given identity authorizes this app, tracked via existing identity
     records, and omit it on subsequent sign-ins.
   - Keep `prompt=consent` only when the app cannot find a currently-valid
     session for the requesting browser (i.e., only actually needed for a
     fresh sign-in, not a same-session repeat), and omit it otherwise.
   This specification records the trade-off (freshness of `email_verified`
   versus re-prompt frequency) without picking a resolution; that decision
   belongs to `VOC-045-T01`'s implementer, subject to review, or to the
   adopting human up front.
3. Is the app-level session itself confirmed working as intended, separate
   from Google's prompt behavior? `VOC-045-T00` must answer this before
   `VOC-045-T01` proceeds; if the app-level session turns out to also be
   broken, `VOC-045-T01`'s scope must be revised to address that too, not
   just `AuthURL`.
