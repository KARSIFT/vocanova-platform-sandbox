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

import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const nextConfig = {
  output: "standalone",
  outputFileTracingRoot: path.join(__dirname, "../.."),
};

export default nextConfig;
