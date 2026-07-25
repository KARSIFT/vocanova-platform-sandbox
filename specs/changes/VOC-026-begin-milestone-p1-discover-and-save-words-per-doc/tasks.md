# VOC-026 — Tasks

Mandatory PR order: `T00 → T01 → T02 → T03 → T04 → T05`. Each PR is independently reviewable,
remains R3-proposed (path floor R3 for migrations/schemas), and requires Claude Code exact-SHA
review. Tasks that depend on open founder decisions (`D01`, `D03`, `D04`, `D05`) may not
proceed past them by guessing; record the adopted resolution in `D06` first.

## VOC-026-T00 — Canonical content and user-words persistence, migrations, and deterministic seed

- Requirement source: `VOC-026-D00`, `VOC-026-D01`, DOC-05 §§7–9,18–20
- Acceptance criteria: `VOC-026-AC-00`, `VOC-026-AC-01`
- Tests: `VOC-026-TEST-00`, `VOC-026-TEST-01`
- Evidence: `VOC-026-EV-00`, `VOC-026-EV-01`
- Status: pending

Add Ent schemas + reviewed versioned Atlas SQL for `canonical_words`, `word_meanings`,
`word_examples`, `usage_notes`, `journey_situations`, `journey_words`, and `user_words` per
DOC-05 §§7–9 (one Ent schema file per table under `apps/api/ent/schema/`, using the existing
`UUIDMixin`/`TimeMixin` and a `SoftDeleteMixin` where DOC-05 §16 soft-deletes). Add the
versioned deterministic seed JSON and a single-transaction Go seed command with fixed UUIDs,
scoped exactly to the adopted `D01` MVP seed-data scope. Rehearse disposable forward migration
and recovery; production migration never runs at API startup. No API routes, no DTOs, no
frontend in this PR.

## VOC-026-T01 — Discovery and word-detail read API with requester-scoped saved overlay

- Requirement source: `VOC-026-D00`, `VOC-026-D02`, `VOC-026-D03`, DOC-05 §8, DOC-07
- Acceptance criteria: `VOC-026-AC-02`
- Tests: `VOC-026-TEST-02`..`VOC-026-TEST-06`, `VOC-026-TEST-14`
- Evidence: `VOC-026-EV-02`..`VOC-026-EV-06`, `VOC-026-EV-14`
- Status: pending

Add the `content` business module (DOC-06 §3) with explicit request/response DTOs and Huma
routes: list situations, a situation drill-down with meanings, and a word-detail read
endpoint — keyed per the adopted `D02` resolution. Apply requester-scoped saved-state overlay
by reading the authenticated requester's `user_words` (read-only, cross-module through a
service query, not direct cross-module table writes — DOC-06 §3). Enforce `RequireAuth`, 401,
404 for unknown slugs/words, cursor pagination for lists, stable operation IDs, and commit
the OpenAPI artifact. Follow the adopted `D03` discovery resolution exactly (show-all-with-overlay
or exclude-saved). No save/unsave, no frontend yet.

## VOC-026-T02 — Save/unsave user words with idempotency, CSRF, and authorization

- Requirement source: `VOC-026-D00`, `VOC-026-D02`, `VOC-026-D04`, DOC-06 §§8–10, DOC-07
- Acceptance criteria: `VOC-026-AC-03`
- Tests: `VOC-026-TEST-07`..`VOC-026-TEST-11`, `VOC-026-TEST-13`
- Evidence: `VOC-026-EV-07`..`VOC-026-EV-11`, `VOC-026-EV-13`
- Status: pending

Add the `learning` business module: `POST /api/v1/user-words` (save a meaning) and
`DELETE /api/v1/user-words/{meaningId}` (unsave, per the adopted `D02` keying) with the
existing `RequireAuth`, `CSRFMiddleware`, and `Idempotency-Key` (user+operation scoped,
DOC-07) glue. Save inserts/restores one `user_words` row in one transaction under the adopted
`D04` resolution and writes no P2/P4 tables; unsave soft-deletes. Use `AuthorizeOwner`/404 for
any owner mismatch. Add a `GET /api/v1/user-words` (the requester's saved meanings, cursor
pagination) for cross-screen consistency. Commit OpenAPI and the matched client. No frontend yet.

## VOC-026-T03 — Wire Discover, Situation drill-down, and Word-Detail screens to the real API

- Requirement source: `VOC-026-D02`, `VOC-026-D03`, DOC-12 §5 P1
- Acceptance criteria: `VOC-026-AC-04`
- Tests: `VOC-026-TEST-15`..`VOC-026-TEST-17`, `VOC-026-TEST-19`
- Evidence: `VOC-026-EV-15`..`VOC-026-EV-17`, `VOC-026-EV-19`
- Status: pending

Replace `MOCK_DISCOVER_SITUATIONS` and `MOCK_SITUATION_WORD_LISTS` (and the word-detail mock
import in `apps/web/src/app/(app)/discover/[situation]/[word]/page.tsx`) with server
components calling `@vocanova/api-client` against the real API under the A1 session. Render
the situation list, the situation word list with saved state (per adopted `D03`), and
word-detail meanings/examples/usage-notes with a working save/unsave control (CSRF + the
client method from T02). Keep stable slug routing and existing accessibility markup; unknown
slugs/words `notFound()`. Remove the mock sources, do not retain them as fallback.

## VOC-026-T04 — Reconcile saved-state consistency on Home and Progress with the retained P4 mocks

- Requirement source: `VOC-026-D05`, VOC-025-D05, DOC-12 §5 P1
- Acceptance criteria: `VOC-026-AC-05`
- Tests: `VOC-026-TEST-18`, `VOC-026-TEST-19`
- Evidence: `VOC-026-EV-18`, `VOC-026-EV-19`
- Status: pending

Per the adopted `D05` resolution, wire Home and Progress to the real saved-words API so a
save/unsave is consistent across home/discover/progress within normal navigation, and
explicitly label/retain any out-of-scope P4 mock fields (mission target, reviewed-today,
streak, due words, confidence points, completion history) as mock-pending-P4 (not real P1
learner data) — or replace them per the founder's choice. Add no P2–P4 tables, routes, or
behavior. This task is blocked until `D05` (and `D06`) are resolved at adoption.

## VOC-026-T05 — Mock-decommission inventory, P1 staging evidence, and gate readiness

- Requirement source: `VOC-026-D00`, VOC-025-D05, DOC-12 §5 P1
- Acceptance criteria: `VOC-026-AC-06`, `VOC-026-AC-07`
- Tests: `VOC-026-TEST-12`, `VOC-026-TEST-20`..`VOC-026-TEST-24`
- Evidence: `VOC-026-EV-12`, `VOC-026-EV-20`..`VOC-026-EV-24`
- Status: implemented; live staging evidence blocked by `VOC-026-DEP-03`

Inventory every VOC-010–VOC-024 mock touched by P1 and map each to its disposition
(decommissioned-to-real-P1 / retained-as-mock-pending-P4); add the deterministic mock-inventory
check confirming decommissioned mocks are gone and no P2–P4 API route/table/behavior was
invented. Where the F3 staging environment exists, exercise discover→word-detail→save→
consistency-across-screens→unsave, cross-user denial, CSRF, idempotency, and the
content/user-words rollback rehearsal under non-production identities; where it does not, record
the in-repository evidence and the documented procedures, and record live staging evidence as
blocked by `VOC-026-DEP-03`. Do not declare the DOC-12 P1 gate complete.

### Deliverables

- `mock-inventory.md`: maps every VOC-010–VOC-024 mock touched by P1 to its VOC
  source and disposition (`decommissioned-to-real-P1` or `retained-as-mock-pending-P4`).
- `staging-evidence.md`: collected in-repository evidence (`EV-12`, `EV-20`,
  `EV-24`) and documented procedures for blocked staging evidence (`EV-21`..`EV-23`).
- `scripts/foundation/mock-inventory.mjs` and `mock-inventory.test.mjs`:
  deterministic check that decommissioned mocks are gone and no P2–P4 API route,
  table, or behavior was invented.

### Blocker

`VOC-026-DEP-03` remains open: F3 staging does not exist, so live staging exercises
(`EV-21`, `EV-22`, `EV-23`) cannot be executed. The implementation provides the
procedures and the in-repository evidence only; it does not declare the DOC-12 P1
gate complete.
