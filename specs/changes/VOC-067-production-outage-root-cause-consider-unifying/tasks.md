# VOC-067 — Tasks

## VOC-067-T00 — Record edge-architecture and lifecycle decision

- Requirement source: issue #485; `VOC-067-DEP-00`–`DEP-03`
- Acceptance criteria: `VOC-067-AC-00` (and `VOC-067-AC-07` if alternate path)
- Tests: `VOC-067-TEST-00`
- Evidence: `VOC-067-EV-00` —
  [`t00-edge-architecture-decision-record.md`](t00-edge-architecture-decision-record.md)
- Status: **complete** (2026-08-11)

Decision-record task only (no compose/workflow behavior change required in
this PR beyond the package evidence file itself). Produce a decision record
under this package directory that states:

1. Chosen path: **shared nginx** or **dual nginx + Cloudflare harden**.
2. If shared: accept/amend/reject the lifecycle defaults in
   `specification.md` (dedicated shared-edge compose; routine deploys only
   `nginx -t` + reload; recreate rare/gated; single `default_server`).
3. Explicit stance on superseding `VOC-037-D00`/`D01`'s separate-ports /
   separate-nginx clause while preserving secrets/directory/deploy-user
   isolation (`VOC-067-DEP-02`).
4. Cloudflare cutover owner and ordered steps (`VOC-067-DEP-03`).
5. If dual-nginx path: the concrete hardening tasks that replace T02–T05
   (or an explicit statement that a follow-up package will carry them), and
   mark T02–T05 cancelled for this package.

Do not open T02–T05 implementation until this record is accepted at adoption
(or updated post-adoption with the same human authority). T01 may proceed in
parallel.

## VOC-067-T01 — Fix permanently broken nginx HEALTHCHECK in both compose files

- Requirement source: issue #485; `VOC-067-DEP-04`
- Acceptance criteria: `VOC-067-AC-01`
- Tests: `VOC-067-TEST-01`
- Evidence: `VOC-067-EV-01` —
  [`t01-healthcheck-evidence.md`](t01-healthcheck-evidence.md)
- Status: **complete** (2026-08-11) — repository fix already landed on
  `develop` via `VOC-066` (commit `545a7ef`, PR #518); T01 evidence records
  satisfaction and local probe verification. Live `docker inspect` on staging
  and production nginx awaits the next normal deploy recreate (`VOC-066-DEP-02`).

In `infra/docker-compose.yml` and `infra/docker-compose.production.yml`
(and, if shared edge already exists when this lands, its compose file):

- Replace the Host-less `wget http://127.0.0.1/` HEALTHCHECK with a probe that
  succeeds when nginx is accepting connections (see `specification.md` open
  question 5 for allowed shapes).
- Do not weaken `05-default.conf`'s unmatched-Host `return 444` behavior for
  external traffic.
- Record before/after `docker inspect` health status (or equivalent) in
  evidence. Prefer a disposable/local or staging verification; do not require
  production downtime for this fix alone.

## VOC-067-T02 — Implement shared-edge nginx layout in repository infra

- Requirement source: issue #485 proposed direction; T00 shared-nginx decision
- Acceptance criteria: `VOC-067-AC-02`
- Tests: `VOC-067-TEST-02`
- Evidence: `VOC-067-EV-02` —
  [`t02-shared-edge-evidence.md`](t02-shared-edge-evidence.md)
- Status: **complete** (2026-08-11) — repository shared-edge layout + temporary
  `:8443` cutover bridge + staging controlled bring-up; live origin `:443`
  Host-routing proof remains a host/ops limitation recorded in EV-02

Implement the repository side of the shared edge per T00:

- Shared compose (or equivalent) binding host `80`/`443`, joining both Docker
  networks, mounting each tier's conf and certs at isolated in-container
  paths.
- Resolve dual `default_server` collision when both conf trees are loaded.
- Stop publishing production's public HTTPS via `8081`/`8443` as the primary
  path (temporary dual-publish during cutover is allowed only if T00's
  cutover plan explicitly requires it and documents removal).
- Update `infra/README.md` so it no longer claims Cloudflare origin-port
  remap is the steady-state design.
- Keep staging and production upstream service definitions otherwise
  isolated.

No deploy-workflow rewrite in this task beyond what is strictly required to
bring the shared container up in a controlled way; T03 owns deploy isolation
and reload semantics.

## VOC-067-T03 — Preserve per-tier deploy write isolation; safe shared reload

- Requirement source: issue #485 isolation bullets; `VOC-067-DEP-01`
- Acceptance criteria: `VOC-067-AC-03`, `VOC-067-AC-04`
- Tests: `VOC-067-TEST-03`, `VOC-067-TEST-04`
- Evidence: `VOC-067-EV-03`
- Status: pending — depends on T02; **authorized** by T00 shared-nginx decision

Update `.github/workflows/deploy-staging.yml` and
`.github/workflows/deploy-production.yml` (and header comments) so that:

- Each pipeline writes only its own nginx conf/certs tree.
- After writing, run `nginx -t` against the shared process; reload only on
  success; fail closed on failure without recreating the container.
- Neither pipeline gains write/`chown`/extract into the other tier's
  `/opt/vocanova/...` nginx or secrets paths.
- Secrets-boundary rehearsal expectations from VOC-037 remain satisfied
  (re-run or cite `rehearse-production-secrets-boundary.sh` as evidence where
  applicable).

## VOC-067-T04 — Remove production :8443 port-qualification workarounds

- Requirement source: issue #485 end state (ordinary `:443`); VOC-041 / VOC-042
  reversal once shared edge is live
- Acceptance criteria: `VOC-067-AC-05`
- Tests: `VOC-067-TEST-05`
- Evidence: `VOC-067-EV-04`
- Status: pending — depends on T02/T03; should not land before origin `:443`
  actually serves production (cutover order in
  [`t00-edge-architecture-decision-record.md`](t00-edge-architecture-decision-record.md);
  coordinate with T05)

Remove `:8443` from production client-facing and deploy-emitted URLs that
exist only because of the dual-nginx port split, including at least:

- `infra/docker-compose.production.yml` `API_BASE_URL`
- `deploy-production.yml` emitted `BASE_URL` / `OAUTH_REDIRECT_URI` /
  `OAUTH_REDIRECT_ALLOWLIST` (and any sibling health-check URL construction)
- Docs under `infra/README.md` / `docs/operations/` that assert remap or
  `:8443` as steady state

Do not rewrite historical VOC-041/VOC-042 package docs as if those packages
were wrong at the time — add forward-looking notes only where operators would
otherwise follow stale runbooks.

## VOC-067-T05 — Live cutover verification and rollback evidence

- Requirement source: issue #485; `VOC-067-DEP-03`; `VOC-067-AC-06`
- Acceptance criteria: `VOC-067-AC-06`
- Tests: `VOC-067-TEST-06`
- Evidence: `VOC-067-EV-05`
- Status: pending — depends on T02–T04; cutover order and Cloudflare API
  operator recorded in
  [`t00-edge-architecture-decision-record.md`](t00-edge-architecture-decision-record.md)

No large new feature code expected. Produce evidence that:

1. Shared edge serves both tiers on origin `:443` with correct Host routing.
2. Cloudflare production origin-port override is removed (or confirmed absent).
3. External checks for staging and production web/API succeed on `:443`.
4. Rollback is documented and rehearsed enough to be credible (restore remap
   and/or prior edge generation; name the accountable owner).

Prefer recording commands and outcomes in `VOC-067-EV-05` rather than adding
permanent debug workflows.

Tasks preserve scope, separation of duties, and rollback safety. No task may
be dispatched before this package is adopted. T02–T05 are cancelled for this
package if T00 selects the dual-nginx alternate path (`VOC-067-AC-07`).
