// VOC-094-T00 — deploy concurrency queue wiring and benign-cancel classifier.

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
const deployStagingPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);
const deployProductionPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-production.yml",
);
const observerWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/operational-failure-monitoring.yml",
);
const classifierPath = path.join(
  repositoryRoot,
  "infra/scripts/classify-deploy-concurrency-cancel.sh",
);
const openIssueHelperPath = path.join(
  repositoryRoot,
  "infra/scripts/open-failure-issue.sh",
);

function readConcurrencyBlock(workflowText) {
  const match = workflowText.match(/^concurrency:\n((?:  .+\n)+)/m);
  assert.ok(match, "expected a concurrency block");
  return match[1];
}

function runClassifier({
  workflowName = "deploy-staging",
  conclusion = "cancelled",
  runId = "32290409156",
  jobsTotalCount = 0,
  apiFails = false,
  missingTotalCount = false,
} = {}) {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "voc094-classifier-"));
  const ghPath = path.join(fixtureRoot, "gh");
  const jobsPayload = missingTotalCount
    ? "{}"
    : JSON.stringify({ total_count: jobsTotalCount, jobs: [] });

  writeFileSync(
    ghPath,
    `#!/usr/bin/env bash
set -euo pipefail
if [ "$1" = "api" ] && [[ " $* " == *" --method GET "* ]] && [[ " $* " == *"/jobs"* ]]; then
  if [ "$API_FAILS" = "1" ]; then
    exit 1
  fi
  printf '%s' "$JOBS_PAYLOAD"
  exit 0
fi
echo "unexpected gh invocation: $*" >&2
exit 64
`,
  );
  chmodSync(ghPath, 0o755);

  return spawnSync("bash", [classifierPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${fixtureRoot}:${process.env.PATH}`,
      API_FAILS: apiFails ? "1" : "0",
      JOBS_PAYLOAD: jobsPayload,
      GH_TOKEN: "fixture-installation-token",
      GH_REPOSITORY: "KARSIFT/vocanova-platform-sandbox",
      FAILURE_WORKFLOW_NAME: workflowName,
      FAILURE_CONCLUSION: conclusion,
      FAILURE_RUN_ID: runId,
    },
  });
}

function runOpenIssueHelper({ conclusion = "failure", jobsTotalCount } = {}) {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "voc094-open-issue-"));
  const callsPath = path.join(fixtureRoot, "calls.jsonl");
  const ghPath = path.join(fixtureRoot, "gh");
  const capturePath = path.join(fixtureRoot, "created.jsonl");

  writeFileSync(
    ghPath,
    `#!/usr/bin/env bash
set -euo pipefail
printf '%s\\n' "$(printf '%s' "$*" | jq -Rs .)" >> "$CALLS_PATH"
if [ "$1" = "api" ] && [[ " $* " == *" --method GET "* ]] && [[ " $* " == *"/issues?"* ]]; then
  printf '[]'
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

  const result = spawnSync("bash", [openIssueHelperPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${fixtureRoot}:${process.env.PATH}`,
      CALLS_PATH: callsPath,
      CAPTURE_PATH: capturePath,
      GH_TOKEN: "fixture-installation-token",
      GH_REPOSITORY: "KARSIFT/vocanova-platform-sandbox",
      FAILURE_WORKFLOW_NAME: "deploy-staging",
      FAILURE_CONCLUSION: conclusion,
      FAILURE_RUN_URL:
        "https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/123456",
    },
  });

  let createdBodies = [];
  try {
    createdBodies = readFileSync(capturePath, "utf8")
      .trim()
      .split("\n")
      .filter(Boolean)
      .map((line) => JSON.parse(line).body);
  } catch {
    // No issue created.
  }

  return { result, createdBodies, jobsTotalCount };
}

test("VOC-094-TEST-00: evidence file records run 32290409156 supersession metadata", () => {
  const evidencePath = path.join(
    repositoryRoot,
    "specs/changes/VOC-094-operational-failure-deploy-staging-cancelled/t00-evidence.md",
  );
  const evidence = readFileSync(evidencePath, "utf8");
  assert.match(evidence, /32290409156/);
  assert.match(evidence, /staging-deploy/);
  assert.match(evidence, /total_count:\s*0|zero jobs/i);
  assert.match(
    evidence,
    /higher priority waiting request|concurrency queue supersession/i,
  );
});

test("VOC-094-TEST-01: deploy-staging.yml declares queue max with cancel-in-progress false", () => {
  const block = readConcurrencyBlock(readFileSync(deployStagingPath, "utf8"));
  assert.match(block, /group: staging-deploy/);
  assert.match(block, /cancel-in-progress: false/);
  assert.match(block, /queue: max/);
  assert.doesNotMatch(block, /cancel-in-progress: true/);
});

test("VOC-094-TEST-02: deploy-production.yml declares queue max with cancel-in-progress false", () => {
  const block = readConcurrencyBlock(
    readFileSync(deployProductionPath, "utf8"),
  );
  assert.match(block, /group: production-deploy/);
  assert.match(block, /cancel-in-progress: false/);
  assert.match(block, /queue: max/);
  assert.doesNotMatch(block, /cancel-in-progress: true/);
});

test("VOC-094-TEST-03: classifier skips concurrency-superseded deploy-staging cancel fixture", () => {
  for (const workflowName of ["deploy-staging", "deploy-production"]) {
    const result = runClassifier({
      workflowName,
      conclusion: "cancelled",
      jobsTotalCount: 0,
    });
    assert.equal(result.status, 0, result.stderr);
    assert.match(result.stdout, /zero jobs started/);
  }
});

test("VOC-094-TEST-04: classifier remains fail-closed for real failures and ambiguous cancels", () => {
  assert.equal(
    runClassifier({ conclusion: "failure", jobsTotalCount: 0 }).status,
    1,
  );
  assert.equal(
    runClassifier({ conclusion: "timed_out", jobsTotalCount: 0 }).status,
    1,
  );
  assert.equal(
    runClassifier({
      workflowName: "scheduled-synthetics",
      conclusion: "cancelled",
      jobsTotalCount: 0,
    }).status,
    1,
  );
  assert.equal(
    runClassifier({ conclusion: "cancelled", jobsTotalCount: 1 }).status,
    1,
  );
  assert.equal(
    runClassifier({ conclusion: "cancelled", apiFails: true }).status,
    1,
  );
  assert.equal(
    runClassifier({ conclusion: "cancelled", missingTotalCount: true }).status,
    1,
  );
});

test("VOC-094-TEST-04b: open-failure-issue still creates issues for deploy failure fixtures", () => {
  const failure = runOpenIssueHelper({ conclusion: "failure" });
  assert.equal(failure.result.status, 0, failure.result.stderr);
  assert.equal(failure.createdBodies.length, 1);

  const cancelled = runOpenIssueHelper({ conclusion: "cancelled" });
  assert.equal(cancelled.result.status, 0, cancelled.result.stderr);
  assert.equal(cancelled.createdBodies.length, 1);
});

test("VOC-094-TEST-05: observer workflow wires classifier before open-failure-issue", () => {
  const workflow = readFileSync(observerWorkflowPath, "utf8");
  const classifierIndex = workflow.indexOf(
    "infra/scripts/classify-deploy-concurrency-cancel.sh",
  );
  const openIssueIndex = workflow.indexOf(
    "infra/scripts/open-failure-issue.sh",
  );
  assert.notEqual(classifierIndex, -1);
  assert.notEqual(openIssueIndex, -1);
  assert.ok(classifierIndex < openIssueIndex);
  assert.match(
    workflow,
    /steps\.classify-cancel\.outputs\.skip_issue != 'true'/,
  );
  assert.match(
    workflow,
    /FAILURE_RUN_ID: \$\{\{ github\.event\.workflow_run\.id \}\}/,
  );
  assert.match(workflow, /actions\/create-github-app-token@v3/);
  assert.match(workflow, /permission-actions: read/);
  assert.match(workflow, /permission-issues: write/);
  assert.doesNotMatch(workflow, /gh\s+run\s+view/i);

  const classifier = readFileSync(classifierPath, "utf8");
  assert.match(classifier, /\/actions\/runs\/\$\{FAILURE_RUN_ID\}\/jobs/);
  assert.doesNotMatch(classifier, /gh\s+run\s+view|step output|job logs/i);
});
