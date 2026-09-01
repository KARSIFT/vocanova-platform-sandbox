# vocanova CI base image — Node + pnpm + Go preinstalled, so ci-web/ci-api
# don't reinstall toolchains on every run.
#
# Deliberately a single combined image, unlike kandev's 3-way
# runtime/build/desktop split: vocanova has no desktop/Rust product and a
# much smaller CI matrix, so one shared image for ci-web + ci-api is simpler
# with no meaningful downside at this scale. Revisit only if the combined
# image's pull time becomes the bottleneck.
#
# Pinned versions (keep in sync with .nvmrc, package.json, apps/api/go.mod):
#   - Node : 24.18.0
#   - pnpm : 11.14.0
#   - Go   : 1.26.0

FROM node:24.18.0-bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive \
    GOPATH=/root/go \
    PATH=/usr/local/go/bin:/root/go/bin:$PATH

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
    && rm -rf /var/lib/apt/lists/*

RUN corepack enable \
    && corepack prepare pnpm@11.14.0 --activate

RUN curl -fsSL https://go.dev/dl/go1.26.0.linux-amd64.tar.gz -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz

WORKDIR /work

RUN set -eux; \
    node --version | grep -E '^v24\.18\.0$'; \
    pnpm --version | grep -E '^11\.14\.0$'; \
    go version | grep -F 'go1.26.0'
