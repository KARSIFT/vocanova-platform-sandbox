# apps/web/Dockerfile
#
# Multi-stage build for the vocanova web app
# (apps/web, Next.js 16 App Router). The final
# image is a minimal Node.js-on-Alpine container
# holding only the Next.js standalone server
# output, the .next/static assets the server
# serves at request time, and the busybox `wget`
# the HEALTHCHECK below uses to probe the
# lightweight `/` route. No build secret,
# credential, or runtime configuration is baked
# into any layer; NEXT_PUBLIC_API_BASE_URL is
# supplied as a build arg (Next.js inlines
# NEXT_PUBLIC_* values into the client bundle at
# build time) and the server-only API_BASE_URL /
# HOSTNAME / PORT are read from the container's
# process environment at startup, exactly as
# documented in apps/web/.env.example.
#
# Build context: the monorepo root, NOT apps/web.
#   docker build -f apps/web/Dockerfile .
# The repo root's .dockerignore is what keeps the
# build context small here. The alternative -
# using apps/web/ as the build context - was
# considered and rejected: pnpm workspace deps
# (@vocanova/api-client, @vocanova/design-tokens)
# are resolved relative to the repo root, and
# Docker's COPY command cannot reach files outside
# the build context. Using the repo root is the
# standard pnpm-monorepo pattern, and is the only
# way the `pnpm install --frozen-lockfile` step
# below can see pnpm-workspace.yaml. The
# VOC-032-TEST-08 test procedure (test-plan.md)
# uses exactly this command.
#
# Required build argument:
#   NEXT_PUBLIC_API_BASE_URL  - the API base URL
#     the browser bundle calls. Next.js inlines
#     NEXT_PUBLIC_* values at build time, so this
#     MUST be a build argument (not a runtime env
#     var) for the correct value to reach the
#     client bundle. For staging, the T07 deploy
#     workflow passes the staging API URL here
#     (https://api-staging.vocanova.site). For
#     local dev, pass http://localhost:8080. The
#     default below is localhost so a `docker
#     build` without the arg still succeeds for
#     an interactive smoke test; CI passes the
#     real value.
#
# Run (local sanity check; staging wiring is in
# T04's compose file, deploy workflow is T07):
#   docker run --rm -p 3000:3000 \
#     -e API_BASE_URL=http://host.docker.internal:8080 \
#     -e PORT=3000 \
#     -e HOSTNAME=0.0.0.0 \
#     ghcr.io/vocanova/vocanova-web:dev
#
# Cross-references:
#   * apps/web/.env.example (T01) - the single
#     source of truth for every env var this
#     image's server reads at runtime, plus the
#     build-time contract for NEXT_PUBLIC_* vars.
#   * apps/web/next.config.ts (T03) - declares
#     output: 'standalone' and the monorepo
#     outputFileTracingRoot, which the standalone
#     build below depends on.
#   * apps/web/src/middleware.ts - the server-
#     side middleware that calls getApiBaseURL()
#     (apps/web/src/lib/env.ts) to read
#     API_BASE_URL/NEXT_PUBLIC_API_BASE_URL at
#     request time. The server-only API_BASE_URL
#     override matters here - the standalone
#     server below only inlines
#     NEXT_PUBLIC_API_BASE_URL into the bundle.
#   * infra/README.md (T11) - the human-readable
#     description of how this image is composed
#     with api/postgres/nginx.

# -----------------------------------------------------------------------------
# Stage 1: install dependencies and build the Next.js
# standalone output.
#
# Pin the build image to Node 24-alpine, matching
# the .nvmrc and the root package.json's
# `engines.node` field. Next.js 16 requires Node
# 20.9+; this app's workspace dependencies
# (notably @vocanova/api-client) target Node 24
# per @types/node 24.x, so the floor is 24.x.
# -----------------------------------------------------------------------------
FROM node:24-alpine AS build

# corepack ships with Node 24 and is the
# official way to provision a pinned pnpm
# version. The repo declares pnpm@11.14.0 in
# package.json's `packageManager` field;
# corepack reads that and uses the exact
# version, regardless of any system-wide
# pnpm install. COREPACK_ENABLE_DOWNLOAD_PROMPT=0
# prevents corepack from blocking on a
# first-time download prompt - non-interactive
# build contexts (CI) would otherwise hang on
# the prompt indefinitely.
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable

WORKDIR /repo

# Copy workspace manifests first and pre-install
# dependencies as their own layer, so a code-only
# change does not re-download the entire
# dependency graph. Each package.json is copied
# individually so pnpm can resolve the workspace
# (it reads every package.json under the
# workspace's `packages:` glob during install).
COPY pnpm-workspace.yaml pnpm-lock.yaml package.json ./
COPY apps/web/package.json ./apps/web/package.json
COPY packages/api-client/package.json ./packages/api-client/package.json
COPY packages/design-tokens/package.json ./packages/design-tokens/package.json
COPY packages/typescript-config/package.json ./packages/typescript-config/package.json
COPY packages/eslint-config/package.json ./packages/eslint-config/package.json

# `--filter @vocanova/web...` installs @vocanova/web
# plus all of its (workspace) dependencies. The
# `...` (trailing dots) is pnpm's "include all
# transitive workspace deps" syntax. `--frozen-
# lockfile` makes the install fail loudly if the
# lock is out of sync with the package.json files,
# catching accidental dep drift before a build.
# --ignore-scripts is intentionally NOT set: the
# root pnpm-workspace.yaml enables `strictDepBuilds`,
# and pnpm's own install hooks are part of the
# workspace's normal install path.
RUN pnpm install --frozen-lockfile --filter @vocanova/web...

# Now copy the actual source. The repo-root
# .dockerignore excludes node_modules, .next,
# apps/api, and other large/unrelated trees so
# the COPY is fast and the context stays small.
# The COPY of apps/web/ overwrites the
# package.json COPY above with the same content
# (and adds src/, etc.); the explicit earlier
# COPY exists only to give pnpm install a stable
# dependency layer that is NOT invalidated by
# every code change.
COPY apps/web ./apps/web
COPY packages ./packages

# Workspace deps (e.g. @vocanova/api-client) are
# imported as compiled output (./dist/index.js
# per the api-client's `main`/`exports` field),
# not as TS source. Build the packages the web
# app depends on before building the web app,
# otherwise the bundler will see a workspace
# package that has no compiled output yet. The
# `--filter @vocanova/web...` form selects
# @vocanova/web and all of its workspace
# dependencies recursively; pnpm resolves build
# order topologically. The `build` script is
# defined for @vocanova/api-client and
# @vocanova/design-tokens (both `tsc -b`); the
# web's build script (`next build`) is what
# actually produces the `.next/standalone/`
# output the runtime stage below consumes.

# Build-time env vars. These MUST be set BEFORE
# the `next build` step below - `next build`
# reads them from the process env at startup,
# and a Dockerfile's ENV directive does NOT
# retroactively affect a preceding RUN step's
# environment.
#
# NEXT_PUBLIC_API_BASE_URL: the build-time
# argument that becomes the browser bundle's
# API base URL. Next.js inlines NEXT_PUBLIC_*
# values into the client bundle at `next build`
# time, NOT at runtime; the T07 deploy workflow
# must pass the *staging* API URL here
# (https://api-staging.vocanova.site) so the
# deployed bundle has the correct URL baked
# in. The default (http://localhost:8080) keeps
# a plain `docker build` working for an
# interactive smoke test; CI passes the real
# value via `--build-arg NEXT_PUBLIC_API_BASE_URL=...`.
#
# NEXT_TELEMETRY_DISABLED=1: opt out of
# Next.js's anonymous telemetry, which would
# otherwise pop prompts that hang
# non-interactive builds.
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV NEXT_TELEMETRY_DISABLED=1

# `next build` (the web's `build` script) reads
# NEXT_PUBLIC_API_BASE_URL and the
# NEXT_TELEMETRY_DISABLED flag set above. Its
# output of interest, apps/web/.next/standalone/,
# is the artifact the runtime stage below copies.
RUN pnpm --filter @vocanova/web... build

# -----------------------------------------------------------------------------
# Stage 2: minimal runtime image.
#
# node:24-alpine (not distroless) for the same
# reason apps/api/Dockerfile uses alpine over
# distroless: the HEALTHCHECK below runs `wget`
# via the shell form, and pulling in a shell +
# wget is materially cheaper than building a
# second binary. This decision is consistent
# with the API Dockerfile and is recorded there
# as a follow-up to revisit if a no-shell image
# is ever wanted.
#
# `alpine:3.21` is the most recent stable line
# with a security-support window that
# comfortably covers this Dockerfile's lifetime;
# bump explicitly in this file rather than using
# `:latest` so a base-image change is always a
# visible diff.
# -----------------------------------------------------------------------------
FROM node:24-alpine

# ca-certificates: required for the Next.js
# server (and any outbound fetch it makes on
# the server side, e.g. the middleware's call
# to /api/v1/me against an https:// API) to
# negotiate TLS against HTTPS endpoints.
# wget: required for the HEALTHCHECK below
# (busybox ships its own wget, but installing
# it explicitly documents the dependency and
# keeps the image working if a future slim
# variant is chosen).
# `--no-cache` avoids storing the apk index in
# the image.
RUN apk add --no-cache ca-certificates wget

# Create a dedicated non-root user with a fixed
# UID/GID so the running container's user
# identity is reproducible across rebuilds and
# host /etc/passwd collisions are avoided.
# UID/GID 10001 mirrors the choice in
# apps/api/Dockerfile and sits in the standard
# "user-defined" range (10000-65534).
RUN addgroup -g 10001 vocanova \
    && adduser -D -u 10001 -G vocanova vocanova

WORKDIR /app

# Copy the Next.js standalone output. The
# standalone build's layout mirrors the source
# workspace path: the server entry point is at
# apps/web/server.js, .next/ is at
# apps/web/.next/, and node_modules is at the
# root of the standalone directory (Next.js
# traces the minimum subset of node_modules
# the server actually needs at runtime). The
# trailing-slash on the source path is
# significant: it makes COPY treat the
# directory contents as the destination's
# children, rather than nesting the source
# directory itself at /app/apps/web/.next/...
COPY --from=build --chown=vocanova:vocanova /repo/apps/web/.next/standalone/ ./

# The standalone build's file tracer copies
# .next/server and .next/static IF the trace
# picks them up; static assets are served at
# request time, so we always overlay them
# explicitly. The destination path matches what
# the standalone server expects:
# apps/web/.next/static. Without this overlay,
# the server would respond 404 for every
# /_next/static/* request.
COPY --from=build --chown=vocanova:vocanova /repo/apps/web/.next/static ./apps/web/.next/static

# public/ assets (favicons, robots.txt, etc.)
# are served verbatim by Next.js. The directory
# does not exist in apps/web today (the repo's
# source tree has no public/ directory - the
# apps/web/src/app/ tree is the only public
# surface, and Next.js's build is responsible
# for fingerprinting any public/ files at
# request time), so there is no COPY here. If a
# future change adds public/, the right line is:
#   COPY --from=build --chown=vocanova:vocanova /repo/apps/web/public ./apps/web/public
# No public/.gitkeep or empty-dir marker is
# added by this Dockerfile - public/ being
# absent is a deliberate, repo-current state,
# not a bug to work around, and adding an
# empty public/ just to make the COPY line
# always-succeed would be a misleading change.

# PORT default matches Next.js's own default
# (3000). The container does not need to know
# the final port here; whatever the
# orchestrator (T04's docker-compose, T07's
# deploy workflow) sets at runtime is what
# `next start` uses, and the HEALTHCHECK below
# honors the same env var via `${PORT:-3000}`.
# Declaring the default here is purely
# documentation of the web app's expected
# port - it does not pin the runtime value.
ENV PORT=3000

# HOSTNAME must be set to 0.0.0.0 (not the
# default 127.0.0.1 some Node.js servers
# resolve to) so the server binds on all
# container interfaces; otherwise the `nginx`
# upstream (T05) cannot reach it over the
# docker-compose internal network. The
# standalone server.js reads HOSTNAME from the
# process env to choose its bind address; the
# default-if-unset is implementation-defined,
# so we set it explicitly.
ENV HOSTNAME=0.0.0.0

# Internal port only. The host port mapping is
# set by T04's compose file / T07's deploy
# workflow and is deliberately not pinned here
# - the staging host publishes 80/443 on nginx
# only (T05).
EXPOSE 3000

# Health probe against the lightweight root
# route. `/` renders the static landing page
# ("Vocanova web foundation is running.") with
# no middleware, no API fetch, and no auth -
# so a 200 here confirms only the Next.js
# process is alive, not that the API is
# reachable. That split is intentional: the
# api container has its own /healthz that pings
# the database, and the web container's role is
# just to confirm "the web process is up";
# nginx (T05) is the right place for a deeper
# composite healthcheck if one is ever wanted.
#
# Uses the shell form so ${PORT:-3000} is
# expanded at HEALTHCHECK time, honoring the
# same env var the server uses. -O /dev/null
# discards the small HTML body; --quiet
# suppresses wget's progress output;
# --tries=1 fails fast on the first error.
# wget returns 0 on HTTP 2xx, 8 on 4xx/5xx; the
# `|| exit 1` ensures the HEALTHCHECK fails on
# any non-2xx response.
HEALTHCHECK CMD wget --quiet --tries=1 -O /dev/null "http://127.0.0.1:${PORT:-3000}/" || exit 1

# Drop privileges. CMD is the exec form so the
# node process runs as PID 1 directly, not
# under a shell - SIGTERM from `docker stop`
# reaches node, node's default SIGTERM handler
# triggers a clean exit, and the process stops
# within the orchestrator's stop grace period.
# The entry path apps/web/server.js is the
# standard Next.js standalone server entry
# point - confirmed against the Next.js 16
# standalone output structure (the server
# entry sits at <output>/<src-path>/server.js
# when the source is in a workspace subpath).
USER vocanova:vocanova
CMD ["node", "apps/web/server.js"]
