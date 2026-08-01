# VOC-037-EV-06 — Production provisioning evidence (T06)

## Standing of `VOC-037-AC-06` at this revision

**`VOC-037-AC-06` is NOT satisfied (corrected 2026-08-01, later the same
day).** An earlier revision of this file claimed SATISFIED based on a
negative-access rehearsal run manually as root, which does not reflect
what the real automated deploy observes. The founder provisioned real
production infrastructure directly (GitHub `production` environment with
required reviewers, all 6 `PRODUCTION_*` secrets, DNS, and SSH access), and
the application itself is genuinely live and healthy - but the rehearsal's
`INS-11` check, once run correctly (as the non-root `vocanova-production`
deploy identity, matching the real workflow), surfaces a confirmed,
disclosed isolation gap: staging's `deploy` user has independent blanket
sudo and can read production's secrets regardless of directory
permissions. See "Confirmed residual risk: `deploy`'s blanket sudo" below.
Closing AC-06 requires either narrowing `deploy`'s sudoers or an explicit
founder-accepted waiver of this specific finding - this document does
neither on its own.

**Correction (2026-08-01, later the same day):** the row below claiming
`INS-9`-`INS-11` "Met... `PASS` on every check" was obtained by running the
rehearsal script manually as root (`sudo bash rehearse-...`), which
trivially bypasses the exact permission boundary INS-11 is supposed to
test - root can always `sudo -u` anyone. The real automated
`deploy-production` workflow runs the script as `vocanova-production`
(correctly not root), and its real run **failed** INS-11: the production
deploy user has no safe way to impersonate staging's `deploy` user for a
live read probe, and investigating why surfaced a real, more significant
finding - `deploy` already has independent, pre-existing blanket
`(ALL:ALL)` sudo on the shared host, which means directory-based isolation
does not actually hold against `deploy` regardless of file permissions.
The rehearsal script was corrected to detect and FAIL on exactly this
condition (not silently pass) - see the updated `INS-11` output further
below, which now correctly reports **FAIL**, not the stale `PASS` originally
recorded in this section before the correction.

| AC-06 clause | Status |
| --- | --- |
| `/opt/vocanova/production/` exists, fully separate from `/opt/vocanova/infra/` | **Met.** Verified live: `mode 750`, owned by `vocanova-production`. |
| `vocanova-production` Compose project with explicit per-service resource limits | **Met.** `infra/docker-compose.production.yml`; all 4 containers (`postgres`/`api`/`web`/`nginx`) running healthy under it. |
| `production` GitHub Actions environment with founder-controlled required reviewers | **Met.** Created 2026-08-01 with `m-e-h-r-d-a-a-d` as required reviewer; both real deploy runs required and received that approval before executing. |
| `deploy-production.yml` deploys without touching staging's tree, user, or Compose project | **Met.** Verified statically, by rehearsal, and by a real run against the real shared host. |
| Negative-access rehearsal proves staging cannot read production secrets (`INS-9`–`INS-11`) | **NOT met.** `INS-9`/`INS-10` pass; `INS-11` correctly FAILS - `deploy`'s independent blanket sudo means directory-based isolation cannot be proven against it. See "Confirmed residual risk" below. |

## Repository deliverables (implemented and verified here)

| Deliverable | Verification performed |
| --- | --- |
| `.github/workflows/deploy-production.yml` | YAML parses; gated on `environment: production`; consumes only `PRODUCTION_*` secrets; writes only under `/opt/vocanova/production/`; operates only on the `vocanova-production` Compose project; deploys the immutable `sha-<short-sha>` image tag so DOC-11 §3's redeploy-by-digest rollback has a specific artifact to name; fails before starting anything if the compose file references a path outside the production root. |
| `infra/docker-compose.production.yml` | `docker compose -f infra/docker-compose.production.yml config` exits 0; project name `vocanova-production`; production-only network and volume names; per-service `mem_limit`/`cpus`; no `build:` block (see below); every path anchored at `${VOCANOVA_PRODUCTION_ROOT}`. |
| `infra/docker-compose.yml` (staging) | `docker compose config` exits 0; per-service `mem_limit`/`cpus` added, budgeted against production's on the shared 2 vCPU / 4 GB host. |
| `.github/workflows/deploy-staging.yml` | Ownership fix narrowed to staging's own subtrees; bundle contents verified against an `infra`/`apps` allowlist before extraction. |
| `infra/nginx-production/` | Production-only config tree with founder-confirmable hostname placeholders. |
| `infra/scripts/rehearse-production-secrets-boundary.sh` | Executed; see `VOC-037-EV-01`. |
| `infra/scripts/rehearse-production-secrets-boundary.selftest.sh` | Executed; eight cases, `SELFTEST PASS`. |
| `infra/README.md` | Production path layout, resource budget, Cloudflare origin-port routing, and isolation conventions documented. |

## Isolation defects found and fixed during this task

1. **Staging's deploy re-owned the production tree.** `deploy-staging.yml`
   ran `sudo chown -R "$(id -un)":"$(id -gn)" /opt/vocanova`. With
   `VOC-037-D00` co-locating both tiers under that root, the next staging
   deploy after any production provision would have handed
   `/opt/vocanova/production/secrets/` to the staging deploy user, defeating
   `VOC-037-D01`'s `INV-4`. The `chown` is now scoped to
   `/opt/vocanova/infra /opt/vocanova/apps`, and the deploy bundle is
   rejected before extraction if it contains any path outside those two
   subtrees. Both the fixed and the pre-fix ownership commands are
   exercised as rehearsal cases, so the regression cannot silently return.
2. **Production compose built from staging's directory tree.** The compose
   file is deployed to `/opt/vocanova/production/`, so its relative build
   contexts (`../apps/api`, `..`) resolved to `/opt/vocanova/apps/api` and
   `/opt/vocanova/`. Production now runs pulled images only, with no
   `build:` block, and the rehearsal fails if one reappears.
3. **The cross-tier compose check was blind to `env_file`.** `docker compose
   config` folds `env_file` entries into `environment:` and drops their
   paths, so a production compose file reading staging's `api.env` rendered
   completely clean. Both the deploy-time guard and the rehearsal script now
   scan the raw compose source in addition to the rendered output.
4. **Secret files were briefly world-readable during a deploy.** The
   AI-provider sync step created `/opt/vocanova/production/secrets/` with
   the default `0755` and only tightened it later. Modes are now set before
   the first secret byte is written, and the file baseline is re-asserted on
   every deploy.
5. **The rehearsal script could pass without checking anything.** It printed
   directory modes without asserting them, ignored `*.env` modes entirely,
   and treated a missing file or a refused `sudo -u` as success. Every check
   now asserts, and an unevaluated check is a failure.

## Shared-host resource budget

`VOC-037-D00` accepted co-location on one 2 vCPU / 4 GB host and required
the contention risk to be addressed rather than assumed away. Memory limits
across both compose files:

| Service | Production | Staging |
| --- | --- | --- |
| postgres | 768m | 512m |
| api | 512m | 384m |
| web | 512m | 384m |
| nginx | 192m | 128m |
| **total** | **~1.9 GB** | **~1.4 GB** |

~3.3 GB of 4 GB committed, leaving headroom for the host itself. CPU limits
are per-service ceilings, so their sum may exceed 2 vCPU by design. Raising
a limit in one file without lowering the other oversubscribes the host; both
files carry that note.

## Founder-provisioned infrastructure (2026-08-01)

- GitHub `production` environment, required reviewer `m-e-h-r-d-a-a-d`.
- `PRODUCTION_SSH_HOST`/`PRODUCTION_SSH_USER`/`PRODUCTION_SSH_PRIVATE_KEY`/
  `PRODUCTION_SSH_KNOWN_HOSTS`/`PRODUCTION_CLOUDFLARE_API_TOKEN`/
  `PRODUCTION_CLOUDFLARE_ACCOUNT_ID` — all 6 set.
- Dedicated OS user `vocanova-production` on the shared host: `docker` group
  membership, narrowly-scoped passwordless sudo
  (`mkdir`/`tar`/`chown`/`touch`/`curl`/`chmod` only, not blanket `ALL`).
- DNS: `production.vocanova.site` / `api-production.vocanova.site`,
  Cloudflare-proxied A records to the shared host.
- Real Cloudflare Origin CA certificate (`*.vocanova.site`/`vocanova.site`,
  15-year validity) installed at
  `/opt/vocanova/production/secrets/nginx/{cert,key}.pem`.

## Live deploy defects found and fixed (2026-08-01)

The real `deploy-production` run did not pass on the first attempt. Each
defect below was found live, fixed, and re-verified - none was anticipated
in advance:

1. **Missing `POSTGRES_PASSWORD`.** First run: Postgres refused to
   initialize with no superuser password (`postgres.env` had never been
   populated on a brand-new host). Fixed by generating and setting a real
   password in `/opt/vocanova/production/secrets/postgres.env` and the
   matching `DATABASE_URL` in `api.env`.
2. **`api.env` ownership mismatch.** The manual fix in (1) created
   `api.env`/`postgres.env` owned by the SSH login user, not
   `vocanova-production`; the AI-provider sync step's non-sudo `touch`
   (correct behavior, since the step chowns the directory to itself first)
   then failed with `Permission denied` against files it didn't own.
   Fixed by re-chowning the secrets tree.
3. **`BASE_URL is required` crash loop.** The api container requires
   `BASE_URL`, `OAUTH_REDIRECT_URI`, and `SESSION_COOKIE_DOMAIN`
   (`apps/api/app/api/production.go`'s `LoadProductionConfig`); nothing had
   ever written them for production. Fixed by adding a new "Write
   production application configuration" step to `deploy-production.yml`
   that derives these (non-secret, workflow-input-derived) values
   automatically on every deploy — `EMAIL_MAGIC_LINK_ENABLED`/
   `GOOGLE_OAUTH_ENABLED`/`NEW_USER_SIGNUP_ENABLED` all default `false`
   until real production-tier email/OAuth credentials exist (this is
   infrastructure verification, not a launch decision - `T05` remains the
   actual go/no-go).
4. **`docker restart` doesn't re-read `env_file`.** Fixing (3) on disk had
   no effect until containers were recreated (`docker compose up -d
   --force-recreate`), not merely restarted — Compose bakes `env_file`
   content into the container at creation time.
5. **Bind-mounted TLS paths auto-created as directories.** Docker creates a
   directory at a bind-mount source path that doesn't exist yet, so
   `cert.pem`/`key.pem` became directories before any cert existed,
   breaking `openssl req`'s output path. Fixed by removing the directories
   before writing real files.
6. **Port collision with staging + Cloudflare edge routing.** Documented in
   `t00-production-hosting-decision-record.md`'s second supersession note:
   production's nginx runs on 8081/8443 (staging owns 80/443 on the same
   host); Cloudflare proxies 8443 automatically with no dashboard change,
   but every client must include the port. `deploy-production.yml`'s health
   checks and `NEXT_PUBLIC_API_BASE_URL` default were fixed to include
   `:8443`.
7. **Self-signed cert rejected by Cloudflare's strict SSL mode (`526`).**
   Fixed by installing a real Cloudflare Origin CA certificate (see above).

## Real verification (2026-08-01)

```
$ curl -sS -o /dev/null -w "web:8443 -> %{http_code}\n" https://production.vocanova.site:8443/
web:8443 -> 200
$ curl -sS https://api-production.vocanova.site:8443/healthz
{"$schema":"https://api-production.vocanova.site/schemas/HealthzOutputBody.json","status":"ok","database":"ok","timestamp":"2026-08-01T19:42:49Z"}
```

All 13 migrations applied cleanly against a fresh production database (the
Atlas `atlas:txmode`/duplicate-index defects that blocked R1's first
attempts were already fixed upstream by `VOC-033` - confirmed live here,
correcting this file's earlier note below that they were still open).

`infra/scripts/rehearse-production-secrets-boundary.sh`, executed on the
real host after the real deploy:

```
[INS-9] production secret tree exists and matches the D01 permission baseline
  ok: /opt/vocanova/production mode 750 is within 750
  ok: /opt/vocanova/production/secrets mode 700 is within 700
  ok: /opt/vocanova/production/secrets/nginx mode 700 is within 700
  ok: /opt/vocanova/production is owned by vocanova-production (not deploy)
  ok: /opt/vocanova/production/secrets is owned by vocanova-production (not deploy)
  ok: /opt/vocanova/production/secrets/api.env mode 600 is within 600
  ok: /opt/vocanova/production/secrets/api.env is owned by vocanova-production (not deploy)
  ok: /opt/vocanova/production/secrets/postgres.env mode 600 is within 600
  ok: /opt/vocanova/production/secrets/postgres.env is owned by vocanova-production (not deploy)
  ok: 2 production env file(s) checked
  ok: /opt/vocanova/production/secrets/nginx/key.pem mode 600 is within 600
  ok: /opt/vocanova/production/secrets/nginx/cert.pem mode 600 is within 600
[INS-10] production compose reads the production tree only
  ok: no rendered compose path outside /opt/vocanova/production
  ok: production compose declares no build context
  ok: no compose source path outside /opt/vocanova/production
  ok: compose references the production secrets tree
[INS-11] neither tier's deploy identity can read the other's secrets
  FAIL: deploy is NOT read-blocked from /opt/vocanova/production/secrets/api.env: deploy already has independent broad sudo (member of the 'sudo' group (blanket sudo via the standard %sudo sudoers default)) and can read anything as root regardless of file permissions - directory-based isolation does NOT hold against deploy on this shared host
  FAIL: deploy is NOT traversal-blocked from /opt/vocanova/production: deploy already has independent broad sudo (member of the 'sudo' group (blanket sudo via the standard %sudo sudoers default)) and can read anything as root regardless of file permissions - directory-based isolation does NOT hold against deploy on this shared host
  ok: staging secrets tree absent on this host; production-to-staging probe not applicable
FAIL: 2 production/staging secret boundary check(s) failed
```

Run for real as the non-root `vocanova-production` identity (the same
identity `deploy-production.yml` actually uses), not as root.

## Confirmed residual risk: `deploy`'s blanket sudo

`deploy` (staging's real deploy user on this shared host) is a member of
the `sudo` group, which grants it blanket `(ALL:ALL)` root access via
Ubuntu's standard `%sudo` sudoers default - independent of, and predating,
anything this package built. This means `deploy` can always read
production's secrets (`sudo cat /opt/vocanova/production/secrets/api.env`)
regardless of file ownership, mode, or any directory-isolation control
`VOC-037-D01`/`T06` put in place. The file-permission checks (`INS-9`) are
real and correct, but they only stop *accidental* cross-tier access (e.g. a
careless script assuming shared paths, the recursive-`chown` regression
this task's own `deploy-staging.yml` fix addressed) - they cannot stop a
*deliberate* read by an identity that already has root.

This was not previously disclosed: earlier revisions of this document
recorded `INS-11` as passing because the rehearsal script was run manually
as root (trivially bypassing the exact boundary being tested), not as the
real `vocanova-production` deploy identity the automated workflow actually
uses. Corrected here.

**Remediation requires one of:**
1. Narrow `deploy`'s sudoers to the minimum commands `deploy-staging.yml`
   actually needs (mirroring the scoped grant already given to
   `vocanova-production`: `mkdir`/`tar`/`chown`/`touch`/etc., not blanket
   `ALL`), removing it from the `sudo` group. This is the only fix that
   makes `INS-11` genuinely pass - a live-staging-affecting change,
   deliberately not made unilaterally in this task.
2. An explicit founder-accepted waiver of this specific finding, if the
   risk is judged acceptable for now (`deploy` is founder/automation-
   controlled, not attacker-reachable, and the real risk is scope-of-
   blast-radius on a compromised or buggy staging deploy, not an external
   threat).

Until one of those happens, `deploy-production.yml`'s final step will
correctly keep failing - this is the intended, honest behavior, not a
regression to fix by loosening the check again.

## Notes

- This task provisions the production deployment shape and isolation
  controls. It does not close R2, authorize launch, or activate autonomous
  production release; founder go/no-go remains `VOC-037-T05`.
- **Correction to this file's earlier draft:** the note previously here
  claimed `apps/api/migrations/*.sql` still carried the invalid
  `-- atlas:txmode transaction` directive and a duplicate-index collision
  that would block a green production deploy. That was stale — `VOC-033`
  already fixed both upstream before this task ran, and the real production
  migration ("Real verification" above) applied all 13 migrations cleanly
  with no such error.
- **Known, disclosed gap:** production's TLS certificate is a real
  Cloudflare Origin CA certificate, but production and staging still share
  one physical host (`VOC-037-D00`'s second supersession). The shared-host
  resource/fault-domain risk that decision explicitly accepted is real and
  in effect, mitigated but not eliminated by the resource limits and
  isolation controls in this task.
