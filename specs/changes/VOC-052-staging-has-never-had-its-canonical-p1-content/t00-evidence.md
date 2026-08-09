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
  from `apps/api`'s `./cmd/seed` during the "Bundle deployable artifacts" step.
- The build runs `go build -C apps/api ... ./cmd/seed`. The Go module root is
  `apps/api` (`apps/api/go.mod`); there is no root `go.mod` or `go.work`, so a build
  invoked from the workspace root without `-C` fails with a missing-module error. This
  matches the repository's existing convention in `package.json`'s `build:api` /
  `test:api` scripts and `apps/api/Dockerfile`'s in-module `WORKDIR`. The `-o` target
  stays an absolute path, so it is unaffected by the directory change.
- A new "Set up Go for the P1 content seed build" step
  (`actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e`, v7.0.0, SHA-pinned in
  the same style as the workflow's existing `actions/setup-node` pin) resolves the Go
  version from `apps/api/go.mod` (`go 1.26.0` / `toolchain go1.26.5`) rather than
  relying on whatever toolchain the `ubuntu-latest` image happens to ship, satisfying
  the implementation plan's "repository's pinned Go toolchain" requirement.
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
- `git diff --check` (pass, no whitespace errors; read-only inspection only — no
  staging, committing, or pushing was performed by this implementer session)
- `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C apps/api -o /tmp/p1-content-seed
  ./cmd/seed` from the repository root (pass; produced an 8.5MB statically linked
  ELF 64-bit linux/amd64 executable, confirmed with `file`). The previously-reviewed
  root-relative form (`go build -o ... ./apps/api/cmd/seed`) is the defect this
  remediation fixes.
- `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/deploy-staging.yml'))"`
  (pass; step order confirmed with the Go setup step immediately before "Bundle
  deployable artifacts")

## Remediation of attempt 1's independent-verification findings

Independent verification of commit `fc51693024209ce0619634fa8aa1d4f45697eafb` returned
`FAIL`. Both findings are addressed in this revision:

- **High — `go build` invoked outside the `apps/api` module root.** Fixed by adding
  `-C apps/api` and changing the package path to `./cmd/seed`; verified by actually
  running the corrected command from the repository root (see above), which previously
  would have failed the bundle step and prevented the seed from ever reaching staging.
- **Medium — Go toolchain not pinned in this workflow.** Fixed by adding the SHA-pinned
  `actions/setup-go` step reading `go-version-file: apps/api/go.mod`.
- **Low — `git diff --check` not executed.** Now executed (read-only) and recorded
  above as passing.

Implementation note: only workflow/package files were changed in this task. Live
staging execution evidence (idempotency across real deploy reruns and real `/discover`
behavior) is explicitly deferred to `VOC-052-T01` as the task list requires.
