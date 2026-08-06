# VOC-040 — Tasks

The task below is not implementation-authorized by this package. Adoption and its
own implementation-authorization are separate, mirroring VOC-037/VOC-038/VOC-039's
convention. This package has exactly one task, deliberately: it does not number a
second `VOC-040-T01`-style entry for the two out-of-scope suggestions below,
because `adopt.yml`'s own task-issue-opening step scans this file for any
`VOC-040-T\d+` pattern and would open a real, unimplementable tracking issue for
it on adoption — see the "Explicitly out of scope" note below instead, which uses
no such pattern on purpose.

## VOC-040-T00 — Document the merge-without-adoption recovery procedure in AGENTS.md

- Requirement source: issue #301's first suggested improvement (documentation half)
- Acceptance criteria: `VOC-040-AC-00`
- Status: pending
- Summary: Add a new section to `AGENTS.md` (placement at the implementer's
  discretion — logically adjacent to the existing "Reporting a bug found outside
  the normal loop" section, since both concern process gaps in the governed
  automation loop) describing the `gh run rerun --failed`-on-the-original-run
  recovery procedure for when a `plan/`-branch PR merges without its `change.yaml`
  being edited to the adopted state first. Must include: how to identify the
  correct failed run, the required `change.yaml` edit before re-running, the exact
  `gh run rerun --failed <run-id>` invocation shape, the explicit dependency on the
  original run not being garbage-collected yet (with a fallback note pointing at
  VOC-039's actual fallback — a manual follow-up PR, per PR #300 — if it has been),
  and a citation of issue #301 and the VOC-039/PR #299/#300 incident as the origin.
  Must also state plainly that this is a documented workaround, not a fix for the
  underlying gap, and reference this package's own open question 1 (the
  `workflow_dispatch`/plan.yml-checklist alternatives, out of this repository's
  scope) so a future reader is not left thinking the gap is fully closed.

## Explicitly out of scope: `workflow_dispatch` entry point / plan.yml checklist

Requirement source: issue #301's second and third suggested improvements. Not
scoped as a task of this package, deliberately, and not given a `VOC-040-T`-numbered
identifier (see the note above). Both suggested improvements require editing files
(`adopt.yml`, `plan.yml`) that exist only in the separate `KARSIFT/karsift-ai-infra`
repository, which this package has no access to and no authority over — see
`specification.md`'s "Scope and non-goals" and open question 1. Recorded here only
so the task roster reflects issue #301's full request rather than silently omitting
two of its three suggestions. If a human reviewer wants either pursued, it needs a
separate change package filed against `KARSIFT/karsift-ai-infra` itself, not this
repository's task-dispatch mechanism (which can only ever open tracking issues and
dispatch `implement.yml` runs against files inside this repository).
