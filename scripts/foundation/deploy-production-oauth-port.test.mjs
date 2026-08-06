import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";

const DEPLOY_WORKFLOW_PATH = path.resolve(".github/workflows/deploy-production.yml");
const STEP_START_MARKER = "- name: Write production application configuration";
const STEP_END_MARKER = "- name: Deploy to production host";

const EXPECTED_CONFIG_LINES = [
  'echo "BASE_URL=https://${PRODUCTION_API_HOST}:8443"',
  'echo "OAUTH_REDIRECT_URI=https://${PRODUCTION_API_HOST}:8443/api/v1/auth/oauth/google/callback"',
  'echo "OAUTH_REDIRECT_ALLOWLIST=https://${PRODUCTION_WEB_HOST}:8443/onboarding,https://${PRODUCTION_WEB_HOST}:8443/home"',
];

function extractConfigStepBlock(workflowSource) {
  const stepStartIndex = workflowSource.indexOf(STEP_START_MARKER);
  assert.notEqual(
    stepStartIndex,
    -1,
    "deploy-production workflow is missing the production config step marker",
  );

  const stepEndIndex = workflowSource.indexOf(STEP_END_MARKER, stepStartIndex);
  assert.notEqual(
    stepEndIndex,
    -1,
    "deploy-production workflow is missing the deploy step marker after the config step",
  );

  return workflowSource.slice(stepStartIndex, stepEndIndex);
}

test("VOC-041-TEST-02: production OAuth/browser URLs stay :8443-qualified in config-writing step", () => {
  const workflowSource = readFileSync(DEPLOY_WORKFLOW_PATH, "utf8");
  const configStepBlock = extractConfigStepBlock(workflowSource);

  for (const expectedLine of EXPECTED_CONFIG_LINES) {
    assert.match(
      configStepBlock,
      new RegExp(expectedLine.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")),
      `Expected deploy-production config step to contain: ${expectedLine}`,
    );
  }
});
