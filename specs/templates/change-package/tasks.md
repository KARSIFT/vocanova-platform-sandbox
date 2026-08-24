# VOC-000 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

The default for a coherent objective is **one end-to-end implementation task**
covering code, tests, documentation, migration/config updates, and acceptance
evidence in the same carrier unless a concrete split boundary requires more.
Every task after the first must record `- Split reason: <allowed-slug>` from the
planner/plan-review allowlist. Packages with more than three tasks require a
`## Package-level multi-task justification` section.

## VOC-000-T00 — REPLACE WITH ONE END-TO-END OUTCOME

- Requirement source: `VOC-000-D00`
- Acceptance criteria: `VOC-000-AC-00`
- Tests: `VOC-000-TEST-00`
- Evidence: `VOC-000-EV-00`
- Status: pending

### Optional: operator-owned live evidence

When a task's acceptance depends on a live GitHub Actions run the implementer
cannot dispatch (production deploy proof, scheduled synthetic on `main`, etc.):

- Add the exact marker inside this task's own `## <task_id>` stanza:
  `- Automation ownership: operator` or `- Automation ownership: live-actions`
  (structural match only; narrative prose is never parsed).
- State in the task body that acceptance requires **operator-owned live
  evidence** (not implementer Actions access).
- Add `<package>/.karsift/live-evidence/<task_id>.yaml` with the allowlisted
  contract fields documented in
  [`docs/operations/live-evidence.md`](../../../docs/operations/live-evidence.md).
- Record in the task's `tNN-evidence.md` what the operator must trigger or wait
  for, using allowlisted metadata only (no logs, secrets, or tokens).
- Do not expand scope into unrelated workflow or pipeline edits to manufacture
  evidence; waiting and reconcile are handled by governed automation after
  VOC-097-T01/T02.

Tasks must preserve scope, separation of duties, and rollback safety.
