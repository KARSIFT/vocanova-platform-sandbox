// VOC-081-T04 — live verification tooling for monitor.vocanova.site.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.
//
// These tests assert repository-side harness + evidence structure. They do
// NOT substitute for live AC-05/AC-03/AC-06 closure: t04-evidence.md must
// record a successful qualifying deploy and passing external probes.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const verifyScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-voc081-monitor.sh",
);
const verifySelftestPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-voc081-monitor.selftest.sh",
);
const cutoverScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-voc067-cutover.sh",
);
const accessPolicyPath = path.join(
  repositoryRoot,
  "infra/monitoring/access-policy.md",
);
const t04EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-081-route-monitor-vocanova-site-through-the/t04-evidence.md",
);
const errorMonitoringWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/error-monitoring.yml",
);
const monitoringComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.monitoring.yml",
);
const deployStagingPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);

const MONITOR_HOSTNAME = "monitor.vocanova.site";

test("VOC-081-TEST-07: live verification script covers HTTPS, WebSocket, and access policy", () => {
  assert.ok(
    existsSync(verifyScriptPath),
    "verify-voc081-monitor.sh must exist",
  );
  assert.ok(
    existsSync(verifySelftestPath),
    "verify-voc081-monitor.selftest.sh must exist",
  );

  const verifyScript = readFileSync(verifyScriptPath, "utf8");

  assert.match(
    verifyScript,
    new RegExp(MONITOR_HOSTNAME.replaceAll(".", "\\.")),
    "verification script must target monitor.vocanova.site",
  );
  assert.match(
    verifyScript,
    /transport=polling|socket\.io/,
    "verification script must probe the Kuma socket.io path",
  );
  assert.match(
    verifyScript,
    /Upgrade|polling handshake|sid/,
    "verification script must verify WebSocket/Engine.IO readiness",
  );
  assert.match(
    verifyScript,
    /login|password|entry-page|SPA/i,
    "verification script must check for Kuma login/SPA boundary (DEP-00)",
  );
  assert.match(
    verifyScript,
    /verify-voc067-cutover\.sh/,
    "verification script must guard the four app hostnames",
  );
  assert.match(
    verifyScript,
    /520|502/,
    "verification script must treat Cloudflare/origin 520/502 as failure",
  );
  assert.match(
    verifyScript,
    /-L|--location|follow/i,
    "verification script must follow Kuma / → /dashboard redirects",
  );
});

test("VOC-081-TEST-08: T04 evidence records deploy, topology, and rollback fields", () => {
  assert.ok(existsSync(t04EvidencePath), "t04-evidence.md must exist");

  const evidence = readFileSync(t04EvidencePath, "utf8");

  assert.match(
    evidence,
    /rollback_owner/i,
    "T04 evidence must name rollback owner",
  );
  assert.match(
    evidence,
    /last_known_good|last-known-good/i,
    "T04 evidence must cite last-known-good SHA",
  );
  assert.match(
    evidence,
    /monitoring window/i,
    "T04 evidence must document monitoring window",
  );
  assert.match(
    evidence,
    /verify-voc081-monitor\.sh/,
    "T04 evidence must reference the live verification script",
  );
  assert.match(
    evidence,
    /actions\/runs\/\d+/,
    "T04 evidence must record a qualifying deploy run URL",
  );
  assert.match(
    evidence,
    /qualifying_deploy_status:/,
    "T04 evidence front matter must declare qualifying_deploy_status",
  );
  assert.match(
    evidence,
    /gate_status:/,
    "T04 evidence front matter must declare gate_status",
  );

  // Structural green must not be confused with live AC closure. When the
  // gate claims live closure, evidence must also show a successful deploy
  // and a passing monitor probe.
  const gateMatch = evidence.match(/^gate_status:\s*(\S+)/m);
  assert.ok(gateMatch, "gate_status front-matter value is required");
  const gateStatus = gateMatch[1];
  if (
    gateStatus === "live-closure-complete" ||
    gateStatus === "pass" ||
    gateStatus === "PASS"
  ) {
    assert.match(
      evidence,
      /qualifying_deploy_status:\s*(success|successful)/,
      "live-closure gate requires qualifying_deploy_status success",
    );
    assert.match(
      evidence,
      /monitor web[^\n]*HTTP 200|https:\/\/monitor\.vocanova\.site\/[^\n]*200/i,
      "live-closure gate requires a recorded HTTP 200 for the monitor hostname",
    );
  }
});

test("VOC-081-TEST-09: error-monitoring workflow unchanged by VOC-081", () => {
  assert.ok(
    existsSync(errorMonitoringWorkflowPath),
    "error-monitoring.yml must exist",
  );

  const workflow = readFileSync(errorMonitoringWorkflowPath, "utf8");
  assert.match(
    workflow,
    /VOC-051/,
    "error-monitoring workflow must remain the VOC-051 hourly Sentry monitor",
  );

  const accessPolicy = readFileSync(accessPolicyPath, "utf8");
  assert.match(
    accessPolicy,
    /verify-voc081-monitor\.sh/,
    "access policy must point operators at the T04 verification script",
  );
});

test("VOC-081-TEST-07: cutover regression script remains available", () => {
  assert.ok(
    existsSync(cutoverScriptPath),
    "verify-voc067-cutover.sh must remain available for app-tier regression checks",
  );
});

test("VOC-081-T04 remediation: monitoring healthcheck invokes JS probe through Node", () => {
  const monitoring = readFileSync(monitoringComposePath, "utf8");
  assert.match(
    monitoring,
    /test:\s*\["CMD",\s*"node",\s*"extra\/healthcheck\.js"\]/,
    "Kuma healthcheck.js has no shebang and must be invoked through node",
  );
  assert.doesNotMatch(
    monitoring,
    /test:\s*\["CMD",\s*"extra\/healthcheck\.js"\]/,
    "bare CMD extra/healthcheck.js must not be used (exec format / shell failure)",
  );
});

test("VOC-081-T04 remediation: staging deploy accepts loopback HTTP as Kuma ready", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  assert.match(
    deployStaging,
    /curl -fsS --max-time 3 -o \/dev\/null http:\/\/127\.0\.0\.1:3001\//,
    "deploy must fall back to loopback HTTP when Docker health lags",
  );
  assert.match(
    deployStaging,
    /for attempt in \$\(seq 1 120\)/,
    "deploy must wait long enough for Kuma start_period + health",
  );
});
