---
evidence_id: VOC-079-EV-01
task_id: VOC-079-T01
acceptance_criteria: VOC-079-AC-04
tests: VOC-079-TEST-03
date: 2026-08-15
related_change: VOC-079
accountable_owner: implementer (repository task)
---

# VOC-079-T01 — Canonical production URLs on ordinary :443

## Summary

Repository operational configuration no longer qualifies production browser,
OAuth, readiness, smoke, or compose `API_BASE_URL` values with `:8443`.
The temporary `vocanova-production-nginx` cutover bridge service remains in
`infra/docker-compose.production.yml` until **VOC-079-T02** (bridge retirement
is explicitly out of scope for this task).

Historical VOC-041/VOC-042/VOC-067 evidence packages are unchanged; only
forward-looking operator docs (`infra/README.md`) and current deploy/compose
paths were updated.

## Files changed

| Path | Change |
| --- | --- |
| `infra/docker-compose.production.yml` | `API_BASE_URL` → `https://api-production.vocanova.site`; header notes VOC-079-T01 URL normalization while bridge remains |
| `.github/workflows/deploy-production.yml` | `BASE_URL` / OAuth allowlist without `:8443`; readiness polls, session-mint, and smoke on canonical `:443`; header/input comments aligned |
| `scripts/foundation/docker-compose-production-api-base-url-port.test.mjs` | Asserts canonical URL; rejects `:8443` (VOC-079-TEST-03) |
| `scripts/foundation/deploy-production-oauth-port.test.mjs` | Asserts unqualified OAuth/browser URLs; rejects `:8443` (VOC-079-TEST-03) |
| `infra/README.md` | Steady-state operator section documents canonical `:443` URLs; bridge retention gate unchanged until T02 |

## Deterministic validation (this revision)

```bash
node --test scripts/foundation/docker-compose-production-api-base-url-port.test.mjs \
  scripts/foundation/deploy-production-oauth-port.test.mjs
docker compose -f infra/docker-compose.production.yml config
```

Expected: foundation tests pass; production compose still defines
`vocanova-production-nginx` with `8443:443` publish (bridge retained for T02).

## Explicit non-changes (T01 scope boundary)

- No removal of production compose `nginx` service (T02).
- No change to `scripts/foundation/voc067-cutover-bridge-gate.test.mjs` (T02).
- No rewrite of historical change-package evidence under `specs/changes/VOC-041*`,
  `VOC-042*`, or VOC-067 narratives.
- Cloudflare rollback script defaults mentioning `:8443` restore target unchanged.

## Limitations

Live production deploy verification of OAuth/CORS on canonical URLs belongs to
post-merge release and **VOC-079-T03** evidence, not this repository-only task.
