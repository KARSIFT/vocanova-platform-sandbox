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

## Outcome

Repository CI has a dedicated controlled-signup callback E2E workflow on pull
requests and `develop`. The job requires a working Docker daemon, runs the real
T00 harness against disposable PostgreSQL, requires named PASS output for both
the allowlisted and unlisted cases, and rejects any controlled-signup test skip.

The workflow has read-only repository permission and no staging or production
secret, host, database, network, or deploy-user dependency.

## Deterministic controls

- The workflow uses the nested `apps/api` Go module and its exact dependency
  cache path.
- `set -euo pipefail` preserves the Go test exit through `tee`.
- Both named test cases must emit PASS; a SKIP causes an explicit failure.
- The foundation suite is included by the existing `pnpm test` foundation glob.
- A separate package script exposes the focused harness command without running
  the already-included API suite twice in the main test aggregate.
- Email-literal validation counts violations without printing an accidental
  address, and source-denylist checks use boolean results rather than echoing
  source text on failure.
- OAuth test logging is restricted to two fixed, scrubbed outcome messages; no
  formatted test logging is allowed.

## Validation

```bash
node --test scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs
bash scripts/governance/validate-governance.sh
git diff --check
```

The focused foundation suite passes locally. The exact final task SHA must pass
the dedicated Docker-backed GitHub Actions check, full repository CI,
governance, and independent review before merge. T03 owns the post-merge
`develop` run URL and live staging synthetic evidence.

No real identity or OAuth/session artifact is recorded here.
