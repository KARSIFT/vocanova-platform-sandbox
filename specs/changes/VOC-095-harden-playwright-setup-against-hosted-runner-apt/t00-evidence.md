# VOC-095-T00 — Root-cause evidence

Recorded at implementation time from the public GitHub Actions REST API
(no authentication required for this repository's public run metadata).

## Failing run

| Item                  | Value                                                                                        |
| --------------------- | -------------------------------------------------------------------------------------------- |
| Workflow              | `deploy-staging` (run number 307)                                                            |
| Run                   | [32299315180](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32299315180) |
| Trigger               | `push` to `develop`                                                                          |
| Head SHA              | `fdde12daeccf521a5a8be171294eba79a138717b`                                                   |
| Display title         | VOC-092: VOC-092-T01 (#790)                                                                  |
| Job                   | `deploy to staging` (`96229608876`)                                                          |
| Job `timeout-minutes` | 40                                                                                           |
| Attempt 1             | `cancelled` (`2026-08-19T20:35:04Z` → `2026-08-19T21:15:35Z`)                                |
| Attempt 2             | `cancelled` (`2026-08-19T21:16:21Z` → `2026-08-19T21:57:03Z`)                                |
| Run conclusion        | `cancelled` (workflow timeout on both attempts)                                              |

## Phase breakdown (job `96229608876`, public GitHub Actions jobs API)

Step timings were read from the public Actions jobs API at implementation time
(no secrets, session values, SSH output, or personal data in the response).

Deploy convergence completed before the Playwright install step stalled:

| Step                                                                    | Duration                                                   | Conclusion    |
| ----------------------------------------------------------------------- | ---------------------------------------------------------- | ------------- |
| Build and push vocanova-api / vocanova-web                              | success                                                    | success       |
| Copy deploy bundle to staging host                                      | 12s                                                        | success       |
| Deploy to staging host                                                  | 28s                                                        | success       |
| Poll api-staging.vocanova.site/healthz                                  | 1s                                                         | success       |
| Poll staging.vocanova.site/                                             | 1s                                                         | success       |
| Verify staging OAuth start initiation                                   | 1s                                                         | success       |
| Mint synthetic smoke-test session for the staging core-loop             | 1s                                                         | success       |
| Install workspace dependencies (`pnpm install --frozen-lockfile`)       | 7s                                                         | success       |
| Install Playwright Chromium (`playwright install --with-deps chromium`) | **2252s (~37m 32s)** on attempt 2 (**2254s** on attempt 1) | **cancelled** |
| Run the staging core-loop journey                                       | 0s                                                         | skipped       |

**Dominant budget consumer:** unbounded `playwright install --with-deps chromium`
waiting on the hosted Ubuntu package mirror during the `--with-deps` apt path.
The staging core-loop Playwright journey never started. This is a CI browser
install wall-clock failure after deploy convergence, not an SSH deploy defect,
migration failure, health-check failure, or core-loop test execution failure
(`VOC-095-D00`).

Related transient: accessibility workflow hit the same mirror stall on the same
integration era; an exact-SHA rerun succeeded (issue #792 drafting note — not
reproduced in this evidence file).

## Remediation applied in T00

1. **Shared bounded install script** — `infra/scripts/install-playwright-chromium.sh`
   separates `playwright install chromium` (CDN browser download) from
   `playwright install-deps chromium` (apt/system deps) with explicit retry limits.
2. **Chosen constants (VOC-095-D03):**
   - `PLAYWRIGHT_DEPS_MAX_ATTEMPTS=3`
   - `PLAYWRIGHT_DEPS_ATTEMPT_TIMEOUT_SECONDS=120`
   - `PLAYWRIGHT_DEPS_RETRY_SLEEP_SECONDS=30`
3. **Post-install verification** — script requires an executable
   `~/.cache/ms-playwright/chromium-*/chrome-linux/chrome` binary and exits
   non-zero when deps or verification fail (fail-closed; no skip flags).
4. **Foundation tests** — `scripts/foundation/voc095-playwright-install.test.mjs`
   locks script structure, constants, denylist semantics, and fixture-based
   bounded-failure behavior.
5. **Runner-only mirror recovery** — after a failed primary dependency attempt,
   the GitHub Linux X64 runner temporarily switches its apt mirror list to the
   verified `https://archive.ubuntu.com/ubuntu/` fallback. The original ephemeral
   runner configuration is restored on exit.
6. **Process and lock cleanup** — GNU `timeout` sends `TERM` to the dependency
   install process tree and escalates after a fixed grace period. Before any
   retry, the script requires all apt/dpkg locks to become idle or fails closed.

Workflow cache restore and inline-command replacement remain **T01** scope
(`VOC-095-D02`, `VOC-095-D05`).

## Deterministic validation (VOC-095-EV-00)

Recorded at implementation time:

```text
$ node --test scripts/foundation/voc095-playwright-install.test.mjs
✔ VOC-095-TEST-00: root-cause evidence references run 32299315180 and apt stall phase
✔ VOC-095-TEST-01: install script declares timeout and retry constants
✔ VOC-095-TEST-02: install script verifies Chromium binary after install
✔ VOC-095-TEST-03: install script has no skip or bypass patterns
✔ VOC-095-TEST-03b: install script fails closed when deps retries exhaust
✔ VOC-095-TEST-03c: install script succeeds when deps succeed and binary exists
✔ VOC-095-TEST-03d: install script fails closed when Chromium binary is missing
✔ VOC-095-TEST-03e: primary timeout activates HTTPS fallback and succeeds
✔ VOC-095-TEST-03f: primary and HTTPS fallback failure remains non-zero
✔ VOC-095-TEST-03g: busy package-manager locks block a poisoned retry
✔ VOC-095-TEST-01b: install script exists at the repository-managed path
ℹ tests 11
ℹ pass 11
ℹ fail 0
```
