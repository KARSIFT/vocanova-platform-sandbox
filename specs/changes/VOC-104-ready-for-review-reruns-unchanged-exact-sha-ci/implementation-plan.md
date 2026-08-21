# VOC-104 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: merge-gate draft blocking; App-only verdict trust; exact-SHA
  stale-run guards; independent implementer/reviewer separation; live-evidence
  attestation when required; calling-repo `pipeline.yml` reuse wiring.
- Prerequisites: caller already lists `ready_for_review`; merge-gate already
  blocks drafts and rechecks base/head; issue #872 records the duplicate-path
  incident; this draft is adopted under A-004.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The implementer
opens PRs there for that behavior; this package authorizes the required outcome.
Do not treat an untracked local `karsift-ai-infra/` checkout as this repository's
tracked tree. Calling-repo pipeline/doc/test changes land here under the same
package.

## File reconciliation and implementation sequence

### T00 — Reuse policy, fail-closed path, docs, deterministic tests

| File / area                                                                        | Action        | Notes                                                                       |
| ---------------------------------------------------------------------------------- | ------------- | --------------------------------------------------------------------------- |
| karsift-ai-infra ready_for_review reuse helper                                     | create/modify | Pure decision output from Actions/check/comment metadata                    |
| karsift-ai-infra pipeline template + ci/review/plan-review/merge-gate coordination | modify        | Skip CI/model review only on safe ready_for_review reuse                    |
| karsift-ai-infra README + policy tests                                             | modify/create | Document reuse-vs-full-path; D02 matrix                                     |
| calling-repo `.github/workflows/pipeline.yml`                                      | modify        | Consume decision; add read-only exact-head proof action                     |
| `scripts/foundation/voc104-*.test.mjs` and/or infra self-ci                        | create/modify | Caller wiring + policy fixtures                                             |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` §17.3    | modify        | Clarify fresh evaluation may reuse exact-SHA evidence or take the full path |
| `specs/changes/VOC-104-.../t00-evidence.md`                                        | create        | Commands + results                                                          |

Ordered steps:

1. Confirm drafting-time diagnosis against current caller pipeline and reusable
   workflows: `ready_for_review` still starts full `ci` and review/plan-review
   even when base/head are unchanged (no secrets).
2. Implement the reusable, read-only eligibility workflow/helper and caller job
   resolved in `VOC-104-D01`; enforce `VOC-104-D02` once in that boundary.
3. Wire caller template and this repo's `pipeline.yml` so safe reuse skips CI and
   model review while merge-gate still runs and remains reachable when review
   siblings are skipped (`VOC-104-D03`).
4. On any D02 failure, select the normal full path (`VOC-104-D04`).
5. Preserve draft never auto-merges and App-only verdict trust (`VOC-104-D00`,
   `D05`, `D06`).
6. Cover both `agent/` and `plan/` per resolved `VOC-104-D10`.
7. Land deterministic shared-infra and calling-repo fixture tests (`VOC-104-D07`).
8. Update infra README and calling-repo DOC-15 §17.3 to distinguish a fresh
   readiness evaluation from an unconditional full CI/model-review replay.
9. Record the current `@main` reusable-workflow consumption (no pin bump
   expected); if repository state differs at implementation time, reconcile it
   explicitly.
10. Add a manually dispatched, read-only proof job to `pipeline.yml` that accepts
    only allowlisted source run metadata, runs on the proof PR ref, reads
    Actions/check metadata but never logs or artifacts, and verifies the
    ready_for_review reuse shape. It has no write, model, deploy, or
    application-secret path.
11. Run applicable tests and governance validation; write `t00-evidence.md`.

### T01 — Controlled draft-to-ready optimized-path proof

| File / area                                 | Action          | Notes                    |
| ------------------------------------------- | --------------- | ------------------------ |
| `specs/changes/VOC-104-.../t01-evidence.md` | create          | Metadata-only live proof |
| `.karsift/live-evidence/VOC-104-T01.yaml`   | already drafted | Operator-owned contract  |

Ordered steps:

1. Ensure T00 is live on the branch the caller pipeline executes from (expected
   after infra merge + calling-repo wiring consumed from `@main`).
2. Prepare or select a controlled draft PR whose exact base/head already has green
   required checks and a trusted App PASS; mark it ready_for_review.
3. Record scrubbed metadata showing the CI reuse marker passed, full application
   validation and model review were skipped, and merge-gate re-evaluated on that
   ready_for_review run.
4. Prove unsafe cases still take the full path via deterministic fixtures (preferred)
   or sanitized observation.
5. Commit allowlisted source metadata to the proof/carrier branch, then manually
   dispatch `verify-ready-for-review-reuse` on the exact PR head. Its caller job
   plus reusable inner job MUST emit the contract's exact display name
   `verify-ready-for-review-reuse / verify`.
6. Reconcile the successful exact-PR-head verifier run through the dedicated
   live-evidence path — never through the general implementer.

## Validation and independent verification

Deterministic (T00):

```bash
# Exact commands depend on where tests land; record actual invocations in evidence.
node --test scripts/foundation/voc104-*.test.mjs
# and/or karsift-ai-infra self-ci for ready_for_review reuse fixtures
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Live (T01): operator-owned; see live-evidence contract. Independent verifier binds
the exact task PR SHA (for T00) and confirms T01 metadata evidence without treating
missing live proof as a code defect to "fix" via unrelated pipeline edits
(VOC-097).

## Deployment and rollback

- **Authorization:** Package adoption + task implementation authorization only;
  this package does not itself authorize production application deployment.
- **Rollout:** Infra reuse-policy change merges and is consumed by calling
  `pipeline.yml`; then T01 controlled proof.
- **Rollback trigger:** Safe unchanged ready_for_review events again run full CI
  and model review; or unsafe events incorrectly skip verification; or draft PRs
  become mergeable.
- **Rollback mechanism:** Revert the reuse-eligibility commit(s) on the infra
  default branch and any calling-repo wiring; re-promote through normal paths.
- **Last-known-good:** commit before T00 merge on each affected repo.
