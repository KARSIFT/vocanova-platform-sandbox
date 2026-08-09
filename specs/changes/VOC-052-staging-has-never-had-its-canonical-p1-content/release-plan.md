# VOC-052 — Release Plan

## Release and deployment authorization

This package requests no new deployment authority. Once adopted and
implemented, each task's pull request follows the existing governed path:
independent review, then automatic merge into `develop` (per
`karsift-ai-infra`'s `merge-gate.yml`, already live), then — per AGENTS.md's
"Release and deployment authority" section — automatic promotion to `main` and
automatic production deployment once the package's task roster closes, with the
founder-approval comment path remaining available as a manual retry mechanism
only. This package does not alter any of that mechanism; it only adds deploy-time
content-seeding steps that run within the existing staging (and, conditionally,
production) deploy workflows.

## Preconditions, monitoring, and outcome

Precondition: `VOC-052-T00` (and, if in scope, `VOC-052-T02`) must be merged to
`develop` before a real staging (or production) deploy run can exercise the new
step. Monitoring: the workflow run's own log output (the seed step's exit code
and `apps/api/cmd/seed`'s row-count summary line) is the direct evidence source;
no new dashboard or alerting is introduced by this package. Outcome owner:
whoever implements `VOC-052-T00`/`T01` is responsible for confirming and
recording, in that task's own evidence file, that a real triggered staging
deploy ran the new step successfully and that
`tests/staging-e2e/core-loop.staging.spec.ts` subsequently passed in full.

## Rollback

Trigger: a staging (or production) deploy failure directly attributable to the
new seed step (e.g. a build failure, a runtime error unrelated to a genuine
schema mismatch). Mechanism: revert the `deploy-staging.yml` (and, if
applicable, `deploy-production.yml`) diff; because the seed's effect is additive
and idempotent, no data-level rollback is required, and doing so does not
disturb the already-verified `migrate.sh` / `seed-synthetic-smoke-user.sh`
sequence. Accountable owner: the implementer of the affected task. Validation:
confirm a subsequent staging deploy succeeds again with the revert applied.
Last-known-good reference: the `deploy-staging.yml` (or `deploy-production.yml`)
revision immediately preceding this package's merge.

## Independent verification, human approvals, and closure

Independent verification (per `CLAUDE.md`) must confirm, against the exact
implemented revision's commit SHA: the seed step is genuinely idempotent and
fail-closed as claimed; the header-comment documentation in the affected
workflow file(s) was updated to match the real new sequence; no unrelated change
was introduced; and (per this repository's active authority model, `a003-active`)
that no standing technical-steward or founder approval is being silently assumed
beyond what A-003 already delegates for routine R3 work, while any R4-level
consequence this drafting pass did not anticipate is escalated rather than
resolved by the implementer or reviewer alone. Repository merge into `develop`
and production release/deployment are not the same event as closure — closure
requires this package's acceptance-criteria results being recorded as passing
with their linked evidence, not merely a merged PR. If `VOC-052-T02` is included,
its own founder-decision precondition (`VOC-052-DEP-02`) must be recorded as
resolved, with the resolving human named, before that task's PR is opened.
