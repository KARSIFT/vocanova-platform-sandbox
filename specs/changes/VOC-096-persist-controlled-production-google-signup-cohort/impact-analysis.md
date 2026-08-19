# VOC-096 — Impact Analysis

## Security and privacy

- **Secrets:** `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` is personal data. The validator,
  deploy logs, `/healthz`, synthetics, tests, issues, PRs, and evidence must never print
  cohort values or cardinality. Only boolean readiness may appear publicly.
- **Tier isolation:** Production secret consumption is limited to `deploy-production.yml`.
  Staging secret and staging deploy path remain unchanged. No cross-tier database or SSH
  user changes.
- **Fail-closed posture:** Prevents OAuth-enabled production from running with an empty
  controlled cohort — reducing accidental total signup lockout for intended L1 users while
  keeping blanket signup disabled.
- **Authorization:** Cohort membership changes remain operator-controlled via GitHub secret
  edit; no new public API to mutate the allowlist.

## Data and migrations

- No database schema or migration.
- Deploy-time `api.env` synchronization only; rollback restores prior workflow behavior
  while the GitHub secret remains available for recovery.
- Existing SignupAllowlist application logic (VOC-038/VOC-084) unchanged.

## Analytics and accessibility

- Evidence-backed non-applicability. No product UI or analytics instrumentation changes.

## Risks, dependencies, and evidence

- `VOC-096-R00`: **Automatic production deploy erases cohort today** — every main push
  deploy converges `NEW_USER_SIGNUP_ALLOWLIST` to empty. Mitigation: AC-00/AC-01/TEST-00/01.
- `VOC-096-R01`: **Empty secret at T00 merge fails production deploy** when OAuth is
  enabled. Mitigation: pre-merge secret confirmation (DEP-00); intentional fail-closed
  behavior documented in operator guide.
- `VOC-096-R02`: **Production synthetic begins failing until cohort persisted** — expected
  during rollout if live cohort was empty; resolves once T00 deploy succeeds with populated
  secret. Mitigation: ordered tasks and T02 evidence.
- `VOC-096-DEP-00`: Existing `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` repository secret;
  confirm populated before T00 merge.
- `VOC-096-DEP-01`: VOC-088-T01 `/healthz` field and Go tests — reuse, do not rewrite.
- `VOC-096-DEP-02`: VOC-088 staging persistence pattern — mirror for production.
- `VOC-096-DEP-03`: VOC-086 production OAuth synthetic infrastructure.
- `VOC-096-EV-00`: T00 workflow/validator/tests evidence (`t00-evidence.md`).
- `VOC-096-EV-01`: T01 harness/synthetic/tests evidence (`t01-evidence.md`).
- `VOC-096-EV-02`: T02 operator doc and live production verification (`t02-evidence.md`).
