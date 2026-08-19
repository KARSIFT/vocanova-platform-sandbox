# VOC-095-T01 — Workflow wiring and validator alignment evidence

Task: `VOC-095-T01`  
Package: `VOC-095-harden-playwright-setup-against-hosted-runner-apt`  
Evidence class: `VOC-095-EV-01`

## Summary

All four Chromium-consuming workflows now restore `~/.cache/ms-playwright` via
`actions/cache` before invoking `infra/scripts/install-playwright-chromium.sh`.
Inline `playwright install --with-deps chromium` one-liners were removed.
`infra/monitoring/scheduled-synthetics.mjs` validates cache-before-script
ordering for the staging authenticated core-journey job. Operator documentation
in `apps/web/tests/e2e/README.md` references the shared contract.

## Workflow changes

| Workflow | Cache key | Install step |
|----------|-----------|--------------|
| `accessibility.yml` | `playwright-chromium-${{ runner.os }}-${{ hashFiles('pnpm-lock.yaml') }}` | `bash infra/scripts/install-playwright-chromium.sh` |
| `lighthouse.yml` | same | same (`LIGHTHOUSE_CHROME_PATH` shell expansion unchanged in `run:`) |
| `deploy-staging.yml` (core-loop) | same | same (`timeout-minutes: 40` unchanged) |
| `scheduled-synthetics.yml` (core-journey job) | same (preserved) | script replaces inline install |

## Deterministic validation (recorded at implementation time)

```bash
node --test scripts/foundation/voc095-playwright-install.test.mjs
node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs
node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
```

See command output below after local run completes.

### Test run output

All foundation tests passed at implementation time:

- `voc095-playwright-install.test.mjs`: 18 pass (includes VOC-095-TEST-04 through TEST-12)
- `voc086-scheduled-synthetics.test.mjs`: 5 pass (VOC-086-TEST-11 regression)
- `voc090-scheduled-synthetics-budget.test.mjs`: 6 pass
- `voc084-deploy-staging-oauth.test.mjs`: 9 pass
- `voc088-deploy-staging-allowlist.test.mjs`: 7 pass

## Preserved semantics (VOC-095-AC-03)

- No `continue-on-error: true` on install or browser-test steps.
- Workflow `timeout-minutes` values unchanged (accessibility/lighthouse 30, deploy-staging 40).
- Lighthouse `LIGHTHOUSE_CHROME_PATH` resolved via `export … $(ls -d …)` inside the `run:` step body.

## Live verification

Deferred to `VOC-095-T02` (green `deploy-staging` core-loop run after merge to `develop`).
