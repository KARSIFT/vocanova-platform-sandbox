# VOC-042 — infra/docker-compose.production.yml Hardcodes web's API_BASE_URL Without the Required :8443 Port, Still Breaking Real Google Sign-In After VOC-041: Specification

## Objective and requirement source

Restore real Google sign-in past the post-OAuth-callback middleware auth check by
making `infra/docker-compose.production.yml`'s `web` service's `API_BASE_URL`
environment value (currently `https://api-production.vocanova.site`, no port,
lines 82-118, specifically line 107) include the `:8443` port production actually
serves its API on — matching the correction VOC-041 already applied to the API
service's own `BASE_URL`/`OAUTH_REDIRECT_URI`/`OAUTH_REDIRECT_ALLOWLIST` values, and
the interim dispatch-time override already used for `NEXT_PUBLIC_API_BASE_URL`.
Grounded in [issue #319](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/319)
in full, including its confirmed-by-live-reproduction root cause (direct SSH
`docker exec` output, nginx access logs from a real sign-in attempt, and a read of
`apps/web/src/lib/env.ts`'s `getApiBaseURL()`) and its named suggested fix. Not yet
approved by a founder or technical steward — see `change.yaml`'s
`requirement_approval_status`.

## Scope and non-goals

In scope:
- `infra/docker-compose.production.yml`'s `web` service `environment:` block
  (currently line 107): change `API_BASE_URL: https://api-production.vocanova.site`
  to `API_BASE_URL: https://api-production.vocanova.site:8443`.
- Confirming (and, if needed, extending) the existing explanatory comment at lines
  93-106 remains accurate after the port is added — the comment explains *why*
  `API_BASE_URL` is set at all (to avoid the internal `http://api:8080` address that
  VOC-038-T03 found silently failing server-side despite being reachable via a
  plain `wget` from inside the container); it does not currently assert anything
  about the port that this fix would contradict, so no correction is expected to be
  needed, only confirmation.
- A regression test (or documented deterministic check, per this repo's existing
  `pnpm validate`/narrower-script convention) that fails if `API_BASE_URL` in this
  service is ever written without `:8443`, so this specific recurrence of the
  same-host-missing-port defect class (now the third instance: VOC-038-T03's
  original CORS bug, VOC-041's `deploy-production.yml` values, and this one) is
  caught deterministically rather than only via a live production reproduction.

Non-goals:
- Any change to `deploy-production.yml` or any workflow input. This value is
  hardcoded directly in the compose file (not templated), so VOC-041's mechanism
  cannot reach it — confirmed by this package's own drafting-time read of both
  files.
- Any change to the `api` service's own `environment:` block in this same file.
  This package's drafting-time read confirms it does not set
  `API_BASE_URL`/`OAUTH_REDIRECT_URI`/`OAUTH_REDIRECT_ALLOWLIST` (those are written
  into `api.env` by `deploy-production.yml`, per VOC-041) — not implicated by this
  issue.
- Any change to the currently-running production `web` container. This package
  changes what the compose file specifies for the *next* recreation of the `web`
  service; it does not itself restart or recreate the running container. Whether an
  immediate manual recreate is needed before the next scheduled deploy is an
  operational decision for the reviewing human (see `README.md`'s recommended next
  action 3) — this package has no production access to perform it.
- A broader audit of every same-host URL construction site across the repository,
  even though this is now the third confirmed instance of the same bug class. See
  "Open questions" below.

## Risk and protected areas

`infra/docker-compose.production.yml` is production infrastructure configuration
that directly sets the server-side value `apps/web/src/middleware.ts` depends on for
every protected route's auth check. Per
`docs/governance/change-risk-classification.md`, "production infrastructure" is
named explicitly as an R3-floor category regardless of diff size (in this case, one
string literal). This package proposes `R3` (see `change.yaml`); it does not touch
schema migrations, billing, or a new secret, so no higher class is proposed, but the
reviewing human's own judgment governs this, not this proposal.
`scripts/governance/classify-change-risk.sh` has not been run against a real,
task-scoped file list at drafting time — consistent with how VOC-039/VOC-040/VOC-041
handled this field, that computation belongs to each task's own implementation PR.

## Decisions, contradictions, security, and privacy

No `VOC-042-D00`-style founder/product decision is defined by this draft — the fix
itself (matching the port every sibling env value in this deploy chain already
uses, or has just been corrected to use) is fully specified by the issue and
requires no product judgment call. If the reviewing human disagrees, they should
record why at adoption time rather than this package inventing a decision.

No contradiction is recorded for this package specifically (unlike VOC-041, no
sibling comment in this file makes an opposing empirical claim about the port) —
but see "Open questions" below for the broader recurring-pattern concern.

Security/privacy: this fix does not change *what* is authenticated, *who* can sign
in, or the auth-check logic in `apps/web/src/middleware.ts` itself — it corrects
the *value* the server-side fetch target is constructed from, from one that does
not reach the real API to one that does. No new attacker-controlled surface,
secret, or personal-data field is introduced. `API_BASE_URL` is a non-sensitive
configuration value (a public hostname), consistent with the file's own existing
treatment of it (plain `environment:` value, not sourced from `secrets/`).

## Data, migrations, analytics, and accessibility

None. This package touches only `infra/docker-compose.production.yml`'s `web`
service `environment:` block and its test coverage; no schema, migration, analytics
event, or accessibility surface is affected.

## Open questions

1. **This exact bug class (a same-host URL hardcoded or generated without the
   `:8443` port production actually serves on) has now recurred at least three
   times** — VOC-038-T03's original CORS bug, VOC-041's `deploy-production.yml`
   `BASE_URL`/`OAUTH_REDIRECT_URI`/`OAUTH_REDIRECT_ALLOWLIST` values, and this
   package's `docker-compose.production.yml` `API_BASE_URL` — all traceable to the
   same underlying interim host-sharing arrangement (production and staging on one
   host, production shifted to `:8443`, per VOC-037/T06's D00 supersession note).
   This package fixes only the one call site the issue names. Flagged for the
   reviewing human to decide whether a broader, deliberate audit of every
   same-host URL construction site (across compose files, workflow files, and
   application code) should be commissioned as a separate follow-up package, rather
   than continuing to discover each instance one live failure at a time.
2. **Whether the currently-running production `web` container needs a one-time
   manual recreate.** This fix only changes what the compose file specifies for the
   *next* recreation of the `web` service (a fresh `docker compose up -d web`, a
   full redeploy, or equivalent). The issue's own evidence shows the *currently
   running* container already has the unqualified (broken) value. Whether real
   sign-in stays broken until the next full dispatch, or whether an operator
   manually recreates the `web` service sooner, is an operational decision outside
   this package's scope (this package has no production access) — flagged in
   `README.md`'s recommended next action 3.
3. **`infra/docker-compose.yml` (the non-production/staging compose file) is not
   read by this package at drafting time beyond confirming it is a distinct file
   from the one this issue names.** Whether it has an equivalent `API_BASE_URL` (or
   similarly-purposed) value with the same or a different port defect is unknown at
   drafting time and is not asserted either way — flagged for the implementer to
   check and, if an equivalent defect exists there, record as a separate follow-up
   rather than silently fixing it inside this package's stated single-file scope.
