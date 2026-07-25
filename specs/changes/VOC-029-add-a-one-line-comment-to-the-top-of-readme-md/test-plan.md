# VOC-029 — Test Plan

No test, fixture, or evidence may contain a real secret or production data;
none is relevant here since the change is a single documentation line.

## VOC-029-TEST-00 — README.md diff is exactly one added top-of-file comment line

- Covers: `VOC-029-AC-00`
- Preconditions: `VOC-029-T00` implemented (if ever adopted)
- Procedure: run `git diff README.md` and confirm the diff contains exactly one
  added line, positioned above the existing `# Vocanova` heading, containing a
  date; confirm no other line in `README.md` is changed; confirm
  `git status`/`git diff --name-only` show no file other than `README.md`
  changed; run `git diff --check` for whitespace/conflict-marker hygiene.
- Expected result: exactly one added line in `README.md`; no other file
  touched; `git diff --check` reports no issues.
- Evidence: `VOC-029-EV-00`.
