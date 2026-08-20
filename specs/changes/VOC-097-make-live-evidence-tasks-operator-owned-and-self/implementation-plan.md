# VOC-097 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected / sensitive areas: `KARSIFT/karsift-ai-infra` reusable workflows
  (`remediate.yml`, `review.yml`, `merge-gate.yml`, `implement.yml`, new reconcile
  workflow), App-token permissions, calling-repo `pipeline.yml` if rewired,
  AGENTS.md / governance docs that describe workflow behavior.
- Prerequisites: issue #823 requirements stable; VOC-093-T00 / VOC-094-T00 already
  merged (stranded items are their T01 evidence tasks #779 / #785).
- Cross-repo: open infra PRs against `KARSIFT/karsift-ai-infra` for T01–T03
  behaviors; do not commit the untracked local `karsift-ai-infra/` checkout into
  this repository's tree.

## File reconciliation and implementation sequence

### T00 — Docs and template declaration

| File / area | Action | Notes |
|-------------|--------|-------|
| `docs/operations/` live-evidence guide | create | Author/operator contract |
| `specs/templates/change-package/` | modify | Declaration guidance |
| `AGENTS.md` / governance docs | modify only if currently false | Keep doc sync rule |
| `t00-evidence.md` | create | Chosen contract shape |

Ordered steps:

1. Resolve DEP-03 at adoption (or record implementer confirmation of `VOC-097-D03`).
2. Write operator + template guidance for allowlisted contracts.
3. Add foundation/doc assertions as needed for TEST-00/01.

### T01 — Waiting lifecycle and remediate skip

| File / area | Action | Notes |
|-------------|--------|-------|
| `karsift-ai-infra` prompts/review.md | modify | Waiting vs code-defect FAIL |
| `karsift-ai-infra` remediate.yml | modify | No retry while waiting |
| `karsift-ai-infra` implement.yml / README | verify/document | No Actions credential grant |
| merge-gate.yml | modify only if required | Fresh review still required |
| Infra / foundation fixtures | create/extend | Waiting ≠ remediate |

Ordered steps:

1. Define machine-readable waiting marker.
2. Teach review prompt + remediate decide step to honor it.
3. Lock least-privilege invariant with tests.
4. Prove genuine FAIL still remediates.

### T02 — Reconciler

| File / area | Action | Notes |
|-------------|--------|-------|
| New infra reconcile workflow + helpers | create | Observe/dispatch allowlist |
| Sanitizer / contract parser | create | Fail closed |
| Timeout + dedup | create | Bounded escalation |
| Calling `pipeline.yml` | modify if needed | Consume new reusable job |
| `t02-evidence.md` | create | Mechanism + pin/wiring |

Ordered steps:

1. Implement contract parse + allowlisted metadata writer.
2. Implement observe path; add optional dispatch only when contract permits.
3. Wire wake → ready-for-re-review + exact-SHA re-review requirement.
4. Add timeout/dedup; keep observer/Sentry separate.

### T03 — Deterministic matrix

| File / area | Action | Notes |
|-------------|--------|-------|
| Infra tests and/or `scripts/foundation/voc097-*.test.mjs` | create/extend | Full matrix |
| `tooling/governance/fixtures/` | extend if needed | Vendored workflow fixtures |
| `t03-evidence.md` | create | Command output (scrubbed) |

### T04 — Stranded #779 / #785

| File / area | Action | Notes |
|-------------|--------|-------|
| VOC-093-T01 / VOC-094-T01 task PRs or clean replacements | migrate | Governed path |
| Evidence contracts for those tasks | create | Match D03 |
| `t04-evidence.md` | create | Closure or migration record |

### T05 — Live proof

| File / area | Action | Notes |
|-------------|--------|-------|
| Controlled waiting + qualify runs | operate | Sandbox |
| Observer health spot-check | operate | Separate from waiting |
| `t05-evidence.md` | create | Scrubbed URLs only |

## Validation and independent verification

Deterministic (as applicable per task):

```bash
# infra self-ci / node fixtures for VOC-097 (exact paths set in T01–T03)
node --test scripts/foundation/voc097-*.test.mjs
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Live (T05):

- Controlled waiting fixture without remediate retry
- One qualifying reconcile + fresh exact-SHA review
- Operational-failure / Sentry separation health check (scrubbed)

Independent verifier binds each task PR to its exact SHA under active A-004;
re-reviews after material remediation.

## Deployment and rollback

- **Authorization:** Merging task PRs updates governance automation; this package
  does not authorize application production deployment.
- **Rollout:** Infra reusable workflows take effect for callers on pin/`@main`
  consumption; calling-repo docs apply on merge to `develop` / promotion.
- **Rollback trigger:** False waiting skips real defects; false wakes merge bad
  evidence; sanitizer leaks; waiting loops despite timeout.
- **Rollback mechanism:** Revert infra workflow/prompt commits and calling-repo
  wiring; restore prior remediate behavior; leave audit evidence files.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** infra and calling-repo SHAs before T01/T02 merges.
