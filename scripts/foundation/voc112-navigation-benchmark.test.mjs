// VOC-112-T04 — navigation benchmark evidence validation and discovery proof checks.
//
// Runs via `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`.
// Regenerate traces: `node scripts/foundation/voc112-navigation-benchmark-run.mjs`

import assert from "node:assert/strict";
import { execFileSync, execSync } from "node:child_process";
import { readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import {
  BENCHMARK_QUESTIONS,
} from "./voc112-navigation-benchmark-run.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const runnerPath = path.join(
  repositoryRoot,
  "scripts/foundation/voc112-navigation-benchmark-run.mjs",
);

test.before(() => {
  execFileSync("node", [runnerPath], {
    cwd: repositoryRoot,
    stdio: "pipe",
  });
});

const tracesPath = path.join(
  repositoryRoot,
  "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
);
const discoveryPath = path.join(
  repositoryRoot,
  "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
);
const agentSkillsDocPath = path.join(
  repositoryRoot,
  "docs/development/agent-skills.md",
);
const agentsPath = path.join(repositoryRoot, "AGENTS.md");

function readJson(filePath) {
  return JSON.parse(readFileSync(filePath, "utf8"));
}

function gitRevision() {
  return execSync("git rev-parse HEAD", {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
}

function aggregateMetrics(session) {
  return {
    totalFiles: session.reduce((sum, row) => sum + row.files_opened.length, 0),
    totalSearches: session.reduce(
      (sum, row) => sum + row.search_operations,
      0,
    ),
    totalElapsedMs: session.reduce((sum, row) => sum + row.elapsed_ms, 0),
    correctCount: session.filter((row) => row.correct).length,
  };
}

function evaluateThresholds(traces) {
  const baseline = aggregateMetrics(traces.sessions.baseline);
  const navigator = aggregateMetrics(traces.sessions.navigator_assisted);
  const regressions = {
    files: navigator.totalFiles - baseline.totalFiles,
    searches: navigator.totalSearches - baseline.totalSearches,
    elapsed_ms: navigator.totalElapsedMs - baseline.totalElapsedMs,
    correctness: navigator.correctCount - baseline.correctCount,
  };

  const thresholds = traces.thresholds;
  const pass =
    regressions.files <= thresholds.max_regression_files &&
    regressions.searches <= thresholds.max_regression_searches &&
    regressions.elapsed_ms <= thresholds.max_regression_time_ms &&
  (thresholds.require_correctness_non_regression
      ? regressions.correctness >= 0
      : true);

  return { baseline, navigator, regressions, pass };
}

test("VOC-112-TEST-12: benchmark traces exist with revision-bound rubric coverage", () => {
  assert.ok(statSync(tracesPath, { throwIfNoEntry: false }), "missing traces fixture");
  const traces = readJson(tracesPath);
  const revision = gitRevision();

  assert.equal(traces.schema_version, 1);
  assert.equal(traces.revision, revision, "traces must bind to current HEAD");
  assert.equal(traces.rubric_version, "voc112-d05");
  assert.ok(Array.isArray(traces.sessions.baseline));
  assert.ok(Array.isArray(traces.sessions.navigator_assisted));

  const expectedIds = BENCHMARK_QUESTIONS.map((q) => q.id).sort();
  const baselineIds = traces.sessions.baseline.map((row) => row.question_id).sort();
  const navigatorIds = traces.sessions.navigator_assisted
    .map((row) => row.question_id)
    .sort();

  assert.deepEqual(baselineIds, expectedIds);
  assert.deepEqual(navigatorIds, expectedIds);

  for (const question of BENCHMARK_QUESTIONS) {
    const baselineRow = traces.sessions.baseline.find(
      (row) => row.question_id === question.id,
    );
    const navigatorRow = traces.sessions.navigator_assisted.find(
      (row) => row.question_id === question.id,
    );
    assert.ok(baselineRow, `missing baseline row for ${question.id}`);
    assert.ok(navigatorRow, `missing navigator row for ${question.id}`);
    assert.deepEqual(
      baselineRow.expected_authoritative_paths,
      navigatorRow.expected_authoritative_paths,
    );
    assert.equal(
      navigatorRow.intent_label,
      question.intentLabel,
      `intent label drift for ${question.id}`,
    );
  }
});

test("VOC-112-TEST-12: navigator-assisted path improves or does not regress cost and correctness", () => {
  const traces = readJson(tracesPath);
  const evaluation = evaluateThresholds(traces);

  assert.ok(
    evaluation.pass,
    `navigator regression detected: files=${evaluation.regressions.files}, searches=${evaluation.regressions.searches}, elapsed_ms=${evaluation.regressions.elapsed_ms}, correctness=${evaluation.regressions.correctness}`,
  );

  for (const row of traces.sessions.navigator_assisted) {
    assert.equal(row.correct, true, `navigator row ${row.question_id} must be correct`);
    assert.ok(
      row.files_opened.length <=
        traces.sessions.baseline.find((b) => b.question_id === row.question_id)
          .files_opened.length,
      `navigator must not open more files than baseline for ${row.question_id}`,
    );
    assert.ok(
      row.search_operations <=
        traces.sessions.baseline.find((b) => b.question_id === row.question_id)
          .search_operations,
      `navigator must not exceed baseline searches for ${row.question_id}`,
    );
  }

  assert.ok(
    evaluation.navigator.totalSearches < evaluation.baseline.totalSearches,
    "navigator-assisted sessions must reduce total search operations versus baseline",
  );
});

test("VOC-112-TEST-12 supplemental: threshold math rejects fabricated hardcoded success", () => {
  const synthetic = {
    thresholds: {
      max_regression_files: 0,
      max_regression_searches: 0,
      max_regression_time_ms: 0,
      require_correctness_non_regression: true,
    },
    sessions: {
      baseline: [
        {
          question_id: "synthetic-q01",
          files_opened: ["apps/web/"],
          search_operations: 2,
          elapsed_ms: 10,
          correct: true,
        },
      ],
      navigator_assisted: [
        {
          question_id: "synthetic-q01",
          files_opened: ["apps/web/", "apps/api/", "infra/"],
          search_operations: 2,
          elapsed_ms: 10,
          correct: true,
        },
      ],
    },
  };
  const evaluation = evaluateThresholds(synthetic);
  assert.equal(
    evaluation.pass,
    false,
    "inflated navigator files must fail zero-regression thresholds",
  );
});

test("VOC-112-TEST-13: runtime discovery evidence is revision-bound and truthful", () => {
  assert.ok(
    statSync(discoveryPath, { throwIfNoEntry: false }),
    "missing discovery evidence fixture",
  );
  const discovery = readJson(discoveryPath);
  const revision = gitRevision();

  assert.equal(discovery.schema_version, 1);
  assert.equal(discovery.revision, revision);
  assert.ok(Array.isArray(discovery.discoveries));

  const cursorRoot = discovery.discoveries.find(
    (row) =>
      row.runtime === "hosted-cursor" && row.context === "repository-root",
  );
  const cursorNested = discovery.discoveries.find(
    (row) =>
      row.runtime === "hosted-cursor" &&
      row.context === "nested-cwd-apps-web",
  );
  assert.ok(cursorRoot, "hosted Cursor root discovery row required");
  assert.ok(cursorNested, "hosted Cursor nested discovery row required");
  assert.equal(cursorRoot.result, "pass");
  assert.equal(cursorNested.result, "pass");
  assert.ok(cursorRoot.canonical_skill_count >= 9);

  for (const row of discovery.discoveries.filter(
    (entry) => entry.runtime === "claude-code",
  )) {
    assert.ok(
      row.result === "pass" ||
        row.result === "not-executed-external-credential-required",
      "Claude discovery must pass or record external credential limitation",
    );
  }
});

test("VOC-112-TEST-14: operator documentation and AGENTS.md pointer preserve precedence", () => {
  const doc = readFileSync(agentSkillsDocPath, "utf8");
  const agents = readFileSync(agentsPath, "utf8");

  assert.match(doc, /## Installation scope/);
  assert.match(doc, /## One-source architecture/);
  assert.match(doc, /## Updating pinned upstream material/);
  assert.match(doc, /## Graphify pilot limitations/);
  assert.match(doc, /## Safe use/);
  assert.doesNotMatch(doc, /Completed in T04/);
  assert.doesNotMatch(doc, /Skeleton \(VOC-112-T00\)/);

  assert.match(agents, /## Agent skills/);
  assert.match(agents, /docs\/development\/agent-skills\.md/);
  assert.match(
    agents,
    /repository sources win|governance precedence/i,
    "AGENTS.md pointer must preserve governance precedence",
  );
  assert.match(agents, /A-004 is the effective/);
  assert.match(agents, /Never self-approve/);
});
