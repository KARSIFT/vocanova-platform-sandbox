# VOC-115-T00 — Evidence

Task: `VOC-115-T00` — Consolidate package/task defaults, causal-remediation policy,
validation, and regression coverage.

Do not record secrets, credentials, session values, OAuth material, personal data,
or complete CI logs.

## Required evidence content

When T00 is implemented, record:

1. The exact canonical docs, template files, prompts, validators, and fixture tests
   changed.
2. The final one-package/one-task default wording adopted.
3. The explicit allowed split-reason rule implemented for tasks after the first.
4. The bounded policy wording for in-scope causal remediation vs unrelated or
   authority-expanding follow-up work.
5. Validation commands run and their outcomes.
6. Any justified multi-task regression fixture used to prove sequential advancement
   remained intact.

## Validation commands

Record the exact command lines and results for:

| Command | Result | Notes |
|---------|--------|-------|
| `bash scripts/governance/validate-governance.sh` | pending | Run if changed paths require it |
| `bash scripts/governance/classify-change-risk.sh` | pending | Expected to confirm protected governance/workflow floor |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pending | Include any narrower targeted governance test command if added |
| `git diff --check` | pending | No whitespace or patch-format errors |

## Acceptance mapping

- `VOC-115-AC-00` / `VOC-115-EV-00` — one coherent request drafts one task by default.
- `VOC-115-AC-01` / `VOC-115-EV-00` — code, tests, docs, and same-carrier evidence
  stay together by default.
- `VOC-115-AC-02` / `VOC-115-EV-00` — extra tasks require explicit split reasons.
- `VOC-115-AC-03` / `VOC-115-EV-00` — justified multi-task sequencing remains correct.
- `VOC-115-AC-04` / `VOC-115-EV-00` — in-scope causal remediation stays bounded under
  the active package while unrelated work still requires a new plan.
- `VOC-115-AC-05` / `VOC-115-EV-00` — exact-SHA review, risk floors, protected-branch
  gates, and fail-closed behavior remain unchanged.
