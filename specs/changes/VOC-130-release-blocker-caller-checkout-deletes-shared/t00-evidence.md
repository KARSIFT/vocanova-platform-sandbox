# VOC-130-T00 — Evidence

Task: `VOC-130-T00` — Pin exact infra #165 and restore shared policy after
caller checkout.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1047)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1047 |
| VOC-129 caller PR | #1046 at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Failing/no-op release wake-up | `33066533397` |
| Selected checkout ref | `develop` |
| Observed failure | `karsift-ai-infra/config/task-completion-runner.py` missing after caller root checkout |
| Observed release result | missing validator treated as safe no-op; no audit/promotion; converge skipped |
| Root cause | root caller checkout deletes nested shared policy; restore must follow caller checkout before lifecycle helpers |
| Authoritative infra merge | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (KARSIFT/karsift-ai-infra#165) |
| Independently reviewed infra head | `e33931d02f7bdbb094ae8177fd88324cd19ac5ce` |
| Infra verification | 429 policy tests plus hosted actionlint, shellcheck, YAML parsing, and policy checks |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| Why VOC-129 is not retried | #1046 already merged; this is a new checkout-lifetime defect |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-130 carrier |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-130 branch/PR from current `develop`; not a VOC-129 rewrite |
| Infra | consume already-merged #165; do not open a replacement infra PR |
| Pin target | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| Pin must not equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Restore | both `identify` and `converge` restore shared policy after caller checkout and before task-completion helpers |
| Restore identity | `job.workflow_repository` + `job.workflow_sha` + `path: karsift-ai-infra` |
| Preserve | #164 missing-`develop` recovery, exact-SHA develop sync, unique-develop fail-closed, promotion checks, serialization, review/implementer split, retry bounds, secret/raw-error controls |
| Exceptional action | live identity remains `reconcile-production-change` |
| Operator identity | adopted authority issue number; no free-form SHA inputs on caller `workflow_dispatch` |
| `existing_pr_number` | remains implement-only |
| VOC-129 | do not re-implement #1046; promote through repaired `release.yml@main` / `reconcile-release` |
| Attempt | VOC-130-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces (implementation)

**Infrastructure (`KARSIFT/karsift-ai-infra`):** already merged as
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (PR #165). This task does not
change infrastructure.

**Caller (`vocanova-platform-sandbox`):**

- `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/release.yml`
  mirrored from infra #165 with `Restore shared lifecycle policy after caller
  checkout` in both `identify` and `converge`.
- `tooling/governance/fixtures/karsift-ai-infra/tests/test_release_policy.py`
  mirrored from infra #165, including
  `test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers`.
- `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` set to
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
- `tooling/governance/fixtures/karsift-ai-infra/README.md` current-state pin
  paragraph advanced to #165 and documents the post-caller-checkout restore.
- `tooling/governance/tests/test_voc130_shared_policy_restore.py` added for pin,
  restore ordering, #164 contract preservation, and VOC-130 carrier identity.
- Caller governance pin literals and foundation tests advanced from
  `863fc1f…` to `8ce2b77…`.
- Live `.github/workflows/pipeline.yml` unchanged (`VOC-130-DEP-07` expected
  no-op; caller already dispatches `release.yml@main`).

**Non-consumption:** unrelated #165 files outside the caller fixture subset were
not copied merely to force the pin.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | KARSIFT/karsift-ai-infra#165 (already merged) |
| Exact infra merge SHA | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| Pin applicable? | **yes** |
| Pin must equal | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| Pin must not equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| `PINNED_SHA.txt` | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| Bootstrap used? | **no** |
| Snapshot-gap task added? | **no** |
| VOC-129 #1046 re-implemented? | **no** |

## Validation commands and results

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node scripts/foundation/voc097-fixture-matrix.test.mjs
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs
python3 -m unittest tooling.governance.tests.test_voc130_shared_policy_restore
git diff --check
```

| Command | Result |
|---------|--------|
| `validate-governance.sh` | pass |
| `classify-change-risk.sh` | pass (path floor only; no PR declaration in local run) |
| caller governance unittest suite | pass (230 tests) |
| fixture unittest suite | pass (379 tests) |
| `voc097-fixture-matrix.test.mjs` | pass (5 tests) |
| `voc104-ready-for-review-reuse.test.mjs` | pass |
| `voc108-authoritative-lifecycle.test.mjs` | pass (5 tests) |
| `test_voc130_shared_policy_restore` | pass (9 tests) |
| `git diff --check` | pass |

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. The
implementer must not approve or merge its own work.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- VOC-129 skipped promotion from run `33066533397` completes through the
  repaired `release.yml@main` path or `reconcile-release` when a valid
  App-authored completion marker exists.
- This package promotes through the same path.
- `develop` is advanced to each successful promotion merge SHA before audit
  close.
- Release/task/requirement records close with audit comments naming both
  exact promotion merges.
- Closed state alone is not completion proof.
