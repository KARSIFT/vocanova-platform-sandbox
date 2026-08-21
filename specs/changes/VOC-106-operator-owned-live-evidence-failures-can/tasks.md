# VOC-106 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
Calling-repo doc/pin/foundation-test changes land in this repository under the
same package.

## VOC-106-T00 — Remediation ownership gate, fail-closed escalation, docs, deterministic tests

- Requirement source: issue #882; `VOC-106-D00`–`D07`
- Acceptance criteria: `VOC-106-AC-00` through `VOC-106-AC-05`, `VOC-106-AC-07`,
  `VOC-106-AC-08`; T00 contributes the verifier implementation and deterministic
  TEST-11 evidence to `VOC-106-AC-06`, which T01 closes with the controlled live
  proof.
- Tests: `VOC-106-TEST-00` through `VOC-106-TEST-07`, `VOC-106-TEST-10`,
  `VOC-106-TEST-11`
- Evidence: `VOC-106-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update `KARSIFT/karsift-ai-infra/config/decide-remediation.py` and
   `.github/workflows/remediate.yml` so that, after exact-head/base guards and
   PR-body task/package identity parsing, remediation resolves task ownership
   from governed package data at the exact reviewed PR head before any
   implementer retry.
2. Authoritative machine-readable source when present at that head:
   `<package_path>/.karsift/live-evidence/<task_id>.yaml`.
   The only secondary expectation signal is an exact
   `- Automation ownership: operator` or
   `- Automation ownership: live-actions` line inside that task's own `## <task_id>`
   stanza in canonical `tasks.md`. Do not infer ownership from prose. Prefer
   reusing VOC-102 ownership-classifier primitives where safe.
3. If a valid contract establishes `ownership` as `operator` or `live-actions`:
   do **not** dispatch `implement.yml` for review `FAIL` or CI failure. Emit a
   sanitized escalate-operator / equivalent outcome per `VOC-106-D02`.
4. Preserve `WAITING` suppression, `STALE` no-op, missing-evidence lifecycle
   handling, and `REVIEW_INFRA_FAILURE` non-retry paths without consuming an
   implementation attempt. Route malformed or mismatched ownership metadata to
   the separate fail-closed path in `VOC-106-D04`.
5. If the task is ordinary (no live-evidence contract and no contradictory
   operator declaration): preserve today's bounded remediate → implement retry
   (attempt capped at 2).
6. Keep merge-gate fail-closed; do not invent PASS from ownership.
7. Add deterministic workflow-policy and decision-helper tests covering operator
   WAITING, FAIL, CI failure, stale result, malformed/mismatched contract, and
   ordinary implementer FAIL and CI failure (infra self-ci and/or
   `scripts/foundation/voc106-*.test.mjs`).
8. Update karsift-ai-infra README and calling-repo
   `docs/operations/live-evidence.md` to document ownership-gated remediation for
   operator-owned FAIL/CI states and the retained ordinary bounded retry. Update
   AGENTS.md only if its current text would otherwise become false.
9. Record that the caller already consumes reusable workflows at `@main`; no pin
   bump is expected. If implementation discovers a different current reference,
   reconcile it explicitly and record the actual consumption mechanism.
10. Add the read-only `verify-remediate-operator-ownership` workflow-dispatch
    action described by `VOC-106-D06`: on the exact carrier branch it validates a
    declared source run using only Actions/issue/PR metadata. Its caller and
    reusable inner job MUST produce the exact contract job display name
    `verify-remediate-operator-ownership / verify`. It must never read logs or
    artifacts and must receive no writes, model keys, deploy secrets, or
    application secrets.
11. Run applicable tests and governance validation for changed calling-repo paths;
    record commands and results in `t00-evidence.md` (no secrets).
12. Do not address Node runtime deprecations, action-input migration, Go cache
    configuration, dependency alerts, or application changes (out of scope per
    `VOC-106-D07`).

### Explicitly out of scope for this task

- Controlled live workflow proof (T01).
- Granting implementer Actions credentials.
- Changing VOC-097 reconcile allowlists or waiting verdict classification beyond
  what remediation must cooperate with.
- Application or monitoring-inventory ID changes.

## VOC-106-T01 — Controlled sanitized workflow proof (operator-owned live evidence)

- Requirement source: issue #882 acceptance; `VOC-106-D06`
- Acceptance criteria: `VOC-106-AC-01`, `VOC-106-AC-02`, `VOC-106-AC-06`
- Tests: `VOC-106-TEST-08`, `VOC-106-TEST-09`
- Evidence: `VOC-106-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-106-T00`
- **Acceptance requires operator-owned live evidence** (not implementer Actions
  access). Contract:
  `.karsift/live-evidence/VOC-106-T01.yaml`.
- Automation ownership: operator

### Required work

1. After T00 is live on the branch remediation executes from, perform a
   **controlled** sanitized workflow proof that when an operator-owned
   live-evidence carrier experiences review `FAIL` or CI failure, remediation
   completes with no executed `implement.yml` job and emits sanitized
   operator/reconcile escalation metadata instead.
2. Prove an ordinary no-contract implementer-owned FAIL still retries as
   intended (deterministic fixture preferred; sanitized live observation
   allowed if needed).
3. Confirm merge remained fail-closed for the failing head and that no
   implementation attempt was consumed for the operator-owned case.
4. Record allowlisted metadata only in `t01-evidence.md` (pipeline/remediate run
   IDs/URLs, conclusions, `should_retry` outcome, absence of implementer job,
   escalation marker presence). Never copy logs, secrets, sessions, OAuth data,
   cookies, tokens, user identifiers, or evidence payloads.
5. Commit those allowlisted source metadata to the carrier, manually dispatch the
   read-only proof action on `agent/voc-106-voc-106-t01`, and require its run HEAD
   to equal the current carrier PR head before reconciliation.
6. Do not manufacture live evidence for unrelated packages (do not re-run VOC-104
   T01 proof as part of this task). Do not expand scope into unrelated pipeline
   edits; waiting/reconcile are handled by governed automation after VOC-097.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/script/test/doc/pin edits).
- Granting implementer Actions credentials.
- Node runtime / action-input / Go cache / dependency-alert follow-ups.

## Task ordering notes

- T00 blocks T01: live proof requires the ownership gate to be what remediation
  actually runs.
- After T00 merges, auto-advance MUST NOT dispatch the general implementer for
  this package's T01 (T01 is itself operator-owned; VOC-102 already covers that
  skip). That skip is expected, not a pipeline defect.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
