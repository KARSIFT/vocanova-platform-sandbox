# VOC-095 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02**.

## VOC-095-T00 — Add bounded Playwright Chromium install script and foundation tests

- Requirement source: issue #792; `VOC-095-D00`–`D04`, `D07`
- Acceptance criteria: `VOC-095-AC-00`, `VOC-095-AC-01`
- Tests: `VOC-095-TEST-00` through `VOC-095-TEST-03`
- Evidence: `VOC-095-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Read run 32299315180 public metadata at implementation time. Record in
   `t00-evidence.md`: workflow, attempts, head SHA `fdde12daeccf521a5a8be171294eba79a138717b`,
   failure phase (Playwright install / apt wait), and that deploy convergence completed
   before timeout. Do not copy secrets, SSH output, session values, or personal data.
2. Create `infra/scripts/install-playwright-chromium.sh` implementing:
   - `set -euo pipefail` and repository-relative invocation of
     `pnpm --filter @vocanova/web exec playwright …`.
   - Browser install (`playwright install chromium`) separate from deps install
     (`playwright install-deps chromium` or equivalent bounded apt path).
   - Named constants for max attempts, per-attempt timeout, and inter-attempt sleep
     (`VOC-095-D03`; document chosen values in script header and evidence).
   - Post-install verification that a `chromium-*/chrome-linux/chrome` binary exists under
     `$HOME/.cache/ms-playwright`.
   - Non-zero exit and stderr diagnostics when verification or deps install fails after
     retries; no fallback that skips deps or pretends success.
3. Add `scripts/foundation/voc095-playwright-install.test.mjs` covering:
   - Script file exists and is referenced by name (workflows wired in T01).
   - Timeout/retry constants are present and within documented bounds.
   - Failure-fixture or static analysis proving the script does not emit skip/bypass
     patterns (`continue-on-error`, `PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD` without verify, etc.).
   - Optional: subprocess test with a mocked failing deps command if practical without
     root/apt in CI unit tests.
4. Run `node --test scripts/foundation/voc095-playwright-install.test.mjs` and applicable
   governance validation for changed paths.

### Explicitly out of scope for this task

- Workflow YAML edits (T01).
- Live deploy-staging verification (T02).
- Apt archive caching (open question 3 — only if adopting human approves in task PR).

## VOC-095-T01 — Wire all four workflows and monitoring validator to the shared contract

- Requirement source: `VOC-095-D02`, `D05`, `D06`; `VOC-095-DEP-02`
- Acceptance criteria: `VOC-095-AC-02` through `VOC-095-AC-04`, `VOC-095-AC-06`
- Tests: `VOC-095-TEST-04` through `VOC-095-TEST-12`, regression `VOC-086-TEST-11`
- Evidence: `VOC-095-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-095-T00`

### Required work

1. In each of `.github/workflows/accessibility.yml`, `.github/workflows/lighthouse.yml`,
   `.github/workflows/deploy-staging.yml` (core-loop install step), and
   `.github/workflows/scheduled-synthetics.yml`:
   - Add or preserve `actions/cache@d4323d4df104b026a6aa633fdb11d772146be0bf` restoring
     `~/.cache/ms-playwright` with key
     `playwright-chromium-${{ runner.os }}-${{ hashFiles('pnpm-lock.yaml') }}` and the
     same restore-keys prefix as scheduled-synthetics.
   - Place cache restore **before** workspace install completes only where consistent with
     existing pnpm cache ordering; cache restore must precede the install script step.
   - Replace inline `playwright install --with-deps chromium` with
     `bash infra/scripts/install-playwright-chromium.sh` (or `chmod +x` + direct path).
2. Preserve lighthouse.yml `LIGHTHOUSE_CHROME_PATH` export inside the Lighthouse `run:`
   step; do not move glob expansion into `env:` (see `.karsift/lessons.md`).
3. Update `infra/monitoring/scheduled-synthetics.mjs` validator to assert:
   - Cache step precedes install script for staging authenticated core-journey.
   - Install step invokes `install-playwright-chromium.sh` (not the removed inline command).
4. Extend `scripts/foundation/voc095-playwright-install.test.mjs` (or
   `voc086-scheduled-synthetics.test.mjs` minimally) to assert all four workflows meet
   AC-02 wiring.
5. Update `apps/web/tests/e2e/README.md` CI install section to document the shared script,
   cache behavior, and bounded deps contract.
6. Run full applicable foundation tests:
   `node --test scripts/foundation/voc095-playwright-install.test.mjs`,
   `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs`, and deploy/accessibility
   regression tests unchanged by semantics.

### Explicitly out of scope for this task

- Changing Playwright spec files or test thresholds.
- Live staging proof (T02).
- Weakening workflow timeouts or making browser jobs optional.

## VOC-095-T02 — Record live staging core-loop verification after deploy-staging merge

- Requirement source: issue #792 required outcome; `VOC-095-D00`
- Acceptance criteria: `VOC-095-AC-05`
- Tests: `VOC-095-TEST-11`
- Evidence: `VOC-095-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-095-T01`

### Required work

1. After T01 merges to `develop`, identify the latest `deploy-staging` run for the
   integration head SHA. Record scrubbed run URL, conclusion `success`, and confirmation
   that the staging core-loop Playwright step ran (step name visible in run UI; no job
   timeout during install).
2. If the run log shows cache hit/miss and install duration, record scrubbed timing only
   (no session cookies, minted tokens, or SSH output).
3. Note whether issue #792 can close under normal roster closure after verification.
4. If live mirror stall cannot be reproduced, document open question 4 fallback explicitly
   in `t02-evidence.md` (green deploy + deterministic bounded-failure proof from T00).

### Explicitly out of scope for this task

- Code changes (T00/T01 own all script and workflow edits).
- Manual issue closure outside the governed roster path.

## Task ordering notes

- T00 blocks T01: workflows must invoke a committed script.
- T01 blocks T02: live proof requires merged workflow wiring on `develop`.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
