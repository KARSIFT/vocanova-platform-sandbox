# VOC-042 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and each task are approved and implementation is
authorized, per this repository's `AGENTS.md` ("a chat prompt or issue alone is not
implementation authority"). `infra/docker-compose.production.yml` is a protected
R3 area (production infrastructure configuration that sets the server-side value
`apps/web/src/middleware.ts` depends on) — see `specification.md`'s risk section.

## File reconciliation and implementation sequence

Existing target: `infra/docker-compose.production.yml` (read in full at drafting
time — 156 lines). The `web` service's `environment:` block (currently lines
92-107, specifically the `API_BASE_URL` value at line 107) is the sole in-scope
line. No conflicting in-flight work against this file is known at drafting time.
This package's diff is a single-line value change plus a new regression check in a
separate file — no other line in this file is touched.

Ordered steps:

1. `VOC-042-T00`: in the `web` service's `environment:` block, change
   `API_BASE_URL: https://api-production.vocanova.site` to
   `API_BASE_URL: https://api-production.vocanova.site:8443`. Leave the existing
   explanatory comment (lines 93-106) as-is unless the implementer finds it no
   longer accurate after the fix (drafting-time review found no such inaccuracy —
   the comment explains why the var is set at all, not what port it uses). Verify
   by reading the post-fix file directly and confirming the line matches
   `VOC-042-AC-00`'s exact expected string, and by exercising
   `apps/web/src/lib/env.ts`'s `getApiBaseURL()` (directly, or via an existing/new
   unit test) with `API_BASE_URL` set to the post-fix value, confirming it resolves
   to `https://api-production.vocanova.site:8443`.
2. `VOC-042-T01`: add a deterministic regression check for this defect class.
   Concretely: a small script or test (implementer to identify this repo's best-fit
   mechanism — e.g. a shell test under an existing `scripts`/CI-test convention, or
   a lightweight YAML-parsing test if `apps/web`'s or a shared package's test stack
   already parses YAML for another purpose, consistent with how VOC-041-T01 solved
   the equivalent problem for `deploy-production.yml`) that extracts the `web`
   service's `API_BASE_URL` value from `infra/docker-compose.production.yml` and
   asserts it contains `:8443` immediately after the hostname. Must fail against
   the pre-fix (unqualified) value and pass against the post-fix one — implementer
   should verify this by temporarily reverting `VOC-042-T00`'s change in a
   throwaway local check, confirming the new check fails, then restoring the fix
   and confirming it passes. Runs via `pnpm validate` or a narrower documented
   script, per this repo's existing CI convention.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
pnpm validate   # or the narrower relevant script, once VOC-042-T01's check exists
```

Plus this package's own `VOC-042-TEST-00`/`01`/`02` procedures.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this specification,
confirm `VOC-042-AC-00`/`01`/`02` are each satisfied with real evidence (not
asserted), and confirm no self-approval occurred. The reviewer should also confirm
this package's stated non-goals (no change to `deploy-production.yml`, the `api`
service's own environment block, or `infra/docker-compose.yml`) were actually
preserved in the diff, not silently expanded.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. This
package's change only takes effect the next time the `web` service is recreated
(a fresh `docker compose up -d web`, a full redeploy via `deploy-production.yml`,
or an operator-initiated manual recreate); it does not itself restart or touch the
currently-running production `web` container.

Rollout sequence (once authorized): merge to `develop`; the fix takes effect the
next time production's `web` service is recreated. Whether an immediate manual
recreate is needed before the next scheduled deploy is an operational decision for
the reviewing human (see `README.md`'s recommended next action 3) — this package
does not perform that recreate and has no production access to do so.

Rollback trigger: a future edit reintroduces an unqualified value, or the appended
port turns out to be wrong for a reason not anticipated at drafting time (e.g. a
future infrastructure change moves the API off `:8443`). Rollback mechanism: revert
the single-line change (or redeploy the prior known-good artifact) — the same
mechanism already documented for VOC-039's single-line revert and VOC-041's
three-line revert. Owner: named explicitly in the implementation PR at deploy time,
not left implicit.
