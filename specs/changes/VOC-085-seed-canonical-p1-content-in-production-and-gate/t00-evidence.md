# VOC-085-EV-00 — T00 production deploy P1 seed-step implementation evidence

Evidence for `VOC-085-T00` (`VOC-085-TEST-00`, `VOC-085-TEST-01`, `VOC-085-TEST-02`,
`VOC-085-TEST-03`, `VOC-085-TEST-08`).

## VOC-085-DEP-01 resolution (reuse-only)

Resolved by mirroring `VOC-052-T00`'s staging mechanism without redesigning
`apps/api/cmd/seed` or `voc026-p1.json`:

- Build `apps/api/cmd/seed` on the GitHub Actions runner as a static
  `linux/amd64` binary (`CGO_ENABLED=0 GOOS=linux GOARCH=amd64`).
- Include `/opt/vocanova/production/apps/api/bin/p1-content-seed` in the
  production deploy bundle and run it on the production host over SSH.
- Preserve the private-database boundary: the seed executes on the production
  host using the same Postgres bridge-IP `DATABASE_URL` rewrite already used
  by `migrate.sh`.

## Implemented workflow changes

File changed: `.github/workflows/deploy-production.yml`

- Added SHA-pinned `actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e`
  with `go-version-file: apps/api/go.mod` immediately before the bundle step.
- The bundle step now creates `apps/api/bin/` and builds
  `p1-content-seed` with `go build -C apps/api -o .../p1-content-seed ./cmd/seed`.
- The production-host SSH deploy script now runs the canonical-content seed
  **after** `migrate.sh` and `seed-synthetic-smoke-user.sh`, and **before**
  `docker compose ... up -d --remove-orphans`.
- The seed run reuses the resolved private-IP `migration_database_url` value
  (unset only after the P1 seed completes).
- Staging's existing `deploy-staging.yml` P1 seed path is unchanged.

## Deterministic tests added

File added: `scripts/foundation/voc085-production-p1-seed.test.mjs`

- `VOC-085-TEST-00`: Go toolchain pin, bundle `apps/api/bin/p1-content-seed`
  build from `./cmd/seed`, staging regression guard.
- `VOC-085-TEST-01`: host-script order migrations → synthetic-user seed →
  P1 seed → `up -d`; `DATABASE_URL="$migration_database_url"` on seed line.
- `VOC-085-TEST-02`: `set -euo pipefail`, no `continue-on-error` on seed;
  disposable bash fixture proves failing seed aborts before convergence marker.
- `VOC-085-TEST-03`: references existing `apps/api/cmd/seed` upsert tests and
  absence of `DELETE FROM` in the seed tool.
- `VOC-085-TEST-08`: confirms `seed-synthetic-smoke-user.sql` still sets
  `onboarding_status='completed'` and `is_synthetic_test_account=true`.

## Fail-closed behavior (VOC-085-AC-01)

- P1 seed invocation runs inside the existing `set -euo pipefail` SSH script block.
- No `continue-on-error` was added.
- If `/opt/vocanova/production/apps/api/bin/p1-content-seed` exits non-zero, the
  SSH script exits before `docker compose up -d`, preserving the fail-closed
  deploy shape (previously-running containers remain untouched).

## Local deterministic checks run for this task

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
node --test scripts/foundation/voc085-production-p1-seed.test.mjs
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C apps/api -o /tmp/p1-content-seed ./cmd/seed
cd apps/api && go test ./cmd/seed/...
```

Record results from the implementation run in the task PR (pass/fail per command).

Implementation run (2026-08-15):

| Command | Result |
|---------|--------|
| `bash scripts/governance/validate-governance.sh` | pass |
| `bash scripts/governance/classify-change-risk.sh --files-from …` | pass (floor R3 via `deploy-production.yml`) |
| `git diff --check` | pass |
| `node --test scripts/foundation/voc085-production-p1-seed.test.mjs` | pass (5 tests) |
| `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C apps/api -o /tmp/p1-content-seed ./cmd/seed` | pass (static ELF 64-bit linux/amd64) |
| `cd apps/api && go test ./cmd/seed/...` | pass |

## Explicit deferrals

- Content-aware smoke body assertions (`VOC-085-T01`).
- Authenticated route sweep and live Cloudflare proof (`VOC-085-T02`).
- Live production idempotency across redeploys (evidence in `VOC-085-EV-02` after
  T02 promotion).
