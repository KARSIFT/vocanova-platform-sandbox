# VOC-113 — Recover missing Actions checks after automated merges and release PR creation: Specification

## Objective and requirement source

Implement a durable, repository-managed recovery mechanism so App-driven task
merges into the integration branch and App-created promotion PRs always obtain
genuine exact-SHA Actions validation — even when GitHub does not emit the usual
event-triggered workflow runs.

**Requirement source:** [GitHub issue #948](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/948).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

### Confirmed problem evidence (issue #948)

| Item | Value |
|------|-------|
| Date | 2026-08-24 |
| Integration merge symptom | Final VOC-112 task PR merged to `develop`; no `push` workflows for the squash commit |
| Promotion symptom | Release automation opened PR #947; no pull-request workflows / required checks |
| Manual recovery attempts | Close/reopen and draft/ready transitions did not recreate checks |
| Existing reconcile | `reconcile-release` reused release audit and PR; no merge because exact-head required checks are absent |
| Active required checks | `governance-policy`, `validate`, `ci / ci` (must remain; must not be fabricated) |

## Scope and non-goals

### In scope

1. Determine and document the precise trigger/token/event behavior that produced
   the missing downstream runs (T00 evidence; fail closed on speculation).
2. Preserve GitHub App authentication for automated mutations (no regression to
   `github.token` merges/PR creation that silently suppress follow-on events).
3. Guarantee that the exact integration squash SHA and the exact promotion PR
   head SHA receive the normal required validation, independent review where
   applicable, and merge-gate / release evidence.
4. Prefer explicit reusable-workflow or `workflow_dispatch` orchestration when
   implicit event triggers are unreliable due to recursion/suppression.
5. Fail closed when evidence is absent, stale, bound to another SHA, pending past
   a bounded timeout, or ambiguous.
6. Prevent duplicate promotion PRs, duplicate release audits, workflow recursion,
   and fabrication of check runs or commit statuses.
7. Preserve the existing branch ruleset and governance risk classification.
8. Add deterministic tests for:
   - task-merge → integration SHA recovery, and
   - release-PR creation → exact-head check recovery.
9. Emit actionable, sanitized diagnostics (run IDs, SHAs, missing check names,
   timeout) without logs, secrets, tokens, or personal data.
10. Use open promotion PR #947 as the controlled live fixture; complete promotion
    only after genuine exact-head required checks succeed.
11. Verify post-promotion workflows on `main` and close the remediation only after
    that live evidence is complete.

### Non-goals / explicitly excluded

- Weakening or removing required ruleset checks.
- Synthesizing successful check runs / commit statuses.
- Changing product runtime behavior, credentials, signup policy, infrastructure
  topology, or branch-protection ruleset membership.
- Broader Actions efficiency improvements unrelated to missing-run recovery.
- Snapshot-then-drift commits that invalidate themselves (see planner prompt /
  karsift-ai-infra#15); package-directory bookkeeping commits must not count as
  promotion drift.
- Self-adoption / self-authorization of this package.
- Treating issue closure alone as completion proof.

## Risk and protected areas

- **Draft package proposal:** **R4** (CI/CD lifecycle orchestration plus a
  repository-governance exact-provenance gate correction).
- **Measured package floor at drafting:** **R4** because T00 must update
  `.github/workflows/repository-governance.yml` and
  `scripts/foundation/voc112-navigation-benchmark.test.mjs` without weakening
  exact-hash verification.
- Protected areas: App-token merge/release mutation paths, exact-SHA authoritative
  check selection (VOC-108), release converge idempotency, branch ruleset
  required contexts, production promotion/deploy wake path.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-113-D00`: Missing downstream Actions runs after a successful App-driven
mutation are an **actionable automation defect**, not an acceptable silent no-op.
Recovery must create or wake **genuine** workflow runs bound to the exact SHA.

`VOC-113-D01`: Recovery MUST NOT fabricate check runs, commit statuses, or
PASS verdicts. The ruleset contexts `governance-policy`, `validate`, and
`ci / ci` remain authoritative only when produced by their real workflows.

`VOC-113-D02`: GitHub App authentication for merge and promotion-PR mutation
MUST be preserved. Falling back to `github.token` for those mutations is
fail-closed forbidden when App credentials are configured (existing merge-gate
posture).

`VOC-113-D03`: When implicit `push` / `pull_request` triggers do not produce the
required runs, automation MUST use explicit reusable-workflow calls or
`workflow_dispatch` (or an equivalent repository-managed orchestration) so the
exact SHA still receives normal validation. Recursion guards must prevent the
recovery path from re-entering itself unboundedly.

`VOC-113-D04`: Evidence selection MUST reuse VOC-108 authoritative exact-head
rules: absent, stale, wrong-SHA, pending-past-timeout, cancelled, or latest-failed
evidence fails closed. `reconcile-release` may wake evaluation but cannot merge
without successful newest exact-head required checks.

`VOC-113-D05`: Idempotency invariants from VOC-043 / VOC-108 remain: at most one
open release audit and one open `develop`→`main` promotion PR per package /
promotion pair; duplicate recovery dispatches must be harmless.

`VOC-113-D06`: Recovery waits are **bounded**. On timeout, automation emits
sanitized diagnostics naming the missing checks / SHA / PR and fails closed
without merging.

`VOC-113-D07`: Deterministic fixtures MUST cover positive recovery (missing runs
→ dispatch/orchestration → exact-SHA checks present) and negatives (wrong SHA,
duplicate promotion, recursion, fabricated-status refusal, timeout fail-closed)
for both task-merge-to-integration and release-PR paths.

`VOC-113-D08`: Promotion PR #947 is the controlled live fixture. It may be
completed only after genuine exact-head required checks pass. Post-promotion
`main` workflows must be verified before the remediation closes.

`VOC-113-D09`: Cross-repo execution follows the VOC-104/VOC-107/VOC-108 pattern.
Primary behavior lands in `KARSIFT/karsift-ai-infra` with `Relates to` (non-closing)
references. Caller pipeline/docs/foundation tests land in this repository.
Do not treat an untracked local `karsift-ai-infra/` checkout as this repo's
tracked tree. Caller already consumes reusable workflows at `@main`.

`VOC-113-D10`: T00 adds a read-only caller `workflow_dispatch` verification
action (VOC-104 verify-* pattern) whose job display name is exactly
`verify-promotion-check-recovery / verify`. On the T01 evidence PR it validates,
using Actions/check metadata only, that a declared promotion PR (live fixture
#947) has genuine required exact-head checks successful — never fabricating
statuses and never merging. T01's live-evidence contract observes that verify
job on the T01 PR (`exact_pr_head`). Completing the promotion merge remains
release converge's job after those genuine checks exist.

`VOC-113-D11`: The VOC-112 provenance gate MUST remain strict across squash
merges and subsequent pull requests. The original capture PR must prove that
its subject commits are ancestors of its reviewed head. A later PR based on the
accepted squash may instead prove that each capture's expected immutable source
hash was already anchored in the pull request merge base and is unchanged at
the reviewed head. A non-ancestor capture that is not anchored in the merge base,
or whose current hash changed, fails closed. T00 adds deterministic original-PR,
post-squash-next-PR, tampered-base, and changed-current fixtures.

`VOC-113-D12`: T00 also adds a read-only
`verify-post-promotion-workflow` dispatch action with job display name
`verify-post-promotion-workflow / verify`. Given allowlisted promotion PR #947,
it resolves the PR's merged result SHA and verifies the expected post-promotion
workflow and job succeeded for that exact SHA. It reads Actions/PR metadata only,
does not dispatch a deploy, and binds its own successful run to the T02 carrier
through `exact_pr_head`.

## Data, migrations, analytics, and accessibility

None. Governance-automation recovery only; no database, product analytics, or UI
accessibility surface changes.

## Security, privacy, and authorization

No new long-lived secrets. App installation tokens remain short-lived and
scoped. Evidence and diagnostics must be metadata-only (SHAs, run/job IDs, check
names, conclusions, timeouts). Forbidden: logs, credentials, OAuth/session
material, tokens, user identifiers, personal data.

## Open questions

1. **Root-cause precision (`VOC-113-DEP-04`):** T00 must document whether the
   missing runs were caused by GitHub event suppression for the mutation token,
   missing App Actions permissions, first-time workflow approval friction, a
   caller wiring gap, or another verified cause. Do not invent a cause in this
   draft.
2. **Concrete timeout bound:** The requirement mandates a bounded wait; the
   exact duration (for example 30–60 minutes for check appearance) is an
   implementer choice recorded in T00 evidence, subject to fail-closed timeout
   behavior.
3. **Historical VOC-112 develop squash without push workflows:** If later
   commits already superseded that SHA on `develop`, backfilling its push
   workflows is out of scope unless T00 proves a still-blocking exact-SHA gap.
   The durable mechanism and PR #947 fixture are in scope.
4. **Caller entrypoint shape:** Whether recovery is embedded solely inside
   merge-gate/release converge, exposed as a new `pipeline.yml` dispatch action,
   or both, is left to implementation provided D03–D06 hold and docs stay true.
5. **Bootstrap provenance:** Plan PR #949 uses a content-neutral merge parent
   containing the original VOC-112 capture commits so the pre-fix validator can
   verify real ancestry. T00 may use the same content-neutral ancestry bridge;
   the permanent D11 correction must merge in T00 so later task PRs need no such
   bridge.
