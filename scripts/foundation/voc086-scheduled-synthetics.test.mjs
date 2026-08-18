// VOC-086-T03 — scheduled synthetics workflow/registry wiring tests.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { redactSecrets } from "../../infra/monitoring/kuma-sync/redact.mjs";
import {
  CANONICAL_CHECK_REFS,
  EXPECT_OAUTH_ENABLED_FROM_SECRETS,
  SCHEDULED_SYNTHETICS_WORKFLOW,
  extractTopLevelJobBlock,
  loadSyntheticsRegistry,
  validateScheduledSyntheticsFiles,
  validateScheduledSyntheticsWorkflow,
} from "../../infra/monitoring/scheduled-synthetics.mjs";
import { CANONICAL_SYNTHETIC_IDS } from "../../infra/monitoring/validate-inventory.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const workflowPath = path.join(repositoryRoot, SCHEDULED_SYNTHETICS_WORKFLOW);
const runSyntheticScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/run-scheduled-synthetic.sh",
);
const mintSessionScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/mint-synthetic-session.sh",
);
const errorMonitoringPath = path.join(
  repositoryRoot,
  ".github/workflows/error-monitoring.yml",
);

test("VOC-086-TEST-11: scheduled synthetics map to required stable IDs", () => {
  const errors = validateScheduledSyntheticsFiles(repositoryRoot);
  assert.deepEqual(errors, [], errors.join("; "));

  const syntheticsDocument = loadSyntheticsRegistry(repositoryRoot);
  const workflowSource = readFileSync(workflowPath, "utf8");

  for (const syntheticId of CANONICAL_SYNTHETIC_IDS) {
    assert.match(
      workflowSource,
      new RegExp(syntheticId.replaceAll(".", "\\.")),
      `workflow must reference stable id ${syntheticId}`,
    );
  }

  for (const checkRef of CANONICAL_CHECK_REFS) {
    const entry = syntheticsDocument.synthetics.find(
      (item) => item.check_ref === checkRef,
    );
    assert.ok(entry, `registry must contain check_ref ${checkRef}`);
    assert.equal(
      entry.workflow_ref,
      SCHEDULED_SYNTHETICS_WORKFLOW,
      `${entry.id} must point at scheduled-synthetics workflow`,
    );
  }

  assert.ok(existsSync(runSyntheticScriptPath));
  const dispatcher = readFileSync(runSyntheticScriptPath, "utf8");
  for (const checkRef of CANONICAL_CHECK_REFS) {
    if (checkRef === "staging-authenticated-core-journey") {
      assert.match(
        dispatcher,
        /staging-authenticated-core-journey/,
        "dispatcher must document Playwright handoff for staging core-loop",
      );
      continue;
    }
    assert.match(
      dispatcher,
      new RegExp(checkRef.replaceAll("-", "\\-")),
      `dispatcher must handle ${checkRef}`,
    );
  }
});

test("VOC-086-TEST-12: synthetic checks reuse mint secrets and mask sessions", () => {
  const workflowSource = readFileSync(workflowPath, "utf8");
  const mintScript = readFileSync(mintSessionScriptPath, "utf8");
  const dispatcher = readFileSync(runSyntheticScriptPath, "utf8");

  assert.match(workflowSource, /STAGING_SMOKE_TEST_SESSION_MINT_TOKEN/);
  assert.match(workflowSource, /PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN/);
  assert.match(workflowSource, /mint-synthetic-session\.sh/);
  assert.match(mintScript, /::add-mask::/);
  assert.match(mintScript, /SMOKE_TEST_SESSION_MINT_TOKEN/);
  assert.match(
    mintScript,
    /empty session or csrf value/,
    "mint script must fail closed when csrf_token is empty (deploy-staging parity)",
  );
  assert.doesNotMatch(
    mintScript,
    /echo\s+["']?\$session/,
    "mint script must not echo the session variable",
  );

  for (const jobId of [
    "production-journey-content",
    "production-authenticated-route-content-sweep",
  ]) {
    const jobBlock = extractTopLevelJobBlock(workflowSource, jobId);
    assert.ok(jobBlock, `workflow must define job ${jobId}`);
    assert.ok(
      jobBlock.includes(EXPECT_OAUTH_ENABLED_FROM_SECRETS),
      `${jobId} must set EXPECT_OAUTH_ENABLED from deploy-production's secrets-present expression`,
    );
    assert.match(jobBlock, /EXPECT_MAGIC_LINK_ENABLED: "false"/);
    assert.match(jobBlock, /EXPECT_NEW_SIGNUPS_ENABLED: "false"/);
    assert.match(jobBlock, /EXPECT_AI_ENABLED: "true"/);
  }

  const fixture =
    "mint failed password=super-secret KUMA_PASSWORD=abc123 Bearer eyJhbGciOiJIUzI1NiJ9";
  const redacted = redactSecrets(fixture);
  assert.ok(!redacted.includes("super-secret"));
  assert.ok(!redacted.includes("abc123"));
  assert.match(redacted, /\[REDACTED\]/);

  const syntheticsDocument = loadSyntheticsRegistry(repositoryRoot);
  const productionSweep = syntheticsDocument.synthetics.find(
    (item) =>
      item.id === "synthetic.production.authenticated-route-content-sweep",
  );
  assert.equal(productionSweep?.mutating, false);

  assert.match(
    dispatcher,
    /SMOKE_TEST_PROFILE=route-sweep/,
    "production route sweep must use non-mutating profile",
  );
  assert.match(dispatcher, /SMOKE_TEST_SESSION_COOKIE/);

  const smokeScript = readFileSync(
    path.join(repositoryRoot, "infra/scripts/smoke-test-production.sh"),
    "utf8",
  );
  assert.match(
    smokeScript,
    /non-mutating GET/,
    "production smoke script documents non-mutating route sweep",
  );
});

test("VOC-086-TEST-13: error-monitoring.yml remains separate", () => {
  const workflowSource = readFileSync(workflowPath, "utf8");
  const errorMonitoring = readFileSync(errorMonitoringPath, "utf8");

  assert.doesNotMatch(
    workflowSource,
    /uses:.*error-monitoring|workflow_call:.*error-monitoring/,
    "scheduled synthetics must not replace or call error-monitoring as a sub-workflow",
  );

  assert.match(errorMonitoring, /name: error-monitoring/);
  assert.match(errorMonitoring, /SENTRY_API_TOKEN/);
  assert.match(errorMonitoring, /schedule:/);

  const errors = validateScheduledSyntheticsWorkflow({
    workflowSource,
    syntheticsDocument: loadSyntheticsRegistry(repositoryRoot),
    repositoryRoot,
  });
  assert.ok(
    !errors.some((error) => error.includes("error-monitoring")),
    `workflow must not call error-monitoring: ${errors.join("; ")}`,
  );
});

test("VOC-086-TEST-12: dispatcher error messages do not echo cookie values", () => {
  const dispatcher = readFileSync(runSyntheticScriptPath, "utf8");
  assert.doesNotMatch(dispatcher, /echo\s+["']?\$SMOKE_TEST_SESSION_COOKIE/);
  assert.doesNotMatch(dispatcher, /printf\s+.*SMOKE_TEST_SESSION_COOKIE/);
});

test("VOC-086-TEST-11: production smoke-profile jobs fail closed without EXPECT_OAUTH_ENABLED", () => {
  const workflowSource = readFileSync(workflowPath, "utf8");
  const syntheticsDocument = loadSyntheticsRegistry(repositoryRoot);
  const stripped = workflowSource.replaceAll(
    EXPECT_OAUTH_ENABLED_FROM_SECRETS,
    "EXPECT_OAUTH_ENABLED: false",
  );
  const journeyBlock = extractTopLevelJobBlock(
    stripped,
    "production-journey-content",
  );
  assert.ok(journeyBlock);
  assert.ok(
    !journeyBlock.includes(EXPECT_OAUTH_ENABLED_FROM_SECRETS),
    "fixture must omit the secrets-present EXPECT_OAUTH_ENABLED expression",
  );

  const errors = validateScheduledSyntheticsWorkflow({
    workflowSource: stripped,
    syntheticsDocument,
    repositoryRoot,
  });
  assert.ok(
    errors.some(
      (error) =>
        error.includes("production-journey-content") &&
        error.includes("EXPECT_OAUTH_ENABLED"),
    ),
    `validator must reject missing secrets-present EXPECT_OAUTH_ENABLED: ${errors.join("; ")}`,
  );
  assert.ok(
    errors.some(
      (error) =>
        error.includes("production-authenticated-route-content-sweep") &&
        error.includes("EXPECT_OAUTH_ENABLED"),
    ),
    `validator must reject missing secrets-present EXPECT_OAUTH_ENABLED on route-sweep: ${errors.join("; ")}`,
  );
});
