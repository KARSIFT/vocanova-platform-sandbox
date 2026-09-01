# VOC-145 — Acceptance Criteria

## VOC-145-AC-00 — Live `roles.yml` matches the authorized current binding set

- Requirement source: `VOC-145-D01`, `VOC-145-D02`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-00`
- Evidence: `VOC-145-EV-00`
- Result: pending

After the independently reviewed infrastructure merge and caller pin,
`KARSIFT/karsift-ai-infra/config/roles.yml` and the caller mirrored fixture
contain exactly the authorized six bindings. Default Path A is the VOC-142
DEP-06 / `8993e867…` lineup (all review roles
`cursor/grok-4.6[effort=high,fast=false]`). Path B applies only when
adoption recorded `VOC-145-DEP-07` Path B. Header comments describe that
authorized current set and do not describe the other path as current.
Unauthorized head `d8720829…` is not treated as authority.

## VOC-145-AC-01 — Historical VOC-117 assertions are not rewritten to bless a later lineup

- Requirement source: `VOC-145-D03`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-01`
- Evidence: `VOC-145-EV-00`
- Result: pending

Path A restores the pre-drift VOC-117 current-state expectations
(`effort=high,fast=false` for planner and all review roles). Path B keeps a
named historical constant equal to those VOC-117 / `8993e867…` bindings and
a separate current-state constant equal to the Path B lineup. Tests do not
claim VOC-117 originally required `xhigh` or `fast=true`. The class of
rewriting `VOC117_BINDINGS` in the same sequence as an ungoverned
`roles.yml` change cannot recur as the acceptance proof for a later
lineup.

## VOC-145-AC-02 — Parameterized Cursor resolution and fail-closed invalid bindings remain

- Requirement source: `VOC-145-D06`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-02`, `VOC-145-TEST-03`
- Evidence: `VOC-145-EV-00`
- Result: pending

`prepare_cursor_model` still passes stored parameterized identifiers to the
Cursor CLI without stripping effort or speed, rejects effort-omitted Grok
4.6 forms, rejects missing `CURSOR_API_KEY` on required paths, and rejects
unsupported prefixes. No silent vendor, model, speed, or effort fallback.

## VOC-145-AC-03 — Current-state docs match the authorized lineup

- Requirement source: `VOC-145-D05`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-04`, `VOC-145-TEST-06`
- Evidence: `VOC-145-EV-00`
- Result: pending

An exhaustive tracked-source search identifies every current claim about
the six active bindings, `xhigh`, pin hashes, and `VOC117_BINDINGS`. Infra
`README.md`, infra `CHANGELOG.md`, fixture README, and every other
current-state match describe the authorized current set. Clearly marked
historical VOC-117 / VOC-142 records remain historical. Those package
directories are not rewritten.

## VOC-145-AC-04 — Caller pin and mirrored fixture bytes match the new infra merge

- Requirement source: `VOC-145-D04`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-05`
- Evidence: `VOC-145-EV-00`
- Result: pending

`PINNED_SHA.txt` equals the independently reviewed infrastructure merge that
contains this reconciliation, not leftover `d8720829…` and not leftover
`8993e867…` if that merge is no longer the live reconciled contract. Every
changed authoritative fixture file is byte-identical to that merge.

## VOC-145-AC-05 — Retry, exact-SHA review, and provider isolation are not weakened

- Requirement source: `VOC-145-D06`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-07`
- Evidence: `VOC-145-EV-00`
- Result: pending

Implementer retry remains capped at two attempts. Independent review still
binds the exact PR head. No OpenAI execution route is added. Path B
`fast=true` on `reviewer_fast_retry`, if adopted, does not expand retry
count or skip exact-SHA review. Credentials are never printed.

## VOC-145-AC-06 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-145-D09`, `VOC-145-D11`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-08`
- Evidence: `VOC-145-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, R4
classification, caller governance suite, mirrored infra suite,
`git diff --check`, and independent exact-revision review that binds the
live head all pass. Evidence does not require a commit to contain its own
SHA.

## VOC-145-AC-07 — No snapshot-gap; #1120 is not used as a carrier; #1124 closes on allowlisted metadata

- Requirement source: `VOC-145-D07`, `VOC-145-D12`
- Tasks: `VOC-145-T00`
- Tests: `VOC-145-TEST-09`
- Evidence: `VOC-145-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. No VOC-112
fixture recapture or #1120 resume helper is added. After exact-SHA review
and merge into `develop`, ordinary later promotion uses existing release
evaluation. Root issue #1124 closes only after allowlisted metadata from a
successful implement/release path exists. Closed state alone is not
completion proof. Named SHAs `8993e867…` / `d8720829…` and run
`33443684483` remain audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
