# VOC-138 — Promotion PR CI fails VOC-112 ancestry when captured commit is not reachable: Specification

## Objective and requirement source

Unblock same-repository `main` <- `develop` promotion PR validation when the
historical VOC-112 capture subject commit object is not reachable. Promotion
PR application checks must use the existing merge-base/hash-bound
`pr-validation` contract in that case, without fetching or hydrating evidence
commits and without weakening ordinary PR fail-closed behavior. Exact-head
check recovery must select or publish one unambiguous successful validation
instead of rerunning a structurally doomed `pull_request` job.

**Requirement source:** [GitHub issue #1091](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1091).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1091)

| Item | Value |
|------|-------|
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090) |
| Release issue | [#1089](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1089) |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `87f0efcb94a213a0ede9fdbca94a707a22d42b86` |
| Reusable CI checkout | `fetch-depth: 0` of `pull/1090/merge` |
| Application-check invocation | `run-app-checks.sh --pr-base-sha <base> --pr-head-sha <head>` |
| Selected mode | `pr-ancestry` (capture fixture differs between `main` and `develop`) |
| Failing tests | `VOC-112-TEST-12`, `VOC-112-TEST-13` |
| Fail-closed message | `PR ancestry mode requires every captured commit object` |
| Unreachable subject | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| PR run | `33122154521` |
| Failed jobs | `98691441027`, rerun `98692552949` |
| Dispatch recovery (pass) | `33122158425` (339 foundation tests; squash-safe non-PR path) |
| reconcile-release | `33122099253` (`selected_required_run_mismatch`), `33122436137` (reran doomed PR job) |
| Production promotion | none; `main` remains `0d0b0cdf…` |
| Issue-creation pin | `b263c0c110591cc798b89277dfc35542abb1597b` (#167) |
| Protected comparison anchor | `b9e74fc2db4691c48c637639b265d527de9f4505` |

## Scope and non-goals

### In scope

1. Coordinated infrastructure change so reusable `ci.yml` / `run-app-checks.sh`
   select `pr-validation` (exact PR base/head SHAs, merge-base anchored
   hashes) for a same-repository production <- integration promotion PR when
   the recorded capture subject cannot be resolved in the existing checkout.
2. Exact-head check recovery that selects or publishes one unambiguous
   successful exact-head validation instead of rerunning a `pull_request`
   application-check whose provenance classification cannot change on rerun.
3. Deterministic regressions for the #1090 class, ordinary fixture-changing
   PRs that must remain `pr-ancestry`, and negative SHA/hash cases.
4. Caller pin/fixture mirror of the new independently reviewed infrastructure
   merge, caller tests, and current-state docs that already describe this
   contract (`docs/operations/11-devops-and-ci-cd.md`,
   `docs/development/agent-skills.md`, fixture README).
5. After exact-SHA review and merge, `reconcile-release` for #1089 can merge
   #1090 (or the live promotion at the then-current `develop` head) and
   converge `develop` to that exact merge SHA.

### Non-goals / explicitly excluded

- Fetching, hydrating, or materializing the historical capture subject.
- Switching the promotion PR application check to `--squash-safe-push`
  (that path unsets exact PR SHAs and cannot enforce the required hash/SHA
  negatives).
- Weakening ordinary (non-promotion) PR `pr-ancestry` when the capture
  fixture is added, modified, or deleted.
- Weakening the required `ci / ci` check, fabricating unbacked statuses, or
  bypassing rulesets.
- Editing the eight VOC-112 no-change paths relative to `b9e74fc2…`.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- Rewriting VOC-112 through VOC-137 package records or manufacturing their
  completion markers.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review; promotion/closure are ordinary release-path
  evidence, not a VOC-097 live-evidence gate.
- Splitting provenance repair, recovery repair, pin/fixture, tests, docs, or
  release reconciliation into separate tasks.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable CI provenance selection and required-check recovery; the eight
  VOC-112 no-change paths (protected *against* change relative to
  `b9e74fc2…`).
- Protected technical effect: whether a same-repository promotion PR can pass
  required application checks when a squash-era capture subject is not in
  the synthetic checkout, and whether recovery may rerun a job whose
  provenance decision cannot change. No application runtime effect is
  intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-test changes still require
  exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-138-D00`: This is one outcome-sized promotion-CI repair. Use one
end-to-end implementation task covering the coordinated infrastructure PR,
caller pin/fixture mirror, tests, docs, evidence, and release handoff.
Repository count, provenance-versus-recovery, and tests-versus-docs are not
split reasons. Merging #1090 after this correction is evidence of this
outcome, not a second task and not a snapshot of the develop/main gap.

`VOC-138-D01`: Promotion identification. A pull-request application check is
a promotion check when it is the same-repository pair whose base ref is the
configured production branch and whose head ref is the configured
integration branch. In this repository that pair is `main` <- `develop`.
Reusable `ci.yml` must pass an explicit promotion signal (for example
`--promotion-pr`) only for that pair. Forks and any other base/head pair
must not receive the signal. Missing or conflicting exact PR SHAs remain
fail-closed.

`VOC-138-D02`: Promotion provenance when the subject is unreachable. When
the promotion signal is present, exact PR SHAs are valid, and the capture
fixture's recorded `subject_revision` commit object cannot be resolved with
`git cat-file -e <subject>^{commit}` in the existing checkout, select
`pr-validation` (keep exporting `PR_BASE_SHA` / `PR_HEAD_SHA`). Do not
fetch. If the subject *is* resolvable, existing fixture-diff behavior may
remain (`pr-ancestry` when the fixture changed; `pr-validation` when it did
not).

`VOC-138-D03`: Ordinary PRs stay fail-closed. When the promotion signal is
absent, fixture add/modify/delete continues to select `pr-ancestry`. An
ordinary fixture-changing PR whose subject cannot be resolved still fails
closed with `PR ancestry mode requires every captured commit object`. Do not
treat "subject missing" as a general escape hatch.

`VOC-138-D04`: Keep the existing `pr-validation` contract. Unchanged capture
fixtures with valid PR SHAs still select `pr-validation`. Tampered
merge-base, tampered current hashes, missing `PR_BASE_SHA`, missing
`PR_HEAD_SHA`, malformed SHAs, and unrelated base/head pairs still fail
closed. Comparison errors from `git diff` of the capture fixture still fail
closed.

`VOC-138-D05`: Do not switch promotion PRs to `--squash-safe-push`. That
mode is correct for non-PR dispatch/push recovery and is why run
`33122158425` passed. The required PR check must keep exact base/head SHAs
so hash/SHA negatives remain enforceable. `docs/operations/11-devops-and-ci-cd.md`
currently claims the promotion PR uses `squash-safe-push`; this package
updates that sentence to the `pr-validation` contract in the same caller PR.

`VOC-138-D06`: No evidence hydration. `run-app-checks.sh`, caller scripts,
and tests must not `git fetch` the subject, add a hydrate/materialize
helper, wrap `VOC112_CAPTURE_PROVENANCE_MODE`, stamp evidence at test time,
or skip the named VOC-112 tests.

`VOC-138-D07`: Exact-head recovery. Recovery must not rerun a selected
`pull_request` `ci / ci` job when that job's provenance classification
cannot change on rerun and a successful exact-head application-check of the
same required workflow already exists (including `workflow_dispatch` run
`33122158425` at head `87f0efcb…`). In that class, select or publish one
unambiguous successful validation: extend the VOC-114-D07 same-SHA
attestation path so genuine exact-head success from the expected workflows
counts even when the GitHub-selected required row is a structurally doomed
`pull_request` run. Identity mismatches that are not this class still raise
`selected_required_run_mismatch`. Do not fabricate statuses without genuine
exact-head success. Do not bypass the ruleset.

`VOC-138-D08`: Required regression. A deterministic test must construct a
`main` <- `develop` style checkout where the capture fixture differs, the
recorded subject commit object is absent, exact PR SHAs are valid, and the
promotion signal is set. Expected: provenance mode `pr-validation`, and the
existing `assertCapturedRevision` path used by `VOC-112-TEST-12` and
`VOC-112-TEST-13` passes without fetching the subject. A parallel ordinary-PR
fixture-changed checkout with the same missing subject and no promotion
signal must still select `pr-ancestry` and fail closed.

`VOC-138-D09`: Eight VOC-112 no-change paths remain byte-identical to
protected comparison anchor `b9e74fc2db4691c48c637639b265d527de9f4505` and
must be absent from the implementation diff against that SHA:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
- `scripts/foundation/voc112-navigation-benchmark.test.mjs`
- `scripts/foundation/voc112-navigation-benchmark-run.mjs`
- `scripts/foundation/validate-workspace.mjs`
- `AGENTS.md`
- `.agents/skills/vocanova-repo-navigator/SKILL.md`
- `package.json`

JSON `subject_revision` remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
Do not retarget, recapture, or weaken those files. Mode selection changes
belong in `run-app-checks.sh`, not in the provenance test.

`VOC-138-D10`: Pin advance. Issue-creation pin
`b263c0c110591cc798b89277dfc35542abb1597b` (#167) is the defective live
contract for this failure class. T00 opens one new `KARSIFT/karsift-ai-infra`
PR, obtains independent exact-revision review, and after that merge sets
`PINNED_SHA.txt` and every changed mirrored fixture file to that exact merge.
Mirror at least `config/run-app-checks.sh`, `.github/workflows/ci.yml`, the
recovery runner/modules touched by D07, and their tests. If exact comparison
proves another authoritative fixture file also changed, mirror it too. Do
not treat the untracked local `karsift-ai-infra/` checkout as this
repository's tracked tree.

`VOC-138-D11`: Docs in the same PR. Update every current-state document that
would otherwise remain false:

- fixture `README.md` PR-context paragraph;
- `docs/operations/11-devops-and-ci-cd.md` promotion-PR provenance sentence
  (replace the incorrect `squash-safe-push` claim with the D02/D05 contract);
- `docs/development/agent-skills.md` sentence that currently says all pull
  requests require each captured commit.

Do not rewrite VOC-135/VOC-136/VOC-137 package records.

`VOC-138-D12`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. Preserve the named
run/job IDs as audit evidence only (no raw logs).

`VOC-138-D13`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`;
- targeted VOC-138 provenance and recovery cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-138-D14`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate the promotion missing-subject case, ordinary
`pr-ancestry` retention, hash/SHA negatives, no-fetch constraint, and
recovery "do not rerun doomed job" behavior. Merge-gate must reject any
mismatch. Committed `t00-evidence.md` records the implementation PR base,
new infra merge, mode-selection change, recovery change, and the contract
that later exact-head binding is published as review/check metadata. A
tracked file must not be required to contain the SHA of the same commit that
contains it.

`VOC-138-D15`: Protected comparison versus implementation PR base.
Protected comparison anchor for the eight VOC-112 paths remains
`b9e74fc2…`. Issue-creation `develop` is `87f0efcb…`. Implementation must
resolve current `develop` to a 40-character SHA before any in-scope edit and
record that SHA as the implementation PR base. Fail closed on
unrelated/material movement of `develop` (any tree change outside this
package directory, in-scope fixture/pin/tests, and the named current-state
docs). This package's own plan/adoption/roster commits after `87f0efcb…` are
governance-only and do not count as protected-file drift.

`VOC-138-D16`: Release handoff. After the exact reviewed caller merge,
ordinary `reconcile-release` for #1089 may merge #1090 once that PR's head
includes this repair (GitHub updates a `develop` head automatically) or a
successor same-repository promotion PR at the then-current `develop` tip.
`develop` is advanced to the exact promotion merge SHA before audit close.
Do not snapshot the current gap. Closed state alone is not completion
proof. Preserve runs `33122154521`, `33122158425`, `33122099253`,
`33122436137` and jobs `98691441027`, `98692552949` as audit evidence.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents a
promotion deadlock without creating a fetch-based evidence-hydration path
that could mask missing capture ancestry on ordinary PRs. Ordinary
fixture-changing PRs remain `pr-ancestry` fail-closed. Recovery still
requires genuine exact-head Actions success before any same-SHA attestation.

Abuse/process risks:

1. Switching promotion PRs to `--squash-safe-push` so hash/SHA negatives no
   longer apply — forbidden by `VOC-138-D05`.
2. Falling back to `pr-validation` for every missing subject, including
   ordinary PRs — forbidden by `VOC-138-D03`.
3. Fetching or hydrating `f9d11e23…` — forbidden by `VOC-138-D06`.
4. Editing the eight VOC-112 no-change paths to make `pr-ancestry` succeed —
   forbidden by `VOC-138-D09`.
5. Rerunning the doomed `pull_request` job as the recovery strategy —
   forbidden by `VOC-138-D07`.
6. Fabricating statuses without genuine exact-head success — forbidden by
   `VOC-138-D07`.
7. Snapshotting the develop/main gap — forbidden by `VOC-138-D16` and
   `karsift-ai-infra#15`.
8. Changing `roles.yml` or adding an OpenAI route — forbidden by
   `VOC-138-D12`.
9. Printing credentials or copying full CI logs into evidence — forbidden.
10. Requiring a commit to contain its own SHA — forbidden by `VOC-138-D14`.

## Contradictions and open questions

1. **Doc versus code versus issue:**
   `docs/operations/11-devops-and-ci-cd.md` says the canonical `develop` →
   `main` promotion PR uses `squash-safe-push`. Live `ci.yml` uses exact PR
   SHAs for every `pull_request` event. Issue #1091 requires
   merge-base/hash-bound `pr-validation` with exact SHAs, not
   `squash-safe-push`. This package follows the issue (`VOC-138-D05`) and
   updates the doc.
2. **Promotion-signal mechanism:** D01 requires an explicit promotion
   signal from reusable `ci.yml` for the configured `main`/`develop` pair.
   If implementation proves those branch names must be inputs rather than
   literals, that is compatible as long as forks and other pairs cannot
   receive the signal. Do not infer promotion solely from "subject missing".
3. **Live caller workflows:** no live `.github/workflows/pipeline.yml` edit
   is expected, because the caller already consumes `ci.yml@main` and
   `release.yml@main`. If dispatch-surface comparison proves a caller
   workflow must change, record that in `t00-evidence.md` and keep the
   change inside T00.
4. **PR #1090 head movement:** after this package merges, GitHub will move
   #1090's `develop` head. Recovery/promotion evidence may name a later
   exact head. The named 2026-08-27 run/job IDs remain the incident audit
   record.
5. **New infrastructure merge SHA:** not available at drafting time; record
   it after the coordinated infra PR merges. Implementation writes that SHA
   into `PINNED_SHA.txt` and `t00-evidence.md`. Do not invent it at planning
   time.
