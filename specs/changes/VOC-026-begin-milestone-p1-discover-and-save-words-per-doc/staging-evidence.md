# VOC-026 — P1 Staging and Rollback Evidence Collection

## Purpose

This document records the evidence required by `VOC-026-AC-07` (via
`VOC-026-TEST-12` and `VOC-026-TEST-20`..`VOC-026-TEST-24`). It collects the
in-repository evidence that can be produced without the F3 staging environment
and documents the staging exercises and rollback rehearsal that can only be run
once F3 exists.

## Current status

Execution of live staging evidence (`VOC-026-EV-21`..`VOC-026-EV-23`) is
**blocked** by `VOC-026-DEP-03`: the F3 staging environment does not yet exist.
T00–T04 have implemented the code; T05 collects every piece of evidence that can
be produced in-repository and documents the remaining staging/rollback work for
F3.

No DOC-12 P1-complete declaration is made here.

## Evidence collected in repository

| Evidence ID     | Test                                                                        | Status       | Location                                                                                                                                                                                                                                          |
| --------------- | --------------------------------------------------------------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `VOC-026-EV-12` | `VOC-026-TEST-12` — Installed deterministic and security suite              | Collected    | `pnpm run validate`, `pnpm run test`, `pnpm run build`, `cd apps/api && go vet ./... && go test ./...`, and the scripts/foundation tests                                                                                                          |
| `VOC-026-EV-20` | `VOC-026-TEST-20` — Mock-decommission inventory                             | Collected    | `scripts/foundation/mock-inventory.mjs`, `scripts/foundation/mock-inventory.test.mjs`, and `specs/changes/VOC-026-begin-milestone-p1-discover-and-save-words-per-doc/mock-inventory.md`                                                           |
| `VOC-026-EV-21` | `VOC-026-TEST-21` — Staging discover→save→consistency→unsave                | **Blocked**  | Requires F3 staging environment with seeded content and two isolated non-production users. See procedure below.                                                                                                                                   |
| `VOC-026-EV-22` | `VOC-026-TEST-22` — Staging authorization, CSRF, and idempotency validation | **Blocked**  | Requires F3 staging environment with two isolated non-production users and CSRF/idempotency tooling. See procedure below.                                                                                                                         |
| `VOC-026-EV-23` | `VOC-026-TEST-23` — Content and user-words rollback rehearsal               | **Blocked**  | Requires a staged candidate, approved rollback procedure, and disposable PostgreSQL. See procedure below.                                                                                                                                         |
| `VOC-026-EV-24` | `VOC-026-TEST-24` — Exact-SHA independent verification                      | **Required** | Performed by Claude Code on each T00–T05 final SHA; reports scope, classifier floor, migration safety, requester scope, idempotency, CSRF, contract drift, secrets/logging, accessibility, staging/rollback evidence, and implementer separation. |

## Staging discover→save→consistency→unsave procedure (EV-21)

When F3 staging is available, perform the following with non-production
identities only:

1. **Seed and authenticate**
   - Ensure the F3 database has the VOC-026 P1 content and idempotency
     migrations applied and the deterministic seed command has run.
   - Authenticate as test user A through the A1 sign-in flow and obtain a valid
     `vocanova_session` and `vocanova_csrf` token pair.

2. **Discover and save**
   - Open `/discover` and verify the seven seeded situations render.
   - Choose a situation and open its drill-down; verify the meanings list shows
     the requester-scoped `saved` overlay (initially `false`).
   - Open a word-detail page and save one meaning using the save control.
   - Verify the save request carries `X-CSRF-Token` and an `Idempotency-Key`.

3. **Consistency across screens**
   - Return to the situation drill-down and verify the saved meaning shows the
     saved badge.
   - Navigate to `/home` and verify the saved word appears in the saved-words
     list.
   - Navigate to `/progress` and verify the saved word appears in the saved
     vocabulary list.
   - Hard-refresh each page and confirm the saved state persists.

4. **Unsave**
   - From the word-detail page, unsave the same meaning.
   - Revisit the situation drill-down, `/home`, and `/progress` and confirm the
     word is no longer flagged as saved.

5. **No P2/P4 side effects**
   - Inspect the `user_words` row and confirm only the P1-owned fields are set
     (`status='new'`, `source='journey'`, `review_step=0`, `added_at`).
   - Confirm no `next_review_at`, Confidence Point ledger, mission, or daily
     activity rows were created.

## Staging authorization, CSRF, and idempotency validation procedure (EV-22)

When F3 staging is available:

1. **Unauthenticated access**
   - With no session cookie, request `GET /api/v1/journey-situations` and
     `GET /api/v1/canonical-words/{wordSlug}`.
   - Confirm each returns `401 Unauthorized` and the `(app)` routes redirect to
     `/signin`.
   - Attempt `POST /api/v1/user-words` and `DELETE /api/v1/user-words/{meaningId}`
     without a session; confirm `401 Unauthorized`.

2. **Cross-user isolation**
   - Create two isolated users, A and B.
   - As A, save a meaning and record its `userWordId` and `meaningId`.
   - As B, attempt to `DELETE /api/v1/user-words/{meaningId}` for A's saved row
     (using the same meaningId) and attempt to read A's saved-words list.
   - Confirm both attempts return `404 Not Found` and do not leak A's saved-word
     choices.

3. **CSRF enforcement**
   - As A, call `POST /api/v1/user-words` without `X-CSRF-Token`: expect
     `403 Forbidden`.
   - Call with a mismatched `X-CSRF-Token`: expect `403 Forbidden`.
   - Call with a valid cookie/header pair: expect `200 OK` or `201 Created`.
   - Repeat for `DELETE /api/v1/user-words/{meaningId}`.

4. **Idempotency enforcement**
   - As A, save a meaning with a fixed `Idempotency-Key` and fingerprint; replay
     the same request and confirm it is idempotent (no duplicate row).
   - Reuse the same key with a different meaningId/fingerprint: expect
     `409 Conflict`.
   - As B, use the exact same key and fingerprint as A; confirm B's request is
     isolated from A's and succeeds with a different saved row.

5. **Redaction**
   - Inspect response bodies, logs, and the OpenAPI document.
   - Confirm no raw session tokens, CSRF tokens, or another learner's saved-word
     choices are exposed.

## Content and user-words rollback rehearsal procedure (EV-23)

When a staged candidate is ready and the rollback procedure is approved:

1. **Pre-rollback**
   - Capture a backup/snapshot of the staging database after the VOC-026 P1
     content and `user_words` migrations and after the seed command has run.
   - Record the last-known-good code revision.
   - Identify all active sessions and any saved `user_words` rows.

2. **Rollback**
   - Stop the API service.
   - Revert the code to the last-known-good revision.
   - Restore the database to the pre-migration state using the approved
     procedure.

3. **Post-rollback safety**
   - Confirm all candidate-era sessions are invalidated (do not trust restored
     revocation/expiry state alone; explicitly revoke or delete them).
   - Confirm canonical content and soft-deleted `user_words` rows are preserved.
   - Confirm no raw bearer tokens, CSRF tokens, or personal saved-content choices
     remain in logs or test artifacts.

4. **Service validation**
   - Start the API service.
   - Run health checks.
   - Confirm the P1 content and saved-word endpoints are unavailable until the
     next forward migration, or are restored to a consistent pre-candidate state
     if the rollback deliberately retained the P1 schema.

## Dependency blocker

`VOC-026-DEP-03`: F3 staging environment must exist before the live staging
exercises (`VOC-026-EV-21`, `VOC-026-EV-22`, `VOC-026-EV-23`) can be executed.

`VOC-026-T05` provides the in-repository evidence and the documented procedures
only; it does not declare the DOC-12 P1 gate complete.
