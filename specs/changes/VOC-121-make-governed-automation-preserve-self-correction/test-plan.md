# VOC-121 — Test Plan

## VOC-121-TEST-00 — Authorized nested source edits are not silently discarded

- Covers: `VOC-121-AC-00`, `VOC-121-AC-06`
- Preconditions: fixture or extracted `implement.yml` commit/bundle steps with
  a nested `karsift-ai-infra` checkout that contains authorized tracked edits
  plus caller edits
- Procedure: Replay the current discard sequence (copy only
  `run-app-checks.sh`, `rm -rf karsift-ai-infra`, caller-only bundle) against
  the pre-fix contract if needed as a negative fixture, then run the corrected
  workflow/helper. Assert the source edits are either present in an isolated
  source bundle/PR representation or the step fails closed with recovery
  instructions naming the nested edits. A successful caller-only publish with
  those edits gone is a failure.
- Expected result: silent loss cannot pass; isolated publish or fail-loud is
  the only success class.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-01 — Caller staging never records a nested gitlink

- Covers: `VOC-121-AC-00`, `VOC-121-AC-01`
- Preconditions: nested checkout with its own `.git` present before staging
- Procedure: After the corrected commit/staging steps, inspect the caller
  index/bundle for `karsift-ai-infra` as a gitlink, submodule, or nested
  `.git`. Confirm the nested checkout was removed before `git add -A`.
- Expected result: caller commit contains no gitlink; source state is isolated
  or the job failed closed before staging.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-02 — Cross-repository publisher races and stale heads fail closed

- Covers: `VOC-121-AC-01`, `VOC-121-AC-06`
- Preconditions: isolated publisher fixture modeled on
  `tests/test_implementer_bundle.py`, extended for a second carrier when that
  path is implemented
- Procedure: Attempt publication against a changed live head, a missing
  bundle, and a lineage that is not an ancestor of the declared integration
  SHA. If the fail-loud fallback is the recorded path instead of a second
  publisher, assert the fail-loud path still refuses to publish a gitlink or
  caller-smuggled source tree.
- Expected result: stale heads, missing bundles, and unverifiable lineage are
  refused. No second-carrier publish proceeds from polluted caller state.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-03 — Self-correction still resolves helpers after nested-checkout deletion

- Covers: `VOC-121-AC-03`, `VOC-121-AC-06`
- Preconditions: extracted self-correction step plus a worktree where
  `karsift-ai-infra/` has already been removed after helper copies
- Procedure: Replay the live failure: after `rm -rf karsift-ai-infra`, invoke
  the same model-preparation command self-correction uses. Assert it runs
  against the preserved copy, not the deleted nested path. Also assert missing
  `CURSOR_API_KEY` still fails closed without printing the secret.
- Expected result: `prepare_cursor_model.py` (and `run-app-checks.sh`) are
  reachable after deletion; the old nested path is not required; credentials
  stay fail-closed.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-04 — Cross-repository PR text is fully qualified and non-closing

- Covers: `VOC-121-AC-02`
- Preconditions: publisher/PR-body construction and
  `config/cross_repo_reference.py`
- Procedure: Assert a source-repo PR body for a caller task uses
  `Relates to OWNER/CALLER#N` and that closing-keyword variants are rejected.
  Assert the caller PR still uses local `Closes #N`.
- Expected result: foreign repository text cannot manufacture caller
  completion.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-05 — Cancelled required check-run is recovered despite an alternate success

- Covers: `VOC-121-AC-04`, `VOC-121-AC-06`
- Preconditions: synthetic exact-head payload with a cancelled
  `governance-policy` pull-request check-run plus a successful
  workflow-dispatch run of the same context
- Procedure: Run recovery planning/selection. Assert the cancelled required
  check-run is not treated as satisfied and that recovery reruns or
  redispatches on the unchanged exact head (`dispatched=none` is incorrect).
- Expected result: the #993 class of evidence does not suppress recovery.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-06 — Same-named successful status does not override a cancelled check-run

- Covers: `VOC-121-AC-04`, `VOC-121-AC-06`
- Preconditions: synthetic payload with a cancelled/failed required check-run
  and a successful commit status of the same context name
- Procedure: Run recovery completeness and attestation gates. Assert the
  status is not treated as overriding the check-run while GitHub required-check
  satisfaction would still report failure. Attestation must refuse to claim
  the context is merge-ready in that state.
- Expected result: status attestation cannot paper over a ruleset-selected
  cancelled or failed check-run.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-07 — Safety gates remain unchanged or stronger

- Covers: `VOC-121-AC-05`
- Preconditions: final implementation branch
- Procedure: Inspect workflows and tests for preserved App-token isolation,
  two-attempt bounds, exact-SHA review language, Cursor fail-closed auth, and
  absence of an OpenAI/Codex or `OPENAI_API_KEY` requirement introduced by this
  package. Confirm caller automated publication still refuses caller
  `.github/workflows/**` changes.
- Expected result: reliability fixes land without weakening existing safety
  gates.
- Evidence: `VOC-121-EV-00`

## VOC-121-TEST-08 — Source self-CI, caller suites, docs, and pin match consumption

- Covers: `VOC-121-AC-07`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: Run infra unit suite and caller governance/fixture suites listed
  in `implementation-plan.md`. Confirm current-state comments no longer claim
  silent discard is safe or that status attestation satisfies a cancelled
  required check-run if GitHub does not. If the fixture consumed the infra
  change, assert the pin equals the exact reviewed infra merge. If not, assert
  the pin is unchanged and evidence records why.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA and only when applicable.
- Evidence: `VOC-121-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
