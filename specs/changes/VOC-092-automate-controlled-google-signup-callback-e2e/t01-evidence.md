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
live_ci_claimed: false
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
- local Docker execution failed closed because this WSL workspace has no usable
  Docker daemon.

No real email address, OAuth code, state value, access token, refresh token, or
session cookie value was used or recorded.

The exact final revision must pass the dedicated Docker-backed GitHub Actions
check and full repository CI. A prior candidate passed that dedicated workflow,
but this evidence does not claim the prior result for the current revision.
Post-merge `develop` run URL and live staging synthetic proof remain T03 scope
(`VOC-092-EV-03`).

## Fail-closed confirmation

| Control | Mechanism | Rehearsal result |
| --- | --- | --- |
| Harness test failure | `set -euo pipefail` + `go test` non-zero exit | Workflow step fails on test failure (no `continue-on-error`, no `\|\| true`; foundation TEST-11 asserts both) |
| Missing named PASS | Post-test `grep` for `TestControlledSignupOAuth_AllowlistedCallbackSucceeds` and `_UnlistedCallbackDenied` PASS lines | Simulated log without PASS line → grep exit 1 |
| Silent SKIP | Post-test rejection when log contains `--- SKIP: TestControlledSignupOAuth_` | Simulated SKIP line → explicit `exit 1` with message *controlled-signup OAuth callback E2E cases must not skip* |
| Missing Docker | Pre-test `docker version` / `docker info` | Local WSL check failed closed; exact-revision CI remains pending |
| Staging/production coupling | Local runner host-argument denylist | Script exits 1 on `*vocanova.site*` / production host patterns |

Operational-failure monitoring eligibility is conditional per AC-08: this workflow
is not in the `operational-failure-monitoring.yml` observed list; merge-gate
fail-closed behavior is the primary CI gate.

## Remaining gates

- exact-SHA CI on the final remediation revision (including this evidence commit);
- independent exact-SHA review;
- T02 documentation/VOC-088 evidence remediation;
- T03 post-merge `develop` CI URL and live staging synthetic proof.
