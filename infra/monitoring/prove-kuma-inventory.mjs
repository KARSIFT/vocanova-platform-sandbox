#!/usr/bin/env node
/**
 * VOC-086-T05 — read-only Socket.IO proof that live Kuma inventory matches
 * infra/monitoring/monitors.yaml. Never mutates monitors or touches SQLite.
 */
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { parseMonitoringYaml } from "./validate-inventory.mjs";
import { createSocketKumaClient } from "./kuma-sync/socket-client.mjs";
import {
  formatProofSummary,
  proveKumaInventory,
} from "./kuma-sync/prove-inventory.mjs";
import { formatSyncFailure } from "./kuma-sync/sync.mjs";

const modulePath = fileURLToPath(import.meta.url);
const monitoringRoot = path.dirname(modulePath);

export {
  proveKumaInventory,
  formatProofSummary,
} from "./kuma-sync/prove-inventory.mjs";

async function main() {
  const monitorsPath = path.join(monitoringRoot, "monitors.yaml");
  const monitorsDocument = parseMonitoringYaml(
    readFileSync(monitorsPath, "utf8"),
  );
  const ownershipMarker =
    monitorsDocument.kuma?.ownership_marker ?? "vocanova:repo-managed";

  const baseUrl = process.env.KUMA_URL ?? "http://127.0.0.1:3001";
  const username = process.env.KUMA_USERNAME;
  const password = process.env.KUMA_PASSWORD;

  if (!username || !password) {
    process.stderr.write(
      "KUMA_USERNAME and KUMA_PASSWORD environment variables are required.\n",
    );
    process.exitCode = 1;
    return;
  }

  let client;
  try {
    client = await createSocketKumaClient({
      baseUrl,
      username,
      password,
    });

    const remoteMonitors = await client.listMonitors();
    const proof = proveKumaInventory({
      monitorsDocument,
      remoteMonitors,
      ownershipMarker,
    });

    process.stdout.write(`${formatProofSummary(proof)}\n`);
    if (proof.failures > 0) {
      process.exitCode = 1;
    }
  } catch (error) {
    process.stderr.write(
      `Kuma inventory proof failed: ${formatSyncFailure(error)}\n`,
    );
    process.exitCode = 1;
  } finally {
    if (client) {
      await client.disconnect();
    }
  }
}

if (process.argv[1] === modulePath) {
  main();
}
