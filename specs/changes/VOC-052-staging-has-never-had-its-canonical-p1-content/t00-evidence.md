# VOC-052-EV-00 — T00 staging deploy seed-step implementation evidence

Evidence for `VOC-052-T00` (`VOC-052-TEST-00`, `VOC-052-TEST-01`, `VOC-052-TEST-04`).

## VOC-052-DEP-01 resolution (required)

Resolved to **mechanism (a)** from `specification.md` open question 1:

- Build `apps/api/cmd/seed` on the GitHub Actions runner with a Linux static-binary
  posture (`CGO_ENABLED=0 GOOS=linux GOARCH=amd64`), then include that binary in the
  existing deploy bundle and run it on the staging host over SSH.
- This preserves the existing private-database boundary because the seed executes on
  the staging host and uses the same private Postgres bridge-IP URL rewrite already
  used by `migrate.sh`; no new public/staging-database ingress path is introduced.
- The alternative runner-direct-to-database mechanism (b) was not chosen because it
  would require exposing staging Postgres reachability beyond the current host-private
  model, which is outside this task's authority and not required to satisfy issue #437.

## Implemented workflow changes

File changed: `.github/workflows/deploy-staging.yml`

- The deploy bundle now includes `/opt/vocanova/apps/api/bin/p1-content-seed`, built
  from `./apps/api/cmd/seed` during the "Bundle deployable artifacts" step.
- The staging-host SSH deploy script now runs the canonical-content seed binary
  **after** `migrate.sh` and `seed-synthetic-smoke-user.sh`, and **before**
  `docker compose up -d`.
- The seed run reuses the same resolved private-IP `DATABASE_URL` value already used
  by migrations (`migration_database_url`), avoiding duplicated resolution logic.
- Header comments were updated so host-layout and cross-reference documentation match
  the new artifact and deploy sequence.

## Fail-closed behavior claim (VOC-052-AC-04)

- The new seed invocation runs inside the existing `set -euo pipefail` SSH script
  block.
- No `continue-on-error` was added.
- If `/opt/vocanova/apps/api/bin/p1-content-seed` exits non-zero, the SSH script exits
  before the `docker compose up -d` step, preserving the current fail-closed deploy
  shape (previously-running containers remain untouched).

## Local deterministic checks run for this task

Command set required by the package plan and repository policy:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Commands executed in this implementation run:

- `bash scripts/governance/validate-governance.sh` (pass)
- `bash scripts/governance/classify-change-risk.sh` (pass, detected floor `R3` via
  `.github/workflows/deploy-staging.yml`)

`git diff --check` is still required by plan policy, but was intentionally not run in
this implementer session because this run's explicit instructions prohibit running any
`git` command directly.

Implementation note: only workflow/package files were changed in this task. Live
staging execution evidence (idempotency across real deploy reruns and real `/discover`
behavior) is explicitly deferred to `VOC-052-T01` as the task list requires.
