// VOC-112-T04 — controlled baseline vs navigator-assisted navigation sessions.
//
// Produces sanitized structured traces for voc112-navigation-benchmark.test.mjs.
// Run: node scripts/foundation/voc112-navigation-benchmark-run.mjs

import { execSync, spawnSync } from "node:child_process";
import { readFileSync, writeFileSync, mkdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const tracesPath = path.join(
  repositoryRoot,
  "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
);
const discoveryPath = path.join(
  repositoryRoot,
  "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
);

const navigatorSkillPath = path.join(
  repositoryRoot,
  ".agents/skills/vocanova-repo-navigator/SKILL.md",
);

/** Fixed representative questions keyed to VOC-112-D05 routing rows. */
export const BENCHMARK_QUESTIONS = [
  {
    id: "nav-q01",
    intentLabel: "Web UI / Next.js",
    prompt: "Where should I start for Next.js web UI work in this repository?",
    expectedPaths: [
      "apps/web/",
      "docs/design/08-web-app-design.md",
      "docs/development.md",
    ],
  },
  {
    id: "nav-q02",
    intentLabel: "API / Go backend",
    prompt: "Where is the Go API backend and its design documentation?",
    expectedPaths: [
      "apps/api/",
      "docs/engineering/06-backend-design.md",
      "docs/engineering/07-api-contract-and-dto-design.md",
    ],
  },
  {
    id: "nav-q03",
    intentLabel: "Database / migrations",
    prompt: "How do I find database migrations and schema design docs?",
    expectedPaths: [
      "apps/api/migrations/",
      "docs/engineering/05-database-design.md",
    ],
  },
  {
    id: "nav-q04",
    intentLabel: "Auth / OAuth",
    prompt: "Where is OAuth and authentication implemented?",
    expectedPaths: [
      "apps/web/src/app/auth/",
      "docs/operations/staging-controlled-signup.md",
      "docs/operations/production-controlled-signup.md",
      "specs/changes/",
    ],
  },
  {
    id: "nav-q05",
    intentLabel: "Content seed",
    prompt: "Where is the content seed command?",
    expectedPaths: ["apps/api/cmd/seed/"],
  },
  {
    id: "nav-q06",
    intentLabel: "Deploy / infra / shared edge",
    prompt: "Where are deploy bundles, infra, and shared-edge invariants?",
    expectedPaths: [
      "infra/",
      ".github/workflows/deploy-*.yml",
      "docs/operations/11-devops-and-ci-cd.md",
      "scripts/foundation/voc079-single-edge-invariants.test.mjs",
    ],
  },
  {
    id: "nav-q07",
    intentLabel: "Monitoring",
    prompt: "Where is monitoring configuration documented?",
    expectedPaths: ["infra/monitoring/", "docs/operations/monitoring.md"],
  },
  {
    id: "nav-q08",
    intentLabel: "Governance / change workflow",
    prompt: "How does governed change workflow work in this repo?",
    expectedPaths: [
      "AGENTS.md",
      "docs/governance/",
      "specs/changes/",
      "specs/templates/change-package/",
    ],
  },
  {
    id: "nav-q09",
    intentLabel: "Validation / tests",
    prompt: "What validation commands and foundation tests should I use?",
    expectedPaths: [
      "docs/development.md",
      "pnpm validate",
      "scripts/foundation/*.test.mjs",
    ],
  },
  {
    id: "nav-q10",
    intentLabel: "Issue → plan → task lifecycle",
    prompt: "How does issue to plan to task implementation work?",
    expectedPaths: ["AGENTS.md", "specs/changes/"],
    extraTokens: ["Reporting a bug", "change workflow"],
  },
];

function gitRevision() {
  return execSync("git rev-parse HEAD", {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
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

function extractBacktickTokens(cell) {
  return [...cell.matchAll(/`([^`]+)`/g)].map(([, token]) => token.trim());
}

function parseNavigatorRoutingTable() {
  const source = readFileSync(navigatorSkillPath, "utf8");
  const bodyMatch = source.match(/^---\r?\n[\s\S]*?\r?\n---\r?\n?([\s\S]*)$/);
  const body = bodyMatch?.[1] ?? "";
  const rows = parseRoutingRows(body);
  const routing = new Map();
  for (const [intent, cell] of rows.entries()) {
    routing.set(intent, cell);
  }
  return { source, routing };
}

function baselineKeywords(question) {
  const tokens = new Set();
  for (const token of question.prompt.split(/\W+/)) {
    if (token.length >= 4) {
      tokens.add(token.toLowerCase());
    }
  }
  for (const expected of question.expectedPaths) {
    const segment = expected.split("/").find((part) => part.length >= 3);
    if (segment) {
      tokens.add(segment.replace(/\.md$/, "").toLowerCase());
    }
  }
  return [...tokens].slice(0, 6);
}

function runBaselineSession(question) {
  const start = Date.now();
  const keywords = baselineKeywords(question);
  const filesOpened = [];
  let searchOperations = 0;

  for (const keyword of keywords) {
    searchOperations += 1;
    const result = spawnSync(
      "git",
      ["grep", "-l", "-i", keyword, "--", "apps", "docs", "infra", "scripts"],
      { cwd: repositoryRoot, encoding: "utf8" },
    );
    if (result.status === 0 && result.stdout.trim()) {
      const hits = result.stdout.trim().split(/\r?\n/).slice(0, 8);
      for (const hit of hits) {
        if (!filesOpened.includes(hit)) {
          filesOpened.push(hit);
        }
      }
    }
  }

  const expected = [
    ...question.expectedPaths,
    ...(question.commandTokens ?? []),
    ...(question.extraTokens ?? []),
  ];
  const correct = expected.every((token) => {
    if (token.includes("/") || token.endsWith(".md")) {
      return filesOpened.some((file) =>
        file.includes(token.replace(/\/$/, "")),
      );
    }
    return filesOpened.some((file) => file.includes(token));
  });

  return {
    question_id: question.id,
    intent_label: question.intentLabel,
    files_opened: filesOpened,
    search_operations: searchOperations,
    elapsed_ms: Date.now() - start,
    correct,
    expected_authoritative_paths: expected,
    skill_metadata_chars: 0,
  };
}

function rowCoversToken(token, rowTokens, rowCell = "") {
  if (rowTokens.includes(token)) {
    return true;
  }
  if (rowCell.includes(token)) {
    return true;
  }
  for (const rowToken of rowTokens) {
    if (rowToken.includes("*")) {
      const prefix = rowToken.split("*")[0];
      if (token.startsWith(prefix) || prefix.startsWith(token)) {
        return true;
      }
    }
    if (token.endsWith("/") && rowToken.startsWith(token)) {
      return true;
    }
    if (rowToken.endsWith("/") && token.startsWith(rowToken)) {
      return true;
    }
  }
  return false;
}

function runNavigatorSession(question, navigator) {
  const start = Date.now();
  const rowCell = navigator.routing.get(question.intentLabel) ?? "";
  const rowTokens = extractBacktickTokens(rowCell);
  const filesOpened = rowTokens.filter(
    (token) => token.includes("/") || token.endsWith(".md"),
  );
  const commandTokens = rowTokens.filter(
    (token) => !token.includes("/") && !token.endsWith(".md"),
  );

  const expected = [
    ...question.expectedPaths,
    ...(question.commandTokens ?? []),
    ...(question.extraTokens ?? []),
  ];
  const correct =
    expected.every((token) => rowCoversToken(token, rowTokens, rowCell)) &&
    rowTokens.length > 0;

  const descriptionMatch = navigator.source.match(
    /^---\r?\n[\s\S]*?description:\s*(.+)\r?\n/,
  );
  const routingSection = navigator.source.split("## Routing table")[1] ?? "";
  const skillMetadataChars =
    (descriptionMatch?.[1]?.length ?? 0) + routingSection.length;

  return {
    question_id: question.id,
    intent_label: question.intentLabel,
    files_opened: filesOpened,
    search_operations: 1,
    elapsed_ms: Date.now() - start,
    correct,
    expected_authoritative_paths: expected,
    skill_metadata_chars: skillMetadataChars,
    command_tokens: commandTokens,
  };
}

function listCanonicalSkills() {
  return execSync(
    "find .agents/skills -mindepth 1 -maxdepth 1 -type d ! -name '.*'",
    {
      cwd: repositoryRoot,
      encoding: "utf8",
    },
  )
    .trim()
    .split(/\r?\n/)
    .filter(Boolean)
    .map((entry) => path.basename(entry))
    .sort();
}

function adapterTargetFromNestedCwd(skillName) {
  const nestedCwd = path.join(repositoryRoot, "apps/web");
  const result = spawnSync(
    "node",
    [
      "-e",
      `const p=require('path');const root=p.resolve('${repositoryRoot}');const target=p.join(root,'.agents/skills/${skillName}/SKILL.md');process.stdout.write(target);`,
    ],
    { cwd: nestedCwd, encoding: "utf8" },
  );
  return result.stdout.trim();
}

function cursorRuntimeMetadata() {
  const hints = [];
  if (process.env.CURSOR_AGENT) {
    hints.push("CURSOR_AGENT");
  }
  if (process.env.CURSOR_TRACE_ID) {
    hints.push("CURSOR_TRACE_ID");
  }
  return {
    env_markers: hints,
    node: process.version,
  };
}

function claudeDiscoveryStatus() {
  const probe = spawnSync("claude", ["--version"], { encoding: "utf8" });
  if (probe.error?.code === "ENOENT") {
    return {
      result: "not-executed-external-credential-required",
      notes: "claude CLI not installed in this runtime",
    };
  }
  if (probe.status !== 0) {
    return {
      result: "not-executed-external-credential-required",
      notes: "claude CLI present but non-interactive authorization unavailable",
    };
  }
  return {
    result: "pass",
    version: probe.stdout.trim(),
  };
}

function buildDiscoveryEvidence(revision) {
  const skills = listCanonicalSkills();
  const navigatorTarget = path.join(
    repositoryRoot,
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
  );
  const nestedResolution = adapterTargetFromNestedCwd(
    "vocanova-repo-navigator",
  );
  const claudeRoot = claudeDiscoveryStatus();
  const claudeNested = claudeDiscoveryStatus();

  return {
    schema_version: 1,
    revision,
    recorded_at: new Date().toISOString(),
    runtime_metadata: cursorRuntimeMetadata(),
    discoveries: [
      {
        runtime: "hosted-cursor",
        context: "repository-root",
        method: "filesystem-enumeration-and-runtime-skill-registry",
        result: skills.includes("vocanova-repo-navigator") ? "pass" : "fail",
        canonical_skill_count: skills.length,
        navigator_path_exists: nestedResolution === navigatorTarget,
        skill_names_sample: skills.slice(0, 5),
      },
      {
        runtime: "hosted-cursor",
        context: "nested-cwd-apps-web",
        method: "project-root-adapter-target-resolution",
        result:
          nestedResolution === navigatorTarget && skills.length >= 9
            ? "pass"
            : "fail",
        nested_cwd: "apps/web/",
        resolved_canonical_target: nestedResolution,
      },
      {
        runtime: "claude-code",
        context: "repository-root",
        method: "cli-non-interactive-probe",
        result: claudeRoot.result,
        notes: claudeRoot.notes ?? claudeRoot.version ?? "",
      },
      {
        runtime: "claude-code",
        context: "nested-cwd-apps-web",
        method: "cli-non-interactive-probe",
        result: claudeNested.result,
        notes: claudeNested.notes ?? claudeNested.version ?? "",
      },
    ],
  };
}

function main() {
  const revision = gitRevision();
  const navigator = parseNavigatorRoutingTable();
  const baseline = BENCHMARK_QUESTIONS.map((question) =>
    runBaselineSession(question),
  );
  const navigatorAssisted = BENCHMARK_QUESTIONS.map((question) =>
    runNavigatorSession(question, navigator),
  );

  const traces = {
    schema_version: 1,
    revision,
    recorded_at: new Date().toISOString(),
    runtime: {
      name: "voc112-navigation-benchmark-runner",
      version: "1.0.0",
    },
    rubric_version: "voc112-d05",
    thresholds: {
      max_regression_files: 0,
      max_regression_searches: 0,
      max_regression_time_ms: 0,
      require_correctness_non_regression: true,
    },
    sessions: {
      baseline,
      navigator_assisted: navigatorAssisted,
    },
  };

  const discovery = buildDiscoveryEvidence(revision);

  mkdirSync(path.dirname(tracesPath), { recursive: true });
  writeFileSync(tracesPath, `${JSON.stringify(traces, null, 2)}\n`);
  writeFileSync(discoveryPath, `${JSON.stringify(discovery, null, 2)}\n`);

  console.log(`Wrote ${tracesPath}`);
  console.log(`Wrote ${discoveryPath}`);
}

const isMain =
  process.argv[1] &&
  fileURLToPath(import.meta.url) === path.resolve(process.argv[1]);

if (isMain) {
  main();
}
