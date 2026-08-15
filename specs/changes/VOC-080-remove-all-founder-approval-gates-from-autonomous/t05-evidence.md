# VOC-080-EV-05 — T05 deterministic regression coverage

Evidence for `VOC-080-AC-08`, `VOC-080-AC-02`, `VOC-080-AC-03`, and
`VOC-080-AC-07`. Tests: `VOC-080-TEST-01`–`VOC-080-TEST-05` and
`VOC-080-TEST-07` (policy regressions; integrated live rehearsal remains
`VOC-080-T06`).

## Task outcome

`VOC-080-T05` adds deterministic workflow-policy regressions in
`KARSIFT/karsift-ai-infra` and this caller repository. The tests assert the
post-T01/T03/T04 no-founder-gate behavior at the YAML/prompt/doc level without
requiring production credentials or live GitHub rehearsal.

## Infra policy tests (`karsift-ai-infra/tests/`)

| File | Surfaces covered |
|------|------------------|
| `test_merge_gate_policy.py` | R0–R4 auto-merge eligibility, unparseable-risk fail-closed, no founder override path, verdict/check gates |
| `test_adoption_handoff.py` (existing) | Autonomous adopt audit fields, reconcile dispatch contract, App-token merge |
| `test_release_policy.py` (existing) | Release without founder comment, dispatch-driven retry |
| `test_remediate_policy.py` | Remediation fail-closed/no founder override, bounded retry via `implement.yml` |
| `test_plan_path_policy.py` | Plan PR vs task PR review routing, merge-gate waits for plan-review |
| `test_role_separation_policy.py` | Distinct implementer/reviewer roles; prompt separation language |

Wired through `karsift-ai-infra/.github/workflows/self-ci.yml` job
`policy-tests` (`python3 -m unittest discover -s tests -p 'test_*.py'`).

## Caller policy tests (`tooling/governance/tests/`)

| File | Surfaces covered |
|------|------------------|
| `test_voc080_workflow_policy.py` | `pipeline.yml` reconcile wiring, `auto_merge_enabled`, no `founder_username`; template `automatic_merge_allowed: true`; deploy push path; live-doc absence of founder-comment engineering gates |

Wired through `.github/workflows/repository-governance.yml` step
`Run repository-foundation unit tests`.

## Deterministic verification

Caller repository:

```bash
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py' -v
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Infra repository checkout:

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py' -v
```

Observed on this task attempt: **26** infra policy tests and **9** caller
VOC-080 policy tests passed locally (71 total in `tooling/governance/tests`).

## Mapping to acceptance tests

| Test ID | Policy coverage in this task |
|---------|-----------------------------|
| `VOC-080-TEST-01` | `test_merge_gate_policy.py`, `test_plan_path_policy.py`, caller pipeline auto-merge wiring |
| `VOC-080-TEST-02` | `test_merge_gate_policy.py`, `test_remediate_policy.py`, `test_release_policy.py` |
| `VOC-080-TEST-03` | `test_merge_gate_policy.py::test_unparseable_risk_fails_closed` |
| `VOC-080-TEST-04` | `test_adoption_handoff.py`, caller reconcile dispatch assertions |
| `VOC-080-TEST-05` | `test_release_policy.py`, caller deploy-production push path |
| `VOC-080-TEST-07` | `test_role_separation_policy.py`, foundation validator suite (existing) |

Live end-to-end proof for TEST-00/04/05/06 remains `VOC-080-T06`.

## Explicitly not done

- Sandbox/live rehearsal (`VOC-080-T06`)
- Authority activation (`VOC-080-T07`)
- Infra prompt/header stale founder-language cleanup outside test scope (follow-up if desired)

## Limitations

- Policy tests inspect committed workflow/doc contracts; they do not substitute
  for live GitHub Actions rehearsal on the sandbox (`VOC-080-T06`).
- Infra tests run against the local `karsift-ai-infra` checkout present in this
  workspace; callers pinned to `@main` receive the behavior once the infra delta
  lands on `karsift-ai-infra` `main`.
