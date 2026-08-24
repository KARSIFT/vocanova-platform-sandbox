# VOC-113 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` merge-gate / release mutation paths,
  App installation token minting, VOC-108 authoritative check selection, branch
  ruleset required contexts, caller `.github/workflows/pipeline.yml`.
- Prerequisites: VOC-108 authoritative check selection and idempotent release
  converge are on the consumed `@main` reusable workflows; App credentials
  (`KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`) remain configured.
- Cross-repo: open infra PRs against `KARSIFT/karsift-ai-infra` with
  `Relates to KARSIFT/vocanova-platform-sandbox#<task>` (non-closing). Do not
  treat an untracked local `karsift-ai-infra/` checkout as this repository's
  tracked tree.

## File reconciliation and implementation sequence

### T00 — Diagnose, implement durable recovery, deterministic tests, docs

| Target | Action | Notes |
|--------|--------|-------|
| `specs/changes/VOC-113-.../t00-evidence.md` | create/update | Diagnosis + validation results (metadata-only) |
| `KARSIFT/karsift-ai-infra/.github/workflows/merge-gate.yml` | modify as needed | Post-merge integration SHA recovery when push workflows absent |
| `KARSIFT/karsift-ai-infra/.github/workflows/release.yml` | modify as needed | Post–PR-create / reconcile recovery when PR checks absent |
| Shared recovery helper(s) under `karsift-ai-infra/config/` | create/modify | Detect missing exact-SHA required runs; dispatch/orchestrate; timeout; diagnostics |
| Infra policy tests | create/modify | Task-merge and release-PR positive/negative fixtures |
| Caller `.github/workflows/pipeline.yml` and/or template | modify as needed | Recovery wake entrypoint if required; add `verify-promotion-check-recovery` dispatch (`VOC-113-D10`) |
| `.github/workflows/repository-governance.yml` | modify | Supply exact pull-request merge-base context to strict provenance validation |
| `scripts/foundation/voc112-navigation-benchmark.test.mjs` + fixtures | modify | Original ancestry or accepted-squash merge-base anchoring; strict negative cases |
| Shared/caller verify reusable workflow (if split like VOC-104) | create | Read-only promotion exact-head check verification |
| `scripts/foundation/voc113-*.test.mjs` | create if caller contract needs it | Mirror infra invariants + verify job naming |
| karsift-ai-infra README; AGENTS.md / ops docs | modify when claims become false | Document recovery + timeout; do not claim close/reopen recovers checks |

Ordered steps:

1. Record issue #948 observations and gather live metadata needed to document
   trigger/token/event behavior (`VOC-113-DEP-04`). Do not invent a root cause.
2. Design recovery that preserves App-token mutations and refuses status
   fabrication.
3. Implement integration-SHA recovery after App-driven task merge when required
   push/validation workflows are missing.
4. Implement promotion-PR exact-head recovery after App PR create and on
   `reconcile-release` when required PR checks are missing.
5. Bound wait with fail-closed timeout and sanitized diagnostics.
6. Add deterministic positive/negative/recursion/duplicate fixtures
   (`VOC-113-D07`).
7. Add read-only `verify-promotion-check-recovery` dispatch + job display name
   contract (`VOC-113-D10`) for T01 live evidence.
8. Update docs whose current claims would become false; record that caller
   consumes `@main` (reconcile pin only if reality differs).
9. Repair the post-squash-next-PR provenance regression per `VOC-113-D11` and
   exercise original ancestry, accepted merge-base anchoring, tampered base, and
   changed-current cases.
10. Run applicable validation; record results in `t00-evidence.md`.

### T01 — Recover and complete promotion PR #947 after genuine exact-head checks

| Target | Action | Notes |
|--------|--------|-------|
| `t01-evidence.md` | update | Allowlisted PR #947 / check / merge metadata |
| `.karsift/live-evidence/VOC-113-T01.yaml` | observe | Operator-owned contract |

Ordered steps:

1. After T00 is live on the branch the caller pipeline executes from, use
   repository-controlled recovery / `reconcile-release` (not implementer
   Actions credentials) against the existing release audit and PR #947.
2. Confirm genuine required checks appear for #947's exact head and succeed.
3. Allow release converge to merge only under VOC-108 authoritative success;
   record metadata-only evidence.
4. Do not fabricate statuses; do not open a second promotion PR.

### T02 — Verify post-promotion workflows and close remediation

| Target | Action | Notes |
|--------|--------|-------|
| `t02-evidence.md` | update | Post-promotion run metadata on `main` |
| `.karsift/live-evidence/VOC-113-T02.yaml` | observe | Operator-owned contract |

Ordered steps:

1. After #947 merges, resolve the promotion result SHA on `main`.
2. Confirm expected post-promotion workflows ran for that SHA (at minimum the
   normal `main` push path such as `deploy-production` when selected by existing
   policy).
3. Record metadata-only evidence; close issue #948 / remediation only after
   verification.

## Validation and independent verification

Deterministic (T00), as applicable to the real file list:

```bash
# In KARSIFT/karsift-ai-infra (infra self-ci / policy tests added by T00)
python3 -m unittest discover -s tests -p 'test_*voc113*'   # or the actual suite names T00 adds
python3 -m unittest tests.test_release_policy tests.test_merge_gate_policy

# In this calling repository when caller paths change
node --test scripts/foundation/voc113-*.test.mjs
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Independent verifier (exact reviewed task PR SHA):

- Confirm recovery never fabricates statuses and preserves App-token mutation
  posture.
- Confirm authoritative exact-head selection (VOC-108) remains fail-closed.
- Confirm deterministic fixtures cover both recovery paths and negatives.
- For T01/T02, confirm operator-owned metadata is allowlisted and bound to the
  declared SHA lineage.

## Deployment and rollback

- **Staging effect:** None intentional beyond existing develop push selection.
- **Production effect:** Completing #947 may promote reviewed history and trigger
  the existing automatic `main` deploy path — recovery of stranded release
  handoff, not new deploy policy.
- **Rollback trigger:** Recovery fabricates evidence, opens duplicate promotion
  PRs, recurses unboundedly, or blocks legitimate merges with false negatives.
- **Rollback mechanism:** Revert the infra/caller task PR(s); `reconcile-release`
  remains available as the pre-change wake path while revert lands.
- **Last-known-good reference:** Pre-T00 `merge-gate.yml` / `release.yml` on the
  consumed karsift-ai-infra `@main` revision immediately before T00 lands.
