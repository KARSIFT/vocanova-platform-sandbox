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
| Ruleset satisfaction probe | `gh pr checks --required --json bucket,event,link,name,state,workflow` on promotion PRs |

## Governed review and remediation

The governed Cursor Composer 2.5 implementer created caller heads
`b7e71afaab5750257b3208ab3cac608aee74f557` (attempt 1) and
`e2cdc1214d1705a0120121c3f6e8595d2408c610` (attempt 2). Independent Cursor
Grok review blocked both exact revisions rather than allowing premature merge:

1. The first review found committed nested source work could still be discarded,
   required-check probe failure could fall back during attestation, publisher
   race coverage was incomplete, and the caller could not yet be pinned.
2. The second review found the remaining #993-class defect: recovery detected
   GitHub's failed required-check row but dispatched an alternate workflow
   instead of rerunning the exact cancelled run selected by GitHub. Review:
   https://github.com/KARSIFT/vocanova-platform-sandbox/pull/998#issuecomment-5418048173

The coordinated infrastructure carrier remediated those findings at reviewed
head `c21c9a8e65c3b3ae1ed135ebcb6bfd33597aede5`:

- already-committed and uncommitted nested edits are captured relative to the
  recorded source base, bundled before checkout deletion, and published from a
  clean App-token job with exact lineage and old-head lease checks;
- immutable helper copies survive nested-checkout deletion and post-bundle
  recreated source work fails loud;
- failed/cancelled required rows are rebound to the original first-attempt
  pull-request Actions run and rerun by exact run ID; only genuinely absent
  contexts use workflow dispatch, while pending, ambiguous, foreign, status,
  and unreadable cases fail closed;
- release, attestation, and the read-only recovery verifier use GitHub's actual
  required PR view; alternate runs and same-named statuses cannot override it;
- executable tests cover nested committed edits, no-gitlink staging, valid and
  invalid source publication, exact old-head remediation, races, and the
  cancelled-selected-run plus successful-alternate case.

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra`, PR #157):**

- `.github/workflows/implement.yml` — helper preservation, base-SHA nested source
  detection/bundle, `publish-source`
- `.github/workflows/recover-actions-checks.yml` — exact selected-run recovery contract
- `.github/workflows/release.yml` — required-check view gate before merge decision
- `config/required_check_satisfaction.py` — PR required-check probe helpers
- `config/implementer_source_carrier.py` — nested-edit detection and cross-repo PR body helpers
- `config/actions_check_recovery.py` — promotion recovery uses PR required-check view
- `config/actions-check-recovery-runner.py` — validates and reruns the selected exact Actions run
- `config/promotion_status_attestation.py` — refuses attestation when PR view unsatisfied
- `config/promotion-status-attestation-runner.py` — fail-closed required-check probe
- `config/verify_promotion_check_recovery.py` and runner — read-only proof uses the same PR view
- `prompts/plan-review.md` — current A-004 R0-R4 automatic-merge drafting rule
- `README.md` — current-state publication and promotion recovery contract
- `tests/test_voc121_*.py` — deterministic VOC-121 coverage including publisher races

**Caller (`vocanova-platform-sandbox`):**

- `tooling/governance/fixtures/karsift-ai-infra/` — reviewed source-tree mirror pinned to the exact merge
- `tooling/governance/tests/test_voc121_implement_policy.py` — fixture regressions
- `tooling/governance/tests/test_voc117_role_bindings.py` — preserves the historical VOC-117 source record without freezing the global fixture pin
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | https://github.com/KARSIFT/karsift-ai-infra/pull/157 — merged |
| Independently reviewed infra head SHA | `c21c9a8e65c3b3ae1ed135ebcb6bfd33597aede5` |
| Reviewed source tree SHA | `87dff514e7897460d16a4877f0a626c8b76b96b5` |
| Exact infra merge SHA | `99476c2a1018e42d4bd442657b5257885ac9f1c9` |
| Pin applicable? | **yes** — workflow, recovery, attestation, and test changes are consumed by the fixture |
| `PINNED_SHA.txt` after source merge | `99476c2a1018e42d4bd442657b5257885ac9f1c9` |
| Fixture/source comparison | **PASS** — every source file changed by PR #157 and present in the fixture is byte-for-byte identical to reviewed tree `87dff514...` |

## Independent source review

| Item | Value |
|------|-------|
| Review record | https://github.com/KARSIFT/karsift-ai-infra/pull/157#issuecomment-5418354614 |
| Role/model | `reviewer` / `cursor/grok-4.6[effort=high,fast=false]` |
| Session / request | `74d63e75-72c2-4cfe-a606-c55be91fc3ea` / `e8edcf7f-b71a-44b9-a9db-22681c81dd42` |
| Verdict | **PASS WITH NON-BLOCKING FINDINGS** |
| Non-blocking finding | P3 unused fallback-prose helper; selected isolated publisher already fails closed |

The review was bound to base `9c01779ce166789d7d32eb47c502d42234d82a4c`,
head `c21c9a8e65c3b3ae1ed135ebcb6bfd33597aede5`, and tree
`87dff514e7897460d16a4877f0a626c8b76b96b5`. Source self-CI separately
confirmed actionlint, shellcheck, YAML parsing, and policy tests on that head.

## Validation commands

```bash
# Infrastructure (exact reviewed source worktree)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
(cd tooling/governance/fixtures/karsift-ai-infra && \
  python3 -m unittest discover -s tests -p 'test_*.py')
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (reviewed infrastructure head) | **PASS** (328 tests) |
| `python3 -m unittest discover -s tests -p 'test_*.py'` (pinned fixture) | **PASS** (247 tests) |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `bash scripts/governance/classify-change-risk.sh` | **PASS** — R4 floor |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** (184 tests) |
| Parse every fixture YAML file with PyYAML | **PASS** |
| `git diff --check` | **PASS** |

## Acceptance mapping

- `VOC-121-AC-00` / `VOC-121-EV-00` — base-SHA nested source detection; isolated `publish-source`
- `VOC-121-AC-01` / `VOC-121-EV-00` — no caller gitlink; publisher race/stale-head tests
- `VOC-121-AC-02` / `VOC-121-EV-00` — `Relates to OWNER/CALLER#N` on infra PR
- `VOC-121-AC-03` / `VOC-121-EV-00` — immutable helper copies after nested removal
- `VOC-121-AC-04` / `VOC-121-EV-00` — exact selected-run rerun; absent-only dispatch; alternate-success/status suppression blocked
- `VOC-121-AC-05` / `VOC-121-EV-00` — preserved App-token isolation, two-attempt bound, Cursor fail-closed
- `VOC-121-AC-06` / `VOC-121-EV-00` — `test_voc121_*` suites including publisher harness
- `VOC-121-AC-07` / `VOC-121-EV-00` — current-state docs reconciled and caller fixture pinned to exact infra merge `99476c2...`
