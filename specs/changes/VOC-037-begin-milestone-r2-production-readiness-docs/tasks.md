# VOC-037 — Tasks

Each task below is independently implementable and reviewable in one pull request,
matching this repository's existing milestone-package convention (e.g. VOC-032's
`T00`–`T15`). Ordering reflects dependency, not priority. None is
implementation-authorized by this package; adoption and each task's own
implementation-authorization are separate.

## VOC-037-T00 — Production hosting/deploy-target decision record

- Requirement source: `VOC-037-D00` (not yet defined — founder decision required;
  see specification.md "Open question 1")
- Acceptance criteria: `VOC-037-AC-00`
- Tests: `VOC-037-TEST-00`
- Evidence: `VOC-037-EV-00`
- Status: pending
- Summary: Draft a decision record (options, cost/operational trade-offs, and a
  recommendation) for whether production reuses the existing single founder-owned-
  server Docker Compose shape or changes it, and scope exactly what would need to
  change (new host, new domain, new Cloudflare configuration, `docker-compose.yml`
  changes, `deploy-production.yml` workflow) once the founder decides. This task
  produces the decision record only; it does not itself provision any host, buy
  any domain, or write a deploy workflow — those become follow-up tasks scoped
  once the decision is recorded.

## VOC-037-T01 — Production credential/secrets management design

- Requirement source: `VOC-037-D01` (not yet defined — founder decision required;
  see specification.md "Open question 2")
- Acceptance criteria: `VOC-037-AC-01`
- Tests: `VOC-037-TEST-01`
- Evidence: `VOC-037-EV-01`
- Status: pending
- Depends on: `VOC-037-T00` (the secrets mechanism may differ depending on the
  chosen hosting target, e.g. a managed platform's own secret store versus a
  self-hosted host's file permissions)
- Summary: Evaluate and record a decision for how production secrets (database
  credentials, AI-provider keys, email-provider keys, Google OAuth client
  secret, Sentry DSN) are stored, injected, and rotated, distinct from staging's
  existing mechanism, and confirm no production secret becomes reachable from
  preview/staging/CI per DOC-11 §1's existing rule. Document the chosen mechanism
  and update `apps/api/.env.example`/`apps/web/.env.example` comments only if the
  variable *names* change — this task does not provision or commit any real
  secret value.

## VOC-037-T02 — Privacy policy and terms of service (draft for founder review)

- Requirement source: `VOC-037-D02` (not yet defined — founder decision required;
  see specification.md "Open question 3"; proposed risk `R4`)
- Acceptance criteria: `VOC-037-AC-02`
- Tests: `VOC-037-TEST-02`
- Evidence: `VOC-037-EV-02`
- Status: pending
- Summary: Draft a privacy policy and terms of service covering the data this
  application actually collects and processes (account/email, saved words, review
  history, submitted sentences and AI feedback, OAuth profile data, standard
  operational/analytics logging), for explicit founder review, revision, and
  publication. This task does not publish a policy the founder has not reviewed,
  and does not itself constitute the founder's required R4 legal/privacy sign-off
  — it produces the draft that sign-off is recorded against.

## VOC-037-T06 — Production infrastructure provisioning

- Requirement source: `VOC-037-D00` (accepted 2026-08-01, Option A-modified)
  and `VOC-037-D01` (accepted 2026-08-01, corrected 4A mechanism) — see
  `t00-production-hosting-decision-record.md` and
  `t01-production-secrets-decision-record.md`
- Acceptance criteria: `VOC-037-AC-06` (new; see acceptance-criteria.md)
- Tests: `VOC-037-TEST-06`
- Evidence: `VOC-037-EV-06`
- Status: pending
- Depends on: `VOC-037-T00`, `VOC-037-T01` (both accepted; this task executes
  their already-decided design, it does not make new decisions)
- Summary: Build the actual production target per T00/T01's accepted
  decisions: a `/opt/vocanova/production/` directory tree fully separate
  from staging's `/opt/vocanova/infra/`, its own `vocanova-production`
  Docker Compose project with explicit per-service resource limits (shared-
  host contention mitigation), a production-only least-privilege deploy
  user, a `production` GitHub Actions environment with founder-controlled
  required reviewers, and a new `.github/workflows/deploy-production.yml`
  workflow that writes only under the production tree and never touches
  staging's. Does not choose a production domain/DNS name itself — uses a
  founder-confirmable placeholder and flags exact hostnames as a T05/founder
  confirmation item, consistent with T00's "final names to be founder-
  confirmed during implementation" note. Executes `VOC-037-T01`'s `INS-9`
  through `INS-11` negative-access rehearsal (staging's deploy path/user
  cannot read production's secrets) and records `VOC-037-EV-01` alongside
  its own `EV-06`.

## VOC-037-T03 — Launch kill-switch and rollback verification (production target)

- Requirement source: `VOC-037-D00` (inherits the hosting decision from `T00`)
- Acceptance criteria: `VOC-037-AC-03`
- Tests: `VOC-037-TEST-03`
- Evidence: `VOC-037-EV-03`
- Status: implemented; live rehearsal outstanding. The rehearsal script,
  its deterministic harness, the rollback deploy mode this repository
  did not previously have, and the kill-switch deploy inputs are
  delivered and verified in repository; `VOC-037-AC-03` stays open until
  the founder-owned live run recorded in
  `t03-killswitch-rollback-evidence.md` is executed.
- Depends on: `VOC-037-T06` (needs the actual production target `T06` builds,
  not just T00's decision, to verify against)
- Summary: Verify, against whatever production target `T00` decides, that the
  four existing kill switches (`AI_FEATURES_ENABLED`, `EMAIL_MAGIC_LINK_ENABLED`,
  `GOOGLE_OAUTH_ENABLED`, `NEW_USER_SIGNUP_ENABLED`) and the roll-forward/redeploy-
  by-digest rollback path (DOC-11 §3) work end to end in production, mirroring
  VOC-032's `EV-21` staging rehearsal but against the production target. Does not
  re-implement the switches if `T00` reuses the existing mechanism unchanged.

## VOC-037-T04 — Monitoring/alerting readiness (production target)

- Requirement source: `VOC-037-D00` (inherits the hosting decision from `T00`)
- Acceptance criteria: `VOC-037-AC-04`
- Tests: `VOC-037-TEST-04`
- Evidence: `VOC-037-EV-04`
- Status: pending
- Depends on: `VOC-037-T06`
- Summary: Configure and verify Sentry error monitoring and Better Stack/
  UptimeRobot uptime monitoring for the production target, with alerts reaching
  the founder, per DOC-11 §1's named tools and §5's "Production-ready" checklist
  requirement that both be "active."

## VOC-037-T05 — R2 release PR and founder go/no-go record

- Requirement source: DOC-12 §5 (R2 gate)
- Acceptance criteria: `VOC-037-AC-05`
- Tests: `VOC-037-TEST-05`
- Evidence: `VOC-037-EV-05`
- Status: pending
- Depends on: `VOC-037-T00` through `VOC-037-T04`, and `VOC-037-T06`
- Summary: Open the R2 release PR, confirm every applicable check passes and
  required review returns `approve` or an explicitly accepted follow-up
  (mirroring the R1-closure pattern on issue #256), and record the founder's
  explicit go/no-go decision per DOC-12 §5's exact gate language. This is the R4
  founder decision this whole package leads to; it is not self-executed by any
  prior task.
