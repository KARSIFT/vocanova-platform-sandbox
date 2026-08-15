#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { parseMonitoringYaml } from "./validate-inventory.mjs";
import { createSocketKumaClient } from "./kuma-sync/socket-client.mjs";
import {
  formatSyncFailure,
  syncKumaMonitors,
} from "./kuma-sync/sync.mjs";

const modulePath = fileURLToPath(import.meta.url);
const monitoringRoot = path.dirname(modulePath);

async function main() {
  const monitorsPath = path.join(monitoringRoot, "monitors.yaml");
  const syntheticsPath = path.join(monitoringRoot, "synthetics.yaml");

  const monitorsDocument = parseMonitoringYaml(readFileSync(monitorsPath, "utf8"));
  const syntheticsDocument = parseMonitoringYaml(readFileSync(syntheticsPath, "utf8"));

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

    const result = await syncKumaMonitors({
      monitorsDocument,
      syntheticsDocument,
      client,
    });

    process.stdout.write(
      `Kuma sync completed (${result.changed} monitor change(s)).\n`,
    );
  } catch (error) {
    process.stderr.write(`Kuma sync failed: ${formatSyncFailure(error)}\n`);
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
