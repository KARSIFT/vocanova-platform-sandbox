# VOC-097 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02 → (T03 ∥ T04) → T05**.
T03 and T04 may proceed in parallel after T02. T05 depends on both.

Cross-repo note: T01–T03 primarily change `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for those behaviors; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree. Calling-repo
doc/template/wiring changes land in this repository under the same package.

## VOC-097-T00 — Declare live-evidence contract and author-facing docs/templates

- Requirement source: `VOC-097-D02`, `VOC-097-D03`; issue #823 outcomes 11
- Acceptance criteria: `VOC-097-AC-00`
- Tests: `VOC-097-TEST-00`, `VOC-097-TEST-01`
- Evidence: `VOC-097-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Document the allowlisted live-evidence contract fields and ownership model in
   calling-repo operator/governance docs (preferred new file under
   `docs/operations/`, plus AGENTS.md / template cross-links only where current
   text would otherwise remain false).
2. Extend `specs/templates/change-package/` guidance (README and/or tasks template
   notes) so future packages know how to mark operator-owned live-evidence tasks
   and where the machine-readable contract lives.
3. Record the chosen contract path/shape from `VOC-097-D03` / adoption resolution
   of DEP-03 in `t00-evidence.md` (no secrets).
4. Do not yet change remediation behavior (T01) or add the reconciler (T02).

### Explicitly out of scope for this task

- `remediate.yml` / reconcile workflow behavior (T01/T02).
- Migrating #779/#785 (T04).
- Application code changes.

## VOC-097-T01 — Waiting lifecycle; stop remediation on operator-owned pending evidence

- Requirement source: `VOC-097-D00`, `VOC-097-D01`; issue #823 outcomes 1, 2, 7
- Acceptance criteria: `VOC-097-AC-01`, `VOC-097-AC-02`, `VOC-097-AC-03`
- Tests: `VOC-097-TEST-02`, `VOC-097-TEST-03`, `VOC-097-TEST-04`, `VOC-097-TEST-05`
- Evidence: `VOC-097-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-097-T00`

### Required work

1. In `KARSIFT/karsift-ai-infra`, introduce an explicit waiting-for-operator-live-evidence
   lifecycle signal distinct from `VERDICT: FAIL` for code/docs defects.
2. Update `prompts/review.md` (and plan-review only if needed) so reviewers classify
   "pending declared live evidence only" as waiting guidance — not as a code-defect
   FAIL that must be remediations-fixed by editing workflows. If a FAIL verdict is
   still required for merge-gate fail-closed behavior, pair it with a stable
   machine-readable waiting marker that `remediate.yml` honors.
3. Update `remediate.yml` so waiting-for-operator-live-evidence does **not** set
   `should_retry=true` and does not re-dispatch implement. Preserve retry behavior
   for genuine FAIL, CI failure, and review-job error.
4. Confirm `implement.yml` does not gain general Actions credentials; document the
   invariant in infra README and tests.
5. While waiting, do not instruct or allow the implementer to make unrelated
   pipeline/workflow edits solely to manufacture evidence.

### Explicitly out of scope for this task

- Observe/dispatch reconciler (T02).
- Stranded task migration (T04).
- Calling-repo application monitoring inventory changes.

## VOC-097-T02 — Allowlisted observe/dispatch reconciler with timeout, dedup, wake + re-review

- Requirement source: `VOC-097-D01`–`VOC-097-D07`; issue #823 outcomes 3–6, 8
- Acceptance criteria: `VOC-097-AC-03`, `VOC-097-AC-04`, `VOC-097-AC-05`,
  `VOC-097-AC-06`, `VOC-097-AC-07`
- Tests: `VOC-097-TEST-06` through `VOC-097-TEST-14`
- Evidence: `VOC-097-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-097-T01`

### Required work

1. Add repository-controlled automation in `karsift-ai-infra` (new reusable workflow
   and/or pipeline job) that:
   - reads the task's declared evidence contract;
   - optionally dispatches **only** the declared workflow with declared inputs when
     the contract permits dispatch;
   - otherwise observes `workflow_run` / Actions API metadata for matching runs;
   - validates workflow identity, jobs, event, branch/SHA lineage, conclusion,
     and staleness;
   - writes only allowlisted sanitized metadata onto the governed task path;
   - wakes waiting → ready-for-re-review and requires a **new** exact-SHA
     independent review before merge-gate can proceed.
2. Fail closed on ambiguous/malformed/non-matching evidence; leave waiting or
   record sanitized rejection without starting remediation-as-fix.
3. Implement bounded timeout/escalation and deduplication per `VOC-097-D06`.
4. Keep permissions least-privilege and separate from implementer credentials.
5. Emit only boolean/state/count-style lifecycle signals plus sanitized run IDs;
   do not integrate with operational-failure observer or Sentry.
6. Wire the calling repository's `pipeline.yml` (or pin bump) only as needed for
   this sandbox to consume the new behavior; record the consumption mechanism in
   evidence.

### Explicitly out of scope for this task

- Full stranded migration of #779/#785 (T04) beyond enabling the mechanism.
- Changing Kuma/synthetic inventory YAML.

## VOC-097-T03 — Deterministic fixture matrix

- Requirement source: issue #823 outcomes 9; `VOC-097-D02`, `VOC-097-D05`, `VOC-097-D06`
- Acceptance criteria: `VOC-097-AC-09`
- Tests: `VOC-097-TEST-02` through `VOC-097-TEST-14` (complete matrix)
- Evidence: `VOC-097-EV-03` (`t03-evidence.md`)
- Status: pending — depends on `VOC-097-T02`

### Required work

1. Land or complete deterministic tests (infra self-ci and/or
   `scripts/foundation/voc097-*.test.mjs` plus any vendored fixtures under
   `tooling/governance/fixtures/`) covering:
   - waiting does not invoke remediation;
   - successful reconciliation (single wake);
   - timeout/escalation;
   - wrong workflow / job / branch / SHA;
   - duplicate events;
   - sanitization / allowlist;
   - no-credential and no-log invariants.
2. Run applicable governance validation and risk classification on calling-repo
   diffs; record commands and results in evidence (no secrets).

### Explicitly out of scope for this task

- Live production/staging proof (T05).
- Closing #779/#785 (T04).

## VOC-097-T04 — Reconcile stranded tasks #779 and #785

- Requirement source: `VOC-097-D08`; issue #823 outcomes 10
- Acceptance criteria: `VOC-097-AC-08`
- Tests: `VOC-097-TEST-15`
- Evidence: `VOC-097-EV-04` (`t04-evidence.md`)
- Status: pending — depends on `VOC-097-T02`

### Required work

1. For issue #779 (`VOC-093-T01` / PR #789) and issue #785 (`VOC-094-T01` / PR #791):
   - place each on the governed waiting path with a valid evidence contract, **or**
   - execute the documented safe migration path (clean evidence-only PR /
     reset-to-in-scope tip) recorded in `t04-evidence.md`.
2. Prefer observe of already-healthy qualifying runs when lineage matches; do not
   invent unrelated workflow edits.
3. After qualifying evidence + fresh exact-SHA review + merge (or documented
   migration closure), record scrubbed run URLs and issue states.
4. Do not grant implementer Actions credentials.

### Explicitly out of scope for this task

- Re-opening VOC-093-T00 / VOC-094-T00 code work unless evidence proves the
  underlying system is still broken (then open a new unlabeled issue per AGENTS.md
  rather than expanding this task).

## VOC-097-T05 — Controlled live proof and observer/Sentry separation health

- Requirement source: `VOC-097-D07`; issue #823 verification + monitoring
- Acceptance criteria: `VOC-097-AC-06`, `VOC-097-AC-10`
- Tests: `VOC-097-TEST-11`, `VOC-097-TEST-16`
- Evidence: `VOC-097-EV-05` (`t05-evidence.md`)
- Status: pending — depends on `VOC-097-T03`, `VOC-097-T04`

### Required work

1. Run a controlled repository fixture proving waiting does not invoke remediation
   (may reuse T03 fixture if already live-proven; otherwise record a sandbox PR
   demonstration).
2. Run a controlled qualifying workflow completion proving one reconciliation and
   a fresh exact-SHA review requirement.
3. Confirm deploy/synthetic failure observation remains healthy and separate
   (scrubbed references only; no coupling of waiting signals into failure-to-issue
   or Sentry).
4. Record results in `t05-evidence.md` without secrets or personal data.

### Explicitly out of scope for this task

- Further infra feature work (belongs in T01–T03).

## Task ordering notes

- T00 establishes the author-facing contract before automation depends on it.
- T01 must land before T02 so waiting exists for the reconciler to clear.
- T03 locks the fixture matrix against the T02 mechanism.
- T04 unblocks historical stranded tasks once the mechanism exists.
- T05 is the live closure proof for the package.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
