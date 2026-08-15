// VOC-086-T01 — Kuma Socket.IO synchronizer protocol mocks and static guards.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { EventEmitter } from "node:events";

import {
  applyOperations,
  SyncApplyError,
} from "../../infra/monitoring/kuma-sync/apply.mjs";
import { inventoryEntryToDesiredMonitor } from "../../infra/monitoring/kuma-sync/monitor-payload.mjs";
import { createMockKumaClient } from "../../infra/monitoring/kuma-sync/mock-client.mjs";
import { planSyncOperations } from "../../infra/monitoring/kuma-sync/plan.mjs";
import {
  createRedactingLogger,
  redactSecrets,
} from "../../infra/monitoring/kuma-sync/redact.mjs";
import {
  authenticateAndLoadMonitors,
  createMonitorListWaiter,
} from "../../infra/monitoring/kuma-sync/auth-handshake.mjs";
import {
  formatSyncFailure,
  syncKumaMonitors,
  SyncValidationError,
} from "../../infra/monitoring/kuma-sync/sync.mjs";
import {
  parseMonitoringYaml,
  validateMonitoringInventoryFiles,
} from "../../infra/monitoring/validate-inventory.mjs";

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

function loadInventoryDocuments() {
  const monitorsDocument = parseMonitoringYaml(
    readFileSync(monitorsPath, "utf8"),
  );
  const syntheticsDocument = parseMonitoringYaml(
    readFileSync(syntheticsPath, "utf8"),
  );
  return { monitorsDocument, syntheticsDocument };
}

function mutatingCalls(client) {
  return client.calls.filter((call) =>
    ["add", "editMonitor", "deleteMonitor"].includes(call.op),
  );
}

test("VOC-086-TEST-02: synchronizer creates missing managed monitors", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const client = createMockKumaClient({});

  const result = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });

  assert.equal(result.changed, monitorsDocument.availability_monitors.length);
  const addCalls = client.calls.filter((call) => call.op === "add");
  assert.equal(
    addCalls.length,
    monitorsDocument.availability_monitors.length,
    "each inventory monitor should be created once",
  );
});

test("VOC-086-TEST-03: synchronizer updates changed monitors and is idempotent", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;

  const seeded = {};
  for (const entry of monitorsDocument.availability_monitors) {
    const desired = inventoryEntryToDesiredMonitor(entry, ownershipMarker);
    const id = entry.id.includes("production")
      ? 100 + Object.keys(seeded).length
      : 10 + Object.keys(seeded).length;
    seeded[id] = { ...desired, id };
  }

  const client = createMockKumaClient(seeded);
  const noopResult = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });
  assert.equal(noopResult.changed, 0);
  assert.equal(mutatingCalls(client).length, 0);

  const changedEntry = monitorsDocument.availability_monitors.find(
    (entry) => entry.id === "kuma.availability.staging.web",
  );
  const targetId = Object.values(client.monitors).find((monitor) =>
    monitor.description?.includes(changedEntry.id),
  )?.id;
  client.monitors[targetId] = {
    ...client.monitors[targetId],
    interval: 120,
    retryInterval: 120,
  };

  const updateResult = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });
  assert.equal(updateResult.changed, 1);
  assert.equal(
    client.calls.filter((call) => call.op === "editMonitor").length,
    1,
  );
});

test("VOC-086-TEST-04: adopt production monitors and preserve unrelated manual monitors", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;

  const productionWeb = monitorsDocument.availability_monitors.find(
    (entry) => entry.id === "kuma.availability.production.web",
  );
  const productionApi = monitorsDocument.availability_monitors.find(
    (entry) => entry.id === "kuma.availability.production.api-healthz",
  );

  const initial = {
    1: {
      id: 1,
      type: "http",
      name: productionWeb.name,
      url: productionWeb.url,
      method: "GET",
      interval: 60,
      timeout: 30,
      maxretries: 0,
      retryInterval: 60,
      keyword: "",
      accepted_statuscodes: ["200-299"],
      description: "manual legacy monitor",
      active: true,
      conditions: [],
    },
    2: {
      id: 2,
      type: "http",
      name: productionApi.name,
      url: productionApi.url,
      method: "GET",
      interval: 60,
      timeout: 30,
      maxretries: 0,
      retryInterval: 60,
      keyword: productionApi.expected_body,
      accepted_statuscodes: ["200-299"],
      description: "",
      active: true,
      conditions: [],
    },
    99: {
      id: 99,
      type: "http",
      name: "Founder manual check",
      url: "https://example.com/manual-only",
      method: "GET",
      interval: 300,
      timeout: 30,
      maxretries: 0,
      retryInterval: 300,
      keyword: "",
      accepted_statuscodes: ["200-299"],
      description: "founder-owned",
      active: true,
      conditions: [],
    },
  };

  const client = createMockKumaClient(initial);
  const result = await syncKumaMonitors({
    monitorsDocument,
    syntheticsDocument,
    client,
    logger: { info() {}, error() {} },
  });

  assert.ok(result.changed > 0);

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

test("VOC-086-TEST-05: invalid inventory blocks mutation before apply", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const brokenMonitors = structuredClone(monitorsDocument);
  brokenMonitors.availability_monitors[0] = {
    ...brokenMonitors.availability_monitors[0],
    coverage: [],
  };

  const client = createMockKumaClient({});
  await assert.rejects(
    () =>
      syncKumaMonitors({
        monitorsDocument: brokenMonitors,
        syntheticsDocument,
        client,
        logger: { info() {}, error() {} },
      }),
    SyncValidationError,
  );

  assert.equal(mutatingCalls(client).length, 0);
});

test("VOC-086-TEST-06: partial apply compensates and incomplete rollback fails non-zero", async () => {
  const { monitorsDocument, syntheticsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;

  const partialInventory = {
    ...monitorsDocument,
    availability_monitors: monitorsDocument.availability_monitors.slice(0, 2),
  };

  const client = createMockKumaClient({});
  client.setFailAfterApplied({
    after: 1,
    operation: "add",
    message: "simulated mid-apply add failure",
  });

  await assert.rejects(
    () =>
      syncKumaMonitors({
        monitorsDocument: partialInventory,
        syntheticsDocument,
        client,
        requireCanonicalIds: false,
        logger: { info() {}, error() {} },
      }),
    SyncApplyError,
  );

  assert.equal(Object.keys(client.monitors).length, 0);
  assert.ok(
    client.calls.some((call) => call.op === "deleteMonitor"),
    "compensation should delete the monitor created before the failure",
  );

  const rollbackClient = createMockKumaClient({});
  rollbackClient.setFailAfterApplied({
    after: 1,
    operation: "add",
    message: "simulated mid-apply add failure",
  });
  rollbackClient.setFailDeleteMonitor(true);

  const operations = partialInventory.availability_monitors.map((entry) => ({
    type: "create",
    inventoryId: entry.id,
    desired: inventoryEntryToDesiredMonitor(entry, ownershipMarker),
  }));

  let rollbackError;
  try {
    await applyOperations({
      client: rollbackClient,
      operations,
      logger: { info() {}, error() {} },
    });
  } catch (error) {
    rollbackError = error;
  }

  assert.ok(rollbackError instanceof SyncApplyError);
  assert.match(
    formatSyncFailure(rollbackError),
    /rollback was incomplete|simulated deleteMonitor rollback failure/,
  );
});

test("VOC-086-TEST-07: sync tooling does not reference SQLite deployment paths", () => {
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

test("VOC-086-TEST-09 (T01 subset): credential-like log fixtures and formatSyncFailure are redacted", () => {
  const fixture =
    "login failed password=super-secret-token KUMA_PASSWORD=abc123 Bearer eyJhbGciOiJIUzI1NiJ9";
  const redacted = redactSecrets(fixture);
  assert.ok(!redacted.includes("super-secret-token"));
  assert.ok(!redacted.includes("abc123"));
  assert.ok(!redacted.includes("eyJhbGciOiJIUzI1NiJ9"));
  assert.match(redacted, /\[REDACTED\]/);

  const lines = [];
  const logger = createRedactingLogger({
    info(message) {
      lines.push(message);
    },
    error(message) {
      lines.push(message);
    },
  });
  logger.error("KUMA_PASSWORD=plain-secret-value");
  assert.ok(!lines[0].includes("plain-secret-value"));

  const formatted = formatSyncFailure(
    new Error("connect failed password=live-cli-secret-token"),
  );
  assert.ok(!formatted.includes("live-cli-secret-token"));
  assert.match(formatted, /\[REDACTED\]/);
});

test("VOC-086-TEST-02 (client): monitorList pushed before login ack is still captured", async () => {
  const socket = new EventEmitter();
  socket.emit = function emitWithLoginRace(event, payload, callback) {
    if (event === "login") {
      // Simulate Uptime Kuma pushing monitorList before the login ack returns.
      EventEmitter.prototype.emit.call(this, "monitorList", {
        7: { id: 7, name: "seeded" },
      });
      callback({ ok: true });
      return;
    }
    return EventEmitter.prototype.emit.call(this, event, payload, callback);
  };

  const monitors = await authenticateAndLoadMonitors(socket, {
    username: "admin",
    password: "not-a-real-password",
    timeoutMs: 1000,
  });

  assert.equal(monitors[7].name, "seeded");
});

test("VOC-086-TEST-02 (client): failed login cancels monitorList waiter and does not hang", async () => {
  const socket = new EventEmitter();
  socket.emit = function emitLoginFailure(event, _payload, callback) {
    if (event === "login") {
      callback({ ok: false, msg: "password=should-not-leak-in-hang" });
      return;
    }
    return EventEmitter.prototype.emit.call(this, event, _payload, callback);
  };

  await assert.rejects(
    () =>
      authenticateAndLoadMonitors(socket, {
        username: "admin",
        password: "not-a-real-password",
        timeoutMs: 5000,
      }),
    /password=should-not-leak-in-hang|Kuma login failed/,
  );

  const waiter = createMonitorListWaiter(socket, 50);
  waiter.cancel(new Error("cancelled for isolation"));
  await assert.rejects(() => waiter.promise, /cancelled for isolation/);
});

test("VOC-086-TEST-04: plan detects duplicate managed monitor_id collisions", () => {
  const { monitorsDocument } = loadInventoryDocuments();
  const ownershipMarker = monitorsDocument.kuma.ownership_marker;
  const entry = monitorsDocument.availability_monitors[0];
  const desired = inventoryEntryToDesiredMonitor(entry, ownershipMarker);

  const { errors } = planSyncOperations({
    inventoryMonitors: monitorsDocument.availability_monitors,
    remoteMonitors: {
      1: { ...desired, id: 1 },
      2: { ...desired, id: 2 },
    },
    ownershipMarker,
  });

  assert.ok(
    errors.some((error) => error.includes("duplicate managed monitor_id")),
    `expected collision error, got: ${errors.join("; ")}`,
  );
});

test("VOC-086-TEST-00 guard: repository inventory still validates before sync tests", () => {
  const errors = validateMonitoringInventoryFiles(repositoryRoot);
  assert.deepEqual(errors, [], errors.join("; "));
});
