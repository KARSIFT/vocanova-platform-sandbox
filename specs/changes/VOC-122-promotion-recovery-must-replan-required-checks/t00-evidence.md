# VOC-122-T00 — Evidence

Task: `VOC-122-T00` — Replan newly appearing required-check rows during
promotion-recovery polling.

Do not record secrets, credentials, session values, OAuth material, personal
data, or complete CI logs.

## Discovery recorded at planning time (issue #1001)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1001 |
| Promotion PR | #1000 |
| `develop` head that opened it | `bb531fc211e91e40cec1847e1e28d98beaedacff` |
| First recovery run / job | `32912818046` / `98010275057` at `2026-08-25T23:54:50Z` |
| Snapshot-1 action | absent-context `governance-policy` dispatch `32912851134` (passed) |
| Later-appearing selected row | pull-request `governance-policy` run `32912850066` at `2026-08-25T23:54:52Z`, `CANCELLED` |
| Required view | `gh pr checks 1000 --required` continued to select that row as `CANCELLED` |
| Defect | `plan_required_check_recovery` / exact rerun / dispatch planning only before the `while` loop in `config/actions-check-recovery-runner.py` |
| Later recovery that worked | convergence run `32912851044` / job `98010820950` reran `32912850066` as attempt 2 |
| Merge | App merged #1000 at `70db9d7dbc4789264732e4fe4f347914dd6f764e` with no ruleset bypass |

## Chosen delivery path

| Item | Value |
|------|-------|
| Replan mechanism | Record during implementation: helper versus inline in `main()`, and whether `plan_required_check_recovery` itself changed |
| Run-ID dedupe | Record the invocation-scoped structure used |
| Absent-context dedupe | Record the invocation-scoped structure used |
| Ruleset satisfaction probe | Must remain GitHub's required PR view (`gh pr checks --required` or the equivalent already used by VOC-121) |

## Changed surfaces

Fill in during implementation. Expected:

**Infrastructure (`KARSIFT/karsift-ai-infra`):**

- `config/actions-check-recovery-runner.py`
- `config/required_check_satisfaction.py` only if needed
- recovery/release comments and `README.md`
- deterministic time-evolving tests

**Caller (`vocanova-platform-sandbox`):**

- `tooling/governance/fixtures/karsift-ai-infra/` pin and consumed files
- caller fixture regressions and any `scripts/foundation/*` pin literals
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending implementation |
| Independently reviewed infra head SHA | pending implementation |
| Exact infra merge SHA | pending implementation |
| Pin applicable? | pending implementation — expected yes if the fixture consumes the runner/tests/docs |
| `PINNED_SHA.txt` after source merge | pending implementation |

## Validation commands

```bash
# Infrastructure (exact reviewed source worktree)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (reviewed infrastructure head) | pending implementation |
| `bash scripts/governance/validate-governance.sh` | pending implementation |
| `bash scripts/governance/classify-change-risk.sh` | pending implementation |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pending implementation |
| `git diff --check` | pending implementation |

## Acceptance mapping

- `VOC-122-AC-00` / `VOC-122-EV-00` — pending implementation
- `VOC-122-AC-01` / `VOC-122-EV-00` — pending implementation
- `VOC-122-AC-02` / `VOC-122-EV-00` — pending implementation
- `VOC-122-AC-03` / `VOC-122-EV-00` — pending implementation
- `VOC-122-AC-04` / `VOC-122-EV-00` — pending implementation
- `VOC-122-AC-05` / `VOC-122-EV-00` — pending implementation
- `VOC-122-AC-06` / `VOC-122-EV-00` — pending implementation
- `VOC-122-AC-07` / `VOC-122-EV-00` — pending implementation
