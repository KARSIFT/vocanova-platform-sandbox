# VOC-114 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
Calling-repo evidence/docs changes land in this repository under the same
package. Infra PRs must say `Relates to
KARSIFT/vocanova-platform-sandbox#<task>` and MUST NOT use a closing keyword.

## VOC-114-T00 — Restore recovery metadata-read token contract, localize errors, add tests

- Requirement source: issue #956; `VOC-114-D00`–`D05`
- Acceptance criteria: `VOC-114-AC-00` through `VOC-114-AC-03`
- Tests: `VOC-114-TEST-00` through `VOC-114-TEST-05`
- Evidence: `VOC-114-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record in `t00-evidence.md` the issue #956 observations (PR #954 merge at
   `b97e9575…`, pipeline runs 32696249484 and 32696549963, immediate
   `github_metadata_read_failed`, blocked PR #947 / VOC-113-T01). Determine and
   document the verified token/installation cause using allowlisted metadata only.
2. Update every path that feeds `actions-check-recovery-runner.py` (merge-gate
   post-merge recovery, release converge recovery, and
   `recover-actions-checks.yml`) so recovery uses the job `GITHUB_TOKEN` with
   Actions write plus Checks/Statuses/Contents/Pull requests read. Preserve App
   tokens solely for their existing Contents/Issues/Pull requests mutations.
3. Record the observed App installation permissions in `t00-evidence.md`. If the
   App lacks Actions permission, do not broaden the installation: keep Actions
   reads/dispatch on the dedicated job-token boundary and fail closed when that
   job grant is absent.
4. Refactor `actions-check-recovery-runner.py` (and shared `gh` adapter if
   extracted) so check-runs, workflow-runs, and commit-metadata failures raise
   sanitized endpoint classes per `VOC-114-D02` instead of one generic
   `github_metadata_read_failed`.
5. Ensure metadata-read failures abort before bounded wait and before any recovery
   dispatch planning or execution (`VOC-114-D00`).
6. Extend infra policy tests (for example `test_voc113_actions_check_recovery.py`
   and focused VOC-114 additions) covering:
   - positive metadata read under declared token contract for both modes;
   - absent read permission fail-closed with correct endpoint class;
   - no dispatch after metadata-read failure;
   - token separation and preserved mutation posture (no App Actions broadening).
7. Update karsift-ai-infra README and calling-repo/package docs wherever the
   recovery credential or App permission claims would otherwise become false.
8. Run applicable validation and record results in `t00-evidence.md`:
   - infra policy / self-ci tests added or updated by this task;
   - `bash scripts/governance/validate-governance.sh` when required for changed
     calling-repo paths;
   - `bash scripts/governance/classify-change-risk.sh` when required;
   - `git diff --check`.

### Explicitly out of scope for this task

- Operator-owned live recovery proof (T01).
- Weakening ruleset checks or synthesizing statuses.
- Manually merging promotion PR #947.
- Unrelated VOC-113 feature work beyond the metadata-read defect.
- Product/runtime/signup/credential/topology changes.

## VOC-114-T01 — Live recovery proof for both modes and unblock promotion PR #947

- Requirement source: issue #956 required outcome; `VOC-114-D06`; VOC-113-T01
  blocked outcome
- Acceptance criteria: `VOC-114-AC-04`, `VOC-114-AC-05`
- Tests: `VOC-114-TEST-06`, `VOC-114-TEST-07`
- Evidence: `VOC-114-EV-01` (`t01-evidence.md`)
- Live-evidence contract: `.karsift/live-evidence/VOC-114-T01.yaml`
- Status: pending — depends on `VOC-114-T00`
- Automation ownership: operator

### Required work

1. After T00 is live on the branch the caller pipeline executes from, confirm
   the job-token Actions/metadata grants and mutation-only App contract from T00
   are active; no App installation expansion is required.
2. Re-run integration_push recovery for the documented merged SHA
   `b97e9575fd30671c336a2e92ca00db6e29b86416` (or the still-blocking SHA recorded
   in T00 evidence if `develop` advanced). Verify the run progresses past
   metadata read and produces or observes genuine push/validation workflow runs
   for that exact SHA — not fabricated statuses.
3. Dispatch package-authorized `reconcile-release` for release issue #946 (or
   observe the equivalent release converge recovery). Verify promotion PR #947's
   exact head receives genuine required pull-request checks through recovery.
4. Confirm required contexts succeed for #947's exact head (`governance-policy`,
   `validate`, `ci / ci` or current equivalents) and release converge remains
   fail-closed until VOC-108 authoritative selection reports success.
5. Dispatch `pipeline.yml` on the exact T01 evidence-carrier branch with
   `action=verify-promotion-check-recovery` and `promotion_pr_number=947`.
   Require job `verify-promotion-check-recovery / verify` to succeed on the
   carrier's exact PR head; the live-evidence contract must not try to prove a
   carrier commit through a `develop` run that cannot contain it.
6. Record only allowlisted metadata in `t01-evidence.md` (SHAs, run/job IDs,
   check names, conclusions, sanitized error classes if any intermediate failure
   occurred, promotion PR head SHA). No logs or secrets.
7. Acceptance requires **operator-owned live evidence** per
   `docs/operations/live-evidence.md`.

### Explicitly out of scope for this task

- Code changes (T00 owns the mechanism).
- Post-promotion `main` workflow verification (remains VOC-113-T02).
- Implementer-owned Actions dispatch or log inspection.
- Status fabrication, ruleset weakening, or manual merge of #947.

## Task ordering notes

- T00 blocks T01: metadata reads and localized fail-closed behavior must be live
  before operator recovery proof.
- Successful T01 unblocks VOC-113-T01; VOC-113-T02 remains under VOC-113 roster.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Do not use a snapshot-then-drift commit under this package directory as
  promotion evidence; package bookkeeping is not new unreviewed scope.

Tasks preserve scope, separation of duties, and rollback safety.
