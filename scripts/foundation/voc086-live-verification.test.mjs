// VOC-086-T05 — live verification tooling and evidence structure.
//
// Repository-side harness tests. VOC-086-TEST-10 / TEST-12 live procedures
// are recorded only in t05-evidence.md when the corresponding
// live_*_claimed flags are true. These tests never treat fixture coverage
// as live Socket.IO or scheduled-synthetic closure.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { proveKumaInventory } from "../../infra/monitoring/kuma-sync/prove-inventory.mjs";
import {
  CANONICAL_AVAILABILITY_MONITOR_IDS,
  CANONICAL_SYNTHETIC_IDS,
  parseMonitoringYaml,
} from "../../infra/monitoring/validate-inventory.mjs";
import { inventoryEntryToDesiredMonitor } from "../../infra/monitoring/kuma-sync/monitor-payload.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const verifyScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-voc086-monitoring.sh",
);
const verifySelftestPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-voc086-monitoring.selftest.sh",
);
const proveScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/prove-kuma-inventory.sh",
);
const proveModulePath = path.join(
  repositoryRoot,
  "infra/monitoring/prove-kuma-inventory.mjs",
);
const syncScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/sync-kuma-inventory.sh",
);
const monitorsPath = path.join(
  repositoryRoot,
  "infra/monitoring/monitors.yaml",
);
const syntheticsPath = path.join(
  repositoryRoot,
  "infra/monitoring/synthetics.yaml",
);
const monitoringDocPath = path.join(
  repositoryRoot,
  "docs/operations/monitoring.md",
);
const t05EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-086-manage-monitoring-inventory/t05-evidence.md",
);
const syncWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/sync-monitoring.yml",
);
const syntheticsWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/scheduled-synthetics.yml",
);
const errorMonitoringWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/error-monitoring.yml",
);

const OWNERSHIP_MARKER = "vocanova:repo-managed";

function frontMatterFlag(source, name) {
  const match = source.match(new RegExp(`^${name}:\\s*(\\S+)`, "m"));
  return match?.[1] ?? "";
}

test("VOC-086-T05 harness: prove tooling is read-only Socket.IO and bans SQLite", () => {
  for (const filePath of [
    verifyScriptPath,
    verifySelftestPath,
    proveScriptPath,
    proveModulePath,
    monitoringDocPath,
  ]) {
    assert.ok(existsSync(filePath), `${filePath} must exist`);
  }

  const proveScript = readFileSync(proveScriptPath, "utf8");
  assert.match(
    proveScript,
    /prove-kuma-inventory\.mjs/,
    "host prove script must invoke prove-kuma-inventory.mjs",
  );
  assert.doesNotMatch(
    proveScript,
    /kuma\.db|\bsqlite\b|\/app\/data/i,
    "prove script must not reference SQLite deployment paths",
  );

  const syncScript = readFileSync(syncScriptPath, "utf8");
  const proveIndex = syncScript.indexOf("node prove-kuma-inventory.mjs");
  const syncIndex = syncScript.indexOf("node sync-kuma.mjs");
  assert.ok(syncIndex >= 0, "sync-kuma-inventory.sh must invoke sync-kuma.mjs");
  assert.ok(
    proveIndex > syncIndex,
    "read-only Socket.IO proof must run after inventory apply in the same host script",
  );

  const verifyScript = readFileSync(verifyScriptPath, "utf8");
  for (const monitorId of CANONICAL_AVAILABILITY_MONITOR_IDS) {
    assert.match(
      verifyScript,
      new RegExp(monitorId.replaceAll(".", "\\.")),
      `verify script must reference ${monitorId}`,
    );
  }
  assert.match(
    verifyScript,
    /verify-voc081-monitor\.sh/,
    "verify script must reuse VOC-081 monitor-host verifier",
  );
});

test("VOC-086-T05 harness: proveKumaInventory passes when live monitors match inventory", () => {
  const monitorsDocument = parseMonitoringYaml(
    readFileSync(monitorsPath, "utf8"),
  );
  const remoteMonitors = {};

  for (const [
    index,
    entry,
  ] of monitorsDocument.availability_monitors.entries()) {
    const desired = inventoryEntryToDesiredMonitor(entry, OWNERSHIP_MARKER);
    remoteMonitors[String(index + 1)] = { id: index + 1, ...desired };
  }

  const proof = proveKumaInventory({
    monitorsDocument,
    remoteMonitors,
    ownershipMarker: OWNERSHIP_MARKER,
  });

  assert.equal(proof.failures, 0);
  assert.equal(proof.results.length, CANONICAL_AVAILABILITY_MONITOR_IDS.length);
  assert.ok(proof.results.every((result) => result.status === "pass"));
});

test("VOC-086-T05 harness: proveKumaInventory fails when a canonical monitor is missing", () => {
  const monitorsDocument = parseMonitoringYaml(
    readFileSync(monitorsPath, "utf8"),
  );
  const remoteMonitors = {};
  const first = monitorsDocument.availability_monitors[0];
  const desired = inventoryEntryToDesiredMonitor(first, OWNERSHIP_MARKER);
  remoteMonitors["1"] = { id: 1, ...desired };

  const proof = proveKumaInventory({
    monitorsDocument,
    remoteMonitors,
    ownershipMarker: OWNERSHIP_MARKER,
  });

  assert.ok(proof.failures > 0);
  assert.ok(
    proof.results.some(
      (result) =>
        result.status === "fail" &&
        result.inventoryId !== first.id &&
        /not present/i.test(result.reason),
    ),
  );
});

test("VOC-086-TEST-17 (harness): verification script covers monitor-host and topology branches", () => {
  const verifyScript = readFileSync(verifyScriptPath, "utf8");
  assert.match(
    verifyScript,
    /voc081-monitoring-topology/,
    "verify script must run topology repository assertions",
  );
  assert.match(
    verifyScript,
    /prove-kuma-inventory\.sh/,
    "verify script must optionally invoke Socket.IO proof",
  );
  assert.match(
    verifyScript,
    /monitor\.vocanova\.site/,
    "verify script must probe monitor hostname reachability",
  );
  assert.match(verifyScript, /8081/, "verify script must probe retired :8081");
  assert.match(verifyScript, /8443/, "verify script must probe retired :8443");
});

test("VOC-086-TEST-13: error-monitoring.yml remains the Sentry path", () => {
  const workflow = readFileSync(errorMonitoringWorkflowPath, "utf8");
  const syntheticsWorkflow = readFileSync(syntheticsWorkflowPath, "utf8");
  assert.match(workflow, /Sentry|sentry/i);
  assert.match(
    syntheticsWorkflow,
    /error-monitoring\.yml — Sentry error monitoring \(unchanged\)/,
    "scheduled synthetics must document Sentry as the separate error channel",
  );
  assert.doesNotMatch(
    syntheticsWorkflow,
    /uses:.*error-monitoring|workflow_call:.*error-monitoring/i,
    "scheduled synthetics must not invoke or replace error-monitoring",
  );
});

test("VOC-086-T05 harness: workflows and synthetics registry remain wired", () => {
  const syntheticsDocument = parseMonitoringYaml(
    readFileSync(syntheticsPath, "utf8"),
  );
  const workflow = readFileSync(syntheticsWorkflowPath, "utf8");
  const syncWorkflow = readFileSync(syncWorkflowPath, "utf8");

  for (const entry of syntheticsDocument.synthetics) {
    assert.match(
      workflow,
      new RegExp(entry.id.replaceAll(".", "\\.")),
      `scheduled workflow must name synthetic ${entry.id}`,
    );
  }

  assert.match(syncWorkflow, /sync-kuma-inventory\.sh/);
  assert.match(syncWorkflow, /rotate_credentials/);
  assert.match(syncWorkflow, /prove-kuma-inventory\.sh/);
});

test("VOC-086-TEST-10 live claim is evidence-gated (not satisfied by this harness)", () => {
  assert.ok(existsSync(t05EvidencePath), "t05-evidence.md must exist");
  const evidence = readFileSync(t05EvidencePath, "utf8");
  const claimed = frontMatterFlag(evidence, "live_socket_proof_claimed");

  assert.match(evidence, /prove-kuma-inventory/);
  assert.match(evidence, /sync-monitoring\.yml/);

  if (claimed === "true") {
    assert.match(
      evidence,
      /actions\/runs\/\d+/,
      "live Socket.IO proof requires a recorded Actions run URL",
    );
    assert.match(
      evidence,
      /sync-monitoring/,
      "live Socket.IO proof requires a sync-monitoring run",
    );
    for (const monitorId of CANONICAL_AVAILABILITY_MONITOR_IDS) {
      assert.match(
        evidence,
        new RegExp(`PASS: ${monitorId.replaceAll(".", "\\.")}`),
        `live Socket.IO proof must record PASS for ${monitorId}`,
      );
    }
  }
});

test("VOC-086-TEST-12 live claim is evidence-gated (mint/mask wiring remains T03)", () => {
  const evidence = readFileSync(t05EvidencePath, "utf8");
  const claimed = frontMatterFlag(evidence, "live_synthetics_claimed");
  assert.match(evidence, /scheduled-synthetics\.yml/);

  if (claimed === "true") {
    assert.match(
      evidence,
      /actions\/runs\/\d+/,
      "live synthetics proof requires a recorded Actions run URL",
    );
    for (const syntheticId of CANONICAL_SYNTHETIC_IDS) {
      assert.match(
        evidence,
        new RegExp(syntheticId.replaceAll(".", "\\.")),
        `live synthetics proof must name ${syntheticId}`,
      );
    }
  }
});

test("VOC-086-T05 evidence records verification commands and rollback owner", () => {
  const evidence = readFileSync(t05EvidencePath, "utf8");
  assert.match(evidence, /rollback_owner/i);
  assert.match(evidence, /verify-voc086-monitoring\.sh/);
  assert.match(evidence, /gate_status:/);
  assert.match(evidence, /docs\/operations\/monitoring\.md/);
  assert.match(evidence, /live_socket_proof_claimed:/);
  assert.match(evidence, /live_synthetics_claimed:/);
});
