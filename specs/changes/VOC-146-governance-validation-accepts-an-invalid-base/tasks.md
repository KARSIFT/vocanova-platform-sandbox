# VOC-146 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1127 is one
fail-closed range-loading repair: unresolved or invalid `--base`/`--head`
must make governance validation return nonzero before it claims success, the
same loader in `classify-change-risk.sh` must fail closed, tests and
current-state docs land with the scripts, and ordinary later promotion is
evidence of the outcome. Script count, wrapper-versus-classifier, and
tests-versus-docs are not split reasons.

No coordinated `karsift-ai-infra` PR is required. Do not treat the untracked
local `karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
No bootstrap exception. T00's first run is attempt `1` on a new VOC-146
carrier from current `develop`.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Emergency PR #1126 is discovery
context only; it is not this package's implementation work. A VOC-097
live-evidence second task is not used.

## VOC-146-T00 — Fail closed on an unresolved or invalid `--base`/`--head` range in governance validation and risk classification

- Requirement source: issue #1127; `VOC-146-D00` through `VOC-146-D15`
- Acceptance criteria: `VOC-146-AC-00` through `VOC-146-AC-08`
- Tests: `VOC-146-TEST-00` through `VOC-146-TEST-07`
- Post-T00 release/closure verification: `VOC-146-TEST-08` (not an
  implementation-task close or retry prerequisite)
- Evidence: `VOC-146-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1127 evidence in `t00-evidence.md` (reproduction head
   `79b2b3f1f4224235bdda3f77ee887c3004978deb`, nonexistent `--base`
   `376e00dd769afb0fe850052b3a5cb48f729e73ad`, Git fatal symmetric-difference
   error, `Governance structure validation passed.`, exit `0`,
   `mapfile < <(git diff … "$base...$head")` in
   `validate-monitoring-impact.sh`, identical pattern in
   `classify-change-risk.sh`, discovery during local exact-range verification
   of emergency PR #1126). Do not copy raw provider responses, credentials,
   or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. This package's own
   plan/adoption/roster commits after issue-creation reproduction commit
   `79b2b3f1…` do not count as protected-file drift. Do not edit
   `specs/changes/VOC-086-…/` or `specs/changes/VOC-112-…/`.
3. Replace the process-substitution changed-file loaders. When both `--base`
   and `--head` are supplied, resolve each as a commit object, then capture
   `git diff --no-renames --name-only --diff-filter=ACDMRTUXB "$base...$head"`
   through a status-preserving path. Fail closed on nonexistent base,
   nonexistent head, no-merge-base, and partial `--base`/`--head`. Do not
   treat a successful empty diff as an invalid range. Preserve `--files-from`,
   `--declarations-only`, VOC-086 missing-range fail-closed, and working-tree
   fallback when no range was requested. Apply the same contract to
   `classify-change-risk.sh`.
4. Add deterministic tests that call the live scripts. Cover the issue #1127
   nonexistent-`--base` reproduction, nonexistent `--head`, unrelated/no-
   merge-base revisions, partial range, valid PR range, `--files-from`, and
   classifier parity. Do not treat a grep-only assertion as covering #1127.
5. Exhaustively search tracked source and current docs. Update `AGENTS.md`
   and every current-state match so unresolved commits and invalid diff
   ranges are fail-closed, not only a missing range. Do not rewrite VOC-086
   or VOC-112 package records. Do not recapture VOC-112 fixtures. Do not
   advance `PINNED_SHA.txt`.
6. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no workflow rewrite unless exhaustive search found a live false
   claim. Confirm no snapshot-gap commit.
7. Record in `t00-evidence.md` the implementation PR base, range-loading
   change, negative-case results, source-search disposition, validation
   commands after commit, and the feasible exact-head binding contract.
   Evidence must not require a commit to contain its own SHA.
8. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head
     (valid implementation PR range; expect success);
   - the issue #1127 class command (expect nonzero; no success line);
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - targeted VOC-146 foundation tests and VOC-086 monitoring-impact tests;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates nonexistent base, nonexistent head, no-merge-base,
     status-preserving diff, preserved valid range and `--files-from`,
     classifier parity, and current-state docs.
9. This package's caller PR `Closes` only its own VOC-146 task issue. Do
   not snapshot the current develop/main gap. Root issue #1127 closes only
   after allowlisted metadata from a successful implementation/promotion
   path exists.

10. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  `roles.yml`, or monitor-inventory changes.
- Changing `monitoring_impact` declaration semantics or inventory files.
- Recapturing VOC-112 JSON fixtures or editing emergency PR #1126.
- Opening a `karsift-ai-infra` PR or advancing `PINNED_SHA.txt`.
- Rewriting VOC-086 or VOC-112 package records.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, or retry caps.
- Splitting wrapper, classifier, tests, or docs into separate tasks.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the two scripts share one fail-closed range contract, and
  tests, docs, evidence, and release handoff complete that same outcome.
- Emergency PR #1126 / VOC-112 is discovery context, not a second VOC-146
  task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
