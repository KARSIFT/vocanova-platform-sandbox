# VOC-098 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Continues to use existing `KARSIFT_BOT_APP_ID` /
  `KARSIFT_BOT_PRIVATE_KEY`. App token scope **narrows** (drops Actions request).
  Job `GITHUB_TOKEN` may gain explicit `actions: read` for classifier metadata only.
- **Issue authorship:** Issue create/dedupe remains App-only so `plan-from-issue`
  still starts. `GITHUB_TOKEN` must not create issues.
- **Personal data / sanitization:** Unchanged allowlist — workflow name, conclusion,
  canonical run URL, fixed summary, HTML marker. No logs, secrets, sessions, OAuth
  data, cookies, tokens, or user identifiers in issues or evidence.
- **Least privilege:** Aligns App mint with installation grants and with
  `error-monitoring.yml` precedent (`permission-issues: write` only).

## Application and operational surface

- **Application code:** No intentional change.
- **Operational effect:** Restores failure-to-issue for watched workflows currently
  blocked by mint failure (notably `scheduled-synthetics` and any deploy
  non-success that never reaches classification).
- **Classifier:** Benign deploy-cancel skip path remains; Actions API reads move to
  job token. Fail-closed on API error/ambiguity preserved.
- **Residual risk:** Mis-wiring `GITHUB_TOKEN` into issue create would break
  plan-from-issue; tests and AC lock against that. Wrong permissions floor would
  make classifier fail closed toward opening issues for benign cancels until fixed.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts workflow/test/doc commits; observer returns to App mint that
  requests Actions (known broken against current installation) unless installation
  permissions change outside this package (out of scope).

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-098-R00`: **Classifier loses Actions access** if job permissions omit
  `actions: read`. Mitigation: explicit permissions floor + TEST-03; fail-closed
  opens issues rather than silently skipping real failures.
- `VOC-098-R01`: **GITHUB_TOKEN accidentally used for issues** breaks
  plan-from-issue. Mitigation: TEST-02; AC-01; code review of env wiring.
- `VOC-098-R02`: **Live proof blocked until default-branch promotion** because the
  observer checks out the default-branch workflow. Mitigation: T01 after T00 on
  `main`; open question 4.
- `VOC-098-R03`: **Open marker already exists** so create path is not exercised.
  Mitigation: AC-04 accepts create-or-dedupe; duplicate re-check still required.
- `VOC-098-DEP-00`: Issue #840 — unnecessary App Actions permission causes mint
  failure.
- `VOC-098-DEP-01`: VOC-088 / VOC-094 observer and classifier predecessors.
- `VOC-098-DEP-02`: Documented controlled cancel fixture for live proof.
- `VOC-098-EV-00`: T00 evidence — workflow permission/token wiring summary,
  foundation test output.
- `VOC-098-EV-01`: T01 evidence — scrubbed observer + watched-run metadata, issue
  marker/dedupe confirmation (operator-owned live evidence).
