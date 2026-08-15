# VOC-085-EV-02 — T02 authenticated route sweep and live production proof

Evidence for `VOC-085-T02` (`VOC-085-TEST-06` through `VOC-085-TEST-10`).

## Harness shape (`VOC-085-DEP-02`)

Extended `infra/scripts/smoke-test-production.sh` with section **# 6**
(authenticated learning-route sweep) rather than a sibling script. Rationale:

- `deploy-production.yml` already invokes this suite after session mint
  (`VOC-085-AC-05`, `VOC-085-D05`).
- Section **# 5** already derives `situation_slug` and `word_slug` from live
  API responses; section **# 6** reuses those identifiers for dynamic discover
  routes instead of hard-coding slugs.
- Curl-level authenticated GETs are sufficient for reachability gating; no
  browser harness is required for this package's read-only AC.

## Implemented route sweep

File changed: `infra/scripts/smoke-test-production.sh`

- **Ten fixed routes** via `PRODUCTION_PUBLIC_WEB_ROUTES` and
  `PRODUCTION_AUTHENTICATED_WEB_ROUTES` (`VOC-085-TEST-06`):
  `/`, `/signin`, `/auth/magic`, `/onboarding`, `/home`, `/discover`,
  `/reviews`, `/progress`, `/settings`, `/settings/account`.
- **Dynamic routes** from section **# 5** API slugs:
  `/discover/{situation_slug}` and `/discover/{situation_slug}/{word_slug}`.
- Reuses `SMOKE_TEST_SESSION_COOKIE` from the workflow mint step; public routes
  are requested without a cookie; protected routes require it.
- Fails closed when the cookie is missing (deploy path always supplies one).
- Protected-route redirects to `/signin` fail the suite; in-app redirects (for
  example completed onboarding `/onboarding` → `/home`) remain acceptable.
- Non-mutating GET only — no magic links, OAuth completion, or learning writes.

## Deterministic selftests

File changed: `infra/scripts/smoke-test-production.selftest.sh`

- Extended the disposable fake server with public/authenticated web routes and
  API-derived discover paths.
- Case 1 / 4: healthy passes now include route-sweep coverage with a valid
  cookie.
- `VOC-085-TEST-07` fixtures:
  - case 9: missing cookie → suite fails (not silent skip).
  - case 10: malformed cookie → authenticated route coverage fails.
  - case 11: protected route HTTP 500 → route sweep fails.

## Foundation tests

File added: `scripts/foundation/voc085-production-route-sweep.test.mjs`

- `VOC-085-TEST-06`: exact ten-route inventory + API-derived dynamic paths.
- `VOC-085-TEST-07`: fail-closed missing-cookie behavior and sign-in redirect
  rejection on protected routes.
- `VOC-085-TEST-10`: deploy still wires minted cookie into
  `smoke-test-production.sh` on canonical `:443`; shared-edge / no `8081`/`8443`
  invariants remain covered by `voc079-single-edge-invariants.test.mjs`.

## Acceptance mapping

| AC / decision | How this task satisfies it |
|---------------|----------------------------|
| `VOC-085-AC-05` | Section **# 6** covers ten fixed + two API-derived discover routes with minted session |
| `VOC-085-AC-06` | No manual DB edits; synthetic onboarding remains seed-owned (`voc085-production-p1-seed.test.mjs` TEST-08) |
| `VOC-085-AC-07` (route tests) | Selftest cases 9–13 + foundation TEST-06/07 |
| `VOC-085-AC-08` | No topology change; foundation `voc079-*` guards + deploy shared-edge reload only |
| `VOC-085-D05` | Recorded harness shape above; read-only GET sweep |
| `VOC-085-D07` | Deploy workflow unchanged except invoking strengthened smoke suite |

## Local deterministic checks run for this task

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
bash infra/scripts/smoke-test-production.selftest.sh
node --test scripts/foundation/voc085-production-route-sweep.test.mjs
```

Record results from the implementation run in the task PR (pass/fail per command).

Implementation run (2026-08-15):

| Command | Result |
|---------|--------|
| `bash scripts/governance/validate-governance.sh` | pass |
| `bash scripts/governance/classify-change-risk.sh --files-from …` | pass (floor R3 via smoke scripts) |
| `git diff --check` | pass |
| `bash infra/scripts/smoke-test-production.selftest.sh` | pass (13 cases after redirect-traversal follow-up) |
| `node --test scripts/foundation/voc085-production-route-sweep.test.mjs` | pass (3 tests) |

## Live Cloudflare verification (`VOC-085-TEST-09`)

Initial protected production deploy:
https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31913634189

- Seed reported 7 situations, 39 words, 42 meanings, 42 examples, 126 notes,
  and 42 journey-word links.
- Authenticated smoke returned 200 for `/api/v1/me`, a non-empty list of 7
  journey situations, real `airport` situation detail with meanings, and real
  `boarding-pass` word detail.
- Read-only host verification matched those canonical counts and confirmed the
  synthetic account was active, verified, marked synthetic, and onboarding
  completed.
- The initial route sweep reached all ten fixed and both API-derived dynamic
  routes, but seven protected routes returned internal 307 responses. Although
  the checker rejected sign-in redirects, it did not traverse other redirects
  to a final rendered response. Issue #706 was reopened rather than claiming
  that weaker evidence complete.

Follow-up behavior in this revision:

- Traverse at most five relative same-origin redirects and require a final
  2xx response.
- Reject sign-in, missing-Location, protocol-relative/external, and looping
  redirects before an explicit session cookie can leave the production origin.
- Final post-follow-up deploy and stable-count/idempotency evidence is recorded
  in issue #706's closure comment so the evidence can refer to the deployed
  revision without creating an endless evidence-only redeploy cycle.

## Topology / isolation confirmation (`VOC-085-TEST-10`)

Repository-side evidence (no deploy topology change in this task):

- Production deploy continues `docker exec vocanova-shared-edge-nginx nginx -t`
  / reload only — no shared-edge recreation (`voc079-single-edge-invariants.test.mjs`).
- Production compose still excludes per-tier nginx and `8081`/`8443` publish.
- Staging/production secret/directory isolation unchanged (no workflow edits beyond
  the existing smoke invocation).

Initial live host verification found exactly one healthy shared-edge nginx, no
production nginx, no listeners on 8081/8443, and healthy staging, production,
and monitoring containers. The post-follow-up confirmation is recorded with
the final deploy in issue #706.
