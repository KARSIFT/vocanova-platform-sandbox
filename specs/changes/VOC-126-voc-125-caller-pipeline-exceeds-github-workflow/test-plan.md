# VOC-126 — Test Plan

## VOC-126-TEST-00 — Every workflow_dispatch block has at most 25 inputs

- Covers: `VOC-126-AC-00`
- Preconditions: final live caller workflows and
  `karsift-ai-infra/templates/project-repo/.github/workflows/*`; no secrets or
  production data
- Procedure: For every workflow file in those two trees that declares
  `on.workflow_dispatch`, count keys directly under `inputs`. Assert each
  count is `<= 25`. Include at least live `.github/workflows/pipeline.yml`,
  the dedicated verifier workflow, and the matching project-repo templates.
  Keep this as a regression of Actions run `32977045898`, not only as a
  comment. Do not treat a 26-input fixture of the VOC-125 template as a
  passing result.
- Expected result: GitHub can accept every live `workflow_dispatch`
  definition; the #1025 26-input class fails the new assertion.
- Evidence: `VOC-126-EV-00`

## VOC-126-TEST-01 — Caller and template pipeline expose existing_pr_number only

- Covers: `VOC-126-AC-01`
- Preconditions: final live `.github/workflows/pipeline.yml` and
  `karsift-ai-infra/templates/project-repo/.github/workflows/pipeline.yml`
- Procedure: Assert both files' `workflow_dispatch` inputs include
  `existing_pr_number` described as implement-only existing-carrier recovery.
  Assert the `implement:` job `with:` block forwards
  `existing_pr_number: ${{ inputs.existing_pr_number }}` (or the live
  equivalent) and still forwards `attempt`. Assert neither file declares
  operator-typed `expected_head_sha` or `expected_base_sha` under
  `workflow_dispatch`.
- Expected result: operator resume identity is a PR number on `pipeline.yml`;
  free-form SHA paste is not the caller interface.
- Evidence: `VOC-126-EV-00`

## VOC-126-TEST-02 — Five read-only verifiers exist on the dedicated workflow

- Covers: `VOC-126-AC-02`
- Preconditions: final dedicated verifier workflow (preferred
  `pipeline-verify.yml`) in live caller and project-repo template
- Procedure: Assert these jobs exist, call the same reusable workflows at
  `@main`, and forward the same named inputs they forwarded from
  `pipeline.yml` before relocation:

  | Job | Reusable workflow | Required forwarded inputs |
  |-----|-------------------|---------------------------|
  | `verify-auto-advance-live-evidence` | `verify-auto-advance-live-evidence.yml` | `verify_source_run_id`, `verify_waiting_pr_number`, `verify_change_id`, `verify_task_id`, `verify_package_path` |
  | `verify-ready-for-review-reuse` | `verify-ready-for-review-reuse.yml` | `verify_ready_run_id`, `verify_prior_run_id`, `verify_reuse_source_pr_number`, `verify_reuse_source_head_sha`, `verify_reuse_source_base_sha` |
  | `verify-remediate-operator-ownership` | `verify-remediate-operator-ownership.yml` | `live_evidence_run_id`, `live_evidence_pr_number`, `change_id`, `task_id`, `package_path` |
  | `verify-promotion-check-recovery` | `verify-promotion-check-recovery.yml` | `promotion_pr_number` |
  | `verify-post-promotion-workflow` | `verify-post-promotion-workflow.yml` | `promotion_pr_number` |

  Assert `pipeline.yml` no longer routes those five `inputs.action` values.
- Expected result: verifier capabilities are relocated, not deleted.
- Evidence: `VOC-126-EV-00`

## VOC-126-TEST-03 — Mutating operator-loop actions remain on pipeline.yml

- Covers: `VOC-126-AC-02`
- Preconditions: final live and template `pipeline.yml`
- Procedure: Assert `workflow_dispatch` action options still include
  `implement`, `plan`, `reconcile`, `reconcile-release`,
  `reconcile-live-evidence`, `recover-integration-push`, and
  `recover-promotion-pr-checks`. Assert the corresponding jobs still exist
  and still call the same reusable workflows. Update any source or caller
  test that currently requires the five verify actions to remain in that
  options list (`test_adoption_handoff.py`,
  `test_voc080_workflow_policy.py`, `test_voc080_adoption_reconcile_policy.py`,
  `test_auto_advance_ownership.py`, `test_remediation_ownership.py`, and
  equivalents) so those tests encode the relocated contract.
- Expected result: recover/reconcile/implement/plan remain reachable on
  `pipeline.yml`.
- Evidence: `VOC-126-EV-00`

## VOC-126-TEST-04 — Dedicated verifier workflow is read-only

- Covers: `VOC-126-AC-03`
- Preconditions: final dedicated verifier workflow
- Procedure: Assert the workflow and its jobs do not use `secrets: inherit`,
  do not mint a GitHub App token, and do not grant `actions: write`. Assert
  job permissions remain read-only (`actions: read`, `contents: read`, plus
  `issues` / `pull-requests` read where those jobs already declared them).
  Assert `recover-integration-push` and `recover-promotion-pr-checks` are
  absent from the dedicated workflow and still present on `pipeline.yml`.
- Expected result: relocation does not expand verifier privilege or mix
  mutating recovery into the read-only surface.
- Evidence: `VOC-126-EV-00`

## VOC-126-TEST-05 — VOC-125 resume and attempt contracts remain

- Covers: `VOC-126-AC-04`
- Preconditions: final `implement.yml`, `remediate.yml`, and caller pipeline
- Procedure: Reuse or extend VOC-125 policy tests to assert
  `implement.yml` still declares `existing_pr_number` and still binds before
  `Create implementation branch`; `remediate.yml` retry still forwards
  event-derived SHAs and `pr_number` as `existing_pr_number`; attempt `3`
  still fails closed; attempt `1` with an existing open carrier still fails
  closed. Do not reopen VOC-125 mismatch-class implementation in this
  package; assert the contracts remain.
- Expected result: the input-limit repair does not regress existing-carrier
  resume identity.
- Evidence: `VOC-126-EV-00`

## VOC-126-TEST-06 — Isolation, leases, roles, and credentials remain

- Covers: `VOC-126-AC-04`
- Preconditions: final implementation branch
- Procedure: Inspect `implement.yml`, `remediate.yml`, and existing
  VOC-121/VOC-123/VOC-124 policy tests. Confirm: named-ref `create-bundle`
  still used; `publish-source` still requires App credentials with
  `permission-workflows: write` and no `|| github.token`; caller `publish`
  still omits workflow-write and still refuses `.github/workflows/**`;
  force-with-lease still uses expected heads; two-attempt bound unchanged;
  source PR body still `Relates to` without a closing keyword; this package's
  caller PR `Closes` only its own VOC-126 task issue; implementer permissions
  still omit `actions`; Cursor Composer remains the implementer route and
  Cursor Grok remains the reviewer route; no OpenAI execution path is
  re-enabled; no App token on the model-controlled runner; no secrets in
  bundles.
- Expected result: dispatch relocation lands without weakening prior safety
  gates.
- Evidence: `VOC-126-EV-00`

## VOC-126-TEST-07 — Source self-CI, caller suites, docs, pin, and handoff

- Covers: `VOC-126-AC-05`, `VOC-126-AC-06`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: Run infra unit suite and caller governance/fixture suites listed
  in `implementation-plan.md`. If the fixture consumed the infra change,
  assert the pin equals the exact reviewed infra merge and is not
  `1f1705dbad41729563b0ad1e878e4154e5511e93`, and that matching foundation
  pin literals match. If not, assert the pin is unchanged and evidence
  records why. Assert current-state comments/README describe attempt `2`
  plus `existing_pr_number` on `pipeline.yml` and verifier dispatch on the
  dedicated workflow. Confirm `t00-evidence.md` records that #1024 is not
  merged, that #1022 / #1020 close only after the live route is valid, that
  #1003 / #1012 remain the existing VOC-122 carrier to resume against that
  revision at attempt `2` with `existing_pr_number=1012`, and that no
  bootstrap exception was used.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA and only when applicable; VOC-122 is not reimplemented here;
  VOC-125 is not retried as attempt `3`.
- Evidence: `VOC-126-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
