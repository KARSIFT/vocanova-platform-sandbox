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

## Remediation (attempt 2 of 2)

Independent verification on `b7e71afaab5750257b3208ab3cac608aee74f557` blocked on:

1. **AC-00** — nested source detection used porcelain only; committed implementer
   work was lost. Fixed by comparing `SOURCE_HEAD_SHA` to
   `steps.infra-checkout.outputs.base_sha` after staging any uncommitted nested
   edits, and adding `nested_worktree_has_source_changes()` for unit coverage.
2. **AC-04** — promotion attestation skipped the PR required-check gate when
   `gh pr checks --required` failed. Fixed by fail-closing with
   `required_pr_checks_read_failed` instead of falling back to check-run summary
   only.
3. **AC-06 / TEST-02** — added `test_voc121_source_carrier_publisher.py`
   exercising the `publish-source` step for missing bundles, stale live heads,
   integration-history races, and unverifiable lineage.
4. **AC-07** — README no longer claims fixture content is synchronized to the
   prior pin while VOC-121 changes are present; `PINNED_SHA.txt` update remains
   deferred to the exact reviewed infra merge SHA per `VOC-121-D10` bootstrap.

Additional fix: source bundle creation uses `"$SOURCE_BASE_SHA"..HEAD` (not
`SHA..SHA`) so `git bundle create` produces a non-empty bundle on hosted runners.

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra`):**

- `.github/workflows/implement.yml` — helper preservation, base-SHA nested source
  detection/bundle, `publish-source`
- `.github/workflows/release.yml` — required-check view gate before merge decision
- `config/required_check_satisfaction.py` — PR required-check probe helpers
- `config/implementer_source_carrier.py` — nested-edit detection and cross-repo PR body helpers
- `config/actions_check_recovery.py` — promotion recovery uses PR required-check view
- `config/actions-check-recovery-runner.py` — loads `gh pr checks --required`
- `config/promotion_status_attestation.py` — refuses attestation when PR view unsatisfied
- `config/promotion-status-attestation-runner.py` — fail-closed required-check probe
- `README.md` — current-state promotion recovery contract
- `tests/test_voc121_*.py` — deterministic VOC-121 coverage including publisher races

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
| `PINNED_SHA.txt` after T00 | **deferred** — remains `37b06aa95030e235b7311b3c14ee23977f62ac76` until the coordinated infra carrier merges; README documents `VOC-121-D10` bootstrap |

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
| `python3 -m unittest discover -s tests -p 'test_voc121_*.py'` (fixture) | **PASS** (20 tests) |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `bash scripts/governance/classify-change-risk.sh` | **PASS** |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** (183 tests) |
| `git diff --check` | **PASS** |

## Acceptance mapping

- `VOC-121-AC-00` / `VOC-121-EV-00` — base-SHA nested source detection; isolated `publish-source`
- `VOC-121-AC-01` / `VOC-121-EV-00` — no caller gitlink; publisher race/stale-head tests
- `VOC-121-AC-02` / `VOC-121-EV-00` — `Relates to OWNER/CALLER#N` on infra PR
- `VOC-121-AC-03` / `VOC-121-EV-00` — immutable helper copies after nested removal
- `VOC-121-AC-04` / `VOC-121-EV-00` — recovery and attestation fail-closed on PR required-check probe
- `VOC-121-AC-05` / `VOC-121-EV-00` — preserved App-token isolation, two-attempt bound, Cursor fail-closed
- `VOC-121-AC-06` / `VOC-121-EV-00` — `test_voc121_*` suites including publisher harness
- `VOC-121-AC-07` / `VOC-121-EV-00` — docs updated; pin deferred with explicit bootstrap note
