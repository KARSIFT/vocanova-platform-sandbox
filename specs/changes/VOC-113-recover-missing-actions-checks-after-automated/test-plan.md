# VOC-113 — Test Plan

## VOC-113-TEST-00 — Evidence documents issue #948 and verified trigger/token/event behavior

- Covers: `VOC-113-AC-00`
- Preconditions: T00 evidence drafted at implementation time
- Procedure: Read `t00-evidence.md`; assert it names the VOC-112 develop merge
  without push workflows, promotion PR #947 without required checks, failed
  close/reopen and draft/ready recovery, reconcile-release reuse without merge,
  and a concrete documented trigger/token/event explanation (not a guess).
- Expected result: Diagnosis is metadata-only and bounds remediation to missing
  downstream workflow activation
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-01 — Task-merge recovery positive path

- Covers: `VOC-113-AC-01`, `VOC-113-D03`, `VOC-113-D07`
- Preconditions: T00 task branch with recovery implementation
- Procedure: Deterministic fixture simulates App-driven integration merge whose
  exact squash SHA lacks required push/validation runs; assert recovery
  orchestrates genuine allowlisted dispatch/reusable-workflow invocation for that
  SHA and records sanitized diagnostics.
- Expected result: Exact-SHA recovery path engages; no status fabrication
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-02 — Task-merge recovery is a no-op when required runs already exist

- Covers: `VOC-113-AC-01`, `VOC-113-D05`
- Preconditions: T00 task branch
- Procedure: Fixture where the exact integration SHA already has authoritative
  successful required runs; assert recovery does not dispatch duplicates.
- Expected result: Idempotent no-op
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-03 — Release-PR recovery positive path

- Covers: `VOC-113-AC-02`, `VOC-113-D03`, `VOC-113-D07`
- Preconditions: T00 task branch
- Procedure: Fixture simulates App-created promotion PR whose exact head lacks
  required pull-request checks; assert recovery starts genuine runs for that head
  and release converge remains non-merging until authoritative success.
- Expected result: Exact-head recovery path engages; merge still fail-closed
  until checks succeed
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-04 — Release-PR recovery on reconcile-release wake

- Covers: `VOC-113-AC-02`, `VOC-113-D04`
- Preconditions: T00 task branch
- Procedure: Fixture models `reconcile-release` against an existing unique
  promotion PR with missing exact-head checks; assert audit/PR are reused (no
  duplicates) and recovery is attempted for the live head SHA.
- Expected result: Idempotent reuse + recovery; no second promotion PR
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-05 — Fail closed on wrong SHA, stale, absent, or fabricated evidence

- Covers: `VOC-113-AC-03`, `VOC-113-D01`, `VOC-113-D04`
- Preconditions: T00 task branch
- Procedure: Negative fixtures for checks bound to another SHA, stale evidence,
  absent required contexts, and any attempt to treat synthesized statuses as
  success; assert refuse/fail-closed and no merge decision.
- Expected result: No merge; diagnostics name the failure class without secrets
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-06 — Bounded timeout and recursion/duplicate guards

- Covers: `VOC-113-AC-03`, `VOC-113-D05`, `VOC-113-D06`
- Preconditions: T00 task branch
- Procedure: Fixtures where recovered runs never appear within the configured
  bound, where a second recovery would recurse through the same wake path, and
  where a second promotion PR/release audit would be created; assert timeout
  fail-closed, recursion halt, and duplicate refusal.
- Expected result: Bounded failure with sanitized diagnostics; invariants hold
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-07 — App-token mutation posture and ruleset contexts preserved

- Covers: `VOC-113-AC-03`, `VOC-113-AC-04`, `VOC-113-D02`
- Preconditions: T00 task branch
- Procedure: Inspect merge-gate/release policy tests and workflow text; assert
  App credentials remain required for workflow-triggering merges/PR creation,
  `github.token` merge fallback remains refused when App secrets are present, and
  required ruleset context names are not removed or replaced by synthetic
  success markers.
- Expected result: Auth and ruleset posture unchanged except for durable recovery
- Evidence: `VOC-113-EV-00`

## VOC-113-TEST-08 — Live completion of promotion PR #947 after genuine exact-head checks

- Covers: `VOC-113-AC-05`
- Preconditions: T00 live; open promotion PR #947; operator-owned contract
  `.karsift/live-evidence/VOC-113-T01.yaml`
- Procedure: Operator/reconcile path recovers genuine required checks on #947's
  exact head; `verify-promotion-check-recovery / verify` succeeds on the T01 PR
  for that fixture; read `t01-evidence.md` for allowlisted metadata proving
  successful exact-head checks and the subsequent single merge decision (or
  correct fail-closed if checks failed).
- Expected result: #947 completes only after genuine checks; no fabricated statuses
- Evidence: `VOC-113-EV-01`

## VOC-113-TEST-09 — Live post-promotion workflow verification

- Covers: `VOC-113-AC-06`
- Preconditions: T01 complete (PR #947 merged); operator-owned contract
  `.karsift/live-evidence/VOC-113-T02.yaml`
- Procedure: Resolve promotion result SHA on `main`; confirm expected
  post-promotion workflow run(s) for that SHA; read `t02-evidence.md`.
- Expected result: Post-promotion evidence complete before remediation closure
- Evidence: `VOC-113-EV-02`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
