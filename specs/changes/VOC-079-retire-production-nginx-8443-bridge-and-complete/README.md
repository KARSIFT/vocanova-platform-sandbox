# VOC-079 — Retire Production nginx :8443 Bridge and Complete the Single Shared-Edge Cutover

**Status: adopted; implementation authorized.** The founder adopted this R4 package
by approving plan PR #625 after independent plan review. It responds to
[issue #624](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/624),
and is implemented only through its ordered governed tasks.

## Identity and lifecycle

- Package ID: VOC-079
- Title: Retire Production nginx :8443 Bridge and Complete the Single Shared-Edge Cutover
- Canonical path:
  `specs/changes/VOC-079-retire-production-nginx-8443-bridge-and-complete`
- Lifecycle state: `adopted` (implementation authorized; no task authority exercised yet)
- Accepted risk: `R4` (see `change.yaml`'s
  `planned_implementation_risk_floor`; measured path floor at drafting is
  **R3** for `infra/*` and deploy workflows)
- Decision owner: `m-e-h-r-d-a-a-d` (founder)
- Approval evidence: founder `approved` comment on PR #625 after independent
  plan review of commit `364333c6e89226b99e96b34a3cda6cd59504b19c`
- Target branch: `develop`
- Linked GitHub issues:
  - [#624](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/624)
    (this package's requirement source)
  - [#485](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/485) /
    [VOC-067](specs/changes/VOC-067-production-outage-root-cause-consider-unifying)
    (shared-edge introduction; bridge kept pending remap API confirm)
  - [#595](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/595)
    (prior retain-bridge closure — superseded by founder direction on
    2026-08-15)
  - [#535](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/535) /
    [VOC-072](specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05)
    (zone-scoped token for repository verify-only)

## Why this exists

VOC-067 introduced `vocanova-shared-edge-nginx` on host `80`/`443` while
keeping a temporary `vocanova-production-nginx` bridge on `8081`/`8443` until
Cloudflare origin-port remap absence is API-confirmed
(`cloudflare_remap_api_status: absent` in VOC-067-EV-05). That frontmatter
is still `unconfirmed`. VOC-067-T04 (strip production `:8443` URLs) remains
pending. Issue #595 retained the bridge; founder direction on 2026-08-15
requires finishing the single-edge design through the repository and normal
deploy workflow — no manual server edits.

Issue #624 records live 2026-08-15 origin `:443` HTTP 200 for all four
hostnames and founder dashboard confirmation that no Origin Rules remap
production to `:8443`. This package still requires repository-controlled
`verify-only` evidence before retiring the bridge.

## What this package does

1. **Cloudflare absence evidence** (`VOC-079-T00`): run
   `deploy-production.yml` `verify-only` with the zone-scoped token; record
   redacted evidence; set VOC-067-EV-05 `cloudflare_remap_api_status` to
   `absent` only on `OK: no origin rules remap…`.
2. **Canonical production URLs** (`VOC-079-T01`): remove operational `:8443`
   qualifications from compose `API_BASE_URL`, deploy-emitted OAuth/BASE_URL
   values, readiness/smoke probes, and current operator docs; invert
   foundation tests that currently require `:8443`.
3. **Bridge retirement + workflow + invariants** (`VOC-079-T02`): remove the
   production compose `nginx` service; add scoped declarative orphan removal
   so the obsolete container disappears on normal deploy; drop
   `vocanova-production-nginx` validate/reload steps; keep fail-closed
   shared-edge `nginx -t` + reload; replace the bridge-retention gate with
   single-edge invariants.
4. **Live verification and rollback** (`VOC-079-T03`): deploy via the normal
   production path; prove one nginx container, no `8081`/`8443` publish, all
   four canonical HTTPS checks; document and rehearse rollback.

## What this package deliberately does NOT do

- Not adding `monitor.vocanova.site` routing or Uptime Kuma.
- Not cookie isolation between staging and production.
- Not general Sentry changes.
- Not unrelated container or host hardening.
- Not manual SSH edits or ad hoc `docker stop/rm` on the server.
- Not rewriting historical VOC-041/VOC-042/VOC-067 evidence as if the bridge
  never existed.
- Not bypassing ordered task authority, independent verification, or release gates.

## Adopted decisions

1. VOC-079 supersedes unfinished VOC-067-T04/T05 and VOC-072-T02 completion
   work without resetting predecessor attempt budgets.
2. Production cleanup uses Compose project-scoped `--remove-orphans`; the
   shared-edge compose project is not targeted or recreated.
3. Repository-controlled Cloudflare `verify-only` evidence is mandatory before
   bridge retirement; dashboard confirmation alone is supporting context.
4. Risk is **R4** (path floor R3; semantic elevation for live edge topology and
   rollback-path change).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This
package carries founder adoption and implementation authorization. Per-task
independent verification of each exact revision and production outcome evidence
remain distinct gates per AGENTS.md and CLAUDE.md.
