# VOC-140-T00 — Evidence

Task: `VOC-140-T00` — Unblock release-convergence CI identity and
production-merge-guard App-token visibility.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

Committed evidence in this file records the implementation PR base, the new
infrastructure merge, the recovery-identity change, the token/API contract,
and validation. It does **not** require this file, in the same Git commit, to
contain that commit's own SHA. The live implementation head is bound by the
App-authored independent-review comment/check on the implementation PR.
Merge-gate must reject any mismatch. Post-merge / root-issue audit may
record the reviewed head and merge SHA.

Independent review must explicitly evaluate the circular-CI identity case,
dedicated promotion-pr-validation requirement, omitted-`bypass_actors` token
shape, empty-bypass acceptance, non-empty-bypass rejection,
exact two-token permissions/repository scope/use isolation in both workflows,
external App approval and hosted explicit-empty-bypass proof, current-state
documentation reconciliation, same-App/private-key residual risk,
no-fabricated-status constraint, and pin advance before merge.

## Discovery recorded at planning time (issue #1102)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1102 |
| Promotion PR | #1090 |
| Release issue | #1089 |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `21eef75549226766fc4f78f62f232ee5fbdb8d6d` |
| Issue-creation pin | `599436835371f27fac52ec6b47a18b36257366ac` |
| Defect 1 | Recovery reports complete without dispatch; selector chooses `ci / ci` from the still-running workflow that contains `release / converge` |
| Circular-CI run / job | `33136633666` / `98738317266` |
| Circular-CI fail-closed | `untrusted_ci_recovery_identity` |
| Dedicated recovery success | run `33136865709` |
| Defect 2 | Merge App token cannot prove full ruleset/no-bypass fields; validator raises `production_merge_guard_missing` while admin-visible verifier reports `ok` |
| Guard-fail release runs / jobs | `33136984634` / `98739074178`; `33137091931` / `98739420310` |
| Guard fail-closed | `production-merge-guard: production_merge_guard_missing` |
| Admin-visible command | `bash tooling/governance/fixtures/karsift-ai-infra/config/verify-production-merge-guard.sh KARSIFT/vocanova-platform-sandbox main tooling/governance/fixtures/karsift-ai-infra/config` |
| Admin-visible result | `production-merge-guard: ok ruleset_id=20575146` |
| Live ruleset | `20575146`; repository-owned; active; `main`-only; strict; requires PRs and `governance-policy` / `validate` / `ci`; `bypass_actors: []` |
| App-token mint | `permission-contents: write`, `permission-issues: write`, `permission-pull-requests: write` |
| Production promotion | none; `main` remains `0d0b0cdf…` |
| Why bootstrap is not required | T00's first run is attempt `1` on a new VOC-140 carrier from current `develop` |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-140 branch/PR from current `develop`; not a rewrite of PR #1090 |
| Infra | open a new independently reviewed `KARSIFT/karsift-ai-infra` PR; pin its exact merge; do not freeze defective `59943683…` |
| Circular CI | never select an in-progress/failed release carrier as attestable `ci / ci`; recovery must not report complete without dispatch on that class |
| Dedicated recovery | dispatch or select `promotion-pr-validation PR #<n>`; require completed/success before attestation |
| Production guard | still requires effective active repository-owned ruleset, pull-request rule, strict non-empty required checks, and `bypass_actors: []` |
| Token/API | mutation token remains exactly Contents/Issues/Pull requests write and sole for merge/mutations; distinct current-repository-scoped Administration-write-only guard token is used only for guard verification immediately before merge in both workflows |
| External activation | `karsift-ai-infra-bot` Administration: Read and write plus owner approval on KARSIFT organization installation `148001476`; installation currently selects all repositories, while guard token stays explicitly caller-repository scoped; no secret rotation; hosted explicit `bypass_actors: []` proof required |
| Tests | real token-visible omitted-field payload and circular-CI parent-run fixture; not helper-only full fixtures |
| Fixture freeze versus advance | issue-creation pin `59943683…` is historical; live pin becomes the new infra merge |
| VOC-139 / VOC-138 records | do not rewrite `specs/changes/VOC-139-…/` or `VOC-138-…/` |
| Docs | exhaustively search source/pin claims; update fixture README, operations CI/CD, repository-settings, activation checklist, DOC-19, and any other current-state matches while preserving historical records |
| Exact-head binding | App-authored independent-review comment/check binds the live PR head; merge-gate rejects mismatch; a commit must not be required to contain its own SHA |
| Attempt | VOC-140-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged: implementer/escalation `cursor/composer-2.5`; planner/reviewer/reviewer_fast_retry/plan_reviewer `cursor/grok-4.6[effort=high,fast=false]` |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Duplicate promotion PR / audit? | **no** |
| VOC-097 live-evidence task? | **no** — evidence-carrier sha_lineage cannot bind to #1090 recovery HEAD |
| Bootstrap | none |
| Secrets | do not print credential values; do not copy full CI logs |
| Evidence truth | name the implementation PR base and new infra merge; bind the live head via App-authored review/check; do not mutate this file at test/runtime |
| Residual risk | both tokens share one App/private key and its permission ceiling; optional dedicated guard App is future hardening, not T00 |

## Changed surfaces (implementation)

Implementation PR base recorded before the first in-scope edit:

pending — resolve current `develop` to a 40-character SHA before any
in-scope edit.

Issue-creation develop was
`21eef75549226766fc4f78f62f232ee5fbdb8d6d`. Plan/adoption/roster commits
after that SHA are governance-only and do not count as protected-file drift.

Issue-creation infra pin (defective for this failure class; historical audit):
`599436835371f27fac52ec6b47a18b36257366ac`.

New independently reviewed infra merge:

pending — record after the coordinated infra PR merges.

In-scope implementation diff paths (expected; record actuals after commit):

- `KARSIFT/karsift-ai-infra` recovery/selection/attestation sources
- `KARSIFT/karsift-ai-infra` `.github/workflows/release.yml` mutation and guard-token mints
- `KARSIFT/karsift-ai-infra` `.github/workflows/merge-gate.yml` production-branch mutation/guard separation
- `KARSIFT/karsift-ai-infra` `config/production_merge_guard.py` and `config/verify-production-merge-guard.sh`
- `KARSIFT/karsift-ai-infra` tests for circular-CI identity and token-visible payload shape
- caller `tooling/governance/fixtures/karsift-ai-infra/**` pin and mirrors
- caller `tooling/governance/tests/` (VOC-140 regressions; current-pin assertions)
- `docs/operations/11-devops-and-ci-cd.md`
- `docs/governance/repository-settings.md`
- `docs/governance/post-merge-activation-checklist.md`
- `docs/operations/19-governance-reconciliation-notes.md`
- `scripts/governance/validate-governance.sh` and relevant tests if stale assertions require reconciliation
- fixture `README.md`
- `specs/changes/VOC-140-release-convergence-cannot-trust-its-own-ci-run/t00-evidence.md`

## Identity and token/API results

| Case | Expected result |
|------|-----------------|
| Newest `ci / ci` SUCCESS on still-running `pipeline.yml` that contains `release / converge` | not attestable; recovery not complete; dedicated `recover-promotion-pr-checks` dispatched |
| Failed or queued release carrier | not attestable; does not suppress dedicated recovery |
| Completed `promotion-pr-validation PR #<n>` exact-head/PR-bound | attestable `ci / ci`; recovery complete; redispatch suppressed |
| Same-head `display_title: pipeline` / `--squash-safe-push` | still rejected |
| Doomed `pull_request` `ci / ci` | still not rerun as a strategy |
| Administrator-visible ruleset `20575146` shape with `bypass_actors: []` | `production-merge-guard: ok ruleset_id=…` |
| App-token-visible ruleset JSON with omitted/non-array `bypass_actors` | distinct `production_merge_guard_payload_incomplete` (or live equivalent), not `production_merge_guard_missing` |
| Non-empty `bypass_actors` | still fail closed |
| Mutation token | exactly Contents/Issues/Pull requests write; sole App token for `gh pr merge` and mutations; no Administration |
| Guard token | exactly Administration write; explicit owner/current caller repository scope; only guard verification; never merge/mutation/status/issues/PR/content |
| Workflow order | guard-only mint and verification immediately precede fresh exact-head/base/ref revalidation and mutation-token merge in release and production merge-gate paths |
| External App activation | pending — record Administration: Read and write registration plus owner approval on KARSIFT organization installation `148001476`; record current `repository_selection: all`, explicit caller-repository runtime token scope, and no secret rotation |
| Hosted guard proof | pending — record sanitized explicit `bypass_actors: []`; omission/non-array/nonempty fail closed |

## Exhaustive source and pin reconciliation

Pending. Record the exact tracked-source searches for the old/current pin,
`CURRENT_PIN` / `AUTHORITATIVE_PIN`, mirrored-file hashes, mutation-only token
claims, active-A-003 claims, and disabled/unimplemented automatic-release or
production-deploy claims. List every match and disposition: updated current
state, preserved clearly historical audit evidence, or irrelevant. At minimum
include fixture README, `docs/operations/11-devops-and-ci-cd.md`,
`docs/governance/repository-settings.md`, the activation checklist, DOC-19,
governance validators that enforce their wording, caller pin-lock tests, and
foundation pin assertions.

## Credential residual risk

Pending confirmation. Record that step-level token isolation does not split
the underlying `karsift-ai-infra-bot` registration/private key: compromise can
still mint up to installation `148001476`'s combined permission ceiling across
its current `repository_selection: all` scope. A separately keyed,
single-repository Administration-only guard App is optional future hardening
and is not implemented or authorized by VOC-140-T00.

## Validation commands (implementation)

Pending. Record against the implementation PR base and the corrected
supervised implementation range after the repair is tracked and committed.
The final exact caller head is bound by the App-authored independent-review
comment/check rather than written into the commit that it identifies.

```bash
bash scripts/governance/validate-governance.sh \
  --base <implementation-pr-base> \
  --head HEAD
bash scripts/governance/classify-change-risk.sh \
  --base <implementation-pr-base> \
  --head HEAD
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

## Independent verification (implementation)

Pending exact-SHA independent review of the infrastructure PR and the caller
implementation PR. The App-authored independent-review comment/check must
bind the live PR head exactly and must explicitly evaluate the circular-CI
identity case, dedicated promotion-pr-validation requirement,
omitted-`bypass_actors` token shape, empty-bypass acceptance,
non-empty-bypass rejection, exact token permissions/scope/use isolation,
guard-before-merge order, external App approval/hosted proof, current-state
documentation, same-App residual risk, no-fabricated-status constraint, and
pin advance. Merge-gate must reject any mismatch. Record a pointer to that
comment/check here after it exists; do not write the live head SHA into this
file as a value that the same commit is required to contain. The implementer
must not approve or merge its own work.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- Genuine staging runs only for the real caller tree change. Tree-equivalent
  post-promotion develop synchronization must not keep staging scheduled.
- Rerun dedicated `recover-promotion-pr-checks` for #1090 if necessary.
- `reconcile-release` for #1089 may merge #1090 (or the live same-repository
  promotion at the then-current `develop` head) once recovered `ci / ci` is a
  completed non-carrier run and the isolated guard token can prove the live
  production merge guard before the mutation-token merge.
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
- Root issue #1102 closes only after that allowlisted recovery/release
  metadata exists.
- Closed state alone is not completion proof.
- Incident runs `33136633666`, `33136865709`, `33136984634`, `33137091931`
  and jobs `98738317266`, `98739074178`, `98739420310` remain audit context
  only.
