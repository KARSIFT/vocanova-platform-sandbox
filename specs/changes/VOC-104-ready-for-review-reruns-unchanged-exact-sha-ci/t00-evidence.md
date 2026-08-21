# VOC-104-T00 — Evidence

Task: `VOC-104-T00` — Ready-for-review reuse policy, fail-closed path, docs, deterministic tests.

## Summary

Implemented VOC-104 ready_for_review exact-SHA evidence reuse across shared infra and the calling
repository:

- Added read-only reuse eligibility workflow (`ready-for-review-reuse.yml`) and policy modules.
- Wired caller `pipeline.yml` and infra template so `reuse-evidence` skips CI/model review while
  merge-gate still runs with validated prior-run identity.
- Extended merge-gate with reuse-aware required-check and publisher evaluation.
- Added read-only live-proof verifier (`verify-ready-for-review-reuse.yml`) for T01.
- Updated DOC-15 §17.3 and infra README to distinguish safe reuse from the full path.
- Added deterministic policy tests and caller foundation coverage.

Cross-repo note: primary reusable-workflow behavior lands in the local `karsift-ai-infra/` checkout
included with this implementation run; the calling repo consumes those workflows at
`KARSIFT/karsift-ai-infra/...@main` after infra merge.

## Commands run

```bash
cd karsift-ai-infra && python3 -m unittest tests.test_ready_for_review_reuse -v
cd .. && node --test scripts/foundation/voc104-ready-for-review-reuse.test.mjs
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## Results

| Check | Result |
| --- | --- |
| `python3 -m unittest tests.test_ready_for_review_reuse -v` (karsift-ai-infra) | PASS — 15 tests |
| `node --test scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | PASS — 4 tests |
| `bash scripts/governance/validate-governance.sh` | PASS |
| `bash scripts/governance/classify-change-risk.sh` | PASS — detected path floor R4 (DOC-15 §17.3) |
| `git diff --check` | PASS |

## Reusable workflow consumption

Calling repo `pipeline.yml` continues to consume shared workflows at `@main` (no pin bump in this
task). New jobs added:

- `KARSIFT/karsift-ai-infra/.github/workflows/ready-for-review-reuse.yml@main`
- `KARSIFT/karsift-ai-infra/.github/workflows/verify-ready-for-review-reuse.yml@main`

Existing merge-gate consumption remains `@main` with new optional inputs `reuse_outcome` and
`reuse_prior_run_id`.

## Live proof

Controlled draft→ready live proof remains operator-owned under `VOC-104-T01`
(`.karsift/live-evidence/VOC-104-T01.yaml`). T00 supplies the read-only verifier implementation and
deterministic TEST-12 coverage only.

## Limitations

- Infra changes in this run are present in the accompanying `karsift-ai-infra/` working tree and
  require a separate infra PR merge before callers observe the behavior on `@main`.
- No production credentials, logs, artifacts, or user identifiers were used in validation.
