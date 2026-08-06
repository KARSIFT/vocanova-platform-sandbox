# VOC-043 — Acceptance Criteria

## VOC-043-AC-00 — At most one open "Release: `<change_id>`" issue is ever created per package

- Requirement source: issue #328's confirmed root cause and suggested fix
- Tasks: `VOC-043-T00`
- Tests: `VOC-043-TEST-00`
- Evidence: `VOC-043-EV-00`
- Result: pending
- Observable outcome: given two (or more) `issues: closed` events for
  different tasks in the same package's roster, firing close enough together
  that both `check-completion` runs would, under the pre-fix logic, observe
  "all roster issues closed" before either had created a release issue, the
  post-fix workflow creates exactly one "Release: `<change_id>`" issue, not
  two. The second (and any subsequent) concurrent run either observes the
  first run's already-created issue and no-ops, or is serialized behind the
  first run by the fix's chosen mechanism and then observes it.

## VOC-043-AC-01 — Every other existing `check-completion` behavior is unchanged

- Requirement source: `specification.md`'s scope ("Preserving every other
  existing behavior... exactly")
- Tasks: `VOC-043-T00`
- Tests: `VOC-043-TEST-01`
- Evidence: `VOC-043-EV-01`
- Result: pending
- Observable outcome: for a single, non-concurrent `issues: closed` event
  (the common case, unaffected by the race), `check-completion` still: (a)
  no-ops immediately for a closed issue whose title doesn't match
  `<CHANGE-ID>: ...`; (b) no-ops for a matching title with no corresponding
  `specs/changes/<change_id>-*` directory; (c) no-ops for a matching package
  directory with no `.karsift/tasks.json`; (d) no-ops when the roster is not
  yet fully closed; (e) creates the `karsift:release` label (idempotently,
  via `--force`) and opens a "Release: `<change_id>`" issue with the same
  title, `karsift:release` label, and body content (roster summary plus the
  existing explanatory text) as before this fix, when the roster is fully
  closed and no such issue already exists.

## VOC-043-AC-02 — The `promote` job and founder-approval semantics are unchanged

- Requirement source: `specification.md`'s non-goals
- Tasks: `VOC-043-T00`
- Tests: `VOC-043-TEST-02`
- Evidence: `VOC-043-EV-02`
- Result: pending
- Observable outcome: the `promote` job's trigger condition, its founder-
  identity/comment-body check, and its promotion-PR open-or-reuse-then-merge
  logic (including its existing "No commits between" duplicate-release
  handling) are byte-for-byte unchanged by this package's diff.
