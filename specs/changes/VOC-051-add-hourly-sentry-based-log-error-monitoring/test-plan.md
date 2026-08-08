# VOC-051 — Test Plan

## VOC-051-TEST-00 — `apps/web` sends a test error to Sentry when a DSN is configured

- Covers: `VOC-051-AC-00`
- Preconditions: a non-production Sentry DSN (test project, not the real
  staging/production DSN) configured in a local/preview environment
- Procedure: trigger a deliberate client-side and a deliberate server-side
  (Route Handler/Server Action) error in a local build with the test DSN
  set; confirm both appear as distinct events in the test Sentry project
  within a reasonable time
- Expected result: both events appear, tagged with the configured
  environment value
- Evidence: `VOC-051-EV-00`

## VOC-051-TEST-01 — `apps/web` does not error, warn noisily, or expose a debug overlay when no DSN is configured, and the debug overlay is disabled in production/staging builds regardless

- Covers: `VOC-051-AC-00`
- Preconditions: `NEXT_PUBLIC_SENTRY_DSN` (or chosen variable name) unset
- Procedure: run `apps/web`'s build and a smoke page load with the variable
  unset; separately, run a production-mode build with a DSN set and confirm
  no Sentry debug/error overlay renders to the end user
- Expected result: app functions normally with no DSN; no debug overlay
  visible in a production-mode build regardless of DSN presence
- Evidence: `VOC-051-EV-00`

## VOC-051-TEST-02 — The Sentry API auth token's configured scope is read-only and least-privilege

- Covers: `VOC-051-AC-01`
- Preconditions: the token's scope as configured in Sentry's own token
  management UI/API, and `VOC-051-T02`'s workflow-file documentation of the
  expected scope
- Procedure: inspect the token's actual granted scopes against Sentry's own
  scope list; confirm no write/admin scope is present and confirm the scope
  is sufficient for (and not broader than justified by) the workflow's
  actual API calls, per the implementer's documented justification
  (`specification.md`'s open question 2)
- Expected result: scope is read-only, minimum sufficient for the workflow's
  actual calls, with the choice documented and justified
- Evidence: `VOC-051-EV-01`

## VOC-051-TEST-03 — The scheduled workflow queries both the staging and production Sentry environments

- Covers: `VOC-051-AC-02`
- Preconditions: test Sentry projects/environments standing in for staging
  and production, each with at least one seeded unresolved issue
- Procedure: run the workflow via `workflow_dispatch:` against the test
  setup; inspect its logs/output for evidence both environments were queried
- Expected result: both environments' seeded issues are found
- Evidence: `VOC-051-EV-01`

## VOC-051-TEST-04 — A genuinely new Sentry problem produces exactly one unlabeled GitHub issue with sufficient reproduction/diagnosis detail

- Covers: `VOC-051-AC-03`
- Preconditions: a test Sentry issue not previously seen by the workflow; a
  scoped test repository or dry-run mode that does not write to the real
  issue tracker
- Procedure: run the workflow against the seeded new issue; inspect the
  created (or dry-run-simulated) GitHub issue's labels and body
- Expected result: exactly one issue created, with no labels applied, and a
  body containing the Sentry issue link, affected environment, and
  available root-cause/stack-trace detail per `AGENTS.md`'s bug-reporting
  requirement
- Evidence: `VOC-051-EV-01`

## VOC-051-TEST-05 — Repeated runs against the same underlying Sentry issue create no second GitHub issue

- Covers: `VOC-051-AC-04`
- Preconditions: same seeded Sentry issue as `VOC-051-TEST-04`, already
  resulting in one open GitHub issue from a prior run
- Procedure: run the workflow again (`workflow_dispatch:`) with no new
  underlying Sentry issue introduced; also test a variant where the
  previously created issue's title has been manually edited by a human,
  confirming the duplicate-check still matches via the embedded stable
  Sentry issue-ID marker, not title text alone
- Expected result: no second issue is created in either variant
- Evidence: `VOC-051-EV-01`

## VOC-051-TEST-06 — The scheduled workflow file declares no SSH-related secret

- Covers: `VOC-051-AC-05`
- Preconditions: the merged `.github/workflows/error-monitoring.yml` file
- Procedure: `grep -i ssh .github/workflows/error-monitoring.yml` (or
  equivalent search for `STAGING_SSH_`/`PRODUCTION_SSH_`-style secret
  references, `appleboy/ssh-action`, `appleboy/scp-action`)
- Expected result: no match
- Evidence: `VOC-051-EV-01`

## VOC-051-TEST-07 — DOC-11 §1's amendment accurately and non-destructively records the new mechanism

- Covers: `VOC-051-AC-06`
- Preconditions: `VOC-051-T04`'s merged documentation change
- Procedure: read the amended `docs/operations/11-devops-and-ci-cd.md` §1;
  confirm the existing row's history is annotated, not deleted, per this
  repository's stated amendment convention, and that the new content
  accurately describes what `VOC-051-T01`-`T03` actually built (not what was
  originally planned, if implementation diverged)
- Expected result: amendment present, accurate, non-destructive of prior
  history
- Evidence: `VOC-051-EV-02`

## VOC-051-TEST-08 — Deterministic validation commands pass against the exact implemented diff

- Covers: `VOC-051-AC-07`
- Preconditions: the final revision of each task's pull request
- Procedure: run `pnpm validate` (`apps/web` scope) for `T01`;
  `bash scripts/governance/validate-governance.sh` and
  `bash scripts/governance/classify-change-risk.sh` for `T04` and for the
  overall task-scoped file list
- Expected result: all commands exit zero; the real detected risk floor is
  recorded and compared against this package's proposed `R3`, with any
  mismatch escalated to the reviewing human rather than silently accepted
- Evidence: `VOC-051-EV-03`

## VOC-051-TEST-09 — Sentry organization/plan capacity is confirmed before dependent tasks proceed

- Covers: (gates `VOC-051-AC-00`, `VOC-051-AC-01` via `VOC-051-T00`)
- Preconditions: access to the founder's actual Sentry organization/plan
  details (a human/founder action, not an agent credential)
- Procedure: confirm project/token-scope availability against the actual
  plan; record the result
- Expected result: a written confirmation (or a documented blocking
  constraint) exists before `T01`/`T02` begin
- Evidence: `VOC-051-EV-04`

Include positive, negative, authorization, failure, migration, accessibility,
and rollback coverage as applicable. Tests must not use secrets or production
data — all Sentry/GitHub-API test procedures above use dedicated test
projects/repositories or `workflow_dispatch:` dry-run modes, never the real
production/staging Sentry projects or the real repository's issue tracker for
throwaway test issues.
