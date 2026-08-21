# VOC-098 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/operational-failure-monitoring.yml`,
  `infra/scripts/classify-deploy-concurrency-cancel.sh` (wiring only),
  VOC-088/VOC-094 foundation tests, operator docs that describe observer auth.
- Prerequisites: VOC-088-T02 observer merged; VOC-094-T00 classifier and
  `permission-actions: read` on App mint present (the bug this package removes);
  issue #840 opened with least-privilege restoration scope.

## File reconciliation and implementation sequence

### T00 — Least-privilege App token and classifier split

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/operational-failure-monitoring.yml` | modify | Drop App `permission-actions`; dual-token wiring; permissions floor |
| `infra/scripts/classify-deploy-concurrency-cancel.sh` | modify only if needed | Prefer env wiring in workflow; preserve fail-closed behavior |
| `scripts/foundation/voc088-failure-to-issue.test.mjs` | modify | Assert no App Actions permission |
| `scripts/foundation/voc094-deploy-concurrency.test.mjs` | modify | Assert classifier/App token split |
| `scripts/foundation/voc098-observer-token-permissions.test.mjs` | create (optional) | Only if cleaner than extending VOC-088/094 |
| `docs/operations/staging-controlled-signup.md` | modify if needed | Dual-token accuracy |
| `specs/changes/VOC-098-.../t00-evidence.md` | create | Commands + results |

Ordered steps:

1. Confirm drafting-time diagnosis against current HEAD of the observer workflow;
   record mint permission list before/after in evidence (no secrets).
2. Remove `permission-actions: read` from App mint; keep `permission-issues: write`.
3. Set workflow/job `permissions` to include `actions: read` (and `contents: read`);
   pass `github.token` to the classifier step only.
4. Keep `open-failure-issue.sh` on App token; keep App credential fail-closed gate.
5. Update foundation tests that currently `assert.match(..., /permission-actions: read/)`.
6. Align operator docs if they would otherwise remain false.
7. Run foundation tests and applicable governance validation.

### T01 — Controlled live proof

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-098-.../t01-evidence.md` | create | Metadata-only live proof |
| `.karsift/live-evidence/VOC-098-T01.yaml` | already drafted | Operator-owned contract; do not invent unrelated workflows |

Ordered steps:

1. Ensure T00 is live on the observer execution branch (expected `main` after
   promotion).
2. Operator follows controlled cancel fixture; record scrubbed metadata.
3. Confirm observer success + create-or-dedupe + no duplicate on repeat.
4. Do not grant implementer Actions credentials or edit unrelated pipelines.

## Validation and independent verification

Deterministic (T00):

```bash
node --test scripts/foundation/voc088-failure-to-issue.test.mjs
node --test scripts/foundation/voc094-deploy-concurrency.test.mjs
# plus any new voc098 foundation test if added
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Live (T01): operator-owned; see live-evidence contract and
`docs/operations/staging-controlled-signup.md` controlled failure fixture.
Independent verifier binds the exact task PR SHA and confirms AC/test/evidence
traceability without treating missing live evidence as a code defect to remediate
via unrelated edits (VOC-097).

## Deployment and rollback

- **Authorization:** Package adoption + task implementation authorization only;
  this package does not itself authorize production application deployment.
- **Rollout:** T00 merges to `develop`, promotes to `main` via normal release path
  so the default-branch observer picks up the fix; then T01 live proof.
- **Rollback trigger:** Observer cannot mint; issues stop appearing for real
  failures; or `GITHUB_TOKEN` is observed creating issues / plan-from-issue stops.
- **Rollback mechanism:** Revert T00 commits through normal PR path.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop`/`main` HEAD before T00 merge (known-broken mint
  against current App installation — rollback restores prior broken mint unless
  paired with out-of-scope installation permission changes).
