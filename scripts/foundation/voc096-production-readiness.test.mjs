// VOC-096-T01 — production controlled-signup readiness synthetic.
// Runs via `node --test scripts/foundation/voc096-production-readiness.test.mjs`.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { loadSyntheticsRegistry } from "../../infra/monitoring/scheduled-synthetics.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const verifyScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-production-oauth-start.sh",
);
const verifySelftestPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-production-oauth-start.selftest.sh",
);
const productionGoPath = path.join(
  repositoryRoot,
  "apps/api/app/api/production.go",
);
const scheduledWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/scheduled-synthetics.yml",
);

const EMAIL_LIKE_PATTERN = /[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}/i;

test("VOC-096-TEST-06/09: production OAuth harness asserts controlled_signup_ready and canonical OAuth start", () => {
  const verifyScript = readFileSync(verifyScriptPath, "utf8");
  const scheduledWorkflow = readFileSync(scheduledWorkflowPath, "utf8");

  assert.match(verifyScript, /controlled_signup_ready is not true/);
  assert.match(
    verifyScript,
    /healthz response must not expose email-like strings/,
  );
  assert.match(verifyScript, /accounts\.google\.com/);
  assert.match(
    verifyScript,
    /https:\/\/api-production\.vocanova\.site\/api\/v1\/auth\/oauth\/google\/callback/,
  );
  assert.match(verifyScript, /OAuth start returned HTTP 200/);
  assert.match(
    scheduledWorkflow,
    /verify-production-oauth-start\.sh also asserts \/healthz\.controlled_signup_ready \(VOC-096-T01\)/,
  );
});

test("VOC-096-TEST-07: synthetic inventory declares production controlled-signup readiness coverage", () => {
  const syntheticsDocument = loadSyntheticsRegistry(repositoryRoot);
  const entry = syntheticsDocument.synthetics.find(
    (item) => item.id === "synthetic.production.oauth-expected-state",
  );

  assert.ok(entry, "production OAuth synthetic must exist in registry");
  assert.ok(
    entry.coverage.includes("api:GET /healthz controlled_signup_ready"),
    "registry must document /healthz controlled_signup_ready coverage",
  );
  assert.ok(
    entry.coverage.includes("feature:production-controlled-signup-readiness"),
    "registry must document production controlled-signup readiness feature",
  );
  assert.ok(
    entry.coverage.includes("feature:production-google-oauth"),
    "registry must retain production Google OAuth feature coverage",
  );
  assert.equal(
    entry.harness_script,
    "infra/scripts/verify-production-oauth-start.sh",
  );
});

test("VOC-096-TEST-08: /healthz exposes controlled_signup_ready without cohort metadata", () => {
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

test("VOC-096-TEST-06/09: disposable harness covers readiness pass and fail paths without email metadata", () => {
  assert.ok(
    existsSync(verifySelftestPath),
    "verify-production-oauth-start.selftest.sh must exist",
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
  assert.doesNotMatch(
    result.stdout,
    EMAIL_LIKE_PATTERN,
    "selftest success output must not contain email-like strings",
  );
  assert.doesNotMatch(
    result.stderr,
    EMAIL_LIKE_PATTERN,
    "selftest stderr must not contain email-like strings",
  );
});
