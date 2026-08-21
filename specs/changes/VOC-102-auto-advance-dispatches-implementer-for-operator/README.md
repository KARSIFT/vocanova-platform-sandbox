# VOC-102 — Stop auto-advance dispatching implementer for operator-owned live-evidence tasks

| Field                     | Value                                                                                |
| ------------------------- | ------------------------------------------------------------------------------------ |
| Package                   | `VOC-102`                                                                            |
| Title                     | Stop auto-advance dispatching implementer for operator-owned live-evidence tasks     |
| Path                      | `specs/changes/VOC-102-auto-advance-dispatches-implementer-for-operator`             |
| Status                    | `adopted`                                                                            |
| Risk                      | `R4` (implemented workflow path floor; independently verified per task)              |
| Authority model           | A-004 active                                                                         |
| Requirement source        | GitHub issue [#863](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/863) |
| Target branch             | `develop`                                                                            |
| Approval                  | `autonomously-adopted-after-independent-verification`                                |
| Implementation authorized | `true`                                                                               |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule)                                           |

## Problem

When an implementation task closes and the next roster task is explicitly
operator-owned live evidence, `auto-advance.yml` still dispatches the general
implementer. That run has no legitimate implementation work, consumes
runner/model capacity, and creates an unnecessary opportunity for evidence files
to be modified before real live proof exists.

Sanitized observed evidence from issue #863: closing `VOC-098-T00` produced
pipeline run `32462184971`. The auto-advance decision succeeded, then
`auto-advance / implement / implement` started for `VOC-098-T01` even though T01
was declared operator-owned live evidence. The run had to be cancelled manually
before it could make changes.

Drafting-time read of `karsift-ai-infra/.github/workflows/auto-advance.yml`
confirms it sets `should_dispatch=true` for the next open roster task after
closed-issue / open-issue / existing-PR guards, without consulting
`<package>/.karsift/live-evidence/<task_id>.yaml` ownership from VOC-097.

## Required outcome (summary)

1. Detect the next task ownership/mode from governed package data before dispatch.
   A strict task-stanza `Automation ownership` marker signals when an operator
   contract is required; narrative prose is never parsed.
2. Do not dispatch the general implementer for operator-owned or live-evidence-only
   tasks.
3. Leave the operator task open and use a deterministic non-LLM clean publisher
   to create/reuse its draft evidence-carrier PR and sanitized waiting marker, so
   the existing PR-centric reconciler has a valid attachment point. Re-entry must
   repair a partial carrier/marker publication without duplicates.
4. Continue automatic dispatch for ordinary tasks with no live-evidence contract.
5. Preserve final-roster release behavior (skipping implementer must not open
   release early).
6. Fail closed on missing, malformed, or contradictory task ownership metadata.
7. Add deterministic positive, negative, malformed-metadata, carrier-idempotency,
   permission-boundary, and regression tests.
8. Prove the behavior through a controlled, sanitized workflow event without
   manufacturing live evidence or exposing secrets.
9. Keep this root focused; duplicate exact-SHA reviews, action-runtime upgrades,
   and cache-path warnings are separate follow-ups.

## Tasks

| Task | Summary                                                                       | Depends on |
| ---- | ----------------------------------------------------------------------------- | ---------- |
| T00  | Auto-advance ownership gate, fail-closed semantics, docs, deterministic tests | —          |
| T01  | Controlled sanitized workflow proof (operator-owned live evidence)            | T00        |

See `tasks.md` for full task definitions.

## What this package deliberately does NOT do

- Grant the implementer general GitHub Actions credentials.
- Change production application behavior, signup policy, secrets, databases, or
  Kuma/synthetic inventory IDs.
- Reopen VOC-098 code work or manufacture live evidence for unrelated packages.
- Address duplicate exact-SHA reviews, action-runtime upgrades, or cache-path
  warnings (explicit follow-ups outside this package).
- Weaken VOC-097 waiting/reconcile semantics or merge-gate fail-closed behavior.
- Self-adopt or self-authorize this package.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
