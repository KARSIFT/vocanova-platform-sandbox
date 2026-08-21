# VOC-104-T00 — Evidence

Task: `VOC-104-T00` — Ready-for-review reuse policy, fail-closed path, docs, deterministic tests.

Evidence date: 2026-08-21

## Outcome

The shared reusable-workflow implementation and the VocaNova caller adoption are complete. A
`ready_for_review` event may reuse an earlier exact-base/exact-head successful pipeline and trusted
App-published review only when every deterministic precondition passes. The current event still
runs merge-gate. Any ineligibility or evaluation uncertainty takes the normal full CI and model
review path; draft PRs remain non-mergeable.

The implementation keeps three identities separate:

1. the prior successful full pipeline run that produced reusable CI/review evidence;
2. the source PR transition run that proves the unchanged draft-to-ready event;
3. the later evidence-carrier head used by the read-only T01 verifier.

Human or implementer comments cannot authorize reuse. The trusted review record binds the prior
pipeline run ID, exact base/head, package/task identity, and authority issue. The merge gate
independently revalidates the prior run and the current intentionally skipped caller jobs.

## Shared infrastructure provenance

The reusable behavior was independently reviewed and merged first:

| Evidence | Result |
| --- | --- |
| Shared PR | `KARSIFT/karsift-ai-infra#88` — merged |
| Independently reviewed head | `b5a6cece5e15294ac2dbdd5d467efdd5e6760a8a` |
| Shared merge commit | `03ac50126be3ef77155d75beaf7aeb4cc3f23df9` |
| Exact-head CI | run `32508879849` — all four policy jobs passed |
| Post-merge main CI | run `32509652055` — passed |
| Shared policy suite | PASS — 126 tests |
| Independent exact-SHA review | PASS — no actionable correctness finding after remediation |

A follow-up exact-SHA review of the calling-repository adoption found two Low shared-contract
gaps. Shared PR `KARSIFT/karsift-ai-infra#89` corrected both before T00 promotion: helper checkout
is pinned to the resolved reusable-workflow revision, and trusted verdict selection now uses the
same `(created_at, id)` ordering in eligibility and merge-gate. Its independently reviewed head
was `8638d8ee3f20623af831b656acf71f7150944907`, exact-head CI run `32512506670` passed all four
jobs, independent review returned PASS with no actionable finding, and the correction merged as
`6d5347b136f1993f8a4c2f6d49787b788a431bf8`. Post-merge main CI run `32512727314`
also passed all four jobs.

Independent review found and the implementation corrected proof-lineage ambiguity, missing
base-binding, unrelated-check eligibility, an unreachable reuse decision, untrusted metadata
lookalikes, missing ready-event provenance, and an empty workflow-run PR-association fallback.
Those corrections are included in the merged shared commit above.

## Calling-repository adoption

- `.github/workflows/pipeline.yml` consumes both reusable workflows at `@main` and supplies distinct
  source PR/base/head plus explicit evidence-carrier head inputs to the verifier.
- The tracked fixture is pinned to shared merge
  `6d5347b136f1993f8a4c2f6d49787b788a431bf8`; 14 copied workflow, helper, template, and test files
  were verified byte-identical to the independently reviewed shared head.
- Fixture tests always execute the tracked copy. They no longer prefer an incidental untracked
  `karsift-ai-infra/` checkout, which previously made results depend on the operator's filesystem.
- Synthetic repository-validator fixtures exclude generated `node_modules`; this preserves their
  tracked-source assertion surface while avoiding a full dependency-tree copy for every test case.
- DOC-15 §17.3 and the fixture README distinguish safe exact-SHA reuse from the full fallback path.

## Commands and results

| Command | Result |
| --- | --- |
| `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py' -v` | PASS — 24 tests |
| `python3 -m unittest tooling.governance.tests.test_ready_for_review_reuse -v` | PASS — 2 calling tests, including an explicit green-evidence/draft auto-merge block |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py' -v` | PASS — 136 governance tests in 15.349 seconds |
| `node --test scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | PASS — 5 tests |
| `node --test scripts/foundation/voc097-fixture-matrix.test.mjs` | PASS — 5 compatibility tests |
| `pnpm validate` | Local environment reached API tests: 231/231 foundation, 28/28 client, and 16/16 web tests passed; the command then stopped because Docker Desktop WSL integration was unavailable to start disposable Postgres for two pre-existing OAuth tests |
| `cd apps/api && go test ./... -skip 'TestControlledSignupOAuth_(AllowlistedCallbackSucceeds\|UnlistedCallbackDenied)'` | PASS — complete non-container API suite |
| `pnpm run build` | PASS — packages, production web build, and API build |
| `bash scripts/governance/validate-governance.sh --files-from <changed-files>` | PASS |
| `bash scripts/governance/classify-change-risk.sh --files-from <changed-files>` | PASS — R4 floor |
| `git diff --check origin/develop` | PASS |

The exact-SHA GitHub CI run remains the required full validation authority because its runner has
Docker available and executes the two disposable-Postgres OAuth tests. This file does not claim the
local `pnpm validate` invocation passed.

The first PR exact-SHA run (`32510529512`) passed full CI and independent review, but Repository
Governance run `32510527837` failed because two older VOC-080 tests froze the workflow-dispatch
option list before the new verifier action existed. Both assertions were updated to the authorized
list. The independent review's Low finding about indirect draft coverage was also corrected with an
explicit test that requires `is_draft == false` even when checks and verdict are green. A later
exact-SHA run is required; the earlier review and CI are not reused across the corrective commit.

## Live proof boundary

The controlled draft-to-ready optimized-path proof remains operator-owned under `VOC-104-T01` and
its `.karsift/live-evidence/VOC-104-T01.yaml` contract. T00 supplies the implementation,
deterministic coverage, and read-only verifier only; it does not claim the T01 live transition.

No production credentials, logs, artifacts, application sessions, OAuth data, or user identifiers
were used or recorded.
