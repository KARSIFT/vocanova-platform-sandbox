# VOC-043 — Test Plan

## VOC-043-TEST-00 — Concurrent roster-completing events produce exactly one release issue

- Covers: `VOC-043-AC-00`
- Preconditions: `VOC-043-T00`'s fix applied to `release.yml`; a test package
  (real or disposable/fixture) with a multi-task roster in
  `.karsift/tasks.json` and no existing open "Release: `<change_id>`" issue.
- Procedure:
  1. Close two (or more) of the test package's roster task issues close
     enough together that, under the pre-fix logic, both `check-completion`
     runs would observe "all roster issues closed" before either created a
     release issue — reproducing the timing VOC-039's #303/#305 actually hit
     (`mergedAt` within the same few seconds).
  2. Observe how many "Release: `<change_id>`" issues exist for the test
     package after both runs complete.
  3. Repeat with three simultaneous closes, if the chosen mechanism's
     robustness to more than two concurrent runs is not already obviously
     implied by the design (e.g. a real `concurrency:` queue is; an ad hoc
     re-check might not obviously be, and should be explicitly reasoned
     about in the implementation PR if a re-check approach is chosen instead
     of a `concurrency:` group).
- Expected result: exactly one "Release: `<change_id>`" issue exists,
  regardless of how many roster-completing events fired concurrently.
- Evidence: `VOC-043-EV-00`

## VOC-043-TEST-01 — Every other check-completion behavior is unchanged (single-event cases)

- Covers: `VOC-043-AC-01`
- Preconditions: `VOC-043-T00`'s fix applied.
- Procedure, each as an isolated single (non-concurrent) `issues: closed`
  event:
  1. Close an issue whose title doesn't match `<CHANGE-ID>: ...` — confirm
     `check-completion` (both jobs, if split) exits with no side effect.
  2. Close a matching-titled issue with no corresponding
     `specs/changes/<change_id>-*` directory on the integration branch —
     confirm no side effect.
  3. Close a matching-titled issue whose package directory has no
     `.karsift/tasks.json` — confirm no side effect.
  4. Close one (but not all) of a real roster's task issues — confirm no
     release issue is created while the roster remains incomplete.
  5. Close the last remaining open roster issue for a package with no
     existing "Release: `<change_id>`" issue — confirm exactly one is
     created, with the `karsift:release` label and the same title/body
     content (roster summary, "reply `approved`" instructions, "Deploy is
     out of scope here" note) as the pre-fix version produces for the same
     inputs.
- Expected result: identical observable behavior to the pre-fix workflow in
  every one of these five single-event cases.
- Evidence: `VOC-043-EV-01`

## VOC-043-TEST-02 — promote job is byte-for-byte unchanged

- Covers: `VOC-043-AC-02`
- Preconditions: `VOC-043-T00`'s fix applied.
- Procedure:
  1. Diff `release.yml`'s `promote` job (pre-fix lines 150-221) against the
     post-fix file's `promote` job.
- Expected result: zero-line diff for `promote`. All diff lines are confined
  to `check-completion` (or its post-split replacement jobs).
- Evidence: `VOC-043-EV-02`

This package introduces no migration, so no migration-rollback test is
applicable. No accessibility surface is affected, so no accessibility test is
applicable. No authorization boundary changes (the founder-identity check in
`promote` is untouched, per `VOC-043-TEST-02`), so no separate
authorization-failure test beyond that confirmation is applicable.
