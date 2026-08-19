// VOC-092-T01 — controlled-signup OAuth callback E2E harness structure,
// CI wiring, isolation, redaction, and bypass-denylist coverage.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const harnessTestPath = path.join(
  repositoryRoot,
  "apps/api/app/api/controlled_signup_oauth_e2e_test.go",
);
const postgresHarnessPath = path.join(
  repositoryRoot,
  "apps/api/app/api/controlled_signup_oauth_postgres_test.go",
);
const localRunnerPath = path.join(
  repositoryRoot,
  "infra/scripts/run-controlled-signup-oauth-e2e.sh",
);
const workflowPath = path.join(
  repositoryRoot,
  ".github/workflows/controlled-signup-oauth-e2e.yml",
);
const packageJsonPath = path.join(repositoryRoot, "package.json");

const HARNESS_GO_TEST_COMMAND =
  "go test ./app/api/... -run ControlledSignupOAuth -count=1";

const STAGING_PRODUCTION_HOST_DENYLIST = [
  "api-staging.vocanova.site",
  "api-production.vocanova.site",
  "staging.vocanova.site",
  "vocanova.site",
  "vocanova.com",
];

const SECRET_ENV_DENYLIST = [
  "DATABASE_URL",
  "STAGING_GOOGLE_OAUTH_CLIENT_SECRET",
  "GOOGLE_OAUTH_CLIENT_SECRET",
  "STAGING_NEW_USER_SIGNUP_ALLOWLIST",
];

const FORBIDDEN_BYPASS_PATTERNS = [
  /\/test-auth\b/i,
  /\?test=1\b/,
  /test_auth/i,
  /runtime.*fake.*provider.*switch/i,
];

const OAUTH_SECRET_LOG_PATTERNS = [
  /t\.Logf?\([^)]*authorization[_ ]?code/i,
  /t\.Logf?\([^)]*access[_ ]?token/i,
  /t\.Logf?\([^)]*refresh[_ ]?token/i,
  /t\.Logf?\([^)]*vocanova_session/i,
  /t\.Logf?\([^)]*vocanova_oauth_state/i,
  /fmt\.Print[f]?\([^)]*authorization[_ ]?code/i,
];

function readHarnessSources() {
  return [harnessTestPath, postgresHarnessPath]
    .filter((filePath) => existsSync(filePath))
    .map((filePath) => readFileSync(filePath, "utf8"))
    .join("\n");
}

function listWorkflowSources() {
  const workflowsDir = path.join(repositoryRoot, ".github/workflows");
  return readdirSync(workflowsDir)
    .filter((name) => name.endsWith(".yml") || name.endsWith(".yaml"))
    .map((name) => readFileSync(path.join(workflowsDir, name), "utf8"))
    .join("\n");
}

test("VOC-092-TEST-11: CI workflow executes harness on pull requests and develop", () => {
  assert.ok(
    existsSync(workflowPath),
    "controlled-signup-oauth-e2e workflow must exist",
  );

  const workflow = readFileSync(workflowPath, "utf8");

  assert.match(workflow, /pull_request:/);
  assert.match(workflow, /branches:\s*\n\s*- develop/);
  assert.match(workflow, /push:\s*\n\s*branches:\s*\n\s*- develop/);
  assert.match(workflow, /docker version/);
  assert.match(workflow, /docker info/);
  assert.match(workflow, /cache-dependency-path: apps\/api\/go\.sum/);
  assert.match(workflow, /TestControlledSignupOAuth_AllowlistedCallbackSucceeds/);
  assert.match(workflow, /TestControlledSignupOAuth_UnlistedCallbackDenied/);
  assert.match(workflow, /controlled-signup OAuth callback E2E cases must not skip/);
  assert.doesNotMatch(workflow, /continue-on-error:/);
  assert.doesNotMatch(workflow, /\|\|\s*true/);
  assert.match(
    workflow,
    new RegExp(HARNESS_GO_TEST_COMMAND.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")),
  );
});

test("VOC-092-TEST-00/01/12: harness artifacts and package scripts are committed", () => {
  assert.ok(
    existsSync(harnessTestPath),
    "controlled_signup_oauth_e2e_test.go must exist",
  );
  assert.ok(
    existsSync(postgresHarnessPath),
    "controlled_signup_oauth_postgres_test.go must exist",
  );
  assert.ok(
    existsSync(localRunnerPath),
    "run-controlled-signup-oauth-e2e.sh must exist",
  );

  const packageJson = JSON.parse(readFileSync(packageJsonPath, "utf8"));
  assert.match(
    packageJson.scripts["test:controlled-signup-oauth-e2e"],
    /ControlledSignupOAuth/,
    "package.json must expose a dedicated harness script",
  );
  assert.match(
    packageJson.scripts.test,
    /node --test scripts\/foundation\/\*\.test\.mjs/,
    "pnpm test must aggregate the VOC-092 foundation test",
  );

  const localRunner = readFileSync(localRunnerPath, "utf8");
  assert.match(localRunner, /ControlledSignupOAuth/);
  assert.match(
    localRunner,
    /refusing to run with staging\/production host argument/,
  );
});

test("VOC-092-TEST-07: harness fixtures use only synthetic.vocanova.invalid identities", () => {
  const harnessSources = readHarnessSources();

  const emailMatches = [
    ...harnessSources.matchAll(
      /[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}/g,
    ),
  ].map((match) => match[0]);

  assert.ok(
    emailMatches.length > 0,
    "harness must declare synthetic fixture emails",
  );
  const nonSyntheticEmailCount = emailMatches.filter(
    (email) => !email.toLowerCase().endsWith("@synthetic.vocanova.invalid"),
  ).length;
  assert.equal(
    nonSyntheticEmailCount,
    0,
    "harness email literals must stay on the synthetic domain",
  );

  assert.match(
    harnessSources,
    /allowlisted-callback-e2e@synthetic\.vocanova\.invalid/,
    "allowlisted fixture must use a dedicated synthetic identity",
  );
  assert.match(
    harnessSources,
    /unlisted-callback-e2e@synthetic\.vocanova\.invalid/,
    "unlisted fixture must use a dedicated synthetic identity",
  );
  assert.doesNotMatch(
    harnessSources,
    /OAuthIdentity\{[^}]*Email:\s*reservedSyntheticSmokeTestEmail/s,
    "reserved deploy smoke-test identity must not be used as a harness signup fixture",
  );
});

test("VOC-092-TEST-15: harness sources exclude staging/production hosts and secret env reads", () => {
  const harnessSources = readHarnessSources();

  for (const host of STAGING_PRODUCTION_HOST_DENYLIST) {
    assert.equal(
      new RegExp(host.replaceAll(".", "\\.")).test(harnessSources),
      false,
      `harness must not reference ${host}`,
    );
  }

  for (const envName of SECRET_ENV_DENYLIST) {
    assert.equal(
      new RegExp(`os\\.Getenv\\(["']${envName}["']\\)`).test(harnessSources),
      false,
      `harness must not read ${envName} from the environment`,
    );
  }

  assert.match(
    harnessSources,
    /docker not on PATH/,
    "local harness must explicitly report Docker absence instead of connecting elsewhere; CI forbids absence",
  );
});

test("VOC-092-TEST-09: no test-auth route or runtime authentication bypass introduced", () => {
  const harnessSources = readHarnessSources();
  const workflowSources = listWorkflowSources();
  const combined = `${harnessSources}\n${workflowSources}`;

  for (const pattern of FORBIDDEN_BYPASS_PATTERNS) {
    assert.equal(
      pattern.test(combined),
      false,
      `forbidden bypass pattern must not appear: ${pattern}`,
    );
  }

  assert.doesNotMatch(
    harnessSources,
    /RegisterAuth\([^)]*FakeOAuthProvider/,
    "harness must use GoogleOAuthProvider, not register a fake provider in production wiring",
  );
});

test("VOC-092-TEST-08: harness logging avoids printing OAuth secrets", () => {
  const harnessSources = readHarnessSources();

  for (const pattern of OAUTH_SECRET_LOG_PATTERNS) {
    assert.equal(
      pattern.test(harnessSources),
      false,
      `harness must not log OAuth secrets: ${pattern}`,
    );
  }

  const e2eSource = readFileSync(harnessTestPath, "utf8");
  const directTestLogs = [
    ...e2eSource.matchAll(/\bt\.Log\(("[^"\n]*")\)/g),
  ].map((match) => match[1]);
  assert.deepEqual(directTestLogs, [
    '"controlled-signup OAuth allowlisted callback succeeded with redirect to onboarding and persisted auth rows"',
    '"controlled-signup OAuth unlisted callback denied with HTTP 503 and no persisted user"',
  ]);
  assert.equal(
    /\bt\.Logf\(/.test(e2eSource),
    false,
    "OAuth E2E harness must not add formatted test logging",
  );

  assert.match(
    harnessSources,
    /session cookie must have a non-empty value/,
    "session presence is asserted without printing cookie values",
  );
});
