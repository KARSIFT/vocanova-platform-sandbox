# VOC-114 — Impact Analysis

## Security and privacy

Recovery orchestration continues to mint short-lived GitHub App installation tokens.
This package adds **read** scopes required for exact-SHA metadata (Checks, Actions,
and existing Contents reads) while preserving the existing narrow mutation set
(contents/issues/PR write; Actions write only on dispatch paths already authorized).

Risk if implemented incorrectly: under-privileged reads leave recovery fail-closed
(blocking promotion); over-broad mutation grants would expand blast radius beyond
recovery needs. Mitigations: minimum read scopes, deterministic absent-permission
tests, and localized fail-closed diagnostics.

Evidence, diagnostics, and task comments remain metadata-only: SHAs, run/job IDs,
check names, conclusions, sanitized error classes. Forbidden: workflow logs, tokens,
OAuth or session material, cookies, user identifiers, email addresses.

## Data and migrations

None. No database schema, seed, or migration changes.

## Analytics and accessibility

None. No product analytics instrumentation or user-facing UI changes.

## Risks, dependencies, and evidence

- `VOC-114-R00`: **High operational risk** if recovery dispatches without
  successful metadata reads — could miss absent required checks or recurse
  incorrectly. Mitigation: fail closed before dispatch; endpoint-class errors;
  deterministic no-dispatch-after-read-failure tests.
- `VOC-114-R01`: **Medium operational risk** — App installation permission changes
  outside git may be required and are operator-owned. Mitigation: document exact
  grants in T00 evidence; T01 fails closed until grants exist.
- `VOC-114-R02`: **Medium release risk** — unblocking PR #947 after genuine checks
  promotes already-reviewed `develop` history and may trigger automatic production
  deploy. Acceptable as recovery of the stranded VOC-112 handoff; not a new deploy
  policy. Rollback remains revert/redeploy prior artifact.
- `VOC-114-R03`: **Low documentation risk** — stale README claims about App token
  scopes. Mitigation: update false claims in the same task PR.
- Protected surfaces: App-token merge/release/recovery paths, VOC-108 authoritative
  check selection, VOC-113 recovery invariants, branch ruleset required contexts.
- `VOC-114-DEP-00`: Issue #956 sanitized observations.
- `VOC-114-DEP-01`: VOC-113-T00 merged but broken at metadata read.
- `VOC-114-DEP-02`: adopt/merge-gate precedent for Checks/Actions read on gate metadata.
- `VOC-114-DEP-03`: Cross-repo karsift-ai-infra ownership pattern.
- `VOC-113`: Blocked T01 predecessor; must not be weakened.
- `VOC-108`: Authoritative exact-head selection predecessor.
- `VOC-114-EV-00`: T00 evidence — diagnosis, token contract, deterministic tests.
- `VOC-114-EV-01`: T01 evidence — live both-mode recovery and #947 unblock metadata.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow merge/adopt/release gates. EHR not triggered.

This draft proposes **R4**; the path classifier and independent verifier remain
authoritative.
