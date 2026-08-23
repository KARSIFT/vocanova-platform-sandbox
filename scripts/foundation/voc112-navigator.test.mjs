// VOC-112-T01 — vocanova-repo-navigator routing table and router-budget assertions.
//
// Runs via `node --test scripts/foundation/voc112-navigator.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { BUDGETS } from "./voc112-agent-skills.test.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const navigatorSkillPath = path.join(
  repositoryRoot,
  ".agents/skills/vocanova-repo-navigator/SKILL.md",
);

/** Router-specific budget (stricter than generic SKILL.md caps in T00). */
export const ROUTER_BUDGETS = {
  bodyMaxBytes: 4096,
  bodyMaxLines: 120,
};

/** Minimum VOC-112-D05 domains with representative path tokens. */
const REQUIRED_ROUTES = [
  {
    id: "web",
    label: "Web UI / Next.js",
    pathTokens: [
      "apps/web/",
      "docs/design/08-web-app-design.md",
      "docs/development.md",
    ],
  },
  {
    id: "api",
    label: "API / Go backend",
    pathTokens: [
      "apps/api/",
      "docs/engineering/06-backend-design.md",
      "docs/engineering/07-api-contract-and-dto-design.md",
    ],
  },
  {
    id: "database",
    label: "Database / migrations",
    pathTokens: [
      "apps/api/migrations/",
      "docs/engineering/05-database-design.md",
    ],
  },
  {
    id: "auth",
    label: "Auth / OAuth",
    pathTokens: [
      "apps/web/src/app/auth/",
      "docs/operations/staging-controlled-signup.md",
      "docs/operations/production-controlled-signup.md",
      "specs/changes/",
    ],
  },
  {
    id: "seed",
    label: "Content seed",
    pathTokens: ["apps/api/cmd/seed/"],
  },
  {
    id: "deploy",
    label: "Deploy / infra / shared edge",
    pathTokens: [
      "infra/",
      ".github/workflows/deploy-",
      "docs/operations/11-devops-and-ci-cd.md",
      "scripts/foundation/voc079-single-edge-invariants.test.mjs",
    ],
  },
  {
    id: "monitoring",
    label: "Monitoring",
    pathTokens: ["infra/monitoring/", "docs/operations/monitoring.md"],
  },
  {
    id: "governance",
    label: "Governance / change workflow",
    pathTokens: [
      "AGENTS.md",
      "docs/governance/",
      "specs/changes/",
      "specs/templates/change-package/",
    ],
  },
  {
    id: "validation",
    label: "Validation / tests",
    pathTokens: ["docs/development.md", "pnpm validate", "scripts/foundation/"],
  },
  {
    id: "lifecycle",
    label: "Issue → plan → task lifecycle",
    pathTokens: ["AGENTS.md", "specs/changes/"],
    extraTokens: ["Reporting a bug", "change workflow"],
  },
];

/** Distinctive AGENTS.md phrases that must not be pasted into the router. */
const GOVERNANCE_PASTE_MARKERS = [
  "A-004 is the effective authority model",
  "automatic_merge_allowed",
  "monitoring_impact",
  "Reconciling a merged plan PR whose adoption handoff was missed",
  "The only bootstrap exception is the initial DOC-16/A-002 adoption",
];

function readNavigatorSkill() {
  return readFileSync(navigatorSkillPath, "utf8");
}

function parseSkillBody(source) {
  const match = source.match(/^---\r?\n[\s\S]*?\r?\n---\r?\n?([\s\S]*)$/);
  assert.ok(match, "navigator SKILL.md must include YAML frontmatter");
  return match[1];
}

function pathExists(candidate) {
  if (candidate.includes("*")) {
    const dir = path.dirname(candidate);
    const base = path.basename(candidate);
    const resolvedDir = path.resolve(repositoryRoot, dir);
    if (!statSync(resolvedDir, { throwIfNoEntry: false })?.isDirectory()) {
      return false;
    }
    return readdirSync(resolvedDir).some((entry) =>
      globBasenameMatches(entry, base),
    );
  }
  return Boolean(
    statSync(path.resolve(repositoryRoot, candidate), {
      throwIfNoEntry: false,
    }),
  );
}

function globBasenameMatches(entry, pattern) {
  const expression = pattern
    .split("*")
    .map((part) => part.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"))
    .join(".*");
  return new RegExp(`^${expression}$`).test(entry);
}

function parseRoutingRows(body) {
  const rows = new Map();
  for (const line of body.split(/\r?\n/)) {
    const match = line.match(/^\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|$/);
    if (!match || match[1] === "Intent" || /^-+$/.test(match[1])) {
      continue;
    }
    rows.set(match[1], match[2]);
  }
  return rows;
}

function missingRequiredRouteMappings(body) {
  const rows = parseRoutingRows(body);
  const missing = [];

  for (const route of REQUIRED_ROUTES) {
    const row = rows.get(route.label);
    if (!row) {
      missing.push(`${route.id}:row`);
      continue;
    }
    for (const token of [...route.pathTokens, ...(route.extraTokens ?? [])]) {
      if (!row.includes(token)) {
        missing.push(`${route.id}:${token}`);
      }
    }
  }

  return missing;
}

function extractBacktickPaths(source) {
  return [...source.matchAll(/`([^`]+)`/g)].map(([, token]) => token.trim());
}

test("VOC-112-TEST-06: navigator covers required domains with existing targets", () => {
  const source = readNavigatorSkill();
  const body = parseSkillBody(source);

  assert.match(body, /## Routing table/);
  assert.match(body, /\| Intent \| Start here \|/);

  assert.deepEqual(
    missingRequiredRouteMappings(body),
    [],
    "every required path and lifecycle token must be in its named routing row",
  );

  const referencedPaths = extractBacktickPaths(body).filter(
    (token) =>
      !token.includes("pnpm") &&
      !token.includes("Reporting a bug") &&
      !token.includes("change workflow") &&
      !token.includes(".env") &&
      !token.startsWith("CLAUDE_"),
  );

  const missing = [];
  for (const candidate of referencedPaths) {
    if (!pathExists(candidate)) {
      missing.push(candidate);
    }
  }
  assert.deepEqual(
    missing,
    [],
    `navigator references paths that do not exist: ${missing.join(", ")}`,
  );
});

test("VOC-112-TEST-06: misplaced routes and false glob matches fail closed", () => {
  const body = parseSkillBody(readNavigatorSkill());
  const misplaced = body
    .replace("`docs/operations/staging-controlled-signup.md`, ", "")
    .replace(
      "| Web UI / Next.js |",
      "| Web UI / Next.js | `docs/operations/staging-controlled-signup.md`,",
    );

  assert.ok(
    missingRequiredRouteMappings(misplaced).includes(
      "auth:docs/operations/staging-controlled-signup.md",
    ),
    "a token moved to another row must not satisfy the auth route",
  );
  assert.equal(globBasenameMatches("example.test.mjs", "*.test.mjs"), true);
  assert.equal(globBasenameMatches("README.md", "*.test.mjs"), false);
});

test("VOC-112-TEST-07: navigator stays compact and states governance precedence", () => {
  const source = readNavigatorSkill();
  const body = parseSkillBody(source);
  const bodyBytes = Buffer.byteLength(body, "utf8");
  const bodyLines = body.split(/\r?\n/).length;
  const descriptionMatch = source.match(
    /^---\r?\n[\s\S]*?description:\s*(.+)\r?\n/,
  );
  assert.ok(descriptionMatch, "navigator must declare description frontmatter");
  const description = descriptionMatch[1].replace(/^["']|["']$/g, "");

  assert.match(body, /Governance precedence/i);
  assert.match(
    body,
    /repository sources win|canonical.*win/i,
    "must state governance precedence outcome",
  );
  assert.match(body, /router only/i, "must identify itself as a router");
  assert.doesNotMatch(
    body,
    /Follow DOC-15, DOC-16/,
    "must not paste governance corpora inline",
  );

  for (const marker of GOVERNANCE_PASTE_MARKERS) {
    assert.doesNotMatch(
      body,
      new RegExp(marker.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")),
      `must not paste AGENTS.md marker: ${marker}`,
    );
  }

  assert.ok(
    bodyBytes <= ROUTER_BUDGETS.bodyMaxBytes,
    `router body exceeds ${ROUTER_BUDGETS.bodyMaxBytes} bytes (${bodyBytes})`,
  );
  assert.ok(
    bodyLines <= ROUTER_BUDGETS.bodyMaxLines,
    `router body exceeds ${ROUTER_BUDGETS.bodyMaxLines} lines (${bodyLines})`,
  );
  assert.ok(
    bodyBytes <= BUDGETS.skillBodyMaxBytes,
    "router must also satisfy generic SKILL.md byte budget",
  );
  assert.ok(
    description.length <= BUDGETS.descriptionMaxChars,
    "description must satisfy generic frontmatter budget",
  );
});

test("VOC-112-TEST-06 supplemental: shared-edge and validation tiers are routed", () => {
  const body = parseSkillBody(readNavigatorSkill());

  assert.match(body, /voc079-single-edge-invariants/);
  assert.match(body, /pnpm validate/);
  assert.match(body, /docs\/development\.md/);
});
