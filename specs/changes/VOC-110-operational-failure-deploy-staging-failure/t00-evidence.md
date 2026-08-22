# VOC-110-T00 evidence — root cause, fix, and validation

Recorded from public Actions metadata, read-only runtime diagnosis, and T00
implementation validation. No secrets, SSH transcripts, session cookies, OAuth
state, tokens, personal data, or complete application logs.

## gate_status

complete — AC-00 through AC-02 satisfied at implementation time (live deploy proof
remains T01)

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

## Fix summary (VOC-110-D04 / D05 / D06)

| Item | Value |
|------|-------|
| Causal link | Direct consequence of Dependabot PR #859 moving `next` 16.3.0 → 16.3.1 |
| Repair | Paired `next` and `@next/eslint-plugin-next` upgraded to stable **16.3.2** |
| Rollback (documented only) | Pin both packages back to **16.3.0** if 16.3.2 cannot pass image-runtime proof |
| Other PR #859 updates | Preserved (no bulk revert) |
| Changed files | `apps/web/package.json`, `pnpm-lock.yaml`, `.github/workflows/pipeline.yml`, `scripts/foundation/voc110-web-container-runtime.test.mjs`, `docs/operations/10-development-workflow.md` |
| Preventive gate | `pipeline.yml` job `web-container-runtime` — fail-closed for every repo-root Docker-context control, web/workspace source, and its own gate definitions; local Docker build/boot/HTTP 2xx; `merge-gate` requires its success |

## T00 completion checklist

- [x] Record failing workflow step name and sanitized failure class
- [x] Document causal link to Dependabot PR #859 Next.js bump
- [x] Apply fix and record changed files
- [x] Run deterministic tests and record command results below
- [x] Set `gate_status` to `complete` when AC-00 through AC-02 satisfied

## commands

```bash
node --test scripts/foundation/voc110-web-container-runtime.test.mjs
node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
node --test scripts/foundation/voc095-playwright-install.test.mjs
docker build -f apps/web/Dockerfile --build-arg NEXT_PUBLIC_API_BASE_URL=http://localhost:8080 -t vocanova-web:voc110 .
# start container, require HTTP 2xx from /, verify still running, remove container
pnpm validate
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## results

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc110-web-container-runtime.test.mjs` | pass (7/7) |
| `node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs` | pass |
| `node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs` | pass |
| `node --test scripts/foundation/voc095-playwright-install.test.mjs` | pass (36/36) |
| `docker build -f apps/web/Dockerfile -t vocanova-web:voc110 .` | pass |
| Local container boot + HTTP smoke (`/` returned HTTP 200, container stayed running) | pass |
| `pnpm validate` | pass (Node engine warning only: runner 24.5.0 vs `.nvmrc` 24.18.0) |
| `bash scripts/governance/validate-governance.sh` | pass |
| `bash scripts/governance/classify-change-risk.sh` | path floor R3 (`.github/workflows/pipeline.yml`) |
| `git diff --check` | pass |

## privacy

Evidence must remain allowlisted metadata only (commands, pass/fail, SHAs, run
IDs, step names). No logs, artifacts, secrets, session values, or user identifiers.
