# VOC-051 — Tasks

## VOC-051-T00 — Confirm Sentry organization/plan capacity and record the exact per-environment DSN/project layout before any code change

- Requirement source: `specification.md`'s open questions 3; `change.yaml`'s
  `VOC-051-DEP-00`
- Acceptance criteria: (gates `VOC-051-AC-00`, `VOC-051-AC-01`)
- Tests: `VOC-051-TEST-09`
- Evidence: `VOC-051-EV-04`
- Status: complete — `t00-evidence.md` (`VOC-051-EV-04`) now records the
  `VOC-051-TEST-09` "written confirmation" outcome. `VOC-051-DEP-00` is
  resolved: the founder confirmed Sentry access, chose Layout B, created all
  four projects, created a scoped internal-integration token, and all five
  Sentry-related GitHub Actions secrets are set. See `t00-evidence.md` §3 for
  the full layout record.

Confirm, against the founder's actual Sentry organization, that adding an
`apps/web` project (or reusing an existing one, if a new project is not
available on the current plan) and issuing an additional read-only org-scoped
API auth token is actually possible. Record the exact resulting layout — how
many Sentry projects exist, which DSN belongs to which environment for both
`apps/api` and `apps/web` — as this task's evidence. This task makes no code
change; it only confirms preconditions and produces a written record other
tasks depend on. If the plan does not support the assumed layout, stop and
record the actual constraint for the reviewing human to resolve before
`VOC-051-T01`/`VOC-051-T02` proceed with a revised scope.

## VOC-051-T01 — Wire the Sentry SDK into `apps/web`, mirroring `apps/api`'s existing pattern

- Requirement source: `specification.md`'s scope item 1
- Acceptance criteria: `VOC-051-AC-00`, `VOC-051-AC-07`
- Tests: `VOC-051-TEST-00`, `VOC-051-TEST-01`, `VOC-051-TEST-08`
- Evidence: `VOC-051-EV-00`, `VOC-051-EV-03`
- Status: unblocked — `VOC-051-T00` is complete and `t00-evidence.md` §3's
  layout table is filled. Follow that table exactly: Layout B (four projects),
  `NEXT_PUBLIC_SENTRY_DSN` reading from `PRODUCTION_WEB_SENTRY_DSN`/
  `STAGING_WEB_SENTRY_DSN`, not the Layout-A/apps/api single-DSN-plus-tag
  pattern this task's own description below still describes generically.

Add the `@sentry/nextjs` SDK to `apps/web`, following the official Next.js
integration wizard's output shape for the App Router (instrumentation files,
`next.config.ts` wrapping), but hand-adapted rather than wizard-generated
where needed to fit this repo's existing conventions. Read the DSN from an
environment variable (name chosen to parallel `apps/api`'s `SENTRY_DSN`, e.g.
`NEXT_PUBLIC_SENTRY_DSN` — Next.js requires the `NEXT_PUBLIC_` prefix for any
value the browser bundle needs, since the DSN is not itself a secret but the
build-time/runtime distinction matters; the implementer must confirm and
document this choice), disabled (no-op, no client/server errors) when unset,
exactly like `apps/api/app/api/production.go`'s `SentryDSN` field already
behaves. Read environment/release tags analogous to `apps/api`'s
`SentryEnvironment`/`SentryRelease`. Update `apps/web/.env.example` with the
new variable(s), documented the same way `apps/api/.env.example` documents
`SENTRY_DSN` (optional, disabled-when-unset, example placeholder value).
Explicitly disable Sentry's client-side error/debug overlay in production
and staging builds. Add or update `.github/workflows/deploy-staging.yml` and
`.github/workflows/deploy-production.yml` to inject the correct per-
environment DSN, following the exact existing pattern
`deploy-production.yml` already uses for `PRODUCTION_SENTRY_DSN` (including
its "both-or-neither" partial-configuration guard) — do not invent a
different injection mechanism for `apps/web` than the one already proven for
`apps/api`. No `apps/web` UI/behavior changes beyond this instrumentation.

## VOC-051-T02 — Add the hourly scheduled workflow that queries Sentry and files deduplicated GitHub issues

- Requirement source: `specification.md`'s scope items 2-4; open questions 1, 2, 4
- Acceptance criteria: `VOC-051-AC-01`, `VOC-051-AC-02`, `VOC-051-AC-03`,
  `VOC-051-AC-04`, `VOC-051-AC-05`, `VOC-051-AC-07`
- Tests: `VOC-051-TEST-02`, `VOC-051-TEST-03`, `VOC-051-TEST-04`,
  `VOC-051-TEST-05`, `VOC-051-TEST-06`, `VOC-051-TEST-08`
- Evidence: `VOC-051-EV-01`, `VOC-051-EV-03`
- Status: unblocked — `VOC-051-T00` is complete. May proceed in parallel with
  `VOC-051-T01`, since this workflow's Sentry-side query does not depend on
  `apps/web`'s own instrumentation existing yet. Query all four projects
  (`prod-api`, `prod-web`, `stage-api`, `stage-web`) per `t00-evidence.md` §3's
  Layout B mapping, org slug `vocanova`, using the `SENTRY_API_TOKEN` secret.

Add a new GitHub Actions workflow (proposed path:
`.github/workflows/error-monitoring.yml`) triggered on an hourly `schedule:
cron:` entry (plus, for manual testing, a `workflow_dispatch:` trigger — this
is a read/query workflow with no destructive effect, so a manual trigger for
verification carries none of the risk a production-deploy manual dispatch
would). The workflow must:

- Use the new Sentry API auth token (`VOC-051-DEP-02`'s scope, confirmed by
  `VOC-051-T00`/this task) to query for new/unresolved issues across both the
  staging and production Sentry environments/projects, using whichever
  time-window or state-tracking approach the implementer proposes and
  justifies explicitly in this task's pull request description (open
  question 4).
- Before creating any GitHub issue, search existing open issues for a match
  keyed on a stable Sentry issue identifier embedded in the created issue's
  body (not solely on free-text title matching, per `impact-analysis.md`'s
  `VOC-051-R01`), following the same idempotency shape as
  `karsift-ai-infra`'s `adopt.yml` "Open one issue per task" step.
- Create genuinely-new-problem issues as plain, **unlabeled** issues (so
  `plan-from-issue` picks them up automatically per `pipeline.yml`), with a
  body containing: the Sentry issue's title/error type, a link back to the
  Sentry issue, the affected environment (staging or production), first-seen
  and event-count information Sentry provides, and any stack-trace/root-cause
  detail Sentry's API surfaces — enough detail that the planner "can act
  without re-deriving your diagnosis," per `AGENTS.md`'s explicit
  bug-reporting requirement.
- Declare a minimum `permissions:` block (at most `issues: write`; no
  `contents: write`, no `deployments:`, no `packages:`) and use the GitHub
  API identity the implementer selects and justifies per open question 1.
- Declare and use no SSH-related secret of any kind (`VOC-051-AC-05`).
- Fail the workflow run (non-zero exit) on a genuine Sentry API error,
  distinct from a legitimate zero-new-issues result, per
  `impact-analysis.md`'s `VOC-051-R02`.

## VOC-051-T03 — Add the Sentry API auth token as a GitHub Actions secret

- Requirement source: `specification.md`'s scope item 2
- Acceptance criteria: `VOC-051-AC-01`
- Tests: `VOC-051-TEST-02`
- Evidence: `VOC-051-EV-01`
- Status: unblocked and largely moot — `VOC-051-T00` is complete, and the
  founder already created and set the token secret directly (ahead of this
  task, through the founder-gate overseer session) as **`SENTRY_API_TOKEN`**,
  not this task's originally proposed `SENTRY_API_AUTH_TOKEN` name. `VOC-051-T02`'s
  workflow must reference the secret by its actual name, `SENTRY_API_TOKEN`.
  This task's remaining scope is just to document that name/scope choice in
  the package's README/impact analysis for the record, not to create anything.

Document the exact secret name the `VOC-051-T02` workflow expects —
**`SENTRY_API_TOKEN`** (already set; internal-integration token scoped to
`project:read` + `event:read` only, per `t00-evidence.md` §3/§4) — so the
record is accurate for future reference. If the secret is absent,
`VOC-051-T02`'s workflow must fail clearly and immediately (not silently
no-op the way `apps/api`'s optional `SENTRY_DSN` is allowed to, since this is
a required secret for the monitoring workflow's entire purpose, not an
optional enhancement) rather than running with a missing credential and
producing a confusing downstream error.

## VOC-051-T04 — Update DOC-11 §1's Error monitoring row to record the new mechanism

- Requirement source: `AGENTS.md`'s documentation-consistency rule
- Acceptance criteria: `VOC-051-AC-06`, `VOC-051-AC-07`
- Tests: `VOC-051-TEST-07`, `VOC-051-TEST-08`
- Evidence: `VOC-051-EV-02`, `VOC-051-EV-03`
- Status: pending (depends on `VOC-051-T01`-`VOC-051-T03` landing, or may land
  in the same pull request as `VOC-051-T02` if the implementer prefers a
  single combined PR for the workflow and its doc update — either ordering is
  acceptable as long as the doc update lands in the same PR as the behavior
  it describes, per `AGENTS.md`)

Add an amendment (following DOC-11's own existing amendment-note convention,
e.g. `VOC-051-§1-amendment`, in the same style as the existing
`VOC-032-§1-amendment` note) recording that DOC-11 §1's "Error monitoring |
Sentry" row now includes both `apps/api`'s existing backend error reporting
and `apps/web`'s new browser-side error reporting, plus the new hourly
scheduled Sentry-to-GitHub-issue monitoring workflow this package adds. Do
not silently rewrite the existing row's history — annotate, consistent with
this repository's stated convention for amending an approved document.

Tasks preserve scope, separation of duties, and rollback safety. No task may
be dispatched before this package is adopted.
