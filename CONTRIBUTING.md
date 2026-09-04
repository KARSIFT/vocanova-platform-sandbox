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

Every non-draft PR is reviewed automatically by Claude (`claude-code-review.yml`)
and Codex. Their comments are advisory, not merge gates: neither can block a
merge, and turning either off doesn't change whether a PR can land. Tag
`@claude` (or `@codex`) in a PR comment for a follow-up review or a question.

CodeRabbit, Cubic, and Greptile were part of the review fleet but are
disabled as of 2026-09-04 - all three need a paid plan to keep reviewing a
repo at this point, and this project doesn't carry paid subscriptions for
them. CodeRabbit's config (`.coderabbit.yaml`) is disabled in place, not
deleted; Cubic and Greptile have no in-repo toggle (pure GitHub Apps) - they
stay installed on the org but idle unless someone re-enables them from each
vendor's own dashboard or a paid plan is bought.

A PR that touches `apps/web` also gets a live preview: Vercel builds it and
posts the preview URL as a PR comment (frontend only, pointed at the shared
staging API — there's no per-PR backend). This is separate from staging/
production, which stay on the VPS via the deploy workflows below.

Merging is automatic. `auto-merge.yml` turns on GitHub auto-merge for every
non-draft PR, and the PR lands by itself once the `require-pr-and-checks` ruleset
is satisfied: the required status checks (`ci-web`, `ci-api`, `controlled-signup
OAuth callback E2E`, `All action refs are SHA-pinned`, `Architecture boundaries do
not regress`) go green and the merge queue re-runs them against the real
post-merge result. No approving review is required, so a green PR can land a few
minutes after CI finishes with no manual step.

To hold a PR back, add the `hold` label (auto-merge is switched off until you
remove it); keep it a draft for the same effect. Squash is the only merge method.

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
