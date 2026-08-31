# VOC-143 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1120 is one
promotion-provenance repair outcome: bind `squash-safe-push` and promotion
`pr-validation` `AGENTS.md` hashes to an immutable historical ancestor, keep
other modes fail-closed, leave fixtures and promotion check identity
unchanged, and let ordinary `reconcile-release` recover #1118. Validate versus
`ci / ci`, tests versus docs, and post-merge promotion are not split reasons.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Interrupted VOC-142 promotion
via #1119 / #1118 is incident evidence that this repair must unblock; it is
not a second VOC-143 task. A VOC-097 live-evidence second task is not used.
No coordinated infrastructure PR is required; live pin
`8993e867640dfb604dec0466c4e0787e68d8e258` is not this defect.

## VOC-143-T00 — Bind promotion-path VOC-112 `AGENTS.md` provenance to an immutable historical ancestor

- Requirement source: issue #1120; `VOC-143-D00` through `VOC-143-D14`
- Acceptance criteria: `VOC-143-AC-00` through `VOC-143-AC-08`
- Tests: `VOC-143-TEST-00` through `VOC-143-TEST-08`
- Post-T00 release/closure verification: `VOC-143-TEST-09` (not an
  implementation-task close or retry prerequisite)
- Evidence: `VOC-143-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1120 evidence in `t00-evidence.md` (promotion PR #1119,
   exact head `376e00dd769253d7a255660f5391fb208781e2f3`, failing required
   `validate` and `ci / ci`, open release audit #1118, VOC-142-DEP-07
   no-recapture constraint, live pin
   `8993e867640dfb604dec0466c4e0787e68d8e258`). Do not copy raw provider
   responses, credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. This package's own
   plan/adoption/roster commits after issue-creation promotion head
   `376e00dd…` do not count as protected-file drift. Do not edit
   `specs/changes/VOC-142-…/` or earlier package directories.
3. In `scripts/foundation/voc112-navigation-benchmark.test.mjs`, implement
   D01–D04:
   - `squash-safe-push` must not require fixture `agents_sha256` to equal
     current working-tree `AGENTS.md`. Bind that hash to an ancestor of
     `PR_HEAD_SHA` (when it is a 40-character SHA) or `HEAD` via
     `resolveFixtureAgentsBase` or equivalent. Fail closed if the hash is
     not found.
   - Promotion `pr-validation` must bind fixture `agents_sha256` to that
     same historical ancestor of `PR_HEAD_SHA`, not to current HEAD
     `AGENTS.md`. Keep navigator hashes HEAD-bound on promotion PRs.
   - Keep `local` and `pr-ancestry` working-tree `AGENTS.md` equality.
   - Keep ordinary `pr-validation` merge-base anchoring for both hashed
     sources.
4. Add deterministic tests that call `assertCapturedRevision`. Cover
   TEST-00 through TEST-07. Update VOC-139 expected error strings only if
   D03 causes the navigator assertion to fire first. Do not treat a comment
   change as covering #1120.
5. Exhaustively search tracked source/docs; update every current-state match
   found by D08. Do not rewrite VOC-142/VOC-139/VOC-114 package records. Do
   not recapture VOC-112 JSON fixtures. Do not change `PINNED_SHA.txt`,
   `config/roles.yml`, promotion mode selection in
   `repository-governance.yml` or `ci.yml`, or `package.json`. Prefer not to
   edit `repository-governance.yml` at all.
6. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no fabricated-status helper, bypass-actor addition, App-token mint
   change, fetch/hydrate helper, or #1119 merge/close/recreate helper was
   added.
7. Record in `t00-evidence.md` the implementation PR base, provenance-assertion
   change, negative-case results, source-search disposition, validation
   commands after commit, and the feasible exact-head binding contract.
   Evidence must not require a commit to contain its own SHA.
8. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head;
   - `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`
     under the modes in D11;
   - `node --test scripts/foundation/voc114-actions-check-recovery.test.mjs`;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates D01–D08.
9. This package's caller PR `Closes` only its own VOC-143 task issue. Do
   not snapshot the current develop/main gap. After the exact reviewed
   caller merge, ordinary `reconcile-release` for #1118 may re-evaluate
   #1119 when it still matches. Root issue #1120 closes only after
   allowlisted metadata from a successful recovery/release run exists. Do
   not create a duplicate promotion PR or release audit.

10. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  `roles.yml`, or monitor-inventory changes.
- Recapturing or editing VOC-112 JSON fixtures or hashed sources.
- Switching promotion `ci.yml` to `--squash-safe-push`, or switching
  promotion `validate` away from `squash-safe-push`.
- Weakening `local`, `pr-ancestry`, ordinary merge-base `pr-validation`, or
  navigator HEAD/working-tree binding.
- Pin-advance or a coordinated `karsift-ai-infra` PR unless an infra contract
  change is proven necessary.
- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging #1119.
- Creating a duplicate promotion PR or release audit.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- OpenAI credentials or execution routes.
- Weakening exact-SHA review, risk floors, protected checks, or retry caps.
- Splitting validate, `ci / ci`, tests, docs, or release reconciliation into
  separate tasks.
- Rewriting VOC-142, VOC-141, VOC-140, VOC-139, VOC-138, VOC-114, VOC-113, or
  VOC-112 package records.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the provenance assertion, both promotion required checks,
  tests, docs, evidence, and release handoff are one repair outcome.
- Live pin `8993e867…` is already merged and is not this defect; do not
  consume a new infra merge inside this task.
- Interrupted VOC-142 promotion (#1119 / #1118 / head `376e00dd…`) is
  incident proof, not a second VOC-143 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
