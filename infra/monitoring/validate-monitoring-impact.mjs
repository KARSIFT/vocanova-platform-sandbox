import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { parseMonitoringYaml } from "./validate-inventory.mjs";

export const VALID_MONITORING_IMPACT_STATES = new Set([
  "none",
  "existing",
  "add",
  "update",
]);

const PAGE_ROUTE_PATTERN = /^apps\/web\/src\/app\/(?:.+\/)?page\.tsx$/;

const NEXT_ROUTE_HANDLER_PATTERN =
  /^apps\/web\/src\/app\/(?:.+\/)?route\.(?:t|j)sx?$/;

const API_ROUTE_HANDLER_PATTERN = /^apps\/api\/app\/api\/[^/]+\.go$/;

function isNonEmptyString(value) {
  return typeof value === "string" && value.trim().length > 0;
}

function parseScalarList(lines, startIndex, parentIndent) {
  const values = [];
  let index = startIndex;
  while (index < lines.length) {
    const line = lines[index];
    if (!line.trim()) {
      index += 1;
      continue;
    }
    const indent = line.match(/^ */)?.[0].length ?? 0;
    if (indent < parentIndent) {
      break;
    }
    const trimmed = line.trim();
    if (!trimmed.startsWith("- ")) {
      break;
    }
    values.push(parseScalar(trimmed.slice(2)));
    index += 1;
  }
  return { values, nextIndex: index };
}

function parseScalar(raw) {
  const trimmed = raw.trim();
  if (trimmed === "null" || trimmed === "~" || trimmed === "") {
    return null;
  }
  if (
    (trimmed.startsWith('"') && trimmed.endsWith('"')) ||
    (trimmed.startsWith("'") && trimmed.endsWith("'"))
  ) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

function parseInlineYamlList(raw) {
  const trimmed = raw.trim();
  if (trimmed === "[]") {
    return [];
  }
  return null;
}

export function parseMonitoringImpactFromChangeYaml(source) {
  if (!isNonEmptyString(source)) {
    return null;
  }

  const lines = source.split(/\r?\n/);
  let start = -1;
  for (let index = 0; index < lines.length; index += 1) {
    if (/^monitoring_impact:\s*$/.test(lines[index])) {
      start = index + 1;
      break;
    }
  }
  if (start === -1) {
    return null;
  }

  const block = [];
  for (let index = start; index < lines.length; index += 1) {
    const line = lines[index];
    if (/^[A-Za-z_][A-Za-z0-9_-]*:\s*/.test(line) && !/^\s/.test(line)) {
      break;
    }
    block.push(line);
  }

  if (block.length === 0) {
    return {};
  }

  const minIndent = block.reduce((currentMin, line) => {
    if (!line.trim()) {
      return currentMin;
    }
    const indent = line.match(/^ */)?.[0].length ?? 0;
    return Math.min(currentMin, indent);
  }, Number.POSITIVE_INFINITY);

  const normalized = block.map((line) =>
    line.trim() ? line.slice(minIndent) : "",
  );

  const declaration = {};
  let index = 0;
  while (index < normalized.length) {
    const line = normalized[index];
    if (!line.trim()) {
      index += 1;
      continue;
    }

    const match = line.match(/^([A-Za-z_][A-Za-z0-9_-]*):\s*(.*)$/);
    if (!match) {
      index += 1;
      continue;
    }

    const key = match[1];
    const inlineValue = match[2];
    if (key === "monitor_ids" || key === "synthetic_ids") {
      const inlineList = parseInlineYamlList(inlineValue);
      if (inlineList) {
        declaration[key] = inlineList;
        index += 1;
      } else if (inlineValue) {
        declaration[key] = [parseScalar(inlineValue)];
        index += 1;
      } else {
        const parsed = parseScalarList(normalized, index + 1, 0);
        declaration[key] = parsed.values;
        index = parsed.nextIndex;
      }
      continue;
    }

    if (inlineValue) {
      declaration[key] = parseScalar(inlineValue);
      index += 1;
      continue;
    }

    const continuation = [];
    index += 1;
    while (index < normalized.length) {
      const nextLine = normalized[index];
      if (!nextLine.trim()) {
        index += 1;
        continue;
      }
      if (/^[A-Za-z_][A-Za-z0-9_-]*:\s*/.test(nextLine)) {
        break;
      }
      if (nextLine.startsWith("- ")) {
        break;
      }
      continuation.push(nextLine.trim());
      index += 1;
    }
    declaration[key] = continuation.join(" ");
  }

  return declaration;
}

export function loadCanonicalMonitoringIds(repositoryRoot) {
  const monitorsPath = path.join(
    repositoryRoot,
    "infra/monitoring/monitors.yaml",
  );
  const syntheticsPath = path.join(
    repositoryRoot,
    "infra/monitoring/synthetics.yaml",
  );

  if (!existsSync(monitorsPath) || !existsSync(syntheticsPath)) {
    throw new Error(
      "canonical monitoring inventory is missing under infra/monitoring/",
    );
  }

  const monitorsDocument = parseMonitoringYaml(
    readFileSync(monitorsPath, "utf8"),
  );
  const syntheticsDocument = parseMonitoringYaml(
    readFileSync(syntheticsPath, "utf8"),
  );

  const monitorIds = new Set(
    (monitorsDocument.availability_monitors ?? []).map((entry) => entry.id),
  );
  const syntheticIds = new Set(
    (syntheticsDocument.synthetics ?? []).map((entry) => entry.id),
  );

  return { monitorIds, syntheticIds };
}

export function validateMonitoringImpactDeclaration(
  declaration,
  { monitorIds, syntheticIds },
  label = "monitoring_impact",
) {
  const errors = [];

  if (!declaration || typeof declaration !== "object") {
    errors.push(`${label}: monitoring_impact block is missing or invalid`);
    return errors;
  }

  const state = declaration.state;
  if (!VALID_MONITORING_IMPACT_STATES.has(state)) {
    errors.push(
      `${label}: state must be one of ${[...VALID_MONITORING_IMPACT_STATES].join(", ")}`,
    );
    return errors;
  }

  const listedMonitorIds = Array.isArray(declaration.monitor_ids)
    ? declaration.monitor_ids
    : [];
  const listedSyntheticIds = Array.isArray(declaration.synthetic_ids)
    ? declaration.synthetic_ids
    : [];

  if (state === "none") {
    if (!isNonEmptyString(declaration.rationale)) {
      errors.push(`${label}: state none requires a non-empty rationale`);
    }
    if (listedMonitorIds.length > 0 || listedSyntheticIds.length > 0) {
      errors.push(
        `${label}: state none must not list monitor_ids or synthetic_ids`,
      );
    }
    return errors;
  }

  if (listedMonitorIds.length === 0 && listedSyntheticIds.length === 0) {
    errors.push(
      `${label}: state ${state} requires at least one monitor_ids or synthetic_ids entry`,
    );
  }

  for (const id of listedMonitorIds) {
    if (!isNonEmptyString(id)) {
      errors.push(`${label}: monitor_ids entries must be non-empty strings`);
      continue;
    }
    if (!monitorIds.has(id)) {
      errors.push(`${label}: unknown monitor id ${id}`);
    }
  }

  for (const id of listedSyntheticIds) {
    if (!isNonEmptyString(id)) {
      errors.push(`${label}: synthetic_ids entries must be non-empty strings`);
      continue;
    }
    if (!syntheticIds.has(id)) {
      errors.push(`${label}: unknown synthetic id ${id}`);
    }
  }

  return errors;
}

export function isRouteOrCriticalEndpointPath(filePath) {
  if (filePath.endsWith("_test.go")) {
    return false;
  }
  return (
    PAGE_ROUTE_PATTERN.test(filePath) ||
    NEXT_ROUTE_HANDLER_PATTERN.test(filePath) ||
    API_ROUTE_HANDLER_PATTERN.test(filePath)
  );
}

export function packageSlugFromPath(filePath) {
  const match = filePath.match(/^specs\/changes\/([^/]+)\//);
  return match ? match[1] : null;
}

export function collectAffectedPackages(
  changedFiles,
  { newPackageSlugs = new Set() } = {},
) {
  const packages = new Map();

  for (const filePath of changedFiles) {
    const slug = packageSlugFromPath(filePath);
    if (!slug) {
      continue;
    }

    const current = packages.get(slug) ?? {
      slug,
      changeYamlTouched: false,
    };

    if (filePath === `specs/changes/${slug}/change.yaml`) {
      current.changeYamlTouched = true;
    }
    packages.set(slug, current);
  }

  return [...packages.values()].filter(
    (entry) => entry.changeYamlTouched || newPackageSlugs.has(entry.slug),
  );
}

export function validateMonitoringImpact({
  repositoryRoot,
  changedFiles = [],
  packageChangeYamlBySlug = new Map(),
  canonicalIds,
}) {
  const errors = [];
  const ids = canonicalIds ?? loadCanonicalMonitoringIds(repositoryRoot);

  const packagesNeedingImpact = collectAffectedPackages(changedFiles, {
    newPackageSlugs: new Set(
      [...packageChangeYamlBySlug.entries()]
        .filter(([, metadata]) => metadata.isNewPackage)
        .map(([slug]) => slug),
    ),
  });

  for (const entry of packagesNeedingImpact) {
    const changeYamlPath = path.join(
      repositoryRoot,
      "specs/changes",
      entry.slug,
      "change.yaml",
    );
    const label = `specs/changes/${entry.slug}/change.yaml`;
    const source =
      packageChangeYamlBySlug.get(entry.slug)?.source ??
      (existsSync(changeYamlPath) ? readFileSync(changeYamlPath, "utf8") : "");

    if (!isNonEmptyString(source)) {
      errors.push(
        `${label}: missing change.yaml for package requiring monitoring_impact`,
      );
      continue;
    }

    const declaration = parseMonitoringImpactFromChangeYaml(source);
    errors.push(
      ...validateMonitoringImpactDeclaration(declaration, ids, label),
    );
  }

  const routeOrCriticalChanges = changedFiles.filter(
    isRouteOrCriticalEndpointPath,
  );
  if (routeOrCriticalChanges.length > 0) {
    const candidateSlugs = new Set(
      changedFiles.map(packageSlugFromPath).filter(Boolean),
    );
    let validRouteImpact = false;

    for (const slug of candidateSlugs) {
      const changeYamlPath = path.join(
        repositoryRoot,
        "specs/changes",
        slug,
        "change.yaml",
      );
      const source =
        packageChangeYamlBySlug.get(slug)?.source ??
        (existsSync(changeYamlPath)
          ? readFileSync(changeYamlPath, "utf8")
          : "");
      const declaration = parseMonitoringImpactFromChangeYaml(source);
      if (
        validateMonitoringImpactDeclaration(declaration, ids, slug).length === 0
      ) {
        validRouteImpact = true;
        break;
      }
    }

    if (!validRouteImpact) {
      errors.push(
        "route/critical-endpoint changes require a change package with valid monitoring_impact " +
          `(touched: ${routeOrCriticalChanges.join(", ")})`,
      );
    }
  }

  return errors;
}

export function validateDeclaredMonitoringImpactFiles(
  repositoryRoot,
  canonicalIds,
) {
  const errors = [];
  const changesRoot = path.join(repositoryRoot, "specs/changes");
  if (!existsSync(changesRoot)) {
    return ["missing specs/changes directory"];
  }

  const ids = canonicalIds ?? loadCanonicalMonitoringIds(repositoryRoot);

  for (const slug of readdirSync(changesRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)) {
    const changeYamlPath = path.join(changesRoot, slug, "change.yaml");
    if (!existsSync(changeYamlPath)) {
      continue;
    }
    const source = readFileSync(changeYamlPath, "utf8");
    if (!parseMonitoringImpactFromChangeYaml(source)) {
      continue;
    }
    const label = `specs/changes/${slug}/change.yaml`;
    errors.push(
      ...validateMonitoringImpactDeclaration(
        parseMonitoringImpactFromChangeYaml(source),
        ids,
        label,
      ),
    );
  }

  return errors;
}

function parseArgs(argv) {
  const options = {
    repositoryRoot: process.cwd(),
    changedFiles: [],
    declarationsOnly: false,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--repository-root") {
      options.repositoryRoot = path.resolve(argv[index + 1] ?? "");
      index += 1;
      continue;
    }
    if (arg === "--changed-file") {
      options.changedFiles.push(argv[index + 1] ?? "");
      index += 1;
      continue;
    }
    if (arg === "--changed-files-file") {
      const filePath = argv[index + 1] ?? "";
      index += 1;
      if (!existsSync(filePath)) {
        throw new Error(`missing changed-files file: ${filePath}`);
      }
      const listed = readFileSync(filePath, "utf8")
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter(Boolean);
      options.changedFiles.push(...listed);
      continue;
    }
    if (arg === "--declarations-only") {
      options.declarationsOnly = true;
      continue;
    }
    if (arg === "-h" || arg === "--help") {
      options.help = true;
      continue;
    }
    throw new Error(`unknown argument: ${arg}`);
  }

  return options;
}

function printHelp() {
  process.stdout.write(`Usage: validate-monitoring-impact.mjs [options]

Options:
  --repository-root PATH   Repository root (default: cwd)
  --changed-file PATH      Changed file path (repeatable)
  --changed-files-file PATH  Newline-delimited changed file paths
  --declarations-only      Validate only existing monitoring_impact declarations
`);
}

const modulePath = fileURLToPath(import.meta.url);
if (process.argv[1] === modulePath) {
  try {
    const options = parseArgs(process.argv.slice(2));
    if (options.help) {
      printHelp();
      process.exit(0);
    }

    const canonicalIds = loadCanonicalMonitoringIds(options.repositoryRoot);
    const errors = options.declarationsOnly
      ? validateDeclaredMonitoringImpactFiles(
          options.repositoryRoot,
          canonicalIds,
        )
      : validateMonitoringImpact({
          repositoryRoot: options.repositoryRoot,
          changedFiles: options.changedFiles,
          canonicalIds,
        });

    if (errors.length > 0) {
      process.stderr.write(`${errors.join("\n")}\n`);
      process.exitCode = 1;
    } else {
      process.stdout.write("Monitoring impact validation passed.\n");
    }
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exitCode = 1;
  }
}
