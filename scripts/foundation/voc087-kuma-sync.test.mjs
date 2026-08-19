// VOC-087-T00 — Live adoption identity and shared URL normalization tests.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { createMockKumaClient } from "../../infra/monitoring/kuma-sync/mock-client.mjs";
import { planSyncOperations } from "../../infra/monitoring/kuma-sync/plan.mjs";
import {
  monitorUrlsEqual,
  normalizeMonitorUrl,
} from "../../infra/monitoring/kuma-sync/url-compare.mjs";
import { syncKumaMonitors } from "../../infra/monitoring/kuma-sync/sync.mjs";
import { parseMonitoringYaml } from "../../infra/monitoring/validate-inventory.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const monitorsPath = path.join(
  repositoryRoot,
  "infra/monitoring/monitors.yaml",
);
const syntheticsPath = path.join(
  repositoryRoot,
  "infra/monitoring/synthetics.yaml",
);

const LIVE_PRODUCTION_API = {
  name: "VocaNova Production API",
  url: "https://api-production.vocanova.site/healthz",
};
const LIVE_PRODUCTION_WEB = {
  name: "VocaNova Production Web",
  url: "https://production.vocanova.site",
};

function loadInventoryDocuments() {
  const monitorsDocument = parseMonitoringYaml(
    readFileSync(monitorsPath, "utf8"),
  );
  const syntheticsDocument = parseMonitoringYaml(
    readFileSync(syntheticsPath, "utf8"),
  );
  return { monitorsDocument, syntheticsDocument };
}

function productionEntries(monitorsDocument) {
  return {
    web: monitorsDocument.availability_monitors.find(
      (entry) => entry.id === "kuma.availability.production.web",
    ),
    api: monitorsDocument.availability_monitors.find(
      (entry) => entry.id === "kuma.availability.production.api-healthz",
    ),
  };
}

function httpMonitorFixture({ id, name, url, keyword = "", description = "" }) {
  return {
    id,
    type: "http",
    name,
    url,
    method: "GET",
    interval: 60,
    timeout: 30,
    maxretries: 0,
    retryInterval: 60,
    keyword,
    accepted_statuscodes: ["200-299"],
    description,
    active: true,
    conditions: [],
  };
}

function mutatingCalls(client) {
  return client.calls.filter((call) =>
    ["add", "editMonitor", "deleteMonitor"].includes(call.op),
  );
}

test("VOC-087-TEST-00: exact live monitor records produce adopt/update, not collision or create", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const { api } = productionEntries(monitorsDocument);

  const initial = {
    1: httpMonitorFixture({
      id: 1,
      name: LIVE_PRODUCTION_WEB.name,
      url: LIVE_PRODUCTION_WEB.url,
      description: "manual legacy",
    }),
    2: httpMonitorFixture({
      id: 2,
      name: LIVE_PRODUCTION_API.name,
      url: LIVE_PRODUCTION_API.url,
      keyword: api.expected_body,
    }),
    99: httpMonitorFixture({
      id: 99,
      name: "Founder manual check",
      url: "https://example.com/manual-only",
      description: "founder-owned",
    }),
  };

  const client = createMockKumaClient(initial);
  const result = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });

  assert.ok(result.changed > 0);
  assert.equal(
    client.calls.filter((call) => call.op === "add").length,
    3,
    "only non-production inventory monitors should be created",
  );

  const manualMutations = client.calls.filter(
    (call) =>
      (call.op === "editMonitor" && Number(call.monitor.id) === 99) ||
      (call.op === "deleteMonitor" && Number(call.monitorID) === 99),
  );
  assert.equal(manualMutations.length, 0);

  const adoptedWeb = client.monitors[1];
  assert.match(
    adoptedWeb.description,
    /monitor_id=kuma\.availability\.production\.web/,
  );
  const adoptedApi = client.monitors[2];
  assert.match(
    adoptedApi.description,
    /monitor_id=kuma\.availability\.production\.api-healthz/,
  );
});

test("VOC-087-TEST-01: trailing-slash-only web URL still adopts rather than duplicate-creating", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const { api } = productionEntries(monitorsDocument);

  const initial = {
    1: httpMonitorFixture({
      id: 1,
      name: LIVE_PRODUCTION_WEB.name,
      url: `${LIVE_PRODUCTION_WEB.url}/`,
    }),
    2: httpMonitorFixture({
      id: 2,
      name: LIVE_PRODUCTION_API.name,
      url: LIVE_PRODUCTION_API.url,
      keyword: api.expected_body,
    }),
  };

  const client = createMockKumaClient(initial);
  const result = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });

  assert.ok(result.changed > 0);
  assert.equal(
    client.calls.some(
      (call) =>
        call.op === "add" &&
        monitorUrlsEqual(call.monitor?.url, LIVE_PRODUCTION_WEB.url),
    ),
    false,
    "production web must not be created when a trailing-slash variant exists",
  );
  assert.match(
    client.monitors[1].description,
    /monitor_id=kuma\.availability\.production\.web/,
  );
});

test("VOC-087-TEST-02: distinct URLs after normalization do not false-positive as collisions", () => {
  const { monitorsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;
  const { web } = productionEntries(monitorsDocument);

  const { operations, errors } = planSyncOperations({
    inventoryMonitors: [web],
    remoteMonitors: {
      7: httpMonitorFixture({
        id: 7,
        name: "Distinct host monitor",
        url: "https://other-production.vocanova.site/",
      }),
    },
    ownershipMarker,
  });

  assert.equal(errors.length, 0);
  assert.equal(operations.length, 1);
  assert.equal(operations[0].type, "create");
  assert.equal(operations[0].inventoryId, web.id);
});

test("VOC-087-TEST-03: name mismatch still fails closed on unmanaged URL collision", () => {
  const { monitorsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;
  const { web } = productionEntries(monitorsDocument);

  const { operations, errors } = planSyncOperations({
    inventoryMonitors: [web],
    remoteMonitors: {
      7: httpMonitorFixture({
        id: 7,
        name: "Wrong production web name",
        url: LIVE_PRODUCTION_WEB.url,
      }),
    },
    ownershipMarker,
  });

  assert.equal(operations.length, 0);
  assert.ok(
    errors.some((error) => error.includes("collision: unmanaged monitor at")),
    `expected URL collision, got: ${errors.join("; ")}`,
  );
  assert.ok(
    !errors.join("; ").includes("KUMA_PASSWORD"),
    "collision errors must not include secrets",
  );
});

test("VOC-087-TEST-04: second sync is a no-op after production adoption", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const { api } = productionEntries(monitorsDocument);

  const initial = {
    1: httpMonitorFixture({
      id: 1,
      name: LIVE_PRODUCTION_WEB.name,
      url: LIVE_PRODUCTION_WEB.url,
    }),
    2: httpMonitorFixture({
      id: 2,
      name: LIVE_PRODUCTION_API.name,
      url: LIVE_PRODUCTION_API.url,
      keyword: api.expected_body,
    }),
  };

  const client = createMockKumaClient(initial);
  const first = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });
  assert.ok(first.changed > 0);

  const second = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });
  assert.equal(second.changed, 0);
  assert.equal(mutatingCalls(client).length, first.changed);
});

test("VOC-087-TEST-05: unrelated manual monitors are not mutated", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const { api } = productionEntries(monitorsDocument);

  const initial = {
    1: httpMonitorFixture({
      id: 1,
      name: LIVE_PRODUCTION_WEB.name,
      url: LIVE_PRODUCTION_WEB.url,
    }),
    2: httpMonitorFixture({
      id: 2,
      name: LIVE_PRODUCTION_API.name,
      url: LIVE_PRODUCTION_API.url,
      keyword: api.expected_body,
    }),
    99: httpMonitorFixture({
      id: 99,
      name: "Founder manual check",
      url: "https://example.com/manual-only",
      description: "founder-owned",
    }),
  };

  const client = createMockKumaClient(initial);
  await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });

  const manualMutations = client.calls.filter(
    (call) =>
      (call.op === "editMonitor" && Number(call.monitor.id) === 99) ||
      (call.op === "deleteMonitor" && Number(call.monitorID) === 99),
  );
  assert.equal(manualMutations.length, 0);
});

test("VOC-087-TEST-13: VOC-086 T01 evidence records VOC-087-D00 live adoption identity", () => {
  const evidence = readFileSync(
    path.join(
      repositoryRoot,
      "specs/changes/VOC-086-manage-monitoring-inventory/t01-evidence.md",
    ),
    "utf8",
  );

  assert.match(evidence, /VocaNova Production Web/);
  assert.match(evidence, /VocaNova Production API/);
  assert.match(evidence, /https:\/\/production\.vocanova\.site/);
  assert.match(evidence, /https:\/\/api-production\.vocanova\.site\/healthz/);
  assert.match(evidence, /url-compare\.mjs|trailing-slash/i);
  assert.doesNotMatch(
    evidence,
    /inventory values: `Production Web` \/ `https:\/\/production\.vocanova\.site\/` and `Production API \/healthz`/,
  );
});

test("VOC-087-TEST-14: package task does not dispatch live inventory apply", () => {
  const evidence = readFileSync(
    path.join(
      repositoryRoot,
      "specs/changes/VOC-087-make-the-first-repository-managed-kuma-sync-adopt/t00-evidence.md",
    ),
    "utf8",
  );

  assert.match(evidence, /live inventory apply.*deferred|not claimed|blocked/i);
  assert.match(evidence, /VOC-086-T05|first live apply/i);
  assert.doesNotMatch(evidence, /sync_inventory:\s*true/i);
});

test("VOC-087 url normalizer: trailing-slash equivalence only for HTTP(S)", () => {
  assert.equal(
    normalizeMonitorUrl("https://production.vocanova.site/"),
    "https://production.vocanova.site",
  );
  assert.ok(
    monitorUrlsEqual(
      "https://production.vocanova.site",
      "https://production.vocanova.site/",
    ),
  );
  assert.ok(
    !monitorUrlsEqual(
      "https://production.vocanova.site",
      "https://api-production.vocanova.site/healthz",
    ),
  );
  assert.ok(
    !monitorUrlsEqual(
      "https://production.vocanova.site",
      "https://production.vocanova.site/other",
    ),
  );
});
