# VOC-141 — Promotion recovery waits 30 minutes when green CI has an unattestable parent: Specification

## Objective and requirement source

Stop promotion recovery from polling for 1,800 seconds with no possible state
transition when GitHub's required `ci / ci` row is already SUCCESS but the
authoritative composed evidence has no uniquely attestable completed
non-carrier parent. Dispatch exactly one dedicated
`pipeline.yml` `action=recover-promotion-pr-checks` immediately. Keep
production/release-carrier fail-closed boundaries unchanged.

**Requirement source:** [GitHub issue #1109](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1109).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1109)

| Item | Value |
|------|-------|
| Repository | `KARSIFT/vocanova-platform-sandbox` |
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090) |
| Exact promotion head | `c3a53bab3035b7f08c0fb959bdf1b56bf330d291` |
| Infrastructure policy | `KARSIFT/karsift-ai-infra@67bdfd13ef875dead23ce4be01d7d0e8b976e289` |
| Failing release carrier | run `33340381776`, job `99334840338` |
| Step | `Recover missing exact-head promotion checks` |
| Started | `2026-08-30T22:54:51Z` |
| Ended | `2026-08-30T23:25:23Z` |
| Timeout diagnostics | `mode: promotion_pr`; `target_sha: c3a53bab…`; `promotion_pr_number: 1090`; `timeout_seconds: 1800`; `missing_checks: none`; `pending: 0`; `failed: 0`; `successful: 6`; then exit 1 |
| Duplicate no-progress carrier | run `33340516672` (force-cancelled after ordinary cancellation did not stop the sleeping reusable job) |
| Workaround dispatch | `gh workflow run pipeline.yml --ref develop -f action=recover-promotion-pr-checks -f promotion_pr_number=1090` |
| Workaround dedicated SUCCESS | run `33341923799`, titled `promotion-pr-validation PR #1090`, exact head `c3a53bab…` |
| Subsequent reconcile-release | run `33342062118` succeeded, including the production merge guard |

Live fixture contract at issue creation (tracked pin
`67bdfd13ef875dead23ce4be01d7d0e8b976e289`):

- `recovery_complete()` returns false when
  `promotion_ci_context_is_attestable()` cannot bind a unique SUCCESS `ci / ci`
  check to an attestable completed non-carrier parent. That half of VOC-140
  is working.
- `apply_promotion_pr_recovery_plan()` calls `plan_required_check_recovery()`
  and keeps only dispatch plans whose context is in that result. SUCCESS
  required rows are skipped, so `ci / ci` is absent from `dispatch_contexts`
  and `pipeline.yml` is filtered out.
- `collect_missing()` / `format_timeout_diagnostics()` report
  `missing_checks: none` from GitHub/gate SUCCESS counts and do not name
  unattestable CI evidence.
- `suppress_active_or_successful_dispatches()` drops `pipeline.yml` when a
  supplied `gate_summary` context state is `SUCCESS` or `PENDING`. Reusing
  that helper unchanged as the dedicated-dispatch filter would recreate the
  hang.

## Scope and non-goals

### In scope

1. Promotion recovery must immediately plan exactly one dedicated
   `recover-promotion-pr-checks` dispatch when the required PR view shows
   SUCCESS `ci / ci` but composed evidence has no uniquely attestable
   completed non-carrier parent.
2. A valid completed dedicated parent must complete recovery without
   redispatch. Active or successful exact dedicated recovery must suppress
   duplicates. Release carriers and other `pipeline.yml` runs must not
   suppress that dispatch.
3. Timeout diagnostics must identify unattestable CI evidence rather than
   `missing_checks: none` alone.
4. Deterministic live-shaped regressions for the five issue classes, plus
   caller pin/fixture mirror, tests, and current-state docs found by an
   exhaustive search. At minimum reconcile the fixture `README.md` and
   `docs/operations/11-devops-and-ci-cd.md`, which currently claim recovery
   dispatches dedicated validation when no completed non-carrier run exists.
5. After exact-SHA review and merge, ordinary `reconcile-release` must not
   hang on this class. Closure of #1109 binds allowlisted metadata from a
   successful recovery/release run.

### Non-goals / explicitly excluded

- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging a promotion PR.
- Changing the VOC-140 two-token contract, App-token mints, or
  Administration permission set.
- Creating a duplicate promotion PR or release audit issue.
- Switching the promotion PR application check to `--squash-safe-push`.
- Rerunning doomed `pull_request` `ci / ci` jobs as a strategy.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- A VOC-097 operator-owned live-evidence second task (see D10).
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Rewriting VOC-140, VOC-139, VOC-138, or earlier package records under
  `specs/changes/`.
- Fetching, hydrating, or recapturing VOC-112 JSON fixtures or hashed sources.
- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Repairing GitHub reusable-job cancellation of sleeping recovery runners
  (incident `33340516672` is audit evidence that the timeout is stuck; it is
  not a second outcome).
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable required-check recovery and status attestation; named
  current-state docs found by the required exhaustive search.
- Protected technical effect: whether promotion recovery dispatches dedicated
  `promotion-pr-validation` when GitHub-required `ci / ci` is SUCCESS and
  composed evidence is unattestable, instead of polling until timeout. No
  application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-test and recovery-dispatch changes
  still require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-141-D00`: This is one outcome-sized residual recovery-dispatch repair.
Use one end-to-end implementation task covering the coordinated
infrastructure PR, caller pin/fixture mirror, unattestable-SUCCESS dispatch,
timeout diagnostics, tests, docs, evidence, and release handoff. Repository
count, dispatch versus diagnostics, and tests-versus-docs are not split
reasons. Later promotion of this repair is evidence of the outcome, not a
second task and not a snapshot of the develop/main gap.

`VOC-141-D01`: When the required PR view has a SUCCESS `ci / ci` row but the
authoritative composed evidence has no uniquely attestable completed
non-carrier parent, promotion recovery must immediately plan exactly one
dedicated `pipeline.yml` dispatch with `action=recover-promotion-pr-checks`
and `promotion_pr_number` bound to the promotion PR, producing
`display_title` `promotion-pr-validation PR #<n>`. It must not wait
`DEFAULT_TIMEOUT_SECONDS` (1800) with no dispatch. It must not rerun a doomed
`pull_request` `ci / ci` carrier as a strategy. The class of run
`33340381776` / job `99334840338` cannot recur.

`VOC-141-D02`: `apply_promotion_pr_recovery_plan` must consult the same
composed attestable-CI rule as `recovery_complete` /
`promotion_ci_context_is_attestable`, not only
`plan_required_check_recovery`. GitHub-required SUCCESS is not sufficient to
omit dedicated CI recovery when the selected row cannot be bound to a unique
attestable completed non-carrier parent, including when the composed
attestable summary filtered that row out. The planner must receive the
current `gate_summary` and `workflow_runs` (or equivalent composed evidence
already loaded by `run_metadata_phase`). Filtering `plan_recovery_dispatches`
exclusively by `plan_required_check_recovery` SUCCESS-skipping is the defect
class of issue #1109.

`VOC-141-D03`: A valid completed dedicated parent completes without
redispatch. When composed evidence already contains a unique exact-head /
PR-bound completed successful `promotion-pr-validation PR #<n>` parent,
`recovery_complete` remains true and dedicated dispatch is not planned.
This preserves the workaround class of run `33341923799`.

`VOC-141-D04`: Active or successful exact dedicated recovery suppresses
duplicates. A queued/in-progress/pending/waiting or completed-success
exact-head `promotion-pr-validation PR #<n>` run may suppress a second
`recover-promotion-pr-checks` dispatch via
`dedicated_recovery_run_covers_dispatch` (or the live equivalent). An
in-progress, failed, cancelled, or successful **release carrier**, a
duplicate native carrier such as `33340516672`, any other `pipeline.yml`
run, and a GitHub-required SUCCESS row whose parent is unattestable must not
suppress that dispatch. If `suppress_active_or_successful_dispatches` is
reused, its current SUCCESS/PENDING `context_states` filter must not drop
`pipeline.yml` for unattestable CI.

`VOC-141-D05`: Timeout diagnostics must identify unattestable CI evidence
rather than `missing_checks: none` alone. When the poll deadline expires and
composed `ci / ci` is still unattestable, diagnostics must include a stable
sanitized token such as `unattestable_ci_evidence` (or the live equivalent)
in addition to existing mode/SHA/PR/timeout/pending/failed/successful
fields. Do not print tokens, secrets, or raw provider payloads.

`VOC-141-D06`: Production/release-carrier fail-closed boundaries remain
unchanged. Never select an in-progress or failed release carrier as trusted
recovered `ci / ci`. Status attestation retains VOC-138/VOC-139
exact-identity `pr-validation` binding: PR number, repository, immutable
base/head SHAs, configured `main` ← `develop` pair, expected workflow/path,
and genuine success. Same-head `--squash-safe-push` remains insufficient.
Doomed `pull_request` `ci / ci` jobs are still not rerun. Required `ci / ci`
remains a required check. Completed exact-bound `pull_request` `ci / ci` may
attest only when its parent workflow run is completed/successful and is not
a release carrier. The VOC-140 two-token production merge guard, mutation
token permission set, isolated Administration-write guard token, omitted
`bypass_actors` distinct failure, and `bypass_actors: []` requirement are
out of scope to change.

`VOC-141-D07`: Tests must exercise the live planner path, not only helpers
that omit SUCCESS required rows. Include at least: required `ci / ci`
SUCCESS plus filtered/unattestable parent causes
`apply_promotion_pr_recovery_plan` (or the live recovery loop) to dispatch
exactly one dedicated validation; a valid completed dedicated parent
completes without redispatch; active/successful exact dedicated recovery
suppresses duplicates; timeout diagnostics name unattestable CI evidence; and
production/release-carrier fail-closed cases still reject. A unit test that
only calls `suppress_active_or_successful_dispatches` without GitHub-required
SUCCESS rows is not sufficient coverage of #1109.

`VOC-141-D08`: Pin advance. Issue-creation pin
`67bdfd13ef875dead23ce4be01d7d0e8b976e289` is the defective live contract
for this failure class. T00 opens one new `KARSIFT/karsift-ai-infra` PR,
obtains independent exact-revision review, and after that merge sets
`PINNED_SHA.txt` and every changed mirrored fixture file to that exact merge.
Mirror at least recovery/attestation sources and their tests. If exact
comparison proves another authoritative fixture file also changed, mirror it
too. Do not treat the untracked local `karsift-ai-infra/` checkout as this
repository's tracked tree. Reconcile all live caller pin-lock tests that
assert the current fixture pin or mirrored hashes. Preserve historical
`AUTHORITATIVE_PIN` / issue-era pin constants and package evidence, including
VOC-140 pin `67bdfd13…` as historical.

`VOC-141-D09`: No VOC-097 live-evidence second task and no snapshot-gap
task. Workaround runs `33341923799` and `33342062118` already proved dedicated
dispatch unblocks this class; they are incident evidence, not this package's
implementation work. Live success after this repair is ordinary
release/closure evidence recorded with allowlisted metadata only. Do not
snapshot the develop/main gap (`karsift-ai-infra#15`). Do not create a
duplicate promotion PR or release audit. If #1090 already merged via the
documented workaround, this package's own later promotion is the recurrence
proof; do not reopen or recreate #1090/#1089.

`VOC-141-D10`: Docs in the same PR. Before editing, exhaustively search
tracked source and current documentation for claims that required-check
SUCCESS completes promotion recovery, that dedicated validation is dispatched
whenever no completed non-carrier run exists, and for the current pin
literal/hash assertions. Record the searched patterns and resulting path
disposition in `t00-evidence.md`. Update every current-state document that
would otherwise remain false, including the fixture `README.md` and
`docs/operations/11-devops-and-ci-cd.md`. Current docs must state that
GitHub-required SUCCESS is not sufficient when composed CI is unattestable
and that recovery dispatches dedicated validation immediately rather than
polling to timeout. Preserve clearly labeled historical VOC-140/VOC-139
records. Do not rewrite those package directories.

`VOC-141-D11`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. No App ID/private-key
secret is rotated. Preserve the named run/job IDs as audit evidence only (no
raw logs).

`VOC-141-D12`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- targeted VOC-141 unattestable-SUCCESS dispatch, duplicate-suppression,
  timeout-diagnostic, and unchanged carrier fail-closed cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-141-D13`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate: SUCCESS-plus-unattestable-parent dedicated dispatch;
completed dedicated parent without redispatch; active/successful exact
dedicated suppression; timeout unattestable-CI diagnostics; unchanged
production/release-carrier fail-closed boundaries; no doomed `ci / ci` rerun;
current-state docs; and pin advance. Merge-gate must reject any mismatch.
Committed `t00-evidence.md` records the implementation PR base, new infra
merge, dispatch change, diagnostic change, and the contract that later
exact-head binding is published as review/check metadata. A tracked file must
not be required to contain the SHA of the same commit that contains it.

`VOC-141-D14`: Protected comparison versus implementation PR base.
Issue-creation promotion head is `c3a53bab3035b7f08c0fb959bdf1b56bf330d291`.
Implementation must resolve current `develop` to a 40-character SHA before
any in-scope edit and record that SHA as the implementation PR base. Fail
closed on unrelated/material movement of `develop` (any tree change outside
this package directory, in-scope fixture/pin/tests, and the named
current-state docs). This package's own plan/adoption/roster commits after
`c3a53bab…` are governance-only and do not count as protected-file drift.

`VOC-141-D15`: Release handoff. After the exact reviewed caller merge,
ordinary `reconcile-release` for the live release audit may merge the live
same-repository promotion at the then-current `develop` tip once required
checks pass under the repaired dispatch. `develop` is advanced to the exact
promotion merge SHA before audit close. Every promotion merge push to `main`
triggers automatic production deployment, whose exact-SHA result is verified.
Do not snapshot the current gap. Closed state alone is not completion proof.
Preserve runs `33340381776`, `33340516672`, `33341923799`, `33342062118` and
job `99334840338` as audit evidence. Root issue #1109 closes only after
allowlisted metadata from a successful recovery/release run exists. Do not
create a duplicate promotion PR or release audit.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents a
release deadlock without letting recovery attest in-progress carriers, rerun
doomed PR jobs, or skip dedicated validation when GitHub-required SUCCESS is
unattestable.

Abuse/process risks:

1. Treating GitHub-required SUCCESS as sufficient to skip dedicated dispatch
   while composed CI is unattestable — forbidden by `VOC-141-D01` and
   `VOC-141-D02`.
2. Waiting the full 1,800-second timeout with no dispatch — forbidden by
   `VOC-141-D01`.
3. Selecting a still-running or failed release carrier as completed `ci / ci`
   — forbidden by `VOC-141-D06`.
4. Treating any in-progress non-dedicated `pipeline.yml` as covering
   `recover-promotion-pr-checks` — forbidden by `VOC-141-D04`.
5. Rerunning doomed `pull_request` `ci / ci` jobs — forbidden by
   `VOC-141-D01` and `VOC-141-D06`.
6. Weakening the production merge guard or changing the two-token contract —
   forbidden by `VOC-141-D06` and `VOC-141-DEP-04`.
7. Fabricating statuses or manually merging a promotion PR — forbidden by
   `VOC-141-DEP-04`.
8. Snapshotting the develop/main gap or adding a self-invalidating
   snapshot-then-check task — forbidden by `VOC-141-D09` and
   `karsift-ai-infra#15`.
9. Changing `roles.yml` or adding an OpenAI route — forbidden by
   `VOC-141-D11`.
10. Printing credentials or copying full CI logs into evidence — forbidden.
11. Requiring a commit to contain its own SHA — forbidden by `VOC-141-D13`.

## Contradictions and open questions

1. **VOC-140 claim versus live dispatch:** VOC-140 required dedicated
   `promotion-pr-validation` when no completed non-carrier run exists, and
   `recovery_complete` now correctly returns false. Live pin `67bdfd13…`
   still omits the dispatch because `apply_promotion_pr_recovery_plan` consults
   only GitHub-required SUCCESS/missing/failed rows. This package follows
   D01/D02 and treats that remaining planner gap as in-scope residual work,
   not as a rewrite of VOC-140 records.
2. **`suppress_active_or_successful_dispatches` SUCCESS filter:** the helper
   used by VOC-140 tests keeps `pipeline.yml` when `gate_summary` is omitted,
   but drops it when context state is SUCCESS. Implementation must not reuse
   that SUCCESS filter as the dedicated-dispatch gate for unattestable CI.
3. **PR #1090 live state after workaround:** issue #1109 records that
   dedicated run `33341923799` succeeded and reconcile-release `33342062118`
   succeeded including the production merge guard. #1090 may already have
   merged by adoption time. This package must not create a duplicate
   promotion PR or release audit. Remaining live proof is that a later
   promotion, including this package's own, does not hang on this class.
4. **New infrastructure merge SHA:** not available at drafting time; record
   it after the coordinated infra PR merges. Implementation writes that SHA
   into `PINNED_SHA.txt` and `t00-evidence.md`. Do not invent it at planning
   time.
5. **Reusable-job cancellation:** duplicate carrier `33340516672` required
   force-cancel because ordinary cancellation did not stop the sleeping
   reusable job. That operational stuckness is evidence of the 1,800-second
   poll, not a second authorized outcome.
