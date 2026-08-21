---
evidence_id: VOC-097-EV-04
task_id: VOC-097-T04
acceptance_criteria:
  - VOC-097-AC-08
tests:
  - VOC-097-TEST-15
date: 2026-08-21
related_change: VOC-097
gate_status: migration-complete-live-closure-pending
live_reconcile_claimed: false
---

# VOC-097-T04 — Reconcile stranded tasks #779 and #785

## Scope of this evidence

This task migrates stranded live-evidence tasks onto the governed waiting/reconcile
path introduced in VOC-097-T01/T02. It does **not** grant implementer Actions
credentials, does not merge the predecessor task PRs, and does not claim live
qualification (`live_reconcile_claimed: false`; closure proof continues in T05
after operator reconcile and fresh exact-SHA review on each task PR).

## Stranded baseline (issue #823)

| Task | Issue | PR | Pollution / blocker |
|------|-------|-----|---------------------|
| `VOC-093-T01` | [#779](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/779) | [#789](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/789) | Branch added a one-off `voc-093-t01-live-verify` pipeline job (~228 lines) to manufacture evidence; live proof still pending |
| `VOC-094-T01` | [#785](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/785) | [#791](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/791) | Evidence-only branch; AC-05(a) green head deploy still pending after run #306 timeout |

Both tasks already had T00 code fixes merged to `develop` (PR #783 and PR #788).

## Safe migration path (VOC-097-D08)

Adoption open question 5 is answered **yes**: a clean evidence-only replacement
on the same task issue/PR is acceptable when history contains out-of-scope
workflow edits. T04 uses **in-place branch reset** on the existing PR numbers so
roster/issue wiring stays intact.

### Shared steps (after this revision merges to `develop`)

1. Ensure `develop` contains this T04 revision (contracts + waiting-path
   `t01-evidence.md` templates + VOC-097 reconcile wiring from T02).
2. Reset each task branch to the current `develop` tip so the PR head SHA equals
   integration (`git push --force-with-lease origin develop:agent/voc-093-voc-093-t01`
   and the analogous command for `agent/voc-094-voc-094-t01`). The PR may show
   an empty diff while remaining open for exact-SHA review.
3. Re-run the task PR pipeline so independent review can post
   `VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE` against the declared contract.
4. Use repository-controlled reconcile only (`pipeline.yml`
   `action=reconcile-live-evidence` or hourly poll). Do not reintroduce
   task-specific pipeline dispatch jobs.

### VOC-093-T01 (#779 / #789)

| Item | Value |
|------|-------|
| Contract | `specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure/.karsift/live-evidence/VOC-093-T01.yaml` |
| Removed scope | Entire `voc-093-t01-live-verify` job from `.github/workflows/pipeline.yml` (never merge) |
| Qualification | Declared `dispatch` on `scheduled-synthetics.yml` with `synthetic_id=synthetic.production.authenticated-route-content-sweep`; `exact_pr_head` on ref `agent/voc-093-voc-093-t01` |
| Operator dispatch example | `workflow_dispatch` `pipeline.yml` with `action=reconcile-live-evidence`, `live_evidence_mode=dispatch`, `live_evidence_pr_number=789` |

Prefer reconcile dispatch over manual workflow edits. Observe-only is supported
once a matching successful `workflow_dispatch` run exists on the task branch.

### VOC-094-T01 (#785 / #791)

| Item | Value |
|------|-------|
| Contract | `specs/changes/VOC-094-operational-failure-deploy-staging-cancelled/.karsift/live-evidence/VOC-094-T01.yaml` |
| Qualification | Observe-only successful `deploy-staging` `push` on `develop` with `integration_contains_pr_head` (requires PR head SHA equal to `develop` tip at observation time) |
| Prior partial proof | Classifier skip + queue posture from the stranded `t01-evidence.md` revision is preserved in the waiting-path template |
| Operator observe example | After the next green head deploy, `workflow_dispatch` `pipeline.yml` with `action=reconcile-live-evidence`, `live_evidence_mode=observe`, `live_evidence_run_id=<scrubbed id>` |

Run #306 (`32293128673`) remains a non-qualifying timeout cancel and must not be
treated as supersession evidence.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| VOC-093 contract | `specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure/.karsift/live-evidence/VOC-093-T01.yaml` |
| VOC-094 contract | `specs/changes/VOC-094-operational-failure-deploy-staging-cancelled/.karsift/live-evidence/VOC-094-T01.yaml` |
| VOC-093 waiting evidence | `specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure/t01-evidence.md` |
| VOC-094 waiting evidence | `specs/changes/VOC-094-operational-failure-deploy-staging-cancelled/t01-evidence.md` |
| Package task cross-links | `specs/changes/VOC-093-*/tasks.md`, `specs/changes/VOC-094-*/tasks.md` |
| TEST-15 lock | `scripts/foundation/voc097-stranded-migration.test.mjs` |
| This evidence | `specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t04-evidence.md` |

## Acceptance mapping

| AC | Result |
| --- | --- |
| AC-08 | **migration complete, live closure pending** — both tasks have valid contracts, waiting-path evidence, documented safe reset for #789 pipeline pollution, and operator reconcile instructions; issue/PR closure awaits qualifying runs + fresh exact-SHA review + merge (or further migration notes if reset fails) |

## Deterministic validation

```bash
node --test scripts/foundation/voc097-stranded-migration.test.mjs
python3 karsift-ai-infra/config/live_evidence_reconcile.py  # via unittest imports in TEST-15
bash scripts/governance/validate-governance.sh
git diff --check
```

No secrets, tokens, logs, OAuth values, or personal identifiers are recorded
here.

## Deferred to T05

- Live proof that waiting skips remediation on the migrated PRs
- Post-reconcile fresh exact-SHA review and merge of #789 / #791
- Issue #779 / #785 closure after roster path completes
