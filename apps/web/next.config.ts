// apps/web/next.config.ts
//
// VOC-032-T03: minimal Next.js config that enables
// `output: 'standalone'`, the production-build output
// shape the apps/web/Dockerfile multi-stage build
// copies into the runtime image. Without this option,
// `next build` produces a full Next.js build that
// requires the entire `next` package plus all of its
// optional-dep binary helpers at runtime; with
// `output: 'standalone'`, the same build traces every
// file the server actually needs and emits a self-
// contained `.next/standalone/` directory the
// Dockerfile copies almost verbatim.
//
// `outputFileTracingRoot` is set to the monorepo
// root so the file tracer can resolve workspace
// dependencies. apps/web imports from
// @vocanova/api-client (its only workspace runtime
// dep - @vocanova/design-tokens is a devDep used by
// the design-token CSS generator at development time
// only, never imported from apps/web/src/), which
// lives at packages/api-client/. The default trace
// root is the directory containing this file
// (apps/web/), which would leave the tracer unable
// to follow the pnpm workspace symlinks to
// /packages/* - the standalone build would then ship
// without the api-client compiled output and fail at
// runtime. Setting the trace root to the monorepo
// root is the standard pnpm-monorepo pattern for
// Next.js and is consistent with how the apps/web
// Dockerfile (T03) treats the build context as the
// repo root, not apps/web/.
//
// Nothing else lives in this file on purpose: every
// other Next.js option (redirects, headers,
// experimental.*) is either already configured
// elsewhere (eslint via apps/web/eslint.config.mjs,
// PostCSS via apps/web/postcss.config.mjs) or has
// not been added by any approved change package.
// Adding options here outside an approved change
// package is out of scope per the scope discipline
// in this package's implementation plan.
//
// `__dirname` below is intentionally the CJS global, not derived from
// `import.meta.url`: Next.js always transpiles next.config.ts to
// CommonJS (via SWC) before executing it, and in that CJS wrapper
// `__dirname` is already an implicit function parameter. Redeclaring it
// with `const __dirname = ...` is a duplicate-declaration SyntaxError
// inside that CJS wrapper, which makes Node's module-syntax detection
// silently fall back to treating this file as an ES module - where
// `exports`/`require` don't exist, breaking `next build` with
// "ReferenceError: exports is not defined in ES module scope".

import path from "node:path";
import { withSentryConfig } from "@sentry/nextjs";

// `output: "standalone"` is for the Dockerfile-based staging/production
// build only (see the module comment above). Vercel (used for apps/web PR
// preview deploys) packages its own Functions from a normal `.next` build
// and does not consume the standalone output; forcing this mode there makes
// Vercel's own build-output step fail looking for artifacts standalone mode
// doesn't produce in the expected shape. Vercel sets `VERCEL=1` on every
// build, Docker never does, so gate on that instead of adding a second
// config file to keep in sync with this one.
const nextConfig = {
  ...(process.env.VERCEL ? {} : { output: "standalone" }),
  outputFileTracingRoot: path.join(__dirname, "../.."),
};

// VOC-051-T01: hand-adapted equivalent of the @sentry/nextjs wizard's
// `withSentryConfig` options block. No `org`/`project`/`authToken` is set and
// source-map upload is disabled outright: this package provisions only a
// read-only Sentry API token for the monitoring workflow, never a build-time
// upload token, so leaving upload enabled would make every build attempt (and
// warn about) an upload it can never perform. `telemetry: false` keeps build
// metadata from being sent to Sentry, and `excludeDebugStatements` strips the
// SDK's own debug/logger statements from the built bundle so no Sentry debug
// output can reach a browser console. (`excludeDebugStatements` is used rather
// than the webpack-only `webpack.treeshake.removeDebugLogging` because
// apps/web builds with Turbopack, where the webpack options are inert.)
export default withSentryConfig(nextConfig, {
  silent: true,
  telemetry: false,
  sourcemaps: { disable: true },
  widenClientFileUpload: false,
  bundleSizeOptimizations: { excludeDebugStatements: true },
});
