# VOC-113 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02**.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
Calling-repo pipeline/doc/foundation-test changes land in this repository under
the same package. Infra PRs must say `Relates to
KARSIFT/vocanova-platform-sandbox#<task>` and MUST NOT use a closing keyword.

## VOC-113-T00 — Diagnose missing-run behavior; implement durable recovery, tests, and docs

- Requirement source: issue #948; `VOC-113-D00`–`D07`, `VOC-113-D09`, `VOC-113-D10`
- Acceptance criteria: `VOC-113-AC-00` through `VOC-113-AC-04`
- Tests: `VOC-113-TEST-00` through `VOC-113-TEST-07`
- Evidence: `VOC-113-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record in `t00-evidence.md` the issue #948 observations (VOC-112 develop merge
   without push workflows; promotion PR #947 without required checks; failed
   close/reopen and draft/ready recovery; reconcile-release reuse without merge).
   Determine and document the precise trigger/token/event behavior
   (`VOC-113-DEP-04`) using allowlisted metadata only — no logs, secrets, tokens,
   or personal data.
2. Implement durable recovery in shared infra so that:
   - after an App-driven task merge to the integration branch, a missing required
     push/validation workflow set for the exact squash SHA is recovered via
     explicit reusable-workflow or dispatch orchestration;
   - after App-created promotion PR creation (and on `reconcile-release`), a
     missing required pull-request check set for the exact head SHA is likewise
     recovered.
3. Preserve GitHub App authentication for automated mutations; keep the existing
   refusal to merge with `github.token` when App credentials are configured.
4. Reuse VOC-108 authoritative exact-head selection; fail closed on absent,
   stale, wrong-SHA, pending-past-timeout, cancelled, or latest-failed evidence.
5. Prevent duplicate promotion PRs, duplicate release audits, unbounded workflow
   recursion, and any fabrication of check runs or commit statuses.
6. Add a bounded wait with actionable sanitized diagnostics on timeout
   (duration chosen in implementation and recorded in evidence).
7. Preserve the active ruleset required contexts (`governance-policy`,
   `validate`, `ci / ci` or their current equivalents) and governance risk
   classification.
8. Add deterministic tests for task-merge-to-integration recovery and release-PR
   recovery, including positive recovery and negatives for wrong SHA, timeout,
   duplicate promotion, recursion, and fabricated-status refusal.
9. Add the read-only `verify-promotion-check-recovery` caller
   `workflow_dispatch` action described by `VOC-113-D10`: it validates a
   declared promotion PR's exact-head required checks using Actions/check
   metadata only, produces job display name
   `verify-promotion-check-recovery / verify`, never merges, never fabricates
   statuses, and receives no model keys, deploy secrets, or application secrets.
10. Update karsift-ai-infra README and calling-repo docs (AGENTS.md and/or ops
    docs) only where current claims would become false — in particular, do not
    leave text implying close/reopen or draft/ready recovers missing checks.
11. Record that the caller already consumes reusable workflows at `@main`; no pin
    bump is expected. If implementation discovers a different current reference,
    reconcile it explicitly and record the actual consumption mechanism.
12. Run applicable commands and record results in `t00-evidence.md`:
    - infra policy / self-ci tests added or updated by this task;
    - caller `node --test scripts/foundation/voc113-*.test.mjs` when present;
    - `bash scripts/governance/validate-governance.sh` when required for changed
      calling-repo paths;
    - `bash scripts/governance/classify-change-risk.sh` when required;
    - `git diff --check`.

### Explicitly out of scope for this task

- Completing promotion PR #947 (T01) and post-promotion verification (T02).
- Weakening ruleset checks or synthesizing statuses.
- Product/runtime/signup/credential/topology changes.
- Unrelated Actions efficiency work.
- Granting the implementer general Actions credentials.

## VOC-113-T01 — Recover and complete promotion PR #947 after genuine exact-head checks

- Requirement source: issue #948 required outcome; `VOC-113-D08`, `VOC-113-D10`
- Acceptance criteria: `VOC-113-AC-02`, `VOC-113-AC-05`
- Tests: `VOC-113-TEST-08`
- Evidence: `VOC-113-EV-01` (`t01-evidence.md`)
- Live-evidence contract: `.karsift/live-evidence/VOC-113-T01.yaml`
- Status: pending — depends on `VOC-113-T00`
- Automation ownership: operator

### Required work

1. After T00 is live on the branch the caller pipeline executes from, enter the
   operator-owned waiting state. Through repository-controlled recovery /
   `reconcile-release` (not implementer Actions access), ensure promotion PR
   #947's **exact head SHA** receives genuine required checks.
2. Confirm required contexts succeed for that exact head (ruleset contexts
   `governance-policy`, `validate`, and `ci / ci`, or their current equivalents).
3. Dispatch or observe the T00 `verify-promotion-check-recovery` action on this
   task PR so the live-evidence contract can qualify (`VOC-113-D10`).
4. Allow release converge to perform the single exact-head merge decision only
   after VOC-108 authoritative selection reports all required newest checks
   successful. Do not open a second promotion PR or release audit.
5. Record only allowlisted metadata in `t01-evidence.md` (PR number, head SHA,
   run/job IDs, conclusions, verify-run metadata, merge outcome). No logs or
   secrets.
6. Acceptance requires **operator-owned live evidence** per
   `docs/operations/live-evidence.md`.

### Explicitly out of scope for this task

- Code changes (T00 owns the mechanism).
- Post-promotion `main` workflow verification (T02).
- Implementer-owned Actions dispatch or log inspection.
- Status fabrication or ruleset weakening.

## VOC-113-T02 — Verify post-promotion workflows and close remediation

- Requirement source: issue #948 required outcome; `VOC-113-D08`
- Acceptance criteria: `VOC-113-AC-06`
- Tests: `VOC-113-TEST-09`
- Evidence: `VOC-113-EV-02` (`t02-evidence.md`)
- Live-evidence contract: `.karsift/live-evidence/VOC-113-T02.yaml`
- Status: pending — depends on `VOC-113-T01`
- Automation ownership: operator

### Required work

1. After PR #947 merges, resolve the promotion result SHA on `main` using
   operator-owned repository metadata only.
2. Confirm expected post-promotion workflows ran for that SHA — at minimum the
   repository's normal `main` push path (for example `deploy-production` when
   selected by existing policy). Record allowlisted run metadata only.
3. Close issue #948 / this remediation only after that live evidence is complete.
   Issue closure is a wake-up/visibility signal, not completion proof by itself.
4. Acceptance requires **operator-owned live evidence** per
   `docs/operations/live-evidence.md`.

### Explicitly out of scope for this task

- Code or workflow mechanism changes.
- Implementer-owned Actions access.
- Changing production deploy policy, signup, or monitoring inventory.

## Task ordering notes

- T00 blocks T01: recovery must be live before #947 can obtain genuine checks
  through the durable mechanism.
- T01 blocks T02: post-promotion verification requires #947 to have merged under
  genuine exact-head success.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Do not use a snapshot-then-drift commit under this package directory as
  promotion evidence; package bookkeeping is not new unreviewed scope.

Tasks preserve scope, separation of duties, and rollback safety.
