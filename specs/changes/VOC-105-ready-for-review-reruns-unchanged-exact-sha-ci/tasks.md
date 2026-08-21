# VOC-105 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
Calling-repo wiring/doc/foundation-test changes land in this repository under
the same package.

## VOC-105-T00 — Unchanged-SHA ready-for-review reuse gate, fail-closed semantics, docs, deterministic tests

- Requirement source: issue #872; `VOC-105-D00`–`D06`, `D08`
- Acceptance criteria: `VOC-105-AC-00` through `VOC-105-AC-05`, `VOC-105-AC-07`;
  T00 contributes the verifier implementation and deterministic TEST-11
  evidence to `VOC-105-AC-06`, which T01 closes with the controlled live proof.
- Tests: `VOC-105-TEST-00` through `VOC-105-TEST-07`, `VOC-105-TEST-09`,
  `VOC-105-TEST-10`, `VOC-105-TEST-11`
- Evidence: `VOC-105-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update `KARSIFT/karsift-ai-infra` reusable workflow/helpers so that on
   `pull_request` action `ready_for_review`, a deterministic classifier evaluates
   whether prior exact-SHA CI and independent-review evidence may be reused
   before starting full CI or model review (`VOC-105-D00`, `VOC-105-D01`).
2. When every `VOC-105-D01` precondition holds: skip full CI and skip model
   review / plan-review; still run deterministic merge-gate re-evaluation for
   the now-ready PR (`VOC-105-D02`). Draft PRs remain non-mergeable.
3. When any precondition fails or cannot be proven: run the normal full CI +
   applicable independent review + merge-gate path (`VOC-105-D03`).
4. Never treat human/implementer/non-App comments as reusable verification
   authority (`VOC-105-D05`).
5. Preserve exact-SHA stale-run protections and independent-role separation
   (`VOC-105-D04`, `VOC-105-D05`). Ensure `opened` / `synchronize` / `reopened`
   still always take the full CI + review path.
6. Add deterministic positive, negative, human-comment rejection, draft-never-
   merge, synchronize-regression, and caller-wiring tests (infra self-ci and/or
   `scripts/foundation/voc105-*.test.mjs`) (`VOC-105-D06`).
7. Update karsift-ai-infra README and any calling-repo docs that would otherwise
   remain false about ready_for_review always re-running full CI and model
   review. Update AGENTS.md only if its current text would otherwise become
   false.
8. Record that the caller already consumes reusable workflows at `@main`; no
   pin bump is expected. If implementation discovers a different current
   reference, reconcile it explicitly and record the actual consumption
   mechanism.
9. Add the read-only `verify-ready-for-review-reuse` workflow-dispatch action
   described by `VOC-105-D07` / the implementation plan: on the exact carrier
   branch it validates declared source run metadata using only Actions/PR
   metadata. Its caller and reusable inner job MUST produce the exact contract
   job display name `verify-ready-for-review-reuse / verify`. It must never
   read logs or artifacts and must receive no writes, model keys, deploy
   secrets, or application secrets.
10. Run applicable tests and governance validation for changed calling-repo
    paths; record commands and results in `t00-evidence.md` (no secrets).
11. Do not address deprecated action inputs, Node runtime warnings, dependency
    alerts, or deterministic remediation preflight (out of scope per
    `VOC-105-D08`).

### Explicitly out of scope for this task

- Controlled live draft → ready proof (T01).
- Granting implementer Actions credentials.
- Weakening merge-gate draft blocking or App-only verdict trust.
- Application or monitoring-inventory ID changes.

## VOC-105-T01 — Controlled draft-to-ready proof (operator-owned live evidence)

- Requirement source: issue #872 acceptance; `VOC-105-D07`
- Acceptance criteria: `VOC-105-AC-01`, `VOC-105-AC-06`
- Tests: `VOC-105-TEST-08`, `VOC-105-TEST-11`
- Evidence: `VOC-105-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-105-T00`
- **Acceptance requires operator-owned live evidence** (not implementer Actions
  access). Contract:
  `.karsift/live-evidence/VOC-105-T01.yaml`.
- Automation ownership: operator

### Required work

1. After T00 is live on the branch the pipeline executes from, perform a
   **controlled** draft → ready transition with unchanged base/head and trusted
   prior exact-SHA PASS evidence. Preferred: dogfood this package's own T00→T01
   carrier path after T00 merges, or an equivalent controlled agent PR prepared
   only for this proof.
2. Confirm the ready_for_review pipeline run skipped full CI and model review
   while still running merge-gate re-evaluation.
3. Confirm a changed-SHA or missing-verdict control (deterministic fixture
   preferred; sanitized live observation allowed) still takes the normal full
   path.
4. Record allowlisted metadata only in `t01-evidence.md` (prior run IDs,
   ready_for_review run IDs, job conclusions, reuse decision, PR number). Never
   copy logs, secrets, sessions, OAuth data, cookies, tokens, or user
   identifiers.
5. Commit those allowlisted source metadata to the carrier, manually dispatch
   the read-only proof action on `agent/voc-105-voc-105-t01`, and require its
   run HEAD to equal the current carrier PR head before reconciliation.
6. Do not expand scope into deprecated-action, Node-runtime, dependency-alert,
   or remediation-preflight work. Do not expand scope into unrelated pipeline
   edits; waiting/reconcile are handled by governed automation after VOC-097.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/script/test/doc/pin edits).
- Granting implementer Actions credentials.
- Out-of-scope roots listed in issue #872 scope boundary.

## Task ordering notes

- T00 blocks T01: live proof requires the reuse gate to be what the pipeline
  actually runs.
- After T00 merges, auto-advance MUST NOT dispatch the general implementer for
  this package's T01 (T01 is itself operator-owned). That skip is part of the
  intended VOC-102 behavior, not a pipeline defect.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
