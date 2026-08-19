# VOC-087 — Make the first repository-managed Kuma sync adopt live monitors safely

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #728](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/728),
prepared for plan review and adoption under **active A-004** (proposed **R3**;
measured path floor is R3 — see `change.yaml`).

## Identity and lifecycle

- Package ID: VOC-087
- Title: Make the first repository-managed Kuma sync adopt live monitors safely
- Canonical path:
  `specs/changes/VOC-087-make-the-first-repository-managed-kuma-sync-adopt`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R3` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; measured path floor at drafting is
  **R3** for `infra/monitoring`, `infra/scripts`, and
  `.github/workflows/sync-monitoring.yml`)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`, `implementation.authorized: false`,
  `repository_adoption_status: not-adopted`
- Target branch: `develop`
- Linked GitHub issues:
  - [#728](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/728)
    (this package's requirement source)
- Related but distinct packages:
  - [VOC-086](../VOC-086-manage-monitoring-inventory)
    — repository-managed inventory, Socket.IO synchronizer, sync workflow
  - [VOC-081](../VOC-081-route-monitor-vocanova-site-through-the)
    — repository-managed Kuma Compose, network, vhost, access policy

## Why this exists

VOC-086 introduced the first repository-managed Kuma inventory apply, but
that path is not safe against the current live inventory. Issue #728 records
fresh 2026-08-19 read-only verification of exactly these active monitors:

1. `VocaNova Production API` at
   `https://api-production.vocanova.site/healthz`
2. `VocaNova Production Web` at `https://production.vocanova.site`

The merged inventory instead declares adoption matches for
`Production API /healthz` and `Production Web`; the production web URL also
includes a trailing slash. `infra/monitoring/kuma-sync/plan.mjs` requires
exact name and URL matches. It therefore cannot adopt either live monitor.
The API URL matches exactly after the name mismatch, so unmanaged URL
collision fails closed. The web URL differs only by trailing slash, so exact
URL collision also misses and a duplicate create is possible.

A second safety defect is in the same first-apply path:
`inventoryEntryToDesiredMonitor` sets `notificationIDList: {}`. Any adopted
or managed monitor update sends that empty value, which can remove existing
notification bindings. Notification ownership is outside the inventory
schema and must be preserved unless explicitly managed.

VOC-086 T01 evidence records the incorrect live adoption names. T02 evidence
describes obsolete rotation ordering after independently reviewed
remediation, and its two shell harnesses are documented but not executed by
`pnpm test`. The merged rotation path can still strand a successful host
reset if proof/metadata transfer fails and scrub deletes the only remaining
password copy.

## What this package does

1. **Live adoption identity + shared URL normalization** (`VOC-087-T00`):
   align production inventory name/URL/adoption keys to the verified live
   monitors; use one URL comparison for adoption and unmanaged-URL
   collision, including trailing-slash equivalence, without weakening
   collision safety; add deterministic tests with the exact live records;
   correct VOC-086 T01 evidence.
2. **Preserve notification bindings** (`VOC-087-T01`): stop sending an empty
   `notificationIDList` on adopt/update unless inventory explicitly takes
   ownership; prove bindings survive adoption, update, and a second no-op
   sync.
3. **Rotation recovery + CI harness execution** (`VOC-087-T02`): close the
   reset-success/proof-transfer-failure hole so rotation remains recoverable
   without exposing credentials; execute the two shell harnesses from the
   normal deterministic test entry point; correct VOC-086 T02 evidence.

Live `sync-monitoring` inventory apply remains deferred until this package
is merged and deployed. This package does not perform that first live apply.

## What this package deliberately does NOT do

- The first live `sync-monitoring` inventory apply (deferred until after
  merge and deploy; VOC-086-T05 remains the live-proof task).
- Direct SQLite read/write.
- Overwriting unrelated manually owned Kuma monitors.
- Logging Kuma credentials or notification configuration.
- Cloudflare, Sentry/`error-monitoring.yml`, application schema, or
  synthetic-registry redesign.
- Adopting, authorizing, implementing, or merging itself.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Accept proposed **R3** (measured path floor R3), or raise to R4 in writing
   if first-live-apply and credential-recovery semantics require it.
2. Confirm implementer shape choices for rotation recovery
   (`VOC-087-DEP-02`) and optional notification-ownership schema
   (`VOC-087-DEP-03`).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment; R3 still requires independent verification and
applicable rollback credibility.
`automatic_merge_allowed: true` is set per AGENTS.md. This draft still
carries no adoption or implementation authority.
