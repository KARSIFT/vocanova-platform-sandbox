# VOC-079 — Acceptance Criteria

## VOC-079-AC-00 — Repository-controlled Cloudflare verification reports no :8443 remap

- Requirement source: issue #624; `VOC-079-D00`; `VOC-079-DEP-02`; VOC-067-AC-06
- Tasks: `VOC-079-T00`
- Tests: `VOC-079-TEST-00`
- Evidence: `VOC-079-EV-00`
- Result: satisfied — verify-only run #39
  (https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31876429297);
  VOC-067-EV-05 `cloudflare_remap_api_status: absent` (recorded with T01
  remediation that closes the T00 gate before URL normalization)

Observable outcome:

1. A real `deploy-production.yml` run (or equivalent production-environment
   invocation of the existing cutover script path) with
   `voc067_cloudflare_origin_cutover=verify-only` exits 0 using
   `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`.
2. Redacted evidence shows zone resolution succeeded and output includes
   `OK: no origin rules remap production hosts to port 8443` (or, if
   initially `FOUND`, a subsequent authorized `--apply` / `apply` dispatch
   followed by a successful verify-only `OK`).
3. `specs/changes/VOC-067-…/t05-live-cutover-evidence.md` frontmatter
   `cloudflare_remap_api_status` is set to `absent` only after that `OK`.
4. Missing credentials, `zone not found`, or dashboard-only claims without
   a repository transcript do **not** satisfy this criterion.

## VOC-079-AC-01 — Production Compose defines only PostgreSQL, API, and web

- Requirement source: issue #624 required change 2
- Tasks: `VOC-079-T02`
- Tests: `VOC-079-TEST-01`, `VOC-079-TEST-02`
- Evidence: `VOC-079-EV-02`
- Result: satisfied — PR #660 / `VOC-079-EV-02`; production Compose contains
  only PostgreSQL, API, and web, with nginx/TLS material preserved for the
  shared-edge read-only mounts

Observable outcome: `infra/docker-compose.production.yml` has no `nginx`
service and no `vocanova-production-nginx` / `8081:80` / `8443:443`
publish. Production `nginx/conf.d` and TLS material remain available for
shared-edge read-only mounts. Secrets/directory/deploy-user/network
isolation is not weakened.

## VOC-079-AC-02 — Exactly one VocaNova nginx remains; ports 8081/8443 unpublished

- Requirement source: issue #624 acceptance criteria
- Tasks: `VOC-079-T02`, `VOC-079-T03`
- Tests: `VOC-079-TEST-02`, `VOC-079-TEST-05`
- Evidence: `VOC-079-EV-02`, `VOC-079-EV-03`
- Result: satisfied — production run 31884987715 removed the orphaned bridge
  declaratively; live inspection found one VocaNova nginx and no 8081/8443
  listeners (`VOC-079-EV-03`)

Observable outcome after normal production deployment:

1. The only VocaNova nginx container is `vocanova-shared-edge-nginx`.
2. Host ports `8081` and `8443` are no longer published by the production
   stack.
3. Convergence used repository deploy + scoped orphan removal — no manual
   SSH / `docker rm` required for acceptance.

## VOC-079-AC-03 — Shared edge serves all four hostnames on origin 443; external checks pass

- Requirement source: issue #624
- Tasks: `VOC-079-T03`
- Tests: `VOC-079-TEST-05`
- Evidence: `VOC-079-EV-03`
- Result: satisfied — all four canonical external checks returned HTTP 200
  after production convergence (`VOC-079-EV-03`)

Observable outcome: shared edge continues Host/SNI routing for
`staging.vocanova.site`, `api-staging.vocanova.site`,
`production.vocanova.site`, and `api-production.vocanova.site`. All four
external web/API checks pass through Cloudflare on canonical HTTPS URLs
**without** explicit ports (ordinary `:443`).

## VOC-079-AC-04 — Canonical production URLs use ordinary HTTPS :443

- Requirement source: issue #624 required change 3; VOC-067-AC-05
- Tasks: `VOC-079-T01`
- Tests: `VOC-079-TEST-03`
- Evidence: `VOC-079-EV-01`
- Result: satisfied — PR #659 / `VOC-079-EV-01`; current compose, OAuth,
  readiness, session-mint, smoke, and operator paths use canonical HTTPS
  without `:8443`

Observable outcome: current operational configuration
(`infra/docker-compose.production.yml` `API_BASE_URL`,
`deploy-production.yml` emitted `BASE_URL` / OAuth redirect/allowlist,
readiness/smoke/session-mint URL construction, and current operator docs)
contains no production `:8443` qualifications. Historical evidence packages
are not rewritten as though the bridge never existed.

## VOC-079-AC-05 — Deploy write isolation and shared-edge reload semantics preserved

- Requirement source: issue #624 required change 4; VOC-067-AC-03 / AC-04
- Tasks: `VOC-079-T02`
- Tests: `VOC-079-TEST-04`
- Evidence: `VOC-079-EV-02`
- Result: satisfied — PR #660 / `VOC-079-EV-02`; scoped project writes and
  orphan removal remain isolated, with fail-closed shared-edge validation and
  reload retained

Observable outcome:

1. Routine staging and production deploys still write only their own nginx
   configuration/certificate trees.
2. No production or staging secret is copied into the other tier's writable
   tree.
3. Fail-closed `nginx -t` followed by reload remains for
   `vocanova-shared-edge-nginx`.
4. Validation/reload steps targeting `vocanova-production-nginx` are gone.
5. Routine application deploys do not recreate the shared-edge container.

## VOC-079-AC-06 — Single-edge invariants replace the bridge-retention gate

- Requirement source: issue #624 required change 5
- Tasks: `VOC-079-T02`
- Tests: `VOC-079-TEST-01`, `VOC-079-TEST-02`, `VOC-079-TEST-03`
- Evidence: `VOC-079-EV-02`
- Result: satisfied — `voc079-single-edge-invariants.test.mjs` replaces the
  bridge-retention gate and passed CI on PR #660 (`VOC-079-EV-02`)

Observable outcome: deterministic checks assert (a) production Compose has
no nginx service, (b) only shared-edge Compose publishes host `80`/`443`,
(c) shared edge remains attached to both tier networks, (d) current
operational configuration has no production `:8443` URLs, and (e) the old
“keep bridge until unconfirmed” gate no longer blocks the completed
end state once AC-00 is met.

## VOC-079-AC-07 — Rollback documented and rehearsed; no manual server surgery

- Requirement source: issue #624 acceptance criteria; risk sequencing
- Tasks: `VOC-079-T03`
- Tests: `VOC-079-TEST-06`
- Evidence: `VOC-079-EV-03`
- Result: satisfied — normal production deployment rehearsed declarative bridge
  removal; the autonomous production workflow owns rollback to last-known-good
  revision `58f803b`, with Cloudflare restore limitations recorded in
  `VOC-079-EV-03`

Observable outcome:

1. Rollback documents that redeploying the prior revision recreates the
   bridge, and that the existing Cloudflare `--restore` path remains
   available if required.
2. Rollback owner is named.
3. Deployment and cleanup require no manual SSH edits or manual Docker
   commands for the happy path or the documented rollback path.
