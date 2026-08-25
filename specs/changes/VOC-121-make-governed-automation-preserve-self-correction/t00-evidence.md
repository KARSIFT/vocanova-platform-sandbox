# VOC-121-T00 — Evidence

Task: `VOC-121-T00` — Preserve self-correction helpers, finish coordinated
carriers without silent loss, and recover required checks from GitHub
satisfaction state.

Do not record secrets, credentials, session values, OAuth material, personal
data, or complete CI logs.

## Discovery recorded at planning time (issue #994)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #994 |
| Adopted task that exposed the failures | `VOC-120-T00` |
| Implementation run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32899479806 |
| Implementation job | `97970681892` |
| Discard mechanism | copy `config/run-app-checks.sh`, `rm -rf karsift-ai-infra`, caller-only Git bundle |
| Self-correction failure | `python3: can't open file .../karsift-ai-infra/config/prepare_cursor_model.py` |
| Promotion PR | #993 |
| Cancelled required check-run | exact-head `governance-policy` pull-request run, cancelled by concurrency |
| Alternate successful evidence | workflow-dispatch run and published status of the same context |
| Release run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32902418610 |
| Recovery log class | `actions-check-recovery: ... dispatched=none` then merge failures |
| Working recovery | rerun the cancelled exact-head governance run; App merge without admin bypass |

## Chosen delivery path

| Item | Value |
|------|-------|
| Multi-carrier path or fail-loud fallback | **Full isolated multi-carrier publication** (`publish-source` job) |
| Safety constraint if fallback | Not used — nested edits are committed, bundled, and published from a clean runner with an App token scoped only to `KARSIFT/karsift-ai-infra` |
| Helper copy location | `/tmp/run-app-checks.sh` and `/tmp/karsift-implement-helpers/prepare_cursor_model.py` |
| Ruleset satisfaction probe | `gh pr checks --required --json name,state` on promotion PRs |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra`):**

- `.github/workflows/implement.yml` — helper preservation, nested source bundle, `publish-source`
- `.github/workflows/release.yml` — required-check view gate before merge decision
- `config/required_check_satisfaction.py` — PR required-check probe helpers
- `config/implementer_source_carrier.py` — nested-edit detection and cross-repo PR body helpers
- `config/actions_check_recovery.py` — promotion recovery uses PR required-check view
- `config/actions-check-recovery-runner.py` — loads `gh pr checks --required`
- `config/promotion_status_attestation.py` — refuses attestation when PR view unsatisfied
- `config/promotion-status-attestation-runner.py` — loads PR required checks before attestation
- `README.md` — current-state promotion recovery contract
- `tests/test_voc121_*.py` — deterministic VOC-121 coverage

**Caller (`vocanova-platform-sandbox`):**

- `tooling/governance/fixtures/karsift-ai-infra/` — mirrored infra changes (pin pending infra merge)
- `tooling/governance/tests/test_voc121_implement_policy.py` — fixture regressions
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending independent review and merge |
| Independently reviewed infra head SHA | pending |
| Exact infra merge SHA | pending |
| Pin applicable? | **yes** — workflow, recovery, attestation, and test changes are consumed by the fixture |
| `PINNED_SHA.txt` after T00 | pending exact reviewed infra merge SHA (currently `37b06aa95030e235b7311b3c14ee23977f62ac76`) |

## Validation commands

```bash
# Infrastructure (nested checkout)
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py'
python3 -m unittest discover -s tests -p 'test_voc121_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (infra) | **PASS** (305 tests) |
| `python3 -m unittest discover -s tests -p 'test_voc121_*.py'` (infra) | **PASS** (15 tests) |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** (182 tests) |
| `git diff --check` | **PASS** |

## Acceptance mapping

- `VOC-121-AC-00` / `VOC-121-EV-00` — implemented; isolated `publish-source` path
- `VOC-121-AC-01` / `VOC-121-EV-00` — implemented; no caller gitlink; clean publisher isolation
- `VOC-121-AC-02` / `VOC-121-EV-00` — implemented; `Relates to OWNER/CALLER#N` on infra PR
- `VOC-121-AC-03` / `VOC-121-EV-00` — implemented; immutable helper copies after nested removal
- `VOC-121-AC-04` / `VOC-121-EV-00` — implemented; PR required-check probe drives recovery/attestation
- `VOC-121-AC-05` / `VOC-121-EV-00` — preserved App-token isolation, two-attempt bound, Cursor fail-closed
- `VOC-121-AC-06` / `VOC-121-EV-00` — `test_voc121_*` suites added
- `VOC-121-AC-07` / `VOC-121-EV-00` — docs updated; fixture pin pending infra merge SHA
