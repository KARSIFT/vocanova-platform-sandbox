// VOC-114-T00 — recovery metadata-read token contract and fail-closed diagnostics.

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
const fixtureInfraRoot = path.join(
  repositoryRoot,
  "tooling/governance/fixtures/karsift-ai-infra",
);
const devopsOperationsPath = path.join(
  repositoryRoot,
  "docs/operations/11-devops-and-ci-cd.md",
);
const repositoryGovernancePath = path.join(
  repositoryRoot,
  ".github/workflows/repository-governance.yml",
);
const governancePolicyPath = path.join(
  repositoryRoot,
  ".github/workflows/governance-policy.yml",
);

function runInfraTests(pattern) {
  const result = spawnSync(
    "python3",
    ["-m", "unittest", "discover", "-s", "tests", "-p", pattern],
    {
      cwd: fixtureInfraRoot,
      encoding: "utf8",
      env: {
        ...process.env,
        PYTHONPATH: path.join(fixtureInfraRoot, "config"),
      },
    },
  );
  assert.equal(
    result.status,
    0,
    result.stderr || result.stdout || `fixture tests failed for ${pattern}`,
  );
}

test("VOC-114-TEST-01 through TEST-05: fixture recovery metadata policy tests", () => {
  runInfraTests("test_*voc114*");
});

test("VOC-114 caller docs describe recovery App read contract", () => {
  const devopsOperations = readFileSync(devopsOperationsPath, "utf8");
  assert.match(devopsOperations, /permission-checks: read/);
  assert.match(devopsOperations, /permission-statuses: read/);
  assert.match(devopsOperations, /permission-actions: write/);
  assert.match(devopsOperations, /check_runs_read_failed/);
  assert.match(devopsOperations, /workflow_runs_read_failed/);
  assert.match(devopsOperations, /commit_metadata_read_failed/);
});

test("VOC-114 fixture mirror declares recovery mint read scopes", () => {
  const mergeGate = readFileSync(
    path.join(fixtureInfraRoot, ".github/workflows/merge-gate.yml"),
    "utf8",
  );
  const release = readFileSync(
    path.join(fixtureInfraRoot, ".github/workflows/release.yml"),
    "utf8",
  );
  const reusable = readFileSync(
    path.join(fixtureInfraRoot, ".github/workflows/recover-actions-checks.yml"),
    "utf8",
  );
  for (const workflow of [mergeGate, release, reusable]) {
    assert.match(workflow, /permission-checks: read/);
    assert.match(workflow, /permission-statuses: read/);
  }
  assert.match(reusable, /permission-actions: write/);
  assert.equal(
    reusable.match(/^\s+permission-actions:/gm)?.length,
    1,
    "the recovery mint must declare Actions once; write includes read capability",
  );
  assert.match(reusable, /permission-contents: read/);
  const runner = readFileSync(
    path.join(fixtureInfraRoot, "config/actions-check-recovery-runner.py"),
    "utf8",
  );
  assert.doesNotMatch(runner, /github_metadata_read_failed/);
});

test("VOC-114 recovery workflows resolve PR metadata before checkout", () => {
  for (const workflowPath of [repositoryGovernancePath, governancePolicyPath]) {
    const workflow = readFileSync(workflowPath, "utf8");
    const resolveStep = workflow
      .split("- name: Resolve ", 2)[1]
      .split("\n      - name:", 1)[0];
    assert.match(resolveStep, /GH_REPO: \$\{\{ github\.repository \}\}/);
    assert.match(resolveStep, /gh pr view/);
  }
});
