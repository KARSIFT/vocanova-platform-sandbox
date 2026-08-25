# VOC-120 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `scripts/governance/`, `tooling/governance/`, caller
  `infra/scripts/` (bytecode untrack only), and `KARSIFT/karsift-ai-infra`
  repository ignore rules/tests.
- Prerequisites: confirm the tracked blob is still present on the implementation
  base, confirm caller `.gitignore` still has `__pycache__/` and `*.py[cod]`,
  and confirm shared infra still lacks a repository-root `.gitignore`.
- Do not treat an untracked local `karsift-ai-infra/` checkout as this
  repository's tracked tree.
- Preserve exact-SHA independent review, risk classification, protected checks,
  and retry limits.
- Do not modify `infra/scripts/cloudflare_origin_port_remap.py`.

## File reconciliation and implementation sequence

### T00 — Untrack caller bytecode, add portable infra ignores, and add hygiene validation

| Target | Action | Notes |
|--------|--------|-------|
| `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc` | untrack/remove from the index | Known blob from issue #987; also untrack any other tracked bytecode/cache paths found at implementation time |
| `infra/scripts/cloudflare_origin_port_remap.py` | preserve | No source-behavior change |
| `.gitignore` | verify/keep | Must retain `__pycache__/` and `*.py[cod]` |
| `KARSIFT/karsift-ai-infra/.gitignore` | create | Repository-root portable Python cache patterns per `VOC-120-D02` / `VOC-120-D03` |
| `KARSIFT/karsift-ai-infra/tests/test_python_cache_ignore.py` | create | Assert required ignore patterns; name may vary in the same directory |
| `scripts/governance/validate-governance.sh` | modify | Invoke the caller bytecode-hygiene check so the documented governance command fails closed |
| `tooling/governance/validate_python_bytecode_hygiene.py` | create | Tracked-path scan plus caller ignore-rule assertions; name may vary in the same directory |
| `tooling/governance/tests/test_python_bytecode_hygiene.py` | create | Positive tree, negative tracked-bytecode, and missing-ignore-pattern coverage |
| `tooling/governance/fixtures/karsift-ai-infra/**` | pin only if consumed | Do not copy `.gitignore` into the policy fixture merely to force a pin |
| `specs/changes/VOC-120-.../t00-evidence.md` | create/update | Record commands, results, infra SHA, and pin applicability |

Ordered steps:

1. In `KARSIFT/karsift-ai-infra`, add a repository-root `.gitignore` covering
   `__pycache__/` and `*.py[cod]` (or equivalent rules that cover `*.pyc` and
   related variants). Add deterministic tests that fail closed when those
   patterns are missing. Run the infra unit suite. Open one reviewed infra PR
   that `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and does not use a
   closing keyword.
2. In the caller, untrack
   `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc` and
   any other tracked bytecode/cache artifacts found by `git ls-files`. Do not
   edit the Python source.
3. Verify caller `.gitignore` still contains `__pycache__/` and `*.py[cod]`.
4. Add caller validation that scans tracked files (not the working tree) and
   asserts the ignore rules. Hook it from `validate-governance.sh` and cover it
   with `tooling/governance/tests/` including negative fixtures.
5. If and only if the caller fixture actually consumes the infra change, merge
   the reviewed infra PR first, then sync and pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA.
   Otherwise leave the existing pin unchanged and record the non-consumption.
6. Run applicable validation in both repositories and record results in
   `t00-evidence.md`.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'
# and a targeted ignore-rule test if one is added, for example:
python3 -m unittest tests.test_python_cache_ignore

# In this caller repository:
git ls-files '**/__pycache__/**' '*.pyc' '*.pyo' '*.pyd'
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

If implementation adds a narrower targeted hygiene command, record the exact
command in `t00-evidence.md` and run it in addition to the suite above.

Independent verifier (exact reviewed task PR SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- the documented `.pyc` is no longer tracked and no other tracked bytecode/cache
  artifacts remain;
- `infra/scripts/cloudflare_origin_port_remap.py` is unchanged;
- caller ignore rules remain;
- infra has a repository-root `.gitignore` covering the required patterns;
- positive and negative hygiene tests pass;
- exact-SHA review, risk floors, protected checks, and retry limits remain;
- the caller fixture pin equals the exact reviewed infra merge when the fixture
  consumes the change, or evidence records why the pin was not applicable.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** Protected branches stop carrying the generated
  bytecode blob; fresh infra clones receive portable cache ignore rules.
- **Rollback trigger:** Hygiene validation false-positives block legitimate
  tracked files, or ignore rules become incorrect/contradictory.
- **Rollback mechanism:** Revert the ignore-rule and validation commits. Do not
  re-commit bytecode as the rollback of the deletion unless reverting the entire
  caller PR for an unrelated blocking defect.
- **Last-known-good reference:** Current `main`/`develop` trees before VOC-120
  implementation lands.
