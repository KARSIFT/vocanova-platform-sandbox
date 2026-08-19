# VOC-093-EV-00 — T00 route-sweep diagnosis and harness fix

Evidence for `VOC-093-T00` (`VOC-093-AC-00` through `VOC-093-AC-03`).

## Failure source (run 32288703894)

| Item | Value |
|------|-------|
| Workflow | `scheduled-synthetics` |
| Run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32288703894 |
| Head SHA | `7851d570f1dc245a3b310b9705b54b10585c9ea0` (`main`) |
| Failing job | `synthetic.production.authenticated-route-content-sweep` |
| Failing step | Run production authenticated route/content sweep synthetic (~5s, exit 1) |
| Sibling | `synthetic.production.journey-content` **success** in the same run |

Job logs are not available from this implementation environment (GitHub Actions
logs API returned 403 without repository admin credentials). Diagnosis below
combines public run metadata, sibling-job bounding, middleware behavior, and
the VOC-085 synthetic-account contract.

## Scope bounding (sibling `journey-content` success)

The route-sweep job uses `SMOKE_TEST_PROFILE=route-sweep`, which runs sections
**# 1** (healthz/kill switches), **# 5** (journey-situations API content), and
**# 6** (authenticated web route sweep). Sibling `journey-content` exercises
sections **# 1** and **# 5** only and passed, so session mint, kill switches,
`/api/v1/me`, and journey API slug derivation were healthy when the route sweep
ran. The failure is confined to section **# 6** web GET checks or harness logic
executed only in that section.

## Root cause (harness false negative)

**Chosen fix path:** harness (`infra/scripts/smoke-test-production.sh`), not
`apps/web/`.

The deploy-seeded synthetic account keeps `onboarding_status='completed'`
(VOC-085). For that account, `apps/web/src/middleware.ts` redirects
`/onboarding` to `/home` when onboarding is already complete — an expected,
in-app redirect that VOC-085-T02 evidence documents as acceptable.

Next.js emits an **absolute** same-origin `Location` header for that redirect
(for example `https://<production-host>/home`), confirmed locally via the
VOC-039 middleware runtime harness:

```text
{"kind":"redirect","location":"https://app.vocanova.test/home"}
```

Section **# 6** lists `/onboarding` among the ten fixed authenticated routes.
The route-sweep checker (`assert_web_route_reachable`) followed relative
redirects only; absolute same-origin locations were rejected with:

```text
FAIL: GET <web>/onboarding (authenticated) returned a non-relative redirect (Location: https://…/home)
```

That false negative explains exit 1 on the first failing authenticated fixed
route without implicating production rendering, session validity, or journey
content.

## Fix

- Added `coerce_same_origin_redirect_path` to normalize absolute `Location`
  values on `web_base_url` to a path before traversal.
- Preserved fail-closed behavior for sign-in, protocol-relative, and external
  redirects.
- Updated the disposable selftest fake server to emit absolute same-origin
  `/onboarding` → `/home` redirects (matching production) and added a regression
  case for absolute same-origin sign-in rejection.

## Deterministic validation

Implementation run (2026-08-19):

| Command | Result |
|---------|--------|
| `bash infra/scripts/smoke-test-production.selftest.sh` | pass (14 cases) |
| `node --test scripts/foundation/voc085-production-route-sweep.test.mjs` | pass (3 tests) |
| `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs` | pass (5 tests) |
| `bash scripts/governance/validate-governance.sh` | pass |
| `git diff --check` | pass |