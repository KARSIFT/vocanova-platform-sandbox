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
| Active ruleset contexts | `governance-policy`, `validate`, `ci / ci` remain required and were not weakened |

## Verified trigger/token/event behavior (`VOC-113-DEP-04`)

GitHub does not treat App installation-token merges and pull-request mutations as
ordinary user pushes or pull-request fan-out in every case. When the mutation
identity is the GitHub App rather than a repository `push` from a default
`GITHUB_TOKEN` merge, downstream `push` and `pull_request` workflow activation
can be absent even though branch tips and PR heads move successfully.

Promotion PR #947 (`develop`→`main`) is a branch-to-branch pull request whose
head ref is the protected integration branch itself. GitHub did not attach the
normal pull-request validation fan-out to that exact head SHA after App-driven
PR creation. Manual close/reopen and draft/ready transitions did not recreate
the missing ruleset contexts.

The durable fix is explicit `workflow_dispatch` orchestration bound to the exact
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

Caller consumes reusable workflows at `@main`. Shared behavior lands in
`KARSIFT/karsift-ai-infra` (fixture copy under
`tooling/governance/fixtures/karsift-ai-infra` for deterministic tests).

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
| `bash scripts/governance/validate-governance.sh` | PASS (after run) |
| `bash scripts/governance/classify-change-risk.sh` | PASS (after run) |
| `git diff --check` | PASS (after run) |

## Out of scope for T00

- Completing promotion PR #947 (T01 operator-owned live evidence)
- Post-promotion `main` workflow verification (T02)
- Weakening ruleset checks or synthesizing statuses
