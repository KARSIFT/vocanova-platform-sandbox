// VOC-108-T00 — authoritative lifecycle and caller wake-path contracts.

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const fixtureRoot = path.join(
  repositoryRoot,
  "tooling/governance/fixtures/karsift-ai-infra",
);
const pipeline = readFileSync(
  path.join(repositoryRoot, ".github/workflows/pipeline.yml"),
  "utf8",
);
const sharedRelease = readFileSync(
  path.join(fixtureRoot, ".github/workflows/release.yml"),
  "utf8",
);
const sharedMerge = readFileSync(
  path.join(fixtureRoot, ".github/workflows/merge-gate.yml"),
  "utf8",
);
const sharedAdvance = readFileSync(
  path.join(fixtureRoot, ".github/workflows/auto-advance.yml"),
  "utf8",
);
const sharedCallerTemplate = readFileSync(
  path.join(
    fixtureRoot,
    "templates/project-repo/.github/workflows/pipeline.yml",
  ),
  "utf8",
);
const authorityDocs = [
  readFileSync(path.join(repositoryRoot, "AGENTS.md"), "utf8"),
  readFileSync(
    path.join(
      repositoryRoot,
      "docs/operations/15-ai-native-product-and-engineering-operating-model.md",
    ),
    "utf8",
  ),
  pipeline,
  readFileSync(path.join(fixtureRoot, "README.md"), "utf8"),
  readFileSync(path.join(fixtureRoot, "prompts/plan.md"), "utf8"),
  sharedCallerTemplate,
].join("\n");

test("VOC-108-TEST-00 through TEST-09: pinned shared policy suite", () => {
  const result = spawnSync(
    "python3",
    ["-m", "unittest", "discover", "-s", "tests", "-p", "test_*.py"],
    {
      cwd: fixtureRoot,
      encoding: "utf8",
      env: { ...process.env, PYTHONPATH: path.join(fixtureRoot, "config") },
    },
  );
  assert.equal(
    result.status,
    0,
    result.stderr || result.stdout || "shared lifecycle policy tests failed",
  );
});

test("VOC-108-TEST-07: terminal gates wake cheap release evaluation", () => {
  assert.match(pipeline, /check_run:\s+types: \[completed\]/);
  assert.match(pipeline, /workflow_run:[\s\S]*workflows: \[deploy-staging,/);
  const workflowRun = pipeline
    .split("  workflow_run:", 2)[1]
    .split("  workflow_dispatch:", 1)[0];
  assert.doesNotMatch(workflowRun, /pipeline/);
  const release = pipeline
    .split("  release:", 2)[1]
    .split("  auto-advance:", 1)[0];
  assert.match(release, /needs: \[merge-gate\]/);
  assert.match(release, /always\(\)/);
  assert.match(release, /github\.event_name == 'pull_request'/);
  assert.match(release, /github\.event_name == 'workflow_run'/);
  assert.match(release, /github\.event_name == 'check_run'/);
  assert.match(release, /github\.event\.check_run\.pull_requests\[0\] != null/);
  assert.match(
    release,
    /github\.event\.check_run\.pull_requests\[0\]\.base\.ref == 'main'/,
  );
  assert.match(
    release,
    /github\.event\.check_run\.pull_requests\[0\]\.head\.ref == 'develop'/,
  );
});

test("VOC-108-TEST-04 through TEST-08: one marker validator and merge authority", () => {
  assert.match(sharedMerge, /task-completion-runner\.py publish/);
  assert.match(
    sharedMerge,
    /Publish task completion marker and close linked task issue\s+if: startsWith\(github\.event\.pull_request\.head\.ref, 'agent\/'\)/,
  );
  assert.match(sharedAdvance, /task-completion-runner\.py validate-task/);
  assert.match(sharedAdvance, /authoritative caller-merge marker; safe no-op/);
  assert.doesNotMatch(sharedAdvance, /check-completion/);
  assert.match(sharedAdvance, /serialized convergence evaluator/);
  assert.match(sharedRelease, /task-completion-runner\.py validate-roster/);
  assert.match(sharedRelease, /authoritative-checks-runner\.py/);
  assert.match(sharedRelease, /--pull-request-file \/tmp\/release-pr\.json/);
  assert.match(sharedMerge, /--pull-request-file \/tmp\/merge-pr\.json/);
  assert.match(sharedMerge, /reject_foreign_repository_closing_text/);
  assert.match(sharedRelease, /group: release-converge-/);
  assert.equal(
    (sharedRelease.match(/gh pr merge/g) ?? []).length,
    1,
    "shared release must expose one final merge command",
  );
});

test("VOC-108-TEST-08: caller and shared docs name marker-bound authority", () => {
  assert.doesNotMatch(
    authorityDocs,
    /roster closes|every task in a package's roster closes|issue closing is what triggers promotion|release-approval issue|release promotion \(one human decision/i,
  );
  assert.match(authorityDocs, /closed state alone cannot advance/i);
  assert.match(authorityDocs, /App-authored completion marker/i);
  assert.match(sharedMerge, /publish the immutable task-completion/i);
  assert.match(
    sharedMerge,
    /Publish task completion marker and close linked task issue/,
  );
  assert.doesNotMatch(sharedMerge, /see "Close linked task issue" below/);
  assert.match(
    sharedCallerTemplate,
    /options: \[[^\]]*verify-remediate-operator-ownership[^\]]*\]/,
  );
  assert.match(
    sharedCallerTemplate,
    /\n  verify-remediate-operator-ownership:\n/,
  );
});

test("VOC-108 fixture is pinned to the consumed shared merge", () => {
  assert.equal(
    readFileSync(path.join(fixtureRoot, "PINNED_SHA.txt"), "utf8").trim(),
    "9d7e334f917643c42bb4b7a062c8fcddecc7927f",
  );
});
