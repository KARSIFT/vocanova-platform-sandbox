# VOC-106 — Stop remediation dispatching the general implementer for operator-owned live-evidence failures

| Field                     | Value                                                                                      |
| ------------------------- | ------------------------------------------------------------------------------------------ |
| Package                   | `VOC-106`                                                                                  |
| Title                     | Stop remediation dispatching the general implementer for operator-owned live-evidence failures |
| Path                      | `specs/changes/VOC-106-operator-owned-live-evidence-failures-can`                          |
| Status                    | `draft`                                                                                    |
| Risk                      | `R3` (draft proposal; independently classified per task)                                   |
| Authority model           | A-004 active                                                                               |
| Requirement source        | GitHub issue [#882](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/882)       |
| Target branch             | `develop`                                                                                  |
| Approval                  | `not-approved`                                                                             |
| Implementation authorized | `false`                                                                                    |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule)                                                 |

## Problem

When an operator-owned live-evidence task receives a genuine independent-review
`FAIL` (or a CI failure) on its evidence carrier, remediation currently selects
the general implementer even though the adopted contract declares
`ownership: operator` and the task excludes code implementation.

`WAITING FOR OPERATOR LIVE EVIDENCE` already suppresses remediation retry.
A genuine `FAIL` on the same operator-owned carrier can still become an
implementation retry because `decide-remediation.py` / `remediate.yml` decide
from CI/review state and PR-body task identity without independently resolving
the adopted `.karsift/live-evidence/<task_id>.yaml` contract.

Sanitized observed evidence from issue #882: during supervised completion of
`VOC-104-T01`, a documentation-only edit after live-evidence reconciliation
invalidated the exact-head attestation. Independent review correctly failed
closed. Remediation then selected the general implementer. Workflow
`32529860337` on carrier PR `#879` was cancelled before branch mutation; the
branch head remained unchanged. Subsequent operator-owned requalification,
exact-head independent review, merge, staging deployment, and production
deployment all passed.

Drafting-time read of `karsift-ai-infra/config/decide-remediation.py` confirms
`RETRY` on `ci_failed` or review `FAIL` with no ownership input.

## Required outcome (summary)

1. Resolve task ownership from the exact reviewed PR head and the adopted
   live-evidence contract before any remediation retry.
2. Never dispatch the general implementer when a valid contract establishes an
   operator-owned or live-evidence-only task and review FAIL or CI failure
   occurs.
3. Keep merge behavior fail-closed.
4. Route operator FAIL/CI to a sanitized, bounded operator/reconcile escalation;
   preserve stale no-op and missing-evidence handling; route malformed or
   mismatched ownership metadata to a separate fail-closed escalation. No path
   may leak logs, identity data, credentials, or evidence payloads.
5. Preserve normal bounded remediation for implementer-owned tasks.
6. Treat stale/missing evidence lifecycle states deliberately so they cannot
   deadlock or consume an implementation attempt.
7. Add deterministic workflow-policy and decision-helper tests covering
   operator WAITING, FAIL, CI failure, stale result, malformed contract, and
   ordinary implementer FAIL and CI failure.
8. Prove the behavior through a controlled, sanitized, non-destructive live
   workflow event without manufacturing unrelated evidence or exposing secrets.
9. Keep this root focused; Node runtime deprecations, action-input migration,
   Go cache configuration, dependency alerts, and application changes are
   separate follow-ups.

## Tasks

| Task | Summary                                                                              | Depends on |
| ---- | ------------------------------------------------------------------------------------ | ---------- |
| T00  | Remediation ownership gate, fail-closed escalation, docs, deterministic tests        | —          |
| T01  | Controlled sanitized workflow proof (operator-owned live evidence)                   | T00        |

See `tasks.md` for full task definitions.

## What this package deliberately does NOT do

- Grant the implementer general GitHub Actions credentials.
- Change production application behavior, signup policy, secrets, databases, or
  Kuma/synthetic inventory IDs.
- Reopen VOC-104 or VOC-102 code work, or manufacture live evidence for
  unrelated packages.
- Address Node runtime deprecations, action-input migration, Go cache
  configuration, or dependency alerts (explicit follow-ups outside this package).
- Weaken VOC-097 waiting/reconcile semantics or merge-gate fail-closed behavior.
- Self-adopt or self-authorize this package.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
