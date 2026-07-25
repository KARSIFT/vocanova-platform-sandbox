# VOC-025 — Tasks

Mandatory PR order: `T00 → T01 → T02 → T03 → T04 → T05`. Each PR is independently reviewable, remains R3-proposed, and requires Claude Code exact-SHA review.

## VOC-025-T00 — Establish identity persistence, migrations, and contract foundation

- Requirement source: `VOC-025-D00`, `VOC-025-D01`, `VOC-025-D04`
- Acceptance criteria: `VOC-025-AC-00`, `VOC-025-AC-07`
- Tests: `VOC-025-TEST-00`, `VOC-025-TEST-01`, `VOC-025-TEST-20`
- Evidence: `VOC-025-EV-00`, `VOC-025-EV-01`, `VOC-025-EV-20`
- Status: pending

Reconcile current Go/Ent/Atlas/Huma commands. Add schemas/migrations for users, external identities, sessions, and magic links using DOC-05 constraints; add minimal explicit auth/current-user DTO/OpenAPI generation foundation, not a login flow. Rehearse disposable forward migration and recovery; production migration never runs at API startup.

## VOC-025-T01 — Implement magic link, session primitives, CSRF, and rate limits

- Requirement source: `VOC-025-D00`, `VOC-025-D01`, `VOC-025-D03`, `VOC-025-D07`
- Acceptance criteria: `VOC-025-AC-01`, `VOC-025-AC-03`, `VOC-025-AC-07`
- Tests: `VOC-025-TEST-02`..`VOC-025-TEST-04`, `VOC-025-TEST-07`..`VOC-025-TEST-10`, `VOC-025-TEST-21`
- Evidence: `VOC-025-EV-02`..`VOC-025-EV-04`, `VOC-025-EV-07`..`VOC-025-EV-10`, `VOC-025-EV-21`
- Status: pending

Add email-provider/clock interfaces, magic-link request/consume, hashed session issue/validate/revoke, secure cookies, logout, CSRF, rate limits, and cleanup. Use fake email only in CI; add service/HTTP/PostgreSQL coverage for replay, expiry, enumeration, cookies, logout, CSRF, thresholds, and log redaction.

## VOC-025-T02 — Add Google OAuth and safe identity linking

- Requirement source: `VOC-025-D01`, `VOC-025-D03`, `VOC-025-D07`
- Acceptance criteria: `VOC-025-AC-02`, `VOC-025-AC-03`, `VOC-025-AC-07`
- Tests: `VOC-025-TEST-05`..`VOC-025-TEST-07`, `VOC-025-TEST-21`
- Evidence: `VOC-025-EV-05`..`VOC-025-EV-07`, `VOC-025-EV-21`
- Status: pending

After approved provider setup, add start/callback with state protection, verified provider identity, duplicate-safe linking, and T01 sessions. Use fake provider in CI and staging credentials only via approved secret delivery. Never expose OAuth access/refresh values.

## VOC-025-T03 — Enforce API authentication and authorization

- Requirement source: `VOC-025-D02`, `VOC-025-D03`, `VOC-025-D04`
- Acceptance criteria: `VOC-025-AC-03`, `VOC-025-AC-04`, `VOC-025-AC-07`
- Tests: `VOC-025-TEST-09`..`VOC-025-TEST-14`, `VOC-025-TEST-22`
- Evidence: `VOC-025-EV-09`..`VOC-025-EV-14`, `VOC-025-EV-22`
- Status: pending

Add auth context/middleware and requester-scoped private route/service patterns. Establish 401/404, CSRF, validation, and redaction behavior; regenerate contract/client. Test two-user guessed IDs, disabled users, invalid cookies, and idempotency isolation where applicable.

## VOC-025-T04 — Integrate authenticated shell and logout journey

- Requirement source: `VOC-025-D02`, `VOC-025-D06`, `VOC-025-D07`
- Acceptance criteria: `VOC-025-AC-03`, `VOC-025-AC-05`, `VOC-025-AC-07`
- Tests: `VOC-025-TEST-15`..`VOC-025-TEST-17`, `VOC-025-TEST-22`
- Evidence: `VOC-025-EV-15`..`VOC-025-EV-17`, `VOC-025-EV-22`
- Status: pending

Implement only adopted sign-in/return/logout behavior. Protect the complete `(app)` group, including dynamic routes, with API-backed identity and no client DB/duplicated authorization. Use existing component/browser tools; a missing tool is a recorded limitation, not a pass or an implicit new dependency.

## VOC-025-T05 — Reconcile adopted mock disposition and collect A1 staging evidence

- Requirement source: `VOC-025-D04`, `VOC-025-D05`, `VOC-025-D06`
- Acceptance criteria: `VOC-025-AC-04`, `VOC-025-AC-06`, `VOC-025-AC-07`
- Tests: `VOC-025-TEST-18`..`VOC-025-TEST-23`
- Evidence: `VOC-025-EV-18`..`VOC-025-EV-23`
- Status: implemented; staging evidence blocked by `VOC-025-DEP-01`

Inventory every VOC-010–VOC-024 mock source and map it to adopted retain/wire/follow-up. Do not add P1–P4 APIs. In staging exercise both auth methods, navigation, logout, unauthorized/cross-user/CSRF/abuse paths, and migration/session-safe rollback; collect evidence without declaring A1 complete.

### Deliverables

- `mock-inventory.md`: maps every retained learning-content mock to its VOC
  source, disposition (`retain` or `wire to A1`), and follow-up package.
- `staging-evidence.md`: collected in-repository evidence (`EV-18`..`EV-20`)
  and documented procedures for blocked staging evidence (`EV-21`..`EV-23`).
- `scripts/foundation/mock-inventory.mjs` and `mock-inventory.test.mjs`:
  deterministic check that retained mocks are unchanged and no P1–P4 API
  routes were invented.
- Minor comment updates in the VOC-022 mock source and word-detail page to
  identify them explicitly as VOC-022 placeholders (data and rendering
  unchanged).

### Blocker

`VOC-025-DEP-01` remains open: F3 staging does not exist, so live staging
exercises (`EV-21`, `EV-22`, `EV-23`) cannot be executed. The implementation
provides the procedures and the in-repository evidence only; it does not
declare A1 complete.
