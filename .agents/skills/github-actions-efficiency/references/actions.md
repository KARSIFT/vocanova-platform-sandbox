# GitHub Actions efficiency — repository reference

Load when auditing `.github/workflows/` in vocanova-platform-sandbox.

## Audit order

1. `.github/workflows/*.yml` — triggers, concurrency, caches, matrices
2. `docs/operations/11-devops-and-ci-cd.md` — documented CI expectations
3. `docs/development.md` — validation commands workflows should invoke
4. `gh run list` / `gh run view --json` metadata when `gh` is available

## Common waste in this repo

- Missing `pnpm` cache keyed on `pnpm-lock.yaml`
- Missing Go module cache for `apps/api`
- Workflows running on unrelated path changes
- Duplicate lint/test jobs across workflows
- Heavy jobs (Lighthouse, Playwright) without path or label gates

## Trigger scoping

- Use workflow-level `paths` / `paths-ignore` when entire workflows are file-class specific.
- Use job-level `if` with path filters when event-level filters are too coarse.
- Prefer explicit changed-file detection for reliability over brittle glob tricks.

## Matrix reduction

- Full matrix for release or explicit compatibility validation
- Single representative leg for ordinary feature work
- Skip expensive browser jobs when only docs or governance change

## Safe-change rules

- Never remove `scripts/governance/` or foundation test gates without an approved package.
- Never hide deploy or monitoring sync workflows required by `infra/`.
- Treat severity drift in accessibility or Lighthouse thresholds as a regression risk.

## Live validation

- `workflow_dispatch` smoke run when available
- Confirm `concurrency` cancels superseded PR runs
- Confirm path-gated jobs skip in the Actions UI — do not assume from YAML alone

Do not paste raw CI logs into issues or chat. Summarize job names, conclusions, and durations only.
