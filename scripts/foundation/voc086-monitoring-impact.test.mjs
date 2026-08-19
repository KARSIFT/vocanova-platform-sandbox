// VOC-086-T04 — monitoring_impact governance validation.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import {
  CANONICAL_AVAILABILITY_MONITOR_IDS,
  CANONICAL_SYNTHETIC_IDS,
} from "../../infra/monitoring/validate-inventory.mjs";
import {
  collectAffectedPackages,
  isRouteOrCriticalEndpointPath,
  loadCanonicalMonitoringIds,
  parseMonitoringImpactFromChangeYaml,
  validateDeclaredMonitoringImpactFiles,
  validateMonitoringImpact,
  validateMonitoringImpactDeclaration,
} from "../../infra/monitoring/validate-monitoring-impact.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const voc086ChangeYaml = readFileSync(
  path.join(
    repositoryRoot,
    "specs/changes/VOC-086-manage-monitoring-inventory/change.yaml",
  ),
  "utf8",
);

function buildFixtureRepository() {
  const root = mkdtempSync(path.join(tmpdir(), "voc086-monitoring-impact-"));
  mkdirSync(path.join(root, "infra/monitoring"), { recursive: true });
  mkdirSync(path.join(root, "specs/changes"), { recursive: true });

  writeFileSync(
    path.join(root, "infra/monitoring/monitors.yaml"),
    readFileSync(
      path.join(repositoryRoot, "infra/monitoring/monitors.yaml"),
      "utf8",
    ),
  );
  writeFileSync(
    path.join(root, "infra/monitoring/synthetics.yaml"),
    readFileSync(
      path.join(repositoryRoot, "infra/monitoring/synthetics.yaml"),
      "utf8",
    ),
  );

  return root;
}

function writePackage(root, slug, changeYaml) {
  const packageDir = path.join(root, "specs/changes", slug);
  mkdirSync(packageDir, { recursive: true });
  writeFileSync(path.join(packageDir, "change.yaml"), changeYaml);
}

test("VOC-086-TEST-14: monitoring_impact positive cases validate", () => {
  const canonicalIds = loadCanonicalMonitoringIds(repositoryRoot);

  const noneErrors = validateMonitoringImpactDeclaration(
    parseMonitoringImpactFromChangeYaml(`
monitoring_impact:
  state: none
  rationale: Documentation-only package with no route or monitoring inventory change.
`),
    canonicalIds,
    "fixture-none",
  );
  assert.deepEqual(noneErrors, []);

  const existingErrors = validateMonitoringImpactDeclaration(
    parseMonitoringImpactFromChangeYaml(`
monitoring_impact:
  state: existing
  monitor_ids:
    - kuma.availability.production.web
  synthetic_ids:
    - synthetic.production.journey-content
`),
    canonicalIds,
    "fixture-existing",
  );
  assert.deepEqual(existingErrors, []);

  const addErrors = validateMonitoringImpactDeclaration(
    parseMonitoringImpactFromChangeYaml(`
monitoring_impact:
  state: add
  monitor_ids:
    - kuma.availability.staging.web
  synthetic_ids:
    - synthetic.staging.oauth-expected-state
`),
    canonicalIds,
    "fixture-add",
  );
  assert.deepEqual(addErrors, []);

  const updateErrors = validateMonitoringImpactDeclaration(
    parseMonitoringImpactFromChangeYaml(`
monitoring_impact:
  state: update
  synthetic_ids:
    - synthetic.production.authenticated-route-content-sweep
`),
    canonicalIds,
    "fixture-update",
  );
  assert.deepEqual(updateErrors, []);

  const emptySyntheticList = validateMonitoringImpactDeclaration(
    parseMonitoringImpactFromChangeYaml(`
monitoring_impact:
  state: update
  monitor_ids:
    - kuma.availability.production.web
  synthetic_ids: []
`),
    canonicalIds,
    "fixture-empty-inline-list",
  );
  assert.deepEqual(emptySyntheticList, [], emptySyntheticList.join("; "));

  const voc086Errors = validateMonitoringImpactDeclaration(
    parseMonitoringImpactFromChangeYaml(voc086ChangeYaml),
    canonicalIds,
    "specs/changes/VOC-086-manage-monitoring-inventory/change.yaml",
  );
  assert.deepEqual(voc086Errors, [], voc086Errors.join("; "));
});

test("VOC-086-TEST-15: monitoring_impact negative cases fail closed", () => {
  const canonicalIds = loadCanonicalMonitoringIds(repositoryRoot);

  const missingBlock = validateMonitoringImpactDeclaration(
    null,
    canonicalIds,
    "fixture-missing",
  );
  assert.ok(
    missingBlock.some((error) => error.includes("missing or invalid")),
    missingBlock.join("; "),
  );

  const noneWithoutRationale = validateMonitoringImpactDeclaration(
    { state: "none" },
    canonicalIds,
    "fixture-none",
  );
  assert.ok(
    noneWithoutRationale.some((error) =>
      error.includes("requires a non-empty rationale"),
    ),
  );

  const noneWithIds = validateMonitoringImpactDeclaration(
    {
      state: "none",
      rationale: "Should not list ids.",
      monitor_ids: [CANONICAL_AVAILABILITY_MONITOR_IDS[0]],
    },
    canonicalIds,
    "fixture-none",
  );
  assert.ok(
    noneWithIds.some((error) => error.includes("must not list monitor_ids")),
  );

  const invalidState = validateMonitoringImpactDeclaration(
    { state: "maybe-later" },
    canonicalIds,
    "fixture-state",
  );
  assert.ok(
    invalidState.some((error) => error.includes("state must be one of")),
  );

  const invalidIds = validateMonitoringImpactDeclaration(
    {
      state: "add",
      monitor_ids: ["kuma.availability.unknown"],
      synthetic_ids: ["synthetic.unknown.check"],
    },
    canonicalIds,
    "fixture-ids",
  );
  assert.ok(invalidIds.some((error) => error.includes("unknown monitor id")));
  assert.ok(invalidIds.some((error) => error.includes("unknown synthetic id")));
});

test("VOC-086-TEST-16: route/critical-endpoint changes require valid monitoring_impact; historical packages grandfathered", () => {
  const fixtureRoot = buildFixtureRepository();
  const canonicalIds = loadCanonicalMonitoringIds(fixtureRoot);

  const historicalSlug = "VOC-050-historical-package";
  writePackage(
    fixtureRoot,
    historicalSlug,
    `id: VOC-050
status: adopted
`,
  );

  const historicalOnlyErrors = validateMonitoringImpact({
    repositoryRoot: fixtureRoot,
    changedFiles: [`specs/changes/${historicalSlug}/tasks.md`],
    canonicalIds,
  });
  assert.deepEqual(
    historicalOnlyErrors,
    [],
    "unchanged historical change.yaml must remain grandfathered",
  );

  const routeChangeErrors = validateMonitoringImpact({
    repositoryRoot: fixtureRoot,
    changedFiles: ["apps/web/src/app/(app)/home/page.tsx"],
    canonicalIds,
  });
  assert.ok(
    routeChangeErrors.some((error) =>
      error.includes("route/critical-endpoint changes require"),
    ),
    routeChangeErrors.join("; "),
  );

  const newPackageSlug = "VOC-999-new-route-package";
  writePackage(
    fixtureRoot,
    newPackageSlug,
    `id: VOC-999
monitoring_impact:
  state: add
  monitor_ids:
    - kuma.availability.staging.web
  synthetic_ids:
    - synthetic.staging.authenticated-core-journey
`,
  );

  const validRouteErrors = validateMonitoringImpact({
    repositoryRoot: fixtureRoot,
    changedFiles: [
      `specs/changes/${newPackageSlug}/change.yaml`,
      "apps/web/src/app/(app)/discover/page.tsx",
    ],
    canonicalIds,
  });
  assert.deepEqual(validRouteErrors, [], validRouteErrors.join("; "));

  writePackage(
    fixtureRoot,
    newPackageSlug,
    `id: VOC-999
monitoring_impact:
  state: none
  rationale: No monitoring inventory change required for this fixture.
`,
  );

  const modifiedWithoutImpactErrors = validateMonitoringImpact({
    repositoryRoot: fixtureRoot,
    changedFiles: [`specs/changes/${newPackageSlug}/change.yaml`],
    canonicalIds,
  });
  assert.deepEqual(modifiedWithoutImpactErrors, []);

  writePackage(
    fixtureRoot,
    newPackageSlug,
    `id: VOC-999
`,
  );

  const modifiedMissingImpactErrors = validateMonitoringImpact({
    repositoryRoot: fixtureRoot,
    changedFiles: [`specs/changes/${newPackageSlug}/change.yaml`],
    canonicalIds,
  });
  assert.ok(
    modifiedMissingImpactErrors.some((error) =>
      error.includes("monitoring_impact block is missing"),
    ),
  );
});

test("VOC-086-TEST-16: route/critical path classifier covers pages and API handlers", () => {
  assert.equal(
    isRouteOrCriticalEndpointPath("apps/web/src/app/(app)/home/page.tsx"),
    true,
  );
  assert.equal(isRouteOrCriticalEndpointPath("apps/api/app/api/auth.go"), true);
  assert.equal(
    isRouteOrCriticalEndpointPath("apps/api/app/api/auth_test.go"),
    false,
  );
  assert.equal(
    isRouteOrCriticalEndpointPath("apps/web/src/components/button.tsx"),
    false,
  );

  const affected = collectAffectedPackages(
    [
      "specs/changes/VOC-100-example/change.yaml",
      "specs/changes/VOC-050-historical/tasks.md",
    ],
    { newPackageSlugs: new Set(["VOC-100-example"]) },
  );
  assert.deepEqual(affected.map((entry) => entry.slug).sort(), [
    "VOC-100-example",
  ]);
});

test("VOC-086-TEST-14: declared monitoring_impact files in the repository validate", () => {
  const errors = validateDeclaredMonitoringImpactFiles(repositoryRoot);
  assert.deepEqual(errors, [], errors.join("; "));
  assert.ok(
    CANONICAL_AVAILABILITY_MONITOR_IDS.length === 5 &&
      CANONICAL_SYNTHETIC_IDS.length === 5,
    "canonical inventory must expose five monitor and five synthetic ids",
  );
});

test("VOC-086-TEST-16: CI passes pull-request base/head into monitoring_impact validation", () => {
  const governancePolicy = readFileSync(
    path.join(repositoryRoot, ".github/workflows/governance-policy.yml"),
    "utf8",
  );
  const repositoryGovernance = readFileSync(
    path.join(repositoryRoot, ".github/workflows/repository-governance.yml"),
    "utf8",
  );
  const validateGovernance = readFileSync(
    path.join(repositoryRoot, "scripts/governance/validate-governance.sh"),
    "utf8",
  );
  const wrapper = readFileSync(
    path.join(
      repositoryRoot,
      "scripts/governance/validate-monitoring-impact.sh",
    ),
    "utf8",
  );

  assert.match(governancePolicy, /github\.event\.pull_request\.base\.sha/);
  assert.match(governancePolicy, /github\.event\.pull_request\.head\.sha/);
  assert.match(
    governancePolicy,
    /validate-governance\.sh[\s\S]*--base "\$BASE_SHA"[\s\S]*--head "\$HEAD_SHA"/,
  );

  assert.match(repositoryGovernance, /fetch-depth: 0/);
  assert.match(
    repositoryGovernance,
    /validate-governance\.sh --base "\$BASE_SHA" --head "\$HEAD_SHA"/,
  );
  assert.match(
    repositoryGovernance,
    /bash -n scripts\/governance\/validate-monitoring-impact\.sh/,
  );

  assert.match(
    validateGovernance,
    /monitoring_impact_args\+=\(--base "\$base" --head "\$head"\)/,
  );
  assert.match(
    validateGovernance,
    /validate-monitoring-impact\.sh "\$\{monitoring_impact_args\[@\]\}"/,
  );

  assert.doesNotMatch(wrapper, /base="\$\{GITHUB_BASE_SHA/);
  assert.match(wrapper, /GITHUB_EVENT_PATH/);
  assert.match(
    wrapper,
    /pull_request monitoring_impact validation requires --base\/--head/,
  );
});

test("VOC-086-TEST-16: pull_request without a changed-file range fails closed", () => {
  const wrapper = path.join(
    repositoryRoot,
    "scripts/governance/validate-monitoring-impact.sh",
  );
  const isolatedEnv = { ...process.env };
  delete isolatedEnv.GITHUB_EVENT_PATH;
  isolatedEnv.GITHUB_EVENT_NAME = "pull_request";

  const missingRange = spawnSync("bash", [wrapper], {
    env: isolatedEnv,
    encoding: "utf8",
  });
  assert.notEqual(missingRange.status, 0);
  assert.match(
    missingRange.stderr,
    /requires --base\/--head, --files-from, or a parseable GITHUB_EVENT_PATH/,
  );

  const filesFrom = path.join(
    mkdtempSync(path.join(tmpdir(), "voc086-impact-files-")),
    "changed.txt",
  );
  writeFileSync(filesFrom, "docs/development.md\n");
  const withFilesFrom = spawnSync(
    "bash",
    [wrapper, "--files-from", filesFrom],
    {
      env: isolatedEnv,
      encoding: "utf8",
    },
  );
  assert.equal(
    withFilesFrom.status,
    0,
    withFilesFrom.stderr || withFilesFrom.stdout,
  );
});
