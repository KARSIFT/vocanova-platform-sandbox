# Monitoring operations (VOC-086)

Operator guide for repository-managed Uptime Kuma availability monitors,
scheduled synthetics and Sentry error monitoring.

> **Current process checked against repository source on 2 October 2026.**
> Follow [AGENTS.md](../../AGENTS.md): PRs target `main`, required checks and the
> merge queue remain mandatory, and draft/`hold` keeps work pending review.
> Staging deploys automatically for matching `main` changes; production deploys
> manually. Historical VOC-086/VOC-087 packages explain the original design;
> their retired governance pipeline and credential-writer assumptions are not
> current instructions. This runbook correction does not claim new owner approval
> or fresh verification of live configuration.

## Responsibility split

| Layer                                                 | Owner                              | Stable IDs            | What it checks                                                   |
| ----------------------------------------------------- | ---------------------------------- | --------------------- | ---------------------------------------------------------------- |
| **Kuma** (`sync-monitoring.yml`)                      | `infra/monitoring/monitors.yaml`   | `kuma.availability.*` | Availability, TLS, basic HTTP/API health                         |
| **Scheduled synthetics** (`scheduled-synthetics.yml`) | `infra/monitoring/synthetics.yaml` | `synthetic.*`         | OAuth expected state, authenticated journeys, production content |
| **Sentry** (`error-monitoring.yml`)                   | Workflow's four-project matrix     | n/a (issue-driven)    | Unresolved application errors → deduplicated GitHub issues        |
| **Workflow failures** (`operational-failure-monitoring.yml`) | Workflow allowlist           | n/a (issue-driven)    | Terminal deploy/scheduled-synthetic failures → GitHub issues      |

Do not replace Sentry with Kuma page checks. Do not put authenticated app
journeys in naive Kuma HTTP monitors.

## Canonical inventory

| File                                        | Purpose                                         |
| ------------------------------------------- | ----------------------------------------------- |
| `infra/monitoring/monitors.yaml`            | Five availability monitors with stable IDs      |
| `infra/monitoring/synthetics.yaml`          | Five scheduled synthetic checks with stable IDs |
| `infra/monitoring/sync-kuma.mjs`            | Socket.IO synchronizer (never SQLite)           |
| `infra/monitoring/prove-kuma-inventory.mjs` | Read-only Socket.IO proof (no mutation)         |

Ownership marker for managed Kuma monitors:
`vocanova:repo-managed` embedded in the monitor description with
`monitor_id=` and `severity=`.

Unrelated manually created Kuma monitors are preserved unless an inventory
entry explicitly adopts them (`adoption.match_name` + `adoption.match_url`).

## Adding monitoring for a page, API, or feature

1. **Choose the layer.** Availability/TLS/basic health → Kuma inventory.
   Authenticated behavior or OAuth state → synthetics registry. Application
   errors → Sentry (separate workflow; open a normal issue if a new error
   class needs tracking).

2. **Add a stable ID** to `monitors.yaml` or `synthetics.yaml` with required
   metadata (environment, owner, URL/harness reference, expected
   status/body, interval/timeout/retries, severity, coverage references).

3. **Describe the monitoring change in the PR:** affected stable IDs, intended
   behavior, live effects and validation. Ordinary work does not require a
   change package or `monitoring_impact` declaration.

4. **Run deterministic validation** before opening the PR:

   ```bash
   node infra/monitoring/validate-inventory.mjs
   node --test scripts/foundation/voc086-monitoring-inventory.test.mjs
   ```

   Use the relevant sync/synthetic tests and repository checks in the
   [development guide](../development.md) for the affected code. The former
   `scripts/governance/validate-governance.sh` command is retired.

5. **After review and merge, apply Kuma changes** through `sync-monitoring.yml`
   (see below); merging the inventory alone does not run that workflow. Add or
   extend scheduled synthetic jobs in `scheduled-synthetics.yml` when the
   check is authenticated or OAuth-specific.

6. **Record proof** using the verification commands in §Alert and check proof.
   After a successful sync, the host script runs a read-only Socket.IO
   monitor-list proof (`prove-kuma-inventory.mjs`) so TEST-10 metadata is
   in the same workflow log as the apply.

## Credentials and supported workflow modes

The [sync workflow](../../.github/workflows/sync-monitoring.yml) consumes Kuma
credentials from the GitHub **`monitoring`** environment:

| Secret          | Purpose                                         |
| --------------- | ----------------------------------------------- |
| `KUMA_USERNAME` | Existing admin username (preserved on rotation) |
| `KUMA_PASSWORD` | Existing password, or securely preprovisioned rotation value |

Host access uses the same repository secrets as staging deploy:
`STAGING_SSH_HOST`, `STAGING_SSH_USER`, `STAGING_SSH_PRIVATE_KEY`,
`STAGING_SSH_KNOWN_HOSTS`.

Confirm the intended `monitoring` environment and required secrets are configured
through the established secret-management process before dispatch. Do not print
values or copy them into commands, issues, documentation or review output.
Manual dispatch requires an authenticated actor with repository Actions write
access; the retired implementer pipeline is not an access path.

The current workflow modes are:

| Mode | Inputs | Current behavior |
| --- | --- | --- |
| Normal inventory sync | `rotate_credentials=false`, `recover_store_only=false`, `preprovisioned_credentials=false`, `sync_inventory=true` | Applies inventory using existing credentials; no password reset |
| Preprovisioned rotation | `rotate_credentials=true`, `recover_store_only=false`, `preprovisioned_credentials=true`; choose `sync_inventory` explicitly | Resets the existing Kuma account password to the preprovisioned value and checks the preserved username; no GitHub environment-secret write |
| Generated-password rotation | `rotate_credentials=true`, `preprovisioned_credentials=false` | Refused: no environment-secret writer credential is configured |
| Store-only recovery | `recover_store_only=true` | Refused for the same missing writer credential |

`preprovisioned_credentials=true` is valid only with rotation enabled and
store-only recovery disabled. Rotation and store-only recovery cannot be
combined. Normal sync does not create a Kuma operator account; existing
authenticated access is a prerequisite, not something this guide establishes.

### Explicit password rotation

For an intentional rotation, securely preprovision the existing username and
the intended strong password in the monitoring environment, then use the
preprovisioned mode above on reviewed `main`. The workflow invokes Kuma's
official reset tool, invalidates existing Kuma sessions and may subsequently
apply inventory when requested. It is not a sign-in test or read-only action.

Generated-password bootstrap and automatic store-only recovery were part of the
historical VOC-086/VOC-087 design. The workflow now refuses them before execution.
Do not bypass that guard or introduce a new credential merely to follow an old
runbook example. Restoring either mode needs a separately reviewed workflow and
credential-management change.

### Normal sync (no credential reset)

```bash
gh workflow run sync-monitoring.yml --ref main \
  -f rotate_credentials=false \
  -f recover_store_only=false \
  -f preprovisioned_credentials=false \
  -f sync_inventory=true
```

This command **changes live monitor inventory**. Normal runs do not reset
credentials. The workflow uploads a bundle even when all action inputs are false;
it has no read-only inspection mode.

### Interrupted rotation recovery

If rotation may have reset Kuma but proof or credential reconciliation did not
finish, preserve retained recovery material. Preflight blocks a new rotation
while that material exists. Do not delete it, blindly reset again or dispatch
`recover_store_only=true` expecting it to work: that mode is currently guarded
off. Coordinate a specific recovery through the established secret-management
process and review any workflow change before execution. Never copy retained
passwords, proof files or username metadata into public evidence.

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
4. **Rotate credentials** when required for compromise response, using the
   supported preprovisioned mode above and preserving interrupted-rotation
   evidence.

Manually owned monitors that were never adopted must remain untouched.
The two unmanaged production monitors and deploy-only synthetics in issue #716
are historical pre-VOC-086 context, not a current rollback target. Select and
verify the actual prior reviewed inventory for the incident.

## Alert and check proof

### External availability and monitor-host reachability

```bash
infra/scripts/verify-voc086-monitoring.sh --skip-socket-proof
```

This runs the VOC-081 monitor-host verifier, probes all five canonical
availability URLs, asserts retired `:8081`/`:8443` do not serve HTTP 2xx,
and checks repository topology invariants (single shared-edge, no
`8081`/`8443` publish in compose, loopback-only Kuma `3001`).
These are public probes and checks of the repository configuration, not an
authenticated inspection of the live host or proof of delivered alerts. The
explicit flag prevents the optional host-side proof from running merely because
Kuma credentials happen to be present in the environment.

Disposable harness (no live hostname required for wiring checks):

```bash
infra/scripts/verify-voc086-monitoring.selftest.sh
```

### Read-only Socket.IO inventory proof (after sync)

On the shared host with appropriately supplied `KUMA_USERNAME` / `KUMA_PASSWORD`,
Docker access and the monitoring network available:

```bash
infra/scripts/prove-kuma-inventory.sh
```

Confirms all five `kuma.availability.*` monitors match
`infra/monitoring/monitors.yaml` via authenticated monitor-list. Output is
redacted (no passwords or session tokens). The wrapper creates temporary files,
starts a disposable container and installs its Socket.IO dependency. The
monitor-list operation does not alter inventory, but the wrapper is not a
zero-write or dependency-free host audit. After applying inventory,
`sync-kuma-inventory.sh` runs the same proof module in its own disposable
container. Record that result separately from the preceding mutations.

The synchronizer preserves existing notification bindings when the inventory
does not own them. Empty bindings can pass inventory equality; this proof does
not establish notification configuration or operator receipt. Confirm receipt
separately through an authorized delivery check or dated authoritative receipt,
keeping destinations and credentials private.

### Scheduled synthetics

The workflow runs hourly. A manual full run includes the mutating staging
synthetic described below; it is not a read-only audit:

```bash
gh workflow run scheduled-synthetics.yml --ref main
```

Single check:

```bash
gh workflow run scheduled-synthetics.yml --ref main \
  -f synthetic_id=synthetic.staging.oauth-expected-state
```

Mint tokens (`STAGING_SMOKE_TEST_SESSION_MINT_TOKEN`,
`PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN`) are masked in logs. Production
route sweep remains non-mutating (`mutating: false` in registry). The staging
core-loop is explicitly `mutating: true`. The separate
[staging journey preparation](../../apps/api/scripts/prepare-synthetic-staging-journey.sh)
creates an absent guarded fixture or retires and replaces the matching active
synthetic account for the full browser run. Making a saved word due
alone is insufficient after a persistent account has completed its daily target.

Preparation requires staging configuration and the pinned staging Compose
scope. In one transaction it retires only the exact active, marked reserved
identity, revokes its live sessions and creates a replacement under the same
address. Retired learning, mission, activity and reward rows remain attached to
their original user ID, with their synthetic marker retained. Each preparation
adds one new fixture identity; no automatic history cleanup is performed.
The normal seed and production checks do not use this operation.

Deployment and scheduled staging journeys share the `staging-synthetic-account`
job concurrency group across preparation, session minting and the complete
browser run. They wait without canceling the running job. Existing workflow
queues remain in place. The
[GitHub concurrency documentation](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency)
describes the shared job lock and queued-run behavior.

A missing preparation script, rejected identity/environment or failed SQL
transaction fails the run before session minting. Do not substitute a counter
reset or remove the requirement to review at least one card. Verify the full
journey after deployment; static wiring checks do not prove hosted execution.

### Sentry and workflow-failure monitoring

[error-monitoring.yml](../../.github/workflows/error-monitoring.yml) queries the
four Sentry projects hourly using `SENTRY_API_TOKEN`, then opens deduplicated
GitHub issues with the workflow's `GITHUB_TOKEN`. It has `issues: write` and no
SSH access. Confirm recent run outcomes without copying raw error payloads into
public evidence. There is no downstream `plan-from-issue` pipeline.

[operational-failure-monitoring.yml](../../.github/workflows/operational-failure-monitoring.yml)
observes completed `deploy-staging`, `deploy-production` and
`scheduled-synthetics` runs, opening or deduplicating sanitized issues for
failure, cancellation or timeout except identified benign deploy supersession.
Its allowlist does not currently include Sentry monitoring, Kuma sync, backup
discovery or the synthetic recovery rehearsal. A created issue proves that
issue path ran; it does not prove the operator received an alert.

## Historical governance context

VOC-086 originally required `monitoring_impact` declarations in change packages
and validation through the former governance pipeline. Those packages remain
historical evidence. Ordinary changes now follow [AGENTS.md](../../AGENTS.md),
the PR description and current checks; there is no required change-package
ceremony or retired governance-script invocation.

## Topology and access policy

Preserve VOC-081 / VOC-067 invariants:

- One `vocanova-shared-edge-nginx` on host `80`/`443`
- Kuma on `vocanova-monitoring-net` only; `127.0.0.1:3001` publish
- No production `8081`/`8443` bridge in compose
- Staging and production app networks isolated from monitoring

Access policy: [monitor access policy](../../infra/monitoring/access-policy.md) (public Kuma login;
proxied DNS is not authorization).

## Related documentation

- [Infrastructure](../../infra/README.md) — host layout, monitoring Compose, shared edge
- [Monitor access policy](../../infra/monitoring/access-policy.md) — monitor-host exposure
- [DevOps plan](11-devops-and-ci-cd.md) — current process and historical amendments
- [VOC-086](../../specs/changes/VOC-086-manage-monitoring-inventory/README.md) — historical package evidence
- [VOC-081](../../specs/changes/VOC-081-route-monitor-vocanova-site-through-the/README.md) — topology predecessor
