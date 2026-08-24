# VOC-115 — Test Plan

## VOC-115-TEST-00 — Planner drafts one task for an ordinary coherent request

- Covers: `VOC-115-AC-00`
- Preconditions: planner prompt / fixture regression harness available
- Procedure: Use the deterministic planner fixture or prompt regression to draft a
  normal coherent feature/bug request; assert the resulting package contains exactly
  one task by default.
- Expected result: one package and one end-to-end task are drafted without extra
  code/docs/tests/evidence task fragmentation.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-01 — Default task contains code, tests, and docs together

- Covers: `VOC-115-AC-00`, `VOC-115-AC-01`
- Preconditions: planner/template regression fixture
- Procedure: Draft a request whose implementation obviously includes code, tests, and
  docs; assert the package still proposes one task unless a valid split boundary is
  recorded.
- Expected result: code, tests, docs, and same-carrier evidence stay together by
  default.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-02 — Evidence-only fragmentation is not drafted by default

- Covers: `VOC-115-AC-01`
- Preconditions: same as above
- Procedure: Draft a request where acceptance evidence can be gathered by the same
  implementation carrier or its governed workflow; assert the package does not create
  an extra task whose only purpose is carrying evidence.
- Expected result: no evidence-only follow-up task when same-carrier evidence is
  feasible.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-03 — Missing split reason after the first task fails closed

- Covers: `VOC-115-AC-02`
- Preconditions: deterministic package-validation or fixture parser tests updated
- Procedure: Validate a synthetic package with at least two tasks where the second
  task has no explicit allowed split reason.
- Expected result: validation/test fails closed with a reason indicating missing
  split justification.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-04 — Invalid split reasons are rejected and >3 tasks are flagged

- Covers: `VOC-115-AC-02`
- Preconditions: updated validator/fixture tests
- Procedure: Validate one synthetic package that uses an invalid reason such as
  "docs vs code" or "small", and another that declares more than three tasks without
  exceptional package-level justification.
- Expected result: invalid reasons are rejected; excessive task count is flagged as
  exceptional and cannot silently pass as ordinary planning.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-05 — Justified multi-task package preserves sequential advancement

- Covers: `VOC-115-AC-03`
- Preconditions: fixture package with a valid split reason and multiple tasks
- Procedure: Run or inspect adoption/auto-advance fixture coverage to confirm ordered
  task issue creation and dependency sequencing remain correct for a justified
  multi-task package.
- Expected result: later tasks remain blocked until predecessor completion proof;
  one-task default does not break legitimate multi-task packages.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-06 — In-scope causal remediation remains under the active package

- Covers: `VOC-115-AC-04`
- Preconditions: updated docs/prompt/fixture policy examples
- Procedure: Validate a scenario modeled on implementation/review/merge/promotion
  failure where the discovered defect is causally related and stays within the
  original objective, acceptance criteria, risk ceiling, and protected-area scope.
- Expected result: package/policy guidance allows remediation to remain under the
  active package/carrier.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-07 — Unrelated or authority-expanding follow-up still requires a new plan

- Covers: `VOC-115-AC-04`
- Preconditions: same policy/fixture coverage
- Procedure: Validate examples where the new work is unrelated, changes product
  intent, increases risk/authority, or cannot satisfy the original acceptance
  criteria.
- Expected result: docs/tests/prompt guidance require a new issue/plan rather than
  silently broadening the active package.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-08 — Safety gates remain unchanged

- Covers: `VOC-115-AC-05`
- Preconditions: final implementation branch
- Procedure: Run governance validation and the updated governance test suite; inspect
  changed docs/prompts for preserved exact-SHA verification, deterministic risk-floor,
  protected-branch, and fail-closed language.
- Expected result: efficiency policy changes land without weakening existing safety
  gates.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-09 — Related skills remain one plan and one task

- Covers: `VOC-115-AC-06`
- Preconditions: updated primary planner prompt, package validator, and pinned
  caller fixture
- Procedure: exercise the planner/validator regression with one request to add
  several related agent skills plus their configuration, adapters, documentation,
  and tests.
- Expected result: one package contains one end-to-end task; the number of skills,
  files, directories, components, and changed lines does not create extra tasks.
- Evidence: `VOC-115-EV-00`

## VOC-115-TEST-10 — Adoption-compatible YAML parsing fails before plan merge

- Covers: `VOC-115-AC-05`
- Preconditions: governance validation runs against a plan diff containing a new or
  modified `change.yaml`
- Procedure: validate one syntactically valid package and one fixture with an
  unescaped apostrophe in a single-quoted YAML scalar using the same PyYAML loader as
  adoption.
- Expected result: the valid package passes; invalid YAML fails the plan gate with a
  localized parse error before merge or adoption.
- Evidence: `VOC-115-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
