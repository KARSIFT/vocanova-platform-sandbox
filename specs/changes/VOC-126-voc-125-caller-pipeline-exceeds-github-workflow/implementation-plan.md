# VOC-126 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` project-repo pipeline template
  and tests; caller `.github/workflows/pipeline.yml` and the dedicated
  verifier workflow; operator implement-resume identity; read-only verifier
  routing; caller `tooling/governance/` fixtures and tests.
- Prerequisites: confirm the VOC-125 source template still has 26
  `workflow_dispatch` inputs including `existing_pr_number`. Confirm live
  caller `pipeline.yml` still has 25 inputs and does not yet declare
  `existing_pr_number`. Confirm `implement.yml@main` still declares
  `existing_pr_number` from infrastructure merge
  `1f1705dbad41729563b0ad1e878e4154e5511e93`. Confirm the five read-only
  verifier jobs still live on `pipeline.yml`. Confirm VOC-121 through
  VOC-125 publisher and resume contracts remain the baseline this change
  must preserve.
- No bootstrap exception. VOC-124 already published
  `permission-workflows: write` on `publish-source`. T00's first run is
  attempt `1`. Do not treat an untracked local `karsift-ai-infra/` checkout
  as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions
  or rotate `KARSIFT_BOT_*` secrets.
- Do not merge PR #1024. Do not dispatch VOC-125-T00 as attempt `3`.

## File reconciliation and implementation sequence

### T00 — Relocate read-only verifier dispatch under the 25-input limit

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/templates/project-repo/.github/workflows/pipeline.yml` | modify | Keep `existing_pr_number`; remove the five verify jobs and their dedicated inputs; keep mutating operator-loop actions; update current-state comments |
| `KARSIFT/karsift-ai-infra/templates/project-repo/.github/workflows/pipeline-verify.yml` | create | Preferred name (`VOC-126-DEP-07`); five read-only verifier jobs; same reusable `@main` calls and named inputs; read-only permissions |
| `KARSIFT/karsift-ai-infra/tests/` | create/extend | Explicit ≤25 input-count regression; update hard-coded `pipeline.yml` action-options assertions (`test_adoption_handoff.py`, `test_auto_advance_ownership.py`, `test_remediation_ownership.py`, and equivalents) |
| `KARSIFT/karsift-ai-infra/README.md` | modify | Current-state verifier-dispatch and implement-resume paragraphs |
| `.github/workflows/pipeline.yml` | modify | Same dispatch split as the repaired template; add `existing_pr_number`; preserve live `live_evidence_mode` unless relocation requires touching it |
| `.github/workflows/pipeline-verify.yml` | create | Match the template dedicated workflow |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge when consumed; do not pin `1f1705d…` |
| `tooling/governance/tests/` | modify/extend | Fixture regressions; caller 25-input assertion; advance pin literal when consumed; update hard-coded action-options lists |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify if they still assert the previous pin | Same-task pin literals (`f406cc95a3f853e8aef5bf8bcf22d37a29d64547` at drafting time in `PINNED_SHA.txt`) |
| `AGENTS.md` | modify only if an existing sentence would become false | Do not add a new verifier-dispatch runbook |
| `specs/changes/VOC-126-.../t00-evidence.md` | update | Record mechanism, validation commands, infra SHA, pin applicability, #1024 / #1022 / #1003 / #1012 handoff |

Ordered steps:

1. In a clean isolated `KARSIFT/karsift-ai-infra` worktree based on current
   `main` (which already contains VOC-125 `existing_pr_number` on the
   26-input template), relocate the five read-only verifier jobs to the
   dedicated template workflow and keep `existing_pr_number` on
   `pipeline.yml`. Count inputs on every template `workflow_dispatch` block
   and fail the new regression if any exceeds 25.
2. Update source tests that currently require the five verify actions to
   remain in the `pipeline.yml` options list. Add the maximum-input-count
   regression. Update current-state comments/README.
3. Run the infra unit/policy suite. Open one reviewed infra PR that
   `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and does not use a
   closing keyword. Merge it first so the project-repo template is
   GitHub-valid before the caller consumes it.
4. After a different actor merges that exact reviewed infra head, update the
   live caller workflows to match: add `existing_pr_number`, relocate the
   five verifier jobs, add the dedicated workflow. Sync and pin the caller
   fixture to that exact merge SHA when consumed; update caller governance
   and foundation pin tests; record evidence in `t00-evidence.md`. Do not
   merge #1024; this caller PR is the replacement.
5. After the exact reviewed caller dispatch is merged and promoted, record
   the superseded #1024 close, the VOC-125 #1022 / #1020 close condition, and
   the repaired dispatch command for existing `VOC-122-T00`. Do not create a
   replacement VOC-122 task or PR. Do not merge #1012 from this package. Do
   not dispatch VOC-125 as attempt `3`.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'

# In this caller repository:
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

If implementation adds narrower targeted commands (for example
`python3 -m unittest tests.test_voc126_workflow_dispatch_input_limit` or the
three foundation pin tests), record the exact commands in `t00-evidence.md`
and run them in addition to the suite above.

Independent verifier (exact reviewed caller SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- every live and template `workflow_dispatch` block has at most 25 inputs;
- caller/template `pipeline.yml` expose and forward `existing_pr_number` and
  do not expose operator SHA inputs;
- the five read-only verifier jobs exist on the dedicated workflow with the
  same reusable calls and named inputs;
- `pipeline.yml` still routes implement/plan/reconcile/recover actions;
- the dedicated verifier workflow is read-only;
- VOC-121 through VOC-125 isolation, named-ref bundle, App-token split,
  lease, retry limits, Cursor Composer/Grok roles, and non-closing source PR
  remain;
- carrier current-state text describes attempt `2` plus `existing_pr_number`
  on `pipeline.yml` and verifier dispatch on the dedicated workflow;
- the caller fixture pin equals the exact reviewed infra merge when the
  fixture consumes the change, or evidence records why the pin was not
  applicable — and the pin is not `1f1705d…`;
- #1024 / VOC-122 / #1012 behavior was not merged in this package;
- VOC-125 was not dispatched as attempt `3`;
- the implementer did not approve or merge its own work on either carrier;
- no bootstrap exception was used.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** GitHub accepts the caller pipeline definition.
  A governed operator can resume an existing implementation PR by dispatching
  `pipeline.yml` `action=implement` with `attempt=2` and
  `existing_pr_number=<PR>`. Read-only verifiers are dispatched through the
  dedicated workflow. The already-authorized VOC-122-T00 carrier
  (#1003 / #1012) can use the repaired implement route.
- **Rollback trigger:** Any live `workflow_dispatch` block exceeds 25 inputs;
  `existing_pr_number` disappears from `pipeline.yml`; a verifier or recover
  capability is missing; the verifier workflow gains `secrets: inherit` or
  `actions: write`; operator SHA paste becomes the caller interface; or
  attempt `3` is accepted.
- **Rollback mechanism:** Revert the infra and caller workflow/fixture/test/
  doc changes to the last reviewed GitHub-valid caller dispatch
  (`pipeline.yml` at 25 inputs without `existing_pr_number`) plus the last
  reviewed `implement.yml` bind contract. That last-known-good still lacks a
  GitHub-valid `existing_pr_number` caller interface; rollback restores a
  known reviewed state, not a working existing-carrier resume.
- **Last-known-good reference:** Live caller `pipeline.yml` on `develop`
  before this package lands (25 inputs, no `existing_pr_number`) and
  infrastructure `implement.yml` after VOC-125 merge `1f1705d…`. Do not roll
  the caller back to PR #1024 / `8621f12…`; that revision is GitHub-invalid.
