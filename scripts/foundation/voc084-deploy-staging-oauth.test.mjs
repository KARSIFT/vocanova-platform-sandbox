// VOC-084-T00 — deploy-staging Google OAuth sync + canonical config.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const deployStagingPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);
const deployProductionPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-production.yml",
);

const CANONICAL_STAGING_CALLBACK =
  "https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback";
const CANONICAL_STAGING_ALLOWLIST =
  "https://staging.vocanova.site/onboarding,https://staging.vocanova.site/home";

const PARTIAL_PAIR_REJECTION =
  /GOOGLE_OAUTH_CLIENT_ID \/ GOOGLE_OAUTH_CLIENT_SECRET partially set - refusing/;

function extractStepBlock(workflowSource, stepName) {
  const marker = `- name: ${stepName}`;
  const stepStartIndex = workflowSource.indexOf(marker);
  assert.notEqual(
    stepStartIndex,
    -1,
    `deploy-staging workflow is missing step: ${stepName}`,
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

function assertCredentialsWrittenOnlyToApiEnv(oauthSyncBlock) {
  assert.match(
    oauthSyncBlock,
    /\{\s*\n\s*echo "GOOGLE_OAUTH_CLIENT_ID=\$\{STAGING_GOOGLE_OAUTH_CLIENT_ID\}"\s*\n\s*echo "GOOGLE_OAUTH_CLIENT_SECRET=\$\{STAGING_GOOGLE_OAUTH_CLIENT_SECRET\}"\s*\n\s*\} >> \/opt\/vocanova\/infra\/secrets\/api\.env/,
    "OAuth credentials must be appended only to the staging api.env secret file",
  );
}

test("VOC-084-TEST-00: partial Google credential pair fails before convergence", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");

  const runnerValidation = extractStepBlock(
    deployStaging,
    "Validate staging Google OAuth credential pair",
  );
  const oauthSync = extractStepBlock(
    deployStaging,
    "Sync staging Google OAuth configuration from GitHub secrets",
  );

  for (const block of [runnerValidation, oauthSync]) {
    assert.match(block, PARTIAL_PAIR_REJECTION);
    assert.match(block, /exit 1/);
  }

  assert.doesNotMatch(
    runnerValidation,
    /\/opt\/vocanova\/infra\/secrets\/api\.env/,
    "runner validation must not write OAuth credentials before SSH convergence",
  );
  assertCredentialsWrittenOnlyToApiEnv(oauthSync);

  const deployIndex = deployStaging.indexOf("- name: Deploy to staging host");
  const oauthSyncIndex = deployStaging.indexOf(
    "- name: Sync staging Google OAuth configuration from GitHub secrets",
  );
  const configIndex = deployStaging.indexOf(
    "- name: Write staging application configuration",
  );
  assert.ok(oauthSyncIndex < configIndex);
  assert.ok(configIndex < deployIndex);
});

test("VOC-084-TEST-01: both credentials present sync safely to staging api.env", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const oauthSync = extractStepBlock(
    deployStaging,
    "Sync staging Google OAuth configuration from GitHub secrets",
  );
  const configStep = extractStepBlock(
    deployStaging,
    "Write staging application configuration",
  );

  assert.match(
    oauthSync,
    /GOOGLE_OAUTH_CLIENT_ID=\$\{STAGING_GOOGLE_OAUTH_CLIENT_ID\}/,
  );
  assert.match(
    oauthSync,
    /GOOGLE_OAUTH_CLIENT_SECRET=\$\{STAGING_GOOGLE_OAUTH_CLIENT_SECRET\}/,
  );
  assert.match(oauthSync, /\/opt\/vocanova\/infra\/secrets\/api\.env/);
  assert.match(
    oauthSync,
    /chmod 600 \/opt\/vocanova\/infra\/secrets\/api\.env/,
  );
  assertCredentialsWrittenOnlyToApiEnv(oauthSync);

  assert.match(
    configStep,
    /STAGING_GOOGLE_OAUTH_ENABLED: \$\{\{ secrets\.GOOGLE_OAUTH_CLIENT_ID != '' && secrets\.GOOGLE_OAUTH_CLIENT_SECRET != '' \}\}/,
  );
  assert.match(
    configStep,
    /echo "GOOGLE_OAUTH_ENABLED=\$\{STAGING_GOOGLE_OAUTH_ENABLED\}"/,
  );
});

test("VOC-084-TEST-02: both credentials absent converges to coherent disabled OAuth", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const oauthSync = extractStepBlock(
    deployStaging,
    "Sync staging Google OAuth configuration from GitHub secrets",
  );
  const configStep = extractStepBlock(
    deployStaging,
    "Write staging application configuration",
  );

  assert.match(oauthSync, /both unset - skipping \(OAuth not yet adopted\)/);
  assert.match(
    configStep,
    /STAGING_GOOGLE_OAUTH_ENABLED: \$\{\{ secrets\.GOOGLE_OAUTH_CLIENT_ID != '' && secrets\.GOOGLE_OAUTH_CLIENT_SECRET != '' \}\}/,
  );
  assert.match(
    configStep,
    /echo "GOOGLE_OAUTH_ENABLED=\$\{STAGING_GOOGLE_OAUTH_ENABLED\}"/,
  );
});

test("VOC-084-TEST-03: canonical staging OAuth callback URI is exact", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const configStep = extractStepBlock(
    deployStaging,
    "Write staging application configuration",
  );

  assert.match(
    configStep,
    new RegExp(
      `echo "OAUTH_REDIRECT_URI=${CANONICAL_STAGING_CALLBACK.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}"`,
    ),
  );
  assert.match(
    configStep,
    new RegExp(
      `echo "OAUTH_REDIRECT_ALLOWLIST=${CANONICAL_STAGING_ALLOWLIST.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}"`,
    ),
  );
  assert.doesNotMatch(configStep, /api-production\.vocanova\.site/);
  assert.doesNotMatch(configStep, /:8443/);
});

test("VOC-084-TEST-04: signup stays false; allowlist defaults empty and is workflow-controlled", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const configStep = extractStepBlock(
    deployStaging,
    "Write staging application configuration",
  );

  assert.match(configStep, /echo "NEW_USER_SIGNUP_ENABLED=false"/);
  assert.match(
    configStep,
    /STAGING_NEW_USER_SIGNUP_ALLOWLIST: \$\{\{ inputs\.new_user_signup_allowlist \|\| '' \}\}/,
  );
  assert.match(
    configStep,
    /echo "NEW_USER_SIGNUP_ALLOWLIST=\$\{STAGING_NEW_USER_SIGNUP_ALLOWLIST\}"/,
  );
  assert.match(deployStaging, /new_user_signup_allowlist:[\s\S]*default: ""/);
});

test("VOC-084-TEST-07 (partial): production OAuth sync semantics are unchanged", () => {
  const deployProduction = readFileSync(deployProductionPath, "utf8");
  const productionOAuthSync = extractStepBlock(
    deployProduction,
    "Sync production Google OAuth configuration from GitHub secrets",
  );

  assert.match(
    productionOAuthSync,
    /\/opt\/vocanova\/production\/secrets\/api\.env/,
  );
  assert.doesNotMatch(
    readFileSync(deployStagingPath, "utf8"),
    /\/opt\/vocanova\/production\/secrets\/api\.env/,
    "staging deploy must not write production secret paths",
  );
});
