# VOC-102 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
Calling-repo doc/pin/foundation-test changes land in this repository under the
same package.

## VOC-102-T00 — Auto-advance ownership gate, fail-closed semantics, docs, deterministic tests

- Requirement source: issue #863; `VOC-102-D00`–`D06`, `D08`
- Acceptance criteria: `VOC-102-AC-00` through `VOC-102-AC-05`, `VOC-102-AC-07`,
  `VOC-102-AC-08`
- Tests: `VOC-102-TEST-00` through `VOC-102-TEST-07`, `VOC-102-TEST-10` through
  `VOC-102-TEST-13`
- Evidence: `VOC-102-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update `KARSIFT/karsift-ai-infra/.github/workflows/auto-advance.yml` so that,
   after resolving the next roster task and applying existing open-issue /
   existing-PR guards, it determines next-task ownership from governed package
   data before setting `should_dispatch=true`.
2. Authoritative machine-readable source when present:
   `<package_path>/.karsift/live-evidence/<next_task_id>.yaml`.
3. If `ownership` is `operator` or `live-actions`: do **not** dispatch
   `implement.yml`; leave the next issue OPEN; select the separate deterministic
   clean evidence-carrier publisher defined by `VOC-102-D01/D02`.
4. That publisher creates/reuses the deterministic task branch/draft PR, writes
   only the strict task-ID-derived `<package_path>/tNN-evidence.md` path, and posts one deduplicated
   sanitized waiting marker. The classifier remains read-only; the publisher
   alone may mint the App for contents/issues/pull-requests writes and receives no
   model credentials or Actions-write permission.
5. If the next task is ordinary (no live-evidence contract and no contradictory
   operator declaration): preserve today's implementer dispatch attempt 1.
6. Fail closed on missing-required / malformed / unrecognized / contradictory
   ownership metadata per `VOC-102-D04` (no dispatch, no carrier from untrusted
   paths, one sanitized publisher escalation).
7. Preserve final-roster release behavior per `VOC-102-D05` (skip ≠ complete).
8. Add deterministic positive, negative, malformed-metadata, carrier-idempotency,
   permission-boundary, and regression tests
   (infra self-ci and/or `scripts/foundation/voc102-*.test.mjs`).
9. Update karsift-ai-infra README and calling-repo
   `docs/operations/live-evidence.md` / AGENTS.md only where current text would
   otherwise claim auto-advance always dispatches implement for every next task.
10. Bump calling-repo `pipeline.yml` pin only if required to consume the fixed
    reusable workflow; record the consumption mechanism in evidence.
11. Add the read-only `verify-auto-advance-live-evidence` workflow-dispatch action
    described by `VOC-102-D07`: on the exact carrier branch it validates a declared
    source run and waiting PR using only Actions/issue/PR metadata. Its caller and
    reusable inner job MUST produce the exact contract job display name
    `verify-auto-advance-live-evidence / verify`. It must never read logs or
    artifacts and must receive no writes, model keys, deploy secrets, or application
    secrets.
12. Run applicable tests and governance validation for changed calling-repo paths;
    record commands and results in `t00-evidence.md` (no secrets).
13. Do not address duplicate exact-SHA reviews, action-runtime upgrades, or
    cache-path warnings (out of scope per `VOC-102-D08`).

### Explicitly out of scope for this task

- Controlled live workflow proof (T01).
- Granting implementer Actions credentials.
- Changing VOC-097 reconcile allowlists or remediation waiting classification
  beyond what auto-advance must cooperate with.
- Application or monitoring-inventory ID changes.

## VOC-102-T01 — Controlled sanitized workflow proof (operator-owned live evidence)

- Requirement source: issue #863 acceptance; `VOC-102-D07`
- Acceptance criteria: `VOC-102-AC-01`, `VOC-102-AC-02`, `VOC-102-AC-06`
- Tests: `VOC-102-TEST-08`, `VOC-102-TEST-09`
- Evidence: `VOC-102-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-102-T00`
- **Acceptance requires operator-owned live evidence** (not implementer Actions
  access). Contract:
  `.karsift/live-evidence/VOC-102-T01.yaml`.

### Required work

1. After T00 is live on the branch auto-advance executes from, perform a
   **controlled** sanitized workflow proof that when the next roster task is
   operator-owned / live-evidence-only, auto-advance completes with no executed
   `implement.yml` job for that task. Dogfood this package's own T00→T01 advance
   after T00 merges (T01 must not receive a general implementer).
2. Prove an ordinary no-contract next task still dispatches as intended
   (deterministic fixture preferred; sanitized live observation allowed if needed).
3. Confirm the operator-owned next task issue remained OPEN, exactly one draft
   evidence-carrier PR was created/reused, the pending evidence path matched the
   strict task-ID-derived convention, and the sanitized waiting marker was deduplicated.
4. Record allowlisted metadata only in `t01-evidence.md` (pipeline/auto-advance run
   IDs/URLs, conclusions, `should_dispatch` outcome, issue number, absence of
   implementer job). Never copy logs, secrets, sessions, OAuth data, cookies,
   tokens, or user identifiers.
5. Commit those allowlisted source metadata to the carrier, manually dispatch the
   read-only proof action on `agent/voc-102-voc-102-t01`, and require its run HEAD
   to equal the current carrier PR head before reconciliation.
6. Do not manufacture live evidence for unrelated packages (do not re-run VOC-098
   T01 proof as part of this task). Do not expand scope into unrelated pipeline
   edits; waiting/reconcile are handled by governed automation after VOC-097.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/script/test/doc/pin edits).
- Granting implementer Actions credentials.
- Duplicate exact-SHA / action-runtime / cache-path follow-ups.

## Task ordering notes

- T00 blocks T01: live proof requires the ownership gate to be what auto-advance
  actually runs.
- After T00 merges, auto-advance MUST NOT dispatch the general implementer for
  this package's T01 (T01 is itself operator-owned). That skip is part of the
  intended proof, not a pipeline defect.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
