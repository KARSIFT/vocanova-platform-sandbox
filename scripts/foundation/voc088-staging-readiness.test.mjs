// VOC-088-T01 — controlled staging signup readiness synthetic.
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const verifyScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-staging-oauth-start.sh",
);
const verifySelftestPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-staging-oauth-start.selftest.sh",
);
const productionGoPath = path.join(
  repositoryRoot,
  "apps/api/app/api/production.go",
);
const syntheticsPath = path.join(
  repositoryRoot,
  "infra/monitoring/synthetics.yaml",
);
const scheduledWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/scheduled-synthetics.yml",
);

test("VOC-088-TEST-07: /healthz exposes controlled_signup_ready without cohort metadata", () => {
  const productionGo = readFileSync(productionGoPath, "utf8");
  const healthzBlock = productionGo.slice(
    productionGo.indexOf("type HealthzOutput struct"),
    productionGo.indexOf("// KillSwitchStatus"),
  );

  assert.match(productionGo, /ControlledSignupReady bool/);
  assert.match(productionGo, /json:"controlled_signup_ready"/);
  assert.match(productionGo, /func ControlledSignupReady\(/);
  assert.doesNotMatch(
    healthzBlock,
    /json:"[^"]*allowlist/i,
    "healthz JSON must not expose allowlist fields",
  );
});

test("VOC-088-TEST-05/06: staging OAuth harness asserts controlled_signup_ready when OAuth is enabled", () => {
  const verifyScript = readFileSync(verifyScriptPath, "utf8");
  const synthetics = readFileSync(syntheticsPath, "utf8");
  const scheduledWorkflow = readFileSync(scheduledWorkflowPath, "utf8");

  assert.match(verifyScript, /controlled_signup_ready is not true/);
  assert.match(
    verifyScript,
    /healthz response must not expose email-like strings/,
  );
  assert.match(synthetics, /api:GET \/healthz controlled_signup_ready/);
  assert.match(synthetics, /feature:staging-controlled-signup-readiness/);
  assert.match(
    scheduledWorkflow,
    /verify-staging-oauth-start\.sh also asserts \/healthz\.controlled_signup_ready/,
  );
});

test("VOC-088-TEST-05/06: disposable harness covers readiness pass and fail paths", () => {
  assert.ok(
    existsSync(verifySelftestPath),
    "verify-staging-oauth-start.selftest.sh must exist",
  );

  const result = spawnSync("bash", [verifySelftestPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: process.env.PATH,
    },
  });

  assert.equal(
    result.status,
    0,
    `selftest failed:\nstdout: ${result.stdout}\nstderr: ${result.stderr}`,
  );
  assert.match(
    result.stdout,
    /controlled_signup_ready=false fails the enabled readiness check/,
  );
  assert.match(
    result.stdout,
    /controlled_signup_ready=true passes the enabled readiness check/,
  );
});
