# VOC-138 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable CI provenance selection and required-check recovery; named
  current-state docs. The eight VOC-112 no-change paths are protected against
  change relative to `b9e74fc2db4691c48c637639b265d527de9f4505`.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `b263c0c110591cc798b89277dfc35542abb1597b`. Confirm fixture
  `run-app-checks.sh` still selects `pr-ancestry` whenever the capture fixture
  differs between PR base and head, without inspecting subject resolvability.
  Confirm reusable `ci.yml` still passes `--pr-base-sha`/`--pr-head-sha` for
  every `pull_request` event and `--squash-safe-push` only otherwise. Confirm
  `validate_selected_workflow_run` still requires `event == pull_request` and
  `run_attempt == 1` before rerun. Confirm PR #1090, run `33122154521`, jobs
  `98691441027` / `98692552949`, dispatch `33122158425`, and reconcile-release
  `33122099253` / `33122436137` remain the incident record. Confirm
  `subject_revision` remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
  Confirm the eight VOC-112 paths still match `b9e74fc2…`. Confirm fixture
  `config/roles.yml` still binds implementer/escalation `cursor/composer-2.5`
  and planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `87f0efcb…` do not count as
  protected-file drift. If any of the eight no-change paths differs from the
  anchor, fail closed.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-138
  carrier from current `develop`. Do not reuse PR #1090 as this package's
  implementation PR. Do not treat the untracked local `karsift-ai-infra/`
  checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not fetch or hydrate the capture subject. Do not edit the eight VOC-112
  no-change paths.

## File reconciliation and implementation sequence

### T00 — Unblock promotion PR CI and exact-head recovery when the VOC-112 subject is unreachable

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra` `config/run-app-checks.sh` | modify | Accept an explicit promotion signal. When that signal is set and the recorded subject cannot be resolved, select `pr-validation` with exact PR SHAs. Do not fetch. Ordinary fixture add/modify/delete without the signal stays `pr-ancestry` |
| `KARSIFT/karsift-ai-infra` `.github/workflows/ci.yml` | modify | Pass the promotion signal only for same-repository `main` <- `develop` (or the configured production/integration pair). Keep `--pr-base-sha`/`--pr-head-sha`. Keep `--squash-safe-push` for non-PR events |
| `KARSIFT/karsift-ai-infra` recovery runner / modules | modify | Do not rerun a structurally doomed `pull_request` `ci / ci` job when successful exact-head application-check evidence already exists; select or publish that success (VOC-114-D07 extension). Keep `selected_required_run_mismatch` for other identity failures |
| `KARSIFT/karsift-ai-infra` tests (`test_app_check_context.py`, recovery tests) | extend | Promotion missing-subject → `pr-validation`; ordinary missing-subject → `pr-ancestry` fail-closed; hash/SHA negatives; no `git fetch`; doomed-job rerun refused when exact-head dispatch success exists |
| Caller `tooling/governance/fixtures/karsift-ai-infra/**` | replace from new infra merge | Pin `PINNED_SHA.txt` to that exact merge; mirror every changed authoritative file |
| Caller `tooling/governance/tests/` | extend as needed | Pin assertions and any caller-owned complete-diff coverage; do not exclude `tooling/governance/tests/` from the scan |
| Eight VOC-112 no-change paths | **do not modify** | Must remain byte-identical to `b9e74fc2…` |
| `docs/operations/11-devops-and-ci-cd.md` | modify | Replace the promotion-PR `squash-safe-push` claim with the `pr-validation` contract |
| `docs/development/agent-skills.md` | modify | State that promotion PRs with an unreachable subject use merge-base/hash-bound `pr-validation`; ordinary fixture-changing PRs still require captured-commit ancestry |
| `specs/changes/VOC-135-…/` and `VOC-136-…/` and `VOC-137-…/` | **do not modify** | Audit evidence |
| `.github/workflows/pipeline.yml` | **do not modify unless proven** | Expected: no live workflow edit |
| `specs/changes/VOC-138-.../t00-evidence.md` | update | Record implementation PR base, new infra merge, mode-selection change, recovery change, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Open the coordinated `KARSIFT/karsift-ai-infra` PR from current infra
   `main`. Implement D01–D08 there with tests. Do not treat an untracked
   nested checkout as already-merged work.
3. Obtain independent exact-revision review of that infra PR and merge it.
   Record the exact merge SHA.
4. From current caller `develop`, create a new VOC-138 implementation branch.
   Set `PINNED_SHA.txt` and mirror every changed authoritative fixture file
   from that exact merge. Update the named current-state docs.
5. Confirm no eight-path VOC-112 file and no VOC-135/136/137 package file is
   staged. Confirm `roles.yml` is untouched. Confirm no fetch/hydrate helper
   was added.
6. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
7. Record evidence in `t00-evidence.md`. This package's caller PR `Closes`
   only its own VOC-138 task issue.
8. After the exact reviewed caller PR merges, `reconcile-release` for #1089
   completes promotion of #1090 (or the live promotion at the then-current
   `develop` head). Do not add a snapshot-gap task.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the repair is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
git diff --check
```

Also record the exact targeted infra and caller commands that prove:

- promotion missing-subject → `application-check provenance mode: pr-validation`;
- ordinary missing-subject fixture change → `pr-ancestry` and fail-closed;
- tampered merge-base / current hashes / missing SHAs fail closed;
- recovery refuses to rerun the doomed `pull_request` job when exact-head
  dispatch success exists.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA, and the infra PR SHA) should
confirm:

- promotion PRs with an unreachable subject select `pr-validation` and keep
  exact PR SHAs;
- `VOC-112-TEST-12` and `VOC-112-TEST-13` pass under that mode without
  fetching the subject;
- ordinary fixture-changing PRs remain `pr-ancestry` fail-closed;
- hash/SHA negatives still fail closed;
- no fetch/hydrate helper exists;
- recovery selects/publishes unambiguous exact-head success instead of
  rerunning jobs `98691441027` / `98692552949` as a strategy;
- `PINNED_SHA.txt` equals the new infra merge, not a stale #167 if that
  merge still fails #1090's class;
- all eight VOC-112 no-change paths are absent from the diff against
  `b9e74fc2…`;
- current-state docs no longer claim promotion PRs use `squash-safe-push`;
- `roles.yml` is unchanged and no OpenAI route was added;
- `t00-evidence.md` names the implementation PR base and new infra merge,
  states that the live head is bound by the App-authored independent-review
  comment/check, and does not require a commit to contain its own SHA;
- the independent-review comment binds this exact live PR head; merge-gate
  would reject a mismatch;
- no snapshot-gap task and no bootstrap exception were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this repair is live, a same-repository
  `main` <- `develop` promotion whose VOC-112 subject is not reachable can
  pass required `ci / ci` under `pr-validation`. Ordinary fixture-changing
  PRs still fail closed without the capture commit. Recovery no longer
  reruns a structurally doomed PR job when exact-head success already
  exists. Ordinary release can then merge #1090 and deploy where applicable.
- **Rollback trigger:** promotion missing-subject still selects
  `pr-ancestry`; ordinary missing-subject no longer fails closed; promotion
  PR switched to `--squash-safe-push`; fetch/hydrate helper added; eight
  VOC-112 paths rewritten; doomed PR job still the recovery strategy;
  evidence mutated at test time; self-referential exact-head SHA required;
  snapshot-gap commit; `roles.yml` / OpenAI route changed.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` merge and revert the coordinated infrastructure
  PR. That last-known-good still has the #1090 deadlock; rollback restores a
  known reviewed state, not a passing promotion PR. Do not roll back by
  hydrating `f9d11e23…`.
- **Last-known-good reference:** caller `develop` before this package's merge
  (issue-creation `87f0efcb94a213a0ede9fdbca94a707a22d42b86`), pin
  `b263c0c110591cc798b89277dfc35542abb1597b`, VOC-112 `subject_revision`
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`, unmodified eight no-change
  paths relative to `b9e74fc2db4691c48c637639b265d527de9f4505`, and `main`
  `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`.
