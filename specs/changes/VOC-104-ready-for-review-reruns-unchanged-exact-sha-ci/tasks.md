# VOC-104 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
Calling-repo pipeline/doc/foundation-test changes land in this repository under
the same package.

## VOC-104-T00 — Ready-for-review reuse policy, fail-closed path, docs, deterministic tests

- Requirement source: issue #872; `VOC-104-D00`–`D07`, `D09`, `D10`
- Acceptance criteria: `VOC-104-AC-00` through `VOC-104-AC-05`, `VOC-104-AC-07`;
  T00 contributes the verifier implementation and deterministic TEST-12 evidence
  to `VOC-104-AC-06`, which T01 closes with the controlled live proof.
- Tests: `VOC-104-TEST-00` through `VOC-104-TEST-07`, `VOC-104-TEST-10` through
  `VOC-104-TEST-12`
- Evidence: `VOC-104-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Add a deterministic ready_for_review reuse-eligibility decision to shared
   infra (helper and/or reusable workflow coordination per `VOC-104-D01`) that
   inspects only Actions/check/comment metadata and emits a machine-readable
   outcome (`reuse-evidence` vs full path).
2. Wire the caller pipeline template and this repository's
   `.github/workflows/pipeline.yml` so that on `ready_for_review`, when reuse is
   allowed, `ci` and model-invoking `review` / `plan-review` are skipped while
   `merge-gate` still runs and remains reachable when review siblings are
   skipped.
3. Enforce all `VOC-104-D02` preconditions; on any failure take the normal full
   CI + review path (`VOC-104-D04`).
4. Preserve draft never auto-merges (`VOC-104-D00`) and reject
   human/implementer comments as reusable authority (`VOC-104-D05`).
5. Preserve exact-SHA stale-run protections and independent-role separation
   (`VOC-104-D06`). Non-`ready_for_review` PR activities keep today's full path.
6. Apply the optimized path to both `agent/` and `plan/` PRs per the resolved
   `VOC-104-D10` decision.
7. Add deterministic shared-infra policy tests and calling-repository
   fixture/foundation coverage (`scripts/foundation/voc104-*.test.mjs` and/or
   infra self-ci) per `VOC-104-D07`.
8. Update karsift-ai-infra README and calling-repo DOC-15 §17.3 so “fresh
   pipeline evaluation” explicitly distinguishes safe exact-SHA reuse from the
   normal full CI/model-review path.
9. Record that the caller already consumes reusable workflows at `@main`; no pin
   bump is expected. If implementation discovers a different current reference,
   reconcile it explicitly and record the actual consumption mechanism.
10. Add the read-only `verify-ready-for-review-reuse` workflow-dispatch action
    described by `VOC-104-D08`: on the exact proof PR branch it validates a
    declared ready_for_review source run using only Actions/check metadata (prior
    run ID, ready_for_review run ID, skipped CI/review jobs, successful
    merge-gate, unchanged base/head). Its caller and reusable inner job MUST
    produce the exact contract job display name
    `verify-ready-for-review-reuse / verify`. It must never read logs or
    artifacts and must receive no writes, model keys, deploy secrets, or
    application secrets.
11. Run applicable tests and governance validation for changed calling-repo
    paths; record commands and results in `t00-evidence.md` (no secrets).
12. Do not address deprecated action inputs, Node runtime warnings, dependency
    alerts, or deterministic remediation preflight (out of scope per
    `VOC-104-D09`).

### Explicitly out of scope for this task

- Controlled live draft→ready proof (T01).
- Granting implementer Actions credentials.
- Application or monitoring-inventory ID changes.

## VOC-104-T01 — Controlled draft-to-ready optimized-path proof (operator-owned live evidence)

- Requirement source: issue #872 acceptance; `VOC-104-D08`
- Acceptance criteria: `VOC-104-AC-01`, `VOC-104-AC-06`
- Tests: `VOC-104-TEST-08`, `VOC-104-TEST-09`
- Evidence: `VOC-104-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-104-T00`
- **Acceptance requires operator-owned live evidence** (not implementer Actions
  access). Contract:
  `.karsift/live-evidence/VOC-104-T01.yaml`.
- Automation ownership: operator

### Required work

1. After T00 is live on the branch the caller pipeline executes from, perform a
   **controlled** draft→ready transition on a PR whose exact base/head already
   has green required checks and a trusted App-authored PASS (preferred: a
   short-lived controlled proof PR or this package's deterministic evidence
   carrier after T00 merge).
2. Confirm the ready_for_review pipeline run skipped full CI and model review,
   still ran merge-gate re-evaluation, and did not merge while draft.
3. Record allowlisted metadata only in `t01-evidence.md`: prior successful run
   ID, ready_for_review run ID, job names/conclusions (CI/review skipped;
   merge-gate success), base/head SHAs, reuse decision boolean. Never copy logs,
   secrets, sessions, OAuth data, cookies, tokens, or user identifiers.
4. Commit those allowlisted source metadata to the carrier/proof branch, manually
   dispatch the read-only `verify-ready-for-review-reuse` pipeline action on that
   exact branch, and require its run HEAD to equal the current PR head before
   reconciliation.
5. Do not expand scope into unrelated pipeline edits; waiting/reconcile are
   handled by governed automation after VOC-097.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/script/test/doc edits).
- Granting implementer Actions credentials.
- Deprecated-action / Node-runtime / dependency-alert / remediation-preflight
  follow-ups.

## Task ordering notes

- T00 blocks T01: live proof requires the reuse policy to be what the caller
  pipeline actually runs.
- After T00 merges, auto-advance MUST NOT dispatch the general implementer for
  this package's T01 (T01 is itself operator-owned). That skip is part of the
  intended VOC-102 behavior, not a pipeline defect.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
