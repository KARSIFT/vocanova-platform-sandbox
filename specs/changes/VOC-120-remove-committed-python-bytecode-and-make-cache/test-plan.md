# VOC-120 — Test Plan

## VOC-120-TEST-00 — Caller tracked tree has no Python bytecode/cache artifacts

- Covers: `VOC-120-AC-00`
- Preconditions: implementation branch with the known blob untracked
- Procedure: Run `git ls-files` for `__pycache__/`, `*.pyc`, `*.pyo`, and
  `*.pyd`. Confirm
  `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc` is
  absent from the index.
- Expected result: no tracked bytecode/cache artifacts remain.
- Evidence: `VOC-120-EV-00`

## VOC-120-TEST-01 — Caller ignore rules still cover Python cache artifacts

- Covers: `VOC-120-AC-01`
- Preconditions: caller `.gitignore` available
- Procedure: Assert the file contains `__pycache__/` and a rule that covers
  `*.pyc` and related variants (`*.py[cod]` is sufficient).
- Expected result: existing portable caller ignore rules remain and are not
  weakened.
- Evidence: `VOC-120-EV-00`

## VOC-120-TEST-02 — Shared infrastructure repository ignore rules exist

- Covers: `VOC-120-AC-02`
- Preconditions: primary `KARSIFT/karsift-ai-infra` checkout of the reviewed
  infra revision
- Procedure: Assert a repository-root `.gitignore` exists and contains
  `__pycache__/`, coverage of `*.pyc`, and related bytecode variants.
- Expected result: portable ignore rules are cloned with the repository and do
  not depend on `.git/info/exclude`.
- Evidence: `VOC-120-EV-00`

## VOC-120-TEST-03 — Tracked bytecode fails caller validation

- Covers: `VOC-120-AC-03`
- Preconditions: synthetic fixture or temporary index containing a tracked
  `.pyc` or `__pycache__/` path
- Procedure: Run the caller hygiene validator against that fixture.
- Expected result: validation fails closed and names the tracked artifact class
  without dumping bytecode contents or secrets.
- Evidence: `VOC-120-EV-00`

## VOC-120-TEST-04 — Missing ignore patterns fail closed

- Covers: `VOC-120-AC-03`
- Preconditions: synthetic `.gitignore` missing `__pycache__/` and/or `*.pyc`
  coverage
- Procedure: Run caller and infra ignore-rule assertions against the incomplete
  file.
- Expected result: both suites fail closed until required patterns are present.
- Evidence: `VOC-120-EV-00`

## VOC-120-TEST-05 — Python source behavior is unchanged

- Covers: `VOC-120-AC-00`
- Preconditions: caller implementation diff
- Procedure: Confirm `infra/scripts/cloudflare_origin_port_remap.py` is not
  modified or deleted. Confirm the task diff does not change application,
  deploy, credential, provider, or model-routing files except ignore/validation
  hygiene in scope.
- Expected result: only generated-artifact cleanup, ignore rules, validation,
  tests, and this package's evidence change.
- Evidence: `VOC-120-EV-00`

## VOC-120-TEST-06 — Safety gates remain unchanged

- Covers: `VOC-120-AC-04`
- Preconditions: final implementation branch
- Procedure: Run governance validation and inspect the diff for preserved
  exact-SHA verification, deterministic risk-floor, protected-check, and retry
  language. Confirm no product/deploy/credential/routing behavior change.
- Expected result: hygiene changes land without weakening existing safety gates.
- Evidence: `VOC-120-EV-00`

## VOC-120-TEST-07 — Fixture pin matches consumption

- Covers: `VOC-120-AC-05`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: If the fixture consumed the infra change, assert the pin equals the
  exact reviewed infra merge. If not, assert the pin is unchanged and evidence
  records why the fixture did not consume `.gitignore`/ignore tests.
- Expected result: no silent fixture drift; pin updates are exact-SHA and only
  when applicable.
- Evidence: `VOC-120-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
