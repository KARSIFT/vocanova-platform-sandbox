# VOC-095 — Test Plan

## VOC-095-TEST-00 — Root-cause evidence references run 32299315180 and apt stall phase

- Covers: `VOC-095-AC-00`
- Preconditions: T00 evidence file drafted
- Procedure: Read `t00-evidence.md`; assert it names run 32299315180, head SHA
  `fdde12daeccf521a5a8be171294eba79a138717b`, and Playwright `--with-deps` / apt wait
  as the timeout phase after deploy convergence.
- Expected result: Evidence bounds remediation to install boundedness + cache, not deploy logic
- Evidence: `VOC-095-EV-00`

## VOC-095-TEST-01 — Install script declares timeout and retry constants

- Covers: `VOC-095-AC-01`
- Preconditions: T00 script committed
- Procedure: Read `infra/scripts/install-playwright-chromium.sh`; assert named constants
  for max attempts, per-attempt timeout, and retry sleep; assert `set -euo pipefail`.
- Expected result: Bounded contract is explicit and auditable
- Evidence: `VOC-095-EV-00`

## VOC-095-TEST-02 — Install script verifies Chromium binary after install

- Covers: `VOC-095-AC-01`
- Preconditions: T00 script committed
- Procedure: Static read or fixture run asserts post-install check for
  `$HOME/.cache/ms-playwright/chromium-*/chrome-linux/chrome`.
- Expected result: Script fails closed when binary missing
- Evidence: `VOC-095-EV-00`

## VOC-095-TEST-03 — Install script has no skip or bypass patterns

- Covers: `VOC-095-AC-01`
- Preconditions: T00 foundation test committed
- Procedure: Run `node --test scripts/foundation/voc095-playwright-install.test.mjs`;
  assert denylist patterns absent (`continue-on-error`, unconditional skip env vars, etc.).
- Expected result: Fail-closed semantics locked by test
- Evidence: `VOC-095-EV-00`

## VOC-095-TEST-04 — accessibility.yml uses cache then install script

- Covers: `VOC-095-AC-02`
- Preconditions: T01 workflow changes
- Procedure: Parse workflow; assert `~/.cache/ms-playwright` cache step precedes
  `install-playwright-chromium.sh`; assert inline `--with-deps` one-liner absent.
- Expected result: Accessibility job uses shared contract
- Evidence: `VOC-095-EV-01`

## VOC-095-TEST-05 — lighthouse.yml uses cache then install script

- Covers: `VOC-095-AC-02`
- Preconditions: T01 workflow changes
- Procedure: Same as TEST-04 for lighthouse.yml.
- Expected result: Lighthouse job uses shared contract
- Evidence: `VOC-095-EV-01`

## VOC-095-TEST-06 — deploy-staging.yml core-loop uses cache then install script

- Covers: `VOC-095-AC-02`
- Preconditions: T01 workflow changes
- Procedure: Parse deploy-staging core-loop job; assert cache + script wiring; assert
  `timeout-minutes: 40` unchanged.
- Expected result: Staging deploy core-loop install hardened without timeout reduction
- Evidence: `VOC-095-EV-01`

## VOC-095-TEST-07 — scheduled-synthetics.yml uses cache then install script

- Covers: `VOC-095-AC-02`
- Preconditions: T01 workflow changes
- Procedure: Parse staging core-loop synthetic job; assert cache precedes script.
- Expected result: Scheduled core-loop retains cache and adopts script
- Evidence: `VOC-095-EV-01`

## VOC-095-TEST-08 — Browser test steps remain mandatory after install step

- Covers: `VOC-095-AC-03`
- Preconditions: T01 workflow changes
- Procedure: Assert accessibility, lighthouse, deploy-staging core-loop, and
  scheduled-synthetics test steps remain without `continue-on-error: true` on install
  or test steps.
- Expected result: Fail-closed browser gates preserved
- Evidence: `VOC-095-EV-01`

## VOC-095-TEST-09 — Lighthouse chrome path resolution stays shell-expanded

- Covers: `VOC-095-AC-03`
- Preconditions: T01 lighthouse.yml changes
- Procedure: Read Lighthouse `run:` step; assert `LIGHTHOUSE_CHROME_PATH` assignment uses
  `export … $(ls -d …)` inside `run:`, not unexpanded `env:` glob.
- Expected result: Lighthouse reuse of Playwright binary unchanged in mechanism
- Evidence: `VOC-095-EV-01`

## VOC-095-TEST-10 — scheduled-synthetics.mjs validates script + cache ordering

- Covers: `VOC-095-AC-04`
- Preconditions: T01 validator changes
- Procedure: Run `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs`
  and VOC-095 workflow wiring tests.
- Expected result: Validator passes with new install contract
- Evidence: `VOC-095-EV-01`

## VOC-095-TEST-11 — Live deploy-staging success with core-loop after T01 merge

- Covers: `VOC-095-AC-05`
- Preconditions: T01 merged to `develop`
- Procedure: Inspect latest green `deploy-staging` run; confirm core-loop step completed;
  record run URL in `t02-evidence.md`.
- Expected result: Staging core-loop no longer cancelled at timeout during install
- Evidence: `VOC-095-EV-02`

## VOC-095-TEST-12 — e2e README documents shared install contract

- Covers: `VOC-095-AC-06`
- Preconditions: T01 documentation update
- Procedure: Read `apps/web/tests/e2e/README.md`; assert reference to
  `install-playwright-chromium.sh` and bounded deps behavior.
- Expected result: Operator docs match CI contract
- Evidence: `VOC-095-EV-01`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.

## Regression tests (existing, must remain green)

- `VOC-086-TEST-11` — scheduled synthetics registry/workflow wiring
- `voc084-deploy-staging-oauth.test.mjs`, `voc088-deploy-staging-allowlist.test.mjs`,
  `voc081-deploy-convergence.test.mjs` — deploy semantics unchanged aside from install step
- Accessibility/lighthouse foundation tests if present and path-triggered
