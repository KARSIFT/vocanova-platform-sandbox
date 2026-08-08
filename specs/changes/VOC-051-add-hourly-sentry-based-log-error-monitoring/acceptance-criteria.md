# VOC-051 — Acceptance Criteria

## VOC-051-AC-00 — `apps/web` reports errors to Sentry, using a distinct DSN per environment tier, and is silently disabled when no DSN is configured

- Requirement source: `specification.md`'s scope item 1
- Tasks: `VOC-051-T01`
- Tests: `VOC-051-TEST-00`, `VOC-051-TEST-01`
- Evidence: `VOC-051-EV-00`
- Result: pending

## VOC-051-AC-01 — A new Sentry API auth token exists as a GitHub Actions secret, and its configured scope is read-only and least-privilege, not a broad/admin scope

- Requirement source: `specification.md`'s scope item 2; open question 2
- Tasks: `VOC-051-T02`
- Tests: `VOC-051-TEST-02`
- Evidence: `VOC-051-EV-01`
- Result: pending

## VOC-051-AC-02 — A hourly scheduled workflow queries the Sentry API for new/unresolved issues across both the staging and production Sentry environments

- Requirement source: `specification.md`'s scope item 3
- Tasks: `VOC-051-T02`
- Tests: `VOC-051-TEST-03`
- Evidence: `VOC-051-EV-01`
- Result: pending

## VOC-051-AC-03 — A genuinely new Sentry problem results in exactly one unlabeled GitHub issue, following this repository's bug-report convention (reproduction detail, failing behavior, root cause if known)

- Requirement source: `specification.md`'s scope item 3; `AGENTS.md`'s
  "Reporting a bug found outside the normal loop"
- Tasks: `VOC-051-T02`
- Tests: `VOC-051-TEST-04`
- Evidence: `VOC-051-EV-01`
- Result: pending

## VOC-051-AC-04 — Repeated hourly runs against the same underlying Sentry issue never open a second GitHub issue for it

- Requirement source: `specification.md`'s scope item 4 (duplicate-check guard)
- Tasks: `VOC-051-T02`
- Tests: `VOC-051-TEST-05`
- Evidence: `VOC-051-EV-01`
- Result: pending

## VOC-051-AC-05 — The new scheduled workflow declares and uses no SSH credential (no `STAGING_SSH_*`/`PRODUCTION_SSH_*`-style secret, and no new equivalent) to either host

- Requirement source: `specification.md`'s scope item 5 (explicit non-goal)
- Tasks: `VOC-051-T02`
- Tests: `VOC-051-TEST-06`
- Evidence: `VOC-051-EV-01`
- Result: pending

## VOC-051-AC-06 — DOC-11 §1's "Error monitoring" row is updated in the same pull request that changes the monitoring mechanism it describes, per `AGENTS.md`'s documentation-consistency rule

- Requirement source: `AGENTS.md`'s "Any change to workflow behavior ... must
  update every doc that describes that behavior in the same pull request"
- Tasks: `VOC-051-T04`
- Tests: `VOC-051-TEST-07`
- Evidence: `VOC-051-EV-02`
- Result: pending

## VOC-051-AC-07 — Deterministic validation (`pnpm validate` for the `apps/web` change; `scripts/governance/validate-governance.sh` and `scripts/governance/classify-change-risk.sh` for the governance-doc change) passes against the exact implemented diff, and the real detected risk floor is recorded, not assumed from this draft's proposal

- Requirement source: `AGENTS.md`'s "Current validation"; `CLAUDE.md`'s required
  review
- Tasks: `VOC-051-T01`, `VOC-051-T02`, `VOC-051-T04`
- Tests: `VOC-051-TEST-08`
- Evidence: `VOC-051-EV-03`
- Result: pending

Acceptance criteria must be observable, stable, security-aware, and
bidirectionally traceable to requirements, tasks, tests, and evidence.
