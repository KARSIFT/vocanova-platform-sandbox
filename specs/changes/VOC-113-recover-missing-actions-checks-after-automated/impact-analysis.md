# VOC-113 — Impact Analysis

## Security and privacy

Recovery orchestration may mint short-lived GitHub App installation tokens and
may receive a narrowly scoped `actions: write` grant solely to dispatch allowlisted
workflows (same separation pattern as live-evidence reconcile: mutation vs
dispatch authority). No new long-lived secrets, production credentials, or
personal-data processing.

Evidence, diagnostics, and task comments remain metadata-only: SHAs, run/job IDs,
check names, conclusions, timeouts. Forbidden: workflow logs, tokens, OAuth or
session material, cookies, user identifiers, email addresses.

Risk if implemented incorrectly: fabricating statuses would bypass the ruleset;
unbounded dispatch recursion could waste Actions capacity; using `github.token`
for merges could reintroduce silent event suppression. Mitigations are explicit
in `VOC-113-D01`–`D06` and deterministic negative tests.

## Data and migrations

None. No database schema, seed, or migration changes.

## Analytics and accessibility

None. No product analytics instrumentation or user-facing UI changes.

## Risks, dependencies, and evidence

- `VOC-113-R00`: **High operational risk** if recovery fabricates checks or merges
  without exact-head evidence — would violate the active ruleset. Mitigation:
  forbid status synthesis; reuse VOC-108 authoritative selection; fail closed.
- `VOC-113-R01`: **Medium operational risk** — recovery dispatches could recurse
  or duplicate promotion PRs/releases. Mitigation: idempotent converge keys,
  recursion guards, duplicate-PR refusal tests.
- `VOC-113-R02`: **Medium release risk** — completing PR #947 after genuine checks
  promotes already-reviewed `develop` history and may trigger automatic
  production deploy. Acceptable as recovery of the stranded VOC-112 handoff;
  not a new deploy policy. Rollback remains revert/redeploy prior artifact.
- `VOC-113-R03`: **Low documentation risk** — stale docs claiming close/reopen or
  draft/ready recover missing checks. Mitigation: update false claims in the
  same task PR.
- `VOC-113-R04`: **High governance availability risk** — the VOC-112 validator
  proves its squash but rejects every later PR because discarded capture commits
  are no longer ancestors. Mitigation: retain ancestry for the original capture
  PR and require exact source hashes anchored in the later PR merge base plus the
  reviewed head; add strict tamper negatives.
- Protected surfaces: App-token merge/release paths, ruleset required contexts,
  VOC-108 check selection, release converge concurrency.
- `VOC-113-DEP-00`: Issue #948 sanitized observations.
- `VOC-113-DEP-01`: Existing App-token merge/release posture.
- `VOC-113-DEP-02`: Cross-repo karsift-ai-infra ownership pattern.
- `VOC-113-DEP-03`: Active ruleset required checks.
- `VOC-113-DEP-04`: Open — precise root cause to document in T00.
- `VOC-108`: Authoritative exact-head check selection predecessor.
- `VOC-113-EV-00`: T00 evidence — diagnosis, mechanism, deterministic tests.
- `VOC-113-EV-01`: T01 evidence — PR #947 genuine checks + promotion outcome.
- `VOC-113-EV-02`: T02 evidence — post-promotion workflow metadata.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow merge/adopt/release gates. EHR not triggered.

This draft proposes **R4**; the path classifier and independent verifier remain
authoritative.
