// VOC-096-T00 — persistent, fail-closed production signup allowlist.
// Runs via `node --test scripts/foundation/voc096-deploy-production-allowlist.test.mjs`.

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
  ".github/workflows/deploy-production.yml",
);
const stagingWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);
const validatorPath = path.join(
  repositoryRoot,
  "infra/scripts/validate-production-signup-allowlist.sh",
);

function validate({ oauthEnabled, allowlist }) {
  return spawnSync("bash", [validatorPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: {
      PATH: process.env.PATH,
      PRODUCTION_GOOGLE_OAUTH_ENABLED: oauthEnabled,
      PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST: allowlist,
    },
  });
}

test("VOC-096-TEST-00/01: every deploy uses the persistent production secret", () => {
  const workflow = readFileSync(workflowPath, "utf8");

  assert.match(
    workflow,
    /PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST: \$\{\{ secrets\.PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST \}\}/,
  );
  assert.match(
    workflow,
    /echo "NEW_USER_SIGNUP_ALLOWLIST=\$\{PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST\}"/,
  );
  assert.match(workflow, /Validate production controlled-signup allowlist/);
  assert.match(
    workflow,
    /bash infra\/scripts\/validate-production-signup-allowlist\.sh/,
  );
  assert.doesNotMatch(workflow, /new_user_signup_allowlist:/);
  assert.doesNotMatch(
    workflow,
    /PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST: \$\{\{ inputs\./,
  );
});

test("VOC-096-TEST-02: workflow dispatch cannot replace or erase the cohort", () => {
  const workflow = readFileSync(workflowPath, "utf8");
  const dispatchBlock = workflow.slice(
    workflow.indexOf("workflow_dispatch:"),
    workflow.indexOf("\npermissions:"),
  );

  assert.doesNotMatch(dispatchBlock, /signup_allowlist/i);
  assert.doesNotMatch(workflow, /inputs\.new_user_signup_allowlist/);
});

test("VOC-096-TEST-03: OAuth enabled with an empty cohort fails closed", () => {
  const result = validate({ oauthEnabled: "true", allowlist: "" });

  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /controlled signup cohort is empty/);
  assert.equal(result.stdout, "");
});

test("VOC-096-TEST-03/04: malformed cohort fails without echoing its value", () => {
  const malformed = "not-an-email@synthetic.vocanova.invalid,broken";
  const result = validate({ oauthEnabled: "true", allowlist: malformed });

  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /controlled signup cohort is malformed/);
  assert.doesNotMatch(
    `${result.stdout}${result.stderr}`,
    /not-an-email|broken/,
  );
});

test("VOC-096-TEST-04: multiline cohort fails without echoing its value", () => {
  const multiline =
    "member@synthetic.vocanova.invalid\nsecond@synthetic.vocanova.invalid";
  const result = validate({ oauthEnabled: "true", allowlist: multiline });

  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /controlled signup cohort is malformed/);
  assert.doesNotMatch(`${result.stdout}${result.stderr}`, /member|second/);
});

test("VOC-096-TEST-05: successful validation logs only a readiness boolean", () => {
  const cohort =
    "member@synthetic.vocanova.invalid, second@synthetic.vocanova.invalid";
  const result = validate({ oauthEnabled: "true", allowlist: cohort });

  assert.equal(result.status, 0);
  assert.equal(result.stdout.trim(), "controlled signup ready: true");
  assert.equal(result.stderr, "");
  assert.doesNotMatch(`${result.stdout}${result.stderr}`, /@|member|second|2/);
});

test("VOC-096-TEST-05: disabled OAuth accepts an empty cohort without data", () => {
  const result = validate({ oauthEnabled: "false", allowlist: "" });

  assert.equal(result.status, 0);
  assert.equal(result.stdout.trim(), "controlled signup ready: false");
  assert.equal(result.stderr, "");
});

test("VOC-096-TEST-10: production deploy never consumes the staging secret", () => {
  const workflow = readFileSync(workflowPath, "utf8");

  assert.doesNotMatch(workflow, /STAGING_NEW_USER_SIGNUP_ALLOWLIST/);
});

test("VOC-096-TEST-11: staging deploy never consumes the production secret", () => {
  const workflow = readFileSync(stagingWorkflowPath, "utf8");

  assert.doesNotMatch(workflow, /PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST/);
});

test("VOC-096-TEST-00: SSH config write refuses empty cohort when OAuth enabled", () => {
  const workflow = readFileSync(workflowPath, "utf8");

  assert.match(
    workflow,
    /controlled signup cohort is empty; refusing production configuration with OAuth enabled/,
  );
  assert.match(workflow, /controlled signup ready: true/);
  assert.match(workflow, /controlled signup ready: false/);
});
