# Monitoring operations (VOC-086)

Operator guide for repository-managed Uptime Kuma availability monitors,
scheduled synthetics, Sentry error monitoring, and `monitoring_impact`
governance on change packages.

## Responsibility split

| Layer | Owner | Stable IDs | What it checks |
| --- | --- | --- | --- |
| **Kuma** (`sync-monitoring.yml`) | `infra/monitoring/monitors.yaml` | `kuma.availability.*` | Availability, TLS, basic HTTP/API health |
| **Scheduled synthetics** (`scheduled-synthetics.yml`) | `infra/monitoring/synthetics.yaml` | `synthetic.*` | OAuth expected state, authenticated journeys, production content |
| **Sentry** (`error-monitoring.yml`) | VOC-051 package | n/a (issue-driven) | Unresolved application errors → governed issues |

Do not replace Sentry with Kuma page checks. Do not put authenticated app
journeys in naive Kuma HTTP monitors.

## Canonical inventory

| File | Purpose |
| --- | --- |
| `infra/monitoring/monitors.yaml` | Five availability monitors with stable IDs |
| `infra/monitoring/synthetics.yaml` | Five scheduled synthetic checks with stable IDs |
| `infra/monitoring/sync-kuma.mjs` | Socket.IO synchronizer (never SQLite) |
| `infra/monitoring/prove-kuma-inventory.mjs` | Read-only Socket.IO proof (no mutation) |

Ownership marker for managed Kuma monitors:
`vocanova:repo-managed` embedded in the monitor description with
`monitor_id=` and `severity=`.

Unrelated manually created Kuma monitors are preserved unless an inventory
entry explicitly adopts them (`adoption.match_name` + `adoption.match_url`).

## Adding monitoring for a page, API, or feature

1. **Choose the layer.** Availability/TLS/basic health → Kuma inventory.
   Authenticated behavior or OAuth state → synthetics registry. Application
   errors → Sentry (separate workflow; open a governed issue if a new error
   class needs tracking).

2. **Add a stable ID** to `monitors.yaml` or `synthetics.yaml` with required
   metadata (environment, owner, URL/harness reference, expected
   status/body, interval/timeout/retries, severity, coverage references).

3. **Declare `monitoring_impact`** in your change package `change.yaml`:
   - `none` — requires non-empty `rationale` (no new monitor/synthetic IDs)
   - `existing` / `add` / `update` — requires valid `monitor_ids` and/or
     `synthetic_ids` from the canonical inventories

4. **Run deterministic validation** before opening the PR:

   ```bash
   node --test scripts/foundation/voc086-monitoring-inventory.test.mjs
   bash scripts/governance/validate-governance.sh
   ```

5. **Apply Kuma changes** through `sync-monitoring.yml` (see below). Add or
   extend scheduled synthetic jobs in `scheduled-synthetics.yml` when the
   check is authenticated or OAuth-specific.

6. **Record proof** using the verification commands in §Alert and check proof.
   After a successful sync, the host script runs a read-only Socket.IO
   monitor-list proof (`prove-kuma-inventory.mjs`) so TEST-10 metadata is
   in the same workflow log as the apply.

## Credential bootstrap and rotation

Kuma credentials live only in the GitHub **`monitoring`** environment:

| Secret | Purpose |
| --- | --- |
| `KUMA_USERNAME` | Existing admin username (preserved on rotation) |
| `KUMA_PASSWORD` | Strong password (generated on first bootstrap) |

Host access uses the same repository secrets as staging deploy:
`STAGING_SSH_HOST`, `STAGING_SSH_USER`, `STAGING_SSH_PRIVATE_KEY`,
`STAGING_SSH_KNOWN_HOSTS`.

Create the `monitoring` environment on this repository before the first
bootstrap. The `sync-monitoring` workflow will not start without it.

### First-time bootstrap

1. Create the GitHub `monitoring` environment on this repository.
2. Dispatch `.github/workflows/sync-monitoring.yml` on `develop` (or `main`
   after promotion):
   - `rotate_credentials`: **true**
   - `sync_inventory`: **true**
3. The workflow invokes Kuma's official `/app/extra/reset-password.js` via
   `docker exec -i` stdin (password never printed). It stores
   `KUMA_PASSWORD` with `gh secret set` and preserves the username.
4. **All existing Kuma sessions are invalidated once** per rotation.

The implementer `pipeline.yml` job has `actions: read` only and cannot
create `workflow_dispatch` events. An operator or App token with
`actions: write` must run the commands below.

### Normal sync (no credential reset)

```bash
gh workflow run sync-monitoring.yml --ref develop \
  -f rotate_credentials=false \
  -f sync_inventory=true
```

Normal runs **never** reset credentials. Only set `rotate_credentials=true`
when bootstrapping or when compromise is suspected.

After a successful apply, the same host container runs
`prove-kuma-inventory.mjs` (read-only monitor-list). Workflow logs must show
`PASS:` for all five `kuma.availability.*` IDs.

## Deploy / sync inventory

The sync workflow SSHes to the shared host and runs
`infra/scripts/sync-kuma-inventory.sh`, which executes the Node synchronizer
then the read-only proof inside a disposable container on
`vocanova-monitoring-net`. Kuma remains loopback-only on `127.0.0.1:3001`
(VOC-081).

Prevalidation runs before any mutation. On partial failure the synchronizer
compensates applied operations and exits non-zero if rollback is incomplete.

## Rollback

Trigger rollback when sync applies the wrong inventory, overwrites unrelated
monitors, leaks credentials, or breaks topology/isolation.

1. **Revert** the responsible repository commit(s) (inventory, workflow, or
   synchronizer).
2. **Re-run sync** from the rolled-back inventory:
   `sync_inventory=true`, `rotate_credentials=false`.
3. **Confirm** read-only proof and external verification (below).
4. **Rotate credentials** only if compromise is suspected
   (`rotate_credentials=true`).

Manually owned monitors that were never adopted must remain untouched.
Last-known-good before VOC-086: two unmanaged production monitors and
deploy-only synthetics (issue #716).

## Alert and check proof

### External availability and monitor-host reachability

```bash
infra/scripts/verify-voc086-monitoring.sh
```

This runs the VOC-081 monitor-host verifier, probes all five canonical
availability URLs, asserts retired `:8081`/`:8443` do not serve HTTP 2xx,
and asserts repository topology invariants (single shared-edge, no
`8081`/`8443` publish in compose, loopback-only Kuma `3001`).

Disposable harness (no live hostname required for wiring checks):

```bash
infra/scripts/verify-voc086-monitoring.selftest.sh
```

### Read-only Socket.IO inventory proof (after sync)

On the shared host with `KUMA_USERNAME` / `KUMA_PASSWORD` set:

```bash
infra/scripts/prove-kuma-inventory.sh
```

Confirms all five `kuma.availability.*` monitors match
`infra/monitoring/monitors.yaml` via authenticated monitor-list. Output is
redacted (no passwords or session tokens). The same command runs at the end
of `sync-kuma-inventory.sh` so a successful `sync-monitoring` run is the
canonical TEST-10 evidence.

### Scheduled synthetics

Manual full run (must target `develop` until the workflow is on `main`):

```bash
gh workflow run scheduled-synthetics.yml --ref develop
```

Single check:

```bash
gh workflow run scheduled-synthetics.yml --ref develop \
  -f synthetic_id=synthetic.staging.oauth-expected-state
```

Mint tokens (`STAGING_SMOKE_TEST_SESSION_MINT_TOKEN`,
`PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN`) are masked in logs. Production
route sweep remains non-mutating (`mutating: false` in registry).

### Sentry error monitoring (non-regression)

`.github/workflows/error-monitoring.yml` remains the hourly Sentry path.
Confirm recent successful runs in GitHub Actions; this package does not
modify that workflow.

## Governance

New or modified change packages must declare `monitoring_impact` in
`change.yaml`. CI validates declarations when a pull-request base/head range is
available. Historical unmodified packages are grandfathered.

See `AGENTS.md` § "Drafting `monitoring_impact` in `change.yaml`" and
`specs/templates/change-package/change.yaml`.

## Topology and access policy

Preserve VOC-081 / VOC-067 invariants:

- One `vocanova-shared-edge-nginx` on host `80`/`443`
- Kuma on `vocanova-monitoring-net` only; `127.0.0.1:3001` publish
- No production `8081`/`8443` bridge in compose
- Staging and production app networks isolated from monitoring

Access policy: `infra/monitoring/access-policy.md` (public Kuma login;
proxied DNS is not authorization).

## Related documentation

- `infra/README.md` — host layout, monitoring Compose, shared edge
- `infra/monitoring/access-policy.md` — `monitor.vocanova.site` exposure
- `docs/operations/11-devops-and-ci-cd.md` — DOC-11 uptime-monitoring amendment
- `specs/changes/VOC-086-manage-monitoring-inventory/` — package evidence
- `specs/changes/VOC-081-route-monitor-vocanova-site-through-the/` — topology predecessor
