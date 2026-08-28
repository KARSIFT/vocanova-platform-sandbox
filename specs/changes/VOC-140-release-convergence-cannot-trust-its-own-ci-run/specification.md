# VOC-140 — Release convergence cannot trust its own CI run or verify the production merge guard with the App token: Specification

## Objective and requirement source

Unblock the already-open same-repository `main` ← `develop` promotion so
release convergence can trust recovered `ci / ci` evidence and can prove the
live production merge guard with the App identity that merges. Do not create a
duplicate promotion PR or release audit. Do not weaken ruleset enforcement.

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
2. Least-privilege token/API contract so the identity that calls
   `verify-production-merge-guard.sh` immediately before `gh pr merge` can
   prove full ruleset/no-bypass fields. Distinct fail-closed when the
   token-visible payload omits those fields, with a precise operator action
   naming the GitHub App setting if the App registration lacks the required
   permission. The same contract applies to `merge-gate.yml` when it calls the
   same verifier against the production branch.
3. Deterministic regressions that exercise the real token-visible payload
   shape (omitted `bypass_actors`) and the circular-CI identity (newest
   `ci / ci` check SUCCESS on a still-running `pipeline.yml` that also
   contains `release / converge`).
4. Caller pin/fixture mirror of the new independently reviewed infrastructure
   merge, caller tests, and current-state docs that would otherwise keep
   claiming the App token is contents/issues/pull-requests-only and that
   recovery completion may ignore parent-run identity.
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
  named current-state docs.
- Protected technical effect: whether release convergence can attest recovered
  `ci / ci` without selecting its own in-progress carrier, and whether the App
  identity that merges to `main` can prove a non-bypassable production
  ruleset. No application runtime effect is intended.
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
contains `release /` jobs or is a `reconcile-release` / release-converge
carrier, is not attestable CI. Newest-check selection that ignores parent-run
completion is the defect class of job `98738317266`.

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

`VOC-140-D06`: Least-privilege token/API contract. The identity that calls
`verify-production-merge-guard.sh` immediately before `gh pr merge` must
receive a token whose visible payload includes the full ruleset/no-bypass
fields. Prefer adding the least-privilege `actions/create-github-app-token`
permission to that existing mint rather than proving the guard with a
different identity than the merger. GitHub Apps currently list
`GET /repos/{owner}/{repo}/rulesets/{ruleset_id}` under Metadata:read, so the
HTTP call can succeed without returning `bypass_actors`. GitHub REST
documents that `bypass_actors` is omitted unless the caller has write access
to the ruleset. Implementer must capture the live App-token JSON and the
administrator-visible JSON at this revision and request the narrowest
supported mint permission that actually returns `bypass_actors` as an array.
Prefer `permission-administration: read` if it returns the field; if GitHub
requires write access to expose it, `permission-administration: write` is the
least privilege that works and must be documented as a GitHub API constraint,
not as a desire to administer the repository. Do not add `permission-workflows`,
`permission-actions`, or other unrelated grants. Apply the same mint change
to `merge-gate.yml` for the production-branch path that already calls the
same verifier. Do not switch the verifier to `github.token` as a way to avoid
diagnosing the App contract.

`VOC-140-D07`: Distinct omitted-field failure. When the fetched ruleset JSON
omits `bypass_actors` or returns a non-array, fail closed with a new
sanitized class such as `production_merge_guard_payload_incomplete` (or the
live equivalent) that is not `production_merge_guard_missing`. The message
must name the precise operator action: grant the diagnosed permission on
GitHub App `karsift-ai-infra-bot` for this repository (Repository permissions
→ Administration: Read, or Write if D06 proved write is required to expose
the field) and re-run `reconcile-release`. Do not print tokens, private-key
material, or full ruleset dumps.

`VOC-140-D08`: Tests must exercise the real token-visible payload shape, not
only helper logic with complete admin fixtures. Include at least: omitted
`bypass_actors`; `bypass_actors: []`; non-empty bypass; and a subprocess of
`verify-production-merge-guard.sh` with a mock `gh` that returns those
shapes. Include a circular-CI fixture whose newest `ci / ci` check is SUCCESS
while the parent workflow run is `in_progress` and contains `release / converge`.
Do not treat a unit test that only calls `validate_production_merge_guard` with
hand-built full fixtures as covering D06/D07.

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

`VOC-140-D11`: Docs in the same PR. Update every current-state document that
would otherwise remain false, including
`docs/operations/11-devops-and-ci-cd.md` (today: App token remains limited to
PR, issue, and content mutations / mutation-only) and the fixture `README.md`
equivalent. State that recovery/selection never treats a still-running
release carrier as attestable `ci / ci`, that dedicated
`promotion-pr-validation` must be completed/successful, and that the merge
App identity requests the least-privilege permission required to prove
ruleset/no-bypass fields. Do not rewrite VOC-139 or VOC-138 package records.

`VOC-140-D12`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. Preserve the named
run/job IDs as audit evidence only (no raw logs).

`VOC-140-D13`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- targeted VOC-140 circular-CI, omitted-`bypass_actors`, empty-bypass, and
  non-empty-bypass cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-140-D14`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate the circular-CI identity case, dedicated
promotion-pr-validation requirement, omitted-`bypass_actors` token shape,
empty-bypass acceptance, non-empty-bypass rejection, no-fabricated-status
constraint, and pin advance. Merge-gate must reject any mismatch. Committed
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
6. Switching the verifier to a different identity so the merger never proves
   the guard — forbidden by `VOC-140-D06`.
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

1. **GitHub field visibility versus issue wording:** issue #1102 asks for a
   least-privilege **read** permission if the App registration lacks it.
   GitHub Apps currently list `GET /repos/{owner}/{repo}/rulesets/{id}` under
   Metadata:read, and GitHub REST documents that `bypass_actors` is omitted
   unless the caller has **write access to the ruleset**. This package does
   not guess which mint permission returns the field. `VOC-140-D06` requires
   the implementer to capture the live App-token payload and request the
   narrowest supported permission that actually returns `bypass_actors` as an
   array. If that is write, document it as a GitHub API constraint and still
   do not add bypass actors.
2. **Docs versus live App-token mint:** `docs/operations/11-devops-and-ci-cd.md`
   currently says the App token remains limited to PR, issue, and content
   mutations / mutation-only. That sentence is true of the issue-creation pin
   and becomes false if D06 adds an administration permission. This package
   follows issue #1102 (`VOC-140-D06`, `VOC-140-D11`) and updates those docs.
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
7. **App installation grant:** whether `karsift-ai-infra-bot` already has
   Administration Read/Write on this repository is not proven at drafting
   time. If the mint requests a permission the installation lacks, D07's
   operator action is the required fail-closed path. This package does not
   authorize changing GitHub App installation settings itself except as that
   documented operator action outside the implementation PR (repository
   settings are the 2026-08-08 settings exception; the workflow mint change
   still lands in the governed PR).
