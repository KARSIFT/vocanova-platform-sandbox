# VOC-045 — Test Plan

## VOC-045-TEST-00 — App-level session persistence, isolated from Google's prompt behavior

- Covers: `VOC-045-AC-00`
- Preconditions: A test/staging environment with OAuth configured (or the
  existing fake OAuth provider used in `apps/api/business/auth` tests), and
  the ability to inspect issued session cookies and their expiry.
- Procedure: Sign in once. Simulate a page reload and, separately, a
  browser restart (new process, same persisted cookie store) without
  signing out. Inspect whether the app's own session cookie is still valid
  and whether the app treats the request as authenticated, independent of
  any Google-side interaction.
- Expected result: The app-level session remains valid and authenticated
  until `SESSION_LIFETIME` elapses or explicit sign-out occurs, regardless
  of Google's own prompt behavior. Document the actual result (pass or
  fail) as this task's finding either way.
- Evidence: `VOC-045-EV-00`

## VOC-045-TEST-01 — No re-prompt for an already-authorized, currently-signed-in user

- Covers: `VOC-045-AC-01`
- Preconditions: `VOC-045-T01` implemented per the chosen replacement
  condition; a test double or real Google OAuth sandbox identity that has
  previously completed sign-in.
- Procedure: Complete a first sign-in. Without signing out, trigger a
  second sign-in attempt (e.g. revisiting a sign-in-gated route). Inspect
  the generated authorization URL and/or the observed flow.
- Expected result: The generated authorization URL does not force Google's
  account-selection/consent UI to display for this case; the user reaches
  the authenticated state without seeing that screen again.
- Evidence: `VOC-045-EV-01`

## VOC-045-TEST-02 — Explicit sign-out still requires a genuine OAuth round-trip

- Covers: `VOC-045-AC-02`
- Preconditions: `VOC-045-T01` implemented; a previously signed-in test
  session.
- Procedure: Sign out explicitly. Attempt to sign in again. Inspect whether
  a real authorization code exchange occurs with the OAuth provider (real
  or fake, per the existing test harness).
- Expected result: The app performs a genuine OAuth code exchange again
  after explicit sign-out; it does not silently restore the prior session
  without re-authenticating with Google.
- Evidence: `VOC-045-EV-01`

## VOC-045-TEST-03 — CSRF state-token check is unaffected

- Covers: `VOC-045-AC-03`
- Preconditions: `VOC-045-T01` implemented.
- Procedure: Run (or add, if missing) a negative test asserting that
  `OAuthCallback` rejects a callback whose `state` does not match the
  `cookieState` value, unchanged by this package's modifications.
- Expected result: The existing rejection behavior in
  `apps/api/business/auth/service.go` (state/cookieState mismatch returns
  `ErrInvalidOAuthState`) is unchanged and still covered by a passing test.
- Evidence: `VOC-045-EV-01`

## VOC-045-TEST-04 — `email_verified` freshness trade-off is explicit, not silently dropped

- Covers: `VOC-045-AC-04`
- Preconditions: `VOC-045-T01` implemented.
- Procedure: Review the implementing pull request's description for an
  explicit statement of how the chosen replacement condition addresses (or
  consciously accepts a trade-off against) the original
  `email_verified`-freshness reasoning documented in `AuthURL`'s prior code
  comment.
- Expected result: The pull request description states the trade-off
  explicitly; it is not silently omitted from the change's rationale.
- Evidence: `VOC-045-EV-01`

Positive coverage (`TEST-01`), negative/regression coverage (`TEST-02`,
`TEST-03`), and documentation/rationale coverage (`TEST-04`) are included.
No migration or accessibility coverage applies, per `impact-analysis.md`'s
determination. No test in this plan uses secrets or production data; all
OAuth interactions are covered via the repository's existing fake OAuth
provider or a disposable sandbox identity.
