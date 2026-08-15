// VOC-079-T02 — single shared-edge invariants after production bridge retirement.
// Supersedes voc067-cutover-bridge-gate.test.mjs (bridge-retention gate no longer
// applies once cloudflare_remap_api_status is absent — VOC-079-AC-06).
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const productionComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.production.yml",
);
const sharedEdgeComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.shared-edge.yml",
);
const stagingComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.yml",
);
const deployProductionPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-production.yml",
);

const HOST_HTTPS_PORT_PATTERN = /["']443:443["']/;
const HOST_HTTP_PORT_PATTERN = /["']80:80["']/;
const BRIDGE_PUBLISH_PATTERNS = [/["']8081:80["']/, /["']8443:443["']/];

test("VOC-079-TEST-01: production compose defines only postgres, api, and web", () => {
  const compose = readFileSync(productionComposePath, "utf8");

  assert.doesNotMatch(
    compose,
    /^\s{2}nginx:\s*$/m,
    "production compose must not declare an nginx service after VOC-079-T02",
  );
  assert.doesNotMatch(
    compose,
    /container_name:\s*vocanova-production-nginx/,
    "production compose must not reference vocanova-production-nginx",
  );
  for (const pattern of BRIDGE_PUBLISH_PATTERNS) {
    assert.doesNotMatch(
      compose,
      pattern,
      `production compose must not publish bridge port mapping ${pattern}`,
    );
  }
});

test("VOC-079-TEST-02: only shared-edge compose publishes host 80/443", () => {
  const sharedEdge = readFileSync(sharedEdgeComposePath, "utf8");
  const production = readFileSync(productionComposePath, "utf8");
  const staging = readFileSync(stagingComposePath, "utf8");

  assert.match(
    sharedEdge,
    HOST_HTTP_PORT_PATTERN,
    "shared-edge compose must publish host port 80",
  );
  assert.match(
    sharedEdge,
    HOST_HTTPS_PORT_PATTERN,
    "shared-edge compose must publish host port 443",
  );

  for (const [label, source] of [
    ["production", production],
    ["staging", staging],
  ]) {
    assert.doesNotMatch(
      source,
      HOST_HTTP_PORT_PATTERN,
      `${label} compose must not publish host port 80`,
    );
    assert.doesNotMatch(
      source,
      HOST_HTTPS_PORT_PATTERN,
      `${label} compose must not publish host port 443`,
    );
    for (const pattern of BRIDGE_PUBLISH_PATTERNS) {
      assert.doesNotMatch(
        source,
        pattern,
        `${label} compose must not publish bridge port mapping ${pattern}`,
      );
    }
  }
});

test("VOC-079-TEST-02: shared edge attaches to both tier networks", () => {
  const sharedEdge = readFileSync(sharedEdgeComposePath, "utf8");

  assert.match(
    sharedEdge,
    /name:\s*vocanova-net/,
    "shared-edge compose must declare vocanova-net",
  );
  assert.match(
    sharedEdge,
    /name:\s*vocanova-production-net/,
    "shared-edge compose must declare vocanova-production-net",
  );

  const nginxServiceMatch = sharedEdge.match(
    /^\s{2}nginx:\s*$\n([\s\S]*?)(?=^\s{2}\w|\n\w)/m,
  );
  assert.ok(
    nginxServiceMatch,
    "shared-edge compose must declare nginx service",
  );
  const nginxBlock = nginxServiceMatch[1];
  assert.match(
    nginxBlock,
    /-\s*vocanova-net/,
    "shared-edge nginx must attach to vocanova-net",
  );
  assert.match(
    nginxBlock,
    /-\s*vocanova-production-net/,
    "shared-edge nginx must attach to vocanova-production-net",
  );
});

test("VOC-079-TEST-03/06: operational production deploy paths have no :8443 URLs", () => {
  const deployProduction = readFileSync(deployProductionPath, "utf8");

  const operationalBlocks = [
    readFileSync(productionComposePath, "utf8"),
    deployProduction,
  ].join("\n");

  const linesWith8443 = operationalBlocks
    .split("\n")
    .map((line, index) => ({ line, index: index + 1 }))
    .filter(({ line }) => /:8443/.test(line));

  const allowedHistoricalOnly = linesWith8443.every(
    ({ line }) =>
      /VOC-067|issue #485|rollback|restore|historical|Cloudflare.*remap/i.test(
        line,
      ) || /^\s*#/.test(line),
  );

  assert.ok(
    linesWith8443.length === 0 || allowedHistoricalOnly,
    `operational production paths must not use :8443 (found: ${linesWith8443
      .map(({ index, line }) => `L${index}: ${line.trim()}`)
      .join("; ")})`,
  );
});

test("VOC-079-TEST-04: deploy-production reloads shared edge only with scoped orphan removal", () => {
  const workflow = readFileSync(deployProductionPath, "utf8");

  assert.match(
    workflow,
    /docker compose -f docker-compose\.production\.yml -p vocanova-production up -d --remove-orphans/,
    "production deploy must use project-scoped compose up --remove-orphans to drop retired bridge containers",
  );
  assert.doesNotMatch(
    workflow,
    /docker exec vocanova-production-nginx/,
    "deploy-production must not validate or reload vocanova-production-nginx after VOC-079-T02",
  );
  assert.match(
    workflow,
    /docker exec vocanova-shared-edge-nginx nginx -t/,
    "deploy-production must keep fail-closed shared-edge nginx -t",
  );
  assert.match(
    workflow,
    /docker exec vocanova-shared-edge-nginx nginx -s reload/,
    "deploy-production must reload vocanova-shared-edge-nginx on success",
  );
  assert.doesNotMatch(
    workflow,
    /docker compose -f docker-compose\.shared-edge\.yml/,
    "routine production deploy must not recreate the shared-edge compose project",
  );
});
