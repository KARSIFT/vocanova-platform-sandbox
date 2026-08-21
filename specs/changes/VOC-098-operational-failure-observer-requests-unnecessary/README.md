# VOC-098 — Remove unnecessary Actions permission from operational-failure observer App token

| Field | Value |
|-------|-------|
| Package | `VOC-098` |
| Title | Remove unnecessary Actions permission from operational-failure observer App token |
| Path | `specs/changes/VOC-098-operational-failure-observer-requests-unnecessary` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#840](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/840) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The repository-managed operational-failure observer
(`.github/workflows/operational-failure-monitoring.yml`, VOC-088-T02) cannot process
a failed or cancelled watched workflow because its GitHub App installation-token
request includes Actions read permission that the installation does not grant.

Drafting-time read of the workflow shows the mint step requests both
`permission-issues: write` and `permission-actions: read`. VOC-094 added Actions
read so `classify-deploy-concurrency-cancel.sh` could call the Actions jobs API.
Token minting fails before `open-failure-issue.sh` runs, so App-authored sanitized
issues are never created or marker-deduplicated — including for
`scheduled-synthetics` conclusions that never need the deploy-cancel classifier.

## Required outcome (summary)

1. Remove the unnecessary Actions permission from the observer **App-token** request.
2. Preserve App-only issue creation, sanitized bodies, stable-marker deduplication,
   serialized concurrency, and fail-closed observer behavior.
3. Keep VOC-094 benign deploy-cancel classification working without putting Actions
   API reads on the App installation token (proposed: job `GITHUB_TOKEN` with
   workflow `actions: read` for the classifier only).
4. Add deterministic coverage proving the observer App token requests only the
   permissions it uses for issue write/dedupe.
5. Validate with a controlled failed/cancelled watched workflow and confirm exactly
   one sanitized issue is created or deduplicated; repeating the same marker creates
   no duplicate.
6. Do not copy workflow logs, secrets, sessions, OAuth data, user identifiers,
   cookies, or tokens into issues or evidence. Do not use `GITHUB_TOKEN` for issue
   creation.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Least-privilege App token + classifier token split and deterministic tests | — |
| T01 | Controlled live observer proof (operator-owned live evidence) | T00 |

See `tasks.md` for full task definitions.

## What this package deliberately does NOT do

- Grant the GitHub App installation broad Actions permissions.
- Use `GITHUB_TOKEN` to create or deduplicate operational-failure issues.
- Weaken sanitization, marker deduplication, concurrency serialization, or
  fail-closed classifier behavior.
- Change application code, signup policy, secrets, databases, or Kuma/synthetic
  inventory IDs.
- Self-adopt or self-authorize this package.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
