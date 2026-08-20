# VOC-097 — Make live-evidence tasks operator-owned and self-reconciling: Specification

## Objective and requirement source

Stop stranding evidence-only tasks that require a live GitHub Actions run after the
underlying system is healthy, as recorded in
[GitHub issue #823](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/823).

Today the implementer cannot dispatch or inspect qualifying Actions runs
(least-privilege by design). Independent review correctly rejects pending live
evidence, but `remediate.yml` treats that rejection like an implementation defect:
it consumes the bounded remediation retry and the implementer often expands into
unrelated pipeline or task-specific workflow edits. Valid later live evidence then
exists while the PR is polluted or permanently stuck.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time grounding:

| Item | Current state |
|------|---------------|
| Implementer Actions access | Intentionally least-privileged; no general Actions dispatch/inspect credentials (issue #823; implement.yml App token for PR/issue mutations only) |
| Remediation trigger | `remediate.yml` retries on posted `VERDICT: FAIL`, CI failure, or review-job error — no waiting-for-operator-evidence state |
| Stranded VOC-093-T01 | Issue #779 / PR #789 — production route-sweep live evidence pending |
| Stranded VOC-094-T01 | Issue #785 / PR #791 — live deploy-verification evidence pending |
| Related precedent | VOC-088 operational-failure observer is a separate sanitized failure-to-issue path; must stay separate from live-evidence waiting |
| Cross-repo ownership | Lifecycle behavior belongs in `KARSIFT/karsift-ai-infra` reusable workflows; this package authorizes the outcome (VOC-080 / VOC-043 pattern) |

## Scope and non-goals

In scope:

1. Model **operator/live-evidence waiting** as an explicit lifecycle state distinct
   from implementation failure and from review FAIL that means "code/docs wrong."
2. Keep the implementer least-privileged; do **not** grant general GitHub Actions
   credentials to the implementer role.
3. Add repository-controlled automation that may dispatch or observe **only** the
   workflows and jobs explicitly declared by the governed task/package evidence
   contract.
4. Collect **only** allowlisted metadata needed for acceptance: workflow identity,
   event, branch, exact SHA, run/job IDs, conclusion, timestamps, and bounded
   duration. Never copy logs, artifacts, secrets, OAuth data, sessions, cookies,
   tokens, user identifiers, or arbitrary output.
5. When qualifying evidence arrives, reconcile/wake the governed task path and
   require a **new exact-SHA independent review** before merge.
6. Fail closed on ambiguous workflow identity, wrong branch/SHA lineage, stale
   runs, missing jobs, non-success conclusions, or malformed evidence contracts.
7. While the only unmet condition is operator-owned live evidence: do **not** spend
   remediation retries and do **not** permit unrelated code/workflow edits as a
   "fix."
8. Bounded timeout/escalation and deduplication so a missing live run cannot loop
   indefinitely.
9. Deterministic tests covering waiting, successful reconciliation, timeout, wrong
   workflow/job/branch/SHA, duplicate events, sanitization, and
   no-credential/no-log invariants.
10. Reconcile stranded tasks #779 (`VOC-093-T01`) and #785 (`VOC-094-T01`) using
    the governed mechanism or an explicitly documented safe migration path.
11. Update governance and operator documentation so task authors can declare
    live-evidence ownership and an allowlisted evidence contract.
12. Preserve existing risk classification, exact-SHA review, branch protection,
    App-authenticated mutations, deployment isolation, and release behavior.

Non-goals / explicitly excluded:

- Granting the Cursor/implementer agent Actions API credentials or broadening its
  token scopes to general `actions: write` / workflow dispatch.
- Changing production application behavior, signup policy, secrets, databases, or
  Kuma/synthetic inventory definitions for availability monitoring.
- Folding live-evidence waiting into `operational-failure-monitoring.yml` or Sentry
  ingestion (must remain separate mechanisms).
- Weakening remediation for genuine code/CI failures.
- Snapshot-then-drift task patterns that invalidate themselves (karsift-ai-infra#15).
- Self-adoption / self-authorization of this package.
- Manual cleanup of unrelated out-of-scope edits already present on stranded PR
  branches except as required to put #779/#785 onto the governed waiting/reconcile
  path (T04).

## Risk and protected areas

- **Draft package proposal:** **R3** (CI/CD / agent-authority / remediation lifecycle;
  AGENTS.md and workflow caller wiring if touched).
- **Measured path floor at drafting:** **R3** for `.github/workflows/`, `AGENTS.md`,
  and related governance docs. Not proposed as R4 unless a task amends authority
  models or `docs/governance/amendments/*` (out of scope).
- Protected areas: karsift-ai-infra reusable workflows (`remediate.yml`,
  `review.yml`, `merge-gate.yml`, `implement.yml`, new reconcile workflow),
  calling-repo `pipeline.yml` if rewired, App-token mutation paths.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

The `risk: R3` value in `change.yaml` is a **draft proposal for the reviewing human
at adoption time, never a determination**. The path-based classifier and independent
verifier govern each task PR.

## Decisions

`VOC-097-D00`: Missing operator-owned live Actions evidence is **not** an
implementation defect. It must use an explicit waiting lifecycle state so
`remediate.yml` does not consume retries and the implementer does not expand scope
to "create" evidence via unrelated workflow edits.

`VOC-097-D01`: The implementer remains least-privileged. Repository-controlled
automation (App-authenticated workflows with narrowly scoped permissions) is the
only path that may dispatch or observe declared workflows/jobs for live evidence.

`VOC-097-D02`: Evidence payloads are allowlisted metadata only. Logs, artifacts,
secrets, OAuth/session/cookie/token material, user identifiers, and arbitrary job
output are forbidden in issues, PR comments, evidence files, and reconcile artifacts.

`VOC-097-D03` (proposed default; confirm at adoption — DEP-03): Each live-evidence
task declares a machine-readable **evidence contract** under the package (preferred:
`<package>/.karsift/live-evidence/<task_id>.yaml` or an equivalent field in the
task's adopted roster metadata) naming:

- `ownership: operator` (or `live-actions`)
- allowlisted `workflow_file` and/or stable workflow `name` / `workflow_id`
- allowlisted `job_names` (optional but required when the workflow has multiple jobs)
- required `event` set (e.g. `push`, `workflow_dispatch`, `schedule`)
- required `branch` / SHA lineage rule (e.g. exact PR head SHA, or integration tip
  that is an ancestor of the evidence run's head SHA)
- success `conclusion` requirement (`success` only unless explicitly documented)
- optional `max_age` / staleness bound
- optional `dispatch` block (workflow + inputs) when repo automation may trigger the
  run; otherwise observe-only

Human-readable guidance also appears in `tasks.md` / operator docs so authors know
what to declare. Exact file shape is an open question if implementers find a
clearer existing karsift convention.

`VOC-097-D04`: Qualifying evidence transitions waiting → ready-for-re-review. Merge
still requires a **new** independent verification comment bound to the **exact**
post-reconcile head SHA. Prior PASS on a SHA that lacked evidence does not carry
forward.

`VOC-097-D05`: Fail closed on ambiguous or non-matching evidence. Rejected evidence
does not wake the task and does not start remediation as if code failed; it leaves
the task waiting (or escalates on timeout) with a sanitized rejection reason.

`VOC-097-D06`: Bounded timeout and escalation: a waiting task that never receives
qualifying evidence within the configured bound escalates to the authority/task
issue with a sanitized summary and stops automatic looping. Duplicate
workflow_run / reconcile events are idempotent (at most one successful wake per
qualifying run identity).

`VOC-097-D07`: Live-evidence lifecycle signals stay separate from
`operational-failure-monitoring.yml` and Sentry. Emit only boolean/state/count-style
signals plus sanitized run identifiers — no application alerts for expected waiting
policy outcomes.

`VOC-097-D08`: Stranded #779 and #785 are in scope for reconciliation after the
mechanism exists (T04). Prefer governed waiting + qualifying observe/reconcile; if a
task PR is already polluted with out-of-scope edits, the safe migration path is
documented explicitly (reset to last in-scope revision or open a clean evidence-only
follow-up on the same task issue) without granting implementer Actions credentials.

## Open questions for the reviewing human

1. Confirm `VOC-097-D03` file/field shape for the evidence contract, or name an
   alternate existing karsift convention to reuse.
2. Confirm default waiting timeout (proposal: 72 hours wall clock from entering
   waiting, with a single escalation comment and no automatic third-party dispatch
   loops) and whether timeout escalation opens a new issue or only comments the
   existing task issue.
3. Confirm proposed **R3**, or raise in writing if changing remediation authority is
   treated as R4.
4. Confirm caller consumption of karsift-ai-infra `@main` vs pin bump required in
   this repository's `pipeline.yml` before T04/T05 claim live effect.
5. For T04: if #779/#785 PR branches contain irreversible out-of-scope history,
   is a clean evidence-only replacement PR on the same task issue acceptable?

## Data, migrations, analytics, and accessibility

- No application schema migration.
- No database mutation.
- No product UI change — evidence-backed non-applicability.
- No analytics change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.
