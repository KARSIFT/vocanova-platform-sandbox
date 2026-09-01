// VOC-112-T04 — validates captured real-agent benchmark and discovery evidence.

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
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
const VOC112_FIXTURES = {
  benchmark:
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
  discovery: "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
};
const VOC112_PINNED_ANCHOR = "587269f547c93a899ca7b5504825ab5304d7a266";
const VOC112_FIXTURE_PINNED_ANCHORS = {
  [VOC112_FIXTURES.benchmark]: VOC112_PINNED_ANCHOR,
  [VOC112_FIXTURES.discovery]: VOC112_PINNED_ANCHOR,
};
const NAVIGATOR_SKILL_PATH = ".agents/skills/vocanova-repo-navigator/SKILL.md";
const AGENTS_PATH = "AGENTS.md";
const SHA256_PATTERN = /^[a-f0-9]{64}$/;
const COMMIT_PATTERN = /^[a-f0-9]{40}$/;
const GIT_FIXTURE_ENV = {
  GIT_AUTHOR_NAME: "VocaNova Fixture",
  GIT_AUTHOR_EMAIL: "fixture.invalid",
  GIT_COMMITTER_NAME: "VocaNova Fixture",
  GIT_COMMITTER_EMAIL: "fixture.invalid",
  GIT_AUTHOR_DATE: "2026-08-24T00:00:00Z",
  GIT_COMMITTER_DATE: "2026-08-24T00:00:00Z",
};
const fixture = (name) =>
  JSON.parse(
    readFileSync(
      path.join(repositoryRoot, "scripts/foundation/fixtures", name),
      "utf8",
    ),
  );
const VOC139_PROMOTION_BASE_SHA = "0d0b0cdf0692d0349f380e9cae3285b4c7916b05";
const VOC139_PROMOTION_HEAD_SHA = "4812fb91ab1b674f9a9ec03906f90c0edf50421d";
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
const reviewedHeadSha = () =>
  execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();

function assertValidSourceHashes(sourceHashes) {
  assert.ok(sourceHashes, "fixture source_hashes are required");
  assert.match(
    sourceHashes.agents_sha256 ?? "",
    SHA256_PATTERN,
    "malformed agents_sha256 in fixture source_hashes",
  );
  assert.match(
    sourceHashes.navigator_skill_sha256 ?? "",
    SHA256_PATTERN,
    "malformed navigator_skill_sha256 in fixture source_hashes",
  );
}

function fixtureBlobExistsAtRevision(revision, fixtureRelativePath) {
  return (
    spawnSync("git", ["cat-file", "-e", `${revision}:${fixtureRelativePath}`], {
      cwd: repositoryRoot,
    }).status === 0
  );
}

function sourceHashesAtPinnedAnchor() {
  return {
    agents_sha256: sha256AtRevision(VOC112_PINNED_ANCHOR, AGENTS_PATH),
    navigator_skill_sha256: sha256AtRevision(
      VOC112_PINNED_ANCHOR,
      NAVIGATOR_SKILL_PATH,
    ),
  };
}

function assertSourceHashesMatchPinnedAnchor(sourceHashes, label) {
  assertValidSourceHashes(sourceHashes);
  const expected = sourceHashesAtPinnedAnchor();
  assert.equal(
    sourceHashes.agents_sha256,
    expected.agents_sha256,
    `${label}: agents_sha256 must bind to pinned fixture anchor ${VOC112_PINNED_ANCHOR}`,
  );
  assert.equal(
    sourceHashes.navigator_skill_sha256,
    expected.navigator_skill_sha256,
    `${label}: navigator_skill_sha256 must bind to pinned fixture anchor ${VOC112_PINNED_ANCHOR}`,
  );
}

function assertAllEmbeddedSourceHashesAtPinnedAnchor(fixtureRelativePath) {
  if (fixtureRelativePath === VOC112_FIXTURES.benchmark) {
    assertSourceHashesMatchPinnedAnchor(
      fixture("voc112-navigation-benchmark-traces.json").source_hashes,
      "benchmark fixture",
    );
    return;
  }
  if (fixtureRelativePath === VOC112_FIXTURES.discovery) {
    for (const [index, row] of fixture(
      "voc112-skill-discovery-evidence.json",
    ).discoveries.entries()) {
      assertSourceHashesMatchPinnedAnchor(
        row.source_hashes,
        `discovery row ${index}`,
      );
    }
    return;
  }
  throw new Error(`unknown fixture path: ${fixtureRelativePath}`);
}

function resolveFixturePinnedAnchor(fixtureRelativePath) {
  const pinnedAnchor = VOC112_FIXTURE_PINNED_ANCHORS[fixtureRelativePath];
  if (!pinnedAnchor) {
    throw new Error(`unknown fixture path: ${fixtureRelativePath}`);
  }
  assert.match(
    pinnedAnchor,
    COMMIT_PATTERN,
    "pinned fixture anchor is malformed",
  );
  assert.equal(
    spawnSync("git", ["cat-file", "-e", `${pinnedAnchor}^{commit}`], {
      cwd: repositoryRoot,
    }).status,
    0,
    `pinned fixture anchor object is missing: ${pinnedAnchor}`,
  );
  return pinnedAnchor;
}

function firstParentChainThroughPinnedAnchor(reviewedHead, pinnedAnchor) {
  const revisions = execFileSync(
    "git",
    ["rev-list", "--first-parent", reviewedHead],
    { cwd: repositoryRoot, encoding: "utf8" },
  )
    .trim()
    .split("\n")
    .filter(Boolean);
  const pinIndex = revisions.indexOf(pinnedAnchor);
  if (pinIndex === -1) {
    throw new Error(
      `pinned fixture anchor ${pinnedAnchor} not found in first-parent ancestry of ${reviewedHead}`,
    );
  }
  return revisions.slice(0, pinIndex + 1);
}

function assertImmutableFixtureChain(fixtureRelativePath, reviewedHead) {
  const pinnedAnchor = resolveFixturePinnedAnchor(fixtureRelativePath);
  assert.match(
    reviewedHead ?? "",
    COMMIT_PATTERN,
    "reviewed head revision is malformed",
  );
  assert.equal(
    spawnSync("git", ["cat-file", "-e", `${reviewedHead}^{commit}`], {
      cwd: repositoryRoot,
    }).status,
    0,
    "reviewed head commit object is missing",
  );

  const workingFixtureBlob = sha256(fixtureRelativePath);
  for (const revision of firstParentChainThroughPinnedAnchor(
    reviewedHead,
    pinnedAnchor,
  )) {
    assert.equal(
      spawnSync("git", ["cat-file", "-e", `${revision}^{commit}`], {
        cwd: repositoryRoot,
      }).status,
      0,
      `missing git object while validating fixture chain: ${revision}`,
    );
    if (!fixtureBlobExistsAtRevision(revision, fixtureRelativePath)) {
      throw new Error(
        `fixture missing at revision ${revision} for ${fixtureRelativePath}`,
      );
    }
    assert.equal(
      sha256AtRevision(revision, fixtureRelativePath),
      workingFixtureBlob,
      `fixture blob must remain immutable from reviewed head through pinned anchor at ${revision}`,
    );
  }
  return pinnedAnchor;
}

function commitTree(treeSha, parentRevision, message) {
  const args = ["commit-tree", treeSha];
  if (parentRevision) {
    args.push("-p", parentRevision);
  }
  return execFileSync("git", args, {
    cwd: repositoryRoot,
    encoding: "utf8",
    input: message,
    env: { ...process.env, ...GIT_FIXTURE_ENV },
  }).trim();
}

function treeWithReplacedBlob(baseRevision, relativePath, blobSha) {
  const tempDir = mkdtempSync(path.join(tmpdir(), "voc112-tree-"));
  const indexPath = path.join(tempDir, "index");
  const indexEnv = { ...process.env, GIT_INDEX_FILE: indexPath };
  try {
    execFileSync("git", ["read-tree", `${baseRevision}^{tree}`], {
      cwd: repositoryRoot,
      env: indexEnv,
    });
    execFileSync(
      "git",
      ["update-index", "--cacheinfo", `100644,${blobSha},${relativePath}`],
      { cwd: repositoryRoot, env: indexEnv },
    );
    return execFileSync("git", ["write-tree"], {
      cwd: repositoryRoot,
      encoding: "utf8",
      env: indexEnv,
    }).trim();
  } finally {
    rmSync(tempDir, { recursive: true, force: true });
  }
}

function buildFirstParentDescendants(startRevision, extraCommits) {
  let revision = startRevision;
  for (let index = 0; index < extraCommits; index += 1) {
    revision = commitTree(
      `${revision}^{tree}`,
      revision,
      `VOC-112 fixture chain ${index}\n`,
    );
  }
  return revision;
}

function buildTamperedThenRevertedHead(startRevision, fixtureRelativePath) {
  const baseTree = execFileSync(
    "git",
    ["rev-parse", `${startRevision}^{tree}`],
    {
      cwd: repositoryRoot,
      encoding: "utf8",
    },
  ).trim();
  const workingContent = execFileSync(
    "git",
    ["show", `${startRevision}:${fixtureRelativePath}`],
    { cwd: repositoryRoot },
  );
  const tamperedBlob = execFileSync("git", ["hash-object", "-w", "--stdin"], {
    cwd: repositoryRoot,
    encoding: "utf8",
    input: `${workingContent.toString("utf8")}\n/* tamper */\n`,
  }).trim();
  const tamperedTree = treeWithReplacedBlob(
    startRevision,
    fixtureRelativePath,
    tamperedBlob,
  );
  const tamperedCommit = commitTree(tamperedTree, startRevision, "tamper\n");
  return commitTree(baseTree, tamperedCommit, "revert\n");
}

function assertWholeFixtureAnchor(evidence, fixtureRelativePath, reviewedHead) {
  assertAllEmbeddedSourceHashesAtPinnedAnchor(fixtureRelativePath);
  assertSourceHashesMatchPinnedAnchor(
    evidence.source_hashes,
    "validated capture record",
  );
  const anchor = assertImmutableFixtureChain(fixtureRelativePath, reviewedHead);
  assert.equal(anchor, VOC112_PINNED_ANCHOR);
  assert.equal(
    spawnSync("git", ["merge-base", "--is-ancestor", anchor, reviewedHead], {
      cwd: repositoryRoot,
    }).status,
    0,
    "pinned fixture anchor must be an ancestor of the reviewed head",
  );
  return anchor;
}

function assertPrValidationMergeBase(evidence, fixtureRelativePath) {
  const prBaseSha = process.env.PR_BASE_SHA;
  const prHeadSha = process.env.PR_HEAD_SHA;
  assert.match(
    prBaseSha ?? "",
    COMMIT_PATTERN,
    "post-squash PR validation requires PR_BASE_SHA",
  );
  assert.match(
    prHeadSha ?? "",
    COMMIT_PATTERN,
    "post-squash PR validation requires PR_HEAD_SHA",
  );
  const mergeBaseResult = spawnSync(
    "git",
    ["merge-base", prBaseSha, prHeadSha],
    {
      cwd: repositoryRoot,
      encoding: "utf8",
    },
  );
  assert.equal(
    mergeBaseResult.status,
    0,
    "PR_BASE_SHA and PR_HEAD_SHA must resolve to a common merge base",
  );
  const mergeBase = mergeBaseResult.stdout.trim();
  assert.match(mergeBase, COMMIT_PATTERN);
  if (process.env.VOC112_PROMOTION_PR === "true") {
    assert.equal(
      spawnSync("git", ["merge-base", "--is-ancestor", prBaseSha, prHeadSha], {
        cwd: repositoryRoot,
      }).status,
      0,
      "promotion PR base must be an ancestor of its head",
    );
  }
  assertWholeFixtureAnchor(evidence, fixtureRelativePath, prHeadSha);
}

function usesWholeFixtureAnchor(mode, subjectAvailable) {
  return (
    mode === "pr-validation" ||
    mode === "squash-safe-push" ||
    (mode === "local" && !subjectAvailable)
  );
}

function assertCapturedRevision(evidence, fixtureRelativePath) {
  assert.match(evidence.subject_revision, COMMIT_PATTERN);
  const mode = process.env.VOC112_CAPTURE_PROVENANCE_MODE ?? "local";
  assert.ok(
    ["local", "pr-ancestry", "pr-validation", "squash-safe-push"].includes(
      mode,
    ),
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
  const subjectAvailable = subjectLookup.status === 0;
  const fixtureAnchorMode = usesWholeFixtureAnchor(mode, subjectAvailable);

  if (!subjectAvailable) {
    assert.notEqual(
      mode,
      "pr-ancestry",
      "PR ancestry mode requires every captured commit object",
    );
  }

  if (fixtureAnchorMode) {
    if (mode === "pr-validation") {
      assertPrValidationMergeBase(evidence, fixtureRelativePath);
    } else {
      assertWholeFixtureAnchor(
        evidence,
        fixtureRelativePath,
        reviewedHeadSha(),
      );
    }
  } else {
    const ancestry = spawnSync(
      "git",
      ["merge-base", "--is-ancestor", evidence.subject_revision, "HEAD"],
      { cwd: repositoryRoot },
    );
    const ancestryProven = ancestry.status === 0;
    if (mode === "pr-ancestry") {
      assert.equal(
        ancestry.status,
        0,
        "original capture subject must be an ancestor of the reviewed head",
      );
    }
    if (ancestryProven) {
      assert.equal(
        evidence.source_hashes.navigator_skill_sha256,
        sha256AtRevision(evidence.subject_revision, NAVIGATOR_SKILL_PATH),
        "navigator hash must bind to the captured revision",
      );
      assert.equal(
        evidence.source_hashes.agents_sha256,
        sha256AtRevision(evidence.subject_revision, AGENTS_PATH),
        "AGENTS.md hash must bind to the captured revision",
      );
    }
  }

  assert.equal(
    evidence.source_hashes.navigator_skill_sha256,
    sha256(NAVIGATOR_SKILL_PATH),
  );
  if (!fixtureAnchorMode) {
    assert.equal(evidence.source_hashes.agents_sha256, sha256(AGENTS_PATH));
  }
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

function withProvenanceEnv(mode, baseSha, headSha, callback, options = {}) {
  const prior = {
    mode: process.env.VOC112_CAPTURE_PROVENANCE_MODE,
    base: process.env.PR_BASE_SHA,
    head: process.env.PR_HEAD_SHA,
    promotion: process.env.VOC112_PROMOTION_PR,
  };
  process.env.VOC112_CAPTURE_PROVENANCE_MODE = mode;
  process.env.PR_BASE_SHA = baseSha;
  process.env.PR_HEAD_SHA = headSha;
  if (options.promotionPr) {
    process.env.VOC112_PROMOTION_PR = "true";
  } else {
    delete process.env.VOC112_PROMOTION_PR;
  }
  try {
    callback();
  } finally {
    for (const [name, value] of [
      ["VOC112_CAPTURE_PROVENANCE_MODE", prior.mode],
      ["PR_BASE_SHA", prior.base],
      ["PR_HEAD_SHA", prior.head],
      ["VOC112_PROMOTION_PR", prior.promotion],
    ]) {
      if (value === undefined) delete process.env[name];
      else process.env[name] = value;
    }
  }
}

test("VOC-112-TEST-12: benchmark is a revision-bound real structured agent capture", () => {
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  assert.equal(evidence.schema_version, 2);
  assert.equal(evidence.capture_kind, "real-agent-structured-trace");
  withProvenanceEnv("local", "", "", () =>
    assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
  );
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
      withProvenanceEnv("local", "", "", () =>
        assertCapturedRevision(row, VOC112_FIXTURES.discovery),
      );
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

test("VOC-113-TEST-10: original capture mode requires and accepts true subject ancestry", () => {
  const headSha = execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
  const baseFixture = fixture("voc112-navigation-benchmark-traces.json");
  const evidence = {
    ...baseFixture,
    // The checked-out head exists even in CI's depth-1 clone and is a true
    // ancestor of itself, so this positive is deterministic without hidden
    // history while exercising the strict pr-ancestry branch.
    subject_revision: headSha,
    // This is a synthetic positive using the current head as its capture;
    // its hash must describe that head rather than the historical fixture.
    source_hashes: {
      ...baseFixture.source_hashes,
      agents_sha256: sha256("AGENTS.md"),
    },
  };
  withProvenanceEnv("pr-ancestry", headSha, headSha, () =>
    assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
  );
});

test("VOC-113-TEST-10: original capture mode rejects a fetchable non-ancestor", () => {
  const headSha = execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
  const nonAncestor = execFileSync("git", ["commit-tree", "HEAD^{tree}"], {
    cwd: repositoryRoot,
    encoding: "utf8",
    input: "deterministic non-ancestor fixture\n",
    env: {
      ...process.env,
      GIT_AUTHOR_NAME: "VocaNova Fixture",
      GIT_AUTHOR_EMAIL: "fixture.invalid",
      GIT_COMMITTER_NAME: "VocaNova Fixture",
      GIT_COMMITTER_EMAIL: "fixture.invalid",
      GIT_AUTHOR_DATE: "2026-08-24T00:00:00Z",
      GIT_COMMITTER_DATE: "2026-08-24T00:00:00Z",
    },
  }).trim();
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: nonAncestor,
  };
  withProvenanceEnv("pr-ancestry", headSha, headSha, () =>
    assert.throws(
      () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
      /original capture subject must be an ancestor/,
    ),
  );
});

test("VOC-113-TEST-10: later post-squash PR accepts pinned fixture anchor", () => {
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    // Models a capture commit discarded by squash and no longer fetchable.
    subject_revision: "0".repeat(40),
  };
  const headSha = reviewedHeadSha();
  withProvenanceEnv("pr-validation", VOC112_PINNED_ANCHOR, headSha, () =>
    assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
  );
});

test("VOC-113-TEST-10: tampered merge base fails closed under pr-validation", () => {
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  const headSha = execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
  const wrongBase = "0".repeat(40);
  const priorMode = process.env.VOC112_CAPTURE_PROVENANCE_MODE;
  const priorBase = process.env.PR_BASE_SHA;
  const priorHead = process.env.PR_HEAD_SHA;
  process.env.VOC112_CAPTURE_PROVENANCE_MODE = "pr-validation";
  process.env.PR_BASE_SHA = wrongBase;
  process.env.PR_HEAD_SHA = headSha;
  try {
    assert.throws(
      () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
      (error) =>
        error instanceof assert.AssertionError &&
        /merge base/.test(error.message),
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
  const baseSha = headSha;
  const priorMode = process.env.VOC112_CAPTURE_PROVENANCE_MODE;
  const priorBase = process.env.PR_BASE_SHA;
  const priorHead = process.env.PR_HEAD_SHA;
  process.env.VOC112_CAPTURE_PROVENANCE_MODE = "pr-validation";
  process.env.PR_BASE_SHA = baseSha;
  process.env.PR_HEAD_SHA = headSha;
  try {
    assert.throws(() =>
      assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
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

test("VOC-139-TEST-00: promotion pr-validation accepts pinned fixture anchor", () => {
  const baseSha = VOC139_PROMOTION_BASE_SHA;
  const incidentHeadSha = VOC139_PROMOTION_HEAD_SHA;
  const reviewedHead = reviewedHeadSha();
  assert.notEqual(
    sha256AtRevision(incidentHeadSha, AGENTS_PATH),
    sha256AtRevision(baseSha, AGENTS_PATH),
    "VOC-139 incident fixture must exercise different base/head hashes",
  );
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
  };
  withProvenanceEnv(
    "pr-validation",
    baseSha,
    reviewedHead,
    () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
    { promotionPr: true },
  );
});

test("VOC-139-TEST-02: ordinary pr-validation accepts pinned fixture anchor despite AGENTS drift", () => {
  const baseSha = VOC139_PROMOTION_BASE_SHA;
  const incidentHeadSha = VOC139_PROMOTION_HEAD_SHA;
  const reviewedHead = reviewedHeadSha();
  const mergeBase = execFileSync("git", ["merge-base", baseSha, reviewedHead], {
    cwd: repositoryRoot,
    encoding: "utf8",
  }).trim();
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
  };
  assert.notEqual(
    sha256AtRevision(reviewedHead, AGENTS_PATH),
    sha256AtRevision(mergeBase, AGENTS_PATH),
    "ordinary PR regression requires different merge-base/head AGENTS.md hashes",
  );
  assert.notEqual(
    sha256AtRevision(incidentHeadSha, AGENTS_PATH),
    sha256AtRevision(baseSha, AGENTS_PATH),
    "VOC-139 incident fixture must exercise different base/head hashes",
  );
  withProvenanceEnv("pr-validation", baseSha, reviewedHead, () =>
    assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
  );
});

test("VOC-139-TEST-05: promotion pr-validation rejects split source hash pair", () => {
  const baseSha = VOC139_PROMOTION_BASE_SHA;
  const headSha = VOC139_PROMOTION_HEAD_SHA;
  const expected = sourceHashesAtPinnedAnchor();
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
    source_hashes: {
      agents_sha256: expected.agents_sha256,
      navigator_skill_sha256: "f".repeat(64),
    },
  };
  withProvenanceEnv(
    "pr-validation",
    baseSha,
    headSha,
    () =>
      assert.throws(
        () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
        /navigator_skill_sha256 must bind to pinned fixture anchor/,
      ),
    { promotionPr: true },
  );
});

test("VOC-139-TEST-04: promotion pr-validation rejects a non-ancestor base", () => {
  const parent = execFileSync(
    "git",
    ["rev-parse", `${VOC139_PROMOTION_BASE_SHA}^`],
    { cwd: repositoryRoot, encoding: "utf8" },
  ).trim();
  const tree = execFileSync(
    "git",
    ["rev-parse", `${VOC139_PROMOTION_BASE_SHA}^{tree}`],
    { cwd: repositoryRoot, encoding: "utf8" },
  ).trim();
  const divergentBase = execFileSync(
    "git",
    ["commit-tree", tree, "-p", parent],
    {
      cwd: repositoryRoot,
      encoding: "utf8",
      input: "VOC-139 divergent base fixture\n",
      env: {
        ...process.env,
        GIT_AUTHOR_NAME: "VOC-139 test",
        GIT_AUTHOR_EMAIL: "test@example.invalid",
        GIT_COMMITTER_NAME: "VOC-139 test",
        GIT_COMMITTER_EMAIL: "test@example.invalid",
      },
    },
  ).trim();
  assert.notEqual(
    spawnSync(
      "git",
      ["merge-base", "--is-ancestor", divergentBase, VOC139_PROMOTION_HEAD_SHA],
      { cwd: repositoryRoot },
    ).status,
    0,
  );
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
    source_hashes: {
      navigator_skill_sha256: sha256AtRevision(
        VOC139_PROMOTION_HEAD_SHA,
        ".agents/skills/vocanova-repo-navigator/SKILL.md",
      ),
      agents_sha256: sha256AtRevision(VOC139_PROMOTION_HEAD_SHA, "AGENTS.md"),
    },
  };
  withProvenanceEnv(
    "pr-validation",
    divergentBase,
    VOC139_PROMOTION_HEAD_SHA,
    () =>
      assert.throws(
        () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
        /promotion PR base must be an ancestor of its head/,
      ),
    { promotionPr: true },
  );
});

test("VOC-112-EHR-01: current develop ordinary PR passes with unchanged fixtures", () => {
  const headSha = reviewedHeadSha();
  assert.equal(
    resolveFixturePinnedAnchor(VOC112_FIXTURES.benchmark),
    VOC112_PINNED_ANCHOR,
  );
  assert.equal(
    resolveFixturePinnedAnchor(VOC112_FIXTURES.discovery),
    VOC112_PINNED_ANCHOR,
  );
  assert.notEqual(
    sha256AtRevision(headSha, AGENTS_PATH),
    fixture("voc112-navigation-benchmark-traces.json").source_hashes
      .agents_sha256,
    "current develop must exercise AGENTS.md drift from the pinned fixture anchor",
  );
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
  };
  withProvenanceEnv("pr-validation", VOC112_PINNED_ANCHOR, headSha, () =>
    assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
  );
});

test("VOC-112-EHR-02: squash-safe-push accepts pinned fixture anchor at develop tip", () => {
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
  };
  withProvenanceEnv("squash-safe-push", "", "", () =>
    assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
  );
});

test("VOC-112-EHR-03: discovery rows bind through the pinned fixture anchor", () => {
  const headSha = reviewedHeadSha();
  const row = fixture("voc112-skill-discovery-evidence.json").discoveries[0];
  assert.equal(
    assertWholeFixtureAnchor(row, VOC112_FIXTURES.discovery, headSha),
    VOC112_PINNED_ANCHOR,
  );
  withProvenanceEnv("pr-validation", VOC112_PINNED_ANCHOR, headSha, () =>
    assertCapturedRevision(
      { ...row, subject_revision: "0".repeat(40) },
      VOC112_FIXTURES.discovery,
    ),
  );
});

test("VOC-112-EHR-04: malformed source hashes fail closed", () => {
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
    source_hashes: {
      agents_sha256: "not-a-sha256",
      navigator_skill_sha256: "also-not-a-sha256",
    },
  };
  const headSha = reviewedHeadSha();
  withProvenanceEnv("pr-validation", headSha, headSha, () =>
    assert.throws(
      () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
      /malformed agents_sha256/,
    ),
  );
});

test("VOC-112-EHR-05: unfound source hash pair fails closed", () => {
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
    source_hashes: {
      agents_sha256: "a".repeat(64),
      navigator_skill_sha256: "b".repeat(64),
    },
  };
  const headSha = reviewedHeadSha();
  withProvenanceEnv("pr-validation", headSha, headSha, () =>
    assert.throws(
      () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
      /agents_sha256 must bind to pinned fixture anchor/,
    ),
  );
});

test("VOC-112-EHR-06: strict local mode rejects stale embedded agents hash", () => {
  const headSha = reviewedHeadSha();
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: headSha,
  };
  assert.notEqual(
    evidence.source_hashes.agents_sha256,
    sha256(AGENTS_PATH),
    "local strictness requires AGENTS drift at develop tip",
  );
  withProvenanceEnv("local", "", "", () =>
    assert.throws(
      () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
      /AGENTS\.md hash must bind to the captured revision/,
    ),
  );
});

test("VOC-112-EHR-07: long first-parent chains still resolve to pinned anchor", () => {
  const headSha = reviewedHeadSha();
  const syntheticHead = buildFirstParentDescendants(headSha, 65);
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  assert.equal(
    assertWholeFixtureAnchor(
      evidence,
      VOC112_FIXTURES.benchmark,
      syntheticHead,
    ),
    VOC112_PINNED_ANCHOR,
  );
  assert.equal(
    assertWholeFixtureAnchor(
      evidence,
      VOC112_FIXTURES.discovery,
      syntheticHead,
    ),
    VOC112_PINNED_ANCHOR,
  );
});

test("VOC-112-EHR-08: tampered then reverted fixture history fails closed", () => {
  const headSha = reviewedHeadSha();
  const syntheticHead = buildTamperedThenRevertedHead(
    headSha,
    VOC112_FIXTURES.benchmark,
  );
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  assert.throws(
    () =>
      assertWholeFixtureAnchor(
        evidence,
        VOC112_FIXTURES.benchmark,
        syntheticHead,
      ),
    /fixture blob must remain immutable/,
  );
});

test("VOC-112-EHR-09: altered discovery row source hash pair fails closed", () => {
  const headSha = reviewedHeadSha();
  const row = fixture("voc112-skill-discovery-evidence.json").discoveries[0];
  const evidence = {
    ...row,
    subject_revision: "0".repeat(40),
    source_hashes: {
      ...row.source_hashes,
      navigator_skill_sha256: "c".repeat(64),
    },
  };
  withProvenanceEnv("pr-validation", VOC112_PINNED_ANCHOR, headSha, () =>
    assert.throws(
      () => assertCapturedRevision(evidence, VOC112_FIXTURES.discovery),
      /navigator_skill_sha256 must bind to pinned fixture anchor/,
    ),
  );
});

test("VOC-112-EHR-10: unknown fixture path fails closed", () => {
  const headSha = reviewedHeadSha();
  const evidence = fixture("voc112-navigation-benchmark-traces.json");
  assert.throws(
    () =>
      assertWholeFixtureAnchor(
        evidence,
        "scripts/foundation/fixtures/missing-fixture.json",
        headSha,
      ),
    /unknown fixture path/,
  );
});

test(
  "VOC-112-EHR-11: pr-ancestry probe rejects missing capture subject object",
  {
    skip:
      process.env.VOC112_SHALLOW_SUBJECT_PROBE !== "true"
        ? "probe-only regression"
        : false,
  },
  () => {
    const evidence = {
      ...fixture("voc112-navigation-benchmark-traces.json"),
      subject_revision: "0".repeat(40),
    };
    process.env.VOC112_CAPTURE_PROVENANCE_MODE = "pr-ancestry";
    assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark);
  },
);

test("VOC-112-EHR-12: pr-ancestry rejects missing capture subject object", () => {
  const evidence = {
    ...fixture("voc112-navigation-benchmark-traces.json"),
    subject_revision: "0".repeat(40),
  };
  withProvenanceEnv("pr-ancestry", reviewedHeadSha(), reviewedHeadSha(), () =>
    assert.throws(
      () => assertCapturedRevision(evidence, VOC112_FIXTURES.benchmark),
      /PR ancestry mode requires every captured commit object/,
    ),
  );
});
