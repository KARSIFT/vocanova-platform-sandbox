# VocaNova maturity delivery — 2026-09-11

## Intent and context

Deliver a coherent learning and account experience on the existing Next.js/Go
product. Main was pulled before work; baseline is 5c756274. The founder authorizes
this post-MVP scope, including design and documentation changes. Ordinary PR/CI
and review workflow in AGENTS.md applies.

Must-see context: product bible/PRD, web design, current app routes and API client,
auth/session implementation, sentence history endpoints, local verification guide.
Should-see context: browser fixtures, accessibility matrix, deployment/synthetics.
Unknowns: live learner retention, provider feedback quality, recent restore proof,
and browser behavior on physical devices. Do not claim these from fixture tests.

## Product acceptance

- One calm visual system across entry, account and learning surfaces, accessible
  at 360px/430px and desktop with visible focus and reduced-motion support.
- Sign-in gives a useful check-email state, edit/resend actions, safe recovery and
  clear session-expiry/logout messages. Sessions remain server-authoritative.
- Learning interruptions preserve sentence drafts only within a bounded,
  authenticated-user-scoped tab lifetime. Explicit logout/deletion clears drafts.
- Sentence practice supports deliberate revision; original and feedback remain
  distinguishable. Network retries retain their idempotency identity.
- Progress exposes paginated retained sentence history, with honest processing
  states and useful empty/error/loading behavior. No unsupported mastery claims.
- Journey shows useful context and backend-confirmed saved state. Home has one
  clear next action and an unpressured completion state.
- Current docs explain the delivered behavior and remaining operational/product
  validation needs. Historical MVP exclusions remain historical, not blockers.

## Verification and release

Review diffs independently. Run relevant helpers, workspace checks, production
build, browser flows and accessibility at all supported layouts, Lighthouse, and
API checks. Local Docker is unavailable; database/OAuth tests must pass in CI.
Keep the PR draft until review and required checks are satisfied; request @claude
review, fix actionable findings, ready the PR, and observe normal auto-merge.
Production deployment is a separate manual action; do not imply a merge is a
production release.

## Design and authentication references

- [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html):
  preserve server-side session invalidation and appropriate cookie/cache controls.
- [W3C accessible authentication](https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum.html):
  retain passwordless methods, permit paste/autofill, and provide accessible controls.

## Product evidence after this delivery

A mature interface does not establish learning effectiveness or product-market fit.
Run a small A2–B1 learner pilot around one use case, observe first-session friction,
and measure completed learning sessions and return visits. Report cohort sizes and
observation windows; do not present proposed targets as achieved metrics. Existing
review, sentence and daily-activity records can support aggregate learning reports,
but do not measure landing-page or onboarding abandonment. Add instrumentation only
with clear event definitions and without sentence text, emails, or session material.

The canonical seed currently contains seven situations, 39 words and 42 meanings.
Expand a coherent reviewed curriculum based on pilot needs; generated examples need
editorial review before becoming authoritative learning content. Pronunciation,
reminders, monetization and native apps require separate evidence-led decisions.

Before widening rollout, retain real-provider AI evaluation evidence and report
review, prove a database restore into an isolated environment, and verify the live
Google/email flow. Fixture tests and CI cannot establish those claims by themselves.

## Implementation details

- Current-user responses include the opaque authenticated learner ID. Drafts are
  optional tab-local session storage, keyed by learner/source/attempt and expiring
  after two hours. They contain sentence text and retry identity, never bearer
  tokens or CSRF values. The app stays usable when storage is unavailable.
- Review-completion context allows an unfinished sentence to be recovered without
  repeating a completed review. Explicit account exit clears draft/context data.
- Google still returns only to configured OAuth entry points. A short-lived,
  one-use tab continuation restores an allowed app destination; storage failure
  falls back to the configured entry. Email links retain validated destinations.
- Auth responses use no-store cache controls. Logout keeps CSRF protection and
  revokes live sessions, handles already-ended sessions idempotently, and propagates
  storage failures instead of claiming a successful sign-out.
- An authenticated browser can recover a missing CSRF cookie through `/api/v1/me`.
  The app shell bootstraps it after reopening the browser, and logout awaits a
  recovery request when needed. Protected writes still require a matching cookie
  and header; server-component fetches are not treated as browser cookie updates.
  If recovery confirms an expired session, sign-in explains the interruption and
  retains the user's draft instead of repeatedly offering a failing logout.
- Sentence history uses existing requester-scoped API pagination. Dates are
  explicitly presented in UTC, and stored feedback remains distinguishable from
  the learner's original text and from incomplete processing states.

## Verification record

- Workspace validation passed formatting, lint/vet, type checking, 242 foundation
  tests, 31 API-client tests and 73 web helper tests. Its API stage stopped at the
  two controlled-signup OAuth tests because Docker is unavailable locally.
- `go test -skip TestControlledSignupOAuth ./...` passed; this does not establish
  live database/OAuth integration. CI must run the full suite with its database.
- Production workspace build, final web rebuild, lint and formatting passed.
- Full Playwright browser/accessibility verification passed 193 tests with 17
  existing layout-specific skips across desktop, 360px and 430px layouts.
- Independent review identified invalid OAuth recovery configuration and draft
  cleanup iteration edge cases. Follow-up regressions cover safe callback failure,
  explicit OAuth-state cookie clearing, and adjacent expired/live drafts; the web
  helper suite now contains 75 passing tests.
- All 12 Lighthouse audits passed: performance 100, accessibility 100 and best
  practices 96 on every audited screen/layout. These are local production-build
  results against fixtures, not claims about production traffic or devices.
- The first CI run passed full API tests, controlled-signup OAuth, web validation,
  container smoke and Lighthouse. Latest-commit CI and independent review are
  tracked in [PR #1464](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1464).
  The PR stays draft until those checks and review are resolved; the PR is the
  authoritative record of the eventual merge.
