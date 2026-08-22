---
evidence_id: VOC-111-EV-01
task_id: VOC-111-T01
acceptance_criteria:
  - VOC-111-AC-01
tests:
  - VOC-111-TEST-10
date: 2026-08-22
related_change: VOC-111
cites: VOC-111-EV-00
gate_status: pending-fixture-merge
live_absence_claimed: false
t00_merge_sha: ca99b614aa5aaafe3322e37652bc4153acb1b945
fixture_revision: 1
---

# VOC-111-T01 — Governed docs-only selector fixture

This task PR changes **only** this evidence file under
`specs/changes/VOC-111-skip-full-staging-deploys-for-documentation-and/`. That
bounded diff is outside the push allowlist in `VOC-111-EV-00` (`specs/**` is not
selected). When merged to `develop`, it must not schedule `deploy-staging` on push.

T02 records operator-owned Actions metadata proving zero matching runs for the
resulting integration SHA. **T01 does not claim that absence** or name the
integration merge SHA.

Do not copy secrets, workflow logs, session values, OAuth data, or personal data.

## T00 dependency

| Item | Value |
|------|-------|
| Task | `VOC-111-T00` merged to `develop` |
| Merge commit | `ca99b614aa5aaafe3322e37652bc4153acb1b945` |
| Merge PR | #926 |
| Task issue | #922 |
| Evidence | `t00-evidence.md` in this package directory |

T00 added push-only path selection to `.github/workflows/deploy-staging.yml` per
`VOC-111-D03`. Deterministic selector tests in
`scripts/foundation/voc111-deploy-staging-paths.test.mjs` prove `specs/**` paths do
not select push-triggered deploy.

## Fixture metadata (pre-merge)

| Item | Value |
|------|-------|
| Fixture task issue | #923 |
| Fixture task PR | #927 |
| Initial published fixture SHA | `9b4bdade260f0d3be125aaa597ca4c539fd44527` |
| Final reviewed head SHA | bound by the exact-head PR checks; T02 records it after merge |
| Integration push SHA | pending — T02 records after this task merges |
| Fixture revision | `1` |
| Exact changed paths | `specs/changes/VOC-111-skip-full-staging-deploys-for-documentation-and/t01-evidence.md` |

## Fixture boundary

| Check | Result |
|-------|--------|
| Sole changed path | this `t01-evidence.md` only |
| Under `specs/changes/**` | yes |
| Outside push allowlist (`VOC-111-D03`) | yes — `specs/**` is not allowlisted |
| `apps/**`, `packages/**`, `infra/**` touched | no |
| Repository-root files touched | no |
| `.github/workflows/deploy-staging.yml` touched | no |
| `tests/staging-e2e/**` touched | no |
| Post-merge absence claimed in this file | no |

The initial published SHA above is the exact pre-merge carrier revision that made
the PR number available. This metadata-binding revision necessarily changes the PR
head; its final exact reviewed head is authoritative in the PR check suite and is
copied into T02 only after merge. This avoids presenting the initial SHA as the
final reviewed revision or creating a self-referential hash claim.

## Validation (implement-time)

| Command | Result |
|---------|--------|
| `bash scripts/governance/validate-governance.sh` | pass |
| `git diff --check` | pass |

Entire task diff is limited to this package path in `specs/**` and therefore
outside the deploy-staging push allowlist.

## Acceptance mapping

| ID | Result |
|----|--------|
| VOC-111-TEST-10 | pass — exact task diff is specs-only and non-circular |
| VOC-111-AC-01 | partial — live absence proof belongs to T02 |

## Boundary for T02

- T02 resolves the integration merge SHA and timestamp only after this task merges.
- T02 performs the read-only Actions metadata query and records the zero-run result.
- No result in this file authorizes a live-absence claim.
