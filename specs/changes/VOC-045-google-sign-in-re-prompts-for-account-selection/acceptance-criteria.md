# VOC-045 — Acceptance Criteria

## VOC-045-AC-00 — App-level session behavior is confirmed, not assumed

- Requirement source: `specification.md`'s "What remains unconfirmed" section
- Tasks: `VOC-045-T00`
- Tests: `VOC-045-TEST-00`
- Evidence: `VOC-045-EV-00`
- Result: pending

The investigation records, with reproducible evidence, whether the app's own
session (`SESSION_LIFETIME=720h` and related cookie config) persists across
page loads and browser restarts as intended, independent of Google's own
OAuth prompt behavior. This criterion is satisfied by a documented finding
either way, not only by finding "no bug."

## VOC-045-AC-01 — Google's account picker/consent screen does not reappear for an already-authorized, already-signed-in user

- Requirement source: `specification.md`'s objective (issue #343's acceptance
  shape)
- Tasks: `VOC-045-T01`
- Tests: `VOC-045-TEST-01`
- Evidence: `VOC-045-EV-01`
- Result: pending

After a user has completed a real Google sign-in once, a second sign-in
attempt initiated while their app session is still valid (or, at minimum,
while Google itself still recognizes prior consent for this app and this
identity) does not force Google's account-selection or consent UI to
display. The user reaches the app's authenticated state without seeing that
screen again.

## VOC-045-AC-02 — Explicit sign-out still requires a fresh Google sign-in

- Requirement source: `specification.md`'s objective (issue #343's acceptance
  shape: "...until either the app's own session expires or the user
  explicitly signs out")
- Tasks: `VOC-045-T01`
- Tests: `VOC-045-TEST-02`
- Evidence: `VOC-045-EV-01`
- Result: pending

After a user explicitly signs out of the app, a subsequent sign-in attempt
still performs a real OAuth round-trip with Google (this criterion does not
require Google's own UI to reappear if Google itself still recognizes the
user's browser session - it requires that the app does not skip
authentication with Google altogether after sign-out).

## VOC-045-AC-03 — CSRF state-token protection is unchanged

- Requirement source: `specification.md`'s "Decisions, contradictions,
  security, and privacy" section
- Tasks: `VOC-045-T01`
- Tests: `VOC-045-TEST-03`
- Evidence: `VOC-045-EV-01`
- Result: pending

The existing `state`/`cookieState` comparison in `OAuthCallback` continues to
reject a callback whose state does not match the value the app itself
generated and stored, unchanged by any task in this package.

## VOC-045-AC-04 — `email_verified` freshness trade-off is explicitly resolved, not silently dropped

- Requirement source: `specification.md`'s open question 2
- Tasks: `VOC-045-T01`
- Tests: `VOC-045-TEST-04`
- Evidence: `VOC-045-EV-01`
- Result: pending

The fix's chosen replacement condition for when `prompt` is (or is not) set
is documented in the implementing pull request, including how it addresses
(or consciously accepts a trade-off against) the original comment's reason
for requiring `prompt=consent` on every sign-in.

Acceptance criteria remain observable, stable, security-aware, and
bidirectionally traceable to requirements, tasks, tests, and evidence. None
of these criteria may be marked satisfied until this package is adopted and
each task is actually implemented.
