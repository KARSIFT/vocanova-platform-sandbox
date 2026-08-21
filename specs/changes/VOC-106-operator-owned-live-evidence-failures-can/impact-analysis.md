# VOC-106 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Does not grant the implementer Actions credentials.
- **Dispatch surface:** Narrows when `implement.yml` may start after CI/review
  failure — operator-owned / live-evidence-only tasks no longer receive a general
  implementer run that could mutate evidence carriers after an exact-head FAIL.
- **Signals:** Escalation comments or markers are sanitized metadata only (task
  IDs, ownership mode, boolean no-retry decision, scrubbed run URLs/reason codes).
  Forbidden: logs, artifacts, secrets, OAuth/session/cookie/token material, user
  identifiers, evidence payloads, raw review body replay.
- **Proof boundary:** The post-carrier verifier is read-only. It reads only
  Actions run/job, issue, PR, and repository metadata; never logs or artifacts;
  and receives no App write token, inherited/model/deploy/application secrets, or
  Actions-write permission.
- **Residual risk:** Mis-classifying an ordinary implementation task as
  operator-owned would stall automatic remediation until corrected (fail-visible).
  Mis-allowing dispatch on malformed metadata is prevented by fail-closed tests
  (`VOC-106-D04`).

## Application and operational surface

- **Application code:** No intentional change.
- **Operational effect:** Remediation cooperates with VOC-097 ownership and
  reconcile instead of waking the general implementer for evidence-only tasks.
- **Merge:** Remains fail-closed; this package must not weaken merge-gate.
- **Cross-repo:** Primary behavior change is in `KARSIFT/karsift-ai-infra`
  `remediate.yml` / `decide-remediation.py`; the calling repo also wires the
  narrow read-only proof action used to bind T01 evidence to its exact carrier
  SHA.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts remediation ownership gating; pre-fix behavior (dispatch
  implementer for operator-owned FAIL/CI failure) would return — known
  undesirable.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-106-R00`: **Ordinary tasks lose automatic remediation** if ownership
  detection false-positives. Mitigation: positive-retry tests; contract presence
  required for skip; only one exact marker inside the matching task stanza can
  declare that a contract is required. Narrative prose is ignored.
- `VOC-106-R01`: **Deadlock** if operator FAIL escalates without a clear
  reconcile/operator path. Mitigation: `VOC-106-D02` requires an explicit
  sanitized escalation; stale/missing states must not consume attempts
  (`VOC-106-D04`); TEST-05/06/07.
- `VOC-106-R02`: **Merge weakened** if ownership skip is mistaken for PASS.
  Mitigation: merge-gate remains independent and fail-closed; AC-01 explicitly
  preserves fail-closed merge; no PASS is invented from ownership.
- `VOC-106-R03`: **Secret/log leakage** via escalation comments. Mitigation:
  allowlisted metadata only; TEST-07/AC-08; continue-on-error patterns must not
  fetch logs/artifacts.
- `VOC-106-R04`: **Docs drift** claiming universal implementer retry on FAIL.
  Mitigation: AC-07 / TEST-10; AGENTS.md doc-consistency rule.
- `VOC-106-DEP-00`: Issue #882 incident (VOC-104-T01 spurious remediate→implement).
- `VOC-106-DEP-01`: VOC-097 ownership contract; VOC-102 auto-advance gate;
  remaining remediation gap.
- `VOC-106-DEP-02`: Cross-repo infra change ownership pattern.
- `VOC-106-DEP-03`: WAITING already suppressed; FAIL/CI/stale/malformed still open.
- `VOC-106-EV-00`: T00 evidence — ownership-gate summary, deterministic test
  output, doc alignment notes (no secrets).
- `VOC-106-EV-01`: T01 evidence — scrubbed remediate / pipeline run metadata
  proving zero implementer dispatch for operator-owned FAIL/CI failure and
  retained retry for ordinary FAIL (operator-owned live evidence).
