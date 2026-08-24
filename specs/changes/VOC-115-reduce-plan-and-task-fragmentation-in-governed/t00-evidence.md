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
- `docs/governance/16-autonomous-development-operating-model.md` — largest-safe-coherent
  package, minimum-sufficient maximal tasks, and bounded in-scope causal remediation.
- `docs/templates/change-specification.md` — task IDs as minimum-sufficient outcome
  groupings; default `T00`.
- `specs/templates/change-package/tasks.md` — one-task default, split-reason rules,
  exceptional multi-task justification heading.

### Shared infra primary source (`KARSIFT/karsift-ai-infra`, same T00)

- `prompts/plan.md` — largest-safe-coherent unit, maximal-task default, and split allowlist.
- `prompts/plan-review.md` — multi-task scrutiny, split reasons, causal-remediation boundary.
- `prompts/implement.md` and `prompts/review.md` — keep causally related implementation,
  tests, docs, configuration, contracts, and evidence together in the named task.
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
  prompts, workflows, validator behavior, justified multi-task advancement, and
  adoption-compatible YAML parsing.

## Adopted policy wording (summary)

- **Default:** choose the largest safe coherent package for the complete outcome. A broad or
  massive plan may contain several tasks, but it uses the minimum sufficient number of maximal
  tasks; one end-to-end task and implementation PR remains the default when possible.
- **Task IDs:** minimum-sufficient outcome traceability groupings; not component/file/layer/skill
  buckets.
- **Stay together by default:** code, tests, docs, migration/config, and same-carrier evidence.
- **Allowed split reasons (tasks after the first, with a concrete explanation):**
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

- a two-task package with a concretely explained `merge-order-dependency` boundary and
  preserved order;
- a four-task package with and without package-level justification.

Caller coverage then carries that justified two-task order into the existing
`next_roster_task` implementation and verifies that adoption writes the predecessor
dependency while auto-advance requires completion proof before selecting the next row.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra source | `KARSIFT/karsift-ai-infra` PRs #134 and #135 under this T00 |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Prior pin | `d3108dfdef34e2f98c028916e95c36130d329132` |
| New pin | `3fd40f52aba602fab8399482bc5b772731675d1a` |

The caller fixture content is synchronized to the exact final shared-infra merge SHA in
`PINNED_SHA.txt`; runtime callers continue to `uses:` `KARSIFT/karsift-ai-infra/...@main`.

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `bash scripts/governance/validate-governance.sh` | pass | Repository foundation + monitoring declarations |
| `bash scripts/governance/classify-change-risk.sh` | pass | Detected path floor `R4` |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass | 160 tests |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc115*.py'` | pass | 14 tests |
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `KARSIFT/karsift-ai-infra` | pass | 242 tests on merge source |
| `node --test scripts/foundation/*.test.mjs` (via `pnpm run test`) | pass | 333 tests, including revision-bound navigation evidence |
| `pnpm run build` | pass | packages, production web build, and Go API build |
| `pnpm run validate` | environment-limited | Format, lint, typecheck, 333 foundation tests, client tests, and middleware tests passed; two existing OAuth integration cases could not start PostgreSQL because Docker is unavailable in this WSL session. Repository CI retains the Docker-backed gate. |
| `git diff --check` | pass | No whitespace or patch-format errors |

## Acceptance mapping

- `VOC-115-AC-00` / `VOC-115-EV-00` — planner prompt, fixture validator, and VOC-115 package
  prove one-task default.
- `VOC-115-AC-01` / `VOC-115-EV-00` — default task keeps code/tests/docs together; no
  evidence-only fragmentation by default.
- `VOC-115-AC-02` / `VOC-115-EV-00` — missing/invalid split reasons and >3-task packages fail
  closed; size-only reasons rejected.
- `VOC-115-AC-03` / `VOC-115-EV-00` — justified multi-task fixture preserves ordered tasks,
  adoption dependency edges, completion-proof gating, and next-roster selection.
- `VOC-115-AC-04` / `VOC-115-EV-00` — docs/prompts distinguish bounded in-scope remediation vs
  new-plan follow-up.
- `VOC-115-AC-05` / `VOC-115-EV-00` — exact-SHA review, risk floors, protected-branch language,
  and fail-closed behavior remain; the plan gate and adoption both use PyYAML
  `safe_load`, with a valid package passing and an unescaped single-quoted apostrophe
  failing before the draft PR is opened.
- `VOC-115-AC-06` / `VOC-115-EV-00` — related multi-skill request remains one task in validator
  regressions.
