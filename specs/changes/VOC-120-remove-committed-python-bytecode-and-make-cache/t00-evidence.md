# VOC-120-T00 — Evidence

Task: `VOC-120-T00` — Untrack caller bytecode, add portable infra ignores, and
add hygiene validation.

Do not record secrets, credentials, session values, OAuth material, personal data,
or complete CI logs.

This file is a draft evidence shell. Implementation fills commands, SHAs, and
results after the package is adopted and the task is authorized. Discovery facts
below are from issue #987 and are not themselves implementation proof.

## Discovery recorded at planning time (issue #987)

| Item | Value |
|------|-------|
| Caller `origin/develop` | `d67e34b149ad45fd8dcd0e3860a0bbfd1d3da0fb` |
| Caller `origin/main` | `48d7bb5a410035d4d1f12e5c34860399509d9215` |
| Tracked path | `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc` |
| Tracked blob | 5,825-byte CPython 3.12 timestamp-based bytecode |
| Introducing commit | `2118865b6acc4bbd533f62b0cf12330f6cf123d0` |
| Caller ignore rules already present | `__pycache__/` and `*.py[cod]` |
| Shared infra `origin/main` | `37b06aa95030e235b7311b3c14ee23977f62ac76` |
| Infra ignore gap | no repository `.gitignore` |
| Caller fixture pin at planning time | `37b06aa95030e235b7311b3c14ee23977f62ac76` |

## Changed surfaces

To be recorded during implementation:

- Caller untracked paths
- Caller `.gitignore` verification
- Caller validator and tests
- `KARSIFT/karsift-ai-infra` `.gitignore` and tests
- Fixture pin update, or explicit non-consumption

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra source | pending |
| Reviewed infra merge SHA | pending |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Pin applicable? | pending (`VOC-120-D05`) |
| `PINNED_SHA.txt` after T00 | pending |

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `git ls-files` bytecode/cache scan | pending | Must report no tracked `__pycache__/`, `*.pyc`, `*.pyo`, or `*.pyd` |
| `bash scripts/governance/validate-governance.sh` | pending | Must invoke the new caller hygiene check |
| `bash scripts/governance/classify-change-risk.sh` | pending | Record detected path floor |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pending | Include new hygiene tests |
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `KARSIFT/karsift-ai-infra` | pending | Include new ignore-rule tests |
| `git diff --check` | pending | |

## Acceptance mapping

- `VOC-120-AC-00` / `VOC-120-EV-00` — pending untrack proof and unchanged Python source.
- `VOC-120-AC-01` / `VOC-120-EV-00` — pending caller ignore-rule verification.
- `VOC-120-AC-02` / `VOC-120-EV-00` — pending infra `.gitignore` proof.
- `VOC-120-AC-03` / `VOC-120-EV-00` — pending positive/negative hygiene tests.
- `VOC-120-AC-04` / `VOC-120-EV-00` — pending confirmation that safety gates and product behavior are unchanged.
- `VOC-120-AC-05` / `VOC-120-EV-00` — pending exact infra SHA and pin applicability.
