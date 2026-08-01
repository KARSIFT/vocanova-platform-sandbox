# VOC-037 — Acceptance Criteria

## VOC-037-AC-00 — Production hosting/deploy-target decision is recorded and scoped

- Requirement source: `VOC-037-D00`
- Tasks: `VOC-037-T00`
- Tests: `VOC-037-TEST-00`
- Evidence: `VOC-037-EV-00`
- Result: pending
- Observable outcome: A decision record exists naming the chosen production
  hosting/deploy target (either "same shape as staging, second host" or a stated
  alternative), the specific files/workflows/infrastructure that must change as a
  consequence, and the founder's explicit approval of that record.

## VOC-037-AC-01 — Production secrets mechanism is decided and documented

- Requirement source: `VOC-037-D01`
- Tasks: `VOC-037-T01`
- Tests: `VOC-037-TEST-01`
- Evidence: `VOC-037-EV-01`
- Result: pending
- Observable outcome: A document states the production secret storage/injection/
  rotation mechanism, distinct from staging's, and confirms (by inspection of the
  chosen mechanism, not assertion alone) that no production secret is reachable
  from preview, staging, or CI.

## VOC-037-AC-02 — Privacy policy and terms of service are drafted and founder-reviewed

- Requirement source: `VOC-037-D02`
- Tasks: `VOC-037-T02`
- Tests: `VOC-037-TEST-02`
- Evidence: `VOC-037-EV-02`
- Result: pending
- Observable outcome: A privacy policy and terms-of-service document each exist,
  accurately describe the data this application actually collects/processes, and
  carry an explicit founder review/approval record before either is published.

## VOC-037-AC-06 — Production target is actually provisioned per T00/T01's accepted decisions

- Requirement source: `VOC-037-D00`, `VOC-037-D01`
- Tasks: `VOC-037-T06`
- Tests: `VOC-037-TEST-06`
- Evidence: `VOC-037-EV-06`
- Result: satisfied (2026-08-01) — see `VOC-037-EV-06` for the real deploy run, defects found/fixed live, and the on-host `INS-9`-`INS-11` rehearsal output (`PASS`)
- Observable outcome: `/opt/vocanova/production/` exists, fully separate
  from `/opt/vocanova/infra/` (staging); a `vocanova-production` Compose
  project runs with explicit per-service resource limits; a `production`
  GitHub Actions environment with founder-controlled required reviewers
  exists; `.github/workflows/deploy-production.yml` deploys to it without
  ever touching staging's directory tree, deploy user, or Compose project;
  and a negative-access rehearsal proves staging's deploy path cannot read
  production's secrets (T01's `INS-9` through `INS-11`).

## VOC-037-AC-03 — Launch kill switches and rollback work against the production target

- Requirement source: `VOC-037-D00`, `VOC-037-T06`'s provisioned target
- Tasks: `VOC-037-T03`
- Tests: `VOC-037-TEST-03`
- Evidence: `VOC-037-EV-03`
- Result: pending
- Observable outcome: Each of the four named kill switches, toggled against the
  production target, observably changes application behavior as documented; a
  rollback-by-redeploy rehearsal against the production target completes without
  data loss beyond an intentionally reverted change.

## VOC-037-AC-04 — Monitoring and alerting are active for production

- Requirement source: `VOC-037-D00`
- Tasks: `VOC-037-T04`
- Tests: `VOC-037-TEST-04`
- Evidence: `VOC-037-EV-04`
- Result: pending
- Observable outcome: A deliberately triggered test error is observed in Sentry
  for the production environment; a deliberate uptime-check failure produces a
  founder-reaching alert from Better Stack/UptimeRobot.

## VOC-037-AC-05 — R2 gate passes and founder records go/no-go

- Requirement source: DOC-12 §5
- Tasks: `VOC-037-T05`
- Tests: `VOC-037-TEST-05`
- Evidence: `VOC-037-EV-05`
- Result: pending
- Observable outcome: The R2 release PR shows all applicable checks passing,
  required review recorded as `approve` or an explicitly accepted follow-up, and a
  founder-authored go/no-go record exists (mirroring the R1 staging-acceptance
  record on issue #256), stating explicitly whether R2 is closed and what, if
  anything, remains as a tracked follow-up.
