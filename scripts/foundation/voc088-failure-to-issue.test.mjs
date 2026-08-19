// VOC-088-T02 — deterministic failure observer, sanitization, identity, and
// deduplication coverage.

import assert from "node:assert/strict";
import { chmodSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const workflowPath = path.join(
  repositoryRoot,
  ".github/workflows/operational-failure-monitoring.yml",
);
const helperPath = path.join(
  repositoryRoot,
  "infra/scripts/open-failure-issue.sh",
);

function runHelper({ existingBody = "", conclusion = "failure" } = {}) {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "voc088-failure-"));
  const callsPath = path.join(fixtureRoot, "calls.jsonl");
  const ghPath = path.join(fixtureRoot, "gh");
  const existingIssues = existingBody
    ? [[{ number: 42, body: existingBody }]]
    : [[]];

  writeFileSync(
    ghPath,
    `#!/usr/bin/env bash
set -euo pipefail
printf '%s\\n' "$(printf '%s' "$*" | jq -Rs .)" >> "$CALLS_PATH"
if [ "$1" = "api" ] && [[ " $* " == *" --method GET "* ]]; then
  printf '%s' "$EXISTING_ISSUES"
  exit 0
fi
if [ "$1" = "api" ] && [[ " $* " == *" --method POST "* ]]; then
  body_arg=""
  previous=""
  for argument in "$@"; do
    if [ "$previous" = "-F" ]; then body_arg="$argument"; fi
    previous="$argument"
  done
  body_file="\${body_arg#body=@}"
  jq -cRs '{body: .}' < "$body_file" >> "$CAPTURE_PATH"
  exit 0
fi
exit 64
`,
  );
  chmodSync(ghPath, 0o755);

  const capturePath = path.join(fixtureRoot, "created.jsonl");
  const result = spawnSync("bash", [helperPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${fixtureRoot}:${process.env.PATH}`,
      CALLS_PATH: callsPath,
      CAPTURE_PATH: capturePath,
      EXISTING_ISSUES: JSON.stringify(existingIssues),
      GH_TOKEN: "fixture-installation-token",
      GH_REPOSITORY: "KARSIFT/vocanova-platform-sandbox",
      FAILURE_WORKFLOW_NAME: "deploy-staging",
      FAILURE_CONCLUSION: conclusion,
      FAILURE_RUN_URL:
        "https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/123456",
    },
  });

  let calls = [];
  try {
    calls = readFileSync(callsPath, "utf8")
      .trim()
      .split("\n")
      .filter(Boolean)
      .map(JSON.parse);
  } catch {
    // Invalid bounded inputs are rejected before any GitHub API call.
  }
  let createdBodies = [];
  try {
    createdBodies = readFileSync(capturePath, "utf8")
      .trim()
      .split("\n")
      .filter(Boolean)
      .map((line) => JSON.parse(line).body);
  } catch {
    // No capture file is the expected deduplication path.
  }
  return { result, calls, createdBodies };
}

test("VOC-088-TEST-08: first failure creates exactly one plain sanitized issue", () => {
  const { result, calls, createdBodies } = runHelper();
  assert.equal(result.status, 0, result.stderr);
  assert.equal(
    calls.filter((call) => call.includes("--method POST")).length,
    1,
  );
  assert.equal(createdBodies.length, 1);
  assert.match(createdBodies[0], /Workflow: `deploy-staging`/);
  assert.match(createdBodies[0], /Conclusion: `failure`/);
  assert.match(createdBodies[0], /actions\/runs\/123456/);
  assert.doesNotMatch(createdBodies[0], /fixture-installation-token/);
  assert.doesNotMatch(
    calls.find((call) => call.includes("--method POST")),
    /label/i,
  );
});

test("VOC-088-TEST-09: an open matching fingerprint prevents a duplicate", () => {
  const { result, calls, createdBodies } = runHelper({
    existingBody: "<!-- operational-failure:deploy-staging:failure -->",
  });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(
    calls.filter((call) => call.includes("--method POST")).length,
    0,
  );
  assert.equal(createdBodies.length, 0);
});

test("VOC-088-TEST-10: helper rejects unbounded metadata", () => {
  const invalidConclusion = runHelper({ conclusion: "success" });
  assert.notEqual(invalidConclusion.result.status, 0);
  assert.equal(
    invalidConclusion.calls.filter((call) => call.includes("--method POST"))
      .length,
    0,
  );

  const helper = readFileSync(helperPath, "utf8");
  assert.match(helper, /Refusing non-canonical workflow run URL/);
  assert.doesNotMatch(helper, /gh\s+run\s+view|\/jobs(?:\/|\?|"|')/i);
});

test("VOC-088-TEST-11: standalone App observer covers exact failure surface", () => {
  const workflow = readFileSync(workflowPath, "utf8");

  assert.match(workflow, /^name: operational-failure-monitoring$/m);
  assert.match(workflow, /^  workflow_run:$/m);
  for (const workflowName of [
    "scheduled-synthetics",
    "deploy-staging",
    "deploy-production",
  ]) {
    assert.match(workflow, new RegExp(`- ${workflowName}`));
  }
  for (const conclusion of ["failure", "cancelled", "timed_out"]) {
    assert.match(workflow, new RegExp(`\\b${conclusion}\\b`));
  }
  assert.match(workflow, /actions\/create-github-app-token@v3/);
  assert.match(workflow, /KARSIFT_BOT_APP_ID/);
  assert.match(workflow, /KARSIFT_BOT_PRIVATE_KEY/);
  assert.match(workflow, /steps\.app-token\.outputs\.token/);
  assert.doesNotMatch(workflow, /github-token:.*github\.token/);
  assert.doesNotMatch(workflow, /secrets\.GITHUB_TOKEN/);
  assert.match(workflow, /head_repository\.full_name == github\.repository/);

  const observedSources = [
    ".github/workflows/scheduled-synthetics.yml",
    ".github/workflows/deploy-staging.yml",
    ".github/workflows/deploy-production.yml",
  ].map((relativePath) =>
    readFileSync(path.join(repositoryRoot, relativePath), "utf8"),
  );
  for (const source of observedSources) {
    assert.doesNotMatch(source, /operational-failure-monitoring/);
    assert.doesNotMatch(source, /open-failure-issue\.sh/);
  }

  const sentryObserver = readFileSync(
    path.join(repositoryRoot, ".github/workflows/error-monitoring.yml"),
    "utf8",
  );
  assert.match(sentryObserver, /Sentry -> GitHub-issue monitoring agent/);
  assert.doesNotMatch(sentryObserver, /operational-failure-monitoring/);
  assert.doesNotMatch(sentryObserver, /open-failure-issue\.sh/);
});
