// VOC-090-T00 — scheduled staging core-journey CI budget and caching tests.
//
// Runs via `node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import {
  PLAYWRIGHT_STAGING_JOURNEY_TIMEOUT_SECONDS,
  SCHEDULED_SYNTHETICS_WORKFLOW,
  STAGING_CORE_JOURNEY_CHECK_REF,
  STAGING_CORE_JOURNEY_JOB_TIMEOUT_MINUTES,
  STAGING_CORE_JOURNEY_REGISTRY_TIMEOUT_SECONDS,
  STAGING_CORE_JOURNEY_SETUP_RESERVE_SECONDS,
  STAGING_CORE_JOURNEY_SYNTHETIC_ID,
  extractTopLevelJobBlock,
  loadSyntheticsRegistry,
  parseJobTimeoutMinutes,
  validateStagingCoreJourneyBudget,
} from "../../infra/monitoring/scheduled-synthetics.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const workflowPath = path.join(repositoryRoot, SCHEDULED_SYNTHETICS_WORKFLOW);
const evidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t00-evidence.md",
);
const playwrightConfigPath = path.join(
  repositoryRoot,
  "apps/web/playwright.staging.config.ts",
);
const coreLoopSpecPath = path.join(
  repositoryRoot,
  "apps/web/tests/staging-e2e/core-loop.staging.spec.ts",
);

function loadWorkflowSource() {
  return readFileSync(workflowPath, "utf8");
}

function loadStagingCoreJourneyJobBlock() {
  return extractTopLevelJobBlock(
    loadWorkflowSource(),
    STAGING_CORE_JOURNEY_CHECK_REF,
  );
}

test("VOC-090-TEST-00: root-cause evidence references run 32271016931 and dominant phase", () => {
  const evidence = readFileSync(evidencePath, "utf8");

  assert.match(evidence, /32271016931/);
  assert.match(evidence, /synthetic\.staging\.authenticated-core-journey/);
  assert.match(evidence, /30m0s|30m 0s|30 minutes/i);
  assert.match(
    evidence,
    /playwright install|Install Playwright Chromium/i,
    "evidence must identify Playwright browser install as the dominant budget consumer",
  );
  assert.doesNotMatch(evidence, /Bearer\s+[A-Za-z0-9._-]+/);
  assert.doesNotMatch(evidence, /vocanova_session=/);
});

test("VOC-090-TEST-01: staging core-journey job declares pnpm caching", () => {
  const jobBlock = loadStagingCoreJourneyJobBlock();
  assert.ok(jobBlock, "workflow must define staging core-journey job");

  const installIndex = jobBlock.indexOf("pnpm install --frozen-lockfile");
  assert.ok(installIndex >= 0, "job must run pnpm install --frozen-lockfile");
  assert.match(
    jobBlock,
    /cache:\s*["']pnpm["']/,
    "setup-node must enable pnpm caching before install",
  );
  assert.ok(
    jobBlock.indexOf('cache: "pnpm"') < installIndex ||
      jobBlock.indexOf("cache: 'pnpm'") < installIndex,
    "pnpm cache must be configured before pnpm install",
  );
});

test("VOC-090-TEST-02: staging core-journey job declares Playwright browser caching", () => {
  const jobBlock = loadStagingCoreJourneyJobBlock();
  const installIndex = jobBlock.indexOf(
    "run: pnpm --filter @vocanova/web exec playwright install --with-deps chromium",
  );

  assert.ok(installIndex >= 0, "job must install Playwright Chromium");
  assert.match(jobBlock, /actions\/cache@/);
  assert.match(jobBlock, /~\/\.cache\/ms-playwright/);
  assert.ok(
    jobBlock.indexOf("actions/cache@") < installIndex,
    "Playwright browser cache restore must precede playwright install",
  );
});

test("VOC-090-TEST-03: registry timeout_seconds matches job timeout-minutes", () => {
  const syntheticsDocument = loadSyntheticsRegistry(repositoryRoot);
  const entry = syntheticsDocument.synthetics.find(
    (item) => item.id === STAGING_CORE_JOURNEY_SYNTHETIC_ID,
  );
  const jobBlock = loadStagingCoreJourneyJobBlock();
  const jobTimeoutMinutes = parseJobTimeoutMinutes(jobBlock);

  assert.ok(entry);
  assert.equal(jobTimeoutMinutes, STAGING_CORE_JOURNEY_JOB_TIMEOUT_MINUTES);
  assert.equal(
    entry.timeout_seconds,
    STAGING_CORE_JOURNEY_REGISTRY_TIMEOUT_SECONDS,
  );
  assert.equal(entry.timeout_seconds, jobTimeoutMinutes * 60);
});

test("VOC-090-TEST-04: job timeout covers Playwright journey timeout plus setup reserve", () => {
  const jobBlock = loadStagingCoreJourneyJobBlock();
  const jobTimeoutMinutes = parseJobTimeoutMinutes(jobBlock);
  const playwrightConfig = readFileSync(playwrightConfigPath, "utf8");

  assert.match(playwrightConfig, /JOURNEY_TIMEOUT_MS = 240_000/);
  assert.equal(PLAYWRIGHT_STAGING_JOURNEY_TIMEOUT_SECONDS, 240);

  const jobTimeoutSeconds = jobTimeoutMinutes * 60;
  const minimumBudget =
    PLAYWRIGHT_STAGING_JOURNEY_TIMEOUT_SECONDS +
    STAGING_CORE_JOURNEY_SETUP_RESERVE_SECONDS;

  assert.ok(
    jobTimeoutSeconds >= minimumBudget,
    `job timeout ${jobTimeoutSeconds}s must be >= journey (${PLAYWRIGHT_STAGING_JOURNEY_TIMEOUT_SECONDS}s) + setup reserve (${STAGING_CORE_JOURNEY_SETUP_RESERVE_SECONDS}s)`,
  );
});

test("VOC-090-TEST-05: core-loop synthetic wiring unchanged", () => {
  const workflowSource = loadWorkflowSource();
  const jobBlock = loadStagingCoreJourneyJobBlock();
  const coreLoopSpec = readFileSync(coreLoopSpecPath, "utf8");

  assert.match(jobBlock, /Refresh reserved staging synthetic review state/);
  assert.match(jobBlock, /seed-synthetic-smoke-user\.sh/);
  assert.match(jobBlock, /mint-synthetic-session\.sh/);
  assert.match(jobBlock, /playwright\.staging\.config\.ts/);
  assert.ok(
    jobBlock.indexOf("seed-synthetic-smoke-user.sh") <
      jobBlock.indexOf("Mint synthetic smoke-test session"),
    "SSH seed must precede session mint",
  );
  assert.match(coreLoopSpec, /const MAX_REVIEW_CARDS = 8;/);
  assert.doesNotMatch(jobBlock, /retries:\s*[1-9]/);

  const errors = validateStagingCoreJourneyBudget({
    workflowSource,
    syntheticsDocument: loadSyntheticsRegistry(repositoryRoot),
  });
  assert.deepEqual(errors, [], errors.join("; "));
});
