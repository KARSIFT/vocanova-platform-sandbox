# VOC-045 — Impact Analysis

## Security and privacy

- Authentication surface: this package's confirmed root cause and proposed
  fix both live in the exact code path (`AuthURL`, `OAuthStart`) that starts
  every real user's Google sign-in. Any change here needs the same
  scrutiny as any other authentication-path change, per
  `docs/governance/change-risk-classification.md`'s R3 floor for this area.
- CSRF protection: the existing `state`/`cookieState` comparison in
  `OAuthCallback` (`apps/api/business/auth/service.go` lines 256-259) is
  unrelated to the `prompt` parameter and must remain unchanged by any task
  in this package - see `acceptance-criteria.md`'s `VOC-045-AC-03`.
- Consent semantics: dropping or conditioning `prompt=consent` means Google
  will, in some cases, silently re-authenticate and re-authorize a
  previously-consented user without displaying its own UI. This is the
  desired behavior change, not a new vulnerability, but it does shift more
  of the "does this user still intend to be signed in with this identity"
  question onto the app's own session validity rather than onto Google's
  UI reappearing. `specification.md`'s decisions section names this
  explicitly so it is not an unstated side effect.
- No secrets, credentials, or new personal-data collection are introduced.
  The change is to which query parameters are sent to Google's existing
  authorization endpoint, not to what data is collected or stored.

## Data and migrations

None. Confirmed during drafting by reading both affected files
(`apps/api/business/auth/google_oauth.go`,
`apps/api/business/auth/service.go`) - neither introduces or requires any
schema, table, or stored-data change. `OAuthStart`'s existing
`CreateOAuthState` call is unaffected by this package's scope.

## Analytics and accessibility

- Analytics: None currently proposed. `specification.md`'s "Data,
  migrations, analytics, and accessibility" section notes that if
  `VOC-045-T00`'s investigation surfaces a future need to distinguish
  re-prompt causes in production telemetry, that would be a separate
  proposal, not assumed here.
- Accessibility: Not applicable, evidence-backed. This is a backend
  URL-construction and session-validation concern confirmed (by reading the
  affected files) to touch no rendered UI, markup, ARIA attribute, or
  keyboard/interaction pattern in `apps/web`.

## Risks, dependencies, and evidence

- `VOC-045-R00`: The app-level session may itself have an undiscovered
  defect that this package's `T00` investigation has not yet run. If so,
  fixing only `AuthURL`'s `prompt` behavior would be an incomplete fix
  that still leaves users re-prompted. Mitigated by scoping `T00`
  (investigation) as a precondition to `T01` (fix), not bundling them.
- `VOC-045-R01`: Removing or conditioning `prompt=consent` could regress the
  original T15 goal of surfacing a fresh `email_verified` status on every
  sign-in, if the replacement condition (open question 2) is chosen
  carelessly. Mitigated by `acceptance-criteria.md`'s `VOC-045-AC-04`
  requiring the trade-off to be explicitly documented, not silently
  dropped.
- `VOC-045-R02`: Because this reverses a previously deliberate, documented
  design decision rather than fixing an accidental bug, there is a risk the
  original reasoning for that decision is no longer fully understood by
  whoever implements the fix. Mitigated by requiring `VOC-045-T01`'s
  implementer to read and explicitly address the existing code comment
  (`google_oauth.go` lines 192-203) rather than deleting it without
  comment.
- `VOC-045-DEP-00`: See `change.yaml` - the root-cause finding could be
  incomplete if a future issue-#343 comment reports additional symptoms.
- `VOC-045-DEP-01`: See `change.yaml` - whether this needs explicit
  founder/product sign-off beyond routine R3 review is unresolved at
  drafting time.
- `VOC-045-EV-00`: `VOC-045-T00`'s investigation findings (documented in its
  own pull request), confirming or ruling out an app-level session defect.
- `VOC-045-EV-01`: `VOC-045-T01`'s implementing pull request, including its
  test results and its explicit resolution of open question 2.
