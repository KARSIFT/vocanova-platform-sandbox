# VOC-096 — Persist controlled production Google signup cohort: Specification

## Objective and requirement source

Remediate the production controlled-signup persistence gap recorded in
[GitHub issue #809](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/809):
production Google OAuth is live, blanket signup remains disabled, but automatic
production deployments erase the controlled cohort because
`deploy-production.yml` reads `NEW_USER_SIGNUP_ALLOWLIST` only from an ephemeral
`workflow_dispatch` input. The repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`
must become the persistent source of truth.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time repo read:

| Item | Current state |
|------|---------------|
| Production OAuth | Enabled when `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` repository secrets are present |
| `NEW_USER_SIGNUP_ENABLED` | hardcoded `false` in deploy-production.yml's config step |
| `NEW_USER_SIGNUP_ALLOWLIST` source | `${{ inputs.new_user_signup_allowlist }}` — workflow_dispatch input only; push runs resolve to empty |
| Persistent secret | `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` exists but is not referenced by deploy-production.yml |
| `/healthz` | `controlled_signup_ready` boolean already implemented (VOC-088-T01); currently `false` on live production |
| Production synthetic | `synthetic.production.oauth-expected-state` uses `verify-production-oauth-start.sh`, which asserts OAuth start/callback only — not `controlled_signup_ready` |
| Staging precedent | VOC-088-T00/T01 solved the same persistence and readiness gap for staging |

This is distinct from VOC-088's failure-to-issue agent (already merged) and from
VOC-092's controlled Google signup callback E2E harness (provider policy remains
covered there; this package does not automate real Google login in CI).

## Scope and non-goals

In scope:

1. Modify `deploy-production.yml` to read `NEW_USER_SIGNUP_ALLOWLIST` from the
   persistent repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` on every
   deploy (push-triggered and manual dispatch).
2. Remove the `new_user_signup_allowlist` dispatch input entirely.
3. Add `infra/scripts/validate-production-signup-allowlist.sh` mirroring the staging
   validator: fail closed when OAuth is enabled and the cohort is missing, empty,
   multiline, or malformed; log only `controlled signup ready: true/false`.
4. Add a runner-side validation step before the SSH config write, plus a redundant
   fail-closed guard in the SSH script (same pattern as VOC-088-T00).
5. Extend `verify-production-oauth-start.sh` to assert `/healthz.controlled_signup_ready
   is true` when `EXPECT_OAUTH_ENABLED=true`, mirroring
   `verify-staging-oauth-start.sh`.
6. Update `infra/monitoring/synthetics.yaml` coverage metadata for
   `synthetic.production.oauth-expected-state`.
7. Add deterministic foundation tests (`scripts/foundation/voc096-*.test.mjs`) for
   secret precedence, dispatch input removal, fail-closed validation, redaction,
   staging/production secret isolation, and production readiness harness behavior.
8. Add `docs/operations/production-controlled-signup.md` operator procedure.
9. Record production deploy-and-verify evidence (T02): live `/healthz`, OAuth start
   harness, scheduled synthetic green, route/content smoke — without secrets or
   personal data.

Non-goals / explicitly excluded:

- Enabling blanket production signup (`NEW_USER_SIGNUP_ENABLED=true`).
- Changing staging signup policy, staging secrets, or `deploy-staging.yml`.
- Modifying the `/healthz` handler implementation (already present from VOC-088-T01).
- Adding or changing the operational-failure observer (VOC-088-T02).
- Real Google interactive login in CI or changes to VOC-092's callback E2E harness
  beyond any shared documentation cross-links.
- Database migrations or application signup logic changes beyond deploy-time config sync.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4** (production deploy signup-policy synchronization
  while Google OAuth is live).
- **Measured path floor at drafting:** **R3** for `.github/workflows/` and
  `infra/scripts/`. Semantic escalation to R4 is proposed because the change restores
  production controlled-launch cohort persistence.
- Protected areas: `.github/workflows/deploy-production.yml`, production repository
  secrets, `infra/scripts/verify-production-oauth-start.sh`,
  `infra/monitoring/synthetics.yaml`, production `api.env` on the host.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-096-D00`: Automatic push deployments erase the production controlled cohort
because `deploy-production.yml` resolves `inputs.new_user_signup_allowlist` to empty
on push events. Remediation targets secret-backed persistence, not OAuth credential
sync or application callback logic.

`VOC-096-D01`: The repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` is the
single source of truth for production `NEW_USER_SIGNUP_ALLOWLIST`. Only
`deploy-production.yml` may consume it; `deploy-staging.yml` must never reference it.

`VOC-096-D02`: Remove the `new_user_signup_allowlist` dispatch input entirely (same
choice as VOC-088-T00 for staging). Cohort changes are made by editing the GitHub
secret, then letting the next production deploy (automatic on main push or manual
dispatch) pick up the value.

`VOC-096-D03`: Validation mirrors staging: when `GOOGLE_OAUTH_ENABLED=true`, missing,
empty, multiline, or malformed allowlist values fail before runtime configuration is
written. When OAuth is disabled, an empty cohort is acceptable and readiness logs
`controlled signup ready: false`.

`VOC-096-D04`: Logs, workflow output, `/healthz`, synthetics, tests, issues, PRs,
and evidence expose only the boolean readiness fact — never cohort values, email
addresses, or cardinality.

`VOC-096-D05`: `NEW_USER_SIGNUP_ENABLED` remains `false` in production. Only the
controlled allowlist may admit first-time production accounts while global signup stays
off.

`VOC-096-D06`: Extend `verify-production-oauth-start.sh` and
`synthetic.production.oauth-expected-state` to assert `controlled_signup_ready=true`
when OAuth is expected enabled, matching the staging harness behavior from VOC-088-T01.

`VOC-096-D07`: Do not automate real Google interactive login in CI. Provider callback
policy remains covered by the repository-managed synthetic E2E harness (VOC-092).

## Open questions for the reviewing human

1. Confirm proposed **R4**, or record in writing if the adopting human treats this
   as routine **R3** given the staging precedent (VOC-088).
2. Confirm secret bootstrap timing: issue #809 states
   `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` already exists — T00 should record
   confirmation only, not the value, before merge triggers the next production deploy.
3. If production must intentionally run OAuth-enabled with an empty cohort (unlikely
   given fail-closed AC), that requires an explicit product decision outside this
   package; default remains fail-closed per issue #809 AC-4.

## Data, migrations, analytics, and accessibility

- No application schema migration.
- No database mutation — only deploy-time `api.env` synchronization.
- No product UI change — evidence-backed non-applicability.
- No analytics change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.
