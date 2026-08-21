# VOC-098 — Acceptance Criteria

## VOC-098-AC-00 — App installation token mints with issues-write only

- Requirement source: `VOC-098-D00`, `VOC-098-D03`
- Tasks: `VOC-098-T00`
- Tests: `VOC-098-TEST-00`, `VOC-098-TEST-01`
- Evidence: `VOC-098-EV-00`
- Result: pending

`.github/workflows/operational-failure-monitoring.yml` mints the App installation
token with `permission-issues: write` and without `permission-actions` (or any
other unused permission). Deterministic tests assert this permission set.

## VOC-098-AC-01 — Issue creation remains App-only; classifier does not use App Actions

- Requirement source: `VOC-098-D01`, `VOC-098-D02`, `VOC-098-D05`
- Tasks: `VOC-098-T00`
- Tests: `VOC-098-TEST-02`, `VOC-098-TEST-03`, regression `VOC-088-TEST-11`,
  `VOC-094-TEST-05`
- Evidence: `VOC-098-EV-00`
- Result: pending

`open-failure-issue.sh` is invoked only with the App token. The benign-cancel
classifier may use the job `GITHUB_TOKEN` (with `actions: read` at the workflow/job
permissions floor) for bounded jobs-API metadata. No step uses `GITHUB_TOKEN` for
issue create or open-issue marker scans. Missing App credentials still fail closed.

## VOC-098-AC-02 — Sanitization, markers, concurrency, and fail-closed behavior preserved

- Requirement source: `VOC-098-D02`, `VOC-098-D05`
- Tasks: `VOC-098-T00`
- Tests: `VOC-098-TEST-04`, regression VOC-088 / VOC-094 foundation suites
- Evidence: `VOC-098-EV-00`
- Result: pending

Observed workflow names/conclusions, unlabeled issues, HTML marker format,
concurrency serialization, and classifier fail-closed skip/create semantics remain
intact. Issue bodies still forbid logs, secrets, sessions, OAuth data, cookies,
tokens, and user identifiers.

## VOC-098-AC-03 — Watched non-success invokes observer successfully after fix is live

- Requirement source: issue #840 acceptance; `VOC-098-D04`
- Tasks: `VOC-098-T01`
- Tests: `VOC-098-TEST-05`
- Evidence: `VOC-098-EV-01`
- Result: pending

After T00 is live on the branch the observer executes from, a controlled watched
workflow non-success (`failure`, `cancelled`, or `timed_out`) causes
`operational-failure-monitoring` to complete with conclusion `success` (App token
mint no longer fails).

## VOC-098-AC-04 — Exactly one sanitized App-authored issue created or deduplicated

- Requirement source: issue #840 acceptance; `VOC-098-D02`, `VOC-098-D04`
- Tasks: `VOC-098-T01`
- Tests: `VOC-098-TEST-06`
- Evidence: `VOC-098-EV-01`
- Result: pending

The resulting issue (or the existing open owner of the same marker) is App-authored,
unlabeled, sanitized to allowlisted fields only, and contains the stable marker
`<!-- operational-failure:{workflow}:{conclusion} -->`. Repeating the same marker
creates no duplicate open issue.

## VOC-098-AC-05 — Operator documentation matches dual-token contract when touched

- Requirement source: AGENTS.md doc-consistency rule; `VOC-098-D01`
- Tasks: `VOC-098-T00`
- Tests: `VOC-098-TEST-07`
- Evidence: `VOC-098-EV-00`
- Result: pending

If `docs/operations/staging-controlled-signup.md` (or any other in-diff doc) describes
observer token minting, it accurately states App token = issue write/dedupe and that
Actions jobs metadata for benign-cancel classification uses the job token — or the
doc is left unchanged only when it already does not claim App Actions permission.
