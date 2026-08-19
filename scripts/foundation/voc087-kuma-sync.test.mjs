// VOC-087 — Kuma sync adoption identity, URL normalization, and notification preservation tests.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { createMockKumaClient } from "../../infra/monitoring/kuma-sync/mock-client.mjs";
import {
  inventoryEntryToDesiredMonitor,
  inventoryOwnsNotificationBindings,
} from "../../infra/monitoring/kuma-sync/monitor-payload.mjs";
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

function httpMonitorFixture({
  id,
  name,
  url,
  keyword = "",
  description = "",
  notificationIDList,
}) {
  const monitor = {
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
  if (notificationIDList !== undefined) {
    monitor.notificationIDList = notificationIDList;
  }
  return monitor;
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

test("VOC-087-TEST-00A: supported Kuma 1.x payload omits newer conditions field", () => {
  const { monitorsDocument } = loadInventoryDocuments();
  const entry = monitorsDocument.availability_monitors[0];
  const desired = inventoryEntryToDesiredMonitor(
    entry,
    monitorsDocument.kuma.ownership_marker,
  );

  assert.ok(!("conditions" in desired));
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

function editPayloads(client) {
  return client.calls
    .filter((call) => call.op === "editMonitor")
    .map((call) => call.monitor);
}

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

test("VOC-087-TEST-04 (notification): second sync is a no-op with preserved bindings", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const { api } = productionEntries(monitorsDocument);
  const syntheticBindings = { 42: true, 99: true };

  const initial = {
    1: httpMonitorFixture({
      id: 1,
      name: LIVE_PRODUCTION_WEB.name,
      url: LIVE_PRODUCTION_WEB.url,
      notificationIDList: syntheticBindings,
    }),
    2: httpMonitorFixture({
      id: 2,
      name: LIVE_PRODUCTION_API.name,
      url: LIVE_PRODUCTION_API.url,
      keyword: api.expected_body,
      notificationIDList: { 7: true },
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

  for (const payload of editPayloads(client)) {
    assert.deepEqual(
      payload.notificationIDList,
      initial[payload.id].notificationIDList,
      "adopt/update must preserve remote notification bindings",
    );
  }

  assert.deepEqual(client.monitors[1].notificationIDList, syntheticBindings);
  assert.deepEqual(client.monitors[2].notificationIDList, { 7: true });

  const second = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });
  assert.equal(second.changed, 0);
  assert.equal(mutatingCalls(client).length, first.changed);
});

test("VOC-087-TEST-06: adopt/update edit payloads retain remote notification bindings", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const { api } = productionEntries(monitorsDocument);
  const webBindings = { 11: true };
  const apiBindings = { 22: true };

  const initial = {
    1: httpMonitorFixture({
      id: 1,
      name: LIVE_PRODUCTION_WEB.name,
      url: LIVE_PRODUCTION_WEB.url,
      notificationIDList: webBindings,
    }),
    2: httpMonitorFixture({
      id: 2,
      name: LIVE_PRODUCTION_API.name,
      url: LIVE_PRODUCTION_API.url,
      keyword: api.expected_body,
      notificationIDList: apiBindings,
    }),
  };

  const client = createMockKumaClient(initial);
  await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });

  const payloads = editPayloads(client);
  assert.ok(payloads.length >= 2);
  assert.deepEqual(
    payloads.find((payload) => Number(payload.id) === 1)?.notificationIDList,
    webBindings,
  );
  assert.deepEqual(
    payloads.find((payload) => Number(payload.id) === 2)?.notificationIDList,
    apiBindings,
  );
  assert.ok(
    !payloads.some(
      (payload) =>
        payload.notificationIDList &&
        Object.keys(payload.notificationIDList).length === 0,
    ),
    "edit payloads must not default to clearing notification bindings",
  );
});

test("VOC-087-TEST-07: create sends Kuma's required empty notification mapping", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;
  const { web } = productionEntries(monitorsDocument);

  const desired = inventoryEntryToDesiredMonitor(web, ownershipMarker);
  assert.deepEqual(
    desired.notificationIDList,
    {},
    "a create must satisfy Kuma's Socket.IO contract without inventing destinations",
  );

  const client = createMockKumaClient({});
  await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    requireCanonicalIds: false,
    logger: { info() {}, error() {} },
  });

  const addCalls = client.calls.filter((call) => call.op === "add");
  assert.ok(addCalls.length > 0);
  for (const call of addCalls) {
    assert.deepEqual(
      call.monitor?.notificationIDList,
      {},
      "create must send the empty mapping required by Kuma",
    );
  }
});

test("VOC-087-TEST-08: explicit inventory notification ownership is honored", () => {
  const { monitorsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;
  const { web } = productionEntries(monitorsDocument);
  const fixtureEntry = {
    ...web,
    notification_id_list: { 5: true },
  };
  const remoteMonitor = httpMonitorFixture({
    id: 1,
    name: web.name,
    url: web.url,
    notificationIDList: { 9: true },
  });

  const desired = inventoryEntryToDesiredMonitor(
    fixtureEntry,
    ownershipMarker,
    {
      remoteMonitor,
    },
  );

  assert.ok(inventoryOwnsNotificationBindings(fixtureEntry));
  assert.deepEqual(desired.notificationIDList, { 5: true });
  assert.notDeepEqual(
    desired.notificationIDList,
    remoteMonitor.notificationIDList,
  );

  const { operations } = planSyncOperations({
    inventoryMonitors: [fixtureEntry],
    remoteMonitors: { 1: remoteMonitor },
    ownershipMarker,
  });

  assert.equal(operations.length, 1);
  assert.equal(operations[0].type, "update");
  assert.deepEqual(operations[0].desired.notificationIDList, { 5: true });

  const otherwiseMatchingRemote = {
    ...inventoryEntryToDesiredMonitor(web, ownershipMarker, {
      remoteMonitor: { notificationIDList: { 9: true } },
    }),
    id: 2,
  };
  const ownershipOnly = planSyncOperations({
    inventoryMonitors: [fixtureEntry],
    remoteMonitors: { 2: otherwiseMatchingRemote },
    ownershipMarker,
  });
  assert.equal(
    ownershipOnly.operations.length,
    1,
    "an explicitly owned binding change must cause an update by itself",
  );
  assert.equal(ownershipOnly.operations[0].type, "update");
  assert.deepEqual(ownershipOnly.operations[0].desired.notificationIDList, {
    5: true,
  });
});

test("VOC-087-TEST-08 (fail-closed): invalid notification ownership is rejected before connect", async () => {
  const invalidValues = [
    null,
    [],
    "5",
    { 5: "true" },
    { 0: true },
    { destination: true },
  ];

  for (const invalidValue of invalidValues) {
    const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
    monitorsDocument.availability_monitors[0].notification_id_list =
      invalidValue;
    let connectCalls = 0;
    const client = {
      async connect() {
        connectCalls += 1;
      },
      async listMonitors() {
        throw new Error("inventory validation must run before remote access");
      },
    };

    await assert.rejects(
      syncKumaMonitors({
        monitorsDocument,
        syntheticsDocument,
        client,
        logger: { info() {}, error() {} },
      }),
      /notification_id_list/,
    );
    assert.equal(connectCalls, 0);
  }
});

test("VOC-087-TEST-11: sync tooling does not reference SQLite deployment paths", () => {
  const syncFiles = [
    "infra/monitoring/sync-kuma.mjs",
    "infra/monitoring/kuma-sync/apply.mjs",
    "infra/monitoring/kuma-sync/auth-handshake.mjs",
    "infra/monitoring/kuma-sync/mock-client.mjs",
    "infra/monitoring/kuma-sync/monitor-metadata.mjs",
    "infra/monitoring/kuma-sync/monitor-payload.mjs",
    "infra/monitoring/kuma-sync/plan.mjs",
    "infra/monitoring/kuma-sync/redact.mjs",
    "infra/monitoring/kuma-sync/socket-client.mjs",
    "infra/monitoring/kuma-sync/sync.mjs",
  ];

  const forbidden = [/kuma\.db/i, /\bsqlite\b/i, /\/app\/data/i];

  for (const relativePath of syncFiles) {
    const source = readFileSync(
      path.join(repositoryRoot, relativePath),
      "utf8",
    );
    for (const pattern of forbidden) {
      assert.ok(
        !pattern.test(source),
        `${relativePath} must not reference SQLite deployment paths (${pattern})`,
      );
    }
  }
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
      "specs/changes/VOC-087-make-the-first-repository-managed-kuma-sync-adopt/t01-evidence.md",
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
