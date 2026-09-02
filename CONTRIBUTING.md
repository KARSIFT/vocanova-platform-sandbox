# Contributing

Vocanova uses a single permanent branch: `main`.

Create a working branch from `main`, using whichever prefix fits — `feature/`,
`fix/`, `docs/`, `refactor/`, `infra/`, `security/` — and open a PR back into `main`.
Squash-merge is the default.

Before opening a PR, run:

```bash
pnpm install
pnpm run validate
```

CI runs `ci-web`, `ci-api`, and (where relevant) `controlled-signup-oauth-e2e`,
`accessibility`, `lighthouse`, and `docker-smoke` on every PR — see `AGENTS.md` for
what each checks. All required checks must pass before merge.

Every non-draft PR is reviewed automatically by several assistants — Claude
(`claude-code-review.yml`), CodeRabbit, Cubic, Greptile, and Codex. Their
comments are advisory, not merge gates. Tag `@claude` (or `@codex`) in a PR
comment for a follow-up review or a question.

Merging goes through GitHub's merge queue, not a direct merge: once required checks
pass and the PR is approved, enqueue it (`gh pr merge --squash --auto`, or the
"Merge when ready" button in the UI). The queue re-runs required checks against the
real post-merge result before landing it on `main`, so a merge can take a few minutes
after approval rather than being instant.

Use the exact checked-in tool versions and a frozen lockfile (`pnpm install
--frozen-lockfile` in CI). Don't claim an unavailable check passed.

## Deploys

There is no separate "release" step. Every push to `main` deploys to **staging**
automatically (`deploy-staging.yml`). **Production** is a deliberate manual
dispatch: Actions → `deploy-production` → Run workflow (defaults are the canonical
hosts). Each production deploy builds SHA-tagged images and runs health checks and
a smoke suite, failing closed. To roll back, dispatch `deploy-production` again
against an earlier commit SHA (every past deploy's commit is in the workflow's run
history).

See `docs/development.md` for local setup and `AGENTS.md` for the full workflow,
deploy process, and repo conventions.
