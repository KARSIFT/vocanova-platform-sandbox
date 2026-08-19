# VOC-086 — Manage monitoring inventory, synthetics, and monitoring-impact governance

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #716](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/716),
prepared for plan review and adoption under **active A-004** (proposed **R4**;
issue text proposed R3, measured path floor is R4 — see `change.yaml`).

## Identity and lifecycle

- Package ID: VOC-086
- Title: Manage monitoring inventory, synthetics, and monitoring-impact governance
- Canonical path: `specs/changes/VOC-086-manage-monitoring-inventory`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R4` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; measured path floor at drafting is
  **R4** for `scripts/governance/*`, else **R3** for monitoring workflows /
  `infra/monitoring`)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`, `implementation.authorized: false`,
  `repository_adoption_status: not-adopted`
- Target branch: `develop`
- Linked GitHub issues:
  - [#716](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/716)
    (this package's requirement source)
- Related but distinct packages:
  - [VOC-081](../VOC-081-route-monitor-vocanova-site-through-the)
    — repository-managed Kuma Compose, network, vhost, access policy
  - [VOC-051](../VOC-051-add-hourly-sentry-based-log-error-monitoring)
    — hourly Sentry error-monitoring (must remain separate and healthy)
  - [VOC-050](../VOC-050-run-core-loop-e2e-against-real-staging-and-gate)
    — synthetic smoke identity + session mint (reused)
  - [VOC-067](../VOC-067-production-outage-root-cause-consider-unifying)
    — staging/production isolation + shared-edge (must be preserved)

## Why this exists

Monitoring topology is repository-managed (VOC-081), but monitor definitions
and monitoring-impact governance are not. Verified 2026-08-16 evidence in
issue #716:

1. Current promoted revision is
   `96a1fd10155c605ba444e624928552f7b62b93d4`.
2. Kuma is healthy at 1.23.17 (`louislam/uptime-kuma:1`), host-bound only on
   `127.0.0.1:3001`, on the isolated monitoring network and shared edge.
3. Repository owns Compose, vhost, access policy, and reachability verifier,
   but has no monitor inventory or synchronizer.
4. Live inventory remains two active 60-second HTTP monitors (Production API
   `/healthz` and Production Web).
5. Missing Kuma coverage: staging web/API and independent
   `https://monitor.vocanova.site` reachability.
6. Deploy synthetics exist for staging/production OAuth and authenticated
   journeys but are deploy-only, not a canonical scheduled inventory.
7. Hourly Sentry `error-monitoring` is healthy and must stay separate.
8. No GitHub secret names exist yet for Kuma username/password.
9. Supported authenticated Socket.IO events (`add`, `editMonitor`,
   monitor-list) and `/app/extra/reset-password.js` are confirmed; no SQLite
   deployment path is needed or allowed.

## What this package does

1. **Canonical inventory** (`VOC-086-T00`): `infra/monitoring/monitors.yaml`
   (and synthetics registry) with stable IDs and required metadata.
2. **Socket.IO synchronizer** (`VOC-086-T01`): idempotent create/update/adopt
   with ownership marker, collision detection, prevalidation, and compensating
   rollback; never SQLite; never log credentials.
3. **Sync workflow + credential bootstrap** (`VOC-086-T02`): explicit
   repository workflow; one-time/rotation-only password reset via Kuma's
   official tool; credentials only in GitHub secrets.
4. **Scheduled synthetics** (`VOC-086-T03`): stable-ID registry/workflow for
   OAuth, content, and authenticated non-mutating journeys.
5. **`monitoring_impact` governance** (`VOC-086-T04`): required field on
   new/changed packages; CI validates; grandfather historical packages.
6. **Docs + live proof** (`VOC-086-T05`): operator docs; deploy inventory;
   Socket.IO status proof; manual synthetic dispatch; independent monitor-host
   reachability; preserve edge/network/secret isolation.

## What this package deliberately does NOT do

- Manual server fixes or direct SQLite read/write for deployment/sync.
- Cloudflare configuration changes.
- Naive unauthenticated Kuma HTTP checks of authenticated app pages.
- Credential logging or printing bootstrap secrets.
- Overwriting unrelated manually owned Kuma monitors unless explicitly adopted.
- Changing Sentry/`error-monitoring.yml` behavior.
- Application schema migrations or real-user mutation.
- Adopting, authorizing, implementing, or merging itself.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Accept proposed **R4** (measured path floor R4 for `scripts/governance/*`),
   or record a written amendment if governance validation is relocated and the
   path floor drops (still not below the semantic monitoring/credential floor).
2. Confirm implementer shape choices for synchronizer runtime
   (`VOC-086-DEP-01`), secret/environment names (`VOC-086-DEP-02`), and
   scheduled-synthetics packaging (`VOC-086-DEP-03`).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment; R4 still requires strengthened evidence, independent
verification, rollout/monitoring, and rollback credibility.
`automatic_merge_allowed: true` is set per AGENTS.md. This draft still
carries no adoption or implementation authority.
