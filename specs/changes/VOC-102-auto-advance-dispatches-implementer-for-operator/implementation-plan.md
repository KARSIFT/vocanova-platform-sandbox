# VOC-102 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `karsift-ai-infra` `auto-advance.yml` dispatch gate; implementer
  least-privilege; release check-completion boundary; VOC-097 live-evidence
  contracts and reconcile path; calling-repo `pipeline.yml` verifier wiring.
- Prerequisites: VOC-097 ownership contract path and reconcile mechanism exist;
  issue #863 records the spurious-dispatch incident; this draft is adopted under
  A-004.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The implementer
opens PRs there for that behavior; this package authorizes the required outcome.
Do not treat an untracked local `karsift-ai-infra/` checkout as this repository's
tracked tree. Calling-repo doc/pin/test changes land here under the same package.

## File reconciliation and implementation sequence

### T00 — Ownership gate, fail-closed, docs, deterministic tests

| File / area                                           | Action           | Notes                                                                         |
| ----------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------- |
| `karsift-ai-infra/.github/workflows/auto-advance.yml` | modify           | Read next-task ownership before `should_dispatch=true`                        |
| karsift-ai-infra classifier helper                    | create/modify    | Pure decision output; no mutation credentials                                 |
| karsift-ai-infra clean carrier publisher helper/job   | create/modify    | Deterministic pending evidence PR; App-scoped writes; no LLM or Actions-write |
| karsift-ai-infra README                               | modify           | Document skip vs dispatch                                                     |
| karsift-ai-infra / calling-repo tests                 | create/modify    | `voc102-*.test.mjs` and/or infra self-ci fixtures                             |
| `docs/operations/live-evidence.md` and/or AGENTS.md   | modify if needed | Only if text would remain false                                               |
| calling-repo `.github/workflows/pipeline.yml`         | modify           | Consume fixed auto-advance; add read-only exact-head proof action             |
| `specs/changes/VOC-102-.../t00-evidence.md`           | create           | Commands + results                                                            |

Ordered steps:

1. Confirm drafting-time diagnosis against current `auto-advance.yml` HEAD; record
   that ownership is not consulted today (no secrets).
2. After next-task resolution and the open-issue guard, load
   `<package_path>/.karsift/live-evidence/<next_task_id>.yaml` when present. Do not
   apply the existing-PR guard globally before ownership is known. Structurally
   parse only the exact matching `tasks.md` stanza for the allowlisted
   `Automation ownership` marker; never infer from prose.
3. If ownership is `operator` or `live-actions`: select
   `prepare-live-evidence`; leave the issue open; do not call `implement.yml`.
4. Run the separate clean publisher to create/reuse the task branch and draft
   evidence-carrier PR with only the governance-derived pending evidence path, and
   post one deduplicated sanitized waiting marker. The classifier stays read-only;
   the publisher alone mints the App for contents/issues/PR writes and receives no
   model key or Actions-write permission. Re-enter it when a deterministic carrier
   already exists so partial publication can repair a missing derived evidence file
   or marker; reject conflicting/untrusted PR state.
5. If no contract and no contradictory operator declaration: preserve today's
   existing-PR guard, then preserve today's `should_dispatch=true` path.
6. If metadata malformed/contradictory/unrecognized: fail closed — no dispatch,
   no carrier from untrusted paths, and one sanitized publisher escalation
   (`VOC-102-D04`).
7. Ensure last-task / no-next-task still no-ops toward release (no early release).
8. Land deterministic positive / negative / malformed / publisher-idempotency /
   permission-boundary / regression tests.
9. Align infra/operator/template docs with skip-vs-dispatch behavior and the exact
   task-stanza automation-ownership marker.
10. Record the current `@main` reusable-workflow consumption (no pin bump expected);
    if repository state differs at implementation time, reconcile it explicitly.
11. Add a manually dispatched, read-only proof job to `pipeline.yml`. It accepts
    only a source run ID and waiting PR number, runs on the T01 carrier ref, reads
    Actions/issue/PR metadata but never logs or artifacts, and verifies: the
    source was the expected `issues: closed` pipeline run on `develop`; the
    ownership decision did not execute the reusable implement job; T01 remains
    open; and exactly one matching carrier and waiting marker exist. It has no
    write, model, deploy, or application-secret path.
12. Run applicable tests and governance validation; write `t00-evidence.md`.

### T01 — Controlled sanitized workflow proof

| File / area                                 | Action          | Notes                    |
| ------------------------------------------- | --------------- | ------------------------ |
| `specs/changes/VOC-102-.../t01-evidence.md` | create          | Metadata-only live proof |
| `.karsift/live-evidence/VOC-102-T01.yaml`   | already drafted | Operator-owned contract  |

Ordered steps:

1. Ensure T00 is live on the branch auto-advance executes from (expected after
   infra merge + calling-repo pin if any).
2. Use this package's T00→T01 advance so the next task is operator-owned; record
   scrubbed auto-advance / pipeline metadata showing `prepare-live-evidence` and
   no executed `implement.yml` job for that next task. This operational run starts
   before it creates the carrier and is not falsely claimed as PR-head lineage.
3. Prove ordinary implementation next-task dispatch still occurs (fixture and/or
   sanitized observation of a non-operator next task). Prefer deterministic
   fixture if a second live package advance is unnecessary.
4. Confirm T01 issue remained open, the deterministic draft carrier PR exists,
   its pending evidence path exactly matches the strict task-ID-derived convention,
   and one waiting marker is present; do not manufacture unrelated package evidence.
5. Commit the allowlisted source metadata to the carrier, then manually dispatch
   the read-only `verify-auto-advance-live-evidence` pipeline action on the exact
   carrier branch. Its caller job plus reusable inner job MUST emit the contract's
   exact display name `verify-auto-advance-live-evidence / verify`.
   It must validate the source event and carrier state without logs/artifacts.
6. Reconcile the successful exact-PR-head verifier run through the dedicated
   live-evidence path — never through the general implementer.

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
