# VOC-102 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Does not grant the implementer Actions credentials.
- **Dispatch surface:** Narrows when `implement.yml` may start after a task close —
  operator-owned / live-evidence-only next tasks no longer receive a general
  implementer run that could edit evidence files before real proof exists.
- **Signals:** Carrier/fail comments or markers are sanitized metadata only (task IDs,
  ownership mode, boolean dispatch decision, scrubbed run URLs). Forbidden: logs,
  artifacts, secrets, OAuth/session/cookie/token material, user identifiers.
- **Mutation boundary:** The classifier remains read-only. A separate clean,
  deterministic publisher (no LLM/model keys, no Actions-write) alone mints the
  existing App for contents/issues/pull-requests write to create/reuse the waiting
  carrier PR and deduplicated issue marker.
- **Proof boundary:** The post-carrier verifier is also read-only. It reads only
  Actions run/job, issue, PR, and repository metadata; never logs or artifacts;
  and receives no App write token, inherited/model/deploy/application secrets, or
  Actions-write permission.
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
  `auto-advance.yml`; the calling repo also wires the narrow read-only proof action
  used to bind T01 evidence to its exact carrier SHA.

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
  Mitigation: positive-dispatch tests; contract presence required for skip; only
  one exact marker inside the matching task stanza can declare that a contract is
  required. Narrative prose is ignored, and absence of marker plus contract stays
  ordinary.
- `VOC-102-R01`: **Early release** if skip is mistaken for task completion.
  Mitigation: issue remains OPEN; TEST-07; release still keyed off issue close /
  check-completion.
- `VOC-102-R02`: **Reconcile assumes waiting PR.** Mitigation: resolved DEP-03
  selects a deterministic non-implementer carrier publisher. Tests require a draft
  PR with the strict task-ID-derived package-local evidence path before declaring
  the operator task safely waiting; an issue comment alone cannot pass. The
  unchanged VOC-097 contract schema receives no arbitrary-path field.
- `VOC-102-R03`: **Docs drift** claiming universal implement dispatch.
  Mitigation: AC-07 / TEST-10; AGENTS.md doc-consistency rule.
- `VOC-102-R04`: **Mutation authority broadens silently.** Mitigation: `advance`
  remains read-only; only the clean publisher mints the App for explicit
  contents/issues/pull-requests writes, with no model or Actions-write credentials.
- `VOC-102-R05`: **Pre-carrier evidence cannot satisfy PR-head lineage.**
  Mitigation: retain the real T00-close operational observation, then run a
  separate metadata-only verifier on the exact carrier head. TEST-13 rejects a
  stale head, mismatched source event, or executed implement job. Existing
  reconciler lineage modes remain unchanged.
- `VOC-102-R06`: **Partial publisher run strands an existing carrier.**
  Mitigation: ownership classification precedes the path-specific existing-PR
  decision; operator paths re-enter the idempotent publisher to validate and
  repair missing derived evidence/marker state, while ordinary paths retain the
  duplicate-implement guard. TEST-11 covers both partial boundaries.
- `VOC-102-DEP-00`: Issue #863 incident (VOC-098-T00→T01 spurious implement).
- `VOC-102-DEP-01`: VOC-097 ownership contract and reconcile path.
- `VOC-102-DEP-02`: Cross-repo infra change ownership pattern.
- `VOC-102-DEP-03`: Resolved — deterministic clean evidence-carrier PR plus a
  deduplicated sanitized task-issue marker.
- `VOC-102-DEP-04`: Resolved — two-stage live proof: real transition first, then a
  read-only verifier whose successful run equals the carrier PR head.
- `VOC-102-EV-00`: T00 evidence — ownership-gate summary, deterministic test output,
  doc alignment notes (no secrets).
- `VOC-102-EV-01`: T01 evidence — scrubbed auto-advance / pipeline run metadata
  proving zero implementer dispatch for operator-owned next task and retained
  dispatch for ordinary next task (operator-owned live evidence).
