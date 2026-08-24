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

test("VOC-114 caller docs describe the recovery job-token contract", () => {
  const devopsOperations = readFileSync(devopsOperationsPath, "utf8");
  assert.match(devopsOperations, /`GITHUB_TOKEN`/);
  assert.match(devopsOperations, /`checks: read`/);
  assert.match(devopsOperations, /`statuses: read`/);
  assert.match(devopsOperations, /`statuses: write`/);
  assert.match(devopsOperations, /`actions: write`/);
  assert.match(
    devopsOperations,
    /App token[\s\S]*limited to PR, issue, and content mutations/,
  );
  assert.match(devopsOperations, /check_runs_read_failed/);
  assert.match(devopsOperations, /workflow_runs_read_failed/);
  assert.match(devopsOperations, /commit_metadata_read_failed/);
});

test("VOC-114 fixture mirror separates recovery and mutation tokens", () => {
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
  const template = readFileSync(
    path.join(
      fixtureInfraRoot,
      "templates/project-repo/.github/workflows/pipeline.yml",
    ),
    "utf8",
  );
  const templateMergeGate = template
    .split("\n  merge-gate:\n", 2)[1]
    .split("\n  release:\n", 1)[0];
  const templateRelease = template
    .split("\n  release:\n", 2)[1]
    .split("\n  auto-advance:\n", 1)[0];
  for (const workflow of [mergeGate, reusable]) {
    assert.match(workflow, /checks: read/);
    assert.match(workflow, /statuses: read/);
    assert.match(workflow, /actions: write/);
  }
  assert.match(release, /checks: read/);
  assert.match(release, /statuses: write/);
  assert.match(release, /promotion-status-attestation-runner\.py/);
  assert.match(release, /actions: write/);
  assert.match(templateMergeGate, /statuses: read/);
  assert.doesNotMatch(templateMergeGate, /statuses: write/);
  assert.match(templateRelease, /statuses: write/);
  assert.doesNotMatch(templateRelease, /statuses: read/);
  assert.doesNotMatch(
    reusable,
    /Mint App installation token for recovery dispatch/,
  );
  assert.match(reusable, /GH_TOKEN: \$\{\{ github\.token \}\}/);
  assert.doesNotMatch(reusable, /steps\.app-token\.outputs\.token/);
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

test("VOC-114 canonical promotion uses squash-safe provenance on its associated check", () => {
  const workflow = readFileSync(repositoryGovernancePath, "utf8");
  const provenance = workflow
    .split("- name: Select strict capture provenance mode", 2)[1]
    .split("\n      - name:", 1)[0];
  assert.match(provenance, /PR_BASE_REF:.*pull_request\.base\.ref/);
  assert.match(provenance, /PR_HEAD_REF:.*pull_request\.head\.ref/);
  assert.match(
    provenance,
    /PR_HEAD_REPOSITORY:.*pull_request\.head\.repo\.full_name/,
  );
  assert.match(provenance, /CURRENT_REPOSITORY:.*github\.repository/);
  assert.match(
    provenance,
    /PR_BASE_REF.*main[\s\S]*PR_HEAD_REF.*develop[\s\S]*PR_HEAD_REPOSITORY.*CURRENT_REPOSITORY[\s\S]*mode=squash-safe-push/,
  );
  assert.match(provenance, /else[\s\S]*mode=pr-validation/);
  assert.match(provenance, /git diff --quiet[\s\S]*mode=pr-ancestry/);
});
