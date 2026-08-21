# VOC-105 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: CI and independent-review entry conditions; merge-gate draft
  and exact-SHA gates; trusted App publisher identity; live-evidence attestation
  binding; calling-repo `pipeline.yml` wiring.
- Prerequisites: issue #872 records the duplicate ready_for_review cost;
  draft-aware `ready_for_review` subscription already exists; this draft is
  adopted under A-004.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package authorizes the
required outcome. Do not treat an untracked local `karsift-ai-infra/` checkout
as this repository's tracked tree. Calling-repo wiring/docs/tests land here
under the same package.

## File reconciliation and implementation sequence

### T00 — Reuse gate, fail-closed semantics, docs, deterministic tests

| File / area | Action | Notes |
| ----------- | ------ | ----- |
| karsift-ai-infra ready_for_review reuse helper / workflow logic | create/modify | Deterministic reuse decision; fail closed |
| karsift-ai-infra `ci.yml` / `review.yml` / `plan-review.yml` and/or caller conditions | modify | Skip full CI/model review only when reuse permitted |
| karsift-ai-infra `merge-gate.yml` | modify only if required | Preserve draft block; still re-evaluate on ready |
| karsift-ai-infra README + policy tests | create/modify | Document safe vs refuse paths |
| calling-repo `.github/workflows/pipeline.yml` | modify | Wire skip/reuse conditions; add read-only verify action |
| `scripts/foundation/voc105-*.test.mjs` and/or infra self-ci | create/modify | Positive/negative/regression coverage |
| docs claiming universal full re-run | modify if needed | Only if text would remain false |
| `specs/changes/VOC-105-.../t00-evidence.md` | create | Commands + results |

Ordered steps:

1. Confirm drafting-time diagnosis against current caller `pipeline.yml` and
   reusable ci/review/plan-review entry conditions (no secrets).
2. Implement a deterministic reuse classifier that evaluates `VOC-105-D01`
   against live PR base/head, required checks, trusted App verdict, scope
   binding, and live-evidence attestation when required.
3. Wire the safe path so full CI and model review are skipped while merge-gate
   still runs for non-draft PRs.
4. Wire the refuse path to today's full CI + review + merge-gate behavior.
5. Preserve draft-never-merge and exact-SHA stale-run protections.
6. Ensure `opened` / `synchronize` / `reopened` still always take the full path.
7. Land deterministic positive, negative, human-comment rejection, draft, and
   synchronize regression tests (infra self-ci and calling-repo foundation).
8. Align infra/caller docs with the optimized and fail-closed paths; do not
   claim out-of-scope roots from issue #872 are fixed.
9. Record current `@main` reusable-workflow consumption (no pin bump expected);
   if repository state differs at implementation time, reconcile it explicitly.
10. Add a manually dispatched, read-only
    `verify-ready-for-review-reuse` proof job to `pipeline.yml`. It accepts only
    allowlisted source run IDs / PR number inputs, runs on the T01 carrier ref,
    reads Actions/PR metadata but never logs or artifacts, and verifies: prior
    green exact-SHA evidence existed; the ready_for_review run skipped full CI
    and model review; merge-gate re-evaluated; base/head were unchanged. It has
    no write, model, deploy, or application-secret path. Caller and reusable
    inner job MUST produce the exact contract job display name
    `verify-ready-for-review-reuse / verify`.
11. Run applicable tests and governance validation; write `t00-evidence.md`.

### T01 — Controlled draft-to-ready live proof

| File / area | Action | Notes |
| ----------- | ------ | ----- |
| `specs/changes/VOC-105-.../t01-evidence.md` | create | Metadata-only live proof |
| `.karsift/live-evidence/VOC-105-T01.yaml` | already drafted | Operator-owned contract |

Ordered steps:

1. After T00 is live on the branch the pipeline executes from, perform a
   controlled draft → ready transition with unchanged base/head (preferred:
   dogfood this package's own evidence-carrier path after T00 merges, or an
   equivalent controlled agent PR prepared for proof only).
2. Confirm the ready_for_review pipeline run skipped full CI and model review
   and still ran merge-gate re-evaluation.
3. Record allowlisted metadata only in `t01-evidence.md` (prior run IDs,
   ready_for_review run IDs, job conclusions, reuse decision). Never copy logs
   or secrets.
4. Commit those allowlisted source metadata to the carrier, manually dispatch
   the read-only proof action on `agent/voc-105-voc-105-t01`, and require its
   run HEAD to equal the current carrier PR head before reconciliation.
5. Do not expand scope into deprecated-action, Node-runtime, dependency-alert,
   or remediation-preflight work.

## Validation and independent verification

- Deterministic: infra self-ci / `scripts/foundation/voc105-*.test.mjs`;
  `bash scripts/governance/validate-governance.sh` and
  `bash scripts/governance/classify-change-risk.sh` for calling-repo governance
  paths as applicable; `git diff --check`.
- Exact-SHA independent verification on each implementable task PR.
- T01 acceptance requires operator-owned live evidence under
  `.karsift/live-evidence/VOC-105-T01.yaml`.

## Deployment and rollback

- Authorization: package adoption + per-task implementation authority only.
  This package does not authorize production application deployment.
- Rollout: infra merge to karsift-ai-infra default branch (caller consumes
  `@main`); calling-repo wiring via develop → main promotion.
- Rollback trigger: unsafe skip observed, draft merge regression, or stalled
  ready PRs that should have reused evidence.
- Rollback mechanism: revert T00 infra (and calling-repo wiring) through normal
  PR paths.
- Last-known-good: commit before T00 merge.
