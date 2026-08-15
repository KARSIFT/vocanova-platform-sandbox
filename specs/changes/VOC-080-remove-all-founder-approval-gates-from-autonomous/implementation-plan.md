# VOC-080 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted under **current**
A-003 / VOC-075 authority (`status: adopted`, `approval_status` per house
convention, `implementation_authorized: true` /
`implementation.authorized: true`) and `VOC-080-DEP-00`–`DEP-04` are
resolved or explicitly deferred in writing.

Additionally:

- Proposed package risk is **R4**. Under pre-transition rules, founder
  adoption of this package is required before task dispatch.
- Protected areas: `docs/governance/amendments/`, transition-state YAML,
  DOC-15, `scripts|tooling/governance/`, `AGENTS.md` / `CLAUDE.md`,
  `.github/workflows/pipeline.yml`, and cross-repo
  `KARSIFT/karsift-ai-infra` reusable workflows.
- Do not mark the successor amendment effective until T07.
- Do not weaken independent verification, CI, or fail-closed risk parsing
  to make autonomy easier.
- Any workflow/settings behavior change must update every doc that
  describes that behavior in the same PR (or immediate settings follow-up
  PR per AGENTS.md).

## File reconciliation and implementation sequence

1. **`VOC-080-T00`** — Author A-004 (or settled vehicle) + inactive
   transition scaffolding; preserve A-003/VOC-075 history.
2. **`VOC-080-T01`** — Infra merge-gate: R0–R4 auto-merge on gates;
   unparseable fail-closed; remove founder override merge path; neutralize
   `automatic_merge_allowed` founder semantics per DEP-02.
3. **`VOC-080-T02`** — Autonomous adoption + idempotent reconcile
   `workflow_dispatch`; eliminate silent merged-as-draft.
4. **`VOC-080-T03`** — Release/remediate/deploy-path founder-gate removal;
   preserve auto-promote and fail-closed retries.
5. **`VOC-080-T04`** — Caller `pipeline.yml` + AGENTS/CLAUDE/DOC-15/16/
   matrices/templates/repository-settings reconciliation.
6. **`VOC-080-T05`** — Deterministic tests (may land with T01–T04).
7. **`VOC-080-T06`** — Sandbox/dry-run rehearsal evidence.
8. **`VOC-080-T07`** — Exact-revision founder approval (final) + activation
   markers + VOC-079 unblock note.

Preserve compatible autonomous-merge / auto-release work from VOC-012 and
the 2026-08-08 delegation. Do not rewrite historical package evidence.

## Validation and independent verification

Deterministic commands on this repository (every caller-repo task PR):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Infra repository: run that repo's installed self-ci / workflow tests for
touched reusable workflows (exact commands from
`KARSIFT/karsift-ai-infra` README / `self-ci.yml` at implementation time —
do not invent unavailable checks).

Missing rehearsal credentials or cross-repo access is a **limitation**, not
a pass for T06 live clauses.

Independent verification (per `CLAUDE.md`) must bind each exact commit SHA,
confirm the implementer-role occupant did not approve/merge its own work,
identify the **active** authority model (A-003 until T07; successor after),
and report every still-required R3/R4/EHR/adoption/activation gate. Before
T07, R4 founder transition approval remains required. After T07, verifier
must confirm founder-comment gates are gone while non-founder controls
remain.

## Deployment and rollback

Authorization: this package does **not** authorize production application
deployment by draft or adoption alone. After activation, existing
auto-promotion / `deploy-production.yml` on `main` push continues per
AGENTS.md 2026-08-08 delegation, now without residual founder-comment
retry-as-gate.

Rollout sequence:

1. Adopt/authorize VOC-080 under current authority; settle DEPs.
2. Land T00–T05 (infra then caller docs/wiring).
3. T06 rehearsal on sandbox/harness; record evidence.
4. T07 exact-revision founder approval + activation.
5. Resume VOC-079 and normal packages on the new path.

Rollback trigger: autonomy merges without verification; silent
unadopted merges recur; docs contradict workflows; activation flipped
without exact-revision founder approval; production deploy blocked by
undocumented residual reviewer settings; or independent review FAIL.

Rollback mechanism:

1. Revert caller-repo commits to last-known-good pre-VOC-080 tips for
   affected docs/workflows/transition-state.
2. Revert or pin `karsift-ai-infra` reusable workflows to the prior known
   good tag/SHA if `@main` already moved (prefer temporary pin over
   silent drift).
3. Restore pre-transition authority markers so founder gates return
   **only** as the documented rollback model — without rewriting audit
   history of what happened under the interim revisions.
4. Re-run governance validation; confirm merge-gate/adopt/release
   behavior matches the restored revision.

Accountable owner: named in T07 evidence (unassigned at drafting).
Last-known-good reference: pre-T01 `karsift-ai-infra` merge-gate/adopt/
release SHAs and this repo's pre-T04 AGENTS.md / pipeline.yml /
A-003-active transition-state tip.
