# VOC-144 — Roster reconciliation resolves carrier before pushed PR head metadata converges: Specification

## Objective and requirement source

Stop documented `reconcile` from treating immediate post-push GitHub REST
PR-head lag as a permanent `MISMATCHED_OPEN_CARRIER`. After `Push roster
branch`, carrier resolution must boundedly wait for the existing
same-repository, same-head-ref, same-base OPEN PR to expose the locally
pushed exact SHA, then reuse that carrier. Keep exact-SHA identity,
VOC-142 complete-required-set wait, production merge-guard, attestation, and
task-reuse fail-closed boundaries unchanged.

**Requirement source:** [GitHub issue #1122](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1122).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1122)

| Item | Value |
|------|-------|
| Repository | `KARSIFT/vocanova-platform-sandbox` |
| Merged plan PR | [#1110](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1110) |
| Generated roster PR | [#1112](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1112) |
| Deterministic branch | `karsift/roster-voc-141` |
| Issue-creation pin | VOC-142 merge `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Reconcile run 1 | `33437239322` |
| Pushed head 1 | `958e0fedf742173320bb89cfe690ec7070b49e93` |
| Run 1 failure | `Open roster PR` / `MISMATCHED_OPEN_CARRIER` |
| After run 1 | PR #1112 exposed `958e0fed…`; local resolution returned `reuse_open` |
| Reconcile run 2 | `33437514152` |
| Pushed head 2 | `0206cb70437cb751a19a2e715d8202b672060b50` |
| Run 2 failure | same `MISMATCHED_OPEN_CARRIER` class |
| After run 2 | PR #1112 exposed `0206cb70…` |

No roster PR was manually merged, closed, recreated, or bypassed. #1112
remains the single exact VOC-141 carrier. Each retry created a fresh roster
commit because the previous run failed after the push.

Live fixture contract at issue creation (tracked pin
`8993e867640dfb604dec0466c4e0787e68d8e258`):

- `Push roster branch` force-with-lease pushes the deterministic
  `karsift/roster-*` ref.
- `Open roster PR` immediately runs `roster-carrier-runner.py` with
  `head_sha=$(git rev-parse HEAD)`.
- The runner lists open and closed pulls once via `gh api` and calls
  `resolve_roster_carrier` on that single snapshot.
- `resolve_roster_carrier` reuses an OPEN PR only when repository, head
  ref, head SHA, and base ref all match. A same-ref OPEN PR whose listed
  SHA differs returns `MISMATCHED_OPEN_CARRIER`.
- That single-snapshot identity predicate is correct for a stable mismatch
  and must remain fail-closed. The adapter does not re-fetch, so transient
  REST lag is indistinguishable from a durable mismatch.

## Scope and non-goals

### In scope

1. After a roster-branch push, carrier resolution must boundedly poll the
   existing same-repository, same-head-ref, same-base OPEN PR until its
   listed head equals the locally pushed exact 40-character SHA.
2. The wait belongs in the GitHub adapter (`roster-carrier-runner.py` or an
   equivalent adapter helper it calls). `resolve_roster_carrier` remains a
   single-snapshot fail-closed identity predicate.
3. Wait only for the transient class: exactly one OPEN PR whose repository,
   head ref, and base ref match, and whose listed head SHA differs from the
   locally pushed SHA. First matching snapshot wins; do not wait the full
   bound after convergence.
4. Finite named timeout with a ceiling of 60 seconds and a named poll
   interval. Exact seconds within that ceiling are implementer-named
   constants. Tests inject sequenced snapshots or a fake clock; they must
   not depend on wall-clock GitHub lag or on VOC-141's 1,800-second recovery
   timeout.
5. Fail closed immediately (no SHA-lag wait) on ambiguous carriers,
   repository mismatch, base mismatch, invalid identity inputs, or API
   failure. After the bound, a still-different SHA on an otherwise matching
   OPEN PR remains `MISMATCHED_OPEN_CARRIER`.
6. Do not weaken exact-SHA identity, create a duplicate PR, retarget an
   unrelated PR, or manually merge #1112.
7. Preserve VOC-142 complete-required-set roster wait, already-merged reuse,
   create-when-zero-matches, existing-task reuse, and single root dispatch.
8. Deterministic live-shaped regressions for stale-then-converge,
   timeout-still-stale, durable-mismatch-does-not-wait, and the preserved
   VOC-142 carrier cases, plus caller pin/fixture mirror, tests, and
   current-state docs found by an exhaustive search. At minimum update
   `AGENTS.md`'s "Reconciling a merged plan PR" procedure, the fixture
   `README.md` roster-reuse paragraph, and the `adopt.yml` header comment.
9. After exact-SHA review and merge, ordinary `reconcile` for plan PR #1110
   must be able to resume #1112 when that PR still matches. Closure of
   #1122 binds allowlisted metadata from a successful adopt or reconcile
   run.

### Non-goals / explicitly excluded

- Weakening exact-SHA identity so a still-stale listed head is reused.
- Creating a second roster PR, a second VOC-141 task, a promotion PR, or a
  release audit.
- Manually merging, closing, recreating, or bypassing #1112.
- Weakening the production merge guard, adding bypass actors, or fabricating
  statuses.
- Changing VOC-142 complete-required-set wait, switching roster wait to
  `statusCheckRollup` or `gh pr checks`, or treating an in-progress parent
  as attestable SUCCESS for merge-gate or release.
- Unbounded sleep, a 1,800-second recovery-style timeout, or tests that
  sleep against live GitHub metadata.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- A VOC-097 operator-owned live-evidence second task.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Rewriting VOC-142, VOC-141, VOC-140, VOC-139, or earlier package records
  under `specs/changes/`.
- Fetching, hydrating, or recapturing VOC-112 JSON fixtures or hashed
  sources.
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
  reusable adoption carrier resolution; named current-state docs found by
  the required exhaustive search, including `AGENTS.md`.
- Protected technical effect: whether adopt may treat post-push REST PR-head
  lag as a durable mismatched carrier, and whether exact-SHA identity still
  gates reuse. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-test and adoption-merge-readiness
  changes still require exact-SHA independent verification and fail-closed
  controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-144-D00`: This is one outcome-sized adoption-recovery repair. Use one
end-to-end implementation task covering the coordinated infrastructure PR,
caller pin/fixture mirror, bounded post-push PR-head convergence wait,
unchanged single-snapshot identity predicate, fail-closed timeout and
durable-mismatch paths, tests, docs, evidence, and release handoff.
Repository count, adapter versus resolver, and tests-versus-docs are not
split reasons. Later promotion of this repair is evidence of the outcome,
not a second task and not a snapshot of the develop/main gap.

`VOC-144-D01`: After `Push roster branch`, `Open roster PR` must not treat
the first REST snapshot of an existing same-repository, same-head-ref,
same-base OPEN PR as a durable SHA mismatch. The class of reconcile runs
`33437239322` and `33437514152` cannot recur: if GitHub later exposes the
locally pushed exact SHA on that same OPEN PR, resolution must reuse it
(`reuse_open`) rather than fail `MISMATCHED_OPEN_CARRIER` and force the
next retry to push yet another head.

`VOC-144-D02`: The bounded wait lives in the GitHub adapter
(`roster-carrier-runner.py` or an equivalent helper it owns). 
`resolve_roster_carrier` remains a single-snapshot fail-closed identity
predicate: reuse only when repository, head ref, head SHA, and base ref all
match; a same-ref OPEN PR whose listed SHA differs is still
`MISMATCHED_OPEN_CARRIER` on that snapshot. `adopt.yml` may keep a single
runner invocation. Do not fold "stale SHA is good enough" into the
predicate.

`VOC-144-D03`: Wait only for the transient SHA-lag class: exactly one OPEN
PR whose `head.repo.full_name` and `base.repo.full_name` equal the caller
repository, whose head ref equals the deterministic `karsift/roster-*`
branch, whose base ref equals the integration branch, and whose listed head
SHA is a different valid 40-character SHA from the locally pushed HEAD.
Re-fetch until listed SHA equals local HEAD or the bound is exhausted. The
adapter may re-list pulls or GET the identified PR by number; each snapshot
must re-check full identity. Do not wait when there are zero same-ref OPEN
PRs (create path), when two or more OPEN matches exist, or when repository
or base ref already mismatch.

`VOC-144-D04`: Timeout and polling use named constants, not magic numbers.
The timeout ceiling is 60 seconds. Exact timeout and poll-interval seconds
within that ceiling are implementer-named constants recorded in evidence.
Do not wait the full bound after a matching snapshot. Do not use VOC-141's
1,800-second recovery timeout. Tests inject sequenced snapshots or a fake
clock; a test that sleeps against live GitHub metadata is not coverage of
#1122.

`VOC-144-D05`: Exact-SHA identity is not weakened. Convergence means the
listed PR head SHA equals local `git rev-parse HEAD` as a 40-character hex
SHA. A still-stale listed head must not be reused, merged, or used to
create a second PR. Do not close, recreate, or retarget #1112 or any
unrelated PR.

`VOC-144-D06`: Durable failures remain fail-closed and must not enter the
SHA-lag wait: `AMBIGUOUS_OPEN_CARRIER`, repository mismatch, base mismatch,
invalid repository/head-ref/head-SHA/base-ref, and GitHub API / parse
failure. After the bound, a still-different SHA on an otherwise matching
OPEN PR remains `MISMATCHED_OPEN_CARRIER`. Already-merged exact-identity
reuse, create-when-zero-matches, existing-task reuse, and single root
dispatch stay as VOC-142 defined them.

`VOC-144-D07`: Tests must exercise the live adapter path, not only the
snapshot predicate. Include at least: stale listed head then exact pushed
head → `reuse_open` (modeled on runs `33437239322` / `33437514152`);
timeout with still-stale head → `MISMATCHED_OPEN_CARRIER` and no `gh pr
create`; wrong-base / wrong-repo / ambiguous OPEN carriers fail without
waiting; exact first-snapshot match still reuses immediately; already-merged
exact reuse and zero-match create still work; API failure fails closed.
A unit test that only asserts `MISMATCHED_OPEN_CARRIER` on one stale
snapshot is not sufficient coverage of #1122.

`VOC-144-D08`: Pin advance. Issue-creation pin
`8993e867640dfb604dec0466c4e0787e68d8e258` is the defective live contract
for this failure class. T00 opens one new `KARSIFT/karsift-ai-infra` PR,
obtains independent exact-revision review, and after that merge sets
`PINNED_SHA.txt` and every changed mirrored fixture file to that exact merge.
Mirror at least `adopt.yml`, `roster-carrier-runner.py`, `roster_carrier.py`
if changed, any new wait helper, and their tests. If exact comparison proves
another authoritative fixture file also changed, mirror it too. Do not treat
the untracked local `karsift-ai-infra/` checkout as this repository's
tracked tree. Reconcile all live caller pin-lock tests. Preserve historical
`AUTHORITATIVE_PIN` / issue-era pin constants and package evidence,
including VOC-142 pin `8993e867…` as historical.

`VOC-144-D09`: No VOC-097 live-evidence second task, no snapshot-gap task,
and no manual merge of #1112. Roster PR #1112 remains the single exact
VOC-141 carrier. After this repair is live, ordinary
`gh workflow run pipeline.yml --ref develop -f action=reconcile -f plan_pr_number=1110`
must be able to reuse #1112 when it still matches, including across
post-push REST lag, wait for the complete required set, merge, clean up,
and dispatch the first not-yet-dispatched task exactly once. Do not
snapshot the develop/main gap (`karsift-ai-infra#15`). Do not create a
duplicate promotion PR or release audit. Live success after this repair is
ordinary adopt/reconcile/release closure evidence recorded with allowlisted
metadata only.

`VOC-144-D10`: Docs in the same PR. Before editing, exhaustively search
tracked source and current documentation for claims that carrier resolution
is a single immediate REST snapshot, that a same-ref SHA difference is
always a durable mismatch with no wait, and for the current pin
literal/hash assertions. Record the searched patterns and resulting path
disposition in `t00-evidence.md`. Update every current-state document that
would otherwise remain false, including `AGENTS.md`'s reconcile procedure,
the fixture `README.md` roster-reuse paragraph, and the `adopt.yml` header
comment. Current docs must state that reconcile waits boundedly for the
existing same-ref OPEN PR to expose the exact pushed SHA, then reuses that
carrier, and still fails closed on a stable mismatch. Preserve clearly
labeled historical VOC-142/VOC-141/VOC-140 records. Do not rewrite those
package directories.

`VOC-144-D11`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. No App ID/private-key
secret is rotated. Preserve the named run/PR/SHA IDs as audit evidence only
(no raw logs).

`VOC-144-D12`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- targeted VOC-144 stale-then-converge, timeout-still-stale,
  durable-mismatch-does-not-wait, and preserved VOC-142 carrier cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-144-D13`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate: bounded post-push SHA-lag wait; stale-then-converge
reuse; timeout-still-stale fail-closed; durable mismatch without wait;
unchanged exact-SHA identity; no duplicate PR; current-state docs; and pin
advance. Merge-gate must reject any mismatch. Committed `t00-evidence.md`
records the implementation PR base, new infra merge, wait change, and the
contract that later exact-head binding is published as review/check
metadata. A tracked file must not be required to contain the SHA of the
same commit that contains it.

`VOC-144-D14`: Protected comparison versus implementation PR base.
Implementation must resolve current `develop` to a 40-character SHA before
any in-scope edit and record that SHA as the implementation PR base. Fail
closed on unrelated/material movement of `develop` (any tree change outside
this package directory, in-scope fixture/pin/tests, and the named
current-state docs). This package's own plan/adoption/roster commits are
governance-only and do not count as protected-file drift. VOC-142, VOC-141,
VOC-140, and VOC-139 package files under `specs/changes/` are out of scope
and must not be edited.

`VOC-144-D15`: Release handoff. After the exact reviewed caller merge,
ordinary `reconcile` for plan PR #1110 may resume #1112. Ordinary later
promotion of this package uses `release.yml` at the then-current `develop`
tip once required checks pass. `develop` is advanced to the exact promotion
merge SHA before audit close. Every promotion merge push to `main` triggers
automatic production deployment, whose exact-SHA result is verified. Do not
snapshot the current gap. Closed state alone is not completion proof.
Preserve runs `33437239322` and `33437514152` and heads `958e0fed…` /
`0206cb70…` as audit evidence. Root issue #1122 closes only after
allowlisted metadata from a successful adopt/reconcile run exists. Do not
create a duplicate promotion PR or release audit.

`VOC-144-D16`: This package's own first native adoption still executes the
defective live `adopt.yml` until T00's independently reviewed infra merge is
pinned. That sequencing tension is not a bootstrap exception, not authority
to manually merge any roster PR, and not authority to self-adopt. Prefer
native adoption succeeding when GitHub has already converged before
resolution. Record remaining bootstrap tension in `t00-evidence.md` if this
package's own adoption hits the same lag; do not weaken gates to avoid it.

`VOC-144-D17`: VOC-142 complete-required-set roster wait remains required.
This package does not change `Wait for roster PR checks`,
`--roster-pr-gate`, or in-progress-parent attestation. After a converged
carrier is reused, wait still requires registered SUCCESS for `ci / ci`,
`governance-policy`, and `validate` on the exact head.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents an
adoption deadlock without letting carrier resolution reuse a still-stale
SHA, without creating duplicate roster carriers, and without skipping
protected checks.

Abuse/process risks:

1. Treating the first post-push REST snapshot of a same-ref OPEN PR as a
   durable SHA mismatch — forbidden by `VOC-144-D01`.
2. Reusing a still-stale listed head, or treating SHA equality as optional
   — forbidden by `VOC-144-D05`.
3. Folding "stale SHA is good enough" into `resolve_roster_carrier` —
   forbidden by `VOC-144-D02`.
4. Waiting on ambiguous, wrong-repo, or wrong-base carriers, or on API
   failure — forbidden by `VOC-144-D03` and `VOC-144-D06`.
5. Unbounded sleep or a 1,800-second recovery-style timeout — forbidden by
   `VOC-144-D04`.
6. Creating a second roster PR when #1112 already exists — forbidden by
   `VOC-144-D05` and `VOC-144-D09`.
7. Manually merging, closing, recreating, or bypassing #1112 — forbidden by
   `VOC-144-D09` and `VOC-144-DEP-04`.
8. Weakening VOC-142 complete-required-set wait, the production merge guard,
   or in-progress-parent attestation — forbidden by `VOC-144-D17` and
   `VOC-144-DEP-04`.
9. Snapshotting the develop/main gap or adding a self-invalidating
   snapshot-then-check task — forbidden by `VOC-144-D09` and
   `karsift-ai-infra#15`.
10. Changing `roles.yml` or adding an OpenAI route — forbidden by
    `VOC-144-D11`.
11. Printing credentials or copying full CI logs into evidence — forbidden.
12. Requiring a commit to contain its own SHA — forbidden by `VOC-144-D13`.
13. Using a bootstrap exception to land this repair outside adoption —
    forbidden by `VOC-144-D16`.

## Contradictions and open questions

1. **VOC-142 reuse promise versus post-push REST lag:** `adopt.yml` and
   `AGENTS.md` state that reconcile reuses a matching open roster PR for
   the deterministic branch/base/exact head. Live resolution compares the
   listed PR head to local HEAD on the first snapshot, so a matching PR
   whose GitHub metadata has not yet caught up fails closed. This package
   follows D01–D03 and treats that gap as in-scope residual work, not as a
   rewrite of VOC-142 package records.
2. **Single-snapshot `MISMATCHED_OPEN_CARRIER` tests remain correct:**
   VOC-142 tests that a stale SHA on one snapshot is a mismatch must keep
   passing for `resolve_roster_carrier`. The new adapter wait is what
   distinguishes lag from a stable mismatch. Do not delete those
   single-snapshot assertions.
3. **Exact timeout seconds within the 60-second ceiling:** issue #1122
   requires a finite bound and observed that lag cleared by the time a
   failed job could be re-queried. The precise named timeout and poll
   interval inside that ceiling are an implementer choice for
   adoption-time review if a shorter or longer bound is preferred; they
   must remain named, finite, and test-injected.
4. **VOC-141 carrier #1112 is still OPEN:** this package must not merge,
   close, or recreate it. After T00, documented reconcile of #1110 is the
   recovery. Do not rewrite VOC-141 or VOC-142 package files.
5. **This package's own first adoption uses defective `adopt.yml`:** see
   D16. Do not invent a founder-comment gate, a manual-merge exception, or a
   snapshot-gap task to paper over that sequencing. New infrastructure merge
   SHA is not available at drafting time; record it after the coordinated
   infra PR merges.
6. **Untracked local `karsift-ai-infra/` checkout:** if present in the
   workspace, it is not this repository's tracked tree and is not
   implementation evidence.
