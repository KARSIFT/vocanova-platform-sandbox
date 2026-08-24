# VOC-115-T00 — Evidence

Task: `VOC-115-T00` — Consolidate package/task defaults, causal-remediation policy,
validation, and regression coverage.

Do not record secrets, credentials, session values, OAuth material, personal data,
or complete CI logs.

## Changed surfaces

### Canonical policy and templates

- `AGENTS.md` — bounded in-scope causal remediation under an active package; unrelated
  work still requires a new issue/plan.
- `docs/operations/10-development-workflow.md` — removed automatic `L`/800-line task/PR
  split commands; size is a review signal subordinate to outcome/risk/rollback boundaries.
- `docs/operations/15-ai-native-product-and-engineering-operating-model.md` — replaced
  seven-task authentication example with one end-to-end task; documented split-reason
  allowlist and exceptional >3-task packages.
- `docs/governance/16-autonomous-development-operating-model.md` — one-package/one-task
  default and bounded in-scope causal remediation.
- `docs/templates/change-specification.md` — task IDs as minimum-sufficient outcome
  groupings; default `T00`.
- `specs/templates/change-package/tasks.md` — one-task default, split-reason rules,
  exceptional multi-task justification heading.

### Shared infra primary source (`KARSIFT/karsift-ai-infra`, same T00)

- `prompts/plan.md` — one-task default and explicit split-reason allowlist.
- `prompts/plan-review.md` — multi-task scrutiny, split reasons, causal-remediation boundary.
- `config/package_task_policy.py` — deterministic parser/validator.
- `config/package-task-policy-runner.py` — CLI used by workflows/tests.
- `.github/workflows/plan.yml` — planner retry loop enforces task policy after YAML parse.
- `tests/test_package_task_policy.py` — positive/negative/sequencing regressions.

Enforcement applies to newly drafted plan packages in `plan.yml`; historical adopted
packages are not retroactively rejected at adoption time.

### Caller mirrored fixture and tests

- `tooling/governance/fixtures/karsift-ai-infra/**` — synced copies of the infra changes above.
- `tooling/governance/fixtures/karsift-ai-infra/README.md` — VOC-115 contract note.
- `tooling/governance/tests/test_voc115_package_task_policy.py` — caller regressions for docs,
  prompts, workflows, and validator behavior.

## Adopted policy wording (summary)

- **Default:** one coherent objective → one package → one end-to-end implementation task →
  one outcome-sized implementation PR per repository when possible.
- **Task IDs:** minimum-sufficient outcome traceability groupings; not component/file/layer/skill
  buckets.
- **Stay together by default:** code, tests, docs, migration/config, and same-carrier evidence.
- **Allowed split reasons (tasks after the first):**
  `merge-order-dependency`, `independently-releasable-unit`,
  `distinct-owner-authority-risk-boundary`, `mutually-exclusive-execution-environment`,
  `post-merge-evidence-not-in-carrier`, `single-pr-review-size-boundary` (requires concrete
  reviewability explanation).
- **Exceptional packages:** more than three tasks require `## Package-level multi-task
  justification`.
- **In-scope causal remediation:** may remain under the active package/carrier only when bounded
  by original objective, acceptance criteria, risk ceiling, and protected-area scope.
- **Still requires new plan:** unrelated scope, changed product intent, authority/risk expansion,
  or work that cannot satisfy original acceptance criteria.

## Justified multi-task regression fixture

`karsift-ai-infra/tests/test_package_task_policy.py` and mirrored fixture copy exercise:

- `VOC-902-T00` / `VOC-902-T01` with `merge-order-dependency` split reason and preserved order.
- four-task package with and without package-level justification.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra source | `karsift-ai-infra/` in this T00 working tree |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Prior pin | `d3108dfdef34e2f98c028916e95c36130d329132` |
| New pin | pending exact reviewed `KARSIFT/karsift-ai-infra` merge for this T00 |

The caller fixture content is synchronized to the primary infra changes in this task. Update
`PINNED_SHA.txt` to the exact reviewed shared-infra merge SHA when that carrier lands; runtime
callers continue to `uses:` `KARSIFT/karsift-ai-infra/...@main`.

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `bash scripts/governance/validate-governance.sh` | pass | Repository foundation + monitoring declarations |
| `bash scripts/governance/classify-change-risk.sh` | pass | Detected path floor `R4` |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass | 156 tests |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc115*.py'` | pass | 10 tests |
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `karsift-ai-infra` | pass | 240 tests |
| `git diff --check` | pass | No whitespace or patch-format errors |

## Acceptance mapping

- `VOC-115-AC-00` / `VOC-115-EV-00` — planner prompt, fixture validator, and VOC-115 package
  prove one-task default.
- `VOC-115-AC-01` / `VOC-115-EV-00` — default task keeps code/tests/docs together; no
  evidence-only fragmentation by default.
- `VOC-115-AC-02` / `VOC-115-EV-00` — missing/invalid split reasons and >3-task packages fail
  closed; size-only reasons rejected.
- `VOC-115-AC-03` / `VOC-115-EV-00` — justified multi-task fixture preserves ordered tasks.
- `VOC-115-AC-04` / `VOC-115-EV-00` — docs/prompts distinguish bounded in-scope remediation vs
  new-plan follow-up.
- `VOC-115-AC-05` / `VOC-115-EV-00` — exact-SHA review, risk floors, protected-branch language,
  fail-closed validation, and adoption-compatible YAML parsing unchanged.
- `VOC-115-AC-06` / `VOC-115-EV-00` — related multi-skill request remains one task in validator
  regressions.
