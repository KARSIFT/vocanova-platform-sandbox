# VOC-120-T00 — Evidence

Task: `VOC-120-T00` — Untrack caller bytecode, add portable infra ignores, and add
hygiene validation.

Do not record secrets, credentials, session values, OAuth material, personal data,
or complete CI logs.

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

## Remediation context (attempt 2)

Independent verification of attempt 1 (`49fbb1fa70ab9b46b2afdbbcc234db7ea9866de4`)
failed on missing coordinated infra delivery, evidence over-claims, and a caller
validator `--deleted` filter that hid indexed bytecode during unstaged deletes.
Attempt 2 remediates those findings on top of the attempt-1 caller commit.

## Changed surfaces

### Caller (`vocanova-platform-sandbox`)

- Untracked path (attempt 1, committed): `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc`
- Preserved: `infra/scripts/cloudflare_origin_port_remap.py` (no source diff).
- Verified caller `.gitignore` retains `__pycache__/` and `*.py[cod]`.
- `tooling/governance/validate_python_bytecode_hygiene.py` scans `git ls-files`
  only (no `--deleted` exclusion).
- `tooling/governance/tests/test_python_bytecode_hygiene.py` includes a working-tree
  delete negative case proving indexed bytecode still fails closed.
- Hooked hygiene validation from `scripts/governance/validate-governance.sh`.

### Shared infrastructure (`KARSIFT/karsift-ai-infra`)

- Added repository-root `.gitignore` with `__pycache__/` and `*.py[cod]`.
- Added `tests/test_python_cache_ignore.py` with positive and negative coverage.
- Infra implementation base: `37b06aa95030e235b7311b3c14ee23977f62ac76`.
- Infra changes are in the local `karsift-ai-infra/` checkout working tree for a
  separate reviewed infra PR (`Relates to KARSIFT/vocanova-platform-sandbox#989`).

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra source | local `karsift-ai-infra/` checkout (separate repository carrier) |
| Infra implementation base SHA | `37b06aa95030e235b7311b3c14ee23977f62ac76` |
| Reviewed infra merge SHA | pending independent infra PR review/merge |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Pin applicable? | **no** (`VOC-120-D05`) |
| `PINNED_SHA.txt` after T00 | unchanged: `37b06aa95030e235b7311b3c14ee23977f62ac76` |

**Pin non-consumption rationale:** the caller fixture is a policy-contract subset
that does not mirror repository-root `.gitignore` or the new infra ignore tests.
Assertions for infra ignore rules run in the primary `karsift-ai-infra` unit suite
only; copying `.gitignore` into the fixture merely to force a pin is explicitly out
of scope.

## Validation commands

Caller reviewed base for attempt 1: `49fbb1fa70ab9b46b2afdbbcc234db7ea9866de4`.
Attempt-2 remediation applies additional caller working-tree edits on that base.
Infra validation runs against the local `karsift-ai-infra/` checkout at base
`37b06aa95030e235b7311b3c14ee23977f62ac76` plus the new `.gitignore` and test file.

| Command | Result | Notes |
|---------|--------|-------|
| `git ls-files '**/__pycache__/**' '*.pyc' '*.pyo' '*.pyd'` | pass | No tracked bytecode/cache artifacts on caller HEAD `49fbb1fa…`. |
| `bash scripts/governance/validate-governance.sh` | pass | Invokes `validate_python_bytecode_hygiene.py`. |
| `bash scripts/governance/classify-change-risk.sh` | pass | Path floor includes `scripts/governance/` and `tooling/governance/`. |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass (179 tests) | Includes hygiene tests and working-tree delete negative case. |
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `karsift-ai-infra` | pass (288 tests) | Includes `test_python_cache_ignore` (4 tests). |
| `git diff --check` | pass | No conflict markers or whitespace errors. |

Targeted hygiene commands:

```bash
python3 tooling/governance/validate_python_bytecode_hygiene.py --repository-root .
python3 -m unittest tooling/governance/tests/test_python_bytecode_hygiene.py
python3 -m unittest tests.test_python_cache_ignore  # from karsift-ai-infra repo root
```

All three passed during attempt-2 remediation.

## Acceptance mapping

- `VOC-120-AC-00` / `VOC-120-EV-00` — known `.pyc` untracked on caller HEAD;
  `infra/scripts/cloudflare_origin_port_remap.py` unchanged.
- `VOC-120-AC-01` / `VOC-120-EV-00` — caller `.gitignore` verified by hygiene
  validator and unit tests.
- `VOC-120-AC-02` / `VOC-120-EV-00` — infra `.gitignore` added with required
  patterns; infra unit tests pass in the coordinated checkout.
- `VOC-120-AC-03` / `VOC-120-EV-00` — positive/negative hygiene tests in both
  repositories; governance hook fails closed on tracked bytecode and missing
  ignore patterns; caller validator inspects indexed paths only.
- `VOC-120-AC-04` / `VOC-120-EV-00` — no product/deploy/credential/routing
  changes; existing governance gates preserved.
- `VOC-120-AC-05` / `VOC-120-EV-00` — fixture pin intentionally unchanged;
  non-consumption recorded above; exact reviewed infra merge SHA pending infra PR.
