# VOC-051 — Release Plan

## Release and deployment authorization

Not authorized by this package. Merging any `VOC-051-T*` task's pull request
into `develop` does not itself authorize production deployment or activation
of the scheduled monitoring workflow's real effect (which depends on the
`SENTRY_API_AUTH_TOKEN` secret and the per-environment DSN secrets actually
being provisioned by a human — see `implementation-plan.md`'s "Deployment and
rollback" section on the safe-by-default rollout).

Per `AGENTS.md`'s "Release and deployment authority" (effective 2026-08-08),
once this package's full task roster closes, promotion from `develop` to
`main` and the resulting production deployment happen automatically via
`karsift-ai-infra`'s `release.yml`/`deploy-production.yml` pipeline — CI and
independent review passing on every task PR is the gate, not a separate
founder `approved` comment. This package does not request or need any
additional manual deploy step beyond that already-delegated automatic path.

## Preconditions, monitoring, and outcome

- Exact revision: the final merged SHA of each `VOC-051-T*` task PR,
  recorded at merge time.
- Checks: `pnpm validate` (T01), workflow-syntax/dry-run verification (T02),
  `scripts/governance/validate-governance.sh` and
  `scripts/governance/classify-change-risk.sh` (T04 and overall).
- Approvals: independent verification per `CLAUDE.md`; no standing
  founder/steward approval required solely for being R3 under active A-003,
  but the reviewing human should confirm the R3 proposal against the actual
  detected risk floor at adoption time (see `change.yaml`'s
  `blocking_reasons`) and confirm the open questions in `specification.md`
  are resolved, not merely acknowledged.
- Staged evidence: `VOC-051-EV-00` through `VOC-051-EV-04` per
  `impact-analysis.md`/`test-plan.md`.
- Monitoring: the scheduled workflow's own GitHub Actions run history is its
  primary health signal — a run that fails outright (per
  `impact-analysis.md`'s `VOC-051-R02`) is itself visible without additional
  tooling; the founder should periodically confirm the hourly cron is
  actually firing (GitHub can silently disable scheduled workflows in
  inactive repositories after 60 days with no other activity — not currently
  a risk for this repository, but worth naming for future reference).
- Outcome owner: unassigned; to be recorded at adoption time.

## Rollback

- Trigger: duplicate/spam issue creation, a Sentry SDK regression in
  `apps/web`, or a leaked secret/overly broad token scope discovered after
  merge.
- Mechanism: revert the relevant task's merge commit; for the scheduled
  workflow specifically, disabling its `schedule:` trigger (or the workflow
  file entirely) stops all further effect immediately with no other system
  depending on it continuing to run.
- Accountable owner: unassigned; to be recorded at adoption time.
- Validation: confirm no further scheduled runs occur post-rollback; confirm
  `apps/web` continues to function normally with Sentry wiring reverted or
  disabled.
- Last-known-good reference: `develop`'s tip immediately prior to this
  package's first merged task PR.

## Independent verification, human approvals, and closure

- Verifier result: recorded per task PR by Claude Code per `CLAUDE.md`,
  bound to the exact reviewed commit SHA.
- R3/R4 approvals: no standing technical-steward or founder approval is
  required solely because this package is R3, under active A-003;
  strengthened applicable controls and independent verification are
  required and are the actual gate here (see `implementation-plan.md`'s
  validation section). No R4 consequence is identified in this package —
  if the reviewing human's read of the open questions surfaces one (e.g. a
  privacy-policy gap from the new browser-data flow), that escalates to
  founder review before adoption, not after.
- Remaining hosted controls: the real `SENTRY_API_AUTH_TOKEN` and
  per-environment DSN secret values must be provisioned by a human with
  repository secret-write access — no agent obtains or enters these values
  (per `AGENTS.md`'s "Safety" section).
- Closure evidence: all `VOC-051-AC-*` acceptance criteria show `Result:
  satisfied` with linked evidence; the scheduled workflow's first several
  real hourly runs (post-secret-provisioning) show correct, non-duplicating
  behavior.

Repository merge, release, activation (secret provisioning), and closure are
tracked separately and are not conflated with one another in this record.
