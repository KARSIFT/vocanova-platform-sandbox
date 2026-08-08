# VOC-051-EV-00 / VOC-051-EV-03 — T01 `apps/web` Sentry wiring evidence

Evidence for `VOC-051-T01` (`VOC-051-AC-00`, `VOC-051-AC-07`) and tests
`VOC-051-TEST-00`, `VOC-051-TEST-01`, `VOC-051-TEST-08`.

All runs below were performed in the implementation environment on the exact
working tree this task's pull request contains. No real Sentry DSN, no real
Sentry API token, and no production or staging credential was used: the
"configured DSN" runs point at a throwaway local HTTP sink
(`http://examplepublickey@127.0.0.1:9999/1`) that records the Sentry envelopes
it receives, per `test-plan.md`'s "tests must not use secrets or production
data" rule.

## 1. Environment-variable choice (task-required confirmation)

`tasks.md`'s `VOC-051-T01` requires the implementer to confirm and document the
DSN variable name. Confirmed choice: **`NEXT_PUBLIC_SENTRY_DSN`**, with
`NEXT_PUBLIC_SENTRY_ENVIRONMENT` and `NEXT_PUBLIC_SENTRY_RELEASE` alongside it,
mapping onto `t00-evidence.md` §3's Layout B secrets
(`PRODUCTION_WEB_SENTRY_DSN` / `STAGING_WEB_SENTRY_DSN`).

The `NEXT_PUBLIC_` prefix is required, not stylistic: the browser bundle needs
the value, and Next.js only inlines `NEXT_PUBLIC_`-prefixed variables into
client code. A Sentry DSN is a public write-only ingest key, not a secret — it
is designed to ship in browser bundles — so the prefix leaks nothing that
Sentry's own model does not already expect to be public. Section 4 below
verifies that the DSN is the *only* Sentry value reaching the client bundle.

The server and edge runtimes additionally accept the unprefixed
`SENTRY_ENVIRONMENT` / `SENTRY_RELEASE` (falling back to the `NEXT_PUBLIC_`
forms), so a server-side deployment can retag without a rebuild — the same
variable names `apps/api/app/api/production.go` already reads.

## 2. `VOC-051-TEST-00` — deliberate errors reach Sentry when a DSN is set

Procedure: built `apps/web` in production mode with the sink DSN,
`NEXT_PUBLIC_SENTRY_ENVIRONMENT=staging`,
`NEXT_PUBLIC_SENTRY_RELEASE=sha-evidence`; ran the resulting **standalone**
output the same way `apps/web/Dockerfile`'s `CMD ["node", "apps/web/server.js"]`
runs it (working directory `.next/standalone`); requested a temporary Route
Handler that throws.

The temporary probe route existed only for this run and is **not** part of the
diff — it was deleted before the change was handed off.

Result: the sink received a Sentry envelope containing the thrown error, tagged
with the configured environment and release:

```
{"exception":{"values":[{"type":"Error",
 "value":"VOC-051-T01 temporary server-side probe error", ... }]}}
... "trace":{"environment":"staging","release":"sha-evidence", ...}
```

A session envelope with `"attrs":{"release":"sha-evidence","environment":
"staging"}` arrived alongside it. Server-side capture therefore works, and the
environment/release tagging required by `VOC-051-AC-00` is confirmed on the
wire, not just in configuration.

### 2a. Why `onRequestError` and `src/` placement are both load-bearing

Two findings from this run, both fixed in the diff:

1. **`instrumentation.ts` must export `onRequestError`.** `Sentry.init` alone
   only covers uncaught/global handlers. Next.js catches Route Handler / Server
   Action / RSC errors itself and turns them into a 500, reporting them only
   through the `onRequestError` hook. With the hook absent, the probe error was
   logged by Next.js and never sent. With
   `export const onRequestError = Sentry.captureRequestError` present, it is
   sent.

2. **The instrumentation files must live in `apps/web/src/`, not
   `apps/web/`.** `apps/web` uses a `src/` directory, which is where Next.js
   expects `instrumentation.ts` / `instrumentation-client.ts`. With the files at
   the app root, `next start` still loaded them, but the `output: "standalone"`
   build — the only shape the Docker image runs — silently omitted
   `.next/server/instrumentation.js` entirely, so `register()` never ran in a
   deployed container and no server-side event was ever sent. Confirmed by
   `ls .next/standalone/apps/web/.next/server/instrumentation*`: absent before
   the move, present after, and the probe event only reaches the sink from the
   standalone server after the move.

Client-side capture is wired through `src/instrumentation-client.ts` (verified
present and initialised in the built client bundle, section 4) plus
`src/app/global-error.tsx`'s `Sentry.captureException`. A browser-driven
end-to-end client event was **not** executed here — no browser is installed in
this environment — so that half of `VOC-051-TEST-00` rests on bundle-level
evidence rather than a delivered event. Recorded as a limitation, not claimed
as a pass.

## 3. `VOC-051-TEST-01` — no DSN, no noise; no overlay in production builds

- **DSN unset**: `pnpm build` with `NEXT_PUBLIC_SENTRY_DSN` unset completed
  cleanly; the built client bundles contain no `ingest.sentry.io` DSN at all
  (`rg -l "ingest.sentry.io" .next/static` → no matches). The standalone server
  started and served `/` with HTTP 200 and no Sentry warning or error in its
  log. This mirrors `apps/api`'s no-op-when-unset behaviour.
- **Overlay disabled**: the two options that can surface a Sentry developer UI
  or console output — `debug` and `spotlight` (Sentry's dev overlay) — are
  pinned `false` in all three init sites rather than left to their defaults.
  Verified in the emitted production bundle, not just in source:

```
{dsn:"https://examplepublickey@o0.ingest.sentry.io/0",environment:"staging",
 release:"sha-evidence",debug:!1,spotlight:!1}
```

  A production-mode page load with the DSN set returned HTTP 200 and its HTML
  contained no overlay markup (searched for `spotlight`, `sentry-overlay`,
  `nextjs-portal` — no matches).
- `next.config.ts`'s `withSentryConfig` options additionally set
  `bundleSizeOptimizations.excludeDebugStatements`, stripping the SDK's own
  debug/logger statements from the built bundle, and `silent` /
  `telemetry: false` / `sourcemaps.disable` so builds neither attempt a
  source-map upload they have no token for nor send build telemetry to Sentry.

## 4. No Sentry value other than the DSN reaches the client bundle

`rg -c "examplepublickey" .next/static` matches exactly one client chunk (the
instrumentation-client module shown above). No Sentry auth token, API token, or
org/project slug is configured anywhere in `apps/web`'s build, so none can leak:
source-map upload is disabled outright and `next.config.ts` sets no `org`,
`project`, or `authToken`.

## 5. `VOC-051-TEST-08` / `VOC-051-AC-07` — deterministic validation

Run against this task's exact working tree:

| Command | Result |
| --- | --- |
| `pnpm validate` (workspace: validate-workspace, format:check, lint, typecheck, test, build for web/packages/api) | exit 0 |
| `bash scripts/governance/validate-governance.sh` | exit 0 — "Repository foundation validation passed. Governance structure validation passed." |
| `bash scripts/governance/classify-change-risk.sh --files-from <T01 file list>` | exit 0 — **detected path-based risk floor: R3** |

The detected floor **R3 matches** `change.yaml`'s declared R3, so there is no
mismatch to escalate. The paths establishing the floor are
`.github/workflows/deploy-production.yml`,
`.github/workflows/deploy-staging.yml`, and `apps/web/.env.example`; everything
else in the task diff classifies R1/R2.

Environment note: the validation ran on Node v24.5.0 while `package.json`
pins `24.18.0`, which pnpm reports as an "Unsupported engine" warning. Every
command still exited zero. CI runs the pinned version and is the authoritative
run.

## 6. Privacy / DPA confirmation (`impact-analysis.md` T01 requirement)

`impact-analysis.md` requires this task's implementer to confirm that an
existing privacy policy or DPA covers the new browser → Sentry data flow, **or
flag it as an open item for the reviewing human.** Result: **policy coverage
confirmed; DPA coverage flagged as an open item for the reviewing human.**

Confirmed in-repository (`docs/legal/privacy-policy.md`):

- §2 "Data We Collect" already lists "Monitoring and uptime/error telemetry" as
  a processed category, without restricting it to server-originated telemetry.
- §6 "Data Sharing" already names "Error and uptime monitoring providers" among
  the service providers data may be shared with.
- §11 "International Processing" already names Sentry specifically, with an
  EU (Germany) processing location.

So the policy as written covers this addition and needs no amendment for it —
which is also why this task does not touch that document. Two caveats the
reviewing human owns, neither resolvable from this repository:

- That policy carries its own "Founder Review Record (Required Before
  Publication)" block and an unresolved cross-border transfer basis pending the
  registered legal jurisdiction. Its coverage is therefore *drafted*, not
  *published*.
- Whether an executed DPA exists with Sentry, and whether the Sentry
  organization's data region actually matches §11's stated EU (Germany)
  location, are facts outside this repository. Flagged, not assumed.

What this task did do to keep the exposure minimal, so the open item is as
small as possible:

- No `Sentry.setUser` call and no user-context configuration anywhere in
  `apps/web`, so no user identifier is attached by this change.
- No session replay, no performance/tracing sample rate, and no analytics
  configuration — error events only, per `impact-analysis.md`'s explicit
  instruction not to broaden scope.
- `sendDefaultPii` is left at its default (`false`), so the SDK does not attach
  IP addresses, cookies, or request headers of its own accord.

Residual exposure is therefore: error messages, stack traces, and the URL an
error occurred on. Those URLs can include path segments that identify content a
user was viewing. A human must confirm processor coverage before the production
DSN is treated as live.

## 7. Limitations

- No event was sent to the real Sentry organization; all delivery evidence uses
  the local sink described above.
- The client-side half of `VOC-051-TEST-00` is evidenced at bundle level only
  (no browser available in this environment).
- The deploy-workflow DSN injection changes cannot be executed here; they are
  evidenced by inspection against `deploy-production.yml`'s existing
  `PRODUCTION_SENTRY_DSN` pattern and by the both-or-neither guard added in the
  same step, and will first execute on a real deploy.
- The `apps/api` staging DSN sync added to `deploy-staging.yml` is in scope per
  `t00-evidence.md` §3's layout table and §5's finding; it is likewise
  unexecutable here.
