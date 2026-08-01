# VOC-037-EV-03 — Kill-switch and rollback evidence (T03)

## Standing of `VOC-037-AC-03` at this revision

**`VOC-037-AC-03` is NOT satisfied at this revision.** `VOC-037-TEST-03`
requires each of the four kill switches to be toggled *against the production
target* and a rollback-by-redeploy rehearsal to be completed *against the
production target*. Neither has happened yet: this task had no SSH access to
the production host and no `PRODUCTION_*` credentials, and no partial or
"repo-side first" split of `TEST-03` is authorized anywhere in
`test-plan.md`.

What this task delivers is the executable rehearsal itself, proven to work
and proven to be able to fail, so the remaining step is one operator run
rather than an unwritten procedure:

| `TEST-03` clause | Status |
| --- | --- |
| Toggle each of the four switches against the production target and confirm the documented effect | **Not done.** Procedure implemented as `infra/scripts/rehearse-production-killswitch-rollback.sh`; requires host access. |
| Confirm unrelated features are unaffected by each toggle | **Not done against production.** Implemented as a cross-check in every scenario; exercised locally (see below). |
| Deploy a disposable prior artifact by digest and confirm rollback with no unintended data loss | **Not done.** Implemented as the rehearsal's `[S5]` phase; requires a previously published `sha-*` tag and host access. |
| Health check passes after rollback | **Not done.** Implemented; requires host access. |

`VOC-037-T05` must read this section directly rather than infer AC-03's
status from a green deploy or from this task's merge.

## What was implemented

| Deliverable | Purpose |
| --- | --- |
| `infra/scripts/lib/killswitch-probes.sh` | The probes and assertions both runners share, so "this switch is observably off" means one thing. |
| `infra/scripts/rehearse-production-killswitch-rollback.sh` | The host procedure: five switch scenarios plus the rollback/roll-forward phase, against the real production target. |
| `infra/scripts/rehearse-production-killswitch-rollback.selftest.sh` | Runs the same probes against a disposable Postgres and a locally built api binary, and proves each assertion fails when the observed state is wrong. |

Each switch is observed through externally visible API behavior, never by
reading the value back out of `api.env` — that would only prove the file was
edited, which is not what `AC-03` asks for:

| Switch | Off | On |
| --- | --- | --- |
| `EMAIL_MAGIC_LINK_ENABLED` | `POST /api/v1/auth/magic-links` → `503` | → `204` |
| `GOOGLE_OAUTH_ENABLED` | `POST /api/v1/auth/oauth/google/start` → `503` | → any non-`503` (`404` while no Google client is configured for production, `200` once one is) |
| `NEW_USER_SIGNUP_ENABLED` | consuming a seeded magic link for a never-seen address → `503`, while a returning identity still gets `200` | new address → `200` |
| `AI_FEATURES_ENABLED` | `POST /api/v1/sentence-feedback` → `200` with `errorCode=AI_FEEDBACK_GENERATION_DISABLED` | → `404` (the generation gate is checked before the attempt is resolved, so the probe never reaches the AI provider) |

Two constraints shaped those probes. Production has no email provider (the
magic-link switch is `false` by deploy default, `VOC-032-DEP-07`'s
production-tier equivalent is unresolved), so no probe can wait for a
delivered email; the rehearsal seeds a `magic_links` row directly, using the
same `sha256(raw token bytes)` storage the API expects
(`apps/api/business/auth/tokens.go`). And the AI probe deliberately submits an
unowned attempt id, so a passing gate returns `404` without ever spending a
provider call.

## Deterministic checks run at this revision

```bash
shellcheck infra/scripts/rehearse-production-killswitch-rollback.sh \
           infra/scripts/rehearse-production-killswitch-rollback.selftest.sh \
           infra/scripts/lib/killswitch-probes.sh
bash infra/scripts/rehearse-production-killswitch-rollback.selftest.sh
```

`shellcheck` is clean. The self-test's output:

```
[L] assertion-library guards
  ok: an inconclusive observation is a failure, not a pass
  ok: ks_expect_not rejects an inconclusive observation
  ok: a run with zero checks reports FAIL
[H] host-only rehearsal steps are present in the host script
  ok: host script: recreates containers instead of restarting them
  ok: host script: restores api.env on every exit path
  ok: host script: pins the deployed image tag on every compose call
  ok: host script: rehearses rollback to the operator-named tag
  ok: host script: restores the originally running tag
  ok: host script: deletes the disposable identities it created
  ok: host script: never touches staging's tree
[E] disposable environment
  ok: disposable Postgres, migrations, and api binary are ready
[S0] baseline - all four switches on
  ok: magic-link request accepted: 204
  ok: oauth start not disabled: 200 (not 503)
  ok: new sign-up accepted: 200
  ok: ai generation gate open: 404/
[S1] EMAIL_MAGIC_LINK_ENABLED=false
  ok: magic-link request disabled: 503
  ok: negative proof: 'magic-link request disabled' fails when the observed state is wrong
  ok: oauth start unaffected: 200 (not 503)
  ok: ai feedback unaffected: 404/
[S2] GOOGLE_OAUTH_ENABLED=false
  ok: oauth start disabled: 503
  ok: negative proof: 'oauth start disabled' fails when the observed state is wrong
  ok: magic-link request unaffected: 204
[S3] NEW_USER_SIGNUP_ENABLED=false
  ok: new sign-up refused: 503
  ok: negative proof: 'new sign-up refused' fails when the observed state is wrong
  ok: returning sign-in still allowed: 200
[S4] AI_FEATURES_ENABLED=false
  ok: ai generation gate closed: 200/AI_FEEDBACK_GENERATION_DISABLED
  ok: negative proof: 'ai generation gate closed' fails when the observed state is wrong
  ok: magic-link request unaffected: 204
[S5] artifact swap preserves data (local stand-in for the host tag rollback)
  ok: disposable identity exists before the swap: 1 (not 0)
  ok: healthz after swapping the running artifact: 200
  ok: no data loss across the swap: 1
  ok: healthz after swapping back: 200
  ok: no data loss across the swap back: 1
[C] cleanup removes every row the rehearsal created
  ok: no rehearsal identity survives cleanup: 0

PASS: all 34 kill-switch rehearsal self-test check(s) succeeded
```

The self-test's own limits, stated rather than glossed: it runs the real
`cmd/api` binary against a real Postgres with the real migration set, so the
switch behavior it observes is genuine, but the environment is a local
process pair, not the production Compose stack. `[S5]` swaps the running
api artifact and shows the database survives it; it does not pull a
`ghcr.io` image by tag, which is what the host script's rollback phase
actually does and what `TEST-03` requires.

## Operator runbook (the remaining step for AC-03)

On the production host, as the production deploy user, with the immutable
tag of a previously published build:

```sh
# The script and its lib/ directory are not in the deploy bundle (adding
# them means editing deploy-production.yml, which is its own protected
# change), so copy both onto the host from a checkout of this revision:
scp -r infra/scripts/rehearse-production-killswitch-rollback.sh \
       infra/scripts/lib <host>:/tmp/t03/
bash /tmp/t03/rehearse-production-killswitch-rollback.sh sha-<previous>
```

Before running it, note what it does to a live tier:

- It briefly turns every switch **on**, including `EMAIL_MAGIC_LINK_ENABLED`,
  `GOOGLE_OAUTH_ENABLED`, and `NEW_USER_SIGNUP_ENABLED`, which the deploy
  workflow writes as `false`. Run it inside a maintenance window, before
  production carries real users; the trap restores `api.env` and the running
  tag on every exit path, and the next deploy rewrites those lines anyway.
- It creates and deletes `t03-rehearsal-*@rehearsal.invalid` identities and
  their magic-link and session rows, and nothing else. The final cleanup line
  reports how many rehearsal identities remain; anything other than `0`
  should be investigated before the run is recorded as evidence.
- It never calls the AI provider, sends email, or prints a secret.

Record the full output, the run timestamp, the rollback tag used, and the
`healthz` responses in this document, then mark `VOC-037-AC-03` satisfied.
Note that the workflow-level path to a rollback is currently "re-run
`deploy-production` at the older commit", which rebuilds rather than
redeploying the existing artifact; the rehearsal exercises the direct
`PRODUCTION_IMAGE_TAG` redeploy the compose file already supports, which is
the artifact-level rollback DOC-11 §3 describes.

## Follow-up notes (not fixed here — outside this task's scope)

1. **The session cookie's domain does not cover the API host.**
   `deploy-production.yml` sets `SESSION_COOKIE_DOMAIN` to the *web* host
   (`production.vocanova.site`), while the browser calls the API at
   `api-production.vocanova.site:8443`
   (`NEXT_PUBLIC_API_BASE_URL`, used by `apps/web/src/lib/env.ts` in the
   browser branch with `credentials: "include"`). A `Set-Cookie` whose
   `Domain` is not a suffix of the responding host is rejected by browsers,
   and `api-production.vocanova.site` is not a subdomain of
   `production.vocanova.site`. Server-side calls through the web container
   (`API_BASE_URL=http://api:8080`) are unaffected. This is an inspection
   finding, not a live observation — no browser session has been exercised
   against production — but it should be settled before `T05`, and it is a
   configuration/domain-shape question rather than a kill-switch defect.
2. **No workflow-level rollback entry point exists.** Rolling back today
   means re-running `deploy-production` at an older ref, which rebuilds the
   images. A small workflow input that redeploys an existing
   `PRODUCTION_IMAGE_TAG` without rebuilding would match DOC-11 §3 more
   directly; it touches `.github/workflows/`, so it belongs in its own
   scoped change rather than here.
3. **The rehearsal is not in the deploy bundle.** The secrets-boundary
   rehearsal is copied to the host by `deploy-production.yml`; this one is
   not, for the same workflow-protection reason, so the operator copies it
   manually (see the runbook above). Bundling it belongs with the change that
   adds the rollback entry point.
