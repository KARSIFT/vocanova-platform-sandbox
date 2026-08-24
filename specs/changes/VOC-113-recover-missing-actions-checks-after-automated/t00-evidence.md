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
| Task merge → integration SHA     | `merge-gate.yml` runs `actions-check-recovery-runner.py` after App merge for agent task PRs; always recovers `repository-governance.yml`, and recovers full `deploy-staging.yml` only when the merged paths match the VOC-111 deploy selector |
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

The first caller CI run after adding the scheduled wake correctly failed an
older VOC-097 least-privilege assertion that allowed exactly one Actions-write
job. The assertion now names and constrains both authorized jobs—the
live-evidence reconciler and exact-SHA integration recovery—while continuing to
prove the implementer and workflow-level permission floor have no Actions-write
authority. The full foundation suite passes with 333 tests.

The final exact-SHA review passed and reported three non-blocking hardening
opportunities; all three were nevertheless resolved. Shared PR #127 now
requires `success` (not `neutral`) for recovered integration workflows and
passes the immutable target SHA into each integration dispatch. Caller
workflows validate that SHA against the dispatched branch tip before any
governance or deployment work, so branch movement fails closed. The caller's
promotion recovery path also runs the reuse-decision prerequisite explicitly
in `recovery` mode, eliminating dependence on `always()` traversing a skipped
job. Shared PR #127 passed 226 policy tests and all four hosted self-CI checks;
the shared `main` merge commit is
`bc2ea04f4cb664c0f4fa16ceaebdf0160c5ba0a0`.
Shared PR #128 mirrored the explicit recovery prerequisite into the canonical
caller template and passed the same 226-test/four-check gate; its shared
`main` merge commit is `4edfb2d979b86df9cdacc10d5de5120e46763417`.
The subsequent exact-SHA review passed with three Low findings; all were
resolved in shared PR #129 and this caller fixture. Neutral terminal workflows
are retryable rather than merely timing out, the canonical caller template
preserves the live caller's null guard for check runs without an associated
pull request, and the remaining reuse-workflow checkout pins are synchronized.
Shared PR #129 passed 227 policy tests and all four hosted checks; its `main`
merge commit is `1f6c1477d708ea89b1d875fa7684e358538d03a5`.

The next exact-SHA review identified two Medium recovery-policy mismatches and
two Low operational gaps; all four were resolved in shared PR #130. Integration
recovery now mirrors the VOC-111 staging path selector, so documentation-only
merges recover governance without forcing a deploy, while runtime/root/deploy
changes still require the full staging workflow. Promotion recovery requires a
literal successful check-run conclusion (neutral is insufficient). The
promotion verifier sanitizes status-API failures, and the hourly wake performs
a read-only exact-head preflight before invoking the reusable recovery. Shared
PR #130 passed 231 policy tests and all four hosted checks; its `main` merge
commit is `a4899e17fc9eab6e37ea802fce129ea30634e8c3`.
Shared PR #131 synchronized the canonical template's operator comment with the
path-aware behavior; all 231 policy tests and four hosted checks passed. Its
`main` merge commit is `255678b41fb29b27c42d4632f807b42682c29430`.
The next caller review found that the live caller's honest empty recovery-only
base-SHA metadata had not been mirrored into the canonical template. Shared PR
#132 synchronized that template and added positive/negative regression checks;
all 231 policy tests and four hosted checks passed. Its `main` merge commit is
`091fabdf9cf074396339750c7a58d215fbdc9aec`.

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
