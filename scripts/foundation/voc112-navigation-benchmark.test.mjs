// VOC-112-T04 — validates captured real-agent benchmark and discovery evidence.

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { BENCHMARK_QUESTIONS } from "./voc112-navigation-benchmark-run.mjs";

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

function assertCapturedRevision(evidence) {
  assert.match(evidence.subject_revision, /^[a-f0-9]{40}$/);
  const subjectLookup = spawnSync(
    "git",
    ["cat-file", "-e", `${evidence.subject_revision}^{commit}`],
    {
      cwd: repositoryRoot,
      encoding: "utf8",
    },
  );
  if (subjectLookup.status === 0) {
    execFileSync(
      "git",
      ["merge-base", "--is-ancestor", evidence.subject_revision, "HEAD"],
      { cwd: repositoryRoot },
    );
  } else {
    assert.equal(
      execFileSync("git", ["rev-parse", "--is-shallow-repository"], {
        cwd: repositoryRoot,
        encoding: "utf8",
      }).trim(),
      "true",
      "a missing captured commit is allowed only in a shallow CI checkout",
    );
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
  }
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

test("VOC-112-TEST-12 supplemental: evidence contains no prompts or raw logs", () => {
  const source = readFileSync(
    path.join(
      repositoryRoot,
      "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    ),
    "utf8",
  );
  assert.doesNotMatch(source, /aggregated_output|raw_(?:log|trace)|prompt/i);
  assert.doesNotMatch(source, /(?:API_KEY|AUTH_TOKEN|COOKIE|OAUTH_CODE)/i);
});

test("VOC-112-TEST-13: Cursor and Claude actually read the canonical skill from root and nested cwd", () => {
  const evidence = fixture("voc112-skill-discovery-evidence.json");
  assert.equal(evidence.schema_version, 2);
  assert.equal(evidence.capture_kind, "real-agent-structured-trace");
  assertCapturedRevision(evidence);
  for (const runtime of ["hosted-cursor", "claude-code"]) {
    for (const context of ["repository-root", "nested-cwd-apps-web"]) {
      const row = evidence.discoveries.find(
        (candidate) =>
          candidate.runtime === runtime && candidate.context === context,
      );
      assert.ok(row, `missing ${runtime}/${context} discovery`);
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

test("VOC-112-TEST-14: operator docs and AGENTS.md preserve one-source precedence", () => {
  const doc = readFileSync(
    path.join(repositoryRoot, "docs/development/agent-skills.md"),
    "utf8",
  );
  const agents = readFileSync(path.join(repositoryRoot, "AGENTS.md"), "utf8");
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
});
