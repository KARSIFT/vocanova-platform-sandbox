# VOC-025 — Acceptance Criteria

## VOC-025-AC-00 — Identity persistence and migration integrity

- Requirement source: `VOC-025-D00`, `VOC-025-D01`, `VOC-025-D04`
- Tasks: `VOC-025-T00`
- Tests: `VOC-025-TEST-00`, `VOC-025-TEST-01`
- Evidence: `VOC-025-EV-00`, `VOC-025-EV-01`
- Result: pending

Ent/Atlas create users, external identities, sessions, and magic links with required uniqueness, FK, status, expiry, one-use/revocation, and hashed-bearer invariants. Empty-db migration and recovery rehearsal preserve integrity.

## VOC-025-AC-01 — Magic-link authentication is secure

- Requirement source: `VOC-025-D00`, `VOC-025-D01`, `VOC-025-D03`, `VOC-025-D07`
- Tasks: `VOC-025-T01`
- Tests: `VOC-025-TEST-02`..`VOC-025-TEST-04`
- Evidence: `VOC-025-EV-02`..`VOC-025-EV-04`
- Result: pending

A valid verified-email link creates/links the correct user and session. Expired, replayed, malformed, tampered, and cross-environment links create no session or enumeration disclosure; rate limits protect provider work.

## VOC-025-AC-02 — Google OAuth identity linking is secure

- Requirement source: `VOC-025-D01`, `VOC-025-D03`, `VOC-025-D07`
- Tasks: `VOC-025-T02`
- Tests: `VOC-025-TEST-05`, `VOC-025-TEST-06`
- Evidence: `VOC-025-EV-05`, `VOC-025-EV-06`
- Result: pending

Validated state/callback and verified provider identity create/link exactly one permitted internal identity and session. Invalid/replayed state, provider failure, disabled user, unverified email, and ambiguous linking fail without credential or account disclosure.

## VOC-025-AC-03 — Session lifecycle and CSRF controls hold

- Requirement source: `VOC-025-D00`, `VOC-025-D03`
- Tasks: `VOC-025-T01`, `VOC-025-T02`, `VOC-025-T03`
- Tests: `VOC-025-TEST-07`..`VOC-025-TEST-10`
- Evidence: `VOC-025-EV-07`..`VOC-025-EV-10`
- Result: pending

Sessions have approved cookie/hashed-server-record controls, survive normal navigation before expiry, reject expiry/revocation/disablement, and logout invalidates them. Invalid CSRF is rejected; valid requests work. No sliding renewal is introduced.

## VOC-025-AC-04 — API authentication and cross-user isolation are enforced

- Requirement source: `VOC-025-D02`, `VOC-025-D03`
- Tasks: `VOC-025-T03`, `VOC-025-T04`
- Tests: `VOC-025-TEST-11`..`VOC-025-TEST-14`
- Evidence: `VOC-025-EV-11`..`VOC-025-EV-14`
- Result: pending

All A1-touched private `/api/v1` operations require active auth; malformed, expired, revoked, and disabled sessions fail. A user cannot read, mutate, infer, or share idempotency state with another; inaccessible private resources return 404. DTO/OpenAPI/client behavior agrees.

## VOC-025-AC-05 — Existing app routes have a working authenticated shell

- Requirement source: `VOC-025-D02`, `VOC-025-D06`
- Tasks: `VOC-025-T04`
- Tests: `VOC-025-TEST-15`..`VOC-025-TEST-17`
- Evidence: `VOC-025-EV-15`..`VOC-025-EV-17`
- Result: pending

The adopted journey protects existing `(app)` routes before rendering content; both auth methods retain a session during ordinary navigation; logout reaches signed-out state and protection resumes. Controls meet the stated accessibility criteria.

## VOC-025-AC-06 — Mock reconciliation does not invent later milestones

- Requirement source: `VOC-025-D05`
- Tasks: `VOC-025-T05`
- Tests: `VOC-025-TEST-18`, `VOC-025-TEST-19`
- Evidence: `VOC-025-EV-18`, `VOC-025-EV-19`
- Result: implemented

Only the adopted retain/wire/follow-up disposition is implemented. Real sources use approved requester-scoped contracts; retained learning mocks remain identified and do not masquerade as learner data. The mock inventory and deterministic inventory test verify this disposition.

## VOC-025-AC-07 — A1 evidence, staging, and rollback readiness are complete

- Requirement source: `VOC-025-D03`, `VOC-025-D04`, `VOC-025-D07`
- Tasks: `VOC-025-T00`..`VOC-025-T05`
- Tests: `VOC-025-TEST-20`..`VOC-025-TEST-23`
- Evidence: `VOC-025-EV-20`..`VOC-025-EV-23`
- Result: in-repository evidence collected; live staging/rollback evidence blocked by `VOC-025-DEP-01`

Applicable checks, migration/auth/contract tests, exact-SHA reviews, and the deterministic mock-inventory test pass. Staging tests for both methods/navigation/logout/unauthorized/cross-user/rate limit and the session-safe rollback rehearsal are documented and ready to run once the F3 staging environment exists (`VOC-025-DEP-01`). This enables—but does not declare—the DOC-12 A1 gate evaluation.
