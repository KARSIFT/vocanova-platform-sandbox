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
Attempt 2 remediates those findings on top of the attempt-1 caller commit and
updates this evidence file with the merged coordinated infra delivery record.

Caller exact final head SHA is bound by the independent exact-revision review
record on caller PR #991 (`https://github.com/KARSIFT/vocanova-platform-sandbox/pull/991`),
not self-recorded in this evidence file.

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
- Coordinated infra PR: `https://github.com/KARSIFT/karsift-ai-infra/pull/156`
  (`Relates to KARSIFT/vocanova-platform-sandbox#989`).
- Infra implementation base: `37b06aa95030e235b7311b3c14ee23977f62ac76`.
- Independently reviewed infra head: `e3d53350f9903a6c364446db698c42e1a1b85e62`.
- Independent review comment:
  `https://github.com/KARSIFT/karsift-ai-infra/pull/156#issuecomment-5417069458`
  — `VERDICT: PASS`, no findings.
- Exact infra merge SHA: `9c01779ce166789d7d32eb47c502d42234d82a4c`.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | `https://github.com/KARSIFT/karsift-ai-infra/pull/156` |
| Infra implementation base SHA | `37b06aa95030e235b7311b3c14ee23977f62ac76` |
| Independently reviewed infra head SHA | `e3d53350f9903a6c364446db698c42e1a1b85e62` |
| Independent review comment | `https://github.com/KARSIFT/karsift-ai-infra/pull/156#issuecomment-5417069458` |
| Exact infra merge SHA | `9c01779ce166789d7d32eb47c502d42234d82a4c` |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Pin applicable? | **no** (`VOC-120-D05`) |
| `PINNED_SHA.txt` after T00 | unchanged: `37b06aa95030e235b7311b3c14ee23977f62ac76` |

**Pin non-consumption rationale:** `tooling/governance/fixtures/karsift-ai-infra/` does
not consume the source repository-root `.gitignore` or `tests/test_python_cache_ignore.py`
from `KARSIFT/karsift-ai-infra`. Assertions for infra ignore rules run in the primary
infra unit suite only; copying those files into the fixture merely to force a pin is
explicitly out of scope (`VOC-120-D05`).

## Validation commands

Attempt-1 reviewed caller head: `49fbb1fa70ab9b46b2afdbbcc234db7ea9866de4`.
Attempt-2 remediation applies additional caller edits on top of that commit. Infra
validation runs against exact merge SHA
`9c01779ce166789d7d32eb47c502d42234d82a4c`.

### Caller (attempt-2 remediation carrier)

| Command | Result | Notes |
|---------|--------|-------|
| `git ls-files '**/__pycache__/**' '*.pyc' '*.pyo' '*.pyd'` | pass | No tracked bytecode/cache artifacts on remediation carrier. |
| `bash scripts/governance/validate-governance.sh` | pass | Invokes `validate_python_bytecode_hygiene.py`. |
| `bash scripts/governance/classify-change-risk.sh` | pass | Path floor includes `scripts/governance/` and `tooling/governance/`. |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass (179 tests) | Includes hygiene tests and working-tree delete negative case. |
| `git diff --check` | pass | No conflict markers or whitespace errors. |

Targeted caller hygiene commands:

```bash
python3 tooling/governance/validate_python_bytecode_hygiene.py --repository-root .
python3 -m unittest tooling/governance/tests/test_python_bytecode_hygiene.py
```

Both passed during attempt-2 remediation (9 tests in targeted hygiene module).

### Shared infra (exact merge SHA `9c01779c…`)

| Command | Result | Notes |
|---------|--------|-------|
| `python3 -m unittest tests.test_python_cache_ignore` | pass (6 tests) | Targeted ignore-rule coverage at merge SHA. |
| `python3 -m unittest discover -s tests -p 'test_*.py'` | pass (290 tests) | Full infra unittest discovery at merge SHA. |
| `git diff --check` | pass | No conflict markers or whitespace errors. |

Targeted infra command:

```bash
python3 -m unittest tests.test_python_cache_ignore  # from karsift-ai-infra repo root at 9c01779c
```

Passed 6/6 targeted tests during coordinated infra validation before merge; the
reviewed head is the sole parent of the exact merge SHA.

## Acceptance mapping

- `VOC-120-AC-00` / `VOC-120-EV-00` — known `.pyc` untracked on caller remediation
  carrier; `infra/scripts/cloudflare_origin_port_remap.py` unchanged.
- `VOC-120-AC-01` / `VOC-120-EV-00` — caller `.gitignore` verified by hygiene
  validator and unit tests.
- `VOC-120-AC-02` / `VOC-120-EV-00` — infra `.gitignore` merged at
  `9c01779ce166789d7d32eb47c502d42234d82a4c` with required patterns; infra unit
  tests pass at that merge SHA (290/290 discovery, 6/6 targeted).
- `VOC-120-AC-03` / `VOC-120-EV-00` — positive/negative hygiene tests in both
  repositories; governance hook fails closed on tracked bytecode and missing
  ignore patterns; caller validator inspects indexed paths only.
- `VOC-120-AC-04` / `VOC-120-EV-00` — no product/deploy/credential/routing
  changes; existing governance gates preserved.
- `VOC-120-AC-05` / `VOC-120-EV-00` — fixture pin intentionally unchanged;
  non-consumption recorded above; coordinated infra delivery merged and
  independently reviewed as recorded in the shared-infra carrier table.
