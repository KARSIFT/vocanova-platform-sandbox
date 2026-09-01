# vocanova CI base image — Node + pnpm + Go + the wider toolset ci-web's
# foundation test suite actually shells out to, preinstalled so ci-web/ci-api
# don't reinstall toolchains on every run.
#
# Deliberately a single combined image, unlike kandev's 3-way
# runtime/build/desktop split: vocanova has no desktop/Rust product and a
# much smaller CI matrix, so one shared image for ci-web + ci-api is simpler
# with no meaningful downside at this scale. Revisit only if the combined
# image's pull time becomes the bottleneck.
#
# The tool list beyond Node/pnpm/Go (jq, gh, docker CLI, python3, ssh/scp/
# rsync, make, tar/unzip/gzip) exists because scripts/foundation/*.test.mjs
# shells out to all of them, even for tests unrelated to what they're
# nominally about (deploy/monitoring scripts get invoked to prove their
# *content* is correct). docker CLI needs the host's socket mounted via the
# consuming workflow's `container: options:` - the CLI alone has no daemon
# to talk to inside the container.
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
        gnupg \
        git \
        jq \
        make \
        openssh-client \
        rsync \
        python3 \
        tar \
        gzip \
        unzip \
    && rm -rf /var/lib/apt/lists/*

RUN corepack enable \
    && corepack prepare pnpm@11.14.0 --activate

RUN curl -fsSL https://go.dev/dl/go1.26.0.linux-amd64.tar.gz -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz

# GitHub CLI, via its official apt repo.
RUN mkdir -p -m 755 /etc/apt/keyrings \
    && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# Docker CLI only (no daemon) - static binary, matches kdlbs/kandev's approach.
RUN curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-27.3.1.tgz \
        -o /tmp/docker.tgz \
    && tar -xzf /tmp/docker.tgz -C /tmp \
    && mv /tmp/docker/docker /usr/local/bin/docker \
    && rm -rf /tmp/docker /tmp/docker.tgz

WORKDIR /work

RUN set -eux; \
    node --version | grep -E '^v24\.18\.0$'; \
    pnpm --version | grep -E '^11\.14\.0$'; \
    go version | grep -F 'go1.26.0'; \
    jq --version; \
    gh --version; \
    docker --version | grep -F '27.3.1'; \
    python3 --version; \
    ssh -V; \
    scp -V 2>&1 | head -1; \
    rsync --version | head -1; \
    make --version | head -1; \
    tar --version | head -1
