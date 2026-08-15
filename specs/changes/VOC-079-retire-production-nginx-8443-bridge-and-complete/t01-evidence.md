---
evidence_id: VOC-079-EV-01
task_id: VOC-079-T01
acceptance_criteria: VOC-079-AC-04
tests: VOC-079-TEST-03
date: 2026-08-15
related_change: VOC-079
accountable_owner: implementer (repository task)
t00_gate: absent
t00_evidence: VOC-079-EV-00
t00_verifying_run: https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31876429297
---

# VOC-079-T01 — Canonical production URLs on ordinary :443

## Dependency gate (VOC-079-T00 / AC-00)

This remediation revision lands **after** VOC-079-AC-00 closed:

- `t00-evidence.md` `gate_status: resolved`
- VOC-067-EV-05 `cloudflare_remap_api_status: absent`
- Verifying repository run:
  [#39](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31876429297)
  (`verify-only` on `main` @ `b599847…`; required
  `OK: no origin rules remap production hosts to port 8443`)

T01’s declared `depends_on: ["VOC-079-T00"]` and package order
T00 → T01 are therefore satisfied. Attempt 1 failed independent review
precisely because this gate was still open; URL content was otherwise
accepted.

## Summary

Repository operational configuration no longer qualifies production browser,
OAuth, readiness, smoke, or compose `API_BASE_URL` values with `:8443`.
The temporary `vocanova-production-nginx` cutover bridge service remains in
`infra/docker-compose.production.yml` until **VOC-079-T02** (bridge retirement
is explicitly out of scope for this task).

Historical VOC-041/VOC-042/VOC-067 evidence packages are not rewritten as
though the bridge never existed; VOC-067-EV-05 receives only the forward
`absent` gate update required by T00/AC-00, plus a dated addendum.

## Files changed

| Path | Change |
| --- | --- |
| `infra/docker-compose.production.yml` | `API_BASE_URL` → `https://api-production.vocanova.site`; header notes VOC-079-T01 URL normalization while bridge remains |
| `.github/workflows/deploy-production.yml` | `BASE_URL` / OAuth allowlist without `:8443`; readiness polls, session-mint, and smoke on canonical `:443`; header/input comments aligned |
| `scripts/foundation/docker-compose-production-api-base-url-port.test.mjs` | Asserts canonical URL; rejects `:8443` (VOC-079-TEST-03) |
| `scripts/foundation/deploy-production-oauth-port.test.mjs` | Asserts unqualified OAuth/browser URLs; rejects `:8443` (VOC-079-TEST-03) |
| `infra/README.md` | Steady-state operator section documents canonical `:443` URLs; bridge retention until T02 |
| `specs/changes/VOC-079-…/t00-evidence.md` | Record successful verify-only run #39; `gate_status: resolved` (AC-00 unlock for this task) |
| `specs/changes/VOC-067-…/t05-live-cutover-evidence.md` | Frontmatter `cloudflare_remap_api_status: absent` + §8 addendum (historical sections preserved) |

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
  `VOC-042*`, or VOC-067 narratives beyond the EV-05 `absent` gate / addendum.
- Cloudflare rollback script defaults mentioning `:8443` restore target unchanged.

## Limitations

Live production deploy verification of OAuth/CORS on canonical URLs belongs to
post-merge release and **VOC-079-T03** evidence, not this repository-only task.
