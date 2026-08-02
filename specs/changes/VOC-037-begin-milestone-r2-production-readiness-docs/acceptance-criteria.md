# VOC-037 — Acceptance Criteria

## VOC-037-AC-00 — Production hosting/deploy-target decision is recorded and scoped

- Requirement source: `VOC-037-D00`
- Tasks: `VOC-037-T00`
- Tests: `VOC-037-TEST-00`
- Evidence: `VOC-037-EV-00`
- Result: **SATISFIED** — `VOC-037-D00` is `status: accepted`; `t00-production-hosting-decision-record.md` names the approved option (Option A-modified, same-host co-location, logically isolated) and the files/workflows that change as a consequence, with explicit founder approval recorded (including two live-corrected supersession notes as real conditions were discovered). Bookkeeping note: this line was previously left `pending` after the decision was actually approved; corrected here, no new work.
- Observable outcome: A decision record exists naming the chosen production
  hosting/deploy target (either "same shape as staging, second host" or a stated
  alternative), the specific files/workflows/infrastructure that must change as a
  consequence, and the founder's explicit approval of that record.

## VOC-037-AC-01 — Production secrets mechanism is decided and documented

- Requirement source: `VOC-037-D01`
- Tasks: `VOC-037-T01`
- Tests: `VOC-037-TEST-01`
- Evidence: `VOC-037-EV-01`
- Result: **SATISFIED, with the same disclosed residual risk as AC-06** — `VOC-037-D01` is `status: accepted`; `t01-production-secrets-decision-record.md` states the 4A mechanism (separate directory tree/deploy user/Compose project/GitHub environment), distinct from staging's, and confirms by live inspection (not assertion) that production secrets are unreachable from preview and CI, and unreachable from staging's directory-permission path (`INS-9` passes). The one gap is the same one recorded under AC-06: staging's real deploy identity (`ubuntu`) has independent, pre-existing OS-level sudo that permission-based isolation cannot stop — founder-waived, not eliminated. Bookkeeping note: this line was previously left `pending` after the decision was actually approved; corrected here.
- Observable outcome: A document states the production secret storage/injection/
  rotation mechanism, distinct from staging's, and confirms (by inspection of the
  chosen mechanism, not assertion alone) that no production secret is reachable
  from preview, staging, or CI.

## VOC-037-AC-02 — Privacy policy and terms of service are drafted and founder-reviewed

- Requirement source: `VOC-037-D02`
- Tasks: `VOC-037-T02`
- Tests: `VOC-037-TEST-02`
- Evidence: `VOC-037-EV-02`
- Result: **SATISFIED (2026-08-02)** — both documents carry an explicit founder review/approval record (data-collection accuracy confirmed, minimum age set to 13, contact email decided). One item — governing law/jurisdiction — is genuinely undecided pending VocaNova's incorporation status and is recorded as blocking *publication* specifically, not blocking this review criterion: the AC requires a review/approval record to exist before publication, and it now does, with the remaining gap disclosed rather than hidden. Publication itself should not proceed until governing law resolves.
- Observable outcome: A privacy policy and terms-of-service document each exist,
  accurately describe the data this application actually collects/processes, and
  carry an explicit founder review/approval record before either is published.

## VOC-037-AC-06 — Production target is actually provisioned per T00/T01's accepted decisions

- Requirement source: `VOC-037-D00`, `VOC-037-D01`
- Tasks: `VOC-037-T06`
- Tests: `VOC-037-TEST-06`
- Evidence: `VOC-037-EV-06`
- Result: **NOT satisfied — accepted as a founder-waived residual risk (2026-08-02)** — an earlier "satisfied" result was based on running the rehearsal manually as root, which doesn't reflect the real deploy's non-root identity. Corrected: `INS-9`/`INS-10` pass; `INS-11` correctly FAILS with a confirmed, disclosed finding (`ubuntu`, staging's real deploy identity — an earlier revision of this record incorrectly named the unused `deploy` account — has independent blanket sudo via group membership, so directory-based isolation cannot be proven against it). The founder was offered (1) migrating staging's deploy identity to a new least-privilege account, or (2) waiving the risk, and chose the waiver rather than the riskier account-migration surgery. `deploy-production.yml` now treats exactly this disclosed finding as a non-blocking warning (any other rehearsal failure still hard-fails the deploy). See `VOC-037-EV-06`'s "Confirmed residual risk" section for the full record.
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
- Result: **NOT satisfied** (2026-08-01), recorded honestly as pending, not as a partial pass — see `VOC-037-EV-03` for exactly what was and wasn't verified. `AI_FEATURES_ENABLED` and `GOOGLE_OAUTH_ENABLED`/`EMAIL_MAGIC_LINK_ENABLED` were verified to change *some* observable signal (startup log / HTTP status respectively) when toggled, but `AI_FEATURES_ENABLED` was never verified at the documented HTTP surface and `NEW_USER_SIGNUP_ENABLED` was never toggled `true` or verified at all. The redeploy rehearsal exercised the recreate/health-check half of the mechanism but not `pull` (registry auth was session-specific, not re-demonstrated) and not a true two-different-versions rollback (only one production artifact exists so far). Closing AC-03 requires either completing these gaps once real credentials and a second production artifact exist, or an explicit founder-accepted waiver — this record does neither on its own.
- Observable outcome: Each of the four named kill switches, toggled against the
  production target, observably changes application behavior as documented; a
  rollback-by-redeploy rehearsal against the production target completes without
  data loss beyond an intentionally reverted change.

## VOC-037-AC-04 — Monitoring and alerting are active for production

- Requirement source: `VOC-037-D00`
- Tasks: `VOC-037-T04`
- Tests: `VOC-037-TEST-04`
- Evidence: `VOC-037-EV-04`
- Result: **SATISFIED (2026-08-02)** — both halves verified live against real production infrastructure: a real Sentry test event (event ID `60c282e455a843ff9151a235ebb71dda`) and a real, rehearsed uptime-down alert (founder-confirmed received via Telegram). Uptime monitoring uses self-hosted Uptime Kuma rather than the literally-named Better Stack/UptimeRobot — UptimeRobot's free tier defaults to `HEAD` probes, which the production API's `GET`-only `/healthz` rejects with `405`, and the founder chose the self-hosted alternative over a paid plan once informed live. See `VOC-037-EV-04`'s "Disclosed deviation" note.
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
