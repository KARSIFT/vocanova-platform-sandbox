# VOC-045 — Google Sign-In Re-Prompts for Account Selection/Consent on Every Sign-In

**Status: proposed, not adopted.** Nothing in this package is
implementation-authorized. It is a draft response to
[issue #343](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/343),
prepared for founder/steward review at adoption time.

## Why this exists

Issue #343 reports founder feedback from VOC-038-T03's live core-loop
validation on 2026-08-07: every time signing in with Google, the account
picker/consent screen appears again, even when there is no reason to expect
it (not signed out, same browser/session). The founder is not blocked by
this and explicitly asked for it to be recorded for future work, not fixed
immediately - this package is that record, plus a scoped investigation and
fix proposal.

## What was found during drafting

Reading `apps/api/business/auth/google_oauth.go`'s `AuthURL` method (the
function that actually builds the Google authorization URL) during drafting
confirms one half of the issue's own hypothesis directly, not just as a
guess:

```213:222:apps/api/business/auth/google_oauth.go
	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", scopes)
	q.Set("state", state)
	q.Set("access_type", "online")
	q.Set("prompt", "consent")
	q.Set("include_granted_scopes", "true")
	return endpoint + "?" + q.Encode()
```

`prompt=consent` is set unconditionally, on every call, for every sign-in
attempt - not only on first-ever authorization for a given user, and not
conditioned on whether the app already has a valid session for that
browser. The function's own doc comment names this as intentional:

```192:203:apps/api/business/auth/google_oauth.go
// AuthURL builds the Google authorization URL for the supplied
// state token and redirect URI. The state is included verbatim
// (URL-encoded) so the auth service can match it against the
// state cookie on callback; the redirect URI is included exactly
// as supplied so the value registered in Google Cloud Console
// matches the value the auth service sends. prompt=consent is
// requested so the user sees the consent screen on every sign-in
// (and Google therefore returns a fresh id_token whose
// email_verified reflects the user's current verified status);
// access_type=online is requested so Google does not return a
// refresh token, which T15 has no use for and which the auth
// service does not persist.
```

So this is not an accidental defect - it is a deliberate prior design
decision (attributed to task "T15" in the comment) trading off "always see
a fresh, verified `email_verified` status from Google" against "never
re-prompt an already-recognized user." The issue's acceptance shape asks for
exactly that trade-off to move the other way: after a real sign-in,
revisiting the app without signing out should not require Google's account
picker again until either the app's own session expires or the user
explicitly signs out.

This package does **not** independently confirm or rule out the issue's
other named hypothesis - whether the app's own session
(`SESSION_LIFETIME=720h` per `apps/api/.env.example`) is itself persisting
correctly across page loads/browser restarts, separate from Google's
prompt behavior. That requires runtime/browser-level observation this
drafting pass did not perform; `specification.md`'s open questions and
`tasks.md`'s `VOC-045-T00` scope that confirmation as the first task, before
any fix task changes `AuthURL`'s behavior.

## What this package deliberately does NOT do

- It does not assume the app-level session is broken. The issue's own
  acceptance shape treats "our own app forgot who the user is" and
  "Google's own prompt re-appears even though our session is fine" as two
  distinct possible causes; this package's `T00` investigates which one (or
  both) is actually true before `T01` changes any code.
- It does not remove `email_verified` freshness entirely as a goal. Any fix
  task must reconcile "don't force `prompt=consent` on every sign-in" with
  the original comment's stated reason for requiring it, not silently drop
  that concern - see `specification.md`'s open questions.
- It does not touch `apps/web`'s own client-side sign-in flow, cookie
  handling, or any other unrelated authentication surface. The confirmed
  root-cause candidate is entirely server-side, in the URL-construction
  function itself.
- It does not touch `access_type=online` or the scopes requested. The issue
  is specifically about the re-prompt behavior, not the token/scope shape.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched
  until a real adoption decision is recorded.

## Open questions flagged for the reviewing human

`specification.md`'s "Open questions" section flags: (1) whether removing
the unconditional `prompt=consent` needs explicit founder/product sign-off,
given it reverses a previously deliberate, documented decision rather than
fixing an unintentional bug; (2) what the correct replacement condition
should be (e.g. omit `prompt` entirely and rely on Google's own default
behavior, versus only set it on a user's first-ever OAuth authorization,
versus some other signal); and (3) whether the app-level session itself
needs any change at all, pending `VOC-045-T00`'s investigation.

## Structure

Mirrors recent packages' convention (e.g. VOC-044, VOC-043, VOC-042):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` risk classification in `change.yaml` (this
   touches authentication/authorization: the Google OAuth authorization URL
   every sign-in attempt depends on).
2. Read `specification.md`'s open questions and decide whether reversing the
   prior `prompt=consent` design decision needs explicit sign-off beyond
   routine R3 review.
3. Decide the replacement condition for when `prompt` should (and should
   not) be requested, or whether that decision should be left to
   `VOC-045-T01`'s implementer to propose for review.
4. Adopt (or request changes to) this package, then dispatch `VOC-045-T00`
   (investigation) before `VOC-045-T01` (fix), consistent with this
   package's own scoping.
