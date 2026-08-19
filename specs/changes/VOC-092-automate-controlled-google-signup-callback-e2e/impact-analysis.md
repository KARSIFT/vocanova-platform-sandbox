# VOC-092 — Impact Analysis

## Security and privacy

- **Auth boundary:** The harness touches OAuth start/callback handlers and
  controlled-signup policy. It must not weaken production wiring or introduce a
  bypass reachable outside ephemeral tests.
- **Fake provider:** Localhost-only fake Google OAuth server; no exposure through
  the shared edge or production `production.go` wiring.
- **Secrets and personal data:** Harness and CI must never log or persist real
  Google credentials, OAuth codes, state, tokens, cookies, or session values.
  Only `@synthetic.vocanova.invalid` fixture identities.
- **Database isolation:** Disposable Postgres only; no staging/production
  `DATABASE_URL` reuse.
- **Allowlist secrets:** No change to `STAGING_NEW_USER_SIGNUP_ALLOWLIST` or
  deploy-staging.yml behavior.

## Data and migrations

- No application schema migration.
- Ephemeral test databases created and destroyed per test run.
- No production or staging data mutation.

## Analytics and accessibility

- **Analytics:** None — evidence-backed non-applicability.
- **Accessibility:** No product UI change.

## Risks, dependencies, and evidence

- **`VOC-092-R00`:** A poorly scoped harness could accidentally become a runtime
  backdoor if wired into production provider selection. Mitigation: fake server
  and injectable URLs exist only in test packages; foundation tests grep for forbidden
  test-auth routes and production env switches.
- **`VOC-092-R01`:** Integration tests could leak OAuth query strings or cookies
  into CI logs. Mitigation: assert redirect status/Location and DB state without
  printing sensitive values; redaction tests in foundation suite.
- **`VOC-092-R02`:** CI job misconfiguration could point at a non-disposable
  database. Mitigation: harness starts its own Postgres container or uses an
  isolated DSN generated in-test; foundation test forbids staging/production host
  strings in harness sources.
- **`VOC-092-DEP-00`:** VOC-088 merged — controlled staging signup operational.
- **`VOC-092-DEP-01`:** `GoogleOAuthProvider` injectable endpoints exist.
- **`VOC-092-DEP-02`:** Disposable Postgres integration patterns exist in repo.
- **`VOC-092-DEP-03`:** Operational-failure observer exists for observed workflows.
- **`VOC-092-EV-00`:** T00 integration harness test results and scrubbed output.
- **`VOC-092-EV-01`:** T01 CI run and foundation test results.
- **`VOC-092-EV-02`:** T02 documentation diff and VOC-088 evidence remediation.
- **`VOC-092-EV-03`:** T03 live CI and staging synthetic verification.

## Risk class note

This package **proposes R3** as a draft for the reviewing human at adoption time.
The path-based classifier (`scripts/governance/classify-change-risk.sh`) and
independent verification bind the authoritative class per task. Semantic auth
sensitivity does not by itself justify R4 when the change is test-only with no
production bypass.
