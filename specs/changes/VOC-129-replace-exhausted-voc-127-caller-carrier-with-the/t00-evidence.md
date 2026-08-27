# VOC-129-T00 — Evidence

Task: `VOC-129-T00` — Recreate the VOC-127 caller contract from current
develop and pin exact infra #164.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

This file is a planning-time evidence contract. Implementation results are
recorded here only after the adopted task runs. Nothing below adopts or
authorizes this package.

## Discovery recorded at planning time (issue #1042)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1042 |
| Exhausted VOC-127 task | #1039 (`VOC-127-T00`); origin #1035 |
| Unpublishable caller PR | #1041 at `9d459813a733f9c6d58ad3352df0db27d33ee7f4` |
| Attempt-2/2 pipeline run | `33058603158` |
| Attempt-2 artifact | `implement-VOC-127-VOC-127-T00-attempt2` (id `9640920826`) |
| Artifact bundle SHA-256 | `724e5547e29283b2701b70b72e000253a5100edf77545807610fce17151f7906` |
| Artifact exact head | `bbdb93aadec461830435771490f5c79ba524fed9` |
| Stale pin on that artifact | `a9df74a63976d5239b84151fd01310835c999e7c` (infrastructure #163) |
| Publisher | clean publisher refused workflow-file publication |
| Authoritative infra merge | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (KARSIFT/karsift-ai-infra#164) |
| Current `develop` pin at drafting | `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| Why local suites missed the stale pin | they did not require equality with the newly merged authoritative infrastructure revision |
| Why VOC-127 cannot retry | final allowed implementation retry already consumed; attempt `3` is forbidden |
| Why #1041 cannot be the carrier | unpublished; pin/evidence still name #163 |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-129 carrier |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-129 branch/PR from current `develop`; not #1041 |
| Infra | consume already-merged #164; do not open a replacement infra PR |
| Pin target | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Pin must not equal | `a9df74a63976d5239b84151fd01310835c999e7c` or `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| Exceptional action | live #164 identity `reconcile-production-change`; do not revive `reconcile-main-to-develop` |
| Operator identity | adopted authority issue number; no free-form SHA inputs on caller `workflow_dispatch` |
| `existing_pr_number` | remains implement-only |
| Staging | tree-equivalent sync must not deploy; retain VOC-111 path selection |
| #1041 | supersede after the governed replacement is promoted; do not merge |
| VOC-127 #1039 / #1035 | close only after promotion, with audit comments naming the VOC-129 merge; no VOC-127 completion marker bound to #1041 |
| Attempt | VOC-129-T00 attempt `1` on this carrier; do not dispatch VOC-127 as attempt `3` |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces (to be filled at implementation)

**Infrastructure (`KARSIFT/karsift-ai-infra`):** already merged as
`863fc1f35b1d35e4981a59166b0e939be1a2b681` (PR #164). This task does not
change infrastructure.

**Caller (`vocanova-platform-sandbox`):** pending T00.

## Shared-infra carrier and fixture pin (to be filled at implementation)

| Item | Value |
|------|-------|
| Coordinated infra PR | KARSIFT/karsift-ai-infra#164 (already merged) |
| Exact infra merge SHA | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Pin applicable? | expected yes — `release.yml`, helpers, tests, template, and current-state comments |
| Pin must equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Pin must not equal | `a9df74a63976d5239b84151fd01310835c999e7c` or `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| `PINNED_SHA.txt` | pending T00 |
| Bootstrap used? | **no** |
| #1041 published? | **no** |
| VOC-127 attempt 3 dispatched? | **no** |

## Validation commands

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node scripts/foundation/voc111-deploy-staging-paths.test.mjs
node scripts/foundation/voc097-fixture-matrix.test.mjs
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs
git diff --check
```

Record exact additional targeted commands and pass/fail results here after
implementation. Do not treat a missing suite as a pass.

## Post-promotion handoff (to be filled after T00 is merged and promoted)

| Item | Value |
|------|-------|
| VOC-129 caller merge SHA | pending |
| Promotion merge SHA | pending |
| `develop` tip after sync | pending |
| `main` tip after sync | pending |
| Tree-equivalent staging skipped? | pending |
| #1041 closed as superseded? | pending |
| #1039 closed with audit comment naming VOC-129 merge? | pending |
| #1035 closed with audit comment naming VOC-129 merge? | pending |
| VOC-127 completion marker bound to #1041? | **must remain no** |

## Acceptance mapping

- `VOC-129-AC-00` / `VOC-129-EV-00` — pending implementation
- `VOC-129-AC-01` / `VOC-129-EV-00` — pending implementation
- `VOC-129-AC-02` / `VOC-129-EV-00` — pending implementation
- `VOC-129-AC-03` / `VOC-129-EV-00` — pending implementation
- `VOC-129-AC-04` / `VOC-129-EV-00` — pending implementation
- `VOC-129-AC-05` / `VOC-129-EV-00` — pending implementation
- `VOC-129-AC-06` / `VOC-129-EV-00` — pending implementation
