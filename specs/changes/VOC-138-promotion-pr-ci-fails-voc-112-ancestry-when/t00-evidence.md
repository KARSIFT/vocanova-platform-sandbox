# VOC-138-T00 — Evidence

Task: `VOC-138-T00` — Unblock promotion PR CI and exact-head recovery when
the VOC-112 subject is unreachable.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

Committed evidence in this file records the implementation PR base, the new
infrastructure merge, the provenance mode-selection change, the recovery
change, and validation. It does **not** require this file, in the same Git
commit, to contain that commit's own SHA. The live implementation head is
bound by the App-authored independent-review comment/check on the
implementation PR. Merge-gate must reject any mismatch. Post-merge /
root-issue audit may record the reviewed head and merge SHA.

Independent review must explicitly evaluate the promotion missing-subject
case, ordinary `pr-ancestry` retention, hash/SHA negatives, no-fetch
constraint, and recovery "do not rerun doomed job" behavior before merge.

## Discovery recorded at planning time (issue #1091)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1091 |
| Promotion PR | #1090 |
| Release issue | #1089 |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `87f0efcb94a213a0ede9fdbca94a707a22d42b86` |
| Reusable CI checkout | `fetch-depth: 0` of `pull/1090/merge` |
| Application-check invocation | `run-app-checks.sh --pr-base-sha <base> --pr-head-sha <head>` |
| Defect | fixture diff selects `pr-ancestry`; historical subject is not in the synthetic promotion checkout |
| Fail-closed message | `PR ancestry mode requires every captured commit object` |
| Failing tests | `VOC-112-TEST-12`, `VOC-112-TEST-13` |
| Captured subject | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| PR run | `33122154521` |
| Failed jobs | `98691441027`, rerun `98692552949` |
| Dispatch recovery (pass) | `33122158425` (339 foundation tests; `--squash-safe-push`) |
| reconcile-release | `33122099253` (`selected_required_run_mismatch`), `33122436137` (reran doomed PR job) |
| Production promotion | none; `main` remains `0d0b0cdf…` |
| Issue-creation pin (#167) | `b263c0c110591cc798b89277dfc35542abb1597b` |
| Protected comparison anchor (VOC-112 eight-path) | `b9e74fc2db4691c48c637639b265d527de9f4505` |
| Doc contradiction | `docs/operations/11-devops-and-ci-cd.md` claims promotion PRs use `squash-safe-push`; live `ci.yml` does not |
| Why bootstrap is not required | T00's first run is attempt `1` on a new VOC-138 carrier from current `develop` |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-138 branch/PR from current `develop`; not a rewrite of PR #1090 |
| Infra | open a new independently reviewed `KARSIFT/karsift-ai-infra` PR; pin its exact merge; do not freeze defective #167 |
| Promotion identification | explicit signal from reusable `ci.yml` only for same-repository `main` <- `develop` |
| Promotion provenance | unreachable subject → `pr-validation` with exact PR SHAs; do not use `--squash-safe-push` for the required PR check |
| Ordinary PRs | fixture add/modify/delete stays `pr-ancestry`; missing subject still fail-closed |
| Negatives | tampered merge-base/current hashes and missing/malformed PR SHAs still fail closed |
| Hydration | **forbidden** — no `git fetch`, hydrate helper, provenance-mode wrapper, or skip |
| Recovery | do not rerun a structurally doomed `pull_request` `ci / ci` job when successful exact-head validation already exists; select or publish that success |
| Fixture freeze versus advance | issue-creation pin `b263c0c…` is historical; live pin becomes the new infra merge |
| Eight no-change paths | byte-identical to `b9e74fc2…`; do not edit `AGENTS.md`, navigator skill, VOC-112 fixtures, provenance test, runner, `validate-workspace.mjs`, or `package.json` |
| Docs | update fixture README, `docs/operations/11-devops-and-ci-cd.md`, and `docs/development/agent-skills.md` in the same caller PR |
| Exact-head binding | App-authored independent-review comment/check binds the live PR head; merge-gate rejects mismatch; a commit must not be required to contain its own SHA |
| Attempt | VOC-138-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged: implementer/escalation `cursor/composer-2.5`; planner/reviewer/reviewer_fast_retry/plan_reviewer `cursor/grok-4.6[effort=high,fast=false]` |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not copy full CI logs |
| Evidence truth | name the implementation PR base and new infra merge; bind the live head via App-authored review/check; do not mutate this file at test/runtime |

## Changed surfaces (implementation)

Implementation PR base recorded before the first in-scope edit:

pending — resolve current `develop` to a 40-character SHA before any
in-scope edit. Issue-creation develop was
`87f0efcb94a213a0ede9fdbca94a707a22d42b86`. Plan/adoption/roster commits
after that SHA are governance-only and do not count as protected-file drift.

Protected comparison anchor for eight-path VOC-112 boundary:
`b9e74fc2db4691c48c637639b265d527de9f4505`.

Issue-creation infra pin (defective for this failure class; replace):
`b263c0c110591cc798b89277dfc35542abb1597b`.

New independently reviewed infra merge (record after it exists):
pending.

In-scope implementation diff paths (expected; record actuals after commit):

- `KARSIFT/karsift-ai-infra` `config/run-app-checks.sh`
- `KARSIFT/karsift-ai-infra` `.github/workflows/ci.yml`
- `KARSIFT/karsift-ai-infra` recovery runner/modules and tests
- caller `tooling/governance/fixtures/karsift-ai-infra/**` pin and mirrors
- caller `tooling/governance/tests/` as needed for pin/scan assertions
- `docs/operations/11-devops-and-ci-cd.md`
- `docs/development/agent-skills.md`
- `specs/changes/VOC-138-promotion-pr-ci-fails-voc-112-ancestry-when/t00-evidence.md`

## Mode-selection and recovery results

| Case | Expected result |
|------|-----------------|
| Promotion PR, fixture changed, subject absent, exact SHAs | `pr-validation`; `VOC-112-TEST-12` / `VOC-112-TEST-13` pass |
| Ordinary PR, fixture changed, subject present | `pr-ancestry` |
| Ordinary PR, fixture changed, subject absent | `pr-ancestry` fail-closed |
| Unchanged fixture, valid SHAs | `pr-validation` |
| Tampered merge-base or current hashes | fail closed |
| Missing or malformed PR SHAs | fail closed |
| `git fetch` / hydrate helper | absent |
| Failed PR job plus successful exact-head dispatch | recovery does not rerun the doomed job; selects/publishes the success |
| Other selected-run identity mismatch | `selected_required_run_mismatch` |

## Validation commands (implementation)

Recorded after the repair is tracked and committed. Base is the
implementation PR base; head is the implementation PR head bound by
App-authored independent review (not recorded here as a self-SHA).

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
bind the live PR head exactly and must explicitly evaluate the promotion
missing-subject case, ordinary `pr-ancestry` retention, hash/SHA negatives,
no-fetch constraint, and recovery behavior. Merge-gate must reject any
mismatch. Record a pointer to that comment/check here after it exists; do
not write the live head SHA into this file as a value that the same commit
is required to contain. The implementer must not approve or merge its own
work.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- Genuine staging runs only for the real caller tree change. Tree-equivalent
  post-promotion develop synchronization must not keep staging scheduled.
- `reconcile-release` for #1089 may merge #1090 (or the live same-repository
  promotion at the then-current `develop` head) once required `ci / ci`
  passes under the repaired contract.
- `develop` is advanced to the successful promotion merge SHA before audit
  close and ends 0 ahead / 0 behind `main` with an identical tree.
- Production deployment proceeds where existing path selection requires it
  and is verified if selected.
- Release/task/requirement records close with audit comments naming the
  exact promotion merge. Post-merge audit may record the independently
  reviewed head and the promotion merge SHA.
- Root issue #1091 closes only after that audit evidence exists.
- Closed state alone is not completion proof.
- Incident runs `33122154521`, `33122158425`, `33122099253`,
  `33122436137` and jobs `98691441027`, `98692552949` remain audit context
  only.
