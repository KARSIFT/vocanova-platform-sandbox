# VOC-039 — middleware.ts's Auth Check Silently Fails in Production Because Next.js Middleware Runs on the Edge Runtime, Not Node: Specification

## Objective and requirement source

Restore the intended behavior of `apps/web/src/middleware.ts`'s authentication gate
in production, and make a future failure of this kind diagnosable from normal
logs/monitoring rather than requiring SSH access and temporary `console.log`
statements. Grounded in [issue #297](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/297)
in full, including its confirmed-by-reproduction root cause and its two explicit
follow-up asks (regression coverage; structured logging). Not yet approved by a
founder or technical steward — see `change.yaml`'s `requirement_approval_status`.

## Scope and non-goals

In scope:
- Forcing `apps/web/src/middleware.ts` onto the Node.js runtime
  (`export const runtime = "nodejs";`), per the issue's own confirmed-by-reproduction
  suggested fix and Next.js 15.2+/16.2's documented support for this (this repo is
  on Next.js 16.2.10 per the issue).
- Structured logging in the auth-check's failure paths (`try/catch` around the
  `/api/v1/me` fetch, the `401` branch, and the "other non-ok status" branch),
  distinguishing those three cases from each other in whatever this repo's existing
  structured-logging/observability convention is (see `VOC-039-T02` — the concrete
  mechanism is an implementer decision within that convention, not invented here).
- A regression test that exercises the real Edge-vs-Node `fetch()` behavior
  difference this issue identified, not just route-matching or logic-level unit
  tests of `middleware.ts` that would pass under Node regardless of which runtime
  directive is set (see `VOC-039-T01`).

Non-goals:
- Re-diagnosing or re-fixing the two already-merged, already-confirmed prior fixes
  in this same investigation (the `SameSite=Strict` cookie fix and the
  `API_BASE_URL` internal-address fix — both already merged per PR #296 and this
  repository's own `git log`, prior to this package being drafted).
- Changing the auth check's actual authorization decision (what counts as
  authenticated, what redirects where, the onboarding-status gating logic below it).
  This package touches only the runtime the check executes under and the
  observability around its existing failure branches.
- Unblocking or re-attempting VOC-038-T03 itself (the blocked core-loop validation
  task). That remains VOC-038's own task, gated on this fix landing and being
  independently verified in production first.

## Risk and protected areas

`apps/web/src/middleware.ts` implements the production authentication gate for
every protected route (`/onboarding`, `/home`, `/discover`, `/discover/:path*`,
`/progress`, `/reviews`, `/reviews/:path*`, `/settings`, `/settings/:path*` per its
own `config.matcher`). Per `docs/governance/change-risk-classification.md`,
authentication/authorization changes are classified at least `R3` regardless of
diff size. This package proposes `R3` (see `change.yaml`); it does not touch
billing, secrets storage, schema migrations, or infrastructure-as-code, so no
higher class is proposed, but the reviewing human's own judgment governs this, not
this proposal.

## Decisions, contradictions, security, and privacy

No `VOC-039-D00`-style decisions are defined by this draft — none of this package's
three tasks require a founder/product judgment call in the way VOC-038's cohort
composition or expansion thresholds did; the fix, test, and logging asks are all
already fully specified by the issue itself. If the reviewing human disagrees, they
should record why in a decision note at adoption time rather than this package
inventing one.

Security/privacy: this fix does not change what is authenticated or who can access
what — it corrects the *mechanism* by which the check that already exists runs, from
one that currently, silently, always fails (Edge-runtime `fetch()` against this
repo's API, per the issue's reproduction) to one that works (Node runtime). No new
attacker-controlled surface, secret, or personal-data field is introduced. The
structured-logging task (`VOC-039-T02`) must not log the raw session cookie value,
the raw `Cookie` header, or any other credential material in its new log lines —
only the failure category (fetch threw / 401 / other non-ok status) and non-sensitive
metadata (status code, route path, timestamp). This constraint is carried into
`VOC-039-AC-02` and `impact-analysis.md`.

## Data, migrations, analytics, and accessibility

None. This package touches only `apps/web/src/middleware.ts` and its test coverage;
no schema, migration, analytics event, or accessibility surface is affected.

## Open questions

1. **Deployment-topology interaction with `runtime = "nodejs"`.** The issue's
   reproduction ran `node -e "..."` *inside* the production web container directly,
   comparing it against middleware's own (Edge-runtime) `fetch()` — it did not run
   an actual build of `middleware.ts` with the fix applied through this repo's real
   nginx/Docker network path before filing the issue. It is very likely (all
   evidence points the same direction) but not yet *proven* that setting
   `runtime = "nodejs"` alone closes the gap with no other topology-specific
   interaction. `VOC-039-T00`'s acceptance criterion requires the same
   direct-reproduction style check (real session cookie, real deployed API,
   this time through the actual middleware execution path with the fix applied) to
   be re-run before the fix is accepted as proven, not just deployed. This is
   flagged for the reviewing human rather than assumed.
2. **Whether Node-runtime middleware has any cold-start/latency or resource-usage
   difference worth monitoring post-deploy**, given Next.js's own documentation
   describes Node middleware as heavier than Edge middleware. Not a blocker for
   this fix (correctness must come first), but worth a monitoring note in
   `release-plan.md` rather than being silently ignored.
