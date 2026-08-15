---
evidence_id: VOC-079-EV-02
task_id: VOC-079-T02
acceptance_criteria:
  - VOC-079-AC-01
  - VOC-079-AC-02
  - VOC-079-AC-05
  - VOC-079-AC-06
tests:
  - VOC-079-TEST-01
  - VOC-079-TEST-02
  - VOC-079-TEST-04
date: 2026-08-15
related_change: VOC-079
reviewed_head: 99cd982b
develop_merge: da8dd5d3b3cabc744f78d9c405c614a3666447b8
production_merge: be8de870b26547b93407b4444f5c01234a77251c
gate_status: resolved
---

# VOC-079-T02 — Single-edge repository convergence evidence

T02 removed the production nginx service and its `8081:80` / `8443:443`
publishes from production Compose. The production deploy now converges with
the explicitly scoped project command:

```text
docker compose -f docker-compose.production.yml -p vocanova-production up -d --remove-orphans
```

That project scope cannot orphan-remove the separately managed shared-edge or
staging projects. Production nginx configuration and certificate material
remain in the production tree for read-only consumption by shared edge.

## Repository and workflow proof

| Evidence | Result |
| --- | --- |
| T02 implementation | PR [#660](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/660), reviewed head `99cd982b` |
| CI and independent review | Run [31877170372](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31877170372), success |
| Governance | Run [31877170050](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31877170050), success |
| Staging deployment after merge | Run [31877326152](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31877326152), success |
| Promotion to production | PR [#662](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/662), `main` merge `be8de870` |
| Production convergence | Run [31884987715](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31884987715), success |

The successor `voc079-single-edge-invariants.test.mjs` verifies the bridge-free
Compose shape, exclusive host port ownership by shared edge, both tier network
attachments, canonical production URLs, scoped orphan removal, fail-closed
shared-edge validation/reload, and removal of production-nginx workflow steps.
The old bridge-retention gate was removed.

## Live realization of the declarative cleanup

The production run logged `vocanova-production-nginx` as an orphan and then
logged both `Removing` and `Removed` while executing the scoped Compose command.
No manual SSH edit, `docker rm`, or unscoped cleanup was used. The same run
passed production Sentry configuration validation, canonical web/API readiness,
and the authenticated production smoke suite. Container and listener evidence
is recorded in `t03-evidence.md`.

## Isolation and security

- Staging and production retain distinct writable configuration, certificate,
  secret, directory, Compose-project, and application-network boundaries.
- Shared edge consumes both tiers' nginx/TLS trees read-only and remains the
  only VocaNova container publishing host ports 80/443.
- Routine application deployment validates shared-edge configuration before
  reload and does not recreate or take ownership of the shared-edge project.
