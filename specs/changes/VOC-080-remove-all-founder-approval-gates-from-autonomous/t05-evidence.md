# VOC-080-EV-05 — T05 deterministic regression coverage

Evidence for `VOC-080-AC-08`, `VOC-080-AC-02`, `VOC-080-AC-03`, and
`VOC-080-AC-07`. Tests: `VOC-080-TEST-01`–`VOC-080-TEST-05` and
`VOC-080-TEST-07` (policy regressions; integrated live rehearsal remains
`VOC-080-T06`).

## Task outcome

`VOC-080-T05` adds deterministic workflow-policy regressions covering the
AC-08 surfaces. Coverage is delivered in **this caller repository** against
pinned infra contract fixtures (so the PR diff is independently reviewable
without a remote infra checkout), and mirrored as native
`karsift-ai-infra/tests/*` modules merged by infra PR #40.

## Caller delivery (this PR — authoritative for review)

### Pinned fixtures

`tooling/governance/fixtures/karsift-ai-infra/` mirrors
`KARSIFT/karsift-ai-infra` at SHA recorded in `PINNED_SHA.txt`
(`489dd82b5403a36082e70c95185463f445d02c13` = post-T03 `main`). Fixtures are
test inputs only; runtime still `uses:` `@main`.

### Policy test modules (wired via `repository-governance.yml`)

| File | Surfaces covered |
|------|------------------|
| `test_voc080_merge_gate_policy.py` | R0–R4 auto-merge eligibility; `automatic_merge_allowed` not a founder gate; unparseable-risk fail-closed; no founder override; FAIL/PENDING/checks block merge |
| `test_voc080_remediate_policy.py` | Remediation fail-closed; no founder override; bounded attempt≤2 retry via `implement.yml` |
| `test_voc080_plan_path_policy.py` | Plan PR (`plan_reviewer`) vs task PR (`reviewer`) routing; merge-gate waits on verification; fixture pin recorded |
| `test_voc080_role_separation_policy.py` | Distinct implementer/reviewer bindings; prompt forbids self-review/self-merge |
| `test_voc080_adoption_reconcile_policy.py` | Adopt header reconcile contract; issue reuse / no duplicate roster; caller reconcile dispatch; release retry without founder comment |
| `test_voc080_workflow_policy.py` | Caller `pipeline.yml` / docs / deploy-production / template drafting assertions |

Command:

```bash
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py' -v
```

Observed on this remediation attempt: **32** VOC-080 policy tests and **94**
total tests under `tooling/governance/tests` passed.

## Infra mirror (merged canonical coverage)

Matching modules under the workspace `karsift-ai-infra/tests/`:

| File | Notes |
|------|-------|
| `test_merge_gate_policy.py` | New |
| `test_remediate_policy.py` | New |
| `test_plan_path_policy.py` | New |
| `test_role_separation_policy.py` | New |
| `test_adoption_handoff.py` | Existing |
| `test_release_policy.py` | Existing |

Merged to `KARSIFT/karsift-ai-infra` by PR #40 at merge commit
`fcf2224b83c4c3060713a7e2944c7a53d5f1110d`. The suite is wired through
`karsift-ai-infra/.github/workflows/self-ci.yml` job `policy-tests`.

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py' -v
```

Observed locally and in infra PR #40 CI: **26** infra policy tests passed;
actionlint, shellcheck, YAML parse, and policy-tests were green before merge.

## Mapping to acceptance tests

| Test ID | Policy coverage in this task |
|---------|-----------------------------|
| `VOC-080-TEST-01` | `test_voc080_merge_gate_policy.py`, `test_voc080_plan_path_policy.py`, caller auto-merge wiring |
| `VOC-080-TEST-02` | `test_voc080_merge_gate_policy.py`, `test_voc080_remediate_policy.py`, release assertions in reconcile module |
| `VOC-080-TEST-03` | `test_voc080_merge_gate_policy.py::test_unparseable_risk_fails_closed` |
| `VOC-080-TEST-04` | `test_voc080_adoption_reconcile_policy.py` (idempotent reuse / no-op language + dispatch wiring; live second-run remains T06) |
| `VOC-080-TEST-05` | release assertions + `test_voc080_workflow_policy.py` deploy-production push path + remediate fail-closed |
| `VOC-080-TEST-07` | `test_voc080_role_separation_policy.py` + foundation validator suite |

Live end-to-end proof for TEST-00/04/05/06 remains `VOC-080-T06`.

## Explicitly not done

- Sandbox/live rehearsal (`VOC-080-T06`)
- Authority activation (`VOC-080-T07`)
- Syncing infra `templates/project-repo/.../pipeline.yml` plan-review job to
  match the sandbox caller (template lag; noted, not expanded here)

## Limitations

- Policy tests inspect committed workflow/doc contracts (fixtures + caller
  files); they do not substitute for live GitHub Actions rehearsal
  (`VOC-080-T06`).
- Fixture pin must be refreshed if VOC-080-related infra contracts change
  after `489dd82`.
- Infra `self-ci` will only run the new modules after they are committed to
  that repository.
