# VOC-110-T00 evidence — drafting-time run metadata

Initially recorded from public Actions metadata, then refined with privacy-safe,
read-only workflow and runtime diagnosis. T00 must add implementation validation.

## gate_status

pending — root cause confirmed; implementation and regression-gate proof pending

## Run 32566405628 (deploy-staging #364)

| Field | Value |
|-------|-------|
| Run URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32566405628 |
| Workflow | `deploy-staging` (run number 364) |
| Event | `push` |
| Head branch | `develop` |
| Head SHA | `f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46` |
| Display title | Merge pull request #859 from KARSIFT/dependabot/npm_and_yarn/develop/… |
| Conclusion | `failure` |
| Status | `completed` |
| Created at | `2026-08-22T09:59:14Z` |
| Wall-clock duration | ~8m 28s (public run page) |
| Job | `deploy to staging` (~8m 23s, exit code 1 per public annotation) |

## Trigger context

- Dependabot PR **#859**: `dependabot/npm_and_yarn/develop/minor-and-patch` — eight
  npm minor/patch updates merged to `develop`.
- This is the first integration commit carrying that dependency group.

## Public annotations (sanitized)

- `deploy to staging`: Process completed with exit code 1.
- Failure artifact upload steps reported missing paths:
  `apps/web/playwright-report-staging/`, `apps/web/test-results-staging/`.
- Benign warning: Go cache restore — `go.mod` not found at repository root (expected;
  Go module lives under `apps/api/`).

## Confirmed failure chain (sanitized)

- The API health poll completed successfully.
- `Poll staging.vocanova.site/` failed after its five-minute budget; independent
  public probing returned HTTP 502.
- Read-only container state showed `vocanova-web` restarting while the staging API,
  production containers, shared edge, and monitoring remained isolated and healthy.
- The bounded runtime error was a missing `@swc/helpers` ESM module from the Next.js
  16.3.1 standalone artifact on Node 24. No full log or server transcript is retained.
- This matches upstream vercel/next.js issue #97358. Next.js 16.3.2 includes the
  16.3 backport from #97372/#97453.

## Drafting-time classification bounds

| Class | Ruled in/out | Rationale |
|-------|----------------|-----------|
| VOC-094 concurrency supersession | **Out** | Conclusion `failure`, job ran ~8m (not `cancelled` ~2m with zero jobs) |
| VOC-095 Playwright install timeout | **Out** | Web health failed before Playwright install/core-loop steps were reached |
| Actionable deploy step failure | **In** | Web health poll failed after image deploy |
| Exact failing step | **Confirmed** | `Poll staging.vocanova.site/` |
| Direct dependency regression | **Confirmed** | Next.js 16.3.1 standalone output omitted required runtime helper |
| Preventive CI gap | **Confirmed** | Source build passed; production image was never booted before merge |

## T00 completion checklist (implementer)

- [x] Record failing workflow step name and sanitized failure class
- [x] Document causal link to Dependabot PR #859 Next.js bump
- [ ] Apply fix and record changed files
- [ ] Run deterministic tests and record command results below
- [ ] Set `gate_status` to `complete` when AC-00 through AC-02 satisfied

## commands

_To be filled at implementation time._

## results

_To be filled at implementation time._

## privacy

Evidence must remain allowlisted metadata only (commands, pass/fail, SHAs, run
IDs, step names). No logs, artifacts, secrets, session values, or user identifiers.
