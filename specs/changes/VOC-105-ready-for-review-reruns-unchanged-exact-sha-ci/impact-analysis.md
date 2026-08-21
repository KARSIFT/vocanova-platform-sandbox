# VOC-105 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Does not grant the implementer Actions
  credentials.
- **Verification surface:** Narrows when full CI and model review re-execute on
  `ready_for_review`. Incorrect reuse could skip independent verification;
  fail-closed preconditions and exact-SHA guards are mandatory.
- **Trust boundary:** Only trusted App-authored PASS / PASS WITH NON-BLOCKING
  FINDINGS bound to the exact base/head and package/task authority may authorize
  reuse. Human and implementer comments never qualify.
- **Signals:** Proof and evidence use allowlisted Actions metadata only (run
  IDs, job names, conclusions, boolean reuse decision). Forbidden: logs,
  artifacts, secrets, OAuth/session/cookie/token material, user identifiers.
- **Residual risk:** Over-broad skip conditions. Mitigation: `VOC-105-D01`
  conjunction, deterministic negative tests, merge-gate draft and stale-run
  protections retained.

## Application and operational surface

- **Application code:** No intentional change.
- **Operational effect:** Unchanged-SHA draft → ready transitions avoid
  duplicate CI and model compute while still re-entering merge-gate.
- **Release / deploy:** Unchanged application release semantics. Package lands
  through normal develop → main promotion for calling-repo wiring; infra merges
  on the karsift-ai-infra path.
- **Cross-repo:** Primary behavior change is in `KARSIFT/karsift-ai-infra`
  reusable workflows/helpers; calling-repo `pipeline.yml` conditions and any
  narrow verify action land here.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts the reuse gate; pre-fix behavior (always re-run CI and
  review on ready_for_review) would return — known costly but safe.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- Accessibility — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-105-R00`: **Unsafe reuse skips verification.** Mitigation: conjunction
  of exact SHA, required checks, App-signed verdict, scope binding, and
  live-evidence attestation rules; negative tests for each failure mode.
- `VOC-105-R01`: **Draft auto-merge regression.** Mitigation: merge-gate draft
  block retained; TEST-07.
- `VOC-105-R02`: **Stale run merges after a newer push.** Mitigation: existing
  expected_base_sha / expected_head_sha guards; TEST-09 synchronize regression.
- `VOC-105-R03`: **Human comment accepted as PASS.** Mitigation: App-identity
  only; TEST-05.
- `VOC-105-R04`: **Docs claim universal full re-run.** Mitigation: AC-07 /
  TEST-10.
- `VOC-105-DEP-00`: Issue #872 incident runs on PR #868 and #869.
- `VOC-105-DEP-01`: Caller already subscribes to ready_for_review for
  draft-aware merge-gate.
- `VOC-105-DEP-02`: Cross-repo infra change ownership pattern.
- `VOC-105-DEP-03`: VOC-102 deferred this exact root.
- `VOC-105-EV-00`: T00 evidence — reuse-gate summary, deterministic test
  output, doc alignment notes (no secrets).
- `VOC-105-EV-01`: T01 evidence — scrubbed run/job metadata proving safe reuse
  skip on controlled draft → ready (operator-owned live evidence).
