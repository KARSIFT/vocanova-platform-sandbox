# VOC-139 — Promotion recovery cannot validate an accumulated develop-to-main PR: Specification

## Objective and requirement source

Unblock same-repository `main` ← `develop` promotion PR validation when VOC-112
source hashes are truthfully bound to the reviewed promotion head but differ
from historical `main`. Make promotion-check recovery resolve immutable PR
metadata without a git checkout. Ordinary non-promotion `pr-validation` and
`pr-ancestry` fail-closed behavior must stay intact.

**Requirement source:** [GitHub issue #1096](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1096).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1096)

| Item | Value |
|------|-------|
| VOC-138 caller merge | [PR #1095](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1095) `4812fb91ab1b674f9a9ec03906f90c0edf50421d` |
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090) |
| Release issue | [#1089](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1089) |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `4812fb91ab1b674f9a9ec03906f90c0edf50421d` |
| Selected mode | `pr-validation` (VOC-138 intent; this is not a mode-selection miss) |
| Failing tests | `VOC-112-TEST-12`, `VOC-112-TEST-13` |
| Fail-closed message | `AGENTS.md hash must be anchored in the PR merge base` |
| Stored/head hash prefix | `b0e65629…` |
| Main/base hash prefix | `5ba216ff…` |
| PR run / job | `33130426061` / `98718413924` |
| Release run | `33130473438` |
| Recovery run / job | `33130527834` / `98718739912` |
| Recovery failure | `fatal: not a git repository` at `gh pr view "$PROMOTION_PR_NUMBER"` |
| Production promotion | none; `main` remains `0d0b0cdf…` |
| Issue-creation pin | `123735c80fec813a5b46a004f3e1122bd425cde2` |
| Protected comparison anchor | `b9e74fc2db4691c48c637639b265d527de9f4505` (seven remaining no-change paths) |

## Scope and non-goals

### In scope

1. Promotion-specific immutable source-hash/provenance rule for authenticated
   same-repository `main` ← `develop` `pr-validation`: stored capture hashes
   bind to the reviewed head/source revision (`PR_HEAD_SHA` and working tree)
   without requiring historical `main` to contain those hashes. Exact PR
   base/head SHAs remain required, and the base must be an ancestor of the
   head.
2. Repository-explicit no-checkout metadata (prefer the pull REST response
   addressed through `repos/$GITHUB_REPOSITORY`, or supported `gh pr view`
   owner/repository fields with `-R "$GITHUB_REPOSITORY"`)
   in live caller `.github/workflows/pipeline.yml`, the infrastructure
   template, and the mirrored fixture, plus a subprocess/end-to-end test that
   executes that step with no git repository.
3. Deterministic regressions for an accumulated promotion whose `AGENTS.md`
   (and navigator skill) differ between base and head, ordinary
   merge-base-anchored `pr-validation`, malformed SHA, unrelated
   repository/PR, wrong refs, unrelated commits, tampered current hashes,
   and missing/nonancestor subject cases.
4. Caller pin/fixture mirror of the new independently reviewed infrastructure
   merge, caller tests (including narrowing the live VOC-138 `NO_CHANGE_PATHS`
   list so the provenance test is no longer frozen), and current-state docs
   that would otherwise keep claiming merge-base hash anchoring for promotion
   PRs.
5. After exact-SHA review and merge, `reconcile-release` for #1089 can merge
   #1090 (or the live promotion at the then-current `develop` head) and
   converge `develop` to that exact merge SHA. Closure of #1096 binds
   allowlisted metadata from the successful recovery/release run.

### Non-goals / explicitly excluded

- Fetching, hydrating, or recapturing VOC-112 JSON fixtures or hashed sources.
- Switching the promotion PR application check to `--squash-safe-push`.
- Weakening ordinary (non-promotion) `pr-validation` merge-base hash
  anchoring, or ordinary fixture-changing `pr-ancestry`.
- Weakening the required `ci / ci` check, fabricating unbacked statuses, or
  bypassing rulesets.
- Editing the seven remaining VOC-112 no-change paths relative to `b9e74fc2…`.
- Rewriting VOC-138 (or earlier) package records under `specs/changes/`.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- A VOC-097 operator-owned live-evidence second task (see D09 / contradictions).
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable CI provenance export and required-check recovery metadata; live
  `.github/workflows/pipeline.yml`; the provenance test (now in scope to
  change); the seven remaining VOC-112 no-change paths (protected *against*
  change relative to `b9e74fc2…`).
- Protected technical effect: whether an accumulated same-repository
  promotion PR whose hashed sources changed on `develop` can pass required
  application checks, and whether recovery metadata can resolve without a
  checkout. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-test changes still require
  exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-139-D00`: This is one outcome-sized promotion-CI repair. Use one
end-to-end implementation task covering the coordinated infrastructure PR,
caller pin/fixture mirror, live pipeline.yml, provenance-test hash rule,
tests, docs, evidence, and release handoff. Repository count, hash-rule
versus metadata, and tests-versus-docs are not split reasons. Merging #1090
after this correction is evidence of this outcome, not a second task and not
a snapshot of the develop/main gap.

`VOC-139-D01`: Keep VOC-138 promotion identification. A pull-request
application check is a promotion check when it is the same-repository pair
whose base ref is the configured production branch and whose head ref is the
configured integration branch (`main` ← `develop` here). Reusable `ci.yml`
already passes `--promotion-pr` only for that pair. Forks and any other
base/head pair must not receive the signal. Missing or conflicting exact PR
SHAs remain fail-closed.

`VOC-139-D02`: Promotion source-hash contract. When the authenticated
promotion signal is present, exact PR SHAs are valid, and `PR_BASE_SHA` is
an ancestor of `PR_HEAD_SHA`,
`VOC112_CAPTURE_PROVENANCE_MODE` remains `pr-validation` (so VOC-138 recovery
attestation still matches). Stored `source_hashes.agents_sha256` and
`source_hashes.navigator_skill_sha256` must equal the files at `PR_HEAD_SHA`
and the working tree. They must **not** be required to equal the files at
the git merge-base / `PR_BASE_SHA` (`main`). Prefer exporting
`VOC112_PROMOTION_PR=true` (or an equivalent explicit env) from
`run-app-checks.sh` when `--promotion-pr` is set, and reading that signal in
`assertPrValidationMergeBase`. Do not infer promotion solely from "hashes
differ between base and head". A distinct provenance mode name is acceptable
only if recovery attestation is updated in the same task to keep requiring
the promotion-specific contract rather than `--squash-safe-push`.

`VOC-139-D03`: Ordinary `pr-validation` stays merge-base anchored. When the
promotion signal is absent, stored hashes must still equal the PR merge-base
files. An ordinary unchanged-fixture PR whose working-tree hashes do not
match merge-base still fails closed with
`AGENTS.md hash must be anchored in the PR merge base` (or the navigator
equivalent). Do not treat "hashes differ from main" as a general escape
hatch.

`VOC-139-D04`: Identity and hash negatives stay fail-closed. Missing or
malformed `PR_BASE_SHA` / `PR_HEAD_SHA`, a base that is not an ancestor of
the head, tampered current/working-tree hashes, unrelated
repository or PR, and wrong base/head refs still fail closed. Capture-fixture
comparison errors still fail closed. Unchanged fixtures with valid SHAs and
no promotion signal still select merge-base-anchored `pr-validation`.
Promotion missing-subject and resolvable-but-non-ancestor subject cases
continue to pass under `pr-validation` (VOC-138 retention). Ordinary
fixture-changing PRs without the promotion signal remain `pr-ancestry` and
still fail closed when the subject is missing.

`VOC-139-D05`: Do not switch promotion PRs to `--squash-safe-push`. Exact
base/head SHAs remain required so hash/SHA negatives remain enforceable.

`VOC-139-D06`: No evidence hydration or recapture. `run-app-checks.sh`,
caller scripts, and tests must not `git fetch` the subject, add a
hydrate/materialize helper, wrap provenance mode as a skip, stamp evidence at
test time, or rewrite the two VOC-112 JSON fixtures. JSON `subject_revision`
remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.

`VOC-139-D07`: No-checkout recovery metadata is repository-explicit and uses
fields actually present in the live GitHub response. Job
`promotion-pr-metadata` must address `$GITHUB_REPOSITORY` explicitly. The
preferred form is `gh api "repos/$GITHUB_REPOSITORY/pulls/$PROMOTION_PR_NUMBER"`
and validation of `.base.repo.full_name`, `.head.repo.full_name`, refs, state,
and exact SHAs. An equivalent `gh pr view ... -R "$GITHUB_REPOSITORY"` form is
acceptable only when it derives owner/name from supported fields such as
`headRepositoryOwner.login` plus `headRepository.name`; it must not read the
absent `.headRepository.nameWithOwner`. It must not depend on a checkout or
implicit git remote. Equivalent behavior must land in:

- live caller `.github/workflows/pipeline.yml`
- `KARSIFT/karsift-ai-infra` `templates/project-repo/.github/workflows/pipeline.yml`
- the caller fixture mirror of that template

The job still rejects a closed or missing PR, a fork/same-name repository,
and a non-`main`/`develop` same-repository pair. Missing/malformed SHAs and
wrong repository or PR identity remain fail-closed. A deterministic test must
execute the real metadata step (or the extracted command body it actually
runs) in a directory that is not a git repository, with `GITHUB_REPOSITORY`
set, using a fixture shaped like the real GitHub response in which
`headRepository` has no `nameWithOwner`. It must prove success depends on
explicit repository context and supported identity fields rather than an
implicit git remote or a nonexistent projection.

`VOC-139-D08`: Recovery attestation remains VOC-138's PR-bound
`pr-validation` contract: PR number, repository, immutable base/head SHAs,
configured `main` ← `develop` pair, expected workflow/path, and genuine
success. Same-head `--squash-safe-push` remains insufficient. Doomed
`pull_request` `ci / ci` jobs are still not rerun as a strategy. After
metadata can resolve, the recovery run must be able to emit a successful
exact run/repository/PR/base/head/ref/path attestation that
`reconcile-release` can accept.

`VOC-139-D09`: No VOC-097 live-evidence second task. Successful #1090
recovery HEAD is the promotion/`develop` tip after this repair merges. A
draft evidence-carrier PR cannot satisfy `exact_pr_head` or
`integration_contains_pr_head` against that SHA without treating the carrier
as the promotion PR. Live success is ordinary release/closure evidence
recorded with allowlisted metadata only (workflow identity, event, branch,
HEAD SHA, run/job IDs, conclusion). Do not snapshot the develop/main gap.

`VOC-139-D10`: Narrowed VOC-112 freeze. These seven paths remain
byte-identical to `b9e74fc2db4691c48c637639b265d527de9f4505` and must be
absent from the implementation diff against that SHA:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
- `scripts/foundation/voc112-navigation-benchmark-run.mjs`
- `scripts/foundation/validate-workspace.mjs`
- `AGENTS.md`
- `.agents/skills/vocanova-repo-navigator/SKILL.md`
- `package.json`

`scripts/foundation/voc112-navigation-benchmark.test.mjs` is in scope because
`assertPrValidationMergeBase` is the remaining hash defect. Do not rewrite
VOC-138 package files. Update live
`tooling/governance/tests/test_voc138_promotion_pr_provenance.py` so
`NO_CHANGE_PATHS` no longer includes the provenance test.

`VOC-139-D11`: Pin advance. Issue-creation pin
`123735c80fec813a5b46a004f3e1122bd425cde2` is the defective live contract for
this failure class. T00 opens one new `KARSIFT/karsift-ai-infra` PR, obtains
independent exact-revision review, and after that merge sets
`PINNED_SHA.txt` and every changed mirrored fixture file to that exact merge.
Mirror at least `config/run-app-checks.sh`, the project-repo `pipeline.yml`
template, and their tests. If exact comparison proves another authoritative
fixture file also changed, mirror it too. Do not treat the untracked local
`karsift-ai-infra/` checkout as this repository's tracked tree.

`VOC-139-D12`: Docs in the same PR. Update every current-state document that
would otherwise remain false:

- `docs/operations/11-devops-and-ci-cd.md` promotion-PR sentence (today:
  "merge-base/hash-bound `pr-validation`");
- `docs/development/agent-skills.md` equivalent sentence;
- fixture `README.md` PR-context / VOC-138 pin paragraph.

State that authenticated promotion PRs use `pr-validation` with hashes bound
to the reviewed head/source revision, while ordinary unchanged-fixture PRs
remain merge-base anchored. Do not rewrite VOC-138 package records.

`VOC-139-D13`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values. Preserve the named
run/job IDs as audit evidence only (no raw logs).

`VOC-139-D14`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`;
- targeted VOC-139 hash-rule, no-checkout metadata, and identity-negative cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-139-D15`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate the accumulated-promotion head-hash case, ordinary
merge-base `pr-validation` retention, no-checkout metadata, identity
negatives, no-fetch/no-recapture constraint, and seven-path freeze.
Merge-gate must reject any mismatch. Committed `t00-evidence.md` records the
implementation PR base, new infra merge, hash-rule change, metadata change,
and the contract that later exact-head binding is published as review/check
metadata. A tracked file must not be required to contain the SHA of the same
commit that contains it.

`VOC-139-D16`: Protected comparison versus implementation PR base.
Protected comparison anchor for the seven VOC-112 paths remains `b9e74fc2…`.
Issue-creation `develop` is `4812fb91…`. Implementation must resolve current
`develop` to a 40-character SHA before any in-scope edit and record that SHA
as the implementation PR base. Fail closed on unrelated/material movement of
`develop` (any tree change outside this package directory, in-scope
fixture/pin/tests, live `pipeline.yml`, the provenance test, the VOC-138 live
test-list narrowing, and the named current-state docs). This package's own
plan/adoption/roster commits after `4812fb91…` are governance-only and do not
count as protected-file drift.

`VOC-139-D17`: Release handoff. After the exact reviewed caller merge,
ordinary `reconcile-release` for #1089 may merge #1090 once that PR's head
includes this repair, or a successor same-repository promotion PR at the
then-current `develop` tip. `develop` is advanced to the exact promotion
merge SHA before audit close. Do not snapshot the current gap. Closed state
alone is not completion proof. Preserve runs `33130426061`, `33130473438`,
`33130527834` and jobs `98718413924`, `98718739912` as audit evidence.
Root issue #1096 closes only after allowlisted metadata from the successful
recovery/release run exists.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents a
promotion deadlock without letting ordinary PRs skip merge-base hash
anchoring. Recovery metadata becomes repository-explicit instead of depending
on ambient git context.

Abuse/process risks:

1. Skipping merge-base hash checks for every `pr-validation` PR, including
   ordinary feature PRs — forbidden by `VOC-139-D03`.
2. Inferring promotion solely from "hashes differ from main" — forbidden by
   `VOC-139-D02`.
3. Switching promotion PRs to `--squash-safe-push` — forbidden by
   `VOC-139-D05`.
4. Fetching, hydrating, or recapturing VOC-112 fixtures — forbidden by
   `VOC-139-D06`.
5. Editing the seven remaining VOC-112 no-change paths — forbidden by
   `VOC-139-D10`.
6. Leaving metadata without explicit repository context or relying on the
   absent `.headRepository.nameWithOwner` projection — forbidden by
   `VOC-139-D07`.
7. Snapshotting the develop/main gap or adding a self-invalidating
   snapshot-then-check task — forbidden by `VOC-139-D09` and
   `karsift-ai-infra#15`.
8. Changing `roles.yml` or adding an OpenAI route — forbidden by
   `VOC-139-D13`.
9. Printing credentials or copying full CI logs into evidence — forbidden.
10. Requiring a commit to contain its own SHA — forbidden by `VOC-139-D15`.

## Contradictions and open questions

1. **VOC-138 freeze versus this defect:** VOC-138-D09 froze
   `voc112-navigation-benchmark.test.mjs` against `b9e74fc2…`. The remaining
   #1090 failure is exactly `assertPrValidationMergeBase` in that file. This
   package supersedes that freeze for that one live file (`VOC-139-D10`) and
   does not rewrite VOC-138 package records. Live VOC-138 `NO_CHANGE_PATHS`
   must be narrowed in the same T00.
2. **Docs versus live hash rule:** `docs/operations/11-devops-and-ci-cd.md`
   and `docs/development/agent-skills.md` currently claim promotion PRs use
   merge-base/hash-bound `pr-validation`. That sentence is true of mode
   selection and false of the hash-anchor revision for accumulated
   promotions. This package follows issue #1096 (`VOC-139-D02`) and updates
   those docs.
3. **Promotion-signal plumbing:** D02 prefers `VOC112_PROMOTION_PR=true`
   exported from `run-app-checks.sh` while keeping mode `pr-validation`. If
   implementation proves a distinct mode name is clearer, that is compatible
   only if recovery attestation is updated in the same task. Do not infer
   promotion from hash mismatch alone.
4. **VOC-097 live evidence:** issue #1096 asks to bind the exact successful
   #1090 recovery/release run before closing. That run cannot exist until
   this repair is on `develop`, and a VOC-097 evidence-carrier PR cannot
   satisfy sha_lineage against the promotion HEAD. This package records that
   live run as ordinary release/closure evidence (`VOC-139-D09`,
   `VOC-139-D17`), not as a second task.
5. **PR #1090 head movement:** after this package merges, GitHub will move
   #1090's `develop` head. Recovery/promotion evidence may name a later
   exact head. The named 2026-08-28 run/job IDs remain the incident audit
   record.
6. **New infrastructure merge SHA:** not available at drafting time; record
   it after the coordinated infra PR merges. Implementation writes that SHA
   into `PINNED_SHA.txt` and `t00-evidence.md`. Do not invent it at planning
   time.
