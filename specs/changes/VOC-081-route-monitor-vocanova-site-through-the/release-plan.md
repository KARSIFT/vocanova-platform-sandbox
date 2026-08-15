# VOC-081 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment by being merged
as a draft or even by being adopted alone. Adoption authorizes
implementation PRs only. Each task PR still requires independent
verification against the exact revision.

Proposed risk is **R4** (draft): live public edge change restoring a
Cloudflare-fronted administrative monitoring UI, plus monitoring Compose
lifecycle on the shared host. Under **active A-004**, engineering-workflow
gates (plan adoption, R4 merge, release promotion, repository-controlled
deploy) require **no** founder `approved` comment. R4 still carries
stronger evidence, validation, verification, rollout, monitoring, and
rollback obligations. `automatic_merge_allowed: true` is set per AGENTS.md
(`VOC-080-DEP-02`); setting true does not bypass path floors, CI,
independent verification, unparseable-risk fail-closed, or EHR.

After task merges, auto-promotion from `develop` to `main` and
`deploy-production.yml` on push to `main` apply per AGENTS.md —
**unless** adoption records a temporary hold for the shared-edge
network/vhost window (recommended between T03 merge and T04 verification).
Interrupted promotion retries via `reconcile-release`; failed gates remain
fail-closed.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with R4 proposal accepted or amended in writing;
  `VOC-081-DEP-00`–`DEP-02` resolved or explicitly deferred with a safe
  interim rule.
- T00–T03 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on exact SHAs.
- VOC-079 single-edge baseline still holds (one nginx; no `8081`/`8443`;
  four app hostnames healthy) before monitor cutover.
- Backup of `/opt/vocanova/monitoring/kuma-data` completed per T00 before
  first live monitoring Compose converge.

Monitoring during/after T04:

- External HTTPS for `monitor.vocanova.site` (and Access challenge behavior
  if DEP-00 is private).
- WebSocket health for the Kuma UI.
- Shared-edge container health; four staging/production app hostnames.
- Absence of public Kuma `3001` and of a second VocaNova nginx.
- Existing Sentry / hourly `error-monitoring` signals (must remain healthy
  and unchanged).

Outcome owner: named in `VOC-081-EV-04` (unassigned at drafting). Success =
`VOC-081-AC-00` through `VOC-081-AC-07` with linked evidence.

## Rollback

Trigger: shared-edge outage or 5xx on app tiers after monitor changes;
monitor hostname stuck on Cloudflare 520 after claimed success; public
publish of Kuma; access-control failure vs DEP-00; data-path miss that
jeopardizes `kuma-data`; mis-scoped orphan removal.

Mechanism:

1. Redeploy last-known-good prior revision that restores pre-monitor
   shared-edge compose/vhost state (primary).
2. Leave `/opt/vocanova/monitoring/kuma-data` intact unless restoring a
   pre-damage backup is required.
3. Re-verify four app hostnames + monitor hostname behavior; confirm
   single nginx and listener invariants.

Validation: external curls / Access checks match the rolled-back policy;
shared-edge healthy; no unexplained manual Docker residue required for
recovery.

Accountable owner: T04 evidence. Last-known-good: post-VOC-079 single-edge
tip with monitor returning Cloudflare 520 and Kuma on loopback only
(VOC-079-EV-03).

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind the exact reviewed commit SHA for each task.
- Confirm the change matches this specification and acceptance criteria.
- Run/inspect applicable deterministic checks; never treat missing
  Cloudflare / SSH / production access as a pass for T04.
- Escalate if semantic risk exceeds the declared class.
- Verify the implementer-role occupant did not approve or merge its own
  implementation.
- Identify active authority model (`a004-active`) and report every still-
  required R4 evidence obligation, EHR, adoption, and activation gate.
- Confirm access exposure was not closed by “proxied DNS exists.”

Closure requires acceptance-criteria results recorded with evidence, not
merely merged PRs or a successful production deploy. Repository merge,
release to `main`, production deploy, Cloudflare/Access verification, and
package closure are distinct events and must not be conflated.

Issue #665 closes only when AC-00–AC-07 are satisfied with evidence.
