# VOC-106 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `karsift-ai-infra` `remediate.yml` / `decide-remediation.py`
  dispatch gate; implementer least-privilege; merge-gate fail-closed behavior;
  VOC-097 live-evidence contracts and reconcile path; calling-repo
  `pipeline.yml` verifier wiring.
- Prerequisites: VOC-097 ownership contract path exists; VOC-102 auto-advance
  ownership gate exists (separate path); issue #882 records the
  spurious-remediation incident; this draft is adopted under A-004.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package authorizes the
required outcome. Do not treat an untracked local `karsift-ai-infra/` checkout
as this repository's tracked tree. Calling-repo doc/pin/test changes land here
under the same package.

## File reconciliation and implementation sequence

### T00 — Ownership gate, fail-closed escalation, docs, deterministic tests

| File / area                                              | Action        | Notes                                                         |
| -------------------------------------------------------- | ------------- | ------------------------------------------------------------- |
| `karsift-ai-infra/config/decide-remediation.py`          | modify        | Accept ownership / contract inputs; emit escalate decisions   |
| `karsift-ai-infra/.github/workflows/remediate.yml`       | modify        | Load exact-head contract before RETRY; escalate operator path |
| karsift-ai-infra ownership/helper reuse                  | create/modify | Prefer reusing VOC-102 classifier primitives where safe       |
| karsift-ai-infra README                                  | modify        | Document ownership-gated remediation                          |
| karsift-ai-infra / calling-repo tests                    | create/modify | `voc106-*.test.mjs` and/or infra self-ci fixtures             |
| `docs/operations/live-evidence.md`                      | modify        | Document operator FAIL/CI escalation and ordinary retry      |
| AGENTS.md                                               | modify if needed | Only if current policy text would become false             |
| calling-repo `.github/workflows/pipeline.yml`            | modify        | Consume fixed remediate; add read-only exact-head proof action |
| `specs/changes/VOC-106-.../t00-evidence.md`              | create/update | Commands + results                                            |

Ordered steps:

1. Confirm drafting-time diagnosis against current `decide-remediation.py` /
   `remediate.yml` HEAD; record that ownership is not consulted before RETRY
   (no secrets).
2. After exact-head/base guards and PR-body task/package identity parsing, load
   `<package_path>/.karsift/live-evidence/<task_id>.yaml` from the exact reviewed
   head when present. Reuse or mirror VOC-102 structural ownership classification
   (contract authoritative; exact `Automation ownership` marker secondary; no
   prose inference).
3. If a valid contract establishes ownership as `operator` or `live-actions`, do
   **not** set `should_retry=true` and do **not** call `implement.yml`; emit a
   sanitized escalate-operator / equivalent outcome for review FAIL or CI
   failure per `VOC-106-D01/D02`.
4. Preserve existing `WAITING` suppression (no attempt consumed).
5. Preserve existing `STALE` and `REVIEW_INFRA_FAILURE` non-retry paths (no
   attempt consumed).
6. Route malformed, mismatched, or missing-required ownership metadata to the
   separate fail-closed path in `VOC-106-D04`.
7. If the task is ordinary (no contract and no contradictory operator
   declaration): preserve today's bounded RETRY path.
8. Ensure merge-gate behavior is untouched and remains fail-closed.
9. Land deterministic tests covering the AC-05 matrix.
10. Align infra/operator docs with ownership-gated FAIL/CI escalation, explicit
    stale/malformed handling, and retained ordinary bounded retry.
11. Record the current `@main` reusable-workflow consumption (no pin bump
    expected); if repository state differs at implementation time, reconcile it
    explicitly.
12. Add a manually dispatched, read-only proof job to `pipeline.yml`. It accepts
    only allowlisted source-run / PR identity inputs, runs on the T01 carrier
    ref, reads Actions/issue/PR metadata but never logs or artifacts, and
    verifies: the source remediate decision did not execute the reusable
    implement job for the operator-owned task; escalation metadata is present;
    ordinary implementer-owned retry remains covered by fixture/live-safe
    evidence. It has no write, model, deploy, or application-secret path.
13. Run applicable tests and governance validation; write `t00-evidence.md`.

### T01 — Controlled sanitized workflow proof

| File / area                                 | Action          | Notes                    |
| ------------------------------------------- | --------------- | ------------------------ |
| `specs/changes/VOC-106-.../t01-evidence.md` | create/update   | Metadata-only live proof |
| `.karsift/live-evidence/VOC-106-T01.yaml`   | already drafted | Operator-owned contract  |

Ordered steps:

1. Ensure T00 is live on the branch remediation executes from (expected after
   infra merge + calling-repo consumption of `@main`).
2. Perform a controlled operator-owned FAIL or CI-failure scenario on an
   evidence carrier (preferred: this package's own T01 carrier after T00
   creates/publishes it). Record scrubbed remediate / pipeline metadata showing
   no executed `implement.yml` job and sanitized escalation.
3. Prove ordinary implementer-owned FAIL still retries as intended
   (deterministic fixture preferred; sanitized live observation allowed if
   needed).
4. Confirm merge remained fail-closed for the failing head; do not manufacture
   unrelated package evidence.
5. Commit the allowlisted source metadata to the carrier, then manually dispatch
   the read-only `verify-remediate-operator-ownership` pipeline action on the
   exact carrier branch. Its caller job plus reusable inner job MUST emit the
   contract's exact display name `verify-remediate-operator-ownership / verify`.
6. Reconcile the successful exact-PR-head verifier run through the dedicated
   live-evidence path — never through the general implementer.

## Validation and independent verification

Deterministic (T00):

```bash
# Exact commands depend on where tests land; record actual invocations in evidence.
node --test scripts/foundation/voc106-*.test.mjs
# and/or karsift-ai-infra self-ci for remediation ownership fixtures
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Live (T01): operator-owned; see live-evidence contract. Independent verifier
binds the exact task PR SHA (for T00) and confirms T01 metadata evidence without
treating missing live proof as a code defect to "fix" via unrelated pipeline
edits (VOC-097).

## Deployment and rollback

- **Authorization:** Package adoption + task implementation authorization only;
  this package does not itself authorize production application deployment.
- **Rollout:** Infra remediate change merges and is consumed by calling
  `pipeline.yml`; then T01 controlled proof.
- **Rollback trigger:** Ordinary implementer-owned FAIL/CI failure stops
  auto-remediating; or operator-owned FAIL/CI failure again receives implementer
  dispatch; or merge becomes non-fail-closed.
- **Rollback mechanism:** Revert the remediate ownership-gate commit(s) on the
  infra default branch and any calling-repo pin; re-promote through normal paths.
- **Last-known-good:** commit before T00 merge on each affected repo.
