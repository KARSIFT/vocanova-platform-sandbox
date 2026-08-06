# VOC-040 — Acceptance Criteria

## VOC-040-AC-00 — AGENTS.md documents the merge-without-adoption recovery procedure

- Requirement source: issue #301's first suggested improvement (documentation half)
- Tasks: `VOC-040-T00`
- Tests: `VOC-040-TEST-00`
- Evidence: `VOC-040-EV-00`
- Result: pending
- Observable outcome: `AGENTS.md` contains a section describing, in enough detail
  for an operator to follow without re-deriving it, how to recover when a
  `plan/`-branch PR merges without its `change.yaml` first being edited to
  `status: adopted` / `implementation.authorized: true`:
  1. Identify the original `adopt` workflow run that failed on that PR's merge
     (its "Verify the package was actually adopted" step fails with a message
     naming the unadopted `change.yaml` path).
  2. Edit `change.yaml` on the target branch to the adopted state (as the human
     should have done before merging).
  3. Re-run that specific failed run via `gh run rerun --failed` (not a fresh
     dispatch — `adopt.yml` has no `workflow_dispatch` trigger at drafting time)
     so its verify step re-reads the now-adopted `change.yaml` from the target
     branch's current tip and proceeds to open the real task issues.
  4. States explicitly that this depends on the original run not yet being
     garbage-collected by GitHub's retention window, and that if it has been,
     this procedure will not work and a different remediation (e.g. a manual
     follow-up PR, as VOC-039 used) is needed instead.
  5. Names issue #301 and the VOC-039/PR #299/#300 incident as the origin of this
     documented procedure, so a future reader has the concrete precedent, not just
     an abstract instruction.
  6. States, without overstating, that this is a documented workaround, not a
     structural fix — the underlying gap (no `workflow_dispatch` entry point on
     `adopt.yml`; no earlier guardrail in `plan.yml`'s draft PR) remains open and
     out of this repository's own control, per `specification.md`'s open question 1.
