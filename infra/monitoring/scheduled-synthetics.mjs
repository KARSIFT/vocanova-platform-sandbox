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
      errors.push(`scheduled synthetics registry missing check_ref ${checkRef}`);
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
    errors.push("scheduled-synthetics workflow must declare name: scheduled-synthetics");
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
