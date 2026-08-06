# VOC-043 — release.yml Opens Two Duplicate "Release: VOC-XXX" Issues for the Same Package

**Status: proposed, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to [issue #328](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/328),
prepared for founder/steward review at adoption time.

## Why this exists

Issue #328 reports that `karsift-ai-infra/.github/workflows/release.yml`'s
`check-completion` job opened two separate "Release: VOC-039" issues (#310 and
#311, identical bodies) for the same package. VOC-039's task roster
(issues #303-305) closed within seconds of each other — two of them
(`VOC-039-T00`/#303 and `VOC-039-T02`/#305) merged via `merge-gate.yml`'s
auto-merge path at essentially the same timestamp
(`mergedAt: 2026-08-06T09:27:12Z` and shortly after).

`check-completion` runs once per `issues: closed` event and, on the "all tasks
closed" branch, does exactly one existing-issue check
(`gh issue list --state open --search ...`) before creating a new "Release:
$change_id" issue. That check-then-create sequence is not atomic across
concurrent workflow runs: two near-simultaneous task-closing events for the
same package can each independently run their own job instance, each observe
zero existing "Release: VOC-039" issues (because neither run's `gh issue
create` has completed yet when the other run's `gh issue list` executes), and
each proceed to create its own release issue. This is a classic
check-then-act race, not a logic bug in the check itself.

## What this package deliberately does NOT do

- It does not touch `promote` (the second job in `release.yml`, which reacts to
  the founder's `approved` comment and opens/merges the promotion PR). That
  job already has its own idempotency handling for a related-but-different
  race (two concurrently-approved releases; see its "No commits between"
  handling) and is not implicated by issue #328's reproduction, which is
  scoped entirely to the issue-*opening* step.
- It does not touch `merge-gate.yml`'s auto-merge path, which is what produced
  the near-simultaneous task closures in the first place. Two tasks merging
  within seconds of each other is expected, intended behavior (that is what
  auto-merge is for) — the defect is `release.yml`'s response to that
  legitimate event, not the event itself.
- It does not change the human-facing approval mechanism (reply `approved` as
  the founder on the release issue) at all. A founder who happens to approve
  one of today's two duplicate issues still promotes correctly, per the
  issue's own "Impact" section — this package's entire scope is preventing the
  second issue from being created, not changing what approval does.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched until
  a real adoption decision is recorded.

## Open question flagged for the reviewing human

`specification.md`'s "Open questions" section flags that the two exact
concurrency-safe mechanisms named in the issue's own suggested fix (an
existing-issue re-check performed atomically with creation, vs. a
concurrency lock keyed on the package/change_id) have different tradeoffs
in GitHub Actions specifically (job-level `concurrency:` groups cannot
reference a step output computed inside the same job, only `needs.*.outputs`
from an upstream job or values computable directly from the event context) —
this package proposes a two-job restructure (an `identify` job whose output
feeds a `concurrency:` group on the job that performs the check-and-create)
as the most direct fix matching the issue's own suggestion, but flags this as
an implementation-level design choice for the implementer/reviewer to confirm
rather than a locked specification detail, since `release.yml` lives in
`karsift-ai-infra` (the reusable infra template, not this repository's own
application code) and this package's task should keep the diff minimal and
behavior-preserving otherwise.

## Structure

Mirrors recent packages' convention (e.g. VOC-042, VOC-041, VOC-040, VOC-039):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` risk classification in `change.yaml` (this
   touches shared CI/CD workflow infrastructure that gates production
   promotion for every package in the fleet, not just this repository).
2. Read `specification.md`'s open question and confirm (or redirect) the
   proposed two-job/`concurrency:`-group implementation approach.
3. Confirm whether the two live duplicate issues (#310, #311) from VOC-039
   need manual cleanup (closing whichever one is not approved) independent of
   this fix — this package does not touch already-open GitHub issues, only
   the workflow logic that would create future ones.
4. Adopt (or request changes to) this package, then dispatch `VOC-043-T00`
   individually, as prior packages' tasks were dispatched one at a time.
