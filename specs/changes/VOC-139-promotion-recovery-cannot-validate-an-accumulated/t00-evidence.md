# VOC-139-T00 — Evidence

Task: `VOC-139-T00` — Unblock accumulated promotion hash validation and
no-checkout recovery metadata.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

Committed evidence in this file records the implementation PR base, the new
infrastructure merge, the promotion head-hash rule, the no-checkout metadata
change, and validation. It does **not** require this file, in the same Git
commit, to contain that commit's own SHA. The live implementation head is
bound by the App-authored independent-review comment/check on the
implementation PR. Merge-gate must reject any mismatch. Post-merge /
root-issue audit may record the reviewed head and merge SHA.

Independent review must explicitly evaluate the accumulated-promotion
head-hash case, ordinary merge-base `pr-validation` retention, no-checkout
metadata, identity negatives, no-fetch/no-recapture constraint, and
seven-path freeze before merge.

## Discovery recorded at planning time (issue #1096)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1096 |
| VOC-138 caller merge | PR #1095 `4812fb91ab1b674f9a9ec03906f90c0edf50421d` |
| Promotion PR | #1090 |
| Release issue | #1089 |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `4812fb91ab1b674f9a9ec03906f90c0edf50421d` |
| Selected mode | `pr-validation` (VOC-138 intent; not a mode-selection miss) |
| Defect 1 | `assertPrValidationMergeBase` requires stored hashes to equal merge-base / `main` files |
| Fail-closed message | `AGENTS.md hash must be anchored in the PR merge base` |
| Stored/head hash prefix | `b0e65629…` |
| Main/base hash prefix | `5ba216ff…` |
| Failing tests | `VOC-112-TEST-12`, `VOC-112-TEST-13` |
| PR run / job | `33130426061` / `98718413924` |
| Recovery dispatch | `gh workflow run pipeline.yml --ref develop -f action=reconcile-release -f release_issue_number=1089` |
| Release run | `33130473438` |
| Recovery run | `33130527834` (`promotion-pr-validation PR #1090`) |
| Recovery job | `98718739912` |
| Defect 2 | `gh pr view "$PROMOTION_PR_NUMBER"` with no `-R`/`--repo` and no checkout → `fatal: not a git repository` |
| Production promotion | none; `main` remains `0d0b0cdf…` |
| Issue-creation pin (VOC-138 infra) | `123735c80fec813a5b46a004f3e1122bd425cde2` |
| Protected comparison anchor (seven VOC-112 paths) | `b9e74fc2db4691c48c637639b265d527de9f4505` |
| Doc contradiction | current-state docs claim promotion PRs are merge-base/hash-bound |
| Why bootstrap is not required | T00's first run is attempt `1` on a new VOC-139 carrier from current `develop` |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-139 branch/PR from current `develop`; not a rewrite of PR #1090 |
| Infra | open a new independently reviewed `KARSIFT/karsift-ai-infra` PR; pin its exact merge; do not freeze defective `123735c80…` |
| Promotion identification | existing `--promotion-pr` signal from reusable `ci.yml` only for same-repository `main` ← `develop` |
| Promotion hashes | authenticated promotion pair → `pr-validation` with hashes bound to `PR_HEAD_SHA` and the working tree; do not require historical `main` hashes; do not use `--squash-safe-push` |
| Ordinary PRs | unchanged-fixture `pr-validation` stays merge-base anchored; fixture add/modify/delete stays `pr-ancestry` |
| Negatives | malformed SHA, unrelated commits/repository/PR, wrong refs, tampered current hashes still fail closed |
| Hydration / recapture | **forbidden** |
| Metadata | `gh pr view … -R "$GITHUB_REPOSITORY"`; no checkout; subprocess test in a non-git directory |
| Recovery | retain VOC-138 PR-bound `pr-validation` attestation; do not rerun doomed `pull_request` jobs |
| Fixture freeze versus advance | issue-creation pin `123735c80…` is historical; live pin becomes the new infra merge |
| Seven no-change paths | byte-identical to `b9e74fc2…`; provenance test is the allowed exception |
| VOC-138 records | do not rewrite `specs/changes/VOC-138-…/`; narrow live `NO_CHANGE_PATHS` only |
| Docs | update fixture README, `docs/operations/11-devops-and-ci-cd.md`, and `docs/development/agent-skills.md` in the same caller PR |
| Exact-head binding | App-authored independent-review comment/check binds the live PR head; merge-gate rejects mismatch; a commit must not be required to contain its own SHA |
| Attempt | VOC-139-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged: implementer/escalation `cursor/composer-2.5`; planner/reviewer/reviewer_fast_retry/plan_reviewer `cursor/grok-4.6[effort=high,fast=false]` |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| VOC-097 live-evidence task? | **no** — evidence-carrier sha_lineage cannot bind to #1090 recovery HEAD |
| Bootstrap | none |
| Secrets | do not print credential values; do not copy full CI logs |
| Evidence truth | name the implementation PR base and new infra merge; bind the live head via App-authored review/check; do not mutate this file at test/runtime |

## Changed surfaces (implementation)

Implementation PR base recorded before the first in-scope edit:

pending — resolve current `develop` to a 40-character SHA at dispatch.

Issue-creation develop was
`4812fb91ab1b674f9a9ec03906f90c0edf50421d`. Plan/adoption/roster commits
after that SHA are governance-only and do not count as protected-file drift.

Protected comparison anchor for seven-path VOC-112 boundary:
`b9e74fc2db4691c48c637639b265d527de9f4505`.

Issue-creation infra pin (defective for this failure class; to be replaced):
`123735c80fec813a5b46a004f3e1122bd425cde2`.

New independently reviewed infra merge:
pending — record after the coordinated infra PR merges. Do not invent it.

In-scope implementation diff paths (expected; record actuals after commit):

- `KARSIFT/karsift-ai-infra` `config/run-app-checks.sh` (promotion hash-anchor export)
- `KARSIFT/karsift-ai-infra` `templates/project-repo/.github/workflows/pipeline.yml` (`-R "$GITHUB_REPOSITORY"`)
- `KARSIFT/karsift-ai-infra` tests for export, no-checkout metadata, and identity negatives
- caller `scripts/foundation/voc112-navigation-benchmark.test.mjs` (head-bound promotion hashes)
- caller `.github/workflows/pipeline.yml` (`promotion-pr-metadata`)
- caller `tooling/governance/fixtures/karsift-ai-infra/**` pin and mirrors
- caller `tooling/governance/tests/` (VOC-139 regressions; VOC-138 `NO_CHANGE_PATHS` narrowing)
- `docs/operations/11-devops-and-ci-cd.md`
- `docs/development/agent-skills.md`
- fixture `README.md`
- `specs/changes/VOC-139-promotion-recovery-cannot-validate-an-accumulated/t00-evidence.md`

## Hash-rule and metadata results

| Case | Expected result |
|------|-----------------|
| Promotion PR, `AGENTS.md` differs from `main`, exact SHAs, promotion signal | `pr-validation`; hashes bind to `PR_HEAD_SHA`; `VOC-112-TEST-12` / `VOC-112-TEST-13` pass |
| Promotion PR, subject absent or non-ancestor | still `pr-validation`; named tests pass |
| Ordinary PR, hashes differ from merge-base, no promotion signal | fail closed: `AGENTS.md hash must be anchored in the PR merge base` |
| Ordinary PR, fixture changed, subject present | `pr-ancestry` |
| Ordinary PR, fixture changed, subject absent | `pr-ancestry` fail-closed |
| Unchanged fixture, valid SHAs, no promotion signal | merge-base-anchored `pr-validation` |
| Tampered current hashes, or promotion hashes matching only `main` | fail closed |
| Missing or malformed PR SHAs / unrelated commits | fail closed |
| Wrong refs / unrelated repository or PR | fail closed |
| Metadata `gh pr view` in a non-git directory with `-R` | succeeds (mock/fixture `gh`) |
| Metadata `gh pr view` without `-R` in a non-git directory | fails (`not a git repository` or equivalent) |
| `git fetch` / hydrate / recapture | absent |
| Failed PR job plus weaker squash-safe dispatch | still rejected as insufficient |
| Other selected-run identity mismatch | `selected_required_run_mismatch` |

## Validation commands (implementation)

Record after the repair is tracked and committed. Do not treat an
untracked-only pass as acceptance.

```bash
bash scripts/governance/validate-governance.sh \
  --base <implementation-pr-base> \
  --head <implementation-pr-head>
bash scripts/governance/classify-change-risk.sh \
  --base <implementation-pr-base> \
  --head <implementation-pr-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
git diff --check
```

`classify-change-risk.sh` is expected to report R4 on the implementation
range.

## Independent verification (implementation)

Pending exact-SHA independent review of the infrastructure PR and the caller
implementation PR. The App-authored independent-review comment/check must
bind the live PR head exactly and must explicitly evaluate the
accumulated-promotion head-hash case, ordinary merge-base retention,
no-checkout metadata, identity negatives, no-fetch/no-recapture constraint,
and seven-path freeze. Merge-gate must reject any mismatch. Record a pointer
to that comment/check here after it exists; do not write the live head SHA
into this file as a value that the same commit is required to contain. The
implementer must not approve or merge its own work.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- Genuine staging runs only for the real caller tree change. Tree-equivalent
  post-promotion develop synchronization must not keep staging scheduled.
- `reconcile-release` for #1089 may merge #1090 (or the live same-repository
  promotion at the then-current `develop` head) once required `ci / ci`
  passes under the repaired hash and metadata contract.
- Recovery must emit a successful exact run/repository/PR/base/head/ref/path
  attestation. Record allowlisted metadata only (workflow identity, event,
  branch, HEAD SHA, run/job IDs, conclusion).
- `develop` is advanced to the successful promotion merge SHA before audit
  close and ends 0 ahead / 0 behind `main` with an identical tree.
- Every promotion merge push to `main` triggers automatic production
  deployment, whose exact-SHA result is verified.
- Release/task/requirement records close with audit comments naming the
  exact promotion merge. Post-merge audit may record the independently
  reviewed head and the promotion merge SHA.
- Root issue #1096 closes only after that allowlisted recovery/release
  metadata exists.
- Closed state alone is not completion proof.
- Incident runs `33130426061`, `33130473438`, `33130527834` and jobs
  `98718413924`, `98718739912` remain audit context only.
