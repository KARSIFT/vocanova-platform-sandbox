# VOC-126 — Acceptance Criteria

## VOC-126-AC-00 — Every live workflow_dispatch block has at most 25 inputs

- Requirement source: `VOC-126-D01`, `VOC-126-D06`
- Tasks: `VOC-126-T00`
- Tests: `VOC-126-TEST-00`
- Evidence: `VOC-126-EV-00`
- Result: pending

On the exact reviewed revisions, every project-repo template workflow and
every live caller workflow that declares `on.workflow_dispatch` has at most
25 keys under `inputs`. GitHub can accept the workflow definition. The
#1025 / run `32977045898` class (26 inputs on `pipeline.yml`) is a failing
result. Deterministic tests assert the maximum; comments alone are not
coverage.

## VOC-126-AC-01 — existing_pr_number remains the operator resume identity

- Requirement source: `VOC-126-D02`
- Tasks: `VOC-126-T00`
- Tests: `VOC-126-TEST-01`
- Evidence: `VOC-126-EV-00`
- Result: pending

Caller `pipeline.yml` (live and template) exposes implement-only
`existing_pr_number` and forwards it to `implement.yml`. Caller
`workflow_dispatch` does not expose operator-typed SHA inputs. A valid
attempt-2 resume still uses `action=implement`, `attempt=2`, and
`existing_pr_number=<open PR>` on `pipeline.yml`.

## VOC-126-AC-02 — Read-only verifier capabilities are relocated, not deleted

- Requirement source: `VOC-126-D01`, `VOC-126-D03`, `VOC-126-D04`
- Tasks: `VOC-126-T00`
- Tests: `VOC-126-TEST-02`, `VOC-126-TEST-03`
- Evidence: `VOC-126-EV-00`
- Result: pending

The five read-only verifier jobs exist on the dedicated caller workflow,
still call the same reusable workflows at `@main`, and still forward the
same named inputs. `pipeline.yml` still exposes and routes `implement`,
`plan`, `reconcile`, `reconcile-release`, `reconcile-live-evidence`,
`recover-integration-push`, and `recover-promotion-pr-checks`. No active
recovery or verifier capability is dropped.

## VOC-126-AC-03 — Verifier workflow stays read-only; recovery stays mutating on pipeline.yml

- Requirement source: `VOC-126-D03`, `VOC-126-D04`
- Tasks: `VOC-126-T00`
- Tests: `VOC-126-TEST-04`
- Evidence: `VOC-126-EV-00`
- Result: pending

The dedicated verifier workflow does not use `secrets: inherit`, does not
mint an App token, and does not grant `actions: write`.
`recover-integration-push` and `recover-promotion-pr-checks` remain on
`pipeline.yml`.

## VOC-126-AC-04 — VOC-125 resume, attempt, lease, and isolation contracts remain

- Requirement source: `VOC-126-D02`, `VOC-126-D05`
- Tasks: `VOC-126-T00`
- Tests: `VOC-126-TEST-05`, `VOC-126-TEST-06`
- Evidence: `VOC-126-EV-00`
- Result: pending

The change does not weaken exact-head/base binding, two-attempt implementer
bound, `remediate.yml` event-derived SHA forwards, `existing_pr_number`
forward on automatic retry, nested-repository isolation, gitlink refusal,
VOC-123 named-ref bundle tips, credential-free bundles, clean
`publish-source` App-token separation, caller workflow-file refusal,
force-with-lease, Cursor Composer implementer, or Cursor Grok exact-revision
review. No OpenAI route is added. No credential values are printed. Attempt
`3` remains forbidden.

## VOC-126-AC-05 — Source merge first, caller pin, docs, and superseded #1024 handoff

- Requirement source: `VOC-126-D07`, `VOC-126-D08`, `VOC-126-D09`
- Tasks: `VOC-126-T00`
- Tests: `VOC-126-TEST-07`
- Evidence: `VOC-126-EV-00`
- Result: pending

The independently reviewed infrastructure template/test repair merges first.
If the caller fixture consumes that change, `PINNED_SHA.txt` equals that
exact infra merge SHA (not `1f1705d…`) and matching caller pin assertions
are advanced. Current-state comments/docs describe verifier dispatch through
the dedicated workflow and operator resume as `pipeline.yml` attempt `2`
plus `existing_pr_number`. Historical audit records are not rewritten. No
bootstrap exception was used. Caller PR #1024 is not merged; evidence
records that it is superseded only after the governed replacement exists.

## VOC-126-AC-06 — VOC-125 closure and existing VOC-122 resume handoff

- Requirement source: `VOC-126-D08`
- Tasks: `VOC-126-T00`
- Tests: `VOC-126-TEST-07`
- Evidence: `VOC-126-EV-00`
- Result: pending

This package's implementation PR closes only its own VOC-126 task issue.
After the live caller route is valid, reviewed, merged, and promoted,
evidence records that #1024, VOC-125 task #1022, and origin #1020 are closed
as superseded-unusable-carrier with audit comments bound to the VOC-126
exact SHA, that no VOC-125 completion marker claims #1024 merged, and that
existing `VOC-122-T00` / #1003 / #1012 is the carrier to resume through the
repaired route at attempt `2` with `existing_pr_number=1012`. No replacement
VOC-122 task or PR was created.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
