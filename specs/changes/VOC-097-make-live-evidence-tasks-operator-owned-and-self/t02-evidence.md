---
evidence_id: VOC-097-EV-02
task_id: VOC-097-T02
acceptance_criteria:
  - VOC-097-AC-03
  - VOC-097-AC-04
  - VOC-097-AC-05
  - VOC-097-AC-06
  - VOC-097-AC-07
tests:
  - VOC-097-TEST-06
  - VOC-097-TEST-07
  - VOC-097-TEST-08
  - VOC-097-TEST-09
  - VOC-097-TEST-10
  - VOC-097-TEST-11
  - VOC-097-TEST-12
  - VOC-097-TEST-13
  - VOC-097-TEST-14
date: 2026-08-21
related_change: VOC-097
gate_status: caller-exact-sha-review-pending
shared_infrastructure_merged: true
caller_exact_sha_reviewed: false
live_fixture_claimed: false
post_merge_source_run_claimed: false
---

# VOC-097-T02 — Allowlisted observe/dispatch reconciler

## Current gate

The shared implementation is complete and merged. Shared infrastructure PR
`KARSIFT/karsift-ai-infra#49` was independently reviewed at exact head
`e173a2e1a8517cfd367f5ef9e1da75f38093abbe` and returned PASS with no
Critical/High findings. Hosted self-CI run `32437257287` passed all four checks,
including 67 policy tests. The squash merge commit
`1225131895157faf4ae83bc959b44fb888b101d0` then passed post-merge `main`
self-CI run `32437762124`. Shared issue #48 was closed only after that evidence.

This caller PR still requires its own exact-SHA CI and independent review, so
the T02 gate remains pending and no caller completion is claimed yet.

This record replaces an earlier incorrect claim that an untracked nested
`karsift-ai-infra/` checkout was bundled in the VocaNova PR. It was not tracked
by this repository and could not provide hosted evidence. The shared change is
now delivered through its own governed issue and PR.

## Shared mechanism

The shared infrastructure PR adds:

| Component | Responsibility |
| --- | --- |
| `config/live_evidence_reconcile.py` | Strict contract parsing, qualification, sanitization, timeout policy |
| `config/live-evidence-reconcile-runner.py` | GitHub metadata adapter and idempotent result commit |
| `.github/workflows/live-evidence-reconcile.yml` | Serialized, App-authenticated reusable operator job |
| `tests/test_live_evidence_reconcile.py` | Deterministic positive and negative policy coverage |

The implementation:

1. Finds only open same-repository `agent/` PRs with a trusted exact-head
   independent-review `WAITING` marker. The matching successful check must
   resolve to the active caller pipeline workflow, exact PR/head/branch, and a
   caller pipeline file that is byte-identical on the PR head and base.
2. Loads the declared contract from that exact PR head.
3. Reads workflow-run and job metadata only. It never requests logs, artifacts,
   step output, sessions, OAuth material, or user identifiers. Repository files
   accept GitHub's CR/LF-wrapped Base64 representation while retaining strict
   rejection of other invalid encoded data.
4. Fail-closes on malformed or ambiguous contracts, workflow identity, jobs,
   event, target PR association, branch, SHA lineage, conclusion, age, or result
   state.
5. Writes one allowlisted sanitized result file to the same PR branch. The new
   head triggers the ordinary PR pipeline, which must perform fresh exact-SHA CI
   and independent review.
6. Requires the result commit to have a GitHub-App-authored attestation bound to
   the exact new head before independent review may accept it.
7. Escalates a still-waiting task once after 72 hours and deduplicates dispatch,
   timeout, and qualification actions.

Optional dispatch is restricted to the workflow and inputs declared by the
contract. It rejects `pipeline.yml`, requires a protected target branch, and
requires the target workflow file to be byte-identical to the default-branch
copy. A trusted App-authored reservation precedes the single API attempt. If the
outcome becomes uncertain, the reservation blocks automatic retry instead of
risking a duplicate; only a successful attempt updates that reservation to the
sanitized requested state. The exact PR head/ref and immutable target/default
branch snapshots are checked before reservation and again immediately before
dispatch.

The implementer has no Actions credential. Read access uses the caller's normal
token; mutations use a separately minted, repository-scoped GitHub App token.
The token action is pinned to a post-fix immutable revision that honors the
three explicit `permission-*` inputs rather than inheriting the App's complete
installation permissions.

## Calling-repository integration

VocaNova's `/.github/workflows/pipeline.yml` uses:

- `ready_for_review` re-entry so an evidence-complete draft is re-evaluated on
  its unchanged exact SHA before merge;
- hourly metadata reconciliation for matching evidence and bounded timeout;
- explicit manual `reconcile`, `observe`, and `dispatch` modes;
- an explicit run ID only for `observe`;
- an explicit waiting PR number only for `dispatch`;
- `actions: read` in the caller's permission floor.

There is intentionally no catch-all `workflow_run` trigger. Since the pipeline
also produces workflow runs, that design could recursively trigger itself and
waste runner time. Hourly reconciliation finds qualifying runs by declared
workflow identity, while manual observe provides immediate controlled recovery.

Deterministic caller locks live in
`scripts/foundation/voc097-reconcile.test.mjs`. Shared behavior is tested and
reviewed in the shared repository rather than through an untracked local folder.

## Validation recorded so far

- Shared policy suite: 67 tests passed locally.
- Shared hosted self-CI: run `32437257287`, all four checks passed on the exact
  reviewed head above; post-merge run `32437762124` passed on shared `main`.
- Independent shared review: exact-SHA GPT-5.6 PASS with no Critical/High
  findings after base/head, publisher, merge, remediation, adoption, retry, and
  live-evidence consumer hardening.
- VocaNova caller focused tests passed. Governance validation passed and path
  classification reported R4 because the governance regression assertion was
  updated. `pnpm validate` passed workspace, formatting, lint, type checks, 216
  foundation tests, package/web tests, and package builds up to the API suite;
  its two
  database-backed OAuth tests could not start because Docker is unavailable in
  this WSL environment. A separate `pnpm build` passed for packages, web, and
  API. Hosted exact-SHA CI remains required and must supply the missing
  database-backed evidence after the shared PR is merged.

No secret, log content, OAuth value, personal identifier, token, or credential
is recorded. T03 owns the complete deterministic cross-repo fixture matrix; T05
owns controlled live proof (`live_fixture_claimed: false`).
