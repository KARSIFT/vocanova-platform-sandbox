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
what each checks. All required checks must pass before merge. Tag `@claude` in a PR
comment for an automated review.

Merging goes through GitHub's merge queue, not a direct merge: once required checks
pass and the PR is approved, enqueue it (`gh pr merge --squash --auto`, or the
"Merge when ready" button in the UI). The queue re-runs required checks against the
real post-merge result before landing it on `main`, so a merge can take a few minutes
after approval rather than being instant.

Use the exact checked-in tool versions and a frozen lockfile (`pnpm install
--frozen-lockfile` in CI). Don't claim an unavailable check passed.

See `docs/development.md` for local setup and `AGENTS.md` for the full workflow,
deploy process, and repo conventions.
