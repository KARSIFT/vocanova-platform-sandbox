// VOC-086-T02 — sync-monitoring workflow wiring, credential bootstrap guards,
// and redaction/static checks.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { redactSecrets } from "../../infra/monitoring/kuma-sync/redact.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const workflowPath = path.join(
  repositoryRoot,
  ".github/workflows/sync-monitoring.yml",
);
const rotateScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/kuma-rotate-credentials.sh",
);
const syncScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/sync-kuma-inventory.sh",
);
const rotateSelftestPath = path.join(
  repositoryRoot,
  "infra/scripts/kuma-rotate-credentials.selftest.sh",
);
const syncSelftestPath = path.join(
  repositoryRoot,
  "infra/scripts/sync-kuma-inventory.selftest.sh",
);

function extractStepBlock(workflowSource, stepName) {
  const marker = `- name: ${stepName}`;
  const stepStartIndex = workflowSource.indexOf(marker);
  assert.notEqual(
    stepStartIndex,
    -1,
    `sync-monitoring workflow is missing step: ${stepName}`,
  );

  const nextStepIndex = workflowSource.indexOf(
    "\n      - name:",
    stepStartIndex + 1,
  );
  return workflowSource.slice(
    stepStartIndex,
    nextStepIndex === -1 ? workflowSource.length : nextStepIndex,
  );
}

test("VOC-086-TEST-08: rotate_credentials is opt-in; normal sync never resets", () => {
  const workflow = readFileSync(workflowPath, "utf8");
  const rotateScript = readFileSync(rotateScriptPath, "utf8");
  const syncScript = readFileSync(syncScriptPath, "utf8");

  assert.match(workflow, /rotate_credentials:/);
  assert.match(workflow, /default: false/);

  const rotateStep = extractStepBlock(
    workflow,
    "Rotate Kuma credentials on host (explicit opt-in)",
  );
  assert.match(rotateStep, /if: \$\{\{ inputs\.rotate_credentials \}\}/);
  assert.match(rotateStep, /kuma-rotate-credentials\.sh/);
  assert.match(rotateScript, /extra\/reset-password\.js/);

  const syncOnlyStep = extractStepBlock(
    workflow,
    "Sync monitor inventory to Kuma",
  );
  assert.doesNotMatch(
    syncOnlyStep,
    /reset-password/,
    "normal sync path must not invoke password reset",
  );
  assert.doesNotMatch(
    syncScript,
    /reset-password/,
    "inventory sync script must not reset credentials",
  );

  assert.doesNotMatch(
    rotateScript,
    /--new[-_]password/,
    "rotation must use stdin/container reset tool, not argv password flags",
  );
});

test("VOC-086-TEST-09: credential redaction and bootstrap stores secrets without printing", () => {
  const workflow = readFileSync(workflowPath, "utf8");
  const rotateScript = readFileSync(rotateScriptPath, "utf8");

  const fixture =
    "bootstrap failed password=super-secret KUMA_PASSWORD=abc123 Bearer eyJhbGciOiJIUzI1NiJ9";
  const redacted = redactSecrets(fixture);
  assert.ok(!redacted.includes("super-secret"));
  assert.ok(!redacted.includes("abc123"));
  assert.match(redacted, /\[REDACTED\]/);

  const storePasswordStep = extractStepBlock(
    workflow,
    "Store rotated Kuma password in monitoring environment",
  );
  assert.match(
    storePasswordStep,
    /gh secret set KUMA_PASSWORD --env monitoring/,
  );
  assert.match(storePasswordStep, /--body-file/);
  assert.doesNotMatch(
    storePasswordStep,
    /echo.*PASSWORD|printf.*password/i,
    "password storage must not echo generated credentials",
  );

  const generateStep = extractStepBlock(
    workflow,
    "Generate strong Kuma password for rotation",
  );
  assert.match(generateStep, /openssl rand/);
  assert.doesNotMatch(
    generateStep,
    /echo.*rand|cat.*password_file/,
    "password generation must not print the credential",
  );

  assert.match(
    rotateScript,
    /redact_sensitive/,
    "rotate script must redact sensitive output on failure",
  );
  assert.doesNotMatch(
    rotateScript,
    /echo\s+["']?\$password/,
    "rotate script must not echo the password variable",
  );

  assert.ok(existsSync(rotateSelftestPath));
  assert.ok(existsSync(syncSelftestPath));
});

test("VOC-086-TEST-07 (T02 extension): workflow and host scripts ban SQLite deployment paths", () => {
  const paths = [
    ".github/workflows/sync-monitoring.yml",
    "infra/scripts/kuma-rotate-credentials.sh",
    "infra/scripts/sync-kuma-inventory.sh",
  ];
  const forbidden = [/kuma\.db/i, /\bsqlite\b/i, /\/app\/data/i];

  for (const relativePath of paths) {
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

test("VOC-086-TEST-02 (workflow): monitoring env secrets and Socket.IO sync wiring", () => {
  const workflow = readFileSync(workflowPath, "utf8");

  assert.match(workflow, /environment: monitoring/);
  assert.match(workflow, /KUMA_USERNAME/);
  assert.match(workflow, /KUMA_PASSWORD/);
  assert.match(workflow, /sync-kuma-inventory\.sh/);
  assert.match(workflow, /vocanova-monitoring-net/);
  assert.match(readFileSync(syncScriptPath, "utf8"), /sync-kuma\.mjs/);
});
