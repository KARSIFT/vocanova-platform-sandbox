import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

export const CANONICAL_AVAILABILITY_MONITOR_IDS = [
  "kuma.availability.staging.web",
  "kuma.availability.staging.api-healthz",
  "kuma.availability.production.web",
  "kuma.availability.production.api-healthz",
  "kuma.availability.monitor-host",
];

export const CANONICAL_SYNTHETIC_IDS = [
  "synthetic.staging.oauth-expected-state",
  "synthetic.production.oauth-expected-state",
  "synthetic.production.journey-content",
  "synthetic.staging.authenticated-core-journey",
  "synthetic.production.authenticated-route-content-sweep",
];

const AVAILABILITY_REQUIRED_FIELDS = [
  "id",
  "name",
  "environment",
  "owner",
  "type",
  "url",
  "expected_status",
  "expected_body",
  "interval_seconds",
  "timeout_seconds",
  "retries",
  "severity",
  "coverage",
];

const SYNTHETIC_REQUIRED_FIELDS = [
  "id",
  "name",
  "environment",
  "owner",
  "type",
  "workflow_ref",
  "check_ref",
  "expected_status",
  "expected_body",
  "schedule",
  "timeout_seconds",
  "retries",
  "severity",
  "coverage",
];

const VALID_ENVIRONMENTS = new Set(["staging", "production", "shared"]);

const VALID_SEVERITIES = new Set(["low", "medium", "high", "critical"]);

const VALID_MONITOR_TYPES = new Set(["http"]);

const VALID_SYNTHETIC_TYPES = new Set(["synthetic"]);

const ID_PATTERN = /^[a-z0-9][a-z0-9.-]*[a-z0-9]$/;

function parseScalar(raw) {
  const trimmed = raw.trim();
  if (trimmed === "null" || trimmed === "~" || trimmed === "") {
    return null;
  }
  if (trimmed === "true") {
    return true;
  }
  if (trimmed === "false") {
    return false;
  }
  if (/^-?\d+$/.test(trimmed)) {
    return Number(trimmed);
  }
  if (
    (trimmed.startsWith('"') && trimmed.endsWith('"')) ||
    (trimmed.startsWith("'") && trimmed.endsWith("'"))
  ) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

function indentOf(line) {
  const match = line.match(/^ */);
  return match ? match[0].length : 0;
}

function parseYamlDocument(source) {
  const lines = source
    .split(/\r?\n/)
    .map((line, index) => ({
      line: line.replace(/\t/g, "  "),
      number: index + 1,
    }))
    .filter(({ line }) => {
      const trimmed = line.trim();
      return trimmed && !trimmed.startsWith("#");
    });

  let pos = 0;

  function peek() {
    return lines[pos];
  }

  function next() {
    return lines[pos++];
  }

  function parseList(minIndent) {
    const items = [];
    while (pos < lines.length) {
      const current = peek();
      const indent = indentOf(current.line);
      if (indent < minIndent) {
        break;
      }
      if (!current.line.trim().startsWith("- ")) {
        break;
      }

      next();
      const afterDash = current.line.trim().slice(2);
      if (
        afterDash.startsWith('"') ||
        afterDash.startsWith("'") ||
        !afterDash.includes(":")
      ) {
        items.push(parseScalar(afterDash));
        continue;
      }

      const colonIndex = afterDash.indexOf(":");
      const key = afterDash.slice(0, colonIndex).trim();
      const rest = afterDash.slice(colonIndex + 1).trim();
      const item = {};
      if (rest) {
        item[key] = parseScalar(rest);
      }

      if (pos < lines.length && indentOf(peek().line) > indent) {
        Object.assign(item, parseBlock(indent + 2));
      }
      items.push(item);
    }
    return items;
  }

  function parseBlock(minIndent) {
    const result = {};
    while (pos < lines.length) {
      const current = peek();
      const indent = indentOf(current.line);
      if (indent < minIndent) {
        break;
      }
      if (indent > minIndent) {
        throw new Error(`unexpected indent at line ${current.number}`);
      }

      const trimmed = current.line.trim();
      if (trimmed.startsWith("- ")) {
        break;
      }

      next();
      const colonIndex = trimmed.indexOf(":");
      if (colonIndex === -1) {
        throw new Error(`missing colon at line ${current.number}`);
      }

      const key = trimmed.slice(0, colonIndex).trim();
      const rest = trimmed.slice(colonIndex + 1).trim();
      if (!rest) {
        if (pos < lines.length && indentOf(peek().line) === indent + 2) {
          if (peek().line.trim().startsWith("- ")) {
            result[key] = parseList(indent + 2);
          } else {
            result[key] = parseBlock(indent + 2);
          }
        } else {
          result[key] = {};
        }
      } else {
        result[key] = parseScalar(rest);
      }
    }
    return result;
  }

  return parseBlock(0);
}

export function parseMonitoringYaml(source) {
  return parseYamlDocument(source);
}

function isNonEmptyString(value) {
  return typeof value === "string" && value.trim().length > 0;
}

function isNonNegativeInteger(value) {
  return Number.isInteger(value) && value >= 0;
}

function validateCoverage(coverage, label, errors) {
  if (!Array.isArray(coverage) || coverage.length === 0) {
    errors.push(`${label}: coverage must be a non-empty array`);
    return;
  }
  for (const entry of coverage) {
    if (!isNonEmptyString(entry)) {
      errors.push(`${label}: coverage entries must be non-empty strings`);
    }
  }
}

function validateAvailabilityMonitor(monitor, errors) {
  const label = `availability monitor ${monitor?.id ?? "<missing-id>"}`;

  for (const field of AVAILABILITY_REQUIRED_FIELDS) {
    if (!(field in monitor)) {
      errors.push(`${label}: missing required field ${field}`);
    }
  }

  if (!isNonEmptyString(monitor.id)) {
    errors.push(`${label}: id must be a non-empty string`);
  } else if (!ID_PATTERN.test(monitor.id)) {
    errors.push(`${label}: id has invalid format`);
  }

  if (!isNonEmptyString(monitor.name)) {
    errors.push(`${label}: name must be a non-empty string`);
  }

  if (!VALID_ENVIRONMENTS.has(monitor.environment)) {
    errors.push(
      `${label}: environment must be one of ${[...VALID_ENVIRONMENTS].join(", ")}`,
    );
  }

  if (!isNonEmptyString(monitor.owner)) {
    errors.push(`${label}: owner must be a non-empty string`);
  }

  if (!VALID_MONITOR_TYPES.has(monitor.type)) {
    errors.push(
      `${label}: type must be ${[...VALID_MONITOR_TYPES].join(", ")}`,
    );
  }

  if (!isNonEmptyString(monitor.url)) {
    errors.push(`${label}: url must be a non-empty string`);
  } else {
    try {
      const parsed = new URL(monitor.url);
      if (!["http:", "https:"].includes(parsed.protocol)) {
        errors.push(`${label}: url must use http or https`);
      }
    } catch {
      errors.push(`${label}: url must be a valid absolute URL`);
    }
  }

  if (!isNonNegativeInteger(monitor.expected_status)) {
    errors.push(`${label}: expected_status must be a non-negative integer`);
  }

  if (
    monitor.expected_body !== null &&
    typeof monitor.expected_body !== "string"
  ) {
    errors.push(`${label}: expected_body must be a string or null`);
  }

  if (
    !isNonNegativeInteger(monitor.interval_seconds) ||
    monitor.interval_seconds <= 0
  ) {
    errors.push(`${label}: interval_seconds must be a positive integer`);
  }

  if (
    !isNonNegativeInteger(monitor.timeout_seconds) ||
    monitor.timeout_seconds <= 0
  ) {
    errors.push(`${label}: timeout_seconds must be a positive integer`);
  }

  if (!isNonNegativeInteger(monitor.retries)) {
    errors.push(`${label}: retries must be a non-negative integer`);
  }

  if (!VALID_SEVERITIES.has(monitor.severity)) {
    errors.push(
      `${label}: severity must be one of ${[...VALID_SEVERITIES].join(", ")}`,
    );
  }

  validateCoverage(monitor.coverage, label, errors);

  if (monitor.adoption !== undefined) {
    if (typeof monitor.adoption !== "object" || monitor.adoption === null) {
      errors.push(`${label}: adoption must be an object when present`);
    } else {
      if (!isNonEmptyString(monitor.adoption.match_name)) {
        errors.push(`${label}: adoption.match_name must be a non-empty string`);
      }
      if (!isNonEmptyString(monitor.adoption.match_url)) {
        errors.push(`${label}: adoption.match_url must be a non-empty string`);
      }
    }
  }

  if (Object.prototype.hasOwnProperty.call(monitor, "notification_id_list")) {
    const notificationIDList = monitor.notification_id_list;
    if (
      notificationIDList === null ||
      typeof notificationIDList !== "object" ||
      Array.isArray(notificationIDList)
    ) {
      errors.push(
        `${label}: notification_id_list must be an object when present`,
      );
    } else {
      for (const [notificationId, enabled] of Object.entries(
        notificationIDList,
      )) {
        if (!/^[1-9]\d*$/.test(notificationId)) {
          errors.push(
            `${label}: notification_id_list keys must be positive integer IDs`,
          );
        }
        if (typeof enabled !== "boolean") {
          errors.push(`${label}: notification_id_list values must be booleans`);
        }
      }
    }
  }
}

function validateSynthetic(synthetic, errors) {
  const label = `synthetic ${synthetic?.id ?? "<missing-id>"}`;

  for (const field of SYNTHETIC_REQUIRED_FIELDS) {
    if (!(field in synthetic)) {
      errors.push(`${label}: missing required field ${field}`);
    }
  }

  if (!isNonEmptyString(synthetic.id)) {
    errors.push(`${label}: id must be a non-empty string`);
  } else if (!ID_PATTERN.test(synthetic.id)) {
    errors.push(`${label}: id has invalid format`);
  }

  if (!isNonEmptyString(synthetic.name)) {
    errors.push(`${label}: name must be a non-empty string`);
  }

  if (!["staging", "production"].includes(synthetic.environment)) {
    errors.push(`${label}: environment must be staging or production`);
  }

  if (!isNonEmptyString(synthetic.owner)) {
    errors.push(`${label}: owner must be a non-empty string`);
  }

  if (!VALID_SYNTHETIC_TYPES.has(synthetic.type)) {
    errors.push(
      `${label}: type must be ${[...VALID_SYNTHETIC_TYPES].join(", ")}`,
    );
  }

  if (!isNonEmptyString(synthetic.workflow_ref)) {
    errors.push(`${label}: workflow_ref must be a non-empty string`);
  }

  if (!isNonEmptyString(synthetic.check_ref)) {
    errors.push(`${label}: check_ref must be a non-empty string`);
  }

  if (!isNonNegativeInteger(synthetic.expected_status)) {
    errors.push(`${label}: expected_status must be a non-negative integer`);
  }

  if (
    synthetic.expected_body !== null &&
    typeof synthetic.expected_body !== "string"
  ) {
    errors.push(`${label}: expected_body must be a string or null`);
  }

  if (!isNonEmptyString(synthetic.schedule)) {
    errors.push(`${label}: schedule must be a non-empty cron string`);
  }

  if (
    !isNonNegativeInteger(synthetic.timeout_seconds) ||
    synthetic.timeout_seconds <= 0
  ) {
    errors.push(`${label}: timeout_seconds must be a positive integer`);
  }

  if (!isNonNegativeInteger(synthetic.retries)) {
    errors.push(`${label}: retries must be a non-negative integer`);
  }

  if (!VALID_SEVERITIES.has(synthetic.severity)) {
    errors.push(
      `${label}: severity must be one of ${[...VALID_SEVERITIES].join(", ")}`,
    );
  }

  validateCoverage(synthetic.coverage, label, errors);

  if (
    synthetic.mutating !== undefined &&
    typeof synthetic.mutating !== "boolean"
  ) {
    errors.push(`${label}: mutating must be a boolean when present`);
  }
}

function collectDuplicateIds(entries, errors, kind) {
  const seen = new Map();
  for (const entry of entries) {
    if (!isNonEmptyString(entry?.id)) {
      continue;
    }
    if (seen.has(entry.id)) {
      errors.push(`${kind}: duplicate id ${entry.id}`);
    } else {
      seen.set(entry.id, true);
    }
  }
}

function assertCanonicalIds(ids, expected, label, errors) {
  const actual = [...ids].sort();
  const canonical = [...expected].sort();
  if (actual.length !== canonical.length) {
    errors.push(
      `${label}: expected ${canonical.length} canonical ids, found ${actual.length}`,
    );
    return;
  }
  for (let index = 0; index < canonical.length; index += 1) {
    if (actual[index] !== canonical[index]) {
      errors.push(
        `${label}: id mismatch at position ${index} (expected ${canonical[index]}, found ${actual[index] ?? "<missing>"})`,
      );
    }
  }
}

export function validateMonitoringInventory({
  monitorsDocument,
  syntheticsDocument,
  requireCanonicalIds = true,
} = {}) {
  const errors = [];

  if (!monitorsDocument || typeof monitorsDocument !== "object") {
    errors.push("monitors document must be an object");
    return errors;
  }

  if (!syntheticsDocument || typeof syntheticsDocument !== "object") {
    errors.push("synthetics document must be an object");
    return errors;
  }

  if (monitorsDocument.schema_version !== 1) {
    errors.push("monitors.schema_version must be 1");
  }

  if (syntheticsDocument.schema_version !== 1) {
    errors.push("synthetics.schema_version must be 1");
  }

  if (!isNonEmptyString(monitorsDocument?.kuma?.ownership_marker)) {
    errors.push("monitors.kuma.ownership_marker must be a non-empty string");
  }

  const availabilityMonitors = monitorsDocument.availability_monitors;
  if (!Array.isArray(availabilityMonitors)) {
    errors.push("monitors.availability_monitors must be an array");
  } else {
    collectDuplicateIds(availabilityMonitors, errors, "availability monitors");
    for (const monitor of availabilityMonitors) {
      validateAvailabilityMonitor(monitor, errors);
    }
    if (requireCanonicalIds) {
      assertCanonicalIds(
        availabilityMonitors.map((monitor) => monitor.id),
        CANONICAL_AVAILABILITY_MONITOR_IDS,
        "availability monitors",
        errors,
      );
    }
  }

  const synthetics = syntheticsDocument.synthetics;
  if (!Array.isArray(synthetics)) {
    errors.push("synthetics.synthetics must be an array");
  } else {
    collectDuplicateIds(synthetics, errors, "synthetics");
    for (const synthetic of synthetics) {
      validateSynthetic(synthetic, errors);
    }
    if (requireCanonicalIds) {
      assertCanonicalIds(
        synthetics.map((synthetic) => synthetic.id),
        CANONICAL_SYNTHETIC_IDS,
        "synthetics",
        errors,
      );
    }
  }

  return errors;
}

export function validateMonitoringInventorySources({
  monitorsSource,
  syntheticsSource,
  requireCanonicalIds = true,
} = {}) {
  let monitorsDocument;
  let syntheticsDocument;

  try {
    monitorsDocument = parseMonitoringYaml(monitorsSource ?? "");
  } catch (error) {
    return [`monitors yaml parse error: ${error.message}`];
  }

  try {
    syntheticsDocument = parseMonitoringYaml(syntheticsSource ?? "");
  } catch (error) {
    return [`synthetics yaml parse error: ${error.message}`];
  }

  return validateMonitoringInventory({
    monitorsDocument,
    syntheticsDocument,
    requireCanonicalIds,
  });
}

export function validateMonitoringInventoryFiles(
  repositoryRoot,
  { requireCanonicalIds = true } = {},
) {
  const monitorsPath = path.join(
    repositoryRoot,
    "infra/monitoring/monitors.yaml",
  );
  const syntheticsPath = path.join(
    repositoryRoot,
    "infra/monitoring/synthetics.yaml",
  );

  if (!existsSync(monitorsPath)) {
    return ["missing infra/monitoring/monitors.yaml"];
  }
  if (!existsSync(syntheticsPath)) {
    return ["missing infra/monitoring/synthetics.yaml"];
  }

  return validateMonitoringInventorySources({
    monitorsSource: readFileSync(monitorsPath, "utf8"),
    syntheticsSource: readFileSync(syntheticsPath, "utf8"),
    requireCanonicalIds,
  });
}

const modulePath = fileURLToPath(import.meta.url);
if (process.argv[1] === modulePath) {
  const repositoryRoot = path.resolve(path.dirname(modulePath), "../..");
  const errors = validateMonitoringInventoryFiles(repositoryRoot);
  if (errors.length) {
    process.stderr.write(`${errors.join("\n")}\n`);
    process.exitCode = 1;
  } else {
    process.stdout.write("Monitoring inventory validation passed.\n");
  }
}
