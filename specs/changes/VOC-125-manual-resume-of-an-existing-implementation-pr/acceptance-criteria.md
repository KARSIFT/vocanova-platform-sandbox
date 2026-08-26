# VOC-125 — Acceptance Criteria

## VOC-125-AC-00 — Operator resume supplies existing-PR recovery identity

- Requirement source: `VOC-125-D01`, `VOC-125-D02`
- Tasks: `VOC-125-T00`
- Tests: `VOC-125-TEST-00`, `VOC-125-TEST-01`
- Evidence: `VOC-125-EV-00`
- Result: pending

On the exact reviewed revisions, caller `pipeline.yml` (live and template)
exposes implement-only `existing_pr_number` and forwards it to
`implement.yml`. `implement.yml` declares that input. A valid attempt-2
resume with that PR number derives `expected_head_sha` and
`expected_base_sha` from the live open PR and binds them before
`Create implementation branch` and before model resolution. Caller
`workflow_dispatch` does not expose operator-typed SHA inputs. The
#1020 / job `98170418081` class (attempt `2`, existing branch, empty SHAs, no
PR number) remains a failing result.

## VOC-125-AC-01 — Existing carrier, issue, branch, and PR are reused

- Requirement source: `VOC-125-D04`
- Tasks: `VOC-125-T00`
- Tests: `VOC-125-TEST-01`, `VOC-125-TEST-02`
- Evidence: `VOC-125-EV-00`
- Result: pending

A valid existing-carrier resume continues the same task issue, deterministic
task branch, and open PR. It does not create a replacement branch or PR,
does not delete the existing branch, and does not reopen a closed PR.
Publication leases continue to use the bound expected heads.

## VOC-125-AC-02 — Attempt number and one-retry maximum are preserved

- Requirement source: `VOC-125-D02`
- Tasks: `VOC-125-T00`
- Tests: `VOC-125-TEST-03`
- Evidence: `VOC-125-EV-00`
- Result: pending

Only attempts `1` and `2` are accepted. Attempt `3` fails closed. Attempt `1`
while an existing open task PR or remote task branch exists fails closed.
A valid resume of an existing carrier uses attempt `2` and does not
reclassify that attempt as `1`.

## VOC-125-AC-03 — Mismatch classes fail closed before model or mutation

- Requirement source: `VOC-125-D03`
- Tasks: `VOC-125-T00`
- Tests: `VOC-125-TEST-04`
- Evidence: `VOC-125-EV-00`
- Result: pending

Missing, malformed, or stale head or base; wrong PR, branch, repository,
task, package, or authority issue; foreign or malformed App-signed review
evidence when a review exists; changed remote head; and already-closed or
completed tasks all fail closed before model resolution, Composer execution,
source-bundle creation, or publication. Absent App-signed review remains
allowed only for the existing CI-failure class, still bound to the live open
PR and remote branch.

## VOC-125-AC-04 — Automatic remediate retry still supplies trusted SHAs

- Requirement source: `VOC-125-D01`, `VOC-125-D02`
- Tasks: `VOC-125-T00`
- Tests: `VOC-125-TEST-05`
- Evidence: `VOC-125-EV-00`
- Result: pending

`remediate.yml` retry continues to pass event-derived `expected_head_sha` and
`expected_base_sha` and also forwards `pr_number` as `existing_pr_number`.
When both are present they must match the live PR. Automatic retry does not
depend on an operator typing SHAs.

## VOC-125-AC-05 — Publication leases, isolation, roles, and credentials remain

- Requirement source: `VOC-125-D05`
- Tasks: `VOC-125-T00`
- Tests: `VOC-125-TEST-06`
- Evidence: `VOC-125-EV-00`
- Result: pending

The change does not weaken nested-repository isolation, gitlink refusal,
VOC-123 named-ref bundle tips, credential-free bundles, clean
`publish-source` App-token separation, `permission-workflows: write` only on
that mint, caller workflow-file refusal, force-with-lease, two-attempt bound,
non-closing source PR, caller `Closes #N`, Cursor Composer implementer, or
Cursor Grok exact-revision review. No OpenAI route is added. No credential
values are printed.

## VOC-125-AC-06 — Current-state docs, pin, and existing VOC-122 resume handoff

- Requirement source: `VOC-125-D07`, `VOC-125-D08`, `VOC-125-D09`
- Tasks: `VOC-125-T00`
- Tests: `VOC-125-TEST-07`
- Evidence: `VOC-125-EV-00`
- Result: pending

Current-state workflow comments/docs describe operator resume as attempt `2`
plus `existing_pr_number`, not free-form SHA paste. Historical audit records
are not rewritten. If the caller fixture consumes the infrastructure change,
`PINNED_SHA.txt` equals the exact infra merge SHA and matching caller pin
assertions are advanced. After that merge is live on `implement.yml@main` and
the caller dispatch contract is merged, evidence records that existing
`VOC-122-T00` / #1003 / #1012 is the carrier to resume through the repaired
route at attempt `2` with `existing_pr_number=1012`, and that no replacement
VOC-122 task or PR was created. No bootstrap exception was used.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
