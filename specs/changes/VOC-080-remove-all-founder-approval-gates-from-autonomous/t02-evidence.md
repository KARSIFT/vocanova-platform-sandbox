# VOC-080-EV-02 — T02 autonomous adoption and reconcile dispatch

Evidence for `VOC-080-AC-00`, `VOC-080-AC-04`, and `VOC-080-AC-05`. Tests:
`VOC-080-TEST-00`, `VOC-080-TEST-04` (policy regressions; live rehearsal is
`VOC-080-T06`).

## Task outcome

`VOC-080-T02` completes autonomous plan-package adoption in
`KARSIFT/karsift-ai-infra` `adopt.yml` and documents the caller-side reconcile
dispatch. A merged `plan/`-branch PR with an exact-revision plan-review PASS,
green plan-PR checks, and (when present) passing
`scripts/governance/validate-governance.sh` transitions to `status: adopted` /
`implementation_authorized: true` through adopt.yml's checked roster PR — no
manual `change.yaml` flip and no founder `approved` comment. Reconcile
dispatch repairs missed merge events idempotently and re-dispatches the root
task when adoption exists but no agent PR has been opened yet.

## Infra delivery

| Item | Location | Notes |
|------|----------|-------|
| Core autonomous adopt handoff | `karsift-ai-infra` PR [#37](https://github.com/KARSIFT/karsift-ai-infra/pull/37) (`da91f64`) | Initial adopt.yml autonomous path, roster PR, implement-first-task |
| T02 completion | `karsift-ai-infra` PR [#38](https://github.com/KARSIFT/karsift-ai-infra/pull/38) (`a9f0c6512bb87b45d4579e2f0bf5de607abc0df0`) | AC-00 audit fields, plan-PR checks + optional governance validation, missed-root-dispatch recovery, adopt/plan/template contract documentation, expanded `tests/test_adoption_handoff.py` |

Caller repos pin `@main`; PR #38 is merged there, so callers receive the T02
behavior rather than relying on an uncommitted local checkout.

### Files changed in the T02 remediation delta

- `.github/workflows/adopt.yml`
- `.github/workflows/plan.yml`
- `templates/project-repo/.github/workflows/pipeline.yml` (header comments)
- `tests/test_adoption_handoff.py`

## Behavioral checklist (VOC-080-D01 / VOC-040)

| Requirement | Implementation |
|-------------|----------------|
| Autonomous adoption after non-human gates | `Verify merged plan and independent review`: PASS bound to `headRefOid`, green `gh pr checks`, optional `validate-governance.sh` |
| No silent merged-as-draft | `Commit task roster` writes adopted/authorized fields atomically with `.karsift/tasks.json` |
| Exact revision + review evidence recorded | `approved_candidate_sha`, `adoption_independent_verification_evidence`, `adoption_evidence` |
| Risk + decisions recorded | `adoption_risk`, `adoption_resolved_decisions`, `adoption_deferred_decisions` |
| Authority provenance recorded | `adoption_authority_provenance` cites active `authority_model` at adoption |
| Idempotent reconcile without old event | Caller `pipeline.yml` `workflow_dispatch` `action=reconcile` + `plan_pr_number`; adopt reuses task issues and no-ops unchanged roster |
| Root task dispatch after reconcile | `Determine whether the root task needs dispatch` sets `root-dispatch.outputs.needed` / `should_dispatch` when adoption is complete but no agent PR exists |
| Reconcile inputs documented | `adopt.yml` header documents `action=reconcile` / `plan_pr_number`; caller template retains the inputs |
| Plan-merge contract | `plan.yml` PR body and header describe automatic adopt — not a manual `change.yaml` flip |

## Reconcile dispatch (caller)

```bash
gh workflow run pipeline.yml --ref develop \
  -f action=reconcile \
  -f plan_pr_number=<merged plan PR number>
```

Same gates as the happy path: PR must be merged, exact head must have PASS
plan-review bound to that SHA, and plan PR checks must be green.

## Deterministic verification

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py' -v
```

Observed on infra PR #38: policy tests, actionlint, shellcheck, and YAML parsing
all passed. Policy coverage includes expanded
`test_adoption_handoff.py` reconcile/adoption assertions for AC-00 fields,
plan-PR checks, header dispatch docs, missed-root dispatch, and plan.yml
manual-flip absence.

## Explicitly not done (other tasks)

- Release/deploy founder-gate removal (`VOC-080-T03`)
- Full caller `pipeline.yml` / AGENTS.md / DOC-15 reconciliation (`VOC-080-T04`)
- Live sandbox rehearsal (`VOC-080-T06`)
- Authority activation (`VOC-080-T07`)

## Limitations

- Live autonomous-adopt rehearsal requires a merged plan PR with plan-review PASS
  on the rehearsal target (`VOC-080-T06`).
- Reconcile still refuses packages whose merged plan head lacks a bound PASS
  verdict or green checks — by design (`VOC-080-D02`).
