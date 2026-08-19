// VOC-084-T00 — deploy-staging Google OAuth sync + canonical config.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
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

test("VOC-084-TEST-04: signup stays false; allowlist is persistent-secret controlled", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const configStep = extractStepBlock(
    deployStaging,
    "Write staging application configuration",
  );

  assert.match(configStep, /echo "NEW_USER_SIGNUP_ENABLED=false"/);
  assert.match(
    configStep,
    /STAGING_NEW_USER_SIGNUP_ALLOWLIST: \$\{\{ secrets\.STAGING_NEW_USER_SIGNUP_ALLOWLIST \}\}/,
  );
  assert.match(
    configStep,
    /echo "NEW_USER_SIGNUP_ALLOWLIST=\$\{STAGING_NEW_USER_SIGNUP_ALLOWLIST\}"/,
  );
  assert.doesNotMatch(deployStaging, /new_user_signup_allowlist:/);
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

const verifyOAuthStartScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-staging-oauth-start.sh",
);
const verifyOAuthStartSelftestPath = path.join(
  repositoryRoot,
  "infra/scripts/verify-staging-oauth-start.selftest.sh",
);
const t02EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-084-restore-repository-managed-google-sign-in-in/t02-evidence.md",
);

test("VOC-084-TEST-06: deploy-staging wires live OAuth-start verification without following Google", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");

  assert.ok(
    existsSync(verifyOAuthStartScriptPath),
    "verify-staging-oauth-start.sh must exist",
  );
  assert.ok(
    existsSync(verifyOAuthStartSelftestPath),
    "verify-staging-oauth-start.selftest.sh must exist",
  );

  const oauthStartStep = extractStepBlock(
    deployStaging,
    "Verify staging OAuth start initiation",
  );

  assert.match(
    oauthStartStep,
    /EXPECT_OAUTH_ENABLED: \$\{\{ secrets\.GOOGLE_OAUTH_CLIENT_ID != '' && secrets\.GOOGLE_OAUTH_CLIENT_SECRET != '' \}\}/,
    "OAuth-start check must derive enabled expectation from repository pair availability",
  );
  assert.match(
    oauthStartStep,
    /verify-staging-oauth-start\.sh/,
    "deploy-staging must invoke the dedicated OAuth-start verification script",
  );

  const verifyScript = readFileSync(verifyOAuthStartScriptPath, "utf8");
  assert.match(
    verifyScript,
    /accounts\.google\.com/,
    "verification script must require accounts.google.com authorization URL",
  );
  assert.match(
    verifyScript,
    /parsed\.hostname != "accounts\.google\.com"/,
    "verification script must reject lookalike hostnames with an exact hostname comparison",
  );
  assert.match(
    verifyScript,
    new RegExp(
      CANONICAL_STAGING_CALLBACK.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"),
    ),
    "verification script must assert the canonical staging callback redirect_uri",
  );
  assert.doesNotMatch(
    verifyScript,
    /curl[^\n]*-L|--location/,
    "verification script must not follow Google redirects",
  );
  assert.match(
    verifyScript,
    /503/,
    "verification script must assert coherent disabled behavior",
  );

  const apiHealthIndex = deployStaging.indexOf(
    "- name: Poll api-staging.vocanova.site/healthz",
  );
  const webHealthIndex = deployStaging.indexOf(
    "- name: Poll staging.vocanova.site/",
  );
  const oauthStartIndex = deployStaging.indexOf(
    "- name: Verify staging OAuth start initiation",
  );
  assert.ok(apiHealthIndex < oauthStartIndex);
  assert.ok(webHealthIndex < oauthStartIndex);
});

test("VOC-084-TEST-07: public health polls remain before OAuth-start gate", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");

  assert.match(deployStaging, /Poll api-staging\.vocanova\.site\/healthz/);
  assert.match(deployStaging, /Poll staging\.vocanova\.site\//);
  assert.match(deployStaging, /Verify staging OAuth start initiation/);
});

test("VOC-084-TEST-08: Google client callback disposition is precisely recorded", () => {
  assert.ok(existsSync(t02EvidencePath), "t02-evidence.md must exist");

  const evidence = readFileSync(t02EvidencePath, "utf8");

  assert.match(
    evidence,
    /Google Cloud Console|Google OAuth client/i,
    "T02 evidence must reference Google Cloud Console disposition",
  );
  assert.match(
    evidence,
    new RegExp(
      CANONICAL_STAGING_CALLBACK.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"),
    ),
    "T02 evidence must cite the exact staging callback URI",
  );
  assert.match(
    evidence,
    /authorized redirect|Authorized redirect|redirect URI/i,
    "T02 evidence must describe the external redirect registration action",
  );
  assert.match(
    evidence,
    /not verified|not complete|external action|Console access unavailable|pending/i,
    "T02 evidence must not claim Console authorization complete without evidence",
  );
});
