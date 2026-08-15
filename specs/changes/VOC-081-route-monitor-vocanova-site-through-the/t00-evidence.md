---
evidence_id: VOC-081-EV-00
task_id: VOC-081-T00
acceptance_criteria: VOC-081-AC-00
tests: VOC-081-TEST-00
date: 2026-08-15
related_change: VOC-081
accountable_owner: unassigned
gate_status: repository-complete-live-converge-pending
live_converge_claimed: false
---

# VOC-081-T00 — Repository-managed Kuma Compose, data ownership, and backup plan

## Scope of this evidence

This task delivers the **repository source of truth** for the monitoring
Compose project and a written backup/migration plan. It does **not** claim
live host converge, public `https://monitor.vocanova.site/` restoration, or
shared-edge network membership — those are T01–T04.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Monitoring Compose | `infra/docker-compose.monitoring.yml` |
| Operator docs + backup plan | `infra/README.md` (Monitoring host layout section) |
| This evidence | `specs/changes/VOC-081-route-monitor-vocanova-site-through-the/t00-evidence.md` |

## VOC-081-TEST-00 — Identity and data-path assertions

Parsed `infra/docker-compose.monitoring.yml`:

| Assertion | Result |
| --- | --- |
| Compose project name `monitoring` | PASS — top-level `name: monitoring` |
| Container name `vocanova-uptime-kuma` | PASS — `container_name: vocanova-uptime-kuma` |
| Data bind `/opt/vocanova/monitoring/kuma-data` | PASS — `${VOCANOVA_MONITORING_ROOT:-/opt/vocanova/monitoring}/kuma-data:/app/data` |
| No public `0.0.0.0:3001` / `[::]:3001` publish | PASS — only `127.0.0.1:3001:3001` |
| Backup/migration notes recorded | PASS — `infra/README.md` § "Backup and first-converge migration" |

## Compose config validation

Command (repo root):

```bash
docker compose -f infra/docker-compose.monitoring.yml config
```

Result: **PASS** (exit 0). Resolved service `uptime-kuma` uses image
`louislam/uptime-kuma:1`, container name `vocanova-uptime-kuma`, loopback
port mapping, and the documented kuma-data bind defaulting to
`/opt/vocanova/monitoring/kuma-data`.

## Backup and first-converge migration plan (summary)

Full steps live in `infra/README.md`. Before T03's first live converge:

1. Tarball snapshot of `/opt/vocanova/monitoring/kuma-data` on the shared host.
2. Verify the existing data directory is non-empty and will be reused by the
   Compose bind — not replaced with an empty mount.
3. First `compose up` after T03 must preserve monitors and admin state; empty
   re-initialization is a FAIL for VOC-081-AC-00.
4. Rollback: stop the new container; restore `kuma-data` from the tarball if
   the bind path was wrong; shared-edge rollback is separate per release plan.

## Live converge status

**Not performed in T00.** The live host still runs the pre-repository Compose
at `/opt/vocanova/monitoring/docker-compose.yml` (issue #665 / VOC-079-EV-03
baseline). Repository deploy convergence is explicitly deferred to T03; public
HTTPS/WebSocket evidence to T04.

## Limitations

- No SSH or production host access in this task run.
- Image tag `louislam/uptime-kuma:1` is the repository pin; the live container
  may differ until T03 converge — data compatibility is via the preserved
  `/app/data` bind, not image equality at T00.
- `vocanova-monitoring-net` and shared-edge attachment are out of scope (T01).
