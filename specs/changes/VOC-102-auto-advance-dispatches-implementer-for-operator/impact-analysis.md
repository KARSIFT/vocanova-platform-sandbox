# VOC-102 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Does not grant the implementer Actions credentials.
- **Dispatch surface:** Narrows when `implement.yml` may start after a task close —
  operator-owned / live-evidence-only next tasks no longer receive a general
  implementer run that could edit evidence files before real proof exists.
- **Signals:** Skip/fail comments or markers are sanitized metadata only (task IDs,
  ownership mode, boolean dispatch decision, scrubbed run URLs). Forbidden: logs,
  artifacts, secrets, OAuth/session/cookie/token material, user identifiers.
- **Residual risk:** Mis-classifying an ordinary implementation task as
  operator-owned would stall progress until corrected (fail-visible). Mis-allowing
  dispatch on malformed metadata is prevented by fail-closed tests (`VOC-102-D04`).

## Application and operational surface

- **Application code:** No intentional change.
- **Operational effect:** Auto-advance cooperates with VOC-097 waiting/reconcile
  instead of wasting implementer capacity on evidence-only successors.
- **Release:** Unchanged final-roster semantics; operator tasks must still close
  before release opens.
- **Cross-repo:** Primary behavior change is in `KARSIFT/karsift-ai-infra`
  `auto-advance.yml`; calling repo may only pin/docs-align.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts auto-advance ownership gating; pre-fix behavior (dispatch
  implementer for operator-owned next tasks) would return — known undesirable.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-102-R00`: **Ordinary tasks stalled** if ownership detection false-positives.
  Mitigation: positive-dispatch tests; contract presence required for skip;
  fail-closed only when metadata is required/contradictory, not when absent for
  ordinary tasks.
- `VOC-102-R01`: **Early release** if skip is mistaken for task completion.
  Mitigation: issue remains OPEN; TEST-07; release still keyed off issue close /
  check-completion.
- `VOC-102-R02`: **Reconcile assumes waiting PR** and skip-without-implement leaves
  no PR for evidence attach. Mitigation: open question 1 / DEP-03; implementer must
  confirm reconcile path works from open task issue or document minimal non-implementer
  waiting bookkeeping without using general implement.yml.
- `VOC-102-R03`: **Docs drift** claiming universal implement dispatch.
  Mitigation: AC-07 / TEST-10; AGENTS.md doc-consistency rule.
- `VOC-102-DEP-00`: Issue #863 incident (VOC-098-T00→T01 spurious implement).
- `VOC-102-DEP-01`: VOC-097 ownership contract and reconcile path.
- `VOC-102-DEP-02`: Cross-repo infra change ownership pattern.
- `VOC-102-DEP-03`: Skip-signaling shape (confirm at adoption).
- `VOC-102-EV-00`: T00 evidence — ownership-gate summary, deterministic test output,
  doc alignment notes (no secrets).
- `VOC-102-EV-01`: T01 evidence — scrubbed auto-advance / pipeline run metadata
  proving zero implementer dispatch for operator-owned next task and retained
  dispatch for ordinary next task (operator-owned live evidence).
