# VOC-113-T00 — Evidence

Task: `VOC-113-T00` — Diagnose missing-run behavior; implement durable recovery,
deterministic tests, and docs.

Evidence date: 2026-08-24

## Issue #948 observations (metadata-only)

| Observation | Result |
| --- | --- |
| Final VOC-112 task PR merged to `develop` | No `push` workflows observed for the integration squash commit |
| Release automation opened promotion PR #947 | No pull-request workflows / required checks on the exact head |
| Close/reopen and draft/ready transitions | Did not recreate missing required checks |
| `reconcile-release` | Reused release audit and PR #947; merge remained fail-closed without exact-head checks |
| User-authenticated merge of roster PR #953 | Branch advanced, but no `push` workflows or promotion-PR synchronize checks were observed for the new integration SHA |
| Active ruleset contexts | `governance-policy`, `validate`, `ci / ci` remain required and were not weakened |

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

| Path | Behavior |
| --- | --- |
| Task merge → integration SHA | `merge-gate.yml` runs `actions-check-recovery-runner.py` after App merge for agent task PRs; dispatches `repository-governance.yml` and `deploy-staging.yml` (`skip_ssh_deploy=true`) when push workflows are absent |
| Promotion PR / reconcile-release | `release.yml` runs the same runner in `promotion_pr` mode before authoritative gate selection; dispatches `governance-policy.yml`, `repository-governance.yml`, and `pipeline.yml` action `recover-promotion-pr-checks` |
| Bounded wait | **1800 seconds** (30 minutes), poll interval 30 seconds; timeout fails closed with sanitized diagnostics naming mode, SHA, missing contexts, and pending/failed counts |
| Read-only verifiers | `verify-promotion-check-recovery / verify` (T01) and `verify-post-promotion-workflow / verify` (T02) validate Actions/PR metadata only |

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

## VOC-112 provenance repair (`VOC-113-D11`)

Repository Governance now runs capture provenance in `pr-validation` mode for
pull requests, supplying `PR_BASE_SHA` and `PR_HEAD_SHA`. Original capture PRs
still prove subject ancestry when present; later post-squash PRs pass when
expected immutable source hashes are anchored in the merge base and unchanged at
the reviewed head.

## Validation results

Commands run on the T00 task branch working tree:

| Command | Result |
| --- | --- |
| `python3 -m unittest tests.test_voc113_actions_check_recovery -v` (fixture infra root) | PASS |
| `node --test scripts/foundation/voc113-actions-check-recovery.test.mjs` | PASS |
| `VOC112_CAPTURE_PROVENANCE_MODE=squash-safe-push node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` | PASS |
| `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` with PR base/head SHAs (post-squash fixture mode) | PASS |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | PASS — 146 tests |
| `bash scripts/governance/validate-governance.sh --base <integration-sha> --head HEAD` | PASS |
| `bash scripts/governance/classify-change-risk.sh --base <integration-sha> --head HEAD` | PASS — R4 floor |
| `git diff --check <integration-sha>..HEAD` | PASS |

## Out of scope for T00

- Completing promotion PR #947 (T01 operator-owned live evidence)
- Post-promotion `main` workflow verification (T02)
- Weakening ruleset checks or synthesizing statuses
