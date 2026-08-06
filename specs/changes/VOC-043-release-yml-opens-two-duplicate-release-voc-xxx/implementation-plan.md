# VOC-043 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and its task are approved and implementation
is authorized, per this repository's `AGENTS.md` ("a chat prompt or issue
alone is not implementation authority").
`karsift-ai-infra/.github/workflows/release.yml` is a protected R3 area —
shared CI/CD/release-gating workflow infrastructure serving every package in
the fleet, not only this repository — see `specification.md`'s risk section.

## File reconciliation and implementation sequence

Existing target: `karsift-ai-infra/.github/workflows/release.yml` (read in
full at drafting time — 221 lines, two jobs: `check-completion` and
`promote`). Only `check-completion` (lines 41-148) is in scope; `promote`
(lines 150-221) is not touched. No conflicting in-flight work against this
file is known at drafting time.

Ordered steps (`VOC-043-T00`):

1. Split `check-completion` into two jobs, per `specification.md`'s open
   question 1 (or the implementer's confirmed alternative, recorded
   explicitly if different):
   - `identify`: runs on `issues: closed`, checkout, and reproduces exactly
     today's "Identify the package this issue belongs to, if any" step logic
     (lines 52-84) — outputs `is_task`, `change_id`, `package_path`. No
     `concurrency:` group needed on this job; it only reads.
   - `check-and-open`: `needs: identify`, `if:` guarded on
     `needs.identify.outputs.is_task == 'true'`, with
     `concurrency: { group: "release-check-${{ needs.identify.outputs.change_id }}", cancel-in-progress: false }`.
     Contains today's "Check whether every roster issue is now closed" step
     (lines 86-102) and "Open the release-approval issue, if complete and
     not already requested" step (lines 104-148), unchanged in their own
     internal logic (same `gh issue view`/`gh issue list`/`gh label
     create`/`gh issue create` calls, same title/body/label content).
2. Confirm (do not change) that `promote` (lines 150-221) is byte-for-byte
   identical before and after this diff — the split must not touch it.
3. Verify by exercising the concurrent-events scenario `VOC-043-TEST-00`
   describes (two `check-and-open` runs for the same `change_id` dispatched
   close together) and confirming exactly one release issue results, plus
   the single-event regression scenario `VOC-043-TEST-01` describes.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Plus this package's own `VOC-043-TEST-00`/`01`/`02` procedures. Because this
change is a GitHub Actions workflow file rather than an `apps/web`/`apps/api`/
`packages/` change, `pnpm validate` is not the primary check here — the
implementer should identify whichever mechanism `karsift-ai-infra` itself
uses (if any) to test its own reusable workflows (e.g. `act`, a workflow
dry-run, or a documented manual GitHub Actions test run against a disposable
repository/branch), and record which one was actually used, rather than
asserting untested YAML is correct.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this
specification, confirm `VOC-043-AC-00`/`01`/`02` are each satisfied with real
evidence (not asserted — in particular, evidence that the concurrent-events
race was actually reproduced and actually closed, not merely that the YAML
"looks" correct), and confirm no self-approval occurred. The reviewer should
also confirm `promote` was not touched, per `VOC-043-AC-02`.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. This
fix changes `karsift-ai-infra`'s reusable workflow; whichever mechanism this
repository (and any other repository this reusable workflow serves) uses to
consume it determines when it takes effect — see `specification.md`'s open
question 2. This package does not itself update any consuming repository's
pinned reference.

Rollout sequence (once authorized): merge to `develop` (or `karsift-ai-infra`'s
own equivalent integration branch, if that repository has a distinct release
process not covered by this package's drafting-time read). The fix takes
effect for future package completions once whatever consumes this reusable
workflow picks up the new revision.

Rollback trigger: the two-job split introduces an unexpected regression in
the common (non-concurrent) case (`VOC-043-AC-01`'s scenarios), or the chosen
concurrency mechanism turns out not to actually close the race in a real
concurrent-events test. Rollback mechanism: revert the job-split commit,
restoring the single `check-completion` job exactly as it exists today (the
same known, if racy, correct-in-the-common-case behavior). Owner: named
explicitly in the implementation PR at merge time, not left implicit.
