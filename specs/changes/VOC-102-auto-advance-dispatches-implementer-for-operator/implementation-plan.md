# VOC-102 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `karsift-ai-infra` `auto-advance.yml` dispatch gate; implementer
  least-privilege; release check-completion boundary; VOC-097 live-evidence
  contracts and reconcile path; calling-repo `pipeline.yml` only for pin bump.
- Prerequisites: VOC-097 ownership contract path and reconcile mechanism exist;
  issue #863 records the spurious-dispatch incident; this draft is adopted under
  A-004.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The implementer
opens PRs there for that behavior; this package authorizes the required outcome.
Do not treat an untracked local `karsift-ai-infra/` checkout as this repository's
tracked tree. Calling-repo doc/pin/test changes land here under the same package.

## File reconciliation and implementation sequence

### T00 — Ownership gate, fail-closed, docs, deterministic tests

| File / area | Action | Notes |
|-------------|--------|-------|
| `karsift-ai-infra/.github/workflows/auto-advance.yml` | modify | Read next-task ownership before `should_dispatch=true` |
| karsift-ai-infra helper (optional extract) | create/modify | Prefer small pure function for contract parse/classify if it keeps YAML readable |
| karsift-ai-infra README | modify | Document skip vs dispatch |
| karsift-ai-infra / calling-repo tests | create/modify | `voc102-*.test.mjs` and/or infra self-ci fixtures |
| `docs/operations/live-evidence.md` and/or AGENTS.md | modify if needed | Only if text would remain false |
| calling-repo `.github/workflows/pipeline.yml` | modify only if pin required | Consume fixed auto-advance |
| `specs/changes/VOC-102-.../t00-evidence.md` | create | Commands + results |

Ordered steps:

1. Confirm drafting-time diagnosis against current `auto-advance.yml` HEAD; record
   that ownership is not consulted today (no secrets).
2. After next-task resolution and existing open-issue / existing-PR guards, load
   `<package_path>/.karsift/live-evidence/<next_task_id>.yaml` when present.
3. If ownership is `operator` or `live-actions`: set `should_dispatch=false`; leave
   issue open; emit sanitized waiting signal (`VOC-102-D02`); do not call
   `implement.yml`.
4. If no contract and no contradictory operator declaration: preserve today's
   `should_dispatch=true` path.
5. If metadata malformed/contradictory/unrecognized: fail closed — no dispatch +
   sanitized escalation (`VOC-102-D04`).
6. Ensure last-task / no-next-task still no-ops toward release (no early release).
7. Land deterministic positive / negative / malformed / regression tests.
8. Align docs that would otherwise claim universal implement dispatch.
9. Pin-bump calling `pipeline.yml` only if required; record consumption in evidence.
10. Run applicable tests and governance validation; write `t00-evidence.md`.

### T01 — Controlled sanitized workflow proof

| File / area | Action | Notes |
|-------------|--------|-------|
| `specs/changes/VOC-102-.../t01-evidence.md` | create | Metadata-only live proof |
| `.karsift/live-evidence/VOC-102-T01.yaml` | already drafted | Operator-owned contract |

Ordered steps:

1. Ensure T00 is live on the branch auto-advance executes from (expected after
   infra merge + calling-repo pin if any).
2. Use this package's T00→T01 advance (and/or a controlled fixture) so next task is
   operator-owned; record scrubbed auto-advance / pipeline metadata showing
   `should_dispatch=false` and **zero** `implement.yml` jobs for that next task.
3. Prove ordinary implementation next-task dispatch still occurs (fixture and/or
   sanitized observation of a non-operator next task). Prefer deterministic
   fixture if a second live package advance is unnecessary.
4. Confirm T01 issue remained open and waiting signal present; do not manufacture
   unrelated package evidence; no secrets in `t01-evidence.md`.
5. Complete acceptance via dedicated reconcile/evidence path — not by dispatching
   the general implementer.

## Validation and independent verification

Deterministic (T00):

```bash
# Exact commands depend on where tests land; record actual invocations in evidence.
node --test scripts/foundation/voc102-*.test.mjs
# and/or karsift-ai-infra self-ci for auto-advance ownership fixtures
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
- **Rollout:** Infra auto-advance change merges and is consumed by calling
  `pipeline.yml`; then T01 controlled proof.
- **Rollback trigger:** Ordinary implementation tasks stop auto-advancing; or
  operator-owned next tasks again receive implementer dispatch; or release opens
  while an operator task is still open.
- **Rollback mechanism:** Revert the auto-advance ownership-gate commit(s) on the
  infra default branch and any calling-repo pin; re-promote through normal paths.
- **Last-known-good:** commit before T00 merge on each affected repo.
