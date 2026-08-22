# VOC-110-T00 evidence — drafting-time run metadata

Recorded at **drafting time** from the public GitHub Actions REST API and run
page (no authentication required for run-level metadata). T00 must extend this
file with log-derived failing-step detail at implementation time.

## gate_status

pending — failing workflow step not yet identified from job logs

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

## Drafting-time classification bounds

| Class | Ruled in/out | Rationale |
|-------|----------------|-----------|
| VOC-094 concurrency supersession | **Out** | Conclusion `failure`, job ran ~8m (not `cancelled` ~2m with zero jobs) |
| VOC-095 Playwright install timeout | **Unlikely** | Duration well below 40m job timeout; install script bounded since VOC-095 |
| Actionable deploy step failure | **Leading** | Exit code 1 after substantial job runtime; Docker build artifacts produced |
| Exact failing step | **Open** | Requires authenticated job log access in T00 |

## T00 completion checklist (implementer)

- [ ] Record failing workflow step name and sanitized failure class from job logs
- [ ] Document causal link (or lack thereof) to Dependabot PR #859 package bumps
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
