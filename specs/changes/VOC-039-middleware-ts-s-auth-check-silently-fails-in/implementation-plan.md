# VOC-039 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and each task are approved and implementation is
authorized, per this repository's `AGENTS.md` ("a chat prompt or issue alone is not
implementation authority"). `apps/web/src/middleware.ts` is a protected R3 area
(authentication/authorization) — see `specification.md`'s risk section.

## File reconciliation and implementation sequence

Existing target: `apps/web/src/middleware.ts` (read at drafting time; reproduced in
full above — 84 lines, single exported `middleware` function plus `config` and a
`CurrentUserResponse` interface). No conflicting in-flight work against this file is
known at drafting time beyond the two already-merged prior fixes (PR #296 and the
`SameSite=Strict` fix referenced by the issue), both of which are preserved as-is by
this package — this package's diff is additive (`export const runtime`; new log
calls in the three existing failure branches) and does not restructure the
function's existing control flow.

Ordered steps:

1. `VOC-039-T00`: add `export const runtime = "nodejs";` near the existing
   `export const config = { ... };` block. Verify locally (`pnpm --filter web build`
   or the equivalent narrower command documented in `docs/development.md`) that this
   does not break the build, since Node-runtime middleware has some documented
   restrictions/differences from a standard Node.js server context even though it
   lifts most of Edge's restrictions. Deploy to a real environment; re-run the
   direct reproduction (see `VOC-039-TEST-00`) before considering this task done.
2. `VOC-039-T01`: add the regression test, verifying it fails pre-fix and passes
   post-fix (see `VOC-039-TEST-01`'s procedure). This can be implemented in parallel
   with or immediately after `T00` — order between them is an implementer choice,
   not a hard dependency in either direction, though `tasks.md` lists `T00` first
   as the natural default.
3. `VOC-039-T02`: add the three structured log lines, using this repo's existing
   server-side logging convention (implementer to identify by inspecting
   `apps/web/src`'s existing usage, e.g. any existing `pino`/`console`-wrapper/
   similar pattern already used elsewhere in this codebase, rather than introducing
   a new logging dependency). Depends on nothing beyond `T00` being in the same
   file (can be implemented as part of the same PR as `T00`, or separately — an
   implementer/reviewer choice, not fixed here).

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
pnpm validate   # or the narrower pnpm lint / typecheck / test / build
```

Plus this package's own `VOC-039-TEST-00`/`01`/`02` procedures, which require real
or realistic deployed-environment checks beyond what `pnpm validate` alone can
exercise (Edge-vs-Node runtime behavior and real OAuth flow are not observable from
a local unit-test run alone).

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this specification,
confirm `VOC-039-AC-00`/`01`/`02` are each satisfied with real evidence (not
asserted), and confirm no self-approval occurred.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. Deployment to
any real environment for the purpose of `VOC-039-TEST-00`/`02`'s procedures requires
whatever this repo's existing deploy authorization already requires (see
`docs/development.md` and existing `deploy-staging.yml`/`deploy-production.yml`
workflows) — this package does not grant or bypass that.

Rollout sequence (once authorized): deploy to staging first if staging can exercise
OAuth; if not, deploy to production behind the existing rollback mechanism used by
prior fixes in this same investigation (PR #296), with the smoke-test suite
(`VOC-038-T02`, if already merged and available) run immediately after.

Rollback trigger: the auth check begins failing for real users in a new way (e.g. a
fail-open condition caught by `VOC-039-TEST-00`'s negative case slipping through, or
a build/runtime error specific to Node-runtime middleware not present in Edge
runtime). Rollback mechanism: revert the `runtime = "nodejs"` export (single-line
revert) or redeploy the prior known-good artifact, whichever this repo's existing
rollback tooling (per `VOC-037`'s rollback evidence and `VOC-038-T06`'s rehearsal
convention) already supports. Owner: whichever human authorizes and monitors the
deployment, named explicitly in the implementation PR, not left implicit.
