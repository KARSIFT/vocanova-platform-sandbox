// VOC-112-T04 — explicit capture of real agent navigation/discovery sessions.
//
// Raw structured runtime output is held in memory only. This script commits only
// sanitized counts, repository paths, runtime identity, usage, and pass/fail data.

import { createHash } from "node:crypto";
import { execFileSync, execSync, spawnSync } from "node:child_process";
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const modulePath = fileURLToPath(import.meta.url);
const repositoryRoot = path.resolve(path.dirname(modulePath), "../..");
const fixturesDirectory = path.join(
  repositoryRoot,
  "scripts/foundation/fixtures",
);
const tracesPath = path.join(
  fixturesDirectory,
  "voc112-navigation-benchmark-traces.json",
);
const discoveryPath = path.join(
  fixturesDirectory,
  "voc112-skill-discovery-evidence.json",
);
const navigatorRelativePath = ".agents/skills/vocanova-repo-navigator/SKILL.md";

export const BENCHMARK_QUESTIONS = [
  {
    id: "nav-q01",
    question: "Where is monitoring configuration documented?",
    expectedPaths: ["infra/monitoring/", "docs/operations/monitoring.md"],
  },
  {
    id: "nav-q02",
    question: "Where are database migrations and schema design documented?",
    expectedPaths: [
      "apps/api/migrations/",
      "docs/engineering/05-database-design.md",
    ],
  },
  {
    id: "nav-q03",
    question: "Where are repository validation commands and foundation tests?",
    expectedPaths: ["docs/development.md", "scripts/foundation/*.test.mjs"],
  },
];

const CODEX_MODEL = "gpt-5.6-sol";
const CURSOR_MODEL = "auto";

function sha256File(relativePath) {
  return createHash("sha256")
    .update(readFileSync(path.join(repositoryRoot, relativePath)))
    .digest("hex");
}

function gitRevision() {
  return execSync("git rev-parse HEAD", {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
}

function commandVersion(command) {
  return execFileSync(command, ["--version"], { encoding: "utf8" }).trim();
}

function parseJsonLines(stdout, runtimeName) {
  try {
    return stdout
      .split(/\r?\n/)
      .filter(Boolean)
      .map((line) => JSON.parse(line));
  } catch {
    throw new Error(`${runtimeName} did not emit valid structured JSON lines`);
  }
}

function normalizePath(value) {
  return value.replaceAll("\\", "/").replace(/^\.\//, "").replace(/\/$/, "");
}

export function pathMatches(actual, expected) {
  const normalizedActual = normalizePath(actual);
  const normalizedExpected = normalizePath(expected);
  if (normalizedExpected.includes("*")) {
    const expression = new RegExp(
      `^${normalizedExpected
        .replace(/[.+?^${}()|[\]\\]/g, "\\$&")
        .replaceAll("*", ".*")}$`,
    );
    return expression.test(normalizedActual);
  }
  return (
    normalizedActual === normalizedExpected ||
    normalizedActual.startsWith(`${normalizedExpected}/`)
  );
}

function extractAnswerJson(text) {
  const trimmed = text.trim();
  try {
    return JSON.parse(trimmed);
  } catch {
    const match = trimmed.match(/\{[\s\S]*\}/);
    if (!match) throw new Error("agent response did not contain a JSON object");
    return JSON.parse(match[0]);
  }
}

function commandRepositoryFiles(command) {
  const candidates =
    command.match(
      /(?:\.agents|\.claude|\.github|apps|docs|infra|scripts|specs)\/[A-Za-z0-9_.*\/:-]+|(?:AGENTS|CLAUDE|README)\.md/g,
    ) ?? [];
  const files = new Set();
  for (const candidate of candidates) {
    const cleaned = candidate.replace(/[,:]+$/, "");
    if (cleaned.includes("*") || cleaned.endsWith("/")) continue;
    try {
      const absolute = path.resolve(repositoryRoot, cleaned);
      const relative = path
        .relative(repositoryRoot, absolute)
        .replaceAll("\\", "/");
      if (!relative.startsWith("../")) {
        readFileSync(absolute);
        files.add(relative);
      }
    } catch {
      // Search roots and nonexistent suggestions are not opened files.
    }
  }
  return [...files];
}

function codexPrompt(question, variant) {
  const posture =
    variant === "navigator_assisted"
      ? "Explicitly use the repository skill vocanova-repo-navigator before answering."
      : "Do not use or read any repository skill; navigate from repository sources only.";
  return [
    "Controlled read-only repository navigation benchmark.",
    posture,
    `Question: ${question.question}`,
    "Inspect only what is necessary. Do not modify files.",
    'Finish with exactly one JSON object and no markdown: {"authoritative_paths":["path"]}',
  ].join(" ");
}

function runCodexSession(question, variant) {
  const started = Date.now();
  const result = spawnSync(
    "codex",
    [
      "exec",
      "-m",
      CODEX_MODEL,
      "-C",
      repositoryRoot,
      "-s",
      "read-only",
      "--ephemeral",
      "--ignore-user-config",
      "--json",
      codexPrompt(question, variant),
    ],
    { encoding: "utf8", maxBuffer: 16 * 1024 * 1024 },
  );
  const elapsedMs = Date.now() - started;
  if (result.status !== 0) {
    throw new Error(
      `codex ${variant}/${question.id} failed with status ${result.status}`,
    );
  }

  const rows = parseJsonLines(result.stdout, "codex");
  const completedCommands = rows
    .filter(
      (row) =>
        row.type === "item.completed" &&
        row.item?.type === "command_execution" &&
        row.item?.status === "completed",
    )
    .map((row) => row.item.command);
  const agentMessage = rows
    .filter(
      (row) =>
        row.type === "item.completed" && row.item?.type === "agent_message",
    )
    .at(-1)?.item?.text;
  if (!agentMessage)
    throw new Error("codex trace has no completed agent message");
  const answer = extractAnswerJson(agentMessage);
  const answerPaths = Array.isArray(answer.authoritative_paths)
    ? answer.authoritative_paths.map(String)
    : [];
  const usage = rows.findLast((row) => row.type === "turn.completed")?.usage;
  if (!usage) throw new Error("codex trace has no usage record");

  const opened = [
    ...new Set(completedCommands.flatMap(commandRepositoryFiles)),
  ].sort();
  const skillFiles = opened.filter((entry) =>
    entry.startsWith(".agents/skills/"),
  );
  const repositoryFiles = opened.filter(
    (entry) => !entry.startsWith(".agents/skills/"),
  );
  const correct = question.expectedPaths.every((expected) =>
    answerPaths.some((actual) => pathMatches(actual, expected)),
  );

  return {
    question_id: question.id,
    variant,
    elapsed_ms: elapsedMs,
    correct,
    answer_paths: answerPaths,
    repository_files_opened: repositoryFiles,
    skill_files_opened: skillFiles,
    search_operations: completedCommands.filter((command) =>
      /\b(?:rg|grep|find)\b/.test(command),
    ).length,
    tool_calls: completedCommands.length,
    usage: {
      input_tokens: usage.input_tokens,
      cached_input_tokens: usage.cached_input_tokens,
      output_tokens: usage.output_tokens,
      reasoning_output_tokens: usage.reasoning_output_tokens,
    },
  };
}

function captureCodexBenchmark() {
  const sessions = [];
  for (const [index, question] of BENCHMARK_QUESTIONS.entries()) {
    const order =
      index % 2 === 0
        ? ["baseline", "navigator_assisted"]
        : ["navigator_assisted", "baseline"];
    for (const variant of order) {
      sessions.push(runCodexSession(question, variant));
    }
  }

  const evidence = {
    schema_version: 2,
    capture_kind: "real-agent-structured-trace",
    subject_revision: gitRevision(),
    recorded_at: new Date().toISOString(),
    source_hashes: {
      navigator_skill_sha256: sha256File(navigatorRelativePath),
      agents_sha256: sha256File("AGENTS.md"),
    },
    runtime: {
      name: "codex",
      version: commandVersion("codex"),
      model: CODEX_MODEL,
      sandbox: "read-only",
      session: "ephemeral",
      user_config: "ignored",
      file_observation:
        "readable-repository-paths-referenced-in-completed-command-events",
    },
    thresholds: {
      max_repository_files_delta: 0,
      max_search_operations_delta: 0,
      max_tool_calls_delta: BENCHMARK_QUESTIONS.length,
      require_correctness_non_regression: true,
      require_navigator_all_correct: true,
    },
    questions: BENCHMARK_QUESTIONS.map(({ id, expectedPaths }) => ({
      id,
      expected_paths: expectedPaths,
    })),
    sessions,
  };
  writeFileSync(tracesPath, `${JSON.stringify(evidence, null, 2)}\n`);
  console.log(`captured ${sessions.length} sanitized Codex sessions`);
}

function cursorCompletedReadPaths(events, cwd) {
  const reads = [];
  for (const event of events) {
    if (event.type !== "tool_call" || event.subtype !== "completed") continue;
    const toolCall = event.tool_call?.readToolCall;
    const requestedPath = toolCall?.args?.path;
    if (typeof requestedPath !== "string" || !toolCall.result?.success)
      continue;
    const absolutePath = path.isAbsolute(requestedPath)
      ? requestedPath
      : path.resolve(cwd, requestedPath);
    reads.push(
      path.relative(repositoryRoot, absolutePath).replaceAll("\\", "/"),
    );
  }
  return reads;
}

function discoveryPrompt() {
  return "Invoke the repository skill named vocanova-repo-navigator. Read its canonical skill file, identify the monitoring documentation, and finish with exactly DISCOVERY_OK.";
}

function readDiscoveryEvidence() {
  try {
    const evidence = JSON.parse(readFileSync(discoveryPath, "utf8"));
    evidence.discoveries = (evidence.discoveries ?? [])
      .filter(
        (row) =>
          typeof row.runtime === "string" &&
          typeof row.context === "string" &&
          typeof row.result === "string" &&
          Array.isArray(row.canonical_skill_reads) &&
          Number.isInteger(row.structured_event_count),
      )
      .map((row) => ({
        ...row,
        captured_at: row.captured_at ?? evidence.recorded_at,
        subject_revision: row.subject_revision ?? evidence.subject_revision,
        source_hashes: row.source_hashes ?? evidence.source_hashes,
      }));
    return evidence;
  } catch {
    return {
      schema_version: 3,
      discoveries: [],
    };
  }
}

export function mergeDiscoveryRows(
  evidence,
  runtimeName,
  runtimeVersion,
  rows,
  capture,
) {
  return {
    schema_version: 3,
    capture_kind: "real-agent-structured-trace",
    recorded_at: capture.captured_at,
    discoveries: [
      ...(evidence.discoveries ?? []).filter(
        (row) => row.runtime !== runtimeName,
      ),
      ...rows.map((row) => ({
        ...row,
        runtime: runtimeName,
        runtime_version: runtimeVersion,
        captured_at: capture.captured_at,
        subject_revision: capture.subject_revision,
        source_hashes: capture.source_hashes,
      })),
    ],
  };
}

function writeDiscoveryRows(runtimeName, runtimeVersion, rows) {
  const evidence = mergeDiscoveryRows(
    readDiscoveryEvidence(),
    runtimeName,
    runtimeVersion,
    rows,
    {
      captured_at: new Date().toISOString(),
      subject_revision: gitRevision(),
      source_hashes: {
        navigator_skill_sha256: sha256File(navigatorRelativePath),
        agents_sha256: sha256File("AGENTS.md"),
      },
    },
  );
  writeFileSync(discoveryPath, `${JSON.stringify(evidence, null, 2)}\n`);
}

export function classifyClaudeFailure(result) {
  const diagnostic = `${result.stderr ?? ""}\n${result.stdout ?? ""}`;
  return /(?:not logged in|please log in|authentication required|unauthorized|missing[^\n]*(?:credential|api key)|invalid[^\n]*api key)/i.test(
    diagnostic,
  )
    ? "not-executed-external-credential-required"
    : "fail";
}

function captureClaudeDiscovery() {
  const version = commandVersion("claude");
  const discoveries = [
    ["repository-root", repositoryRoot],
    ["nested-cwd-apps-web", path.join(repositoryRoot, "apps/web")],
  ].map(([context, cwd]) => {
    const result = spawnSync(
      "claude",
      [
        "-p",
        "--output-format",
        "stream-json",
        "--verbose",
        "--permission-mode",
        "plan",
        "--no-session-persistence",
        discoveryPrompt(),
      ],
      { cwd, encoding: "utf8", maxBuffer: 16 * 1024 * 1024 },
    );
    if (result.status !== 0) {
      return {
        context,
        cwd: path.relative(repositoryRoot, cwd).replaceAll("\\", "/") || ".",
        result: classifyClaudeFailure(result),
        model: "unavailable",
        structured_event_count: 0,
        canonical_skill_reads: [],
      };
    }
    const events = parseJsonLines(result.stdout, "claude-code");
    const skillReads = [];
    let model = "unknown";
    for (const event of events) {
      if (event.type === "system" && event.subtype === "init") {
        model = event.model ?? model;
      }
      for (const content of event.message?.content ?? []) {
        if (content.type === "tool_use" && content.name === "Read") {
          const relative = path
            .relative(repositoryRoot, content.input?.file_path ?? "")
            .replaceAll("\\", "/");
          if (relative === navigatorRelativePath) skillReads.push(relative);
        }
      }
    }
    const succeeded = events.some(
      (event) =>
        event.type === "result" &&
        String(event.result).includes("DISCOVERY_OK"),
    );
    return {
      context,
      cwd: path.relative(repositoryRoot, cwd).replaceAll("\\", "/") || ".",
      result: succeeded && skillReads.length > 0 ? "pass" : "fail",
      model,
      structured_event_count: events.length,
      canonical_skill_reads: [...new Set(skillReads)],
    };
  });
  writeDiscoveryRows("claude-code", version, discoveries);
  console.log("captured sanitized Claude Code root/nested discovery sessions");
}

function captureCursorDiscovery() {
  if (!process.env.CURSOR_API_KEY && !process.env.CURSOR_AUTH_TOKEN) {
    throw new Error(
      "authorized hosted Cursor credential is required for explicit capture",
    );
  }
  const version = commandVersion("cursor-agent");
  const discoveries = [
    ["repository-root", repositoryRoot],
    ["nested-cwd-apps-web", path.join(repositoryRoot, "apps/web")],
  ].map(([context, cwd]) => {
    const result = spawnSync(
      "cursor-agent",
      [
        "-p",
        "--output-format",
        "stream-json",
        "--mode",
        "ask",
        "--trust",
        "--model",
        CURSOR_MODEL,
        discoveryPrompt(),
      ],
      { cwd, encoding: "utf8", maxBuffer: 16 * 1024 * 1024 },
    );
    if (result.status !== 0) {
      throw new Error(`Cursor discovery failed in ${context}`);
    }
    const events = parseJsonLines(result.stdout, "hosted-cursor");
    const completedReadPaths = cursorCompletedReadPaths(events, cwd);
    const skillReads = completedReadPaths.filter(
      (readPath) => readPath === navigatorRelativePath,
    );
    const initialization = events.find(
      (event) => event.type === "system" && event.subtype === "init",
    );
    const initializedInRequestedCwd =
      path.resolve(initialization?.cwd ?? "") === cwd;
    const succeeded = events.some(
      (event) =>
        event.type === "result" &&
        event.subtype === "success" &&
        String(event.result).includes("DISCOVERY_OK"),
    );
    return {
      context,
      cwd: path.relative(repositoryRoot, cwd).replaceAll("\\", "/") || ".",
      result:
        succeeded && initializedInRequestedCwd && skillReads.length > 0
          ? "pass"
          : "fail",
      model: initialization?.model ?? CURSOR_MODEL,
      permission_mode: initialization?.permissionMode ?? "ask",
      structured_event_count: events.length,
      completed_read_tool_call_count: completedReadPaths.length,
      canonical_skill_reads: [...new Set(skillReads)],
    };
  });
  writeDiscoveryRows("hosted-cursor", version, discoveries);
  console.log(
    "captured sanitized hosted Cursor root/nested discovery sessions",
  );
}

function runCli(action) {
  if (action === "--capture-codex") captureCodexBenchmark();
  else if (action === "--capture-claude-discovery") captureClaudeDiscovery();
  else if (action === "--capture-cursor-discovery") captureCursorDiscovery();
  else if (action) throw new Error(`unknown action: ${action}`);
}

// Importing BENCHMARK_QUESTIONS from a test or another tool must never start a
// credentialed capture merely because that parent process has extra argv data.
if (process.argv[1] && path.resolve(process.argv[1]) === modulePath) {
  runCli(process.argv[2]);
}
