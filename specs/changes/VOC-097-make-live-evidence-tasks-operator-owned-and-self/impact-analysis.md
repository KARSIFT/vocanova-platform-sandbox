# VOC-097 — Impact Analysis

## Security and privacy

- **Secrets / credentials:** Do not grant the implementer general GitHub Actions
  credentials. Any new Actions observe/dispatch capability lives only in a
  repository-controlled reconciler with narrowly scoped permissions and
  App-authenticated mutations consistent with existing karsift patterns.
- **Evidence sanitization:** Allowlisted metadata only. Forbidden: logs, artifacts,
  secrets, OAuth data, sessions, cookies, tokens, user identifiers, arbitrary job
  output (`VOC-097-D02`).
- **Least privilege:** Preserve builder/verifier separation, branch protection, and
  existing App-token mutation boundaries.
- **Abuse:** Malformed or spoofed evidence contracts fail closed; wrong
  workflow/job/branch/SHA cannot wake a task.

## Application and operational surface

- **Application code:** No intentional product behavior change.
- **Governance automation:** New waiting lifecycle and reconcile path in
  karsift-ai-infra; calling-repo docs/templates and optional `pipeline.yml` wiring.
- **Remediation:** Waiting no longer consumes remediation retries; genuine FAIL/CI
  failure remediation preserved.
- **Monitoring inventory:** No Kuma monitor or synthetic ID add/update. Lifecycle
  signals are workflow-internal boolean/state/count style only.
- **Observer separation:** Operational-failure-monitoring and Sentry remain separate
  (`VOC-097-D07`).

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts infra workflow/prompt/doc commits; stranded-task migration
  records remain audit evidence.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- No accessibility change — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-097-R00`: **False waiting** hides a real code defect and skips remediation.
  Mitigation: waiting only when a declared live-evidence contract exists and the
  review/reconcile path classifies the sole gap as pending operator evidence;
  genuine FAIL without waiting marker still remediates; tests TEST-03/TEST-04.
- `VOC-097-R01`: **False wake** accepts non-qualifying runs. Mitigation: fail-closed
  identity/lineage/conclusion/staleness checks; TEST-06–TEST-08, TEST-12.
- `VOC-097-R02`: **Sensitive data leakage** via reconcile comments. Mitigation:
  allowlist sanitizer + denylist tests TEST-09/TEST-10.
- `VOC-097-R03`: **Indefinite waiting** without escalation. Mitigation: bounded
  timeout + single escalation TEST-13.
- `VOC-097-R04`: **Cross-repo lag** — calling repo still pins old infra and never
  receives waiting/reconcile. Mitigation: T02 records consumption/pin; open
  question 4; T05 live proof on this sandbox.
- `VOC-097-DEP-00`–`DEP-04`, `VOC-093`, `VOC-094`, `VOC-088`: see `change.yaml`.
- `VOC-097-EV-00` … `VOC-097-EV-05`: per-task evidence files `t00`–`t05-evidence.md`.
