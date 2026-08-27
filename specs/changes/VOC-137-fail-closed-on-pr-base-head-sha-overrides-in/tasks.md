# VOC-137 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1083 is one
scanner-correction outcome: remove the filename heuristic from PR base/head
SHA override detection, keep source-safe construction, add the required
arbitrary-filename negative cases and benign controls, freeze the #167 pin
and fixture bytes, preserve VOC-136 audit records, and let ordinary release
continue after the exact reviewed merge. Workflow-versus-tests-versus-docs,
VOC-136-versus-VOC-137 file split, and scanner-versus-reproduction are not
split reasons.

Cross-repo note: infrastructure PR KARSIFT/karsift-ai-infra#167 is already
merged as `b263c0c110591cc798b89277dfc35542abb1597b` and already pinned by
VOC-136. T00 does not open an infra PR and does not re-pin. Do not treat the
untracked local `karsift-ai-infra/` checkout (if present) as this repo's
tracked tree. No bootstrap exception. T00's first run is attempt `1` on a new
VOC-137 carrier from current `develop`. Do not reuse PR #1080 as this
package's review or implementation source.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`).

## VOC-137-T00 — Fail closed on PR base/head SHA overrides in every scannable caller executable

- Requirement source: issue #1083; `VOC-137-D00` through `VOC-137-D14`
- Acceptance criteria: `VOC-137-AC-00` through `VOC-137-AC-09`
- Tests: `VOC-137-TEST-00` through `VOC-137-TEST-12`
- Evidence: `VOC-137-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1083 evidence in `t00-evidence.md` (VOC-136 PR #1080,
   exact reviewed head `5d0c2350ab9a20ace586eaadd1169203140ffad0`, develop
   merge `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`, filename heuristic,
   `scripts/arbitrary-wrapper.sh` counterexample, supervisor comment
   5444745877, reviewer comment 5444825638, post-merge audit comment
   5444861382, canceled release runs `33113425829` / `33113547909`, draft
   release PR #1082, issue-creation develop
   `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`, issue-creation main
   `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`, live pin
   `b263c0c110591cc798b89277dfc35542abb1597b`). Do not copy raw provider
   responses, credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. Retain protected
   comparison anchor `b9e74fc2db4691c48c637639b265d527de9f4505` for the eight
   VOC-112 no-change paths. This package's own plan/adoption/roster commits
   after issue-creation develop `0cee20c…` do not count as protected-file
   drift.
3. From current `develop`, create a new VOC-137 implementation branch. Do not
   reuse, merge, cherry-pick, or rewrite PR #1080. Do not rewrite
   `specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/`.
4. Remove the filename gate around `PR_SHA_SET_PATTERN` in
   `tooling/governance/tests/voc136_bypass_scan.py`. Apply the pattern to
   every `should_scan_path` / non-excluded executable, including `*.py`,
   `*.sh`, `*.mjs`, `*.js`, `scripts/**`, and `package.json`. Reconstruct the
   pattern from `_PR_BASE_ENV` / `_PR_HEAD_ENV` (or equivalent non-contiguous
   parts). Do not add `tooling/governance/tests/` to `SCAN_EXCLUDE_PREFIXES`.
5. Make existing contiguous literals that would match the widened scan
   source-safe, including `test_voc136_caller_replacement.py`'s
   `export PR_BASE_SHA=` fixture assertion (`VOC-137-D14`). Do not skip or
   xfail VOC-136 scanner tests.
6. Add deterministic negative unit cases that reconstruct contiguous
   forbidden payloads only in memory and reject at least the four classes in
   `VOC-137-D03`, using arbitrary filenames that do not contain
   `validate-workspace` and do not end in `.test.mjs`. Include a direct
   reproduction of the issue payload at `scripts/arbitrary-wrapper.sh`.
7. Add positive controls for benign discussion, source-safe pattern
   construction, and the excluded fixture `run-app-checks.sh`. Assert
   `should_scan_path` is false for that fixture path and true for the
   arbitrary wrapper names and for added/changed `*.py` outside the mirror.
8. Confirm `PINNED_SHA.txt` still equals
   `b263c0c110591cc798b89277dfc35542abb1597b` and that no fixture-mirror path
   appears in the implementation diff. Confirm the eight VOC-112 no-change
   paths remain byte-identical to `b9e74fc2…`. Confirm fixture `roles.yml` is
   unchanged and no OpenAI route is added.
9. Record in `t00-evidence.md` the implementation PR base, the scan-scope
   change, negative-case and positive-control results, pin freeze, validation
   commands after commit, and the feasible exact-head binding contract.
   Evidence must not require a commit to contain its own SHA.
10. Run applicable validation **after the regression is tracked and
    committed** and record results in `t00-evidence.md`:
    - `bash scripts/governance/validate-governance.sh` with exact base/head;
    - `bash scripts/governance/classify-change-risk.sh` with exact base/head
      (expect R4);
    - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
    - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
    - targeted VOC-137 scanner cases including the arbitrary-wrapper
      reproduction;
    - `git diff --check`;
    - hosted required checks and independent exact-revision review that
      explicitly evaluates the arbitrary-filename shell/Node/Python cases and
      benign controls.
11. This package's caller PR `Closes` only its own VOC-137 task issue. Do not
    manufacture a VOC-136 completion marker. Do not snapshot the current
    develop/main gap.
12. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch. Do not
    treat PR #1080's review as sufficient for this correction.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes.
- Re-implementing VOC-136 or re-mirroring infrastructure #167.
- Opening a replacement infra PR.
- Rewriting VOC-136 package records or manufacturing a VOC-136 completion
  marker.
- Reusing PR #1080 review verdicts as this package's independent review.
- Changing any of the eight VOC-112 no-change paths.
- Adding a capture-commit fetch helper, hydrate helper, provenance-mode
  wrapper, or test-time evidence mutation under any filename.
- Excluding `tooling/governance/tests/**` or another executable directory
  wholesale from the complete-diff scan.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or promoting "current develop"
  to `main` as this package's work.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting scanner logic, tests, docs, or release reconciliation into
  separate tasks.
- Editing live caller workflows.
- Cleaning unrelated VOC-128 work or user worktrees.
- Operator-owned live evidence contracts.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the scanner correction, source-safe literals, negative and
  positive cases, pin freeze, evidence, and release handoff are one repair
  outcome.
- Infrastructure #167 is already merged and already pinned; consume that
  freeze. Do not wait on a new infra PR.
- VOC-136 is already merged on `develop`; its promotion waits on this
  correction and then proceeds through ordinary release. Do not add a second
  VOC-137 task to wait on that event.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
