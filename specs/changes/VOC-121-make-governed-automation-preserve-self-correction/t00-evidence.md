# VOC-121-T00 — Evidence

Task: `VOC-121-T00` — Preserve self-correction helpers, finish coordinated
carriers without silent loss, and recover required checks from GitHub
satisfaction state.

Do not record secrets, credentials, session values, OAuth material, personal
data, or complete CI logs.

Implementation results are pending adoption and task authorization. The tables
below record planning-time discovery from issue #994 only.

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

## Chosen delivery path (fill during T00)

| Item | Value |
|------|-------|
| Multi-carrier path or fail-loud fallback | pending implementation |
| Safety constraint if fallback | pending implementation |
| Helper copy location | pending implementation |
| Ruleset satisfaction probe | pending implementation |

## Changed surfaces

Pending T00. Expected primary source:

- `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml`
- recovery/attestation modules and release/merge-gate callers
- infra tests and current-state README/comments

Expected caller:

- `tooling/governance/fixtures/karsift-ai-infra/` pin and mirrored files when consumed
- `tooling/governance/tests/`
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending T00 |
| Independently reviewed infra head SHA | pending T00 |
| Exact infra merge SHA | pending T00 |
| Pin applicable? | pending T00 |
| `PINNED_SHA.txt` after T00 | pending T00 |

## Validation commands

Pending T00. Record exact commands and results here, including:

```bash
python3 -m unittest discover -s tests -p 'test_*.py'
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

plus any narrower targeted commands added by the implementation.

## Acceptance mapping

- `VOC-121-AC-00` / `VOC-121-EV-00` — pending T00
- `VOC-121-AC-01` / `VOC-121-EV-00` — pending T00
- `VOC-121-AC-02` / `VOC-121-EV-00` — pending T00
- `VOC-121-AC-03` / `VOC-121-EV-00` — pending T00
- `VOC-121-AC-04` / `VOC-121-EV-00` — pending T00
- `VOC-121-AC-05` / `VOC-121-EV-00` — pending T00
- `VOC-121-AC-06` / `VOC-121-EV-00` — pending T00
- `VOC-121-AC-07` / `VOC-121-EV-00` — pending T00
