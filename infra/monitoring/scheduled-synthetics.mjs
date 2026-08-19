import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import {
  CANONICAL_SYNTHETIC_IDS,
  parseMonitoringYaml,
} from "./validate-inventory.mjs";

export const SCHEDULED_SYNTHETICS_WORKFLOW =
  ".github/workflows/scheduled-synthetics.yml";

export const CANONICAL_CHECK_REFS = [
  "staging-oauth-expected-state",
  "production-oauth-expected-state",
  "production-journey-content",
  "staging-authenticated-core-journey",
  "production-authenticated-route-content-sweep",
];

// OAuth is an expected staging and production capability. Scheduled checks
// must alert when it becomes disabled; they must not redefine a missing or
// partial credential pair as a healthy disabled state.
export const EXPECT_OAUTH_ENABLED_TRUE = 'EXPECT_OAUTH_ENABLED: "true"';

export const STAGING_CORE_JOURNEY_CHECK_REF =
  "staging-authenticated-core-journey";
export const STAGING_CORE_JOURNEY_SYNTHETIC_ID =
  "synthetic.staging.authenticated-core-journey";
export const STAGING_CORE_JOURNEY_JOB_TIMEOUT_MINUTES = 40;
export const STAGING_CORE_JOURNEY_REGISTRY_TIMEOUT_SECONDS = 2400;
export const PLAYWRIGHT_STAGING_JOURNEY_TIMEOUT_SECONDS = 240;
// SSH command_timeout (2m) + session mint + conservative warm-cache install
// ceiling documented in VOC-090-T00 evidence.
export const STAGING_CORE_JOURNEY_SETUP_RESERVE_SECONDS = 16 * 60;

const OAUTH_AVAILABILITY_JOBS = [
  "staging-oauth-expected-state",
  "production-oauth-expected-state",
];

const PRODUCTION_SMOKE_PROFILE_JOBS = [
  "production-journey-content",
  "production-authenticated-route-content-sweep",
];

export function extractTopLevelJobBlock(workflowSource, jobId) {
  const escapedJobId = jobId.replaceAll(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const jobStart = new RegExp(`^  ${escapedJobId}:\\s*$`, "m");
  const match = jobStart.exec(workflowSource);
  if (!match) {
    return "";
  }
  const start = match.index;
  const afterHeader = start + match[0].length;
  const rest = workflowSource.slice(afterHeader);
  const nextJob = /^  [A-Za-z0-9_-]+:\s*$/m.exec(rest);
  const end = nextJob ? afterHeader + nextJob.index : workflowSource.length;
  return workflowSource.slice(start, end);
}

export function parseJobTimeoutMinutes(jobBlock) {
  const match = /timeout-minutes:\s*(\d+)/.exec(jobBlock ?? "");
  return match ? Number(match[1]) : null;
}

export function validateStagingCoreJourneyBudget({
  workflowSource,
  syntheticsDocument,
} = {}) {
  const errors = [];
  const jobBlock = extractTopLevelJobBlock(
    workflowSource,
    STAGING_CORE_JOURNEY_CHECK_REF,
  );
  if (!jobBlock) {
    errors.push(
      `scheduled-synthetics workflow missing job block for ${STAGING_CORE_JOURNEY_CHECK_REF}`,
    );
    return errors;
  }

  const entry = (syntheticsDocument?.synthetics ?? []).find(
    (item) => item.id === STAGING_CORE_JOURNEY_SYNTHETIC_ID,
  );
  if (!entry) {
    errors.push(
      `scheduled synthetics registry missing id ${STAGING_CORE_JOURNEY_SYNTHETIC_ID}`,
    );
    return errors;
  }

  const jobTimeoutMinutes = parseJobTimeoutMinutes(jobBlock);
  if (jobTimeoutMinutes === null) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must declare timeout-minutes`,
    );
  } else if (jobTimeoutMinutes < entry.timeout_seconds / 60) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} timeout-minutes (${jobTimeoutMinutes}) must be >= registry timeout_seconds / 60 (${entry.timeout_seconds / 60})`,
    );
  } else if (entry.timeout_seconds !== jobTimeoutMinutes * 60) {
    errors.push(
      `${STAGING_CORE_JOURNEY_SYNTHETIC_ID} timeout_seconds (${entry.timeout_seconds}) must equal job timeout-minutes * 60 (${jobTimeoutMinutes * 60})`,
    );
  }

  const jobTimeoutSeconds = (jobTimeoutMinutes ?? 0) * 60;
  const minimumBudget =
    PLAYWRIGHT_STAGING_JOURNEY_TIMEOUT_SECONDS +
    STAGING_CORE_JOURNEY_SETUP_RESERVE_SECONDS;
  if (jobTimeoutSeconds < minimumBudget) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} job wall clock (${jobTimeoutSeconds}s) must cover Playwright journey timeout (${PLAYWRIGHT_STAGING_JOURNEY_TIMEOUT_SECONDS}s) plus setup reserve (${STAGING_CORE_JOURNEY_SETUP_RESERVE_SECONDS}s)`,
    );
  }

  if (!/cache:\s*["']pnpm["']/m.test(jobBlock)) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must enable pnpm dependency caching via setup-node cache: pnpm`,
    );
  }

  const pnpmActivateIndex = jobBlock.indexOf("corepack prepare pnpm@");
  const pnpmCacheIndex = jobBlock.search(/cache:\s*["']pnpm["']/);
  const pnpmInstallIndex = jobBlock.indexOf("pnpm install --frozen-lockfile");
  if (
    pnpmCacheIndex >= 0 &&
    pnpmActivateIndex >= 0 &&
    pnpmCacheIndex < pnpmActivateIndex
  ) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must activate pnpm via corepack before setup-node cache: pnpm`,
    );
  }
  if (
    pnpmInstallIndex >= 0 &&
    pnpmCacheIndex >= 0 &&
    pnpmCacheIndex > pnpmInstallIndex
  ) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must configure setup-node cache: pnpm before pnpm install`,
    );
  }

  const playwrightCacheAction =
    "actions/cache@d4323d4df104b026a6aa633fdb11d772146be0bf";
  if (!jobBlock.includes(playwrightCacheAction)) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must restore Playwright browser cache via the pinned actions/cache revision`,
    );
  }

  if (!jobBlock.includes("~/.cache/ms-playwright")) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must cache Playwright browsers under ~/.cache/ms-playwright`,
    );
  }

  const installCommand =
    "run: bash infra/scripts/install-playwright-chromium.sh";
  const installIndex = jobBlock.indexOf(installCommand);
  if (installIndex < 0) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must invoke infra/scripts/install-playwright-chromium.sh`,
    );
  }
  const cacheIndex = jobBlock.indexOf(playwrightCacheAction);
  if (installIndex >= 0 && cacheIndex >= 0 && cacheIndex > installIndex) {
    errors.push(
      `${STAGING_CORE_JOURNEY_CHECK_REF} must restore Playwright browser cache before install-playwright-chromium.sh`,
    );
  }

  return errors;
}

export function loadSyntheticsRegistry(repositoryRoot) {
  const syntheticsPath = path.join(
    repositoryRoot,
    "infra/monitoring/synthetics.yaml",
  );
  if (!existsSync(syntheticsPath)) {
    throw new Error("missing infra/monitoring/synthetics.yaml");
  }
  return parseMonitoringYaml(readFileSync(syntheticsPath, "utf8"));
}

export function indexSyntheticsByCheckRef(syntheticsDocument) {
  const byCheckRef = new Map();
  for (const synthetic of syntheticsDocument.synthetics ?? []) {
    byCheckRef.set(synthetic.check_ref, synthetic);
  }
  return byCheckRef;
}

export function validateScheduledSyntheticsRegistry(syntheticsDocument) {
  const errors = [];
  const synthetics = syntheticsDocument?.synthetics;
  if (!Array.isArray(synthetics)) {
    return ["synthetics.synthetics must be an array"];
  }

  const byCheckRef = indexSyntheticsByCheckRef(syntheticsDocument);
  for (const checkRef of CANONICAL_CHECK_REFS) {
    const entry = byCheckRef.get(checkRef);
    if (!entry) {
      errors.push(
        `scheduled synthetics registry missing check_ref ${checkRef}`,
      );
      continue;
    }
    if (entry.workflow_ref !== SCHEDULED_SYNTHETICS_WORKFLOW) {
      errors.push(
        `${entry.id}: workflow_ref must be ${SCHEDULED_SYNTHETICS_WORKFLOW}`,
      );
    }
  }

  for (const syntheticId of CANONICAL_SYNTHETIC_IDS) {
    const entry = synthetics.find((item) => item.id === syntheticId);
    if (!entry) {
      errors.push(`scheduled synthetics registry missing id ${syntheticId}`);
    }
  }

  const productionSweep = synthetics.find(
    (item) =>
      item.id === "synthetic.production.authenticated-route-content-sweep",
  );
  if (productionSweep && productionSweep.mutating !== false) {
    errors.push(
      "synthetic.production.authenticated-route-content-sweep must declare mutating: false",
    );
  }

  return errors;
}

export function validateScheduledSyntheticsWorkflow({
  workflowSource,
  syntheticsDocument,
  repositoryRoot,
} = {}) {
  const errors = validateScheduledSyntheticsRegistry(syntheticsDocument);
  if (!workflowSource || typeof workflowSource !== "string") {
    errors.push("workflow source must be a string");
    return errors;
  }

  if (!workflowSource.includes("name: scheduled-synthetics")) {
    errors.push(
      "scheduled-synthetics workflow must declare name: scheduled-synthetics",
    );
  }

  if (!/schedule:\s*\n\s*- cron:/m.test(workflowSource)) {
    errors.push("scheduled-synthetics workflow must define a schedule cron");
  }

  if (!workflowSource.includes("workflow_dispatch:")) {
    errors.push("scheduled-synthetics workflow must support workflow_dispatch");
  }

  for (const checkRef of CANONICAL_CHECK_REFS) {
    if (!workflowSource.includes(checkRef)) {
      errors.push(
        `scheduled-synthetics workflow missing job wiring for check_ref ${checkRef}`,
      );
    }
  }

  for (const syntheticId of CANONICAL_SYNTHETIC_IDS) {
    if (!workflowSource.includes(syntheticId)) {
      errors.push(
        `scheduled-synthetics workflow missing stable synthetic id ${syntheticId}`,
      );
    }
  }

  if (!workflowSource.includes("STAGING_SMOKE_TEST_SESSION_MINT_TOKEN")) {
    errors.push(
      "scheduled-synthetics workflow must reuse STAGING_SMOKE_TEST_SESSION_MINT_TOKEN",
    );
  }

  if (!workflowSource.includes("PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN")) {
    errors.push(
      "scheduled-synthetics workflow must reuse PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN",
    );
  }

  if (!workflowSource.includes("run-scheduled-synthetic.sh")) {
    errors.push(
      "scheduled-synthetics workflow must invoke run-scheduled-synthetic.sh",
    );
  }

  if (!workflowSource.includes("mint-synthetic-session.sh")) {
    errors.push(
      "scheduled-synthetics workflow must invoke mint-synthetic-session.sh",
    );
  }

  const mintScriptPath = repositoryRoot
    ? path.join(repositoryRoot, "infra/scripts/mint-synthetic-session.sh")
    : null;
  const mintScriptSource =
    mintScriptPath && existsSync(mintScriptPath)
      ? readFileSync(mintScriptPath, "utf8")
      : "";
  if (!mintScriptSource.includes("::add-mask::")) {
    errors.push(
      "mint-synthetic-session.sh must mask minted session values with ::add-mask::",
    );
  }

  if (
    /uses:.*error-monitoring|workflow_call:.*error-monitoring/m.test(
      workflowSource,
    )
  ) {
    errors.push(
      "scheduled-synthetics workflow must not replace or call error-monitoring.yml",
    );
  }

  for (const jobId of [
    ...OAUTH_AVAILABILITY_JOBS,
    ...PRODUCTION_SMOKE_PROFILE_JOBS,
  ]) {
    const jobBlock = extractTopLevelJobBlock(workflowSource, jobId);
    if (!jobBlock) {
      errors.push(
        `scheduled-synthetics workflow missing job block for ${jobId}`,
      );
      continue;
    }
    if (!jobBlock.includes(EXPECT_OAUTH_ENABLED_TRUE)) {
      errors.push(
        `${jobId} must require EXPECT_OAUTH_ENABLED: "true" so loss of expected OAuth availability alerts`,
      );
    }
  }

  for (const jobId of PRODUCTION_SMOKE_PROFILE_JOBS) {
    const jobBlock = extractTopLevelJobBlock(workflowSource, jobId);
    if (!jobBlock) {
      continue;
    }
    if (!jobBlock.includes('EXPECT_MAGIC_LINK_ENABLED: "false"')) {
      errors.push(`${jobId} must set EXPECT_MAGIC_LINK_ENABLED: "false"`);
    }
    if (!jobBlock.includes('EXPECT_NEW_SIGNUPS_ENABLED: "false"')) {
      errors.push(`${jobId} must set EXPECT_NEW_SIGNUPS_ENABLED: "false"`);
    }
    if (!jobBlock.includes('EXPECT_AI_ENABLED: "true"')) {
      errors.push(`${jobId} must set EXPECT_AI_ENABLED: "true"`);
    }
  }

  errors.push(
    ...validateStagingCoreJourneyBudget({
      workflowSource,
      syntheticsDocument,
    }),
  );

  return errors;
}

export function validateScheduledSyntheticsFiles(repositoryRoot) {
  const syntheticsDocument = loadSyntheticsRegistry(repositoryRoot);
  const workflowPath = path.join(repositoryRoot, SCHEDULED_SYNTHETICS_WORKFLOW);
  if (!existsSync(workflowPath)) {
    return [`missing ${SCHEDULED_SYNTHETICS_WORKFLOW}`];
  }
  const workflowSource = readFileSync(workflowPath, "utf8");
  return validateScheduledSyntheticsWorkflow({
    workflowSource,
    syntheticsDocument,
    repositoryRoot,
  });
}
