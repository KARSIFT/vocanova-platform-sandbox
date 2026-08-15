# VOC-079 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment by being merged
as a draft or even by being adopted alone. Adoption authorizes
implementation PRs only. Each task PR still requires independent
verification against the exact revision.

Proposed risk is **R4** (draft): live production edge topology and rollback
path change completing VOC-067's shared-edge supersession. Founder adoption
of this package (and acceptance of the R4 proposal) is required.
`automatic_merge_allowed: false` is set per AGENTS.md for R4 (redundant with
merge-gate's R4 hard block; keeps the package self-describing).

Under active A-003 and this repository's 2026-08-08 founder delegation (see
`AGENTS.md`), merged tasks on `develop` can auto-promote to `main` and
trigger `deploy-production.yml` once the package task roster closes —
**unless** the adopting human records a temporary hold for the bridge-
retirement window (recommended between T02 merge and T03 verification).
Founder comment-based promotion retry remains available if auto-promotion
fails.

Routine R3 path-floor work does not by itself recreate standing
technical-steward approval under A-003; R4 founder authority remains.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with R4 acceptance; `VOC-079-DEP-00`–`DEP-02` resolved or
  explicitly deferred in writing.
- T00: repository verify-only `OK` and VOC-067-EV-05
  `cloudflare_remap_api_status: absent`.
- T01/T02 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on exact SHAs.
- Shared edge already healthy on origin `:443` for all four hostnames
  (issue #624 2026-08-15 baseline; reconfirm in T03).

Monitoring during/after T03:

- External HTTPS for staging and production web/API (canonical URLs, no
  port).
- Shared-edge container health; absence of `vocanova-production-nginx` and
  host `8081`/`8443`.
- OAuth start / callback success on production after URL normalization.
- Existing Sentry / uptime signals for both tiers (shared edge = shared
  fault domain on reload).

Outcome owner: named in `VOC-079-EV-03` (unassigned at drafting). Success =
`VOC-079-AC-00` through `VOC-079-AC-07` with linked evidence.

## Rollback

Trigger: either tier unreachable or 5xx on `:443` after bridge retirement;
OAuth/CORS failure after `:8443` stripping; shared nginx crash loop;
accidental removal of shared-edge via mis-scoped orphans.

Mechanism:

1. Redeploy prior revision that still defines
   `vocanova-production-nginx` on `8081`/`8443` (primary application
   rollback).
2. If edge mapping must again target `:8443`, run Cloudflare `--restore` /
   workflow `restore`.
3. Re-verify with `infra/scripts/verify-voc067-cutover.sh` (add
   `--include-8443-bridge` only when the bridge revision is live again).

Validation: external curls succeed; shared-edge healthy; no unexplained
manual Docker residue required for recovery.

Accountable owner: T03 evidence. Last-known-good: pre-T02 dual-publish
repository tip plus VOC-067 remap restore tooling.

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind the exact reviewed commit SHA for each task.
- Confirm the change matches this specification and acceptance criteria.
- Run/inspect applicable deterministic checks; never treat missing
  Cloudflare / SSH / production access as a pass for T00/T03.
- Escalate if semantic risk exceeds the declared class.
- Verify the implementer-role occupant did not approve or merge its own
  implementation.
- Identify active authority model (`a003-active`) and report every still-
  required R4 / EHR / adoption / activation gate.

Closure requires acceptance-criteria results recorded with evidence, not
merely merged PRs or a successful production deploy. Repository merge,
release to `main`, production deploy, Cloudflare verification, and package
closure are distinct events and must not be conflated.

Issue #624 closes only when AC-00–AC-07 are satisfied with evidence.
Historical issue #595 retain-bridge rationale remains historical; do not
rewrite it — cite supersession by #624.
