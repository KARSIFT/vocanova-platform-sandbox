# VOC-120 — Acceptance Criteria

## VOC-120-AC-00 — Caller tracked tree contains no Python bytecode/cache artifacts

- Requirement source: `VOC-120-D00`, `VOC-120-D03`
- Tasks: `VOC-120-T00`
- Tests: `VOC-120-TEST-00`, `VOC-120-TEST-05`
- Evidence: `VOC-120-EV-00`
- Result: pending

On the exact reviewed caller revision, `git ls-files` reports no tracked
`__pycache__/` paths and no tracked `*.pyc`, `*.pyo`, or `*.pyd` files, including
the documented path
`infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc`.
`infra/scripts/cloudflare_origin_port_remap.py` is unchanged in behavior and is
not deleted.

## VOC-120-AC-01 — Caller repository ignore rules remain portable

- Requirement source: `VOC-120-D01`, `VOC-120-D03`
- Tasks: `VOC-120-T00`
- Tests: `VOC-120-TEST-01`
- Evidence: `VOC-120-EV-00`
- Result: pending

Caller repository `.gitignore` still contains `__pycache__/` and a rule that
covers `*.pyc` and related bytecode variants. The existing `*.py[cod]` form is
sufficient. Those rules are not removed or weakened.

## VOC-120-AC-02 — Shared infrastructure has portable repository-level ignore rules

- Requirement source: `VOC-120-D02`, `VOC-120-D03`
- Tasks: `VOC-120-T00`
- Tests: `VOC-120-TEST-02`
- Evidence: `VOC-120-EV-00`
- Result: pending

`KARSIFT/karsift-ai-infra` has a repository-root `.gitignore` that covers
`__pycache__/`, `*.pyc`, and related Python bytecode variants. Fresh clones no
longer depend on private `.git/info/exclude` or `PYTHONDONTWRITEBYTECODE` for
this hygiene.

## VOC-120-AC-03 — Deterministic validation fails closed

- Requirement source: `VOC-120-D04`
- Tasks: `VOC-120-T00`
- Tests: `VOC-120-TEST-03`, `VOC-120-TEST-04`
- Evidence: `VOC-120-EV-00`
- Result: pending

Deterministic caller and infra checks fail closed when:

1. a protected source tree contains a tracked Python bytecode/cache artifact; or
2. required ignore patterns are missing from the relevant repository `.gitignore`.

Caller checks run from `bash scripts/governance/validate-governance.sh` and from
`tooling/governance/tests/`. Infra checks run from the infra unit suite.

## VOC-120-AC-04 — Existing safety gates and product behavior remain unchanged

- Requirement source: `VOC-120-D06`
- Tasks: `VOC-120-T00`
- Tests: `VOC-120-TEST-06`
- Evidence: `VOC-120-EV-00`
- Result: pending

The change does not weaken exact-SHA independent verification, deterministic risk
floors, protected checks, risk classification, or retry limits. It does not
change application behavior, deployment topology, credentials, providers, or
model routing.

## VOC-120-AC-05 — Coordinated repositories are validated and pinned only when consumed

- Requirement source: `VOC-120-D05`
- Tasks: `VOC-120-T00`
- Tests: `VOC-120-TEST-07`
- Evidence: `VOC-120-EV-00`
- Result: pending

Both repositories are validated on their exact reviewed revisions. If the caller
fixture/pin consumes the infrastructure change, infrastructure merges first and
`tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` equals that exact
infra merge SHA. If the fixture does not consume the change, the pin is left
unchanged and that non-consumption is recorded in `t00-evidence.md`.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
