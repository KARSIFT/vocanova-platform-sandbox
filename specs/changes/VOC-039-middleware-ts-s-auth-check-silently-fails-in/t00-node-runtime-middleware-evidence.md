# VOC-039-EV-00 — Node-runtime middleware evidence (T00)

Produced by the implementer role on 2026-08-06 during the remediation attempt for
`VOC-039-T00`, in response to the independent verification of commit
`842fbbe1994e9bb19001ba7006d02423d6b2be92`, which found `VOC-039-EV-00` absent
(High) and no local build/validation note (Low).

## Standing of `VOC-039-AC-00`

**PARTIALLY SATISFIED — the deployed-environment element is NOT satisfied and
cannot be satisfied by the implementer role.** Read this section before treating
anything below as acceptance of the fix.

| AC-00 element | Standing after this evidence |
| --- | --- |
| `apps/web/src/middleware.ts` exports `runtime = "nodejs"` | Satisfied — see "Code change" below |
| The export actually changes the compiled runtime (not a no-op under Next.js 16.2.10) | Satisfied — proven at the build-artifact level, see "Build-artifact proof" |
| The Node-runtime build compiles and serves protected routes without regression | Satisfied locally — see "Local production-build behavioral check" |
| Negative case (no/invalid cookie) still redirects to `/signin` | Satisfied locally only — see the same section; **not** yet re-run against a real deployed environment |
| Positive case: a real allowlisted user (or a real, valid session cookie against the real deployed API through the actual middleware execution path) reaches a protected route in a **real deployed environment** | **NOT SATISFIED** — see "What remains and why the implementer cannot do it" |

`VOC-039-TEST-00`'s procedure is explicitly a deployed-environment procedure. The
local checks below are strictly weaker than it and are recorded as supporting
evidence, not as a substitute. The package's own open question 1
(`specification.md`) and `VOC-039-R00` (`impact-analysis.md`) remain open.

## Code change

The whole T00 diff is one additive line in `apps/web/src/middleware.ts`
(line 20, alongside the pre-existing `config` export at lines 6–18):

```ts
export const runtime = "nodejs";
```

`config.matcher`, the `middleware` function's auth control flow (fetch-throw /
`401` / other non-ok all still redirect to `/signin`), and the onboarding-status
gating below it are byte-for-byte unchanged. No fail-open path is introduced; every
failure branch still redirects.

## Build-artifact proof that the export is load-bearing under Next.js 16.2.10

This is the part a plain-Node unit test cannot show, and it is the reason the Low
"no build note" finding is worth more than a bare "the build passed": the two builds
below differ **only** in the presence of line 20, and they produce structurally
different server artifacts.

Commands (run from `apps/web`, after `pnpm install --frozen-lockfile` and
`pnpm run build:packages` at the repo root):

```bash
pnpm build                                  # Next.js 16.2.10, Turbopack
cat .next/server/middleware-manifest.json
ls .next/server/middleware.js
```

With `export const runtime = "nodejs";` present (the committed fix):

```
✓ Compiled successfully in 3.9s
  Finished TypeScript in 3.3s
✓ Generating static pages using 3 workers (12/12)

$ cat .next/server/middleware-manifest.json
{ "version": 3, "middleware": {}, "sortedMiddleware": [], "functions": {} }

$ ls .next/server/middleware.js
.next/server/middleware.js
```

With that single line removed (pre-fix state, reconstructed locally and reverted
immediately afterward — the working tree was restored byte-for-byte):

```
$ cat .next/server/middleware-manifest.json
{
  "version": 3,
  "middleware": {
    "/": {
      "files": [
        "server/edge/chunks/0uss_next_dist_esm_build_templates_edge-wrapper_16hw-nn.js",
        "server/edge/chunks/[root-of-the-server]__0l5mq5h._.js",
        "server/edge/chunks/1n1x_next_dist_esm_build_templates_edge-wrapper_1oak6g7.js"
      ],
      "name": "middleware",
      "page": "/",
      "entrypoint": "server/edge/chunks/1n1x_next_dist_esm_build_templates_edge-wrapper_1oak6g7.js",
      "matchers": [ ... the nine config.matcher patterns ... ],
      ...
    }
  },
  "sortedMiddleware": ["/"],
  "functions": {}
}

$ ls .next/server/middleware.js
ls: cannot access '.next/server/middleware.js': No such file or directory
```

Pre-fix, the middleware is compiled through Next.js's **edge-wrapper** template into
`server/edge/chunks/` and registered as an edge function in
`middleware-manifest.json` — i.e. issue #297's diagnosis (the auth `fetch()` runs
inside the Edge runtime sandbox in this repo's own self-hosted build, not only on
Vercel) is confirmed against this repository's actual build output. Post-fix, no
edge entry exists at all and a plain Node server module
(`.next/server/middleware.js`) is emitted instead. The export is therefore genuinely
load-bearing on Next.js 16.2.10 and is not silently ignored.

Also observed and recorded as a follow-up, not acted on (out of T00's scope): the
build prints `⚠ The "middleware" file convention is deprecated. Please use "proxy"
instead.` Next.js 16 renames this convention; migrating `middleware.ts` to
`proxy.ts` is a separate change, not part of this fix.

## Local production-build behavioral check (supporting evidence only)

A real `next start` production server was run against a local stub API that returns
`200 {"onboardingStatus":"in_progress"}` when the request carries a valid session
cookie and `401` otherwise, with `API_BASE_URL` pointed at it. Requests went through
the actual middleware execution path (a real HTTP request to `/onboarding`), not by
calling the exported function directly.

Post-fix (Node runtime) build:

| Case | Request | Result |
| --- | --- | --- |
| Negative | `GET /onboarding`, no cookie | `307`, `location: /signin?returnTo=%2Fonboarding` |
| Positive | `GET /onboarding`, valid session cookie | `200`, no redirect |

The stub API's own log confirms the middleware really called `/api/v1/me` in both
cases (`cookiePresent=false` then `cookiePresent=true`), so the positive case is a
genuine successful auth check, not a skipped one.

**Crucially, the same two checks were re-run against the pre-fix (Edge runtime)
build and produced identical results** (`307` to `/signin` without a cookie, `200`
with one). This is not a failure of the fix — it is direct local proof of the claim
that motivated `VOC-039-AC-00`'s deployed-environment requirement and
`VOC-039-T01`'s "not a plain unit test" requirement: the Edge-runtime `fetch()`
failure reported in issue #297 does **not** reproduce against a loopback API on a
local machine. Any purely local behavioral or unit test of `middleware.ts` therefore
passes with or without this fix. This is recorded here because it is evidence
against over-claiming, and it is a concrete input for whoever implements
`VOC-039-T01` (a local behavioral harness of this shape is demonstrably insufficient
for that task's fail-pre-fix/pass-post-fix requirement).

## Local validation commands run

From the repository root, on Node `v24.5.0` (the repo's `packageManager`/engines
field wants `24.18.0`; the mismatch produced only a pnpm warning, no failure):

```bash
corepack prepare pnpm@latest --activate   # pnpm 11.14.0
pnpm install --frozen-lockfile            # lockfile up to date, supply-chain policies pass
pnpm run build:packages                   # tsc -b packages/api-client packages/design-tokens
pnpm --filter web build                   # Next.js 16.2.10 production build, incl. TypeScript
```

All succeeded. `implementation-plan.md` step 1's "verify locally that this does not
break the build" is satisfied by the `pnpm --filter web build` result above. The
full `pnpm validate` (format/lint/typecheck/test/build across every workspace,
including the Go API) was **not** run in this sandbox and is left to CI; the web
build it gates on is the part T00's diff can affect, and that part passed.

## What remains and why the implementer cannot do it

`VOC-039-TEST-00` requires a real deployed environment (staging if it can exercise
OAuth, otherwise a controlled production check with an existing allowlisted account),
which requires deploy authorization, production/staging access, and real OAuth
credentials. Per this repository's `AGENTS.md` ("agents do not receive production
secrets and do not deploy directly to production") and this package's own
`change.yaml` (`production_deployment: disabled`,
`release.deployment: not-authorized-by-this-package`), the implementer role has
none of these and must not fabricate or assert them. This mirrors the precedent in
`../VOC-037-begin-milestone-r2-production-readiness-docs/t03-killswitch-rollback-evidence.md`,
where a deployed-environment acceptance criterion was correctly refused by the
implementer and later satisfied by the founder-gate delegate against the real target.

Still required before `VOC-039-AC-00` can be marked satisfied:

1. Deploy this revision to a real environment (staging preferred).
2. Run `VOC-039-TEST-00`'s positive case: complete real Google OAuth sign-in as an
   allowlisted account, then reach `/onboarding` without being bounced to `/signin`.
3. Run `VOC-039-TEST-00`'s negative case against the same deployment: no cookie and
   an invalid/expired cookie must both still redirect to `/signin`.
4. Run `VOC-039-TEST-00`'s step 3 cross-check (the issue's own `node -e` in-container
   reproduction) and confirm the two paths now agree.
5. Append those results to this file, then obtain a **separate** independent
   verification of that exact revision — this implementer cannot self-verify, and
   this document is evidence, not approval.

Until step 5 completes, `VOC-039-AC-00`'s result stays `pending` in
`acceptance-criteria.md`, which is why this remediation does not flip it.
