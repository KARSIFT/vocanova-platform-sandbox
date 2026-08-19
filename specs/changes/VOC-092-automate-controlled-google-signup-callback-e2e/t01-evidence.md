---
evidence_id: VOC-092-EV-01
task_id: VOC-092-T01
acceptance_criteria:
  - VOC-092-AC-05
  - VOC-092-AC-06
  - VOC-092-AC-08
  - VOC-092-AC-11
tests:
  - VOC-092-TEST-07
  - VOC-092-TEST-08
  - VOC-092-TEST-09
  - VOC-092-TEST-11
  - VOC-092-TEST-12
  - VOC-092-TEST-15
date: 2026-08-19
related_change: VOC-092
accountable_owner: unassigned
gate_status: repository-complete-exact-sha-ci-pending
live_ci_claimed: true
remediation_of: 83d2a6b159ccfdb9483fc56bb08bcf3daaea9b8e
---

# VOC-092-T01 — Controlled-signup OAuth callback E2E CI wiring

## Scope and outcome

Repository CI now runs the T00 controlled-signup OAuth callback E2E harness on
every pull request targeting `develop`/`main` and on every push to `develop`.
The job requires a working Docker daemon, executes the real HTTP handlers and
`GoogleOAuthProvider` boundary against disposable loopback PostgreSQL only, and
fails closed when either named case does not PASS or any controlled-signup case
SKIPs.

No staging or production host, secret, database, network, deploy user, or public
test-auth route is introduced.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Dedicated Docker-backed CI workflow | `.github/workflows/controlled-signup-oauth-e2e.yml` |
| Harness/CI foundation tests | `scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs` |
| Focused package script | `package.json` → `test:controlled-signup-oauth-e2e` |
| T00 harness (dependency) | `apps/api/app/api/controlled_signup_oauth_e2e_test.go` |
| Disposable Postgres helper (dependency) | `apps/api/app/api/controlled_signup_oauth_postgres_test.go` |
| Local operator wrapper (dependency) | `infra/scripts/run-controlled-signup-oauth-e2e.sh` |
| This evidence | `specs/changes/VOC-092-automate-controlled-google-signup-callback-e2e/t01-evidence.md` |

## Remediation notes (attempt 2)

Independent verification of `83d2a6b159ccfdb9483fc56bb08bcf3daaea9b8e` failed on:

1. **Medium (M1)** — Missing `t01-evidence.md` (`VOC-092-EV-01`). This file
   closes that traceability gap with repository validation, a green dedicated
   workflow run URL, and fail-closed rehearsal results.

Technical CI wiring from attempt 1 was retained; no harness or workflow behavior
change was required beyond the evidence artifact.

## Acceptance mapping

| Acceptance criterion | Repository result |
| --- | --- |
| AC-05 | Foundation tests lock `@synthetic.vocanova.invalid` fixtures and deny OAuth-secret log patterns in harness sources. |
| AC-06 | Denylist covers forbidden test-auth/bypass strings; workflow and harness introduce no runtime fake-provider switch. |
| AC-08 | Workflow triggers on PR (`develop`/`main`) and `push` to `develop`; foundation tests assert wiring; `pnpm test` aggregates the foundation suite via the existing glob. |
| AC-11 | Host/secret-env denylist and explicit Docker absence handling are asserted; local runner refuses staging/production host arguments. |

## Deterministic validation

Commands (repository root):

```bash
node --test scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs
bash scripts/governance/validate-governance.sh
git diff --check
cd apps/api && go test ./app/api/... -run ControlledSignupOAuth -count=1 -v
```

Results on the reviewed working tree:

- six VOC-092 foundation tests passed;
- governance structure validation passed;
- local Docker-backed harness run passed both named cases (scrubbed excerpt below).

No real email address, OAuth code, state value, access token, refresh token, or
session cookie value was used or recorded.

### Scrubbed local harness excerpt

Both cases emit only the two fixed outcome log lines allowed by the foundation
redaction test:

```
--- PASS: TestControlledSignupOAuth_AllowlistedCallbackSucceeds
    controlled-signup OAuth allowlisted callback succeeded with redirect to onboarding and persisted auth rows
--- PASS: TestControlledSignupOAuth_UnlistedCallbackDenied
    controlled-signup OAuth unlisted callback denied with HTTP 503 and no persisted user
```

## Green CI run (dedicated workflow)

| Field | Value |
| --- | --- |
| Workflow | `controlled-signup-oauth-e2e.yml` |
| Run number | **#5** |
| Run ID | `32295470850` |
| Job URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32295470850/job/96205647779 |
| Head SHA | `ac462d15d2db77637d1e4d4f5c79a1cf82131610` |
| Pull request | #790 |
| Conclusion | **success** |

The job required Docker, ran
`go test ./app/api/... -run ControlledSignupOAuth -count=1 -v`, and enforced
named PASS output for both allowlisted and unlisted cases with no SKIP. Earlier
attempt-1 revisions on the same branch (`4a41bca7`, `c160dc96`, `e6bc028a`,
`83d2a6b1`) also passed runs #1–#4 of this workflow.

The exact final remediation commit must re-pass this check and full repository
CI before merge. Post-merge `develop` run URL and live staging synthetic proof
remain T03 scope (`VOC-092-EV-03`).

## Fail-closed confirmation

| Control | Mechanism | Rehearsal result |
| --- | --- | --- |
| Harness test failure | `set -euo pipefail` + `go test` non-zero exit | Workflow step fails on test failure (no `continue-on-error`, no `\|\| true`; foundation TEST-11 asserts both) |
| Missing named PASS | Post-test `grep` for `TestControlledSignupOAuth_AllowlistedCallbackSucceeds` and `_UnlistedCallbackDenied` PASS lines | Simulated log without PASS line → grep exit 1 |
| Silent SKIP | Post-test rejection when log contains `--- SKIP: TestControlledSignupOAuth_` | Simulated SKIP line → explicit `exit 1` with message *controlled-signup OAuth callback E2E cases must not skip* |
| Missing Docker | Pre-test `docker version` / `docker info` | CI runner provides Docker; harness sources require explicit *docker not on PATH* locally instead of connecting elsewhere |
| Staging/production coupling | Local runner host-argument denylist | Script exits 1 on `*vocanova.site*` / production host patterns |

Operational-failure monitoring eligibility is conditional per AC-08: this workflow
is not in the `operational-failure-monitoring.yml` observed list; merge-gate
fail-closed behavior is the primary CI gate.

## Remaining gates

- exact-SHA CI on the final remediation revision (including this evidence commit);
- independent exact-SHA review;
- T02 documentation/VOC-088 evidence remediation;
- T03 post-merge `develop` CI URL and live staging synthetic proof.
