# VOC-096 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **secret bootstrap confirmation → T00 → T01 → T02**.

## VOC-096-T00 — Persist production allowlist secret and remove ephemeral dispatch input

- Requirement source: `VOC-096-D00`, `VOC-096-D01`, `VOC-096-D02`, `VOC-096-D03`, `VOC-096-D04`
- Acceptance criteria: `VOC-096-AC-00`, `VOC-096-AC-01`, `VOC-096-AC-02`, `VOC-096-AC-03`, `VOC-096-AC-07`, `VOC-096-AC-08`
- Tests: `VOC-096-TEST-00` through `VOC-096-TEST-05`, `VOC-096-TEST-10`, `VOC-096-TEST-11`
- Evidence: `VOC-096-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

0. Before this task PR merges, confirm the repository secret
   `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` exists and is populated. Record only
   confirmation that it exists, never its value. Issue #809 states the secret already
   exists; re-verify at implementation time.
1. Create `infra/scripts/validate-production-signup-allowlist.sh` mirroring
   `validate-staging-signup-allowlist.sh` with production-prefixed env vars
   (`PRODUCTION_GOOGLE_OAUTH_ENABLED`, `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`).
2. Modify `deploy-production.yml`:
   - Remove the `new_user_signup_allowlist` dispatch input entirely.
   - Add a runner-side step **Validate production controlled-signup allowlist** before
     **Write production application configuration**, invoking the validator with
     `secrets.PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` and OAuth-enabled derivation from
     the Google credential pair (same expression as today).
   - Change the config step env to
     `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST: ${{ secrets.PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST }}`.
   - Add SSH-side fail-closed guard and boolean-only readiness logging (mirror staging).
3. Keep `NEW_USER_SIGNUP_ENABLED=false` unchanged.
4. Update the deploy-production.yml header comment to document the persistent secret
   and removal of the dispatch input.
5. Add `scripts/foundation/voc096-deploy-production-allowlist.test.mjs` covering:
   - Secret precedence on every deploy path.
   - Dispatch input removal.
   - Fail-closed: OAuth enabled + empty/malformed/multiline cohort → non-zero exit.
   - Redaction: no email values in validator stdout/stderr.
   - Isolation: production workflow never references `STAGING_NEW_USER_SIGNUP_ALLOWLIST`;
     staging workflow never references `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`.
6. Extend `voc088-deploy-staging-allowlist.test.mjs` isolation test or add equivalent
   assertion in the new voc096 test file for bidirectional isolation.

### Explicitly out of scope for this task

- Production OAuth harness / synthetic extension (T01).
- Operator docs or live production evidence (T02).
- Staging workflow changes.
- Application code or database changes.

## VOC-096-T01 — Extend production OAuth/readiness synthetic for controlled signup readiness

- Requirement source: `VOC-096-D06`, `VOC-096-D07`
- Acceptance criteria: `VOC-096-AC-04`, `VOC-096-AC-05`, `VOC-096-AC-06`, `VOC-096-AC-08`
- Tests: `VOC-096-TEST-06` through `VOC-096-TEST-09`
- Evidence: `VOC-096-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-096-T00`

### Required work

1. Extend `infra/scripts/verify-production-oauth-start.sh` to mirror
   `verify-staging-oauth-start.sh`: when `EXPECT_OAUTH_ENABLED=true`, fetch `/healthz`
   and assert `controlled_signup_ready` is `true` without email-like strings in the
   response body.
2. Update `infra/monitoring/synthetics.yaml` for
   `synthetic.production.oauth-expected-state` coverage to include
   `api:GET /healthz controlled_signup_ready` and
   `feature:production-controlled-signup-readiness`.
3. Update `.github/workflows/scheduled-synthetics.yml` header comment for the
   production OAuth job if needed so it documents the readiness assertion (no behavioral
   change required if the harness alone gains the check).
4. Add `scripts/foundation/voc096-production-readiness.test.mjs` covering:
   - Harness contains controlled_signup_ready assertion when OAuth expected enabled.
   - Synthetic inventory declares the new coverage strings.
   - Harness still asserts canonical callback and accounts.google.com target.
   - No email metadata patterns in harness success output fixtures.
5. Run applicable foundation tests including any scheduled-synthetics inventory validators.

### Explicitly out of scope for this task

- deploy-production.yml allowlist persistence (T00).
- Real Google login automation or VOC-092 harness changes.
- `/healthz` handler changes (already implemented).
- Operator docs or live production proof (T02).

## VOC-096-T02 — Document operator procedure and production deploy-and-verify evidence

- Requirement source: issue #809; `VOC-096-D02`, `VOC-096-D04`
- Acceptance criteria: `VOC-096-AC-09`, `VOC-096-AC-10`, `VOC-096-AC-03`, `VOC-096-AC-05`
- Tests: `VOC-096-TEST-12`, `VOC-096-TEST-13`
- Evidence: `VOC-096-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-096-T01`

### Required work

1. Add `docs/operations/production-controlled-signup.md` mirroring
   `docs/operations/staging-controlled-signup.md` for production hosts, secret name,
   deploy workflow, validation steps, and fail-closed behavior. Cross-link staging doc;
   never duplicate cohort values.
2. Add or extend a foundation doc test (e.g. extend `voc092-operator-docs.test.mjs`
   or add `voc096-operator-docs.test.mjs`) asserting the production operator doc exists
   and references the correct secret/workflow names without example real emails.
3. After T01 merges, use a separately reviewed controlled activation promotion from
   `develop` to `main`, then monitor the automatic production deployment. This
   activation is required before T02 may claim live evidence; it does not bypass the
   normal exact-SHA review, merge, or deployment gates.
4. After the controlled activation is green, record deploy-and-verify evidence in
   `t02-evidence.md`:
   - Successful production deploy run URL (scrubbed).
   - Live `/healthz` jq output showing `controlled_signup_ready: true` without cohort
     metadata.
   - `EXPECT_OAUTH_ENABLED=true bash infra/scripts/verify-production-oauth-start.sh` pass.
   - Scheduled `synthetic.production.oauth-expected-state` green.
   - Production route/content smoke references as applicable (no new interactive OAuth).
5. All evidence must be scrubbed of email addresses, credentials, session material, and
   personal data.

### Explicitly out of scope for this task

- Code changes to workflows or harnesses (T00/T01 own those).
- Staging operator procedure changes beyond cross-links.

## Task ordering notes

- Secret bootstrap confirmation blocks T00 merge and its automatic production deploy.
- T00 blocks the allowlist persistence required for live `controlled_signup_ready: true`.
- T01 blocks the synthetic readiness assertion and harness checks used in T02 evidence.
- A separately reviewed controlled activation promotion after T01 supplies the live
  production state required by T02 without waiting for roster completion.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
