# VOC-129 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/pipeline.yml` exceptional production-change dispatch;
  `deploy-staging.yml` path selection if a job-level skip is required;
  current-state release and branch documentation.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `60afda3a44fd06b8c00b219771de7112f1aded6e`. Confirm fixture `release.yml`
  still restores a missing integration ref at `CHECKED_HEAD_SHA`. Confirm live
  `pipeline.yml` action options still omit `reconcile-production-change` and
  that `release` still `needs: [merge-gate]` only. Confirm PR #1041 remains
  at `9d459813a733f9c6d58ad3352df0db27d33ee7f4` and is not merged. Confirm
  KARSIFT/karsift-ai-infra#164 remains merged as
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Confirm VOC-111
  `deploy-staging.yml` path allowlist remains the staging selector. Confirm
  VOC-126 `pipeline.yml` is at most 25 `workflow_dispatch` inputs with
  `existing_pr_number` implement-only and verifiers on `pipeline-verify.yml`.
- No bootstrap exception. VOC-124 already published
  `permission-workflows: write` on `publish-source`. T00's first run is
  attempt `1` on a new VOC-129 carrier. Do not treat an untracked local
  `karsift-ai-infra/` checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not merge PR #1041. Do not dispatch VOC-127-T00 as attempt `3`. Do not
  add operator SHA inputs. Do not revive `reconcile-main-to-develop`.

## File reconciliation and implementation sequence

### T00 — Recreate the VOC-127 caller contract and pin exact infra #164

| Target | Action | Notes |
|--------|--------|-------|
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact merge `863fc1f35b1d35e4981a59166b0e939be1a2b681`; include `release.yml`, `reconcile-production-change.yml`, checkout-ref and branch-sync helpers, template `pipeline.yml`, and #164 tests |
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` | replace | Must equal `863fc1f35b1d35e4981a59166b0e939be1a2b681`; must not equal `a9df74a6…` or `60afda3a…` |
| `.github/workflows/pipeline.yml` | modify | Add `reconcile-production-change`; forward `issue_number`; `release` needs `[merge-gate, reconcile-production-change]`; `auto-advance` waits on successful production-change reconcile; ≤25 inputs; no operator SHAs |
| `.github/workflows/deploy-staging.yml` | modify only if needed | Job-level skip when `on.push.paths` cannot skip tree-equivalent merge-commit fast-forwards; do not broaden VOC-111 allowlist |
| `scripts/foundation/voc111-deploy-staging-paths.test.mjs` | extend | Empty diff and specs-only paths do not select deploy; allowlisted paths still do |
| `tooling/governance/tests/` | create/extend | Exact pin equality; inequality with stale pins; checkout-ref/missing-develop coverage; `reconcile-production-change` dispatch; update VOC-126 action-options lists without dropping the 25-input bound |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify | Advance pin literals from `60afda3a…` to `863fc1f…` |
| `AGENTS.md` | modify current-state reconcile-release / release-authority sentences that would become false | Do not add a new internals runbook |
| `docs/operations/11-devops-and-ci-cd.md` | modify current-state | Post-promotion develop equals merge SHA; staging path selection retained |
| `docs/operations/10-development-workflow.md` | modify current-state branch strategy | After promotion merge, `develop` is advanced to that merge commit |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | modify current-state release paragraphs only | Historical correction notes stay |
| `docs/governance/16-autonomous-development-operating-model.md` | modify current-state branch/merge bullets if they would become false | Historical records unchanged |
| `specs/changes/VOC-129-.../t00-evidence.md` | update | Record pin SHA, validation commands, and #1041 / #1039 / #1035 handoff |

Ordered steps:

1. From current `develop`, create a new VOC-129 implementation branch. Do not
   check out, cherry-pick, or publish #1041 / `9d459813a733f9c6d58ad3352df0db27d33ee7f4`.
2. Fetch KARSIFT/karsift-ai-infra at exact
   `863fc1f35b1d35e4981a59166b0e939be1a2b681` into an untracked nested
   checkout. Mirror the in-scope fixture subset from that SHA. Write
   `PINNED_SHA.txt` with that SHA only. Remove the nested checkout before
   commit so it cannot stage as a gitlink.
3. Update live `pipeline.yml` to the #164 `reconcile-production-change`
   contract. Count `workflow_dispatch` inputs and fail if any block exceeds
   25. Preserve `existing_pr_number` and live `live_evidence_mode`.
4. Prove staging skip for tree-equivalent develop sync. Add or extend VOC-111
   tests. Add caller pin-consistency and checkout-ref regressions. Advance
   foundation pin literals.
5. Update current-state docs listed above. Record evidence in
   `t00-evidence.md`. This package's caller PR `Closes` only its own VOC-129
   task issue.
6. After the exact reviewed caller PR merges and is promoted, close #1041 as
   superseded, then close #1039 and #1035 with audit comments naming the
   VOC-129 merge. Do not manufacture a VOC-127 completion marker from #1041.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

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

If implementation adds narrower targeted commands (for example
`python3 -m unittest tooling.governance.tests.test_voc129_pin_consistency`),
record the exact commands in `t00-evidence.md` and run them in addition to
the suite above. Do not treat a missing suite as a pass.

Independent verifier (exact reviewed caller SHA) should confirm:

- `PINNED_SHA.txt` equals `863fc1f35b1d35e4981a59166b0e939be1a2b681` and does
  not equal `a9df74a6…` or `60afda3a…`;
- the fixture includes #164 checkout-ref ordering / missing-`develop`
  coverage and no longer restores `CHECKED_HEAD_SHA` as the post-merge
  integration tip;
- live `pipeline.yml` exposes `reconcile-production-change`, stays at most
  25 inputs, and does not expose operator SHAs;
- tree-equivalent sync does not keep staging scheduled; VOC-111 allowlist
  is not broadened;
- VOC-113 through VOC-126 isolation, recovery, App-token split, lease,
  retry limits, Cursor Composer/Grok roles, unchanged `roles.yml`, and
  non-closing source PR remain;
- current-state docs describe exact-merge-SHA sync and
  `reconcile-production-change`;
- #1041 was not merged or published; VOC-127 was not dispatched as attempt
  `3`; no snapshot-gap task and no bootstrap exception were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected. Exceptional production-change recon may introduce an already
  authorized main tree onto `develop`; allowlisted paths may then deploy
  staging as they do today.
- **Operational effect:** After each governed promotion merge performed by
  `release.yml@main` (#164), `develop` equals that merge SHA before the audit
  closes. The caller fixture and live dispatch match that contract.
  Operators can reconcile an adopted production-target task onto `develop`
  through `reconcile-production-change`.
- **Rollback trigger:** pin not equal to `863fc1f…`; #1041 published; VOC-127
  attempt `3` dispatched; unique develop commits erased; operator SHA paste
  accepted; tree-equivalent sync performing a full staging deploy; or
  `roles.yml` / OpenAI route changed.
- **Rollback mechanism:** Revert the caller workflow/helper/fixture/test/doc
  changes to the last reviewed `develop` pin
  `60afda3a44fd06b8c00b219771de7112f1aded6e`. That last-known-good still
  lacks the #164 caller contract; rollback restores a known reviewed state,
  not a fully pinned #164 caller. Do not roll back by publishing #1041.
- **Last-known-good reference:** caller `develop` before this package's merge,
  with pin `60afda3a44fd06b8c00b219771de7112f1aded6e`. Infrastructure last
  known good for the consumed contract is #164 merge
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`, which this package does not
  revert.
