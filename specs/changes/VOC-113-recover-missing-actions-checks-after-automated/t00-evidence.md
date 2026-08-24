# VOC-113-T00 — Evidence

Task: `VOC-113-T00` — Diagnose missing-run behavior; implement durable recovery,
deterministic tests, and docs.

Evidence date: 2026-08-24

## Issue #948 observations (metadata-only)

| Observation                                 | Result                                                                                                                |
| ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| Final VOC-112 task PR merged to `develop`   | No `push` workflows observed for the integration squash commit                                                        |
| Release automation opened promotion PR #947 | No pull-request workflows / required checks on the exact head                                                         |
| Close/reopen and draft/ready transitions    | Did not recreate missing required checks                                                                              |
| `reconcile-release`                         | Reused release audit and PR #947; merge remained fail-closed without exact-head checks                                |
| User-authenticated merge of roster PR #953  | Branch advanced, but no `push` workflows or promotion-PR synchronize checks were observed for the new integration SHA |
| Active ruleset contexts                     | `governance-policy`, `validate`, `ci / ci` remain required and were not weakened                                      |

## Verified trigger/token/event behavior (`VOC-113-DEP-04`)

Repository metadata proves that downstream event fan-out is unreliable for the
API-driven mutations used in this lifecycle. It was absent after the GitHub App
task/plan mutations and also after the operator merged roster PR #953 with the
authenticated GitHub CLI OAuth identity. This rules out an App-only explanation.
Issue events and explicit `workflow_dispatch` runs continued to activate, while
the affected branch tips and PR heads still moved successfully.

Promotion PR #947 (`develop`→`main`) is a branch-to-branch pull request whose
head ref is the protected integration branch itself. GitHub did not attach the
normal pull-request validation fan-out to that exact head SHA after App-driven
PR creation. Manual close/reopen and draft/ready transitions did not recreate
the missing ruleset contexts.

GitHub metadata does not expose a more specific server-side suppression cause;
the verified contract is therefore the event/token behavior above, not a guess
about an undocumented internal cause. The durable fix is explicit
`workflow_dispatch` orchestration bound to the exact
SHA (VOC-113-D03): recovery dispatches genuine allowlisted workflows and polls
for real check-run metadata. It never fabricates statuses or weakens the
ruleset.

App-token mutation posture is preserved: merge-gate and release converge continue
to refuse `github.token` merges when App credentials are configured; recovery
dispatches use the App installation token with `actions: write` only for
declared workflow dispatches.

## Recovery mechanism (T00)

| Path                             | Behavior                                                                                                                                                                                                                |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Task merge → integration SHA     | `merge-gate.yml` runs `actions-check-recovery-runner.py` after App merge for agent task PRs; dispatches `repository-governance.yml` and the full `deploy-staging.yml` workflow when push workflows are absent           |
| Promotion PR / reconcile-release | `release.yml` runs the same runner in `promotion_pr` mode before authoritative gate selection; dispatches `governance-policy.yml`, `repository-governance.yml`, and `pipeline.yml` action `recover-promotion-pr-checks` |
| Bounded wait                     | **1800 seconds** (30 minutes), poll interval 30 seconds; timeout fails closed with sanitized diagnostics naming mode, SHA, missing contexts, and pending/failed counts                                                  |
| Read-only verifiers              | `verify-promotion-check-recovery / verify` (T01) and `verify-post-promotion-workflow / verify` (T02) validate Actions/PR metadata only                                                                                  |

Caller consumes reusable workflows at `@main`. Shared behavior merged through
`KARSIFT/karsift-ai-infra` PR #121 after all four exact-head self-CI checks
passed; the shared `main` merge commit is
`d9c4b3ed2fa6aa5b3cbd5db18c4858ec83409be1` (fixture copy under
`tooling/governance/fixtures/karsift-ai-infra` for deterministic tests).

Pre-publication review caught and corrected two live-contract defects: recovery
now reuses the existing caller `ci` job ID so the genuine required context stays
`ci / ci`, and the read-only promotion verifiers consume the REST PR fields
returned by their actual GitHub API calls. REST-shaped positive and negative
fixtures cover both corrections. The shared repository passed all 216 policy
tests before PR #121 was opened.

Independent review of caller head `bed8670c438920da83dadd95b6bfaeb9b5f2e7b0`
found blocking verifier/carrier binding, contract-input, provenance-mode, test
coverage, recovery-concurrency, and completion-publication ordering defects.
All findings were remediated in shared follow-up PR #122, whose four exact-head
self-CI checks passed before merge; the corrected shared `main` commit is
`b8d244a6d0a87ddf775675516b7bc8444d73300e`. Recovery now also performs the
full staging deploy when the normal integration push deploy is absent, rather
than treating a validation-only run as deployment equivalence.

Semantic Actions lint then caught GitHub's 25-input `workflow_dispatch` limit
before a corrected caller push: the new promotion input had raised the caller
to 26. Shared PR #123 removed the redundant caller-supplied proof-head input and
derives that trusted value from `github.sha`; all four self-CI checks passed and
the final shared `main` commit is
`6a24385c8b16ee77318572022b13f91669705cef`. Deterministic policy now enforces
the 25-input ceiling.

Independent review of caller head `37f6386f9445c43224bac332a6b12255cde3e3bc`
then found four fail-open or evidence-integrity defects and one dead helper:
original-capture provenance could fall back to merge-base validation for a
fetchable non-ancestor, queued/in-progress workflows could satisfy integration
recovery, skipped required checks could count as success, and historical
capture subjects had been rewritten without a recapture. Shared PR #124
corrected workflow/check conclusion handling and removed the dead helper; all
four exact-head self-CI checks passed before its merge. The final shared `main`
commit is `3c156312d50856aa937b161b45d0132308b59040`. Caller-side tests now reject a
fetchable non-ancestor in original-capture mode, and the immutable historical
capture subjects are restored rather than falsely reissued.

The next exact-SHA review passed with only two Low findings. Shared follow-up
PR #125 synchronized the caller fixture to the current checkout pin and added
explicit second-wake coverage: an invocation that observes active exact-SHA
recovery workflows waits without dispatching duplicates, while completion
still requires successful terminal evidence. All four shared self-CI checks
passed before merge; the final shared `main` commit is
`d82a905fbecb497ce8346e5cf13e1001a0f13f85`.

A subsequent exact-SHA review found one Medium fixture-coherence defect and two
Low completeness gaps. The caller fixture now includes the shared completion
publisher/parser contract and the current `ci.yml` checkout pin, so the local
mirror is executable rather than only structurally representative. Shared PR
#126 also serializes recovery by mode and exact target SHA and adds an hourly
caller-template wake that resolves the current integration head. If the
original post-merge attempt is interrupted or times out, the next scheduled
wake retries missing/failed genuine workflows; successful terminal evidence is
a mutation-free no-op. All four shared self-CI checks passed before merge; the
final shared `main` commit is
`97b380b77d08adbf14952df17ee8d11e6f521b15`.

## VOC-112 provenance repair (`VOC-113-D11`)

Repository Governance supplies `PR_BASE_SHA` and `PR_HEAD_SHA` and selects
`pr-ancestry` when a pull request changes a capture fixture. That mode now
requires the recorded subject to be a true ancestor of the reviewed head and
has no merge-base fallback. Later pull requests that leave the historical
fixtures unchanged use `pr-validation`: their expected immutable source hashes
must be anchored in the merge base and unchanged at the reviewed head.

## Validation results

Commands run on the T00 task branch working tree:

| Command                                                                                                               | Result           |
| --------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `python3 -m unittest tests.test_voc113_actions_check_recovery -v` (fixture infra root)                                | PASS             |
| `node --test scripts/foundation/voc113-actions-check-recovery.test.mjs`                                               | PASS             |
| `VOC112_CAPTURE_PROVENANCE_MODE=squash-safe-push node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` | PASS             |
| `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` with PR base/head SHAs (post-squash fixture mode)                      | PASS             |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`                                             | PASS — 146 tests |
| `bash scripts/governance/validate-governance.sh --base <integration-sha> --head HEAD`                                 | PASS             |
| `bash scripts/governance/classify-change-risk.sh --base <integration-sha> --head HEAD`                                | PASS — R4 floor  |
| `git diff --check <integration-sha>..HEAD`                                                                            | PASS             |

## Out of scope for T00

- Completing promotion PR #947 (T01 operator-owned live evidence)
- Post-promotion `main` workflow verification (T02)
- Weakening ruleset checks or synthesizing statuses
