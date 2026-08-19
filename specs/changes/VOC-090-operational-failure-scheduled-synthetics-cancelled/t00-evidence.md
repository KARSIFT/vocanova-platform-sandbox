# VOC-090-T00 — Root-cause evidence

## Failing run

| Item | Value |
|------|-------|
| Workflow | `scheduled-synthetics` |
| Run | [#22](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32271016931) (`32271016931`) |
| Trigger | `schedule` (2026-08-19 15:35 UTC) |
| Conclusion | `cancelled` (workflow wall clock ~30m 19s) |
| Failing job | `synthetic.staging.authenticated-core-journey` |
| Public annotation | exceeded maximum execution time **30m0s** |

Other jobs in the same run (`synthetic.staging.oauth-expected-state`,
`synthetic.production.oauth-expected-state`, `synthetic.production.journey-content`,
`synthetic.production.authenticated-route-content-sweep`) completed without a
timeout annotation on the public run page.

## Phase breakdown (job `96127180388`, public GitHub Actions API)

Step timings were read from the public Actions jobs API at implementation time
(no secrets, session values, or personal data in the response).

| Step | Duration | Conclusion |
|------|----------|------------|
| Check out reviewed revision | 2s | success |
| Refresh reserved staging synthetic review state | 3s | success |
| Mint synthetic smoke-test session for staging core-loop | &lt;1s | success |
| Set up Node for the staging core-loop synthetic | 4s | success |
| Install pnpm for the staging core-loop synthetic | 2s | success |
| Install workspace dependencies (`pnpm install --frozen-lockfile`) | 7s | success |
| Install Playwright Chromium (`playwright install --with-deps chromium`) | **1795s (~29m 55s)** | **cancelled** |
| Run staging authenticated core-loop synthetic | 0s | skipped |

**Dominant budget consumer:** uncached Playwright Chromium browser install.
The Playwright journey never started; this is a CI setup wall-clock failure,
not evidence of `playwright.staging.config.ts`'s 240s per-test timeout being
exceeded (`VOC-090-D03` follow-up is not triggered).

## Remediation applied in T00

1. **Playwright browser caching** — `actions/cache` on `~/.cache/ms-playwright`
   keyed by `pnpm-lock.yaml`, restored before `playwright install`.
2. **pnpm dependency caching** — `actions/setup-node` `cache: "pnpm"` before
   `pnpm install --frozen-lockfile`.
3. **Timeout alignment** — job `timeout-minutes` raised from 30 to **40**
   (parity with `deploy-staging.yml`'s proven post-deploy core-loop gate) and
   registry `synthetic.staging.authenticated-core-journey.timeout_seconds`
   updated to **2400** (job wall clock, not Playwright per-test timeout).

## Setup reserve documented for budget tests

`STAGING_CORE_JOURNEY_SETUP_RESERVE_SECONDS = 960` (16 minutes):

- SSH seed `command_timeout`: 120s
- Session mint: 60s conservative ceiling
- Warm-cache dependency/browser install ceiling: 780s conservative ceiling

With Playwright journey timeout 240s, minimum required job wall clock is
1200s (20 minutes). The chosen 2400s (40 minutes) matches deploy-staging and
leaves margin for `MAX_REVIEW_CARDS = 8` review-loop headroom.
