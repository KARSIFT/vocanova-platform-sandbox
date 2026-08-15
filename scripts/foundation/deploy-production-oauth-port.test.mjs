import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";

const DEPLOY_WORKFLOW_PATH = path.resolve(
  ".github/workflows/deploy-production.yml",
);
const STEP_START_MARKER = "- name: Write production application configuration";
const STEP_END_MARKER = "- name: Deploy to production host";

const EXPECTED_CONFIG_LINES = [
  'echo "BASE_URL=https://${PRODUCTION_API_HOST}"',
  'echo "OAUTH_REDIRECT_URI=https://${PRODUCTION_API_HOST}/api/v1/auth/oauth/google/callback"',
  'echo "OAUTH_REDIRECT_ALLOWLIST=https://${PRODUCTION_WEB_HOST}/onboarding,https://${PRODUCTION_WEB_HOST}/home"',
];

const FORBIDDEN_PORT_QUALIFICATION = /:8443/;

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

test("VOC-079-TEST-03: production OAuth/browser URLs use canonical HTTPS without :8443 (supersedes VOC-041-TEST-02 bridge-era assertion)", () => {
  const workflowSource = readFileSync(DEPLOY_WORKFLOW_PATH, "utf8");
  const configStepBlock = extractConfigStepBlock(workflowSource);

  for (const expectedLine of EXPECTED_CONFIG_LINES) {
    assert.match(
      configStepBlock,
      new RegExp(expectedLine.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")),
      `Expected deploy-production config step to contain: ${expectedLine}`,
    );
  }

  const emittedConfigLines = configStepBlock
    .split("\n")
    .filter(
      (line) =>
        /^\s*echo "BASE_URL=/.test(line) ||
        /^\s*echo "OAUTH_REDIRECT_URI=/.test(line) ||
        /^\s*echo "OAUTH_REDIRECT_ALLOWLIST=/.test(line),
    )
    .join("\n");
  assert.doesNotMatch(
    emittedConfigLines,
    FORBIDDEN_PORT_QUALIFICATION,
    "deploy-production config-writing step must not emit production :8443 URLs after VOC-079-T01",
  );
});
