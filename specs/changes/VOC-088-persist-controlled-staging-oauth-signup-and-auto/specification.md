# VOC-088 — Persist controlled staging OAuth signup and auto-file operational failures: Specification

## Objective and requirement source

Close the operational gaps reported in
[GitHub issue #746](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/746):
the controlled staging signup mechanism exists in application code
(`killswitches.go` `SignupAllowlist`) and in `deploy-staging.yml`'s "Write staging
application configuration" step, but its deployment control is ephemeral rather
than persistently secret-backed, its ready/not-ready state is absent from the
operational synthetic contract, and workflow failures are not connected to the
issue-to-plan remediation loop.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary context (issue #746 + drafting-time repo read 2026-08-19):

| Item | Value |
|------|-------|
| Staging OAuth pair | `GOOGLE_OAUTH_CLIENT_ID` / `GOOGLE_OAUTH_CLIENT_SECRET` present as repository secrets |
| `GOOGLE_OAUTH_ENABLED` | derived `true` in deploy-staging.yml when both secrets present |
| `NEW_USER_SIGNUP_ENABLED` | hardcoded `false` in deploy-staging.yml's config step |
| `NEW_USER_SIGNUP_ALLOWLIST` source | `${{ inputs.new_user_signup_allowlist || '' }}` — workflow_dispatch input only |
| Automatic push deploy behavior | no dispatch inputs → allowlist written as empty string → cohort erased |
| Persistent GitHub secret | none exists for the allowlist |
| Staging OAuth synthetic | `synthetic.staging.oauth-expected-state` asserts 200 + accounts.google.com; does not assert controlled signup readiness |
| Failed workflow → issue | not implemented; error-monitoring.yml reads Sentry only |
| 503 on first-time signup | expected behavior of `ErrSignupsDisabled` in `killswitches.go` when `signupAllowed` returns false |

## Scope and non-goals

In scope:

1. Add a persistent GitHub secret (staging-scoped) for `NEW_USER_SIGNUP_ALLOWLIST`
   that `deploy-staging.yml` reads and syncs to `api.env` on every deploy —
   automatic push deploys and manual dispatches alike — without printing email
   addresses in logs or workflow output.
2. Constrain or remove the ephemeral `workflow_dispatch` input
   `new_user_signup_allowlist` so it cannot silently override or erase the
   persistent secret. The secret is the single source of truth for the controlled
   cohort.
3. Fail closed on malformed allowlist configuration (e.g. secret present but empty
   when controlled signup is expected). Expose only a non-sensitive readiness
   boolean in the deploy log, never a cohort size or email values.
4. Extend `synthetic.staging.oauth-expected-state` (or add a sibling synthetic) to
   assert that when `GOOGLE_OAUTH_ENABLED=true` and `NEW_USER_SIGNUP_ENABLED=false`,
   the controlled cohort is non-empty (i.e. controlled first-time signup is
   operational). A pass requires the API's `/healthz` (or a new non-sensitive
   readiness endpoint) to expose a boolean `controlled_signup_ready` fact.
5. Add a repository-managed, deduplicated GitHub issue path for failed
   `scheduled-synthetics.yml`, `deploy-staging.yml`, and `deploy-production.yml`
   runs. Use a standalone `workflow_run` observer that mints an installation token
   from the existing automation GitHub App credentials; the default `GITHUB_TOKEN`
   cannot trigger `plan-from-issue`. Never include secrets, session values, OAuth
   state, raw user identifiers, or unfiltered container logs in the issue body.
6. Deterministic tests for: secret precedence over dispatch input, redaction of
   email addresses, readiness synthetic failure/success, issue deduplication,
   App-token identity, and both staging/production environment mappings.
7. Document the secure operator procedure for adding/removing a staging cohort
   member (GitHub secret update → next deploy picks it up) and prove with evidence
   that a later normal push deploy preserves the cohort.
8. Deploy-and-verify evidence: a permitted first-time staging Google user can sign
   in; an unlisted user remains blocked; scheduled checks pass; a safe controlled
   failure fixture opens exactly one sanitized issue and a repeat creates no
   duplicate.

Non-goals / explicitly excluded:

- Any production signup-policy change. Production `NEW_USER_SIGNUP_ENABLED` and
  allowlist behavior is unchanged.
- Turning expected signup denials into Sentry exceptions. The 503 on unauthorized
  first-time signup is correct behavior; Sentry is for unexpected application
  errors.
- Raw-log scraping that could transfer personal data into GitHub.
- Manual server configuration as the permanent fix.
- Application schema migrations.
- Changes to `error-monitoring.yml`'s Sentry monitoring (it remains Sentry-only).
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Builder/issue proposal:** none stated in issue #746 beyond constraints.
- **Measured path floor:** **R3** for `.github/workflows/*`, `infra/scripts/*`,
  `infra/monitoring/*`. Not R4 unless a task touches `scripts/governance/*`.
- **Draft package proposal:** **R3** (highest of builder class and measured floor).
  This is a draft proposal for the reviewing human at adoption time, never a
  determination.
- Protected areas: `.github/workflows/deploy-staging.yml`,
  `.github/workflows/deploy-production.yml`,
  `.github/workflows/operational-failure-monitoring.yml`, staging secrets,
  `apps/api/app/api/production.go`, `apps/api/app/api/production_test.go`, and
  `infra/monitoring/synthetics.yaml`.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-088-D00`: The controlled staging signup allowlist must be persistently backed
by the repository secret `STAGING_NEW_USER_SIGNUP_ALLOWLIST` per
`VOC-088-DEP-00`. `deploy-staging.yml` reads this secret and syncs it to
`api.env` on every deploy. The `workflow_dispatch` input
`new_user_signup_allowlist` must either be removed entirely or made additive-only
(union with the secret), never a replacement. The secret is the single source of
truth.

`VOC-088-D01`: `NEW_USER_SIGNUP_ENABLED` remains `false` in staging. Only
explicitly admitted normalized emails in the persistent allowlist may create
staging accounts. No blanket-enable path.

`VOC-088-D02`: Deploy-staging.yml must fail closed when the allowlist secret is
missing or empty and `GOOGLE_OAUTH_ENABLED` is true, because that state means
controlled first-time signup is expected but no one can use it. The deploy log
must expose only a boolean, never a cohort size or email values.

`VOC-088-D03`: The staging OAuth/readiness synthetic must assert controlled
signup readiness: when OAuth is enabled, global signup is disabled, and controlled
signup is an expected capability, the synthetic fails if the controlled cohort is
empty. This requires the API's `/healthz` or a dedicated non-sensitive endpoint to
expose `controlled_signup_ready` as a boolean. The synthetic asserts that it is
`true`; neither email values nor cohort cardinality are exposed.

`VOC-088-D04`: Failed scheduled synthetics and deploy workflows open exactly one
deduplicated, sanitized, plain unlabeled GitHub issue through a standalone
`workflow_run` observer. The observer mints an installation token using
`KARSIFT_BOT_APP_ID` and `KARSIFT_BOT_PRIVATE_KEY`; the default `GITHUB_TOKEN`
must not be used to create the issue because its events do not trigger downstream
workflows. The issue body contains the workflow name, conclusion, and run URL plus
a sanitized fixed summary. It never contains secrets, session values, OAuth state,
raw user identifiers, unfiltered logs, or failure-step output. Deduplication: if
an open issue with the same stable workflow + conclusion fingerprint already
exists, no new issue is created.

`VOC-088-D05`: Responsibility separation is preserved:
- Sentry (error-monitoring.yml): unexpected application errors
- Kuma (sync-monitoring.yml): availability/TLS/basic API health
- Synthetics (scheduled-synthetics.yml): expected behavior/configuration readiness
- Failure-to-issue agent: governed remediation for failed operational workflows

`VOC-088-D06`: The failure-to-issue agent covers `scheduled-synthetics.yml`,
`deploy-staging.yml`, and `deploy-production.yml`, including failure, cancellation,
and timeout conclusions. It does not cover
`error-monitoring.yml` (which already opens issues from Sentry) or
`sync-monitoring.yml` (which is a controlled manual dispatch, not scheduled).

`VOC-088-D07`: No email addresses or OAuth credentials in git, issues, workflow
logs, or test fixtures that resemble real users. Test fixtures use
`@synthetic.vocanova.invalid` addresses only.

## Resolved design choices

1. **Secret scope** (`VOC-088-DEP-00`): use repository secret
   `STAGING_NEW_USER_SIGNUP_ALLOWLIST`; only deploy-staging.yml may consume it.
2. **Dispatch input behavior** (`VOC-088-D00`): remove the input; the persistent
   secret is the sole source of truth.
3. **Readiness endpoint shape** (`VOC-088-D03`): extend `/healthz` with the
   informational boolean `controlled_signup_ready`; it does not change HTTP
   health semantics.
4. **Failure observer** (`VOC-088-DEP-01`): use a standalone `workflow_run`
   observer and the existing automation GitHub App installation token.
5. **Risk class:** measured path floor is R3. No R4 path is touched. Semantic
   risk (staging secrets, operational workflow changes) is R3-proportionate.

## Security and privacy

- Email addresses only in GitHub secrets; never committed, never printed in logs,
  issues, workflow output, or evidence.
- The readiness endpoint exposes only a boolean, never cohort size or email values.
- Failure-to-issue bodies are sanitized: no secrets, session values, OAuth state,
  raw user identifiers, or unfiltered container logs.
- No production signup-policy change.
- Test fixtures use `@synthetic.vocanova.invalid` addresses only.
- Preserve staging/production database, secret, directory, network, and deploy-user
  isolation.

## Data, migrations, analytics, and accessibility

- **Data/migrations:** None. No application schema migration. No database mutation.
  The only data change is a GitHub secret (staging allowlist) and workflow
  configuration.
- **Analytics:** None — evidence-backed non-applicability.
- **Accessibility:** No product UI change. Operator docs and workflow changes only.
