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

## Cutting a release

`main` is always deployable, but production deploys are deliberate. To cut a
versioned checkpoint, run the **release** workflow (Actions → release → Run
workflow) and pick a `patch` / `minor` / `major` bump. It:

1. computes the next `vX.Y.Z` from the latest tag,
2. bumps `package.json` versions and regenerates `CHANGELOG.md` (via `cliff.toml`),
3. opens and admin-merges a `release/vX.Y.Z` PR,
4. pushes the tag and publishes a GitHub Release with generated notes,
5. dispatches `deploy-production` for the new tag (unless you clear the `deploy` box).

`CHANGELOG.md` released sections are generated — never hand-edit them. Commit
grouping keys off Conventional Commit prefixes, so keep squash-merge subjects
conventional (`pr-title` already enforces this on PR titles).

Requires the `RELEASE_PR_TOKEN` repository secret (a GitHub App / fine-grained PAT
that can merge into `main` past the ruleset). Run with `dry_run` to preview the
version and changelog without pushing anything.

See `docs/development.md` for local setup and `AGENTS.md` for the full workflow,
deploy process, and repo conventions.
