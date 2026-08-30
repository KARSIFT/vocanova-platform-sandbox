# VOC-140 — Release convergence cannot trust its own CI run or verify the production merge guard with the App token: Specification

## Objective and requirement source

Unblock the already-open same-repository `main` ← `develop` promotion so
release convergence can trust recovered `ci / ci` evidence and a separate,
current-repository-scoped guard token can prove the live production merge guard
before the unchanged mutation token merges. Do not create a duplicate promotion
PR or release audit. Do not weaken ruleset enforcement.

**Requirement source:** [GitHub issue #1102](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1102).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1102)

| Item | Value |
|------|-------|
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090) |
| Release issue | [#1089](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1089) |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `21eef75549226766fc4f78f62f232ee5fbdb8d6d` |
| Issue-creation pin | `599436835371f27fac52ec6b47a18b36257366ac` |
| Dedicated recovery success | run `33136865709` |
| Reproduction 1 dispatch | `gh workflow run pipeline.yml --ref develop -f action=reconcile-release -f release_issue_number=1089` |
| Reproduction 1 run / job | `33136633666` / `98738317266` |
| Reproduction 1 behavior | recovery reports complete without dispatch; selector chooses `ci / ci` from the still-running workflow that contains `release / converge` |
| Reproduction 1 fail-closed | `untrusted_ci_recovery_identity` |
| Reproduction 2 dispatch | `gh workflow run pipeline.yml --ref develop -f action=recover-promotion-pr-checks -f promotion_pr_number=1090` then `reconcile-release` for #1089 |
| Reproduction 2 release runs / jobs | `33136984634` / `98739074178`; `33137091931` / `98739420310` |
| Reproduction 2 attestation | required-context attestation succeeds, including `ci / ci` from `33136865709` |
| Reproduction 2 fail-closed | `production-merge-guard: production_merge_guard_missing` before `gh pr merge` |
| Admin-visible command | `bash tooling/governance/fixtures/karsift-ai-infra/config/verify-production-merge-guard.sh KARSIFT/vocanova-platform-sandbox main tooling/governance/fixtures/karsift-ai-infra/config` |
| Admin-visible result | `production-merge-guard: ok ruleset_id=20575146` |
| Live ruleset | `20575146`; repository-owned; active; `main`-only; strict; requires PRs and `governance-policy` / `validate` / `ci`; `bypass_actors: []` |
| App-token mint | `permission-contents: write`, `permission-issues: write`, `permission-pull-requests: write` |
| Production promotion | none; `main` remains `0d0b0cdf…` |

## Scope and non-goals

### In scope

1. Recovery and authoritative selection must never treat an in-progress or
   failed release carrier as trusted recovered `ci / ci`. The dedicated
   `promotion-pr-validation PR #<n>` workflow is the recovery carrier that may
   be dispatched or selected, and that run must be `completed`/`success`
   before attestation.
2. A strict two-token contract immediately before production merge. Preserve
   the existing mutation token with exactly `permission-contents: write`,
   `permission-issues: write`, and `permission-pull-requests: write`; it remains
   the sole App token supplied to `gh pr merge` and mutation steps. Separately
   mint an ephemeral guard-only token scoped with `owner` plus `repositories`
   to the current caller repository and with only
   `permission-administration: write`. Inject that token only into
   `verify-production-merge-guard.sh`, immediately before merge, and never into
   mutation, status, issue, pull-request, or `gh pr merge` execution. Apply the
   same separation to the production-branch path in `merge-gate.yml`.
3. Deterministic regressions that exercise the real token-visible payload
   shape (omitted `bypass_actors`) and the circular-CI identity (newest
   `ci / ci` check SUCCESS on a still-running `pipeline.yml` that also
   contains `release / converge`).
4. Caller pin/fixture mirror of the new independently reviewed infrastructure
   merge, caller tests, and all current-state docs found by an exhaustive
   repository search. At minimum reconcile the fixture `README.md`,
   `docs/operations/11-devops-and-ci-cd.md`,
   `docs/governance/repository-settings.md`,
   `docs/governance/post-merge-activation-checklist.md`, and
   `docs/operations/19-governance-reconciliation-notes.md`. Preserve clearly
   marked historical records while correcting current claims to active A-004,
   enabled automatic release/production deployment, and the two-token contract.
   Reconcile any validator/test that enforces the stale wording, including
   `scripts/governance/validate-governance.sh` if its current assertions remain.
5. After exact-SHA review and merge, rerun dedicated promotion recovery if
   necessary, then `reconcile-release` for #1089 can merge #1090 (or the live
   promotion at the then-current `develop` head) and converge `develop` to
   that exact merge SHA. Closure of #1102 binds allowlisted metadata from the
   successful recovery/release run.

### Non-goals / explicitly excluded

- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging #1090.
- Creating a duplicate promotion PR or release audit issue.
- Switching the promotion PR application check to `--squash-safe-push`.
- Rerunning doomed `pull_request` `ci / ci` jobs as a strategy.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- A VOC-097 operator-owned live-evidence second task (see D10 / contradictions).
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Rewriting VOC-139, VOC-138, or earlier package records under `specs/changes/`.
- Fetching, hydrating, or recapturing VOC-112 JSON fixtures or hashed sources.
- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable required-check recovery, status attestation, and
  production-merge-guard authorization; GitHub App token mint on release (and
  the production-branch merge-gate path that uses the same verifier);
  named current-state docs and every additional current-state source found by
  the required exhaustive search.
- Protected technical effect: whether release convergence can attest recovered
  `ci / ci` without selecting its own in-progress carrier, and whether the
  isolated guard token can prove a non-bypassable production ruleset before
  the unchanged mutation token merges to `main`. No application runtime
  effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-test and App-token-mint changes
  still require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-140-D00`: This is one outcome-sized release-convergence repair. Use one
end-to-end implementation task covering the coordinated infrastructure PR,
caller pin/fixture mirror, recovery/selection identity, production-merge-guard
token/API contract, tests, docs, evidence, and release handoff. Repository
count, CI-identity versus token contract, and tests-versus-docs are not split
reasons. Merging #1090 after this correction is evidence of this outcome, not
a second task and not a snapshot of the develop/main gap.

`VOC-140-D01`: Never select an in-progress or failed release carrier as
trusted recovered `ci / ci`. A check named `ci / ci` whose parent workflow
run is not `status=completed` and `conclusion=success`, or whose workflow run
actually executed a non-skipped `release /` job or is identified by event,
inputs, or title as a `reconcile-release` / release-converge carrier, is not
attestable CI. Merely defining a skipped `release` job in `pipeline.yml` does
not disqualify the dedicated D02 recovery run. Newest-check selection that
ignores parent-run completion is the defect class of job `98738317266`.

`VOC-140-D02`: Dedicated promotion-PR recovery remains the recovery carrier.
When promotion recovery cannot attest a completed successful `ci / ci` run
that is not a release carrier, it must dispatch or wait for
`pipeline.yml` `action=recover-promotion-pr-checks` whose `display_title` is
`promotion-pr-validation PR #<n>`. That dedicated run must be
`completed`/`success` and exact-head/PR-bound before attestation. Do not
treat any in-progress `pipeline.yml` run, including the release carrier
itself, as covering that dispatch. Completed exact-bound `pull_request`
`ci / ci` may remain attestable only when its parent workflow run is
completed/successful and is not a release carrier; it must not win over, or
be confused with, a newer in-progress carrier check.

`VOC-140-D03`: Recovery completion and dispatch suppression must use the same
attestable-run rule as D01/D02. `recovery_complete` must not return true for
`ci / ci` merely because `gh pr checks --required` shows SUCCESS. Suppressing
`recover-promotion-pr-checks` because some `pipeline.yml` run at the SHA is
queued, in progress, or successful is forbidden unless that run is the
dedicated promotion-pr-validation dispatch and is itself completed/successful
or still actively running as that dedicated dispatch.

`VOC-140-D04`: Status attestation retains VOC-138/VOC-139 exact-identity
`pr-validation` binding: PR number, repository, immutable base/head SHAs,
configured `main` ← `develop` pair, expected workflow/path, and genuine
success. For `ci / ci` `workflow_dispatch`, `display_title` remains
`promotion-pr-validation PR #<n>`. Same-head `--squash-safe-push` remains
insufficient. Doomed `pull_request` `ci / ci` jobs are still not rerun as a
strategy. Required `ci / ci` remains a required check.

`VOC-140-D05`: Production merge guard strength is unchanged. A qualifying
ruleset must still be repository-owned, active, production-branch-effective,
strict, have a non-empty required-check list, include a pull-request rule,
and expose `bypass_actors: []`. Live ruleset `20575146` is the incident
proof that the guard exists; the defect is token-visible payload shape, not
the ruleset. Do not add bypass actors. Do not accept `null`, omitted, or
non-array `bypass_actors` as empty.

`VOC-140-D06`: Strict two-token least-privilege contract. The existing
mutation-token mint remains byte-for-byte equivalent in permission scope:
exactly `permission-contents: write`, `permission-issues: write`, and
`permission-pull-requests: write` (plus GitHub's implicit Metadata read). It
remains the sole App token supplied to `gh pr merge`, PR/issue/content
mutation, and completion-marker publication. It must not receive
Administration permission. Immediately before production-guard verification,
mint a distinct ephemeral token from `karsift-ai-infra-bot` with explicit
`owner: ${{ github.repository_owner }}`, `repositories` limited to the current
caller repository name, and exactly `permission-administration: write` (plus
implicit Metadata read). The guard token must be injected only as the
credential for `verify-production-merge-guard.sh`; it must never be exposed to
`gh pr merge`, mutation, Commit-status, issue, or pull-request steps. The
following merge step must revalidate the exact PR head/base/refs and use only
the mutation token. Apply the identical two-token separation to the
production-branch path in `merge-gate.yml`. Do not add Actions, Workflows,
Contents, Issues, Pull requests, or Statuses permission to the guard token and
do not switch either guard or merge to `github.token`.

`VOC-140-D07`: External activation prerequisite and distinct omitted-field
failure. Before hosted acceptance, GitHub App `karsift-ai-infra-bot` must have
Repository permissions → Administration: **Read and write**, and the owner of
KARSIFT organization installation `148001476` must approve the pending
permission change. Read-only live inspection at drafting time shows that
installation uses `repository_selection: all` and currently grants Contents,
Issues, Pull requests, and Workflows write plus Metadata read, but no
Administration permission. The runtime guard mint must therefore keep its
explicit owner-plus-current-repository restriction; that token restriction does
not narrow the App/private-key ceiling across the installation's selected
repositories. This is a known external repository/App setting, not a permission
to guess at implementation time, and it requires no App-ID or private-key secret
rotation. When the fetched ruleset JSON
omits `bypass_actors` or returns a non-array, fail closed with a new
sanitized class such as `production_merge_guard_payload_incomplete` (or the
live equivalent) that is not `production_merge_guard_missing`. The message
must name the precise operator action: update `karsift-ai-infra-bot` to
Administration: Read and write, have the installation owner approve that
permission on KARSIFT installation `148001476`, retain the workflow's explicit
single-repository token scope for `KARSIFT/vocanova-platform-sandbox`, then
rerun the failed guard / `reconcile-release`; do not rotate secrets. Hosted
evidence must show the guard-only token returns an explicit
`bypass_actors: []` for the effective production ruleset before merge is
allowed. Do not print tokens, private-key material, or full ruleset dumps.
This operator-owned activation occurs after the reviewed infra and caller
repairs merge and before #1090 promotion; it must not consume or exhaust T00
implementer retries. T00 is responsible for the fail-closed code, tests,
documentation, and precise operator action, while the hosted empty-bypass probe
is release/closure evidence.

`VOC-140-D08`: Tests must exercise the real token-visible payload shape and
token isolation, not
only helper logic with complete admin fixtures. Include at least: omitted
`bypass_actors`; `bypass_actors: []`; non-empty bypass; and a subprocess of
`verify-production-merge-guard.sh` with a mock `gh` that returns those
shapes. Include a circular-CI fixture whose newest `ci / ci` check is SUCCESS
while the parent workflow run is `in_progress` and contains `release / converge`.
Do not treat a unit test that only calls `validate_production_merge_guard` with
hand-built full fixtures as covering D06/D07. Parse both workflows and prove
the mutation mint has exactly Contents/Issues/Pull requests write, the guard
mint has exactly Administration write and current-caller-repository scope, the
guard step immediately precedes the exact-head merge decision, and no guard
token expression or output reaches `gh pr merge`, status publication, issue,
PR, content mutation, or completion-marker steps.

`VOC-140-D09`: Pin advance. Issue-creation pin
`599436835371f27fac52ec6b47a18b36257366ac` is the defective live contract for
this failure class. T00 opens one new `KARSIFT/karsift-ai-infra` PR, obtains
independent exact-revision review, and after that merge sets
`PINNED_SHA.txt` and every changed mirrored fixture file to that exact merge.
Mirror at least recovery/attestation/guard sources, `release.yml`, the
production-branch `merge-gate.yml` mint if changed, and their tests. If exact
comparison proves another authoritative fixture file also changed, mirror it
too. Do not treat the untracked local `karsift-ai-infra/` checkout as this
repository's tracked tree. Reconcile all live caller pin-lock tests that
assert the current fixture pin or mirrored hashes. Preserve historical
`AUTHORITATIVE_PIN` / issue-era pin constants and package evidence.

`VOC-140-D10`: No VOC-097 live-evidence second task and no snapshot-gap
task. Successful #1090 recovery/release HEAD is the promotion/`develop` tip
after this repair merges. A draft evidence-carrier PR cannot satisfy
`exact_pr_head` or `integration_contains_pr_head` against that SHA without
treating the carrier as the promotion PR. Live success is ordinary
release/closure evidence recorded with allowlisted metadata only (workflow
identity, event, branch, HEAD SHA, run/job IDs, conclusion). Do not snapshot
the develop/main gap (`karsift-ai-infra#15`).

`VOC-140-D11`: Docs in the same PR. Before editing, exhaustively search tracked
source and current documentation for the old pin, pin/hash assertions,
`mutation-only` / contents-issues-pull-requests token claims, active-A-003
claims, and disabled/unimplemented automatic-release or production-deployment
claims. Record the searched patterns and resulting path disposition in
`t00-evidence.md`. Update every current-state document that would otherwise
remain false, including the fixture `README.md`,
`docs/operations/11-devops-and-ci-cd.md`,
`docs/governance/repository-settings.md`,
`docs/governance/post-merge-activation-checklist.md`, and
`docs/operations/19-governance-reconciliation-notes.md`. The repository-settings
document must state active A-004 and the current enabled repository-controlled
release/production-deploy path while keeping RL1/RL2 disabled. Token docs must
describe the unchanged mutation-only token and separate guard-only,
Administration-write, caller-repository-scoped token; they must not call the
combined App-token system mutation-only. Preserve clearly labeled historical
A-003 and issue-era pin evidence. Do not rewrite VOC-139 or VOC-138 package
records. Update `scripts/governance/validate-governance.sh` and any relevant
tests in the same PR when they enforce the stale current-state wording; do not
weaken an unrelated governance invariant.

`VOC-140-D12`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. No App ID/private-key
secret is rotated. Token separation reduces accidental credential exposure,
but both tokens are minted from the same App registration and private key, so
compromise of that key or a workflow context able to mint tokens retains the
App installation's combined permission ceiling across installation `148001476`'s
current `repository_selection: all` scope. Record this residual risk. An
optional dedicated guard App with only Administration write and a
single-repository installation could further separate keys, but is future
hardening and is not required or authorized by T00. Preserve the named run/job
IDs as audit evidence only (no raw logs).

`VOC-140-D13`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- targeted VOC-140 circular-CI, omitted-`bypass_actors`, empty-bypass,
  non-empty-bypass, exact token-permission/scope, step-order, token-use
  isolation, and current-state source/pin reconciliation cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-140-D14`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate the circular-CI identity case, dedicated
promotion-pr-validation requirement, omitted-`bypass_actors` token shape,
empty-bypass acceptance, non-empty-bypass rejection, exact two-token
permission/scope/use isolation in both workflows, external App approval and
hosted proof, current-state docs, same-App residual risk,
no-fabricated-status constraint, and pin advance. Merge-gate must reject any mismatch. Committed
`t00-evidence.md` records the implementation PR base, new infra merge,
recovery-identity change, token/API contract, and the contract that later
exact-head binding is published as review/check metadata. A tracked file must
not be required to contain the SHA of the same commit that contains it.

`VOC-140-D15`: Protected comparison versus implementation PR base.
Issue-creation `develop` is `21eef75549226766fc4f78f62f232ee5fbdb8d6d`.
Implementation must resolve current `develop` to a 40-character SHA before
any in-scope edit and record that SHA as the implementation PR base. Fail
closed on unrelated/material movement of `develop` (any tree change outside
this package directory, in-scope fixture/pin/tests, and the named
current-state docs). This package's own plan/adoption/roster commits after
`21eef755…` are governance-only and do not count as protected-file drift.

`VOC-140-D16`: Release handoff. After the exact reviewed caller merge, rerun
the existing dedicated promotion recovery if necessary, then ordinary
`reconcile-release` for #1089 may merge #1090 once that PR's head includes
this repair, or a successor same-repository promotion PR at the then-current
`develop` tip. `develop` is advanced to the exact promotion merge SHA before
audit close. Every promotion merge push to `main` triggers automatic
production deployment, whose exact-SHA result is verified. Do not snapshot
the current gap. Closed state alone is not completion proof. Preserve runs
`33136633666`, `33136865709`, `33136984634`, `33137091931` and jobs
`98738317266`, `98739074178`, `98739420310` as audit evidence. Root issue
#1102 closes only after allowlisted metadata from the successful
recovery/release run exists. Do not create a duplicate promotion PR or
release audit.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents a
release deadlock without letting recovery attest in-progress carriers or
letting a token that cannot see `bypass_actors` claim the production guard
is present.

Abuse/process risks:

1. Selecting the still-running release carrier as completed `ci / ci` —
   forbidden by `VOC-140-D01`.
2. Treating any in-progress `pipeline.yml` as covering
   `recover-promotion-pr-checks` — forbidden by `VOC-140-D03`.
3. Weakening the production merge guard or adding bypass actors — forbidden
   by `VOC-140-D05`.
4. Treating omitted `bypass_actors` as empty — forbidden by `VOC-140-D05`
   and `VOC-140-D07`.
5. Granting unrelated App permissions or rotating `KARSIFT_BOT_*` secrets —
   forbidden by `VOC-140-D06` and `VOC-140-D12`.
6. Passing the guard token to merge/mutation steps, passing Administration to
   the mutation token, or separating verification from the immediately
   following fresh exact-head/base/ref revalidation and mutation-token merge —
   forbidden by `VOC-140-D06`.
7. Fabricating statuses or manually merging #1090 — forbidden by
   `VOC-140-DEP-04`.
8. Snapshotting the develop/main gap or adding a self-invalidating
   snapshot-then-check task — forbidden by `VOC-140-D10` and
   `karsift-ai-infra#15`.
9. Changing `roles.yml` or adding an OpenAI route — forbidden by
   `VOC-140-D12`.
10. Printing credentials or copying full CI logs into evidence — forbidden.
11. Requiring a commit to contain its own SHA — forbidden by `VOC-140-D14`.

## Contradictions and open questions

1. **GitHub field visibility versus issue wording:** issue #1102 suggested a
   least-privilege read permission, but `bypass_actors` requires ruleset write
   visibility. The resolved contract is a separate, current-repository-scoped
   token with only `permission-administration: write`; Administration is never
   added to the mutation token. Hosted explicit `bypass_actors: []` proof is
   required after App-installation approval.
2. **Docs versus live App-token mint:** `docs/operations/11-devops-and-ci-cd.md`
   currently says the App token remains limited to PR, issue, and content
   mutations / mutation-only. That remains true for the mutation token but is
   incomplete for the App-token system after the separate guard token lands.
   This package follows D06/D11 and documents both isolated tokens.
3. **Completed pull_request `ci / ci` versus dedicated recovery:** VOC-138
   retained exact-bound `pull_request` `ci / ci` as attestable when the run
   itself is completed/successful. Issue #1102 requires dispatch or selection
   of dedicated `promotion-pr-validation` and forbids selecting the
   still-running release carrier. `VOC-140-D02` keeps both: completed
   non-carrier `pull_request` `ci / ci` may attest, but it must not be
   confused with a newer in-progress carrier check, and recovery must
   dispatch dedicated validation when no completed non-carrier run exists.
4. **VOC-097 live evidence:** issue #1102 asks to verify #1090 merges, develop
   synchronizes, and production deploy succeeds. Those runs cannot exist until
   this repair is on `develop`, and a VOC-097 evidence-carrier PR cannot
   satisfy sha_lineage against the promotion HEAD. This package records that
   live run as ordinary release/closure evidence (`VOC-140-D10`,
   `VOC-140-D16`), not as a second task.
5. **PR #1090 head movement:** after this package merges, GitHub will move
   #1090's `develop` head. Recovery/promotion evidence may name a later
   exact head. The named 2026-08-28 run/job IDs remain the incident audit
   record.
6. **New infrastructure merge SHA:** not available at drafting time; record
   it after the coordinated infra PR merges. Implementation writes that SHA
   into `PINNED_SHA.txt` and `t00-evidence.md`. Do not invent it at planning
   time.
7. **App installation grant:** Administration: Read and write plus
   owner approval on KARSIFT organization installation `148001476` is a known
   external activation prerequisite. The installation currently uses
   `repository_selection: all`; D06 separately requires each runtime guard
   token to be restricted to `KARSIFT/vocanova-platform-sandbox`. D07 uses the
   repository-settings exception for that operator action; workflow code still
   lands through the governed PR. No secret rotation is required, and the path
   stays fail closed until hosted explicit empty-bypass proof exists.
