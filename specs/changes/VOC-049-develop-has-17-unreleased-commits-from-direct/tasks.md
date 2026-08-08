# VOC-049 — Tasks

## VOC-049-T00 — Re-verify the actual current commit gap between `main` and `develop`

- Requirement source: `specification.md`'s "Drafting-time finding" section
- Acceptance criteria: `VOC-049-AC-00`, `VOC-049-AC-03`
- Tests: `VOC-049-TEST-00`
- Evidence: `VOC-049-EV-00`
- Status: completed (2026-08-08; see `t00-evidence.md`)

Re-run the `main`/`develop` compare at implementation time
(`git log origin/main..origin/develop` or the GitHub compare view) and record
the exact resulting commit list, SHAs, and comparison timestamp. Treat this
result — not issue #375's original "17 commits" figure, and not this
package's own drafting-time finding of 1 commit (as of 2026-08-08) — as
authoritative for what `VOC-049-T01` (if needed at all) actually promotes. If
the re-verified gap is already zero, this task's finding is the package's
closing evidence and `VOC-049-T01` is marked not-needed rather than
dispatched. No production code change is made by this task under any outcome.

### `VOC-049-T00` implementation finding (2026-08-08T02:25:16Z)

The implementation-time `main`/`develop` comparison was re-verified and is
non-zero: `develop` is ahead of `main` by 23 commits and behind by 11 commits
(`status=diverged`), so `VOC-049-T01` remains required and must promote the
exact re-verified set recorded in `VOC-049-EV-00` (`t00-evidence.md`), with a
fresh re-run if `develop` moves again before promotion.

## VOC-049-T01 — Promote the re-verified `develop` content to `main` through an explicit, governed mechanism

- Requirement source: `specification.md`'s objective; open question 1;
  gated on `VOC-049-T00`'s non-zero finding
- Acceptance criteria: `VOC-049-AC-01`, `VOC-049-AC-02`, `VOC-049-AC-04`
- Tests: `VOC-049-TEST-01`, `VOC-049-TEST-02`, `VOC-049-TEST-03`
- Evidence: `VOC-049-EV-01`, `VOC-049-EV-02`
- Status: pending

Only dispatched if `VOC-049-T00` finds a non-zero gap. Promote exactly the
commit set `T00` recorded — no more, no less — from `develop` to `main`,
using the mechanism the reviewing human selected at adoption time from
`specification.md`'s open question 1. Record the promotion's evidence (PR or
equivalent, diff, merge commit SHA) and confirm no new application, workflow,
or governance-document content is introduced beyond what was already
reviewed and merged into `develop`. This task does not author any new
content of its own; it makes already-reviewed content reachable on `main`.
Depends on `VOC-049-T00`'s finding; if `develop` has moved further between
`T00`'s snapshot and this task's promotion, re-run `T00` against the new tip
before proceeding rather than silently expanding the promoted set.

Tasks preserve scope, separation of duties, and rollback safety. Neither task
may be dispatched before this package is adopted.
