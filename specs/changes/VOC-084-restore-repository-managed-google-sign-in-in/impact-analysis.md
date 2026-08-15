# VOC-084 — Impact Analysis

## Security and privacy

- **Secrets:** repository `GOOGLE_OAUTH_CLIENT_ID` /
  `GOOGLE_OAUTH_CLIENT_SECRET` are synchronized into the staging secret file
  only. Mode `0600`. Never echo, mask-print, or commit credential values.
  Partial-pair states fail closed before convergence.
- **OAuth redirect integrity:** staging callback is fixed to the canonical
  API `/api/v1` path on `api-staging.vocanova.site`. Live checks must not
  follow Google or complete login.
- **Account creation:** `NEW_USER_SIGNUP_ENABLED` stays `false`. Allowlist
  defaults empty so first-time signup stays closed unless an operator
  explicitly expands the cohort via the repository workflow control.
- **UI honesty:** do not advertise Google when capability is disabled
  (prevents user-facing dead ends and reduces confused-deputy probing).
- **No real-user data mutation** required for acceptance. Synthetic /
  allowlisted tester paths only where already supported.
- **Production boundary:** do not change production OAuth sync semantics;
  do not share staging and production secret files, directories, deploy
  users, databases, or Docker networks.

## Data and migrations

- No schema migration.
- No database seed changes required for OAuth restore.
- Config-only convergence of env flags and secrets on the staging host via
  `deploy-staging.yml`.
- Rollback returns staging to a coherent disabled OAuth state when the
  pair is absent after revert/redeploy; partial states must remain
  impossible.

## Analytics and accessibility

- **Analytics:** None expected — evidence-backed non-applicability for
  product analytics instrumentation.
- **Accessibility:** Hiding the Google button when disabled must leave
  remaining sign-in methods operable with existing labels/focus behavior.
  Preserve the sign-in page `max-w-[28rem]` workaround from
  `.karsift/lessons.md`.

## Risks, dependencies, and evidence

- `VOC-084-R00`: **Credential leak** via workflow logs or committed files.
  Mitigation: AC-01/AC-05; never log secret values; mode `0600` only on
  staging secret file.
- `VOC-084-R01`: **Partial-pair half-config** enabling broken OAuth.
  Mitigation: AC-00/TEST-00 fail before convergence.
- `VOC-084-R02`: **Wrong callback URI** (port, host, or path drift) causing
  Google rejection after start returns 200. Mitigation: AC-02/TEST-03/TEST-06
  exact URI assertions.
- `VOC-084-R03`: **Open signup** if allowlist defaults wrong or
  `NEW_USER_SIGNUP_ENABLED` flips true. Mitigation: AC-03/TEST-04.
- `VOC-084-R04`: **UI still advertises Google** while API is disabled.
  Mitigation: AC-04/TEST-05.
- `VOC-084-R05`: **Production regression** from shared workflow edits or
  shared secret paths. Mitigation: AC-06; staging-only file scope; preserve
  VOC-067 isolation.
- `VOC-084-R06`: **False "done"** when Google Console still lacks the staging
  callback. Mitigation: AC-07/TEST-08; record exact external action when
  access is unavailable (`VOC-084-DEP-01`).
- `VOC-084-DEP-00`: Root cause resolved at drafting (missing staging sync +
  UI advertise-while-disabled).
- `VOC-084-DEP-01`: Google client staging-callback authorization (open).
- `VOC-084-DEP-02`: Staging allowlist control surface (open; recommended
  default documented).
- `VOC-084-EV-00`: T00 deploy/config/test evidence (`t00-evidence.md`).
- `VOC-084-EV-01`: T01 UI capability evidence (`t01-evidence.md`).
- `VOC-084-EV-02`: T02 live OAuth-start + Google callback disposition
  (`t02-evidence.md`).
