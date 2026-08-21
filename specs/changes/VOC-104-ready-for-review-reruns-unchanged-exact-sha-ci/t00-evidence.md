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

A deeper independent exact-SHA audit of the pinned adoption then found five additional proof and
efficiency gaps: live-evidence attestations were not task-bound, the queued package-path job could
force an unnecessary full rerun, a skipped auto-merge job could satisfy the verifier, the final
head proof was incomplete, and negative/uncertain metadata states lacked a complete matrix. Shared
PR `KARSIFT/karsift-ai-infra#90` repaired those paths and then closed two further review findings:
the verifier now requires the authenticated source PR to be recorded as merged, and behavioral
tests exercise the real attestation producer plus the workflow's exact task classifier. Its exact
reviewed head was `21f09e993e5579e24dcd409239e893f41480eab1`; all 132 shared policy tests passed,
exact-head CI run `32515578972` passed all four jobs, and a fresh independent review returned PASS
with no actionable findings. The repair merged as
`d625b40f05b9b860dbf938de41f8ec837740a9fc`; post-merge main CI run `32516435803` passed.

Independent review found and the implementation corrected proof-lineage ambiguity, missing
base-binding, unrelated-check eligibility, an unreachable reuse decision, untrusted metadata
lookalikes, missing ready-event provenance, and an empty workflow-run PR-association fallback.
Those corrections are included in the merged shared commit above.

## Calling-repository adoption

- `.github/workflows/pipeline.yml` consumes both reusable workflows at `@main` and supplies distinct
  source PR/base/head plus explicit evidence-carrier head inputs to the verifier.
- The tracked fixture is pinned to shared merge
  `d625b40f05b9b860dbf938de41f8ec837740a9fc`; 16 copied workflow, helper, template, and test files
  were verified byte-identical to the independently reviewed shared head.
- Fixture tests always execute the tracked copy. They no longer prefer an incidental untracked
  `karsift-ai-infra/` checkout, which previously made results depend on the operator's filesystem.
- Synthetic repository-validator fixtures exclude generated `node_modules`; this preserves their
  tracked-source assertion surface while avoiding a full dependency-tree copy for every test case.
- DOC-15 §17.3 and the fixture README distinguish safe exact-SHA reuse from the full fallback path.

## Commands and results

| Command | Result |
| --- | --- |
| `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py' -v` | PASS — 28 tests |
| `python3 -m unittest tooling.governance.tests.test_ready_for_review_reuse -v` | PASS — 2 calling tests, including an explicit green-evidence/draft auto-merge block |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py' -v` | PASS — 137 governance tests after the final fixture-permission regression was added |
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

A final read-only exact-SHA audit caught one omitted file in the first repin: the tracked
`live-evidence-reconcile.yml` still granted Actions write to the App and only Actions read to the
workflow token, while the pinned runner dispatches with the workflow token. The shared merge already
contained the correct least-privilege split. The fixture was recopied from that merge, the
byte-identical file count was corrected from 15 to 16, and deterministic caller tests now require
Actions write on the dedicated workflow token while forbidding Actions permission on the App token.

The first PR exact-SHA run (`32510529512`) passed full CI and independent review, but Repository
Governance run `32510527837` failed because two older VOC-080 tests froze the workflow-dispatch
option list before the new verifier action existed. Both assertions were updated to the authorized
list. The independent review's Low finding about indirect draft coverage was also corrected with an
explicit test that requires `is_draft == false` even when checks and verdict are green. Exact-head
pipeline run `32513043481` then passed full hosted CI, App-published independent review, remediation,
and merge-gate reporting on head `fa0e1e21afc5a32f20851c0f78667dd1c64eb759`; the separate
Repository Governance, Governance policy, and controlled-signup OAuth E2E workflows also passed.
The PR deliberately remained draft because the subsequent deeper audit identified the shared
contract gaps repaired in PR #90. This evidence update does not reuse the older verdict for the new
shared pin: the final PR head must receive its own hosted checks and App-bound independent review.

The final current commit SHA is intentionally not self-recorded in this file: a commit cannot contain
its own hash without changing that hash. GitHub's immutable check-run records and the trusted
App-published review comment are the final exact-head bindings; this file records the preceding
code/provenance evidence and never claims that a stale run reviewed a newer commit.

## Live proof boundary

The controlled draft-to-ready optimized-path proof remains operator-owned under `VOC-104-T01` and
its `.karsift/live-evidence/VOC-104-T01.yaml` contract. T00 supplies the implementation,
deterministic coverage, and read-only verifier only; it does not claim the T01 live transition.

No production credentials, logs, artifacts, application sessions, OAuth data, or user identifiers
were used or recorded.
