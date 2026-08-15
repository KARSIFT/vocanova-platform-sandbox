# VOC-080-EV-02 — T02 autonomous adoption and reconcile dispatch

Evidence for `VOC-080-AC-00`, `VOC-080-AC-04`, and `VOC-080-AC-05`. Tests:
`VOC-080-TEST-00`, `VOC-080-TEST-04` (policy regressions; live rehearsal is
`VOC-080-T06`).

## Task outcome

`VOC-080-T02` makes plan-package adoption autonomous in
`KARSIFT/karsift-ai-infra` `adopt.yml` and documents the caller-side reconcile
dispatch. A merged `plan/`-branch PR with an exact-revision plan-review PASS,
green CI, and (when present) passing governance validation transitions to
`status: adopted` / `implementation_authorized: true` through adopt.yml's
checked roster PR — no manual `change.yaml` flip and no founder `approved`
comment. Reconcile dispatch repairs missed merge events idempotently.

## Infra delivery

| Item | Location | Notes |
|------|----------|-------|
| Core autonomous adopt handoff | `karsift-ai-infra` PR [#37](https://github.com/KARSIFT/karsift-ai-infra/pull/37) (`da91f64`) | Initial adopt.yml autonomous path, roster PR, implement-first-task |
| T02 polish + policy tests | This task's infra working-tree delta | Reconcile docs, governance re-check, audit fields, root-task redispatch, plan.yml/template text |

Caller repos pin `@main`; adoption behavior is effective once the infra delta
lands on `karsift-ai-infra` `main`.

## Behavioral checklist (VOC-080-D01 / VOC-040)

| Requirement | Implementation |
|-------------|----------------|
| Autonomous adoption after non-human gates | `Verify merged plan and independent review` + green `gh pr checks`; optional `validate-governance.sh` |
| No silent merged-as-draft | `Commit task roster` writes adopted/authorized fields atomically with `.karsift/tasks.json` |
| Exact revision + review evidence recorded | `approved_candidate_sha`, `adoption_independent_verification_evidence`, `adoption_evidence` |
| Authority provenance recorded | `adoption_authority_provenance` cites active `authority_model` at adoption |
| Idempotent reconcile without old event | Caller `pipeline.yml` `workflow_dispatch` `action=reconcile` + `plan_pr_number`; adopt reuses task issues and no-ops unchanged roster |
| Root task dispatch after reconcile | `needs_root_dispatch` fires implement-first-task when adoption is complete but no agent PR exists yet |
| Reconcile inputs documented | `adopt.yml` header + caller template `plan_pr_number` input |

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

Expected: all policy tests pass, including expanded
`test_adoption_handoff.py` reconcile/adoption assertions.

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
