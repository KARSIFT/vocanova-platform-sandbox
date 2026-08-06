# VOC-040 — Test Plan

## VOC-040-TEST-00 — AGENTS.md documentation review against the required elements

- Covers: `VOC-040-AC-00`
- Preconditions: `VOC-040-T00`'s `AGENTS.md` edit implemented.
- Procedure:
  1. Read the new section in full and confirm all six elements
     `VOC-040-AC-00` requires are present: (a) how to identify the correct failed
     `adopt` run, (b) the required `change.yaml` edit before re-running, (c) the
     exact `gh run rerun --failed` invocation shape, (d) the explicit dependency
     on the original run not being garbage-collected, with the manual-follow-up-PR
     fallback named, (e) a citation of issue #301 and the VOC-039/PR #299/#300
     incident, and (f) an explicit statement that this is a workaround, not a
     structural fix, referencing the open `workflow_dispatch`/plan.yml-checklist
     alternatives as still-open.
  2. Cross-check the documented procedure's description of `adopt.yml`'s actual
     behavior (re-reading `change.yaml` from the target branch's current tip, not
     a frozen snapshot) against `adopt.yml`'s own current source at implementation
     time, in case it has changed since this package's drafting — a documentation
     task must describe the real current behavior, not this draft's snapshot of it,
     if the two have diverged by the time `VOC-040-T00` is actually implemented.
  3. Run `bash scripts/governance/validate-governance.sh`,
     `bash scripts/governance/classify-change-risk.sh`, and `git diff --check` per
     `AGENTS.md`'s own "Current validation" section; confirm none reports a new
     failure caused by this change.
- Expected result: all six elements present and accurate against `adopt.yml`'s
  actual current behavior at implementation time; all three deterministic checks
  pass (or report the same pre-existing state they reported before this change, if
  any pre-existing failure is unrelated to this diff).
- Evidence: `VOC-040-EV-00`

This package introduces no code, migration, or user-facing surface, so no
unit/integration/accessibility/migration-rollback test is applicable —
`VOC-040-TEST-00`'s documentation-accuracy review is this package's only
applicable test.
