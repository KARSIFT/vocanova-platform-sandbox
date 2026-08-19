---
evidence_id: VOC-092-EV-00
task_id: VOC-092-T00
acceptance_criteria:
  - VOC-092-AC-00
  - VOC-092-AC-01
  - VOC-092-AC-02
  - VOC-092-AC-03
  - VOC-092-AC-04
  - VOC-092-AC-05
  - VOC-092-AC-06
  - VOC-092-AC-11
tests:
  - VOC-092-TEST-00
  - VOC-092-TEST-01
  - VOC-092-TEST-02
  - VOC-092-TEST-03
  - VOC-092-TEST-04
  - VOC-092-TEST-05
  - VOC-092-TEST-06
  - VOC-092-TEST-07
  - VOC-092-TEST-08
  - VOC-092-TEST-09
  - VOC-092-TEST-15
date: 2026-08-19
related_change: VOC-092
accountable_owner: unassigned
gate_status: repository-complete-exact-sha-review-pending
live_google_ui_claimed: false
---

# VOC-092-T00 — Ephemeral controlled-signup OAuth callback E2E

## Outcome

The repository contains a full application-owned OAuth callback integration
harness. It drives the real start and callback HTTP handlers, the real
`GoogleOAuthProvider` token and userinfo request boundary, and the PostgreSQL auth
repository. Google and PostgreSQL dependencies are disposable and bound to the
test host only.

The allowlisted synthetic case asserts onboarding redirect plus persisted user,
external-identity, and session rows. The unlisted synthetic case asserts the
stable HTTP 503 policy response and verifies that no auth rows were created.
Global signup remains disabled in both cases.

## Privacy and isolation

- Identity fixtures use only the reserved invalid synthetic domain.
- Provider configuration, authorization code, and access token are fixed
  test-only values; generated state and cookies remain ephemeral. None are
  written to test output or evidence.
- Session-cookie presence is reduced to a boolean before assertion so a failure
  cannot make Testify render the cookie value.
- The fake provider rejects an unexpected token form or missing bearer header
  using generic errors that do not echo request values.
- The harness ignores environment database URLs and creates its own loopback-only
  disposable PostgreSQL container.
- No application route, runtime provider switch, deploy configuration, public
  hostname, repository secret, or live database is used.
- This evidence does not claim to automate Google's interactive login UI.

## Repository validation

The predecessor revision passed repository CI at
https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32291802092/job/96194078245.
That job ran `cd apps/api && go test ./...`; the API package containing both
controlled-signup cases passed. The exact final task revision must pass the same
CI gate and a fresh independent review before merge.

Local execution on the operator workstation failed closed because its installed
Docker command is an unavailable environment shim. The tests did not skip or
connect elsewhere. GitHub Actions is the authoritative disposable-container
execution environment for this task.

## Plan-review reconciliation

This task reconciles the independent plan review before T00 merge:

- T03 ownership is recorded for AC-03, AC-04, and AC-08.
- T01 TEST-12 no longer depends on T02 operator documentation; TEST-16 owns the
  T02 documentation foundation test.
- The optional operations index is represented in `affected_areas`.
- The documented Go command runs from the nested `apps/api` module.

## Remaining gates

- exact-SHA CI after this evidence/redaction commit;
- independent exact-SHA review;
- T01 CI/foundation wiring and T03 live scheduled-synthetic evidence.
