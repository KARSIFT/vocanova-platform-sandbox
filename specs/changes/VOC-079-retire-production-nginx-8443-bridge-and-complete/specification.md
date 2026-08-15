# VOC-079 — Retire Production nginx :8443 Bridge and Complete the Single Shared-Edge Cutover: Specification

## Objective and requirement source

Complete the repository-driven single-nginx cutover described in
[GitHub issue #624](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/624):
retire the obsolete production nginx `:8081`/`:8443` bridge so the final host
topology has exactly one VocaNova nginx process —
`vocanova-shared-edge-nginx` — serving staging and production on ordinary
origin ports `80`/`443`, with all convergence performed by the repository and
normal deployment workflow (no manual SSH edits or ad hoc container removal).

Primary context (issue #624 and drafting-time repo read):

| Item | Value |
|------|--------|
| Predecessor | [VOC-067](../VOC-067-production-outage-root-cause-consider-unifying) — shared edge + temporary bridge |
| Credential predecessor | [VOC-072](../VOC-072-same-request-as-github-issue-535-voc-067-t05) — `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` |
| VOC-067-EV-05 gate | `cloudflare_remap_api_status: unconfirmed` — bridge must stay until `absent` |
| Live origin `:443` (2026-08-15) | HTTP 200 for staging/production web + API healthz |
| Founder dashboard | No Origin Rules for `vocanova.site` remapping production to `:8443` |
| Superseded retain decision | Issue #595 closed retain-bridge; founder direction 2026-08-15 finishes cutover |

**Objective:** after this package's implementation and normal production
deploy, (a) repository verify-only reports no production origin-port remap to
8443 and EV-05 records `absent`, (b) production Compose defines only
PostgreSQL, API, and web, (c) exactly one nginx container remains and host
ports `8081`/`8443` are unpublished, (d) operational configuration uses
ordinary HTTPS `:443` URLs, and (e) rollback is documented without requiring
manual server surgery.

## Confirmed findings (drafting-time)

- `infra/docker-compose.production.yml` still defines `nginx` /
  `vocanova-production-nginx` publishing `"8081:80"` / `"8443:443"` and sets
  `API_BASE_URL: https://api-production.vocanova.site:8443`.
- `infra/docker-compose.shared-edge.yml` already binds `80`/`443`, joins
  `vocanova-net` and `vocanova-production-net`, and mounts isolated staging /
  production conf and cert trees read-only.
- `.github/workflows/deploy-production.yml` still validates/reloads
  `vocanova-production-nginx`, polls readiness on `:8443`, and emits
  `:8443`-qualified `BASE_URL` / OAuth values; it already fail-closes
  shared-edge `nginx -t` + reload and does not recreate shared-edge on
  routine deploys.
- `scripts/foundation/voc067-cutover-bridge-gate.test.mjs` fails if compose
  drops the bridge while VOC-067-EV-05 status is not `absent`.
- Foundation tests
  `docker-compose-production-api-base-url-port.test.mjs` and
  `deploy-production-oauth-port.test.mjs` currently **require** `:8443`
  (VOC-041/VOC-042 era) and must be inverted, not left asserting the bridge
  era.
- VOC-067-T04 remains pending; VOC-072-T02 verify-only evidence was still
  pending operator execution at drafting. Issue #624 supplies a fresh
  governed path rather than guessing those task attempt budgets.

## Scope and non-goals

In scope:

1. **Cloudflare absence evidence (T00):** repository
   `deploy-production.yml` `voc067_cloudflare_origin_cutover=verify-only`
   with the zone-scoped token; redacted transcript; update VOC-067-EV-05
   `cloudflare_remap_api_status` to `absent` only on success output
   `OK: no origin rules remap production hosts to port 8443`. If output is
   `FOUND:…`, run authorized `--apply` (or dispatch `apply`) per existing
   VOC-067 tooling **before** setting `absent` or retiring the bridge.
2. **Canonical URLs (T01):** remove operational `:8443` from production
   `API_BASE_URL`, deploy-emitted `BASE_URL` / OAuth redirect/allowlist,
   readiness probes, session-mint calls, smoke tests, and current operator
   documentation; use ordinary HTTPS port 443; preserve historical package
   evidence wording.
3. **Compose + workflow + invariants (T02):** remove production compose
   `nginx` service while preserving production `conf.d` / TLS material for
   shared-edge read-only mounts; scoped declarative orphan removal on
   production deploy; remove production-nginx validate/reload steps; retain
   shared-edge fail-closed `nginx -t` + reload; replace bridge-retention
   gate with single-edge invariants (no production nginx service; only
   shared-edge publishes host `80`/`443`; shared edge on both networks; no
   operational production `:8443` URLs).
4. **Live verification (T03):** normal production deploy converges host
   state; external checks on canonical HTTPS URLs; confirm ports/container
   absence; rollback rehearsal and monitoring window with named owner.

Non-goals / explicitly excluded:

- `monitor.vocanova.site` / Uptime Kuma.
- Cookie isolation between tiers.
- General Sentry work.
- Unrelated hardening.
- Manual SSH or manual Docker commands as the cleanup path.
- Rewriting historical VOC-041/VOC-042/VOC-067 narratives.
- Snapshot-then-recheck-drift promotion tasks (not applicable; this package
  introduces new infra/workflow content, then live evidence).
- Adopting or authorizing this package from within the draft.

## Risk and protected areas

Builder assessment: expected paths include `infra/*` and
`.github/workflows/deploy-*.yml`, which the path classifier floors at
**R3**. Drafting-time
`scripts/governance/classify-change-risk.sh --files-from` against the
expected list reported **Detected path-based risk floor: R3**.

This package **proposes R4** for the change as a whole because it retires
the temporary production edge listener, changes the live rollback path, and
completes the VOC-067 shared-edge supersession of `VOC-037-D00` that was
itself R4. This is a **draft proposal for the reviewing human at adoption
time, not a determination**. Each task's real file list must be re-measured;
the independent verifier may raise or lower.

Protected areas: production infrastructure (`infra/`), deployment workflows,
secrets-boundary / per-tier write isolation (must not regress VOC-037-D01 /
VOC-067-AC-03), Cloudflare origin routing (ops), and production TLS material
(never commit keys).

Under active A-003, routine R3 does not require standing technical-steward
approval merely for being R3; **R4 founder authority remains required**. EHR
is not triggered by this drafting pass.

## Decisions, contradictions, security, and privacy

`VOC-079-D00` (recorded here for traceability; formal acceptance at
adoption): The final steady-state edge is **only**
`vocanova-shared-edge-nginx` on host `80`/`443`. The production compose
bridge is temporary cutover scaffolding, not a permanent dual-nginx design.
Issue #595's retain-bridge closure is superseded by issue #624 / founder
direction 2026-08-15. Bridge retirement is forbidden until repository
verify-only reports remap absence (dashboard confirmation is supporting
context only).

`VOC-079-D01`: Cleanup must be declarative via the production deploy path
(including scoped orphan removal). Manual `docker stop/rm` is out of scope
and must not be required for acceptance.

Contradiction with issue #595: explicit supersession, not silent rewrite —
cite #624 in evidence.

Open questions for the reviewing human:

1. **`VOC-079-DEP-00` — Predecessor task disposition.** Recommended default:
   VOC-079 supersedes unfinished VOC-067-T04 and the remaining VOC-067-AC-05/
   AC-06 / VOC-072-T02 cutover proof; update VOC-067 evidence/AC Results from
   VOC-079 evidence; do not redispatch those exhausted tasks.
2. **`VOC-079-DEP-01` — Orphan removal shape.** Recommended default: on the
   production compose project only, use a declarative mechanism equivalent to
   `docker compose … up -d --remove-orphans` (or the workflow's existing
   compose invocation plus an explicit, project-scoped orphan flag) so
   `vocanova-production-nginx` is removed when dropped from the compose file,
   without touching `vocanova-shared-edge` or staging projects.
3. **`VOC-079-DEP-02` — Verify-only mandatory.** Recommended default: T00
   repository transcript required before T02 merges; founder dashboard note
   alone does not unlock bridge retirement.
4. **Risk.** Accept proposed R4 (path floor R3).

Security / privacy: no new personal-data handling. Zone-scoped Cloudflare
token remains founder/ops-held in the GitHub `production` environment; never
commit token values. Staging/production secrets, directory trees, deploy
users, and Docker networks stay isolated; shared edge continues to mount
each tier's conf/certs read-only from that tier's tree only.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None. No schema or seed change.
- **Analytics:** None.
- **Accessibility:** None. No UI change.
