# apps/api/Dockerfile
#
# Multi-stage build for the vocanova API server
# (apps/api/cmd/api). The final image is a minimal Alpine
# container holding only the compiled Go binary, CA
# certificates (required by lib/pq against any Postgres
# deployment using sslmode=require or higher), and the
# busybox `wget` the HEALTHCHECK below uses to probe
# /healthz. No build secret, credential, or runtime
# configuration is baked into any layer; the binary reads
# every environment variable from the container's process
# environment at startup, exactly as documented in
# apps/api/.env.example.
#
# Build context: apps/api/
#   docker build -f apps/api/Dockerfile apps/api
#   docker build -f Dockerfile .                  (from inside apps/api)
#
# Run (local sanity check; staging wiring is in T04's compose
# file, deploy workflow is T07):
#   docker run --rm -p 8080:8080 \
#     -e DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable \
#     -e BASE_URL=https://api-staging.vocanova.site \
#     -e OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/auth/oauth/google/callback \
#     -e SESSION_COOKIE_DOMAIN=staging.vocanova.site \
#     ghcr.io/vocanova/vocanova-api:dev
#
# Cross-references:
#   * apps/api/.env.example (T01) - the single source of truth
#     for every env var this image's binary reads at runtime.
#   * apps/api/cmd/api/main.go (T00) - the entry point this
#     Dockerfile compiles.
#   * apps/api/app/api/production.go - LoadProductionConfig
#     defines the env-var contract; RegisterHealthz defines
#     the /healthz route the HEALTHCHECK below probes.
#   * infra/README.md (T11) - the human-readable description
#     of how this image is composed with web/postgres/nginx.

# -----------------------------------------------------------------------------
# Stage 1: build the Go binary.
#
# Pin the build image to the exact toolchain apps/api/go.mod
# declares (`go 1.26.0` language, `go1.26.5` toolchain) so a
# future language-feature use anywhere in the apps/api tree is
# compiled with the same Go release the local `go test` path
# and CI's `pnpm test:api` use. The `-alpine` variant keeps
# the build stage small and matches the CGO_ENABLED=0 +
# static-binary choice for the final image below.
# -----------------------------------------------------------------------------
FROM golang:1.26.5-alpine AS build

# CGO is disabled because the final image (alpine) does not
# need a C toolchain at runtime. Combined with
# GOOS=linux this yields a fully static binary that runs
# unchanged on any libc. GOFLAGS=-mod=readonly makes the
# build fail loudly if go.mod / go.sum ever drift out of
# sync during the build, rather than silently rewriting them.
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOFLAGS=-mod=readonly

WORKDIR /src

# Copy module manifests first and pre-download dependencies
# as their own layer, so a code-only change does not
# re-download the entire dependency graph.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the API source tree. .dockerignore keeps
# tests, migration down-files, the openapi/ spec, and any
# local .env out of the build context.
COPY . .

# Build a stripped, static binary:
#   * -trimpath  removes local filesystem paths from the
#                binary (no leaked developer machine paths
#                in stack traces).
#   * -ldflags   -s strips the symbol table; -w strips
#                DWARF debug info. Reduces image size; the
#                final image is a runtime, not a debugging
#                surface.
#   * -o /out/api writes the binary to a stable path the
#                next stage can COPY from.
RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/api \
    ./cmd/api

# -----------------------------------------------------------------------------
# Stage 2: minimal runtime image.
#
# Alpine (rather than distroless) so the HEALTHCHECK below can
# run `wget --spider` without us compiling a second Go binary.
# The alternative (distroless static + a separate healthcheck
# binary) was considered and deferred: a no-shell image is a
# real security win, but for a non-production staging tier
# behind Cloudflare's WAF the operational simplicity of a
# `wget` probe in plain shell outweighs the small attack-
# surface benefit. This decision is recorded as a future
# follow-up in the PR description, not silently chosen.
#
# `alpine:3.21` is the most recent stable line with a
# security-support window that comfortably covers this
# Dockerfile's lifetime; bump explicitly in this file rather
# than using `:latest` so a base-image change is always a
# visible diff.
# -----------------------------------------------------------------------------
FROM alpine:3.21

# ca-certificates is required for the lib/pq driver to
# negotiate TLS against any Postgres host that uses
# sslmode=require or higher. wget is required for the
# HEALTHCHECK below (busybox ships its own wget, but
# installing it explicitly documents the dependency and
# keeps the image working if a future slim variant is
# chosen).
# `--no-cache` avoids storing the apk index in the image.
RUN apk add --no-cache ca-certificates wget

# Create a dedicated non-root user with a fixed UID/GID so
# the running container's user identity is reproducible
# across rebuilds and host /etc/passwd collisions are
# avoided. UID/GID 10001 sits in the standard
# "user-defined" range (10000-65534) and is not used by any
# in-repo tooling.
RUN addgroup -g 10001 vocanova \
    && adduser -D -u 10001 -G vocanova vocanova

# Copy the compiled binary from the build stage, owned by
# the non-root user. The binary is the only thing in the
# runtime image this package produces; everything else is
# system-installed by the apk line above.
COPY --from=build --chown=vocanova:vocanova /out/api /usr/local/bin/api

# PORT default matches apps/api/app/api/production.go's
# LoadProductionConfig default (8080). The container does
# not need to know the final port here; whatever the
# orchestrator (T04's docker-compose, T07's deploy
# workflow) sets at runtime is what the binary uses, and
# the HEALTHCHECK below honors the same env var via
# `${PORT:-8080}`. Declaring the default here is purely
# documentation of the API server's expected port - it
# does not pin the runtime value.
ENV PORT=8080

# Internal port only. The host port mapping is set by
# T04's compose file / T07's deploy workflow and is
# deliberately not pinned here - the staging host
# publishes 80/443 on nginx only (T05).
EXPOSE 8080

# Health probe against the server's own /healthz, which
# pings the database and returns 200 only when the
# database is reachable (apps/api/app/api/production.go:
# RegisterHealthz). Uses the shell form so ${PORT:-8080}
# is expanded at HEALTHCHECK time, honoring the same env
# var the binary uses.
#
# The probe is a GET (not a HEAD) because the /healthz
# route is registered with huma as MethodGet only; the
# huma/chi combination does not auto-implement HEAD for
# GET routes, so a HEAD probe returns 405 Method Not
# Allowed. -O /dev/null discards the tiny JSON body
# ({"status":"ok","database":"ok",...} - a few hundred
# bytes at most) and keeps the container log clean.
# --quiet suppresses wget's progress output; --tries=1
# fails fast on the first error so Docker's --timeout
# (default 30s) is the upper bound. wget returns 0 on
# HTTP 2xx, 8 on 4xx/5xx; the `|| exit 1` ensures the
# HEALTHCHECK fails on any non-2xx response.
HEALTHCHECK CMD wget --quiet --tries=1 -O /dev/null "http://127.0.0.1:${PORT:-8080}/healthz" || exit 1

# Drop privileges. ENTRYPOINT is the exec form so the
# binary runs as PID 1 directly, not under a shell - SIGTERM
# from `docker stop` reaches the binary, the binary's
# SIGTERM handler in cmd/api/main.go fires, and the process
# exits cleanly within the orchestrator's stop grace period.
USER vocanova:vocanova
ENTRYPOINT ["/usr/local/bin/api"]
