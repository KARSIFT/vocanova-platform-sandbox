// VOC-088-T00 — persistent, fail-closed staging signup allowlist.
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const workflowPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);
const validatorPath = path.join(
  repositoryRoot,
  "infra/scripts/validate-staging-signup-allowlist.sh",
);

function validate({ oauthEnabled, allowlist }) {
  return spawnSync("bash", [validatorPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: {
      PATH: process.env.PATH,
      STAGING_GOOGLE_OAUTH_ENABLED: oauthEnabled,
      STAGING_NEW_USER_SIGNUP_ALLOWLIST: allowlist,
    },
  });
}

test("VOC-088-TEST-00/01: every deploy uses the persistent staging secret", () => {
  const workflow = readFileSync(workflowPath, "utf8");

  assert.match(
    workflow,
    /STAGING_NEW_USER_SIGNUP_ALLOWLIST: \$\{\{ secrets\.STAGING_NEW_USER_SIGNUP_ALLOWLIST \}\}/,
  );
  assert.match(
    workflow,
    /echo "NEW_USER_SIGNUP_ALLOWLIST=\$\{STAGING_NEW_USER_SIGNUP_ALLOWLIST\}"/,
  );
  assert.doesNotMatch(workflow, /new_user_signup_allowlist:/);
  assert.doesNotMatch(
    workflow,
    /STAGING_NEW_USER_SIGNUP_ALLOWLIST: \$\{\{ inputs\./,
  );
});

test("VOC-088-TEST-02: workflow dispatch cannot replace or erase the cohort", () => {
  const workflow = readFileSync(workflowPath, "utf8");
  const dispatchBlock = workflow.slice(
    workflow.indexOf("workflow_dispatch:"),
    workflow.indexOf("\npermissions:"),
  );

  assert.doesNotMatch(dispatchBlock, /signup_allowlist/i);
});

test("VOC-088-TEST-03: OAuth enabled with an empty cohort fails closed", () => {
  const result = validate({ oauthEnabled: "true", allowlist: "" });

  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /controlled signup cohort is empty/);
  assert.equal(result.stdout, "");
});

test("VOC-088-TEST-03: malformed cohort fails without echoing its value", () => {
  const malformed = "not-an-email@synthetic.vocanova.invalid,broken";
  const result = validate({ oauthEnabled: "true", allowlist: malformed });

  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /controlled signup cohort is malformed/);
  assert.doesNotMatch(
    `${result.stdout}${result.stderr}`,
    /not-an-email|broken/,
  );
});

test("VOC-088-TEST-04: successful validation logs only a readiness boolean", () => {
  const cohort =
    "member@synthetic.vocanova.invalid, second@synthetic.vocanova.invalid";
  const result = validate({ oauthEnabled: "true", allowlist: cohort });

  assert.equal(result.status, 0);
  assert.equal(result.stdout.trim(), "controlled signup ready: true");
  assert.equal(result.stderr, "");
  assert.doesNotMatch(`${result.stdout}${result.stderr}`, /@|member|second|2/);
});

test("VOC-088-TEST-04: disabled OAuth accepts an empty cohort without data", () => {
  const result = validate({ oauthEnabled: "false", allowlist: "" });

  assert.equal(result.status, 0);
  assert.equal(result.stdout.trim(), "controlled signup ready: false");
  assert.equal(result.stderr, "");
});

test("VOC-088 isolation: production deploy never consumes the staging secret", () => {
  const production = readFileSync(
    path.join(repositoryRoot, ".github/workflows/deploy-production.yml"),
    "utf8",
  );

  assert.doesNotMatch(production, /STAGING_NEW_USER_SIGNUP_ALLOWLIST/);
});
