# VOC-102 — Acceptance Criteria

## VOC-102-AC-00 — Ownership detected from governed package data before dispatch

- Requirement source: `VOC-102-D00`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-00`, `VOC-102-TEST-01`
- Evidence: `VOC-102-EV-00`
- Result: pending

Before auto-advance selects any path-specific PR/dispatch outcome, it reads
next-task ownership from governed package data. When
`<package_path>/.karsift/live-evidence/<next_task_id>.yaml` exists, that contract
is the authoritative machine-readable ownership source, while readable canonical
`tasks.md` and its matching task stanza remain a mandatory roster cross-check;
missing or unreadable roster data fails closed. The only secondary
expectation source is one exact allowlisted `Automation ownership` marker inside
the matching task stanza; narrative prose is never interpreted.

## VOC-102-AC-01 — No implementer dispatch for operator-owned or live-evidence-only next tasks

- Requirement source: `VOC-102-D01`, `VOC-102-D02`
- Tasks: `VOC-102-T00`, `VOC-102-T01`
- Tests: `VOC-102-TEST-02`, `VOC-102-TEST-08`, `VOC-102-TEST-11`
- Evidence: `VOC-102-EV-00`, `VOC-102-EV-01`
- Result: pending

When next-task ownership is `operator` or `live-actions`, auto-advance does **not**
start `implement.yml`. The next task issue remains open. A deterministic clean
publisher creates/reuses a draft evidence-carrier PR containing only the
governance-derived `<package_path>/tNN-evidence.md` path and posts one deduplicated sanitized
waiting marker, giving the existing PR-centric reconciler a valid attachment
point. No non-schema `evidence_path` field is added to the live-evidence contract.
If a prior publisher attempt created the deterministic carrier but stopped before
the derived evidence file or marker was complete, a repeated auto-advance event
re-enters the clean publisher and repairs that state without creating duplicates.

## VOC-102-AC-02 — Ordinary implementation next tasks still dispatch

- Requirement source: `VOC-102-D03`
- Tasks: `VOC-102-T00`, `VOC-102-T01`
- Tests: `VOC-102-TEST-03`, `VOC-102-TEST-09`
- Evidence: `VOC-102-EV-00`, `VOC-102-EV-01`
- Result: pending

When the next task is ordinary (no live-evidence contract and no contradictory
operator-owned declaration), auto-advance still sets
`should_dispatch=true` and invokes `implement.yml` attempt 1 after existing
open-issue and existing-PR guards.

## VOC-102-AC-03 — Fail closed on missing, malformed, or contradictory ownership metadata

- Requirement source: `VOC-102-D04`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-04`, `VOC-102-TEST-05`, `VOC-102-TEST-06`
- Evidence: `VOC-102-EV-00`
- Result: pending

Malformed YAML, missing/unrecognized `ownership`, `task_id` mismatch, unreadable
contracts, or an exact task-stanza operator/live-actions marker without a valid
matching contract do **not** dispatch the implementer. Duplicate, invalid, or
conflicting markers also fail closed. Absence of both marker and contract is the
ordinary path. A sanitized, deduplicated failure marker is emitted instead of
guessing. No carrier is created from malformed/untrusted path metadata.

## VOC-102-AC-04 — Final-roster release behavior preserved

- Requirement source: `VOC-102-D05`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-07`
- Evidence: `VOC-102-EV-00`
- Result: pending

Skipping implementer for an operator-owned next task does not open or advance
release. No-next-task / last-task behavior still defers to existing release
check-completion. Release still waits until the operator task actually closes.

## VOC-102-AC-05 — Deterministic test matrix landed

- Requirement source: `VOC-102-D06`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-00` through `VOC-102-TEST-07`, `VOC-102-TEST-11`
  through `VOC-102-TEST-13`
- Evidence: `VOC-102-EV-00`
- Result: pending

Positive dispatch, negative skip, deterministic carrier creation/reuse,
least-privilege separation, malformed/contradictory fail-closed, and last-task /
no-next regression coverage exist and pass in CI or infra self-ci.

## VOC-102-AC-06 — Controlled sanitized workflow proof

- Requirement source: `VOC-102-D07`
- Tasks: `VOC-102-T00`, `VOC-102-T01`
- Tests: `VOC-102-TEST-08`, `VOC-102-TEST-09`, `VOC-102-TEST-13`
- Evidence: `VOC-102-EV-00`, `VOC-102-EV-01`
- Result: pending

The real T00-close event proves no implementer execution for the operator-owned
T01 and creates its carrier. A later read-only verifier, manually dispatched on
the exact carrier head, validates that source run/job metadata and carrier state;
its successful run supplies the reconcilable `exact_pr_head` evidence. An ordinary
implementation next task still dispatches as intended. Evidence is metadata-only;
no logs, artifacts, secrets, or manufactured unrelated-package live evidence.

## VOC-102-AC-07 — Docs match skip-vs-dispatch behavior when touched

- Requirement source: AGENTS.md doc-consistency rule; `VOC-102-D01`, `VOC-102-D03`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-10`
- Evidence: `VOC-102-EV-00`
- Result: pending

Infra README and any in-diff calling-repo operator/governance docs accurately
state that auto-advance skips general implementer dispatch for operator-owned /
live-evidence-only next tasks and leaves them on the reconcile path — or those
docs are left unchanged only when they already do not claim universal implement
dispatch.

## VOC-102-AC-08 — Classifier and clean publisher remain least-privileged

- Requirement source: `VOC-102-D02`, `VOC-102-D04`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-11` through `VOC-102-TEST-13`
- Evidence: `VOC-102-EV-00`
- Result: pending

The ownership classifier keeps read-only contents/issues/pull-request access and
cannot mutate repository state. Only a separate clean, non-LLM publisher may
mint the existing App for explicit contents/issues/pull-requests writes needed
to create/reuse the carrier and marker. It has no Actions-write permission,
model credentials, or `secrets: inherit` path into `implement.yml`.
The proof verifier is independently read-only and receives no App write token,
model/deploy/application secret, or Actions-write permission.
