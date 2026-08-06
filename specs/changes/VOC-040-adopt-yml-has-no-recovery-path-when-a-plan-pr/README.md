# VOC-040 — adopt.yml Has No Recovery Path When a Plan PR Merges Without Being Adopted First

**Status: proposed, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to [issue #301](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/301),
prepared for founder/steward review at adoption time.

## Why this exists

`adopt.yml` (a reusable workflow owned by `KARSIFT/karsift-ai-infra`, invoked from
this repository's own `.github/workflows/pipeline.yml`) correctly refuses to open
task issues when a `plan/`-branch PR is merged without its `change.yaml` first
being edited to the adopted state — that refusal is intentional and this package
does not change it. The gap issue #301 reports is that there is no *supported* way
to retry adoption after that happens, short of independently discovering that
`gh run rerun --failed` against the original failed run happens to work (because
`adopt.yml`'s verify step re-reads `change.yaml` from the target branch's current
tip, not a frozen snapshot). This happened live on VOC-039's adoption (PR #299
merged without the adoption edit; recovered via follow-up PR #300 plus a manual
rerun of the original failed job).

## Why this package is smaller than issue #301's full request

Issue #301 offers three alternative suggested improvements. Reading `adopt.yml`
and `plan.yml` at drafting time confirms both live entirely in the separate
`KARSIFT/karsift-ai-infra` repository, consumed here only via `uses:` references in
`pipeline.yml` — this package's own scope discipline forbids writing outside its
package directory in *this* repository (the untracked `karsift-ai-infra/` directory
present in this working tree is a local reference checkout, not part of this
repository's own tracked files). That leaves exactly one of the three suggestions
achievable here: documenting the existing recovery procedure in this repository's
own `AGENTS.md` (`VOC-040-T00`). The other two (a `workflow_dispatch` entry point
on `adopt.yml`; a merge-blocking checklist item in `plan.yml`'s draft PR) are
recorded in `tasks.md` as explicitly out of scope, with a pointer to
`specification.md`'s open question 1 for what filing a follow-up against
`KARSIFT/karsift-ai-infra` directly would require — not silently dropped, but not
something this package can implement.

## What this package deliberately does NOT do

- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. `VOC-040-T00` may not be dispatched until a real
  adoption decision is recorded.
- It does not change `adopt.yml`'s refusal behavior, and does not attempt to
  implement the two out-of-scope suggested improvements from a different
  repository it has no access to.
- It does not number a second task for those out-of-scope suggestions with a
  `VOC-040-T`-style identifier, because `adopt.yml`'s own task-issue-opening step
  scans `tasks.md` for exactly that pattern and would otherwise open a real,
  unimplementable tracking issue for it on adoption.

## Open question flagged for the reviewing human

`specification.md`'s "Open questions" section flags that closing issue #301 with
only `VOC-040-T00` resolves its immediate ask (a documented recovery path exists)
but not its deeper motivation (reducing how often the underlying mistake happens at
all) — that would need a separate change package filed against
`KARSIFT/karsift-ai-infra` itself.

## Structure

Mirrors recent packages' convention (e.g. VOC-037, VOC-038, VOC-039):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Read `specification.md`'s open questions and either accept this package's
   documentation-only scope or decide whether a separate `karsift-ai-infra`-side
   follow-up should also be requested now.
2. Confirm the proposed `R1` risk classification in `change.yaml` (a purely
   additive documentation change to a governance file).
3. Adopt (or request changes to) this package, then dispatch `VOC-040-T00`.
