# VOC-124-T00 — Evidence

Task: `VOC-124-T00` — Request workflow-write on the coordinated source
publisher token.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1013)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1013 |
| Blocked adopted task | #1003 (`VOC-122-T00`) |
| Pipeline run / job | `32958526215` / `98147443377` |
| Nested bundle head | `f90eb630743c8c523e2e6e8dff017acbb31a7f43` |
| Infrastructure base | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` |
| Bundle download / verify / token mint | succeeded |
| Error | `refusing to allow a GitHub App to create or update workflow ... without workflows permission` |
| Changed workflow file | `.github/workflows/recover-actions-checks.yml` |
| Caller PR | #1012 — intentionally draft; staged mirror/evidence; still pins the prior infrastructure SHA; must remain unmerged until the authoritative source PR merges |
| Defect | `publish-source` mint requests `permission-contents`, `permission-issues`, and `permission-pull-requests`, and omits `permission-workflows: write` |
| Installation | `karsift-ai-infra-bot` installation `148001476` already has repository `workflows: write` |
| Why a VOC-122 retry cannot self-repair | caller executes `implement.yml@main`; the same mint would reject its own workflow-file fix |
| Why bootstrap is required | the repair itself changes `.github/workflows/implement.yml` and must land through a bounded supervised infra PR, then be exhausted |

## Chosen delivery path

To be recorded during implementation. Required constraints already settled:

| Item | Constraint |
|------|------------|
| Token change | `permission-workflows: write` on the `publish-source` mint only |
| Caller publisher | keep omitting `permission-workflows`; keep rejecting `.github/workflows/**` |
| Bootstrap | `VOC-124-D04` one-time supervised infra PR from current `main`; no PATH/Git interception; no direct `main` push; no VOC-122 bundle publication |
| VOC-122 | retry existing #1003 / #1012 after the exact infra merge is live; do not create a replacement task or PR; do not hand-push `f90eb630743c8c523e2e6e8dff017acbb31a7f43` |
| A-004 text | correct `implement.yml` caller PR body / current-state comments; do not rewrite historical CHANGELOG |
| Secrets | do not print credential values; do not rotate App secrets; do not change installation permissions |

## Changed surfaces

To be recorded during implementation.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending implementation |
| Bootstrap branch | pending implementation |
| Review base SHA | pending; current fixture pin at planning time is `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` |
| Independently reviewed infra head SHA | pending implementation |
| Exact infra merge SHA | pending implementation |
| Separate merger | pending implementation |
| `VOC-124-D04` bootstrap status | **not started** — must be exhausted by the exact infra merge before normal caller work resumes |
| VOC-122 nested head `f90eb63…` published by bootstrap? | **must remain no** |
| Pin applicable? | expected yes — `implement.yml` is in the policy fixture subset; confirm during implementation |
| `PINNED_SHA.txt` after source merge | pending implementation |

## Independent source review

To be recorded during implementation. Bind the review to the exact final
infrastructure head. The implementer must not approve or merge that carrier.

## Dependent #1003 / #1012 (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| Caller PR #1012 | Existing VOC-122 draft carrier; must remain unmerged until the authoritative VOC-122 source PR merges |
| This package's duty | Repair publisher token permission; exhaust bootstrap; record the exact reviewed infra SHA that #1003 should be re-dispatched or reconciled against |
| Re-dispatch against | pending exact VOC-124 infra merge SHA |
| Reconstruct `f90eb630743c8c523e2e6e8dff017acbb31a7f43` by hand? | No |
| Treat VOC-124 merge as VOC-122 completion? | No |

## Validation commands

```bash
# Infrastructure (exact reviewed source worktree)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

Record command results, narrower targeted tests, and pin-literal updates
during implementation. Do not report an unavailable check as passing.

## Acceptance mapping

- `VOC-124-AC-00` / `VOC-124-EV-00` — pending: `publish-source` mint requests `permission-workflows: write`
- `VOC-124-AC-01` / `VOC-124-EV-00` — pending: caller `publish` still omits that permission and still refuses workflow files
- `VOC-124-AC-02` / `VOC-124-EV-00` — pending: authorized `.github/workflows/**` source bundle coverage
- `VOC-124-AC-03` / `VOC-124-EV-00` — pending: missing credentials, invalid bundles, stale bases/leases fail closed
- `VOC-124-AC-04` / `VOC-124-EV-00` — pending: VOC-121/VOC-123 isolation, named-ref bundle, lease, retry limits preserved
- `VOC-124-AC-05` / `VOC-124-EV-00` — pending: A-004 current-state text corrected; historical records unchanged
- `VOC-124-AC-06` / `VOC-124-EV-00` — pending: bootstrap exhausted; pin follows exact infra merge; existing VOC-122 carrier retried, not replaced
