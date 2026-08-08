# VOC-050 — Add Hourly Sentry-Based Log/Error Monitoring Agent: Specification

## Objective and requirement source

[Issue #392](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/392) asks
for a scheduled process that notices new production/staging application errors
without waiting for a human or an oversight loop to spot them, and turns each
genuinely new problem into a governed GitHub issue that `plan-from-issue` (per
`AGENTS.md`'s "Reporting a bug found outside the normal loop") can pick up.

`docs/operations/11-devops-and-ci-cd.md` (DOC-11) §1 already names Sentry as this
repository's chosen error-monitoring tool ("Error monitoring | Sentry | Unchanged",
present in both the superseded v1.0 and the current v1.1 infrastructure table) and
requires "Separate ... Sentry environments per environment tier." DOC-11 does not
itself specify a mechanism for turning Sentry data into action — this package
proposes that mechanism as new scope, not something DOC-11 already authorizes on
its own.

`apps/api` already has a working, narrower piece of this: `apps/api/go.mod` depends
on `github.com/getsentry/sentry-go v0.48.0`, `apps/api/app/api/production.go` reads
`SENTRY_DSN`/`SENTRY_ENVIRONMENT`/`SENTRY_RELEASE` and exposes a manual
`/ops/monitoring/sentry-test` probe endpoint (gated to the `production` environment
and a `MONITORING_TEST_TOKEN`), and `.github/workflows/deploy-production.yml`
already injects `PRODUCTION_SENTRY_DSN` and `PRODUCTION_SENTRY_RELEASE` onto the
production host at deploy time. `apps/web` has none of this today (confirmed:
`grep -ri sentry apps/web` at drafting time returns nothing, and
`.github/workflows/deploy-staging.yml` never references any `SENTRY_*` secret at
all — staging has no Sentry wiring on either side yet).

## Scope and non-goals

In scope:

1. Wire the `@sentry/nextjs` SDK into `apps/web`, mirroring `apps/api`'s existing
   pattern: DSN read from environment, a distinct value expected per environment
   tier (staging vs production), no-op/disabled behavior when the DSN is unset
   (same fail-safe default `apps/api/app/api/production.go`'s `SentryDSN` field
   already uses), and release/environment tagging consistent with
   `apps/api`'s `SentryEnvironment`/`SentryRelease` fields.
2. A new Sentry API auth token, scoped read-only and to the minimum access the
   scheduled workflow actually needs (see open question 2 below for the exact
   scope this drafting pass could not pin down further), stored as a new GitHub
   Actions secret — not a production application secret, since only the
   monitoring workflow reads it, never `apps/api` or `apps/web` at runtime.
3. A new scheduled (`cron`) GitHub Actions workflow, run hourly, that queries the
   Sentry API for new/unresolved issues across both the staging and production
   Sentry environments since its last recorded check, and — for each genuinely
   new problem — opens an unlabeled GitHub issue describing it, following this
   repository's existing bug-report convention (per `AGENTS.md`'s "Reporting a
   bug found outside the normal loop": reproduction detail/link, failing
   behavior, root cause if apparent from the Sentry event) so `plan-from-issue`
   can act on it without re-deriving the diagnosis.
4. A duplicate-check guard before opening an issue: search existing open GitHub
   issues for a matching title (or a stable Sentry issue-ID marker embedded in
   the body) before creating a new one, so repeated hourly runs do not spam
   duplicate issues for the same underlying Sentry issue across runs. This
   mirrors the idempotency pattern issue #392 itself points to in
   `karsift-ai-infra`'s `adopt.yml` "Open one issue per task" step, which reads
   an equivalent "does an issue for this already exist" guard before creating
   one (see `adopt.yml`'s duplicate-check logic near its "Open one issue per
   task" step).
5. Explicitly withholding SSH access to the staging/production hosts from this
   new scheduled workflow. `.github/workflows/deploy-staging.yml` already
   demonstrates the pattern this package deliberately does NOT reuse here:
   `STAGING_SSH_HOST`/`STAGING_SSH_USER`/`STAGING_SSH_PRIVATE_KEY` secrets used
   by `appleboy/ssh-action`/`appleboy/scp-action` for the deploy job's own
   on-demand, human-triggered/PR-triggered deploy steps. The new monitoring
   workflow must not declare or reference any of those secrets, or any new
   equivalent SSH credential, at all.

Out of scope (per issue #392's own "Out of scope" section, confirmed unchanged by
this drafting pass):

- Uptime/liveness monitoring (Better Stack/UptimeRobot per DOC-11 §1's separate
  "Uptime monitoring" row) — a distinct concern from application-level error
  tracking; not touched by this package.
- Any change to release-gating behavior. This package is purely observability
  that feeds the existing issue → `plan-from-issue` pipeline the same way any
  other live-found bug does; it does not touch `release.yml`,
  `deploy-production.yml`'s promotion logic, or any approval gate.
- Raw log access / SSH-based deep-dive tooling. That stays exactly as it exists
  today — a manual, human-triggered, on-demand tool used during deploys — and is
  not modified, replaced, or supplemented by this package's scheduled workflow.

## Risk and protected areas

Builder assessment: R3, proposed (see `change.yaml`'s `risk` field and its full
reasoning). This touches:

- A new secret (Sentry API auth token) — squarely
  `docs/governance/change-risk-classification.md`'s R3 "secrets" category.
- A new scheduled GitHub Actions workflow (CI/CD surface) — R3 "CI/CD ...
  governance enforcement, or agent authority" category, since this workflow
  autonomously opens GitHub issues without a human in the loop for each run
  (though note: it only *opens issues*, the lowest-authority action in this
  repository's own governed loop — it does not adopt, approve, merge, or deploy
  anything; `plan-from-issue` and a human founder/steward remain the actual
  gate on any resulting change).
- Production/staging Sentry data access (read-only) and browser-side error data
  starting to flow to a third-party vendor for the first time from `apps/web`
  (privacy consideration — see `impact-analysis.md`).

No database migration, no authentication/authorization code path, and no billing
code path is touched. `apps/web`'s Sentry wiring (T01) is additive and disabled by
default (mirrors `apps/api`'s existing no-op-when-unset pattern) — in isolation it
would likely classify closer to R2 (moderate blast radius, non-destructive,
independently reversible), but this package proposes the higher R3 floor for the
combined change per `docs/governance/change-risk-classification.md`'s "splitting a
change does not lower the classification" rule, since the workflow only becomes
useful once the web-side DSN exists.

Protected areas touched: `.github/workflows/` (new file), GitHub Actions secrets
(new secret), and (informationally, not edited without a governance-doc PR) DOC-11
§1's monitoring row, which this package's `T04` proposes to update to record the
new mechanism, consistent with `AGENTS.md`'s "Any change to workflow behavior...
must update every doc that describes that behavior in the same pull request" rule.

## Decisions, contradictions, security, and privacy

No `VOC-050-D00`-numbered decisions are defined here — this package is not
adopted, and per the template's own instruction, decisions are only defined
"after approval." The following are recorded as **open questions** for the
reviewing human instead of decisions this drafting pass makes unilaterally:

1. **GitHub API identity for issue-filing.** The scheduled workflow needs to call
   the GitHub API to search existing issues (duplicate-check) and create new
   ones. Options: (a) the workflow's own default `GITHUB_TOKEN` (simplest, but
   its default permissions/audit trail differ from a named actor), or (b) a
   GitHub App installation token, the same pattern `karsift-ai-infra`'s
   `adopt.yml` already uses (`steps.app-token.outputs.token || github.token`
   fallback). This package does not decide between them; `VOC-050-T02`'s
   implementer must propose one explicitly in that task's pull request
   description, not leave it implicit in the diff (same convention VOC-048-T01
   already follows for its own undecided design choice).
2. **Exact Sentry API auth token scope.** Sentry's own token-scope model (org
   auth tokens vs. project-level tokens, `project:read`/`event:read`-style
   scopes) is coarser-grained than "read-only issues/events for exactly the
   staging and production environments of exactly these two projects." This
   drafting pass did not verify against the founder's actual Sentry
   plan/organization which concrete scope combination satisfies "org-scoped,
   read-only, least privilege" from the issue while still being sufficient for
   the workflow's actual API calls. `VOC-050-T02`'s implementer must record the
   exact scope chosen and why in that task's pull request.
3. **Sentry plan/tier capacity.** Whether the founder's existing Sentry
   organization/plan supports adding an `apps/web` project and issuing an
   additional read-only org-level API token (some free/starter tiers restrict
   project count or token scopes) is unconfirmed by this drafting pass — see
   `VOC-050-DEP-00`. If the plan does not support it, `T01`'s scope may need to
   change (e.g. reusing an existing project instead of creating a new one) and
   must be revisited with the reviewing human before implementation, not
   silently worked around.
4. **"Since its last check" state.** The scheduled workflow needs some
   persisted marker of what it already scanned, so it does not either miss
   issues (query window too narrow) or re-scan everything every hour (relying
   purely on duplicate-check to filter, which works but is less efficient).
   Sentry's own issue API supports time-window and `is:unresolved`-style
   queries; whether the workflow persists a "last checked" timestamp (e.g. as
   a workflow artifact, a committed file, or simply always querying "last
   90 minutes" with the hourly cadence overlapping slightly for safety) is left
   to `VOC-050-T02`'s implementer to propose, with the duplicate-check guard
   (requirement 4 above) as the safety net regardless of which windowing
   approach is chosen.

Security and privacy: the new Sentry API auth token must be read-only and
org/least-privilege-scoped per the issue's own explicit requirement (item 2 of
the "Requested change" list) — the acceptance criteria and test plan both bind
this as an observable, checkable property (the token's configured scope, not
just a description of intent). No SSH credential is added for this workflow
(explicit non-goal, item 5). Browser-side Sentry events from `apps/web` may
include PII incidental to error context (user-visible URLs, possibly user IDs if
`apps/web` sets a Sentry user-context — `impact-analysis.md` covers this).

## Data, migrations, analytics, and accessibility

- **Data/migrations**: None. No database schema change. The "last checked"
  state (open question 4) is the only new persisted state this package
  introduces, and it lives in CI infrastructure (a workflow artifact, cache, or
  a small committed marker file), not the application database.
- **Analytics**: Not applicable — this package adds error-monitoring
  instrumentation, not product analytics. `apps/web`'s Sentry SDK integration
  must not be conflated with or repurposed as an analytics pipeline.
- **Accessibility**: Not applicable — no user-facing UI is added or changed by
  this package. The only end-user-visible surface risk is if the Sentry SDK's
  own client-side error overlay (development-mode only, per `@sentry/nextjs`'s
  own defaults) were accidentally left enabled in production, which
  `test-plan.md`'s `VOC-050-TEST-01` checks against explicitly.
