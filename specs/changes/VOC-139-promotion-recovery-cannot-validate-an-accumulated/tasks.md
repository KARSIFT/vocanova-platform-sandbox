# VOC-139 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1096 is one
promotion-CI repair outcome: bind promotion `pr-validation` source hashes to
the reviewed head/source revision, keep ordinary merge-base anchoring, make
no-checkout recovery metadata repository-explicit, pin the new infrastructure
merge, update current-state docs, and let `reconcile-release` for #1089 merge
#1090. Repository count, hash-rule-versus-metadata, workflow-versus-tests-
versus-docs, and pin-versus-release are not split reasons.

Cross-repo note: coordinated PRs in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Infrastructure merge
`123735c80fec813a5b46a004f3e1122bd425cde2` is the defective live contract
for this failure class; T00 must open a new infra PR and pin its exact
merge. Do not treat the untracked local `karsift-ai-infra/` checkout (if
present) as this repo's tracked tree. No bootstrap exception. T00's first
run is attempt `1` on a new VOC-139 carrier from current `develop`. Do not
reuse PR #1090 as this package's implementation PR.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Promotion of already-completed
work through #1090 is evidence of this outcome after the repair lands. A
VOC-097 live-evidence second task is not used: an evidence-carrier PR cannot
satisfy sha_lineage against promotion PR #1090's recovery HEAD.

## VOC-139-T00 — Unblock accumulated promotion hash validation and no-checkout recovery metadata

- Requirement source: issue #1096; `VOC-139-D00` through `VOC-139-D17`
- Acceptance criteria: `VOC-139-AC-00` through `VOC-139-AC-09`
- Tests: `VOC-139-TEST-00` through `VOC-139-TEST-13`
- Evidence: `VOC-139-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1096 evidence in `t00-evidence.md` (PR #1090, head
   `4812fb91ab1b674f9a9ec03906f90c0edf50421d`, base
   `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`, PR run `33130426061`, job
   `98718413924`, tests `VOC-112-TEST-12` / `VOC-112-TEST-13`, merge-base
   hash fail `b0e65629…` versus `5ba216ff…`, release run `33130473438`,
   recovery run `33130527834`, job `98718739912`, `fatal: not a git
   repository`, pin `123735c80fec813a5b46a004f3e1122bd425cde2`). Do not copy
   raw provider responses, credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. Retain protected
   comparison anchor `b9e74fc2db4691c48c637639b265d527de9f4505` for the seven
   remaining VOC-112 no-change paths. This package's own plan/adoption/roster
   commits after issue-creation develop `4812fb91…` do not count as
   protected-file drift.
3. Open a new `KARSIFT/karsift-ai-infra` PR. Export an explicit promotion
   hash-anchor signal from `run-app-checks.sh` when `--promotion-pr` is set
   (prefer `VOC112_PROMOTION_PR=true`) while keeping mode `pr-validation` and
   exact PR SHAs. Make template `promotion-pr-metadata` repository-explicit
   and validate fields present in the live GitHub response rather than absent
   `.headRepository.nameWithOwner`. Do not fetch. Keep
   `--squash-safe-push` for generic non-PR events. Ordinary fixture
   add/modify/delete without the signal stays `pr-ancestry`. Ordinary
   `pr-validation` without the signal stays merge-base hash-anchored.
4. In the caller provenance test, implement the promotion-specific hash
   rule: when the promotion signal is present, stored hashes must equal
   `PR_HEAD_SHA` files and the working tree, not the merge-base files. Keep
   requiring valid exact SHAs and require the base to be an ancestor of the
   head. Keep ordinary
   `pr-validation` merge-base anchoring when the signal is absent.
5. Update live caller `.github/workflows/pipeline.yml` `promotion-pr-metadata`
   to the same repository-explicit, supported-field metadata lookup. Add
   deterministic tests for
   the #1090 accumulated-hash class, ordinary merge-base fail-closed, malformed
   SHA, unrelated commits/repository/PR, wrong refs, tampered current hashes,
   missing/nonancestor subject retention, base-not-ancestor rejection, and a
   subprocess execution of the metadata step with no git repository and a
   realistic GitHub response shape. Narrow live VOC-138 `NO_CHANGE_PATHS`
   so it no longer freezes the provenance test. Do not rewrite VOC-138
   package records.
6. Obtain independent exact-revision review of the infra PR and merge it.
   Record the exact merge SHA. From current caller `develop`, create a new
   VOC-139 implementation branch. Pin `PINNED_SHA.txt` to that infra merge
   and mirror every changed authoritative fixture file. Update fixture
   README, `docs/operations/11-devops-and-ci-cd.md`, and
   `docs/development/agent-skills.md`. Do not edit the seven remaining
   VOC-112 no-change paths.
7. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no capture-commit fetch helper, hydrate helper, recapture, or
   test-time evidence mutation was added.
8. Record in `t00-evidence.md` the implementation PR base, new infra merge,
   hash-rule change, metadata change, negative-case results, pin advance,
   validation commands after commit, and the feasible exact-head binding
   contract. Evidence must not require a commit to contain its own SHA.
9. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
   - `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`;
   - targeted VOC-139 hash-rule and no-checkout metadata cases;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates the accumulated-promotion head-hash case, ordinary
     merge-base retention, no-checkout metadata, identity negatives,
     no-fetch/no-recapture constraint, and seven-path freeze.
10. This package's caller PR `Closes` only its own VOC-139 task issue. Do
    not snapshot the current develop/main gap. After the exact reviewed
    caller merge, `reconcile-release` for #1089 may merge #1090 (or the live
    promotion at the then-current `develop` head) and converge `develop` to
    that exact merge SHA. Root issue #1096 closes only after allowlisted
    metadata from the successful recovery/release run exists.

11. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes.
- Fetching, hydrating, or recapturing VOC-112 JSON fixtures or hashed sources.
- Switching the promotion PR application check to `--squash-safe-push`.
- Weakening ordinary merge-base-anchored `pr-validation` or fixture-changing
  `pr-ancestry`.
- Changing any of the seven remaining VOC-112 no-change paths.
- Adding a capture-commit fetch helper, hydrate helper, provenance-mode
  skip, or test-time evidence mutation under any filename.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting hash-rule, metadata, pin, tests, docs, or release reconciliation
  into separate tasks.
- Rewriting VOC-138 package records.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the infra contract, caller pin, provenance-test hash rule,
  live pipeline.yml, tests, docs, evidence, and release handoff are one
  repair outcome.
- Infrastructure merge `123735c80…` is already merged and is the defective
  live contract; consume a new independently reviewed infra merge inside this
  same task.
- Merging #1090 is ordinary release after this correction, not a second
  VOC-139 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
