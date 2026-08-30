# VOC-127 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1035 is one
release-convergence outcome: exact-merge-SHA develop sync after promotion plus
an exceptional adopted path for main-only reconciliation. Coordinated caller
and infrastructure pull requests remain one task; repository count,
workflow-versus-tests-versus-docs, and fixture/pin work are not split reasons.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for `release.yml`,
helpers, source tests, current-state README/comments, and the project-repo
pipeline template if the exceptional action is added there, and changes this
caller's `.github/workflows/pipeline.yml`, staging path-selection proof or
narrow skip, docs, fixture/pin, and tests. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree. Infra
PRs must say `Relates to OWNER/CALLER#<task>` and MUST NOT use a closing
keyword. No bootstrap exception: VOC-124 already enables nested workflow-file
publication; T00's first run is attempt `1`.

This is not a "promote current develop to main" package. Do not add a task
that snapshots the current (already repaired) commit gap as a later drift
gate.

## VOC-127-T00 — Advance develop to the exact promotion merge SHA and add exceptional main-only reconciliation

- Requirement source: issue #1035; `VOC-127-D00` through `VOC-127-D09`
- Acceptance criteria: `VOC-127-AC-00` through `VOC-127-AC-06`
- Tests: `VOC-127-TEST-00` through `VOC-127-TEST-11`
- Evidence: `VOC-127-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1035 pre-repair evidence in `t00-evidence.md` (`develop`
   `883c4a544f24fb9840694a834ebe9c665d4160b5`, `main` / PR #1033 merge
   `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`, `0 46` ahead/behind, ancestor
   proof, VOC-113-T01 PR #955 tree diff, `release.yml` restore of
   `CHECKED_HEAD_SHA`, post-repair equal refs, #1032 audit). State that the
   2026-08-27 settings repair is historical evidence, not implementation
   authority, and that this task does not snapshot that gap.
2. In `KARSIFT/karsift-ai-infra/.github/workflows/release.yml`, after the
   single exact-head `--merge` (and when the same PR is already `MERGED`),
   resolve `mergeCommit.oid`, bind checked head, checked production base,
   merged PR identity, and live tips (`VOC-127-D02`), then advance or recreate
   `develop` at that exact SHA before closing the release audit. Preserve the
   promotion PR and merge commit. Stop treating `ahead_by == 0` with unequal
   SHAs as fully promoted. Recreate auto-deleted integration refs at the merge
   SHA, not `CHECKED_HEAD_SHA`. Fail closed on races, malformed merges, moved
   `main`, or unique develop commits; never erase concurrent integration work.
3. Extract bind-and-sync into a testable helper under
   `karsift-ai-infra/config/` if that is the cleanest way to cover the
   `VOC-127-D08` matrix; inline shell is acceptable only if the same
   fail-closed cases are still deterministically tested. Exact helper name is
   `VOC-127-DEP-07`.
4. Keep one `gh pr merge` call. After a bound merge exists, later converge
   wakes must not open another promotion PR or merge again (`VOC-127-D04`).
   `reconcile-release` must retry bind-and-sync when the PR is merged and the
   audit is still open or `develop` is not yet at the merge SHA.
5. Add the exceptional adopted main-only reconciliation action on the mutating
   caller `pipeline.yml` surface (live and project-repo template), not on
   `pipeline-verify.yml`. Identity is an adopted package/task plus a merged
   main-targeting PR number. Do not add operator SHA inputs. Do not overload
   `existing_pr_number`. Keep every `workflow_dispatch` block at most 25
   inputs. Do not schedule this action against arbitrary `main` deltas.
6. Prove that a tree-equivalent `develop` synchronization does not schedule an
   unnecessary staging deployment. Retain VOC-111 path selection for actual
   runtime/deploy tree changes. If `on.push.paths` is insufficient for
   merge-commit fast-forwards, add a fail-closed job-level skip in
   `deploy-staging.yml` without broadening the allowlist. Extend
   `scripts/foundation/voc111-deploy-staging-paths.test.mjs` (or an adjacent
   foundation test in the same task) accordingly.
7. Preserve VOC-113 through VOC-126 contracts: roster markers, exact-head
   checks, recovery, App-token isolation, match-head-commit, two-attempt
   bound, Cursor Composer implementer, Cursor Grok review, no OpenAI,
   unchanged `config/roles.yml`. Never bundle secrets. Never print credential
   values.
8. Update current-state comments/docs: `release.yml`, pipeline template/live
   comments as needed, `karsift-ai-infra/README.md`, `AGENTS.md` reconcile-release
   and release-authority text, `docs/operations/11-devops-and-ci-cd.md`,
   `docs/operations/10-development-workflow.md`, and current-state
   branch/release paragraphs in
   `docs/operations/15-ai-native-product-and-engineering-operating-model.md`
   and `docs/governance/16-autonomous-development-operating-model.md` so they
   describe exact-merge-SHA develop sync, `reconcile-release` retry of that
   sync, and exceptional main-only reconciliation. Do not rewrite historical
   CHANGELOG or A-003/VOC-075 audit records. Update any other live contract
   that would become false.
9. Land the infrastructure change through the normal coordinated source
   carrier. Merge the independently reviewed infra PR first so
   `release.yml@main` performs exact-merge-SHA sync before the caller consumes
   it. Pin `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge
   SHA when consumed — not to `60afda3a44fd06b8c00b219771de7112f1aded6e`.
   Update caller fixture regressions and any `scripts/foundation/*` pin
   literals that still assert the previous pin, in the same task.
10. After that exact reviewed infra merge is live, update the live caller
    dispatch, staging selector if required, docs, tests, and evidence. This
    package's caller PR `Closes` only its own VOC-127 task issue.
11. Run applicable validation and record results in `t00-evidence.md`:
    - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
      `KARSIFT/karsift-ai-infra` checkout;
    - `bash scripts/governance/validate-governance.sh`;
    - `bash scripts/governance/classify-change-risk.sh`;
    - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
    - targeted foundation tests if those files change (including VOC-111 path
      selection);
    - `git diff --check`;
    - exact reviewed infra SHA and pin applicability;
    - any narrower targeted commands added by the implementation.
12. Preserve independent exact-SHA review for each carrier, risk
    classification, protected checks, and App-token isolation. Do not weaken
    review, risk classification, required checks, or automatic-merge gates.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes, except
  the staging skip for tree-equivalent develop-sync.
- Snapshotting the current develop/main gap, re-running the 2026-08-27
  settings repair, or promoting "current develop" to `main` as this package's
  work.
- Fast-forwarding `main` instead of preserving `--merge` commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Implementing VOC-122, merging unrelated carriers, or rewriting historical
  CHANGELOG / A-003 / VOC-075 records.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: ordinary post-promotion sync, exceptional main-only
  reconciliation, staging path-selection proof, tests, docs, and caller pin
  are one release-convergence outcome.
- Infra should merge first so `release.yml@main` syncs `develop` to the merge
  SHA before the caller fixture consumes that revision.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
