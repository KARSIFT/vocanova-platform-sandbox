// VOC-112-T04 — validates captured real-agent benchmark and discovery evidence.

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import test from "node:test";
import {
  BENCHMARK_QUESTIONS,
  classifyClaudeFailure,
  mergeDiscoveryRows,
  pathMatches,
} from "./voc112-navigation-benchmark-run.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const fixture = (name) =>
  JSON.parse(
    readFileSync(
      path.join(repositoryRoot, "scripts/foundation/fixtures", name),
      "utf8",
    ),
  );
const sha256 = (relativePath) =>
  createHash("sha256")
    .update(readFileSync(path.join(repositoryRoot, relativePath)))
    .digest("hex");
const sha256AtRevision = (revision, relativePath) =>
  createHash("sha256")
    .update(
      execFileSync("git", ["show", `${revision}:${relativePath}`], {
        cwd: repositoryRoot,
      }),
    )
    .digest("hex");

function assertCapturedRevision(evidence) {
  assert.match(evidence.subject_revision, /^[a-f0-9]{40}$/);
  const mode = process.env.VOC112_CAPTURE_PROVENANCE_MODE ?? "local";
  assert.ok(
    ["local", "pr-ancestry", "pr-validation", "squash-safe-push"].includes(mode),
    `unknown capture provenance mode: ${mode}`,
  );
  const subjectLookup = spawnSync(
    "git",
    ["cat-file", "-e", `${evidence.subject_revision}^{commit}`],
    {
      cwd: repositoryRoot,
      encoding: "utf8",
    },
  );
  if (subjectLookup.status !== 0) {
    assert.notEqual(
      mode,
      "pr-ancestry",
      "PR ancestry mode requires every captured commit object",
    );
    if (mode === "local") {
      assert.equal(
        execFileSync("git", ["rev-parse", "--is-shallow-repository"], {
          cwd: repositoryRoot,
          encoding: "utf8",
        }).trim(),
        "true",
        "a full local checkout must already contain the captured commit",
      );
    }
    if (mode === "pr-validation") {
      const prBaseSha = process.env.PR_BASE_SHA;
      const prHeadSha = process.env.PR_HEAD_SHA;
      assert.match(
        prBaseSha ?? "",
        /^[a-f0-9]{40}$/,
        "post-squash PR validation requires PR_BASE_SHA",
      );
      assert.match(
        prHeadSha ?? "",
        /^[a-f0-9]{40}$/,
        "post-squash PR validation requires PR_HEAD_SHA",
      );
      const mergeBase = execFileSync(
        "git",
        ["merge-base", prBaseSha, prHeadSha],
        { cwd: repositoryRoot, encoding: "utf8" },
      ).trim();
      assert.match(mergeBase, /^[a-f0-9]{40}$/);
      assert.equal(
        evidence.source_hashes.navigator_skill_sha256,
        sha256AtRevision(
          mergeBase,
          ".agents/skills/vocanova-repo-navigator/SKILL.md",
        ),
        "navigator hash must be anchored in the PR merge base",
      );
      assert.equal(
        evidence.source_hashes.agents_sha256,
        sha256AtRevision(mergeBase, "AGENTS.md"),
        "AGENTS.md hash must be anchored in the PR merge base",
      );
    }
  } else {
    const ancestry = spawnSync(
      "git",
      ["merge-base", "--is-ancestor", evidence.subject_revision, "HEAD"],
      { cwd: repositoryRoot },
    );
    const ancestryProven = ancestry.status === 0;
    if (!ancestryProven && mode !== "squash-safe-push") {
      const prBaseSha = process.env.PR_BASE_SHA;
      const prHeadSha = process.env.PR_HEAD_SHA;
      assert.match(
        prBaseSha ?? "",
        /^[a-f0-9]{40}$/,
        "post-squash PR validation requires PR_BASE_SHA",
      );
      assert.match(
        prHeadSha ?? "",
        /^[a-f0-9]{40}$/,
        "post-squash PR validation requires PR_HEAD_SHA",
      );
      const mergeBase = execFileSync(
        "git",
        ["merge-base", prBaseSha, prHeadSha],
        { cwd: repositoryRoot, encoding: "utf8" },
      ).trim();
      assert.match(mergeBase, /^[a-f0-9]{40}$/);
      assert.equal(
        evidence.source_hashes.navigator_skill_sha256,
        sha256AtRevision(
          mergeBase,
          ".agents/skills/vocanova-repo-navigator/SKILL.md",
        ),
        "navigator hash must be anchored in the PR merge base",
      );
      assert.equal(
        evidence.source_hashes.agents_sha256,
        sha256AtRevision(mergeBase, "AGENTS.md"),
        "AGENTS.md hash must be anchored in the PR merge base",
      );
    }
    if (ancestryProven && mode === "pr-ancestry") {
      // Original capture PR path: ancestry plus captured-revision binding below.
    }
    if (ancestryProven && mode !== "squash-safe-push") {
      assert.equal(
        evidence.source_hashes.navigator_skill_sha256,
        sha256AtRevision(
          evidence.subject_revision,
          ".agents/skills/vocanova-repo-navigator/SKILL.md",
        ),
        "navigator hash must bind to the captured revision",
      );
      assert.equal(
        evidence.source_hashes.agents_sha256,
        sha256AtRevision(evidence.subject_revision, "AGENTS.md"),
        "AGENTS.md hash must bind to the captured revision",
      );
    }
  }
  assert.equal(
    evidence.source_hashes.navigator_skill_sha256,
    sha256(".agents/skills/vocanova-repo-navigator/SKILL.md"),
  );
  assert.equal(evidence.source_hashes.agents_sha256, sha256("AGENTS.md"));
}

function totals(rows) {
  return {
    repositoryFiles: rows.reduce(
      (sum, row) => sum + row.repository_files_opened.length,
      0,
    ),
    searches: rows.reduce((sum, row) => sum + row.search_operations, 0),
    toolCalls: rows.reduce((sum, row) => sum + row.tool_calls, 0),
    elapsedMs: rows.reduce((sum, row) => sum + row.elapsed_ms, 0),
    correct: rows.filter((row) => row.correct).length,
    inputTokens: rows.reduce((sum, row) => sum + row.usage.input_tokens, 0),
    outputTokens: rows.reduce((sum, row) => sum + row.usage.output_tokens, 0),
  };
}

test("VOC-112-TEST-12: benchmark is a revision-bound real structured agent capture", () => {
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  assert.equal(evidence.schema_version, 2);
  assert.equal(evidence.capture_kind, "real-agent-structured-trace");
  assertCapturedRevision(evidence);
  assert.equal(evidence.runtime.name, "codex");
  assert.equal(evidence.runtime.model, "gpt-5.6-sol");
  assert.equal(evidence.runtime.sandbox, "read-only");
  assert.equal(evidence.runtime.session, "ephemeral");
  assert.equal(evidence.runtime.user_config, "ignored");
  assert.equal(
    evidence.runtime.file_observation,
    "readable-repository-paths-referenced-in-completed-command-events",
  );

  const expectedIds = BENCHMARK_QUESTIONS.map((question) => question.id).sort();
  for (const variant of ["baseline", "navigator_assisted"]) {
    assert.deepEqual(
      evidence.sessions
        .filter((row) => row.variant === variant)
        .map((row) => row.question_id)
        .sort(),
      expectedIds,
    );
  }
  for (const row of evidence.sessions) {
    assert.ok(row.elapsed_ms > 0);
    assert.ok(row.usage.input_tokens > 0);
    assert.ok(row.usage.output_tokens > 0);
    assert.ok(Array.isArray(row.repository_files_opened));
    assert.ok(Array.isArray(row.answer_paths));
    const question = BENCHMARK_QUESTIONS.find(
      (candidate) => candidate.id === row.question_id,
    );
    assert.ok(question, `unknown benchmark question: ${row.question_id}`);
    assert.equal(
      row.correct,
      question.expectedPaths.every((expected) =>
        row.answer_paths.some((actual) => pathMatches(actual, expected)),
      ),
      `stored correctness does not match the keyed rubric for ${row.question_id}/${row.variant}`,
    );
    if (row.variant === "navigator_assisted") {
      assert.deepEqual(row.skill_files_opened, [
        ".agents/skills/vocanova-repo-navigator/SKILL.md",
      ]);
    } else {
      assert.deepEqual(row.skill_files_opened, []);
    }
  }
});

test("VOC-112-TEST-12: importing the capture module has no CLI side effects", () => {
  const runnerUrl = pathToFileURL(
    path.join(
      repositoryRoot,
      "scripts/foundation/voc112-navigation-benchmark-run.mjs",
    ),
  ).href;
  const result = spawnSync(
    process.execPath,
    [
      "--input-type=module",
      "--eval",
      `process.argv[2] = "--capture-codex"; await import(${JSON.stringify(runnerUrl)});`,
    ],
    { cwd: repositoryRoot, encoding: "utf8" },
  );
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, "");
});

test("VOC-112-TEST-12: navigator does not regress keyed correctness or bounded cost", () => {
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  const baseline = totals(
    evidence.sessions.filter((row) => row.variant === "baseline"),
  );
  const navigator = totals(
    evidence.sessions.filter((row) => row.variant === "navigator_assisted"),
  );
  const thresholds = evidence.thresholds;
  assert.ok(navigator.correct >= baseline.correct);
  assert.equal(navigator.correct, BENCHMARK_QUESTIONS.length);
  assert.ok(
    navigator.repositoryFiles - baseline.repositoryFiles <=
      thresholds.max_repository_files_delta,
  );
  assert.ok(
    navigator.searches - baseline.searches <=
      thresholds.max_search_operations_delta,
  );
  assert.ok(
    navigator.toolCalls - baseline.toolCalls <= thresholds.max_tool_calls_delta,
  );
});

test("VOC-112-TEST-12 supplemental: evidence fixtures contain no prompts or raw logs", () => {
  for (const name of [
    "voc112-navigation-benchmark-traces.json",
    "voc112-skill-discovery-evidence.json",
  ]) {
    const source = readFileSync(
      path.join(repositoryRoot, "scripts/foundation/fixtures", name),
      "utf8",
    );
    assert.doesNotMatch(source, /aggregated_output|raw_(?:log|trace)|prompt/i);
    assert.doesNotMatch(source, /(?:API_KEY|AUTH_TOKEN|COOKIE|OAUTH_CODE)/i);
  }
});

test("VOC-112-TEST-13: Cursor and Claude actually read the canonical skill from root and nested cwd", () => {
  const evidence = fixture("voc112-skill-discovery-evidence.json");
  assert.equal(evidence.schema_version, 3);
  assert.equal(evidence.capture_kind, "real-agent-structured-trace");
  for (const runtime of ["hosted-cursor", "claude-code"]) {
    for (const context of ["repository-root", "nested-cwd-apps-web"]) {
      const row = evidence.discoveries.find(
        (candidate) =>
          candidate.runtime === runtime && candidate.context === context,
      );
      assert.ok(row, `missing ${runtime}/${context} discovery`);
      assertCapturedRevision(row);
      assert.equal(row.result, "pass", `${runtime}/${context} did not pass`);
      assert.deepEqual(row.canonical_skill_reads, [
        ".agents/skills/vocanova-repo-navigator/SKILL.md",
      ]);
      assert.ok(row.structured_event_count > 0);
      if (runtime === "hosted-cursor") {
        assert.ok(row.completed_read_tool_call_count > 0);
      }
      assert.ok(row.runtime_version);
    }
  }
});

test("VOC-112-TEST-13: later runtime capture preserves truthful credential limitations", () => {
  const sourceHashes = {
    navigator_skill_sha256: sha256(
      ".agents/skills/vocanova-repo-navigator/SKILL.md",
    ),
    agents_sha256: sha256("AGENTS.md"),
  };
  const limitedClaudeRow = {
    runtime: "claude-code",
    runtime_version: "fixture",
    context: "repository-root",
    cwd: ".",
    result: "not-executed-external-credential-required",
    model: "unavailable",
    structured_event_count: 0,
    canonical_skill_reads: [],
    captured_at: "2026-08-24T00:00:00.000Z",
    subject_revision: "a".repeat(40),
    source_hashes: sourceHashes,
  };
  const merged = mergeDiscoveryRows(
    { discoveries: [limitedClaudeRow] },
    "hosted-cursor",
    "fixture",
    [
      {
        context: "repository-root",
        cwd: ".",
        result: "pass",
        structured_event_count: 1,
        canonical_skill_reads: [
          ".agents/skills/vocanova-repo-navigator/SKILL.md",
        ],
      },
    ],
    {
      captured_at: "2026-08-24T00:01:00.000Z",
      subject_revision: "b".repeat(40),
      source_hashes: sourceHashes,
    },
  );
  assert.deepEqual(
    merged.discoveries.find((row) => row.runtime === "claude-code"),
    limitedClaudeRow,
  );
});

test("VOC-112-TEST-13: Claude failures are not all mislabeled as credential limitations", () => {
  assert.equal(
    classifyClaudeFailure({ stderr: "Authentication required: please log in" }),
    "not-executed-external-credential-required",
  );
  assert.equal(
    classifyClaudeFailure({ stderr: "process crashed while parsing output" }),
    "fail",
  );
});

test("VOC-112-TEST-14: operator docs and AGENTS.md preserve one-source precedence", () => {
  const doc = readFileSync(
    path.join(repositoryRoot, "docs/development/agent-skills.md"),
    "utf8",
  );
  const agents = readFileSync(path.join(repositoryRoot, "AGENTS.md"), "utf8");
  const governanceWorkflow = readFileSync(
    path.join(repositoryRoot, ".github/workflows/repository-governance.yml"),
    "utf8",
  );
  for (const heading of [
    "## Installation scope",
    "## One-source architecture",
    "## Updating pinned upstream material",
    "## Graphify pilot limitations",
    "## Safe use",
  ]) {
    assert.match(doc, new RegExp(heading));
  }
  assert.match(doc, /\.agents\/skills\/<skill-name>\/SKILL\.md/);
  assert.match(doc, /\.claude\/skills\/<skill-name>\/SKILL\.md/);
  assert.match(doc, /loader-only/);
  assert.match(agents, /## Agent skills/);
  assert.match(agents, /docs\/development\/agent-skills\.md/);
  assert.match(agents, /repository sources win/i);
  assert.match(agents, /A-004 is the effective/);
  assert.match(agents, /Never self-approve/);
  assert.match(governanceWorkflow, /fetch-depth: 0/);
  assert.match(governanceWorkflow, /VOC112_CAPTURE_PROVENANCE_MODE:/);
  assert.match(governanceWorkflow, /pr-validation/);
  assert.match(governanceWorkflow, /squash-safe-push/);
  assert.match(governanceWorkflow, /PR_BASE_SHA:/);
  assert.match(governanceWorkflow, /PR_HEAD_SHA:/);
  assert.match(
    governanceWorkflow,
    /node --test scripts\/foundation\/voc112-navigation-benchmark\.test\.mjs/,
  );
});

test("VOC-113-TEST-10: tampered merge base fails closed under pr-validation", () => {
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  const headSha = execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
  let wrongBase = "";
  for (let depth = 1; depth < 50; depth += 1) {
    const candidate = execFileSync("git", ["rev-parse", `HEAD~${depth}`], {
      cwd: repositoryRoot,
      encoding: "utf8",
    }).trim();
    if (
      sha256AtRevision(candidate, "AGENTS.md") !==
      evidence.source_hashes.agents_sha256
    ) {
      wrongBase = candidate;
      break;
    }
  }
  assert.ok(
    wrongBase,
    "need a parent commit with a different agents hash for tampered fixture",
  );
  const priorMode = process.env.VOC112_CAPTURE_PROVENANCE_MODE;
  const priorBase = process.env.PR_BASE_SHA;
  const priorHead = process.env.PR_HEAD_SHA;
  process.env.VOC112_CAPTURE_PROVENANCE_MODE = "pr-validation";
  process.env.PR_BASE_SHA = wrongBase;
  process.env.PR_HEAD_SHA = headSha;
  try {
    assert.throws(
      () => assertCapturedRevision(evidence),
      (error) =>
        error instanceof assert.AssertionError &&
        /merge base|anchored/.test(error.message),
    );
  } finally {
    if (priorMode === undefined) {
      delete process.env.VOC112_CAPTURE_PROVENANCE_MODE;
    } else {
      process.env.VOC112_CAPTURE_PROVENANCE_MODE = priorMode;
    }
    if (priorBase === undefined) {
      delete process.env.PR_BASE_SHA;
    } else {
      process.env.PR_BASE_SHA = priorBase;
    }
    if (priorHead === undefined) {
      delete process.env.PR_HEAD_SHA;
    } else {
      process.env.PR_HEAD_SHA = priorHead;
    }
  }
});

test("VOC-113-TEST-10: changed current hash fails closed under pr-validation", () => {
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    source_hashes: {
      navigator_skill_sha256: "f".repeat(64),
      agents_sha256: sha256("AGENTS.md"),
    },
  };
  const headSha = execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
  const baseSha = execFileSync("git", ["rev-parse", "HEAD~1"], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
  const priorMode = process.env.VOC112_CAPTURE_PROVENANCE_MODE;
  const priorBase = process.env.PR_BASE_SHA;
  const priorHead = process.env.PR_HEAD_SHA;
  process.env.VOC112_CAPTURE_PROVENANCE_MODE = "pr-validation";
  process.env.PR_BASE_SHA = baseSha;
  process.env.PR_HEAD_SHA = headSha;
  try {
    assert.throws(() => assertCapturedRevision(evidence));
  } finally {
    if (priorMode === undefined) {
      delete process.env.VOC112_CAPTURE_PROVENANCE_MODE;
    } else {
      process.env.VOC112_CAPTURE_PROVENANCE_MODE = priorMode;
    }
    if (priorBase === undefined) {
      delete process.env.PR_BASE_SHA;
    } else {
      process.env.PR_BASE_SHA = priorBase;
    }
    if (priorHead === undefined) {
      delete process.env.PR_HEAD_SHA;
    } else {
      process.env.PR_HEAD_SHA = priorHead;
    }
  }
});
