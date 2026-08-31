# VOC-142 — Adoption roster wait ignores pending required CI and reconcile cannot reuse the open roster PR: Specification

## Objective and requirement source

Stop native adoption from attempting a protected roster merge while
ruleset-required `ci / ci` is still unregistered or IN_PROGRESS, and stop
documented `reconcile` from failing because `Open roster PR` always creates a
new PR when the exact open roster carrier already exists. Keep production
merge-guard, attestation, and task-reuse fail-closed boundaries unchanged.

**Requirement source:** [GitHub issue #1113](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1113).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1113)

| Item | Value |
|------|-------|
| Repository | `KARSIFT/vocanova-platform-sandbox` |
| Source issue | [#1109](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1109) |
| Merged plan PR | [#1110](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1110) |
| Reviewed plan head | `178ae81ed4f0224fbb12359c90ddb67c6687ca9a` |
| Plan merge | `bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2` |
| Adoption run | `33343125733`, job `99342230038` |
| Generated task | [#1111](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1111) |
| Generated roster PR | [#1112](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1112) |
| Exact roster head | `98dd0936a73b64a6b548da6cf2000a6d000917ac` |
| Roster pipeline run | `33343147453` |
| Required CI job | `99342299218` |
| Wait started | `2026-08-30T23:57:08Z` |
| Wait reported SUCCESS | `2026-08-30T23:58:46Z` |
| Merge started | `2026-08-30T23:58:46Z` |
| Merge failed | `2026-08-30T23:58:47Z` |
| `ci / ci` started | `2026-08-30T23:57:21Z` |
| `ci / ci` SUCCESS | `2026-08-30T23:59:16Z` |
| Reconcile run | `33343250178`, job `99342577393` |
| Reconcile failure step | `Open roster PR` |

At the failed merge instant, #1112 was OPEN/MERGEABLE on the exact head but
its required `ci / ci` row was still IN_PROGRESS. Afterward, all three
required rows (`ci / ci`, `governance-policy`, and `validate`) were green.
Reconcile reused task #1111 and the deterministic roster branch, then failed
because #1112 already exists. No duplicate task or PR was created.

Live fixture contract at issue creation (tracked pin
`67bdfd13ef875dead23ce4be01d7d0e8b976e289`):

- `Wait for roster PR checks` breaks when `check_count > 0`, `pending == 0`,
  `failed == 0`, and that count plus head SHA are unchanged across two
  consecutive snapshots (`stable_green_count >= 2`).
- That wait invokes `authoritative-checks-runner.py` with
  `--workflow-event "pull_request"` and without `--ordinary-pr-gate`.
- `_workflow_runs` skips GitHub Actions check-runs whose parent fails
  `parent_run_is_attestable`. `parent_run_is_attestable` is false while the
  parent is not terminal, so an IN_PROGRESS `ci / ci` row is omitted rather
  than counted pending.
- `--ordinary-pr-gate` would not cover this class even if added unchanged:
  `_ordinary_pr_pipeline_parent` requires `head_ref` to start with `agent/`
  or `plan/`, and roster branches are `karsift/roster-*`.
- `Open roster PR` always runs `gh pr create` when
  `steps.commit.outputs.changed == 'true'`. Existing-open lookup exists for
  other workflows (`release.yml`, `implement.yml`) but not here.
- `Recover cleanup for an already-merged roster` runs only when
  `changed == 'false'`. Interrupted merge leaves `changed == 'true'` on
  reconcile because `develop` still lacks the unmerged roster files.

## Scope and non-goals

### In scope

1. Roster adoption must not attempt merge until the complete
   ruleset-required check set for the exact roster head is registered and
   the newest logical attempt of every required row, including `ci / ci`, is
   SUCCESS.
2. Registration stability must be fail-closed. A partial green snapshot must
   not count as complete merely because its current count is stable.
3. Reconcile must find and validate an existing open roster PR for the
   deterministic branch, base, and exact head SHA, then resume
   waiting/merge/cleanup from that carrier.
4. Reconcile must create a PR only when no matching open or already-merged
   carrier exists, and must reject ambiguous or mismatched carriers.
5. Existing task issues must be reused. Implementation must dispatch exactly
   once after the checked roster merges.
6. Deterministic live-shaped regressions for late registration or continued
   execution of required `ci / ci`, exact open-PR reuse, already-merged
   reuse, mismatched/ambiguous rejection, and duplicate task/dispatch
   suppression, plus caller pin/fixture mirror, tests, and current-state docs
   found by an exhaustive search. At minimum reconcile `AGENTS.md`'s
   "Reconciling a merged plan PR" procedure, the fixture `README.md` roster
   wait paragraph, and the `adopt.yml` header comment that already promises
   reuse.
7. After exact-SHA review and merge, ordinary `reconcile` for plan PR #1110
   must be able to resume #1112 when that PR still matches. Closure of #1113
   binds allowlisted metadata from a successful adopt or reconcile run.

### Non-goals / explicitly excluded

- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging a roster PR (including #1112).
- Switching the roster wait to `statusCheckRollup` or `gh pr checks`.
- Treating an in-progress parent workflow as attestable SUCCESS for
  merge-gate or release.
- Creating a duplicate VOC-141 task, roster PR, promotion PR, or release
  audit issue.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- A VOC-097 operator-owned live-evidence second task.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Rewriting VOC-141, VOC-140, VOC-139, or earlier package records under
  `specs/changes/`.
- Fetching, hydrating, or recapturing VOC-112 JSON fixtures or hashed sources.
- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Repairing VOC-141's promotion-recovery hang (issue #1109 / VOC-141). That
  is a separate package; this package only unblocks its interrupted
  adoption carrier after T00 is live.
- A supervised bootstrap exception for this package's own first adoption.
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable adoption wait/reuse; named current-state docs found by the
  required exhaustive search, including `AGENTS.md`.
- Protected technical effect: whether adoption may merge a roster PR before
  the complete required check set is SUCCESS, and whether reconcile reuses
  an existing exact carrier. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-test and adoption-merge-readiness
  changes still require exact-SHA independent verification and fail-closed
  controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-142-D00`: This is one outcome-sized adoption-recovery repair. Use one
end-to-end implementation task covering the coordinated infrastructure PR,
caller pin/fixture mirror, complete-required-set wait, open and
already-merged carrier reuse, mismatched/ambiguous rejection, tests, docs,
evidence, and release handoff. Repository count, wait versus reuse, and
tests-versus-docs are not split reasons. Later promotion of this repair is
evidence of the outcome, not a second task and not a snapshot of the
develop/main gap.

`VOC-142-D01`: Roster adoption must not attempt merge until the complete
GitHub-ruleset-required logical-context set for the exact roster head is
registered and the newest logical attempt of every required row is SUCCESS.
At issue creation that required set is `ci / ci`, `governance-policy`, and
`validate`. The class of run `33343125733` / job `99342230038` cannot recur:
wait must not report complete while required `ci / ci` is IN_PROGRESS or
absent.

`VOC-142-D02`: Registration stability is fail-closed. Two consecutive
snapshots with `pending == 0` and an unchanged `total_count` are not
sufficient when the snapshot is missing a required logical context. A
partial green set must not count as complete merely because its current
count is stable. Fail closed if the required set cannot be uniquely
resolved. Do not infer completeness from "any non-empty SUCCESS-only
snapshot."

`VOC-142-D03`: The wait-completeness predicate is distinct from attestation.
IN_PROGRESS or queued required rows are not-ready, not SUCCESS. Do not
change merge-gate or release so that an in-progress parent becomes
attestable completed `ci / ci`. Do not switch the roster wait to
`statusCheckRollup` or `gh pr checks`. Keep paginated check-runs/statuses,
exact PR base/head binding, and newest-logical-attempt selection. If
`--ordinary-pr-gate` is reused, it must cover `karsift/roster-*` heads or
the wait must use a roster-specific completeness path; the live helper
rejects those refs today.

`VOC-142-D04`: When `steps.commit.outputs.changed == 'true'`, `Open roster
PR` must first resolve existing carriers for the deterministic
`karsift/roster-<change-id>` head, the integration-branch base, and the
exact pushed head SHA. Exactly one matching OPEN PR is reused as
`pr_number`; wait/merge/cleanup resume from that carrier. The class of
reconcile run `33343250178` / job `99342577393` cannot recur.

`VOC-142-D05`: Create a roster PR only when no matching OPEN or
already-merged carrier exists. Exactly one already-merged PR with the same
repository, head ref, exact head SHA, and base ref is a matching
already-merged carrier: do not create another PR; continue into the existing
merged-roster cleanup/dispatch path as applicable. Zero matches may create
exactly one new PR. Two or more matches, or a single carrier whose
repository, base, head ref, or head SHA does not match the pushed identity,
must fail closed. Do not close, recreate, or retarget an unrelated PR.

`VOC-142-D06`: Existing task issues remain reused (`gh issue list` / live
equivalent already in `adopt.yml`). After the checked roster merges,
`Determine whether the root task needs dispatch` still dispatches only when
the first task issue is OPEN and no implementation PR exists for that
deterministic head. An interrupted adoption that later reconciles through
the reused roster PR must not open a second task or dispatch a second root
implementation. Task #1111 must remain the VOC-141 task; this package must
not create a replacement for it.

`VOC-142-D07`: Tests must exercise the live wait and open-PR paths, not only
helpers that omit required `ci / ci` or assume `gh pr create` always
succeeds. Include at least: late registration of required `ci / ci`;
continued IN_PROGRESS `ci / ci` while other required rows are already
SUCCESS; existing exact open roster PR reuse; already-merged roster reuse;
mismatched and ambiguous carrier rejection; and duplicate task/dispatch
suppression. A unit test that only asserts `stable_green_count` exists in
YAML is not sufficient coverage of #1113.

`VOC-142-D08`: Pin advance. Issue-creation pin
`67bdfd13ef875dead23ce4be01d7d0e8b976e289` is the defective live contract
for this failure class. T00 opens one new `KARSIFT/karsift-ai-infra` PR,
obtains independent exact-revision review, and after that merge sets
`PINNED_SHA.txt` and every changed mirrored fixture file to that exact merge.
Mirror at least `adopt.yml`, authoritative-check sources used by the wait,
and their tests. If exact comparison proves another authoritative fixture
file also changed, mirror it too. Do not treat the untracked local
`karsift-ai-infra/` checkout as this repository's tracked tree. Reconcile
all live caller pin-lock tests. Preserve historical `AUTHORITATIVE_PIN` /
issue-era pin constants and package evidence, including VOC-140 pin
`67bdfd13…` as historical.

`VOC-142-D09`: No VOC-097 live-evidence second task, no snapshot-gap task,
and no manual merge of #1112. Roster PR #1112 remains the single exact
VOC-141 carrier. After this repair is live, ordinary
`gh workflow run pipeline.yml --ref develop -f action=reconcile -f plan_pr_number=1110`
must be able to reuse #1112 when it still matches, wait for the complete
required set, merge, clean up, and dispatch the first not-yet-dispatched
task exactly once. Do not snapshot the develop/main gap
(`karsift-ai-infra#15`). Do not create a duplicate promotion PR or release
audit. Live success after this repair is ordinary adopt/reconcile/release
closure evidence recorded with allowlisted metadata only.

`VOC-142-D10`: Docs in the same PR. Before editing, exhaustively search
tracked source and current documentation for claims that two stable
zero-pending snapshots complete roster wait, that reconcile always opens a
new roster PR, and for the current pin literal/hash assertions. Record the
searched patterns and resulting path disposition in `t00-evidence.md`.
Update every current-state document that would otherwise remain false,
including `AGENTS.md`'s reconcile procedure, the fixture `README.md` roster
wait paragraph, and the `adopt.yml` header comment. Current docs must state
that roster wait requires the complete required set including `ci / ci`, and
that reconcile reuses a matching open or already-merged carrier. Preserve
clearly labeled historical VOC-141/VOC-140 records. Do not rewrite those
package directories.

`VOC-142-D11`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. No App ID/private-key
secret is rotated. Preserve the named run/job/PR/SHA IDs as audit evidence
only (no raw logs).

`VOC-142-D12`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- targeted VOC-142 wait-completeness, open-PR reuse, already-merged reuse,
  mismatch/ambiguous rejection, and duplicate-suppression cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-142-D13`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate: complete-required-set wait including `ci / ci`;
fail-closed partial green snapshots; exact open-PR reuse; already-merged
reuse; mismatched/ambiguous rejection; duplicate task/dispatch suppression;
current-state docs; and pin advance. Merge-gate must reject any mismatch.
Committed `t00-evidence.md` records the implementation PR base, new infra
merge, wait change, reuse change, and the contract that later exact-head
binding is published as review/check metadata. A tracked file must not be
required to contain the SHA of the same commit that contains it.

`VOC-142-D14`: Protected comparison versus implementation PR base.
Issue-creation plan merge is `bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2`.
Implementation must resolve current `develop` to a 40-character SHA before
any in-scope edit and record that SHA as the implementation PR base. Fail
closed on unrelated/material movement of `develop` (any tree change outside
this package directory, in-scope fixture/pin/tests, and the named
current-state docs). This package's own plan/adoption/roster commits after
`bb4ffdf…` are governance-only and do not count as protected-file drift.
VOC-141 package files under `specs/changes/VOC-141-…/` are out of scope and
must not be edited.

`VOC-142-D15`: Release handoff. After the exact reviewed caller merge,
ordinary `reconcile` for plan PR #1110 may resume #1112. Ordinary later
promotion of this package uses `release.yml` at the then-current `develop`
tip once required checks pass. `develop` is advanced to the exact promotion
merge SHA before audit close. Every promotion merge push to `main` triggers
automatic production deployment, whose exact-SHA result is verified. Do not
snapshot the current gap. Closed state alone is not completion proof.
Preserve runs `33343125733`, `33343147453`, `33343250178` and jobs
`99342230038`, `99342299218`, `99342577393` as audit evidence. Root issue
#1113 closes only after allowlisted metadata from a successful
adopt/reconcile run exists. Do not create a duplicate promotion PR or
release audit.

`VOC-142-D16`: This package's own first native adoption still executes the
defective live `adopt.yml` until T00's independently reviewed infra merge is
pinned. That sequencing tension is not a bootstrap exception, not authority
to manually merge any roster PR, and not authority to self-adopt. Prefer
native adoption succeeding when the complete required set is already SUCCESS
before wait completes. Record remaining bootstrap tension in
`t00-evidence.md` if this package's own adoption hits the same race; do not
weaken gates to avoid it.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents an
adoption deadlock without letting wait treat in-progress `ci / ci` as
complete, without creating duplicate roster carriers, and without skipping
protected checks.

Abuse/process risks:

1. Treating two stable zero-pending snapshots as complete while required
   `ci / ci` is missing or IN_PROGRESS — forbidden by `VOC-142-D01` and
   `VOC-142-D02`.
2. Switching the wait to `statusCheckRollup` or `gh pr checks` — forbidden
   by `VOC-142-D03`.
3. Treating an in-progress parent as attestable SUCCESS for merge-gate or
   release — forbidden by `VOC-142-D03`.
4. Calling `gh pr create` when the exact open roster PR already exists —
   forbidden by `VOC-142-D04`.
5. Creating a second PR when an already-merged exact carrier exists, or
   accepting a mismatched/ambiguous carrier — forbidden by `VOC-142-D05`.
6. Opening a second task issue or dispatching the root task twice —
   forbidden by `VOC-142-D06`.
7. Manually merging, closing, recreating, or bypassing #1112 — forbidden by
   `VOC-142-D09` and `VOC-142-DEP-04`.
8. Weakening the production merge guard or adding bypass actors — forbidden
   by `VOC-142-DEP-04`.
9. Snapshotting the develop/main gap or adding a self-invalidating
   snapshot-then-check task — forbidden by `VOC-142-D09` and
   `karsift-ai-infra#15`.
10. Changing `roles.yml` or adding an OpenAI route — forbidden by
    `VOC-142-D11`.
11. Printing credentials or copying full CI logs into evidence — forbidden.
12. Requiring a commit to contain its own SHA — forbidden by `VOC-142-D13`.
13. Using a bootstrap exception to land this repair outside adoption —
    forbidden by `VOC-142-D16`.

## Contradictions and open questions

1. **Header promise versus live `Open roster PR`:** `adopt.yml` states that
   re-running with the same merged PR reuses existing artifacts. Live
   `Open roster PR` still always creates when `changed == 'true'`. This
   package follows D04/D05 and treats that gap as in-scope residual work,
   not as a rewrite of historical adoption records.
2. **Wait comment versus live completeness:** the wait comment claims
   stabilizing a complete green logical-name count prevents a freshly opened
   PR from passing while checks are still registering. Live completeness is
   "whatever is currently visible and stable," which dropped IN_PROGRESS
   `ci / ci`. This package requires the complete required set, not a stable
   subset.
3. **`--ordinary-pr-gate` does not cover roster refs:** reusing that flag
   unchanged would still reject `karsift/roster-*`. Implementation must not
   treat adding the flag without a roster-head rule as covering #1113.
4. **VOC-141 carrier #1112 is still OPEN:** this package must not merge,
   close, or recreate it. After T00, documented reconcile of #1110 is the
   recovery. Do not rewrite VOC-141 package files.
5. **This package's own first adoption uses defective `adopt.yml`:** see
   D16. Do not invent a founder-comment gate, a manual-merge exception, or a
   snapshot-gap task to paper over that sequencing. New infrastructure merge
   SHA is not available at drafting time; record it after the coordinated
   infra PR merges.
6. **Untracked local `karsift-ai-infra/` checkout:** if present in the
   workspace, it is not this repository's tracked tree and is not
   implementation evidence.
