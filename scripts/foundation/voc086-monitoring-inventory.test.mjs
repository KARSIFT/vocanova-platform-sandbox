// VOC-086-T00 — canonical monitoring inventory schema validation.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import {
  CANONICAL_AVAILABILITY_MONITOR_IDS,
  CANONICAL_SYNTHETIC_IDS,
  parseMonitoringYaml,
  validateMonitoringInventoryFiles,
  validateMonitoringInventorySources,
} from "../../infra/monitoring/validate-inventory.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

test("VOC-086-TEST-00: canonical inventory schema accepts complete entries", () => {
  const errors = validateMonitoringInventoryFiles(repositoryRoot);
  assert.deepEqual(errors, [], errors.join("; "));

  const monitorsPath = path.join(
    repositoryRoot,
    "infra/monitoring/monitors.yaml",
  );
  const syntheticsPath = path.join(
    repositoryRoot,
    "infra/monitoring/synthetics.yaml",
  );

  const monitors = parseMonitoringYaml(readFileSync(monitorsPath, "utf8"));
  assert.equal(monitors.schema_version, 1);
  assert.equal(monitors.kuma.ownership_marker, "vocanova:repo-managed");

  const availabilityIds = monitors.availability_monitors.map(
    (monitor) => monitor.id,
  );
  assert.deepEqual(
    [...availabilityIds].sort(),
    [...CANONICAL_AVAILABILITY_MONITOR_IDS].sort(),
  );

  const synthetics = parseMonitoringYaml(readFileSync(syntheticsPath, "utf8"));
  const syntheticIds = synthetics.synthetics.map((entry) => entry.id);
  assert.deepEqual(
    [...syntheticIds].sort(),
    [...CANONICAL_SYNTHETIC_IDS].sort(),
  );

  for (const monitor of monitors.availability_monitors) {
    assert.ok(monitor.name, `monitor ${monitor.id} must have a name`);
    assert.ok(monitor.url, `monitor ${monitor.id} must have a url`);
    assert.ok(
      Array.isArray(monitor.coverage) && monitor.coverage.length > 0,
      `monitor ${monitor.id} must declare coverage`,
    );
  }

  for (const synthetic of synthetics.synthetics) {
    assert.ok(
      synthetic.workflow_ref,
      `synthetic ${synthetic.id} must have workflow_ref`,
    );
    assert.ok(
      synthetic.check_ref,
      `synthetic ${synthetic.id} must have check_ref`,
    );
    assert.ok(
      Array.isArray(synthetic.coverage) && synthetic.coverage.length > 0,
      `synthetic ${synthetic.id} must declare coverage`,
    );
  }
});

test("VOC-086-TEST-01: inventory schema rejects missing fields", () => {
  const monitorsSource = `
schema_version: 1
kuma:
  ownership_marker: vocanova:repo-managed
availability_monitors:
  - id: kuma.availability.staging.web
    name: Staging Web
    environment: staging
    owner: vocanova-platform
    type: http
    url: https://staging.vocanova.site/
    expected_status: 200
    expected_body: null
    interval_seconds: 60
    timeout_seconds: 30
    retries: 2
    severity: high
`;

  const syntheticsSource = `
schema_version: 1
synthetics:
  - id: synthetic.staging.oauth-expected-state
    name: Staging OAuth
    environment: staging
    owner: vocanova-platform
    type: synthetic
    workflow_ref: .github/workflows/scheduled-synthetics.yml
    check_ref: staging-oauth-expected-state
    expected_status: 200
    expected_body: null
    schedule: "0 * * * *"
    timeout_seconds: 120
    retries: 1
    severity: high
`;

  const errors = validateMonitoringInventorySources({
    monitorsSource,
    syntheticsSource,
    requireCanonicalIds: false,
  });

  assert.ok(
    errors.some((error) => error.includes("missing required field coverage")),
    `expected missing coverage errors, got: ${errors.join("; ")}`,
  );
});

test("VOC-086-TEST-01: inventory schema rejects duplicate ids", () => {
  const monitorsSource = `
schema_version: 1
kuma:
  ownership_marker: vocanova:repo-managed
availability_monitors:
  - id: kuma.availability.staging.web
    name: Staging Web A
    environment: staging
    owner: vocanova-platform
    type: http
    url: https://staging.vocanova.site/
    expected_status: 200
    expected_body: null
    interval_seconds: 60
    timeout_seconds: 30
    retries: 2
    severity: high
    coverage:
      - "page:staging.vocanova.site/"
  - id: kuma.availability.staging.web
    name: Staging Web B
    environment: staging
    owner: vocanova-platform
    type: http
    url: https://staging.vocanova.site/
    expected_status: 200
    expected_body: null
    interval_seconds: 60
    timeout_seconds: 30
    retries: 2
    severity: high
    coverage:
      - "page:staging.vocanova.site/"
`;

  const syntheticsSource = `
schema_version: 1
synthetics:
  - id: synthetic.staging.oauth-expected-state
    name: OAuth A
    environment: staging
    owner: vocanova-platform
    type: synthetic
    workflow_ref: .github/workflows/scheduled-synthetics.yml
    check_ref: staging-oauth-expected-state
    expected_status: 200
    expected_body: null
    schedule: "0 * * * *"
    timeout_seconds: 120
    retries: 1
    severity: high
    coverage:
      - "api:POST /api/v1/auth/oauth/google/start"
  - id: synthetic.staging.oauth-expected-state
    name: OAuth B
    environment: staging
    owner: vocanova-platform
    type: synthetic
    workflow_ref: .github/workflows/scheduled-synthetics.yml
    check_ref: staging-oauth-expected-state
    expected_status: 200
    expected_body: null
    schedule: "0 * * * *"
    timeout_seconds: 120
    retries: 1
    severity: high
    coverage:
      - "api:POST /api/v1/auth/oauth/google/start"
`;

  const errors = validateMonitoringInventorySources({
    monitorsSource,
    syntheticsSource,
    requireCanonicalIds: false,
  });

  assert.ok(
    errors.some((error) =>
      error.includes("duplicate id kuma.availability.staging.web"),
    ),
    `expected duplicate availability id error, got: ${errors.join("; ")}`,
  );
  assert.ok(
    errors.some((error) =>
      error.includes("duplicate id synthetic.staging.oauth-expected-state"),
    ),
    `expected duplicate synthetic id error, got: ${errors.join("; ")}`,
  );
  assert.ok(
    !errors.some((error) => /password|secret|token/i.test(error)),
    "validation errors must not echo secret-like fixture values",
  );
});
