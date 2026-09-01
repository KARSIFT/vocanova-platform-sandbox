# VOC-146 — Acceptance Criteria

## VOC-146-AC-00 — Nonexistent `--base` makes governance validation fail before claiming success

- Requirement source: `VOC-146-D01`, `VOC-146-D02`, `VOC-146-D03`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-00`
- Evidence: `VOC-146-EV-00`
- Result: pending

`validate-governance.sh --base <unresolved-commit> --head <resolvable-commit>`
returns nonzero and does not print `Governance structure validation passed.`
The class of issue #1127 (`376e00dd769afb0fe850052b3a5cb48f729e73ad` /
`79b2b3f1f4224235bdda3f77ee887c3004978deb`, Git fatal then exit 0) cannot
recur.

## VOC-146-AC-01 — Nonexistent `--head` makes governance validation fail before claiming success

- Requirement source: `VOC-146-D01`, `VOC-146-D02`, `VOC-146-D04`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-01`
- Evidence: `VOC-146-EV-00`
- Result: pending

`validate-governance.sh --base <resolvable-commit> --head <unresolved-commit>`
returns nonzero and does not print `Governance structure validation passed.`

## VOC-146-AC-02 — Unrelated/no-merge-base revisions fail closed

- Requirement source: `VOC-146-D02`, `VOC-146-D05`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-02`
- Evidence: `VOC-146-EV-00`
- Result: pending

Two resolvable commits with no merge base (`git diff A...B` fails) make
range loading return nonzero. Resolving both revisions is not treated as
success if the three-dot diff fails.

## VOC-146-AC-03 — A failed `git diff` status is preserved; empty output from a failed range is not accepted

- Requirement source: `VOC-146-D02`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-00`, `VOC-146-TEST-03`
- Evidence: `VOC-146-EV-00`
- Result: pending

Changed files for a requested `--base`/`--head` range are not loaded via
`mapfile < <(git diff …)`. A nonzero Git status becomes the script's
nonzero status. A successful diff with zero names remains a valid empty
change set.

## VOC-146-AC-04 — Partial `--base`/`--head` does not fall through to working-tree discovery

- Requirement source: `VOC-146-D06`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-04`
- Evidence: `VOC-146-EV-00`
- Result: pending

If either `--base` or `--head` is supplied without the other, the monitoring-impact
wrapper and the risk classifier fail closed. Working-tree fallback remains
only when neither `--files-from` nor any `--base`/`--head` argument was
requested.

## VOC-146-AC-05 — Valid ranges and `--files-from` keep working

- Requirement source: `VOC-146-D07`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-05`
- Evidence: `VOC-146-EV-00`
- Result: pending

A resolvable `--base`/`--head` pair still allows governance validation to
succeed when structure checks pass. `--files-from` still loads the supplied
list, including under `GITHUB_EVENT_NAME=pull_request`. VOC-086
`pull_request` missing-range fail-closed remains. `--declarations-only`
without a range remains usable.

## VOC-146-AC-06 — `classify-change-risk.sh` uses the same fail-closed range contract

- Requirement source: `VOC-146-D00`, `VOC-146-D08`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-06`
- Evidence: `VOC-146-EV-00`
- Result: pending

An unresolved or invalid `--base`/`--head` range makes
`classify-change-risk.sh` exit nonzero. It does not print `No changed files
to classify.` for that class. A valid empty range may still report no
changed files.

## VOC-146-AC-07 — Current-state docs match the live fail-closed range contract

- Requirement source: `VOC-146-D10`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-07`
- Evidence: `VOC-146-EV-00`
- Result: pending

An exhaustive tracked-source search identifies every current claim that a
pull_request changed-file range is fail-closed only when missing, or that
governance validation may succeed after Git reports an invalid symmetric
difference. `AGENTS.md` and every other current-state match state that an
unresolved commit or invalid diff range is fail-closed. Clearly marked
historical VOC-086 records remain historical. VOC-086 and VOC-112 package
directories are not rewritten.

## VOC-146-AC-08 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-146-D11`, `VOC-146-D12`, `VOC-146-D13`
- Tasks: `VOC-146-T00`
- Tests: `VOC-146-TEST-08`
- Evidence: `VOC-146-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation on the
implementation PR's valid range, R4 classification, VOC-146 and VOC-086
foundation tests, `git diff --check`, and independent exact-revision review
that binds the live head all pass. Evidence does not require a commit to
contain its own SHA. No infrastructure pin, VOC-112 recapture, or
develop/main gap snapshot is introduced.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
