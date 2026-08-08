# VOC-049 — Acceptance Criteria

## VOC-049-AC-00 — The actual current commit gap between `main` and `develop` is re-verified, not assumed from the issue's original count

- Requirement source: `specification.md`'s "Drafting-time finding" section
- Tasks: `VOC-049-T00`
- Tests: `VOC-049-TEST-00`
- Evidence: `VOC-049-EV-00`
- Result: satisfied (re-verified at 2026-08-08T02:25:16Z; see `t00-evidence.md`)

## VOC-049-AC-01 — If the re-verified gap is non-zero, every commit named in issue #375's batch (or its current equivalent) is promoted to `main` and no additional, unreviewed content rides along

- Requirement source: `specification.md`'s objective and scope
- Tasks: `VOC-049-T01`
- Tests: `VOC-049-TEST-01`
- Evidence: `VOC-049-EV-01`
- Result: pending

## VOC-049-AC-02 — The promotion mechanism used is explicit, governed, and recorded — not an ungoverned direct push to `main`

- Requirement source: `specification.md`'s open question 1;
  `implementation-plan.md`'s sequence
- Tasks: `VOC-049-T01`
- Tests: `VOC-049-TEST-02`
- Evidence: `VOC-049-EV-01`
- Result: pending

## VOC-049-AC-03 — If the re-verified gap is already zero, the package closes with that finding as evidence and opens no promotion PR

- Requirement source: `specification.md`'s "Drafting-time finding" section;
  open question 3
- Tasks: `VOC-049-T00`
- Tests: `VOC-049-TEST-00`
- Evidence: `VOC-049-EV-00`
- Result: not-triggered (gap is non-zero at T00 snapshot; see `t00-evidence.md`)

## VOC-049-AC-04 — Independent verification confirms the exact final promoted revision's SHA matches what `main` actually received, with no silent drift between reviewed and promoted content

- Requirement source: `CLAUDE.md`'s required review; `release-plan.md`
- Tasks: `VOC-049-T01`
- Tests: `VOC-049-TEST-03`
- Evidence: `VOC-049-EV-02`
- Result: pending

Acceptance criteria must be observable, stable, security-aware, and
bidirectionally traceable to requirements, tasks, tests, and evidence.
