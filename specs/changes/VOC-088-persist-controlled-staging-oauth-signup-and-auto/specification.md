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
   count (number of allowlisted entries) or boolean in the deploy log, never email
   values.
4. Extend `synthetic.staging.oauth-expected-state` (or add a sibling synthetic) to
   assert that when `GOOGLE_OAUTH_ENABLED=true` and `NEW_USER_SIGNUP_ENABLED=false`,
   the controlled cohort is non-empty (i.e. controlled first-time signup is
   operational). A pass requires the API's `/healthz` (or a new non-sensitive
   readiness endpoint) to expose a `signup_allowlist_count > 0` fact.
5. Add a repository-managed, deduplicated GitHub issue path for failed
   `scheduled-synthetics.yml`, `deploy-staging.yml`, and `deploy-production.yml`
   runs. Use the automation App identity (`GITHUB_TOKEN`) so plain unlabeled
   issues trigger `plan-from-issue`. Never include secrets, session values, OAuth
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
  `.github/workflows/deploy-production.yml`, staging secrets,
  `infra/monitoring/synthetics.yaml`.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-088-D00`: The controlled staging signup allowlist must be persistently backed
by a GitHub secret (staging-scoped or repository-level, implementer choice per
`VOC-088-DEP-00`). `deploy-staging.yml` reads this secret and syncs it to
`api.env` on every deploy. The `workflow_dispatch` input
`new_user_signup_allowlist` must either be removed entirely or made additive-only
(union with the secret), never a replacement. The secret is the single source of
truth.

`VOC-088-D01`: `NEW_USER_SIGNUP_ENABLED` remains `false` in staging. Only
explicitly admitted normalized emails in the persistent allowlist may create
staging accounts. No blanket-enable path.

`VOC-088-D02`: Deploy-staging.yml must fail closed when the allowlist secret is
present but empty and `GOOGLE_OAUTH_ENABLED` is true, because that state means
controlled first-time signup is expected but no one can use it. The deploy log
must expose only a count or boolean ("allowlist: N entries"), never email values.

`VOC-088-D03`: The staging OAuth/readiness synthetic must assert controlled
signup readiness: when OAuth is enabled, global signup is disabled, and controlled
signup is an expected capability, the synthetic fails if the controlled cohort is
empty. This requires the API's `/healthz` or a dedicated non-sensitive endpoint to
expose a `signup_allowlist_count` field (integer, no email values). The synthetic
asserts `signup_allowlist_count > 0`.

`VOC-088-D04`: Failed scheduled synthetics and deploy workflows open exactly one
deduplicated, sanitized, plain unlabeled GitHub issue using the `GITHUB_TOKEN`
App identity. The issue body contains the workflow name, run URL, failure step,
and a sanitized summary. It never contains secrets, session values, OAuth state,
raw user identifiers, or unfiltered container logs. Deduplication: if an open
issue with the same fingerprint (workflow + failure category) already exists, no
new issue is created.

`VOC-088-D05`: Responsibility separation is preserved:
- Sentry (error-monitoring.yml): unexpected application errors
- Kuma (sync-monitoring.yml): availability/TLS/basic API health
- Synthetics (scheduled-synthetics.yml): expected behavior/configuration readiness
- Failure-to-issue agent: governed remediation for failed operational workflows

`VOC-088-D06`: The failure-to-issue agent covers `scheduled-synthetics.yml`,
`deploy-staging.yml`, and `deploy-production.yml`. It does not cover
`error-monitoring.yml` (which already opens issues from Sentry) or
`sync-monitoring.yml` (which is a controlled manual dispatch, not scheduled).

`VOC-088-D07`: No email addresses or OAuth credentials in git, issues, workflow
logs, or test fixtures that resemble real users. Test fixtures use
`@synthetic.vocanova.invalid` addresses only.

## Contradictions / open questions

1. **Secret scope** (`VOC-088-DEP-00`): Whether the allowlist secret is
   repository-level or environment-scoped to staging is an implementer/operator
   choice. Either satisfies the requirement as long as `deploy-staging.yml` can
   read it and `deploy-production.yml` cannot accidentally use it.
2. **Dispatch input behavior** (`VOC-088-D00`): removing the dispatch input
   entirely is the simplest safe option. Making it additive (union with secret)
   is acceptable but adds complexity. The implementer should choose based on
   operator need; the reviewing human settles this at adoption.
3. **Readiness endpoint shape** (`VOC-088-D03`): extending `/healthz` with a
   `signup_allowlist_count` field vs. adding a new `/readiness` endpoint is an
   implementer choice. `/healthz` already exposes kill-switch state; adding a
   count is minimal. The reviewing human may prefer a separate endpoint if
   `/healthz` is used by Docker HEALTHCHECK and the count could affect health
   semantics.
4. **Risk class:** measured path floor is R3. No R4 path is touched. Semantic
   risk (staging secrets, operational workflow changes) is R3-proportionate.

## Security and privacy

- Email addresses only in GitHub secrets; never committed, never printed in logs,
  issues, workflow output, or evidence.
- The readiness endpoint exposes only a count (integer), never email values.
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
