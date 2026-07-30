# Staging infrastructure

This directory is the staging-tier infrastructure layout the
**VOC-032** change package built and ships for the
`vocanova.site` staging environment.

> **DOC-11 contradiction caveat (`VOC-032-D02`, resolved at
> adoption 2026-07-28).** This layout is **not** the
> target-infrastructure baseline currently described in
> `docs/operations/11-devops-and-ci-cd.md` §1, which still
> calls for Cloudflare Workers (frontend) + Render Web
> Service (backend) + Render PostgreSQL (database) on the
> `vocanova.com` domain set. The package's founder-directed
> deploy shape — self-hosted Docker Compose + nginx on the
> founder's own 2 vCPU / 4 GB server, Cloudflare for DNS /
> TLS / WAF / CDN only — is the staging-tier reality now.
> `VOC-032-T13` amends DOC-11 §1 to this shape once
> `T00`–`T09` have actually landed; until that amendment
> merges, treat this README as describing the staging tier
> **as built by VOC-032**, not as the project's permanent,
> approved target.

## Layout

```
infra/
├── README.md                  # this file (VOC-032-T11)
├── docker-compose.yml         # four-service stack: postgres + api + web + nginx
├── nginx/
│   ├── nginx.conf             # main config (events + http, includes conf.d/*.conf)
│   ├── conf.d/
│   │   ├── 00-cloudflare-real-ip.conf  # real-IP restoration scoped to CF ranges
│   │   ├── 01-tls.conf                 # shared TLS / ssl_protocols / ssl_ciphers
│   │   ├── 02-docker-dns.conf          # resolver + Docker-embedded DNS variables
│   │   ├── 05-default.conf             # catch-all server, 444 closes
│   │   ├── 10-staging-web.conf         # staging.vocanova.site -> web
│   │   └── 20-api-staging.conf         # api-staging.vocanova.site -> api
│   └── generate-dev-cert.sh   # self-signed cert for local compose validation
└── secrets/
    └── .gitignore             # untracked env files + TLS material; see below
```

Related artifacts this layout depends on but does not own:

```
apps/api/Dockerfile                                  # T02
apps/api/.env.example                                # T01 (api env schema)
apps/api/atlas.hcl                                   # T06
apps/api/scripts/migrate.sh                          # T06
apps/api/migrations/                                 # T06's tool reads this
apps/api/cmd/api/main.go                             # T00 (real server)
apps/api/cmd/eval-live/main.go                       # T10 (one live AI eval)
apps/api/business/aifeedback/                        # T08 (mock-eval gate) lives in CI
apps/api/foundation/email/http.go                    # T14 (real Sender)
apps/api/business/auth/google_oauth.go               # T15 (real OAuthProvider)
apps/web/Dockerfile                                  # T03
apps/web/next.config.ts                              # T03 (output: 'standalone')
apps/web/.env.example                                # T01 (web env schema)
.github/workflows/deploy-staging.yml                 # T07
```

The actual docker-compose service definitions are in
`infra/docker-compose.yml`; this README documents what they
mean, not their full syntax.

## Service architecture

`docker compose -f infra/docker-compose.yml` brings up four
services on a single internal Docker network (`vocanova-net`):

| Service   | Image                              | Port (host) | Reachable from           |
|-----------|------------------------------------|-------------|--------------------------|
| `postgres`| `postgres:16-alpine`               | none        | internal network only    |
| `api`     | `apps/api/Dockerfile` (T02)        | none        | internal network only    |
| `web`     | `apps/web/Dockerfile` (T03)        | none        | internal network only    |
| `nginx`   | `nginx:1.27-alpine`                | 80, 443     | the internet (via CF)    |

**Only `nginx` publishes host ports.** The database and the
two app services are reachable only on the internal network
— this is the explicit `VOC-032-R03` mitigation (a misconfigured
`ports:` block exposing `postgres`'s `5432` to the host would
expose the database to the internet, which this package's
impact analysis explicitly forbids). Cross-service addressing
uses Docker's embedded DNS: `postgres:5432`, `api:8080`,
`web:3000`. Every service has a `HEALTHCHECK`; the api's
`/healthz` pings the live database connection; nginx
intentionally exposes a shallow 200 probe only (the real
routing checks are `infra/nginx/conf.d/10-staging-web.conf`
and `20-api-staging.conf`).

The `web` image is built with the monorepo root as its build
context (so pnpm can resolve workspace dependencies). The
`NEXT_PUBLIC_API_BASE_URL` build argument is the value
inlined into the client bundle at `next build` time — Next.js
does not honor a runtime override for `NEXT_PUBLIC_*` values.
For staging this build arg must be
`https://api-staging.vocanova.site`; for local dev it defaults
to `http://localhost:8080`.

## Staging subdomains and TLS

`VOC-032-D03` proposes two staging subdomains, mirroring
DOC-11 §1's existing `*-staging` naming convention adapted to
the real `vocanova.site` domain:

- `staging.vocanova.site` — the `web` service, browser-facing
- `api-staging.vocanova.site` — the `api` service, browser
  and server-side fetch target

The apex `vocanova.site` is **not** used by this milestone;
keeping the apex unused avoids any appearance that a staging
deploy is a production activation, and avoids colliding with
whatever the apex domain is eventually used for. The founder
must create the corresponding Cloudflare DNS A/AAAA records
pointed at the staging server's IP and confirm proxy
("orange-cloud") status — see `VOC-032-DEP-01`.

TLS is terminated at nginx using a **Cloudflare origin
certificate** (Cloudflare-issued, signed for the
`*.vocanova.site` and `vocanova.site` names), mounted from
`infra/secrets/nginx/cert.pem` and `key.pem`. The
configuration assumes Cloudflare's "Full (strict)" SSL mode
(origin presents a certificate Cloudflare validates). If the
founder instead wants a different Cloudflare TLS mode, that
is a `VOC-032-DEP-01` follow-up, not a silent assumption
this README should paper over.

The real client IP is restored from Cloudflare's
`CF-Connecting-IP` header via `set_real_ip_from` entries
scoped to **Cloudflare's published IP ranges only** (see
`infra/nginx/conf.d/00-cloudflare-real-ip.conf`). A direct,
non-Cloudflare-routed request to the origin server's IP
**cannot** spoof its way past the restoration — that is
the explicit `VOC-032-R01` mitigation against bypassing
`apps/api/business/auth`'s IP-based rate limiting.

## Secrets and founder-populated host files

`infra/secrets/.gitignore` keeps the contents of `infra/secrets/`
out of git. The founder must populate the following files on
the staging host before the first `docker compose up`; the
schemas are documented in the corresponding `.env.example`
files, never in `docker-compose.yml` itself:

| File                            | Owner    | Schema reference            |
|---------------------------------|----------|-----------------------------|
| `infra/secrets/postgres.env`    | founder  | inline (just `POSTGRES_PASSWORD=…`) |
| `infra/secrets/api.env`         | founder  | `apps/api/.env.example` (T01)|
| `infra/secrets/web.env`         | founder  | `apps/web/.env.example` (T01); may be empty/absent — the web has no required runtime env |
| `infra/secrets/nginx/cert.pem`  | founder  | Cloudflare origin cert      |
| `infra/secrets/nginx/key.pem`   | founder  | Cloudflare origin private key |

The four staging-tier repository secrets
`STAGING_SSH_HOST` / `STAGING_SSH_USER` /
`STAGING_SSH_PRIVATE_KEY` / `STAGING_SSH_KNOWN_HOSTS` are
GitHub Actions secrets, not host files; they are documented
in `.github/workflows/deploy-staging.yml`'s own header and
added by the founder under `VOC-032-DEP-00`. The
`apps/api/cmd/api` startup path fails fast on any missing
required env var (see `apps/api/.env.example`'s `[REQUIRED]`
markers) — a half-configured `api.env` is a loud, immediate
startup error, never a silent fallback.

The DOC-11 §3 kill switches
`AI_FEATURES_ENABLED` / `EMAIL_MAGIC_LINK_ENABLED` /
`GOOGLE_OAUTH_ENABLED` / `NEW_USER_SIGNUP_ENABLED` are also
set in `api.env`. The real email `Sender` (T14) and real
Google `OAuthProvider` (T15) both fall back to their
respective `Fake{}` implementations when the matching
provider credential is absent or the corresponding kill
switch is off, so staging can still run with magic-link
delivery off or Google sign-in off rather than crashing at
startup.

## Staging host layout (deploy target)

The `deploy-staging` workflow (T07) and the `T09` manual
rehearsal both assume the staging host has the following
base path layout (this is the load-bearing convention
referenced by the deploy workflow's `scp` and `ssh` steps;
changing the path here is a coordinated change with the
workflow file, not a one-sided edit):

```
/opt/vocanova/
├── infra/                              # this directory on the host
│   ├── docker-compose.yml              # SCPed by deploy-staging
│   ├── nginx/...                       # SCPed by deploy-staging
│   └── secrets/                        # founder-populated, NOT SCPed
├── apps/api/scripts/migrate.sh         # SCPed by deploy-staging
├── apps/api/migrations/                # SCPed by deploy-staging
└── usr/local/bin/atlas                 # installed by deploy-staging (idempotent)
```

The `apps/api/migrations/atlas.sum` integrity file is part
of the deploy bundle; the `migrate.sh` wrapper refuses to
proceed if it is missing.

## Local development and validation

```bash
# 1. Validate the compose file's schema (does not require
#    any of the secret files to exist; compose's
#    `required: false` on every env_file relaxes only the
#    config-time existence check, not runtime semantics).
docker compose -f infra/docker-compose.yml config

# 2. Build the api and web images locally.
docker compose -f infra/docker-compose.yml build

# 3. Bring the four-service stack up locally. This requires
#    a populated infra/secrets/ directory (postgres.env +
#    api.env + a self-signed cert/key generated by
#    infra/nginx/generate-dev-cert.sh into
#    infra/secrets/nginx/) and a database the api can reach
#    (the compose `postgres` service is the obvious target).
docker compose -f infra/docker-compose.yml up -d

# 4. Verify the api's /healthz reports 200 and the database
#    ping passes.
docker compose -f infra/docker-compose.yml exec api wget -q -O- http://127.0.0.1:8080/healthz

# 5. Stop the stack. Does NOT remove the postgres named
#    volume; data persists across `down` / `up` cycles.
docker compose -f infra/docker-compose.yml down
```

The `nginx -t` syntax-check used in `T05`'s
`VOC-032-TEST-12` is:

```bash
docker run --rm \
  -v $(pwd)/infra/nginx/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v $(pwd)/infra/nginx/conf.d:/etc/nginx/conf.d:ro \
  -v $(pwd)/infra/secrets/nginx/cert.pem:/etc/nginx/certs/cert.pem:ro \
  -v $(pwd)/infra/secrets/nginx/key.pem:/etc/nginx/certs/key.pem:ro \
  nginx:1.27-alpine nginx -t
```

The Atlas forward-apply command (used by `T07`'s deploy
workflow and by `T09`'s rehearsal) is the wrapper at
`apps/api/scripts/migrate.sh`, not Atlas directly:

```bash
DATABASE_URL=postgres://vocanova:vocanova@127.0.0.1:5432/vocanova?sslmode=disable \
  sh apps/api/scripts/migrate.sh
```

The `T08` AI-evaluation-threshold CI gate is **not** an
infra artifact — it is a Go test in
`apps/api/business/aifeedback/` (mock-provider golden-set
assertions) wired as a required check in CI. The matching
`T10` live-provider evaluation pass is the one-shot
`apps/api/cmd/eval-live` command, run against the real
OpenCode provider in staging, gated on `VOC-032-DEP-03` and
recorded as `EV-22` in `staging-evidence.md`.

## How to reach staging (once it exists)

Once the founder has provisioned the host, the DNS records,
the Cloudflare origin certificate, and the GitHub Actions
secrets, the staging environment is reachable at:

- `https://staging.vocanova.site` — the deployed web app
  (browser-facing; `staging.vocanova.site` is the staging
  apex)
- `https://api-staging.vocanova.site` — the deployed API
  (browser and server-side fetch target)
- `https://api-staging.vocanova.site/healthz` — the api's
  unauthenticated health check, returning 200 only when the
  database ping succeeds

A live magic-link sign-in and a live Google sign-in each
require the third-party accounts the founder must provision
under `VOC-032-DEP-07` (a transactional-email provider
account; a Google Cloud OAuth 2.0 client). Until those
accounts exist, magic-link and Google sign-in both fall back
to their `Fake{}` implementations and cannot be fully
demonstrated end-to-end — this is the same gap `T14`/`T15`
record as a unit-tested code path plus a one-time live
evidence step.

## Dependencies and open blockers

This package's staging-tier reality is fully operational
once each of the following resolves. None of them can be
provisioned by the planner, implementer, or independent
reviewer — they are founder-owned:

- **`VOC-032-DEP-00`** — `STAGING_SSH_HOST` /
  `STAGING_SSH_USER` / `STAGING_SSH_PRIVATE_KEY` /
  `STAGING_SSH_KNOWN_HOSTS` (T07 / T09)
- **`VOC-032-DEP-01`** — Cloudflare origin TLS cert/key and
  the two staging-subdomain DNS A/AAAA records
  (T05 / T09)
- **`VOC-032-DEP-03`** — staging AI-provider credentials for
  the live `T10` evaluation pass
- **`VOC-032-DEP-07`** — a real transactional-email-provider
  account (T14) and a real Google Cloud OAuth 2.0 client
  (T15)

`T05`, `T07`, `T09`, `T10`, `T14`, `T15` are all
additionally blocked on the matching dependency above;
`T00`–`T04`, `T06`, `T08`, and `T11` itself are not.

The `T06` follow-ups — an invalid
`-- atlas:txmode transaction` directive in the existing
migration files and a duplicate
`streak_states_user_id_key` unique-index in
`apps/api/migrations/20260725130002_voc030_p4_gamification_tables.sql` —
are recorded in
`staging-evidence.md`'s T06 section as pre-existing
incompatibilities between the migration files and the
Atlas tool this package introduces. They do not block
`T00`–`T04` / `T06` / `T08` / `T11` build/CI-wiring; they
do block `EV-14`'s end-to-end Atlas-apply pass against the
existing migration set until either the directives are
changed in place (a protected-area edit, separate package
or narrow T06 exception) or a disposable scratch database
is used that does not exercise the duplicate-index file.

## Production deployment (not authorized)

This package authorizes **no production deployment**. RL1
and RL2 technical activation and autonomous production
release remain disabled per
`docs/governance/a003-transition-state.yaml`; the deploy
target this README documents is **staging only**, on the
founder-provisioned host.

## Cross-references

- VOC-032 change package:
  `specs/changes/VOC-032-begin-milestone-r1-staging-readiness-docs-product/`
  (specification, acceptance criteria, implementation plan,
  tasks, impact analysis, staging evidence, mock inventory)
- DOC-11 §1 (target-infrastructure baseline, still pre-amendment
  at the time of writing — see caveat at the top of this
  file): `docs/operations/11-devops-and-ci-cd.md`
- DOC-12 §5 (R1 gate, including "stable in staging, no
  unresolved critical/high blocker, all required tests pass,
  migration + rollback rehearsed, AI evaluation thresholds
  pass, founder completes staging acceptance, scope frozen
  after"): `docs/product/12-mvp-implementation-plan.md`
- DOC-12 §3 (F3 "Staging Foundation" milestone; VOC-032
  folds F3's undone infrastructure scope into R1 per
  `VOC-032-D04`)
- DOC-09 §23 (AI evaluation thresholds the T08 gate asserts):
  `docs/engineering/09-ai-features.md`
- Apps-owned conventions: `apps/api/.env.example` (T01),
  `apps/api/migrations/README.md`, `apps/api/ent/README.md`,
  `apps/web/.env.example` (T01)
- T12 gate-readiness summary:
  `specs/changes/VOC-032-.../staging-evidence.md`
  (the canonical record of what this package's evidence
  satisfies vs. what remains founder-owned)
