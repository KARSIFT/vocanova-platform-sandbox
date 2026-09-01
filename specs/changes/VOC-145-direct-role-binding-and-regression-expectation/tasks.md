# VOC-145 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1124 is one
authority-reconciliation outcome: make live `roles.yml`, VOC-117
current-state tests, caller fixtures/pin, README, CHANGELOG, and every other
current-state contract describe the same authorized binding set. Repository
count, restore-versus-docs, workflow-versus-tests-versus-docs, and
pin-versus-release are not split reasons.

Cross-repo note: coordinated PRs in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Infrastructure merge
`8993e867640dfb604dec0466c4e0787e68d8e258` is the last governed
role-binding contract; unauthorized head
`d8720829b176cf1287e633f9382989fc8f258105` must not be pinned as-is. T00
must open a new infra PR and pin its exact merge. Do not treat the
untracked local `karsift-ai-infra/` checkout (if present) as this repo's
tracked tree. No bootstrap exception. T00's first run is attempt `1` on a
new VOC-145 carrier from current `develop`.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). A VOC-097 live-evidence second
task is not used. Issue #1120 is not this task.

## VOC-145-T00 — Governed reconciliation of the unauthorized role-binding and VOC-117 expectation drift

- Requirement source: issue #1124; `VOC-145-D00` through `VOC-145-D12`
- Acceptance criteria: `VOC-145-AC-00` through `VOC-145-AC-07`
- Tests: `VOC-145-TEST-00` through `VOC-145-TEST-08`
- Post-T00 release/closure verification: `VOC-145-TEST-09` (not an
  implementation-task close or retry prerequisite)
- Evidence: `VOC-145-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1124 evidence in `t00-evidence.md` (last governed base
   `8993e867640dfb604dec0466c4e0787e68d8e258`, unauthorized head
   `d8720829b176cf1287e633f9382989fc8f258105`, self-CI run `33443684483`,
   the three role deltas, the VOC-117 test rewrite, and unreconciliation of
   README/CHANGELOG/caller fixtures/pin/VOC-142 DEP-06 current-state
   evidence). Do not copy raw provider responses, credentials, or the full
   log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation.
   Fail closed on unrelated/material movement of `develop`. This package's
   own plan/adoption/roster commits do not count as protected-file drift.
   Do not edit `specs/changes/VOC-142-…/`, `VOC-141-…/`, `VOC-140-…/`, or
   `VOC-117-…/`.
3. Read adoption records. Implement Path A unless adoption recorded
   `VOC-145-DEP-07` Path B. Do not select Path B from issue #1124's
   "either/or" wording.
4. Open a new `KARSIFT/karsift-ai-infra` PR from current infra `main`.
   Implement D01–D08 there:
   - Path A: restore `config/roles.yml` and VOC-117 current-state
     expectations to the `8993e867…` / VOC-142 DEP-06 bindings; restore
     header comments to explicit high-effort Standard (`fast=false`) for
     planner and review roles; update README and CHANGELOG so they no
     longer omit the ungoverned drift and the governed restore.
   - Path B only if adopted: keep the `xhigh` current-state lineup;
     preserve historical VOC-117 expected bindings as a named historical
     constant; add separate current-state constants; update README and
     CHANGELOG to describe the new current lineup without claiming VOC-117
     originally required it.
   Do not pin or consume `d8720829…` as the reconciliation merge.
5. Add or restore deterministic tests that read live `roles.yml` and call
   `prepare_cursor_model`. Cover the six exact authorized current bindings,
   historical-versus-current VOC-117 split, effort-omitted and missing-key
   fail-closed paths, and unchanged retry / exact-SHA / no-OpenAI
   controls. Do not treat a green self-CI run against rewritten VOC-117
   expectations as covering #1124.
6. Obtain independent exact-revision review of the infra PR and merge it.
   Record the exact merge SHA and confirm it is not `d8720829…`. From
   current caller `develop`, create a new VOC-145 implementation branch.
   Pin `PINNED_SHA.txt` to that infra merge and mirror every changed
   authoritative fixture file. Update fixture README and all current-state
   docs and caller tests found by exhaustive tracked-source search,
   including `test_voc117_role_bindings.py`,
   `test_voc136_caller_replacement.py`, and `test_voc137_pr_sha_scan.py`
   when those files still hard-code a live reviewer binding. Do not rewrite
   VOC-142, VOC-141, VOC-140, or VOC-117 package records. Reconcile every
   live caller pin-lock/current-pin assertion and mirrored-file hash table
   with the new merge. Preserve issue-era `AUTHORITATIVE_PIN` values and
   historical package evidence.
7. Confirm no OpenAI route is added. Confirm retry caps and exact-SHA
   review wiring are unchanged. Confirm no VOC-112 recapture or #1120
   resume helper was added. Confirm no fabricated-status helper, bypass-actor
   addition, or App-token mint change was added.
8. Record in `t00-evidence.md` the implementation PR base, new infra merge,
   authorized path, binding table, historical-versus-current split,
   negative-case results, source-search disposition, pin advance,
   validation commands after commit, and the feasible exact-head binding
   contract. Evidence must not require a commit to contain its own SHA.
9. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
   - targeted VOC-145 binding, historical-assertion, pin, and safety cases;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates the authorized path, six exact current bindings,
     historical-versus-current VOC-117 split, README/CHANGELOG/fixture
     reconciliation, pin advance not equal to `d8720829…`, unchanged
     retry/exact-SHA/fail-closed controls, and exclusion of #1120.
10. This package's caller PR `Closes` only its own VOC-145 task issue. Do
    not snapshot the current develop/main gap. After the exact reviewed
    caller merge, ordinary later promotion uses existing release
    evaluation. Root issue #1124 closes only after allowlisted metadata
    from a successful implement/release path exists. Do not create a
    duplicate promotion PR or release audit. Do not resume #1120.

11. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider
  addition, or monitor-inventory changes.
- Weakening retry caps, exact-SHA independent review, provider isolation,
  or fail-closed model resolution.
- Adding an OpenAI execution route or requesting `OPENAI_API_KEY`.
- Pinning unauthorized head `d8720829…`.
- Rewriting VOC-142, VOC-141, VOC-140, or VOC-117 package records.
- Fetching, hydrating, or recapturing VOC-112 JSON fixtures.
- Using this package as a carrier for issue #1120.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- A supervised bootstrap exception.
- Splitting restore, pin, tests, docs, or release reconciliation into
  separate tasks.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the infra contract, caller pin, authorized bindings,
  historical-versus-current tests, docs, evidence, and release handoff are
  one reconciliation outcome.
- Unauthorized head `d8720829…` is already on infra `main` and is the
  defective live contract; consume a new independently reviewed infra merge
  inside this same task. Do not treat that head as the pin.
- Path A versus Path B is an adoption record, not a second task.
- Issue #1120 is incident-adjacent only as an exclusion; it is not a second
  VOC-145 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
