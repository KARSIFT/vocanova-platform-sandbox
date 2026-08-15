# VOC-084 — Acceptance Criteria

## VOC-084-AC-00 — Partial Google credential pair fails before convergence

- Requirement source: issue #691; `VOC-084-D00`
- Tasks: `VOC-084-T00`
- Tests: `VOC-084-TEST-00`
- Evidence: `VOC-084-EV-00`
- Result: pending

Observable outcome:

1. Staging deploy fails before application convergence when exactly one of
   `GOOGLE_OAUTH_CLIENT_ID` / `GOOGLE_OAUTH_CLIENT_SECRET` is present.
2. Failure does not write a half-configured OAuth state and does not log
   credential values.

## VOC-084-AC-01 — Complete pair syncs safely; absent pair disables coherently

- Requirement source: issue #691; `VOC-084-D00`; `VOC-084-D02`
- Tasks: `VOC-084-T00`
- Tests: `VOC-084-TEST-01`, `VOC-084-TEST-02`
- Evidence: `VOC-084-EV-00`
- Result: pending

Observable outcome:

1. When both repository credentials are present, they are written only to
   the staging secret file with mode `0600` and without log exposure.
2. When both are absent, staging converges to a coherent disabled OAuth state
   (`GOOGLE_OAUTH_ENABLED=false` / equivalent), not a stale contradictory
   mix.
3. Live staging API has OAuth enabled only when both credentials are available.

## VOC-084-AC-02 — Canonical staging OAuth callback URI

- Requirement source: issue #691; `VOC-084-D01`
- Tasks: `VOC-084-T00`, `VOC-084-T02`
- Tests: `VOC-084-TEST-03`, `VOC-084-TEST-06`
- Evidence: `VOC-084-EV-00`, `VOC-084-EV-02`
- Result: pending

Observable outcome:

1. Deploy-derived staging config sets
   `OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`
   exactly.
2. Live OAuth start returns HTTP 200 and a Google authorization URL whose
   `redirect_uri` query/parameter value is exactly that URI.

## VOC-084-AC-03 — Staging signup remains controlled

- Requirement source: issue #691; `VOC-084-D03`; `VOC-084-DEP-02`
- Tasks: `VOC-084-T00`
- Tests: `VOC-084-TEST-04`
- Evidence: `VOC-084-EV-00`
- Result: pending

Observable outcome:

1. `NEW_USER_SIGNUP_ENABLED` remains `false` on staging after deploy
   convergence.
2. An explicit, repository-workflow-controlled staging
   `NEW_USER_SIGNUP_ALLOWLIST` exists, defaults empty, and supports approved
   first-time testers when non-empty.
3. Empty allowlist means no first-time signup; existing users remain able to
   sign in.

## VOC-084-AC-04 — UI does not advertise disabled Google sign-in

- Requirement source: issue #691; `VOC-084-D04`
- Tasks: `VOC-084-T01`
- Tests: `VOC-084-TEST-05`
- Evidence: `VOC-084-EV-01`
- Result: pending

Observable outcome:

1. When live/deployed OAuth capability reports disabled, the sign-in UI
   does not display "Continue with Google".
2. The gate uses a deploy-derived / auth-capability signal (prefer
   `/healthz` `kill_switches.oauth_enabled`), not a staging-only hardcoded lie
   and not "build succeeded ⇒ show button".

## VOC-084-AC-05 — Deterministic tests and live check without real login

- Requirement source: issue #691; `VOC-084-D05`
- Tasks: `VOC-084-T00`, `VOC-084-T01`, `VOC-084-T02`
- Tests: `VOC-084-TEST-00` through `VOC-084-TEST-07`
- Evidence: `VOC-084-EV-00`, `VOC-084-EV-01`, `VOC-084-EV-02`
- Result: pending

Observable outcome:

1. Deterministic tests cover config/workflow/UI branches listed in the
   issue (both-present, both-absent, partial rejection, canonical callback,
   allowlist, disabled-method rendering).
2. The live OAuth-start check is part of `deploy-staging.yml`.
3. No real OAuth callback/login is performed by CI.
4. No secrets appear in repository, issue, PR, or workflow output.

## VOC-084-AC-06 — Topology and public health remain healthy

- Requirement source: issue #691; `VOC-084-D06`
- Tasks: `VOC-084-T00`, `VOC-084-T02`
- Tests: `VOC-084-TEST-07`
- Evidence: `VOC-084-EV-02`
- Result: pending

Observable outcome:

1. Existing healthy shared-edge topology and public health endpoints remain
   healthy after staging deployment of this package's changes.
2. Staging/production secret, directory, deploy-user, database, and
   Docker-network isolation is preserved.
3. Production OAuth behavior is not intentionally changed and does not
   regress from in-scope edits.

## VOC-084-AC-07 — Google client callback authorization verified or precisely recorded

- Requirement source: issue #691; `VOC-084-DEP-01`
- Tasks: `VOC-084-T02`
- Tests: `VOC-084-TEST-08`
- Evidence: `VOC-084-EV-02`
- Result: pending

Observable outcome:

1. From available access/evidence, either verify that the existing Google
   OAuth client authorizes the staging callback URI, **or**
2. Report the exact one-time external Google Cloud Console/API configuration
   action required, without claiming that external step complete when access
   was unavailable.
3. Do not invent Console mutations or credentials in the repository.

Acceptance criteria are observable, stable, security-aware, and
bidirectionally traceable to requirements, tasks, tests, and evidence.
