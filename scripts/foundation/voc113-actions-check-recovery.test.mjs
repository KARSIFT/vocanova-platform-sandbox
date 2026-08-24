// VOC-113-T00 — missing Actions check recovery wiring and deterministic policy tests.

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
const pipelinePath = path.join(
  repositoryRoot,
  ".github/workflows/pipeline.yml",
);
const governancePolicyPath = path.join(
  repositoryRoot,
  ".github/workflows/governance-policy.yml",
);
const repositoryGovernancePath = path.join(
  repositoryRoot,
  ".github/workflows/repository-governance.yml",
);
const fixtureInfraRoot = path.join(
  repositoryRoot,
  "tooling/governance/fixtures/karsift-ai-infra",
);
const mergeWorkflowPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/merge-gate.yml",
);
const releaseWorkflowPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/release.yml",
);
const fixtureCiPath = path.join(fixtureInfraRoot, ".github/workflows/ci.yml");
const completionRunnerPath = path.join(
  fixtureInfraRoot,
  "config/task-completion-runner.py",
);
const verifyPromotionWorkflowPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/verify-promotion-check-recovery.yml",
);
const verifyPostPromotionWorkflowPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/verify-post-promotion-workflow.yml",
);
const contractT01Path = path.join(
  repositoryRoot,
  "specs/changes/VOC-113-recover-missing-actions-checks-after-automated/.karsift/live-evidence/VOC-113-T01.yaml",
);
const contractT02Path = path.join(
  repositoryRoot,
  "specs/changes/VOC-113-recover-missing-actions-checks-after-automated/.karsift/live-evidence/VOC-113-T02.yaml",
);
const fixtureTestsRoot = path.join(fixtureInfraRoot, "tests");

function runInfraTests(extraArgs = []) {
  const result = spawnSync(
    "python3",
    [
      "-m",
      "unittest",
      "tests.test_voc113_actions_check_recovery",
      ...extraArgs,
      "-v",
    ],
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
    result.stderr || result.stdout || "VOC-113 policy tests failed",
  );
}

test("VOC-113-TEST-00 through TEST-07 and TEST-10: infra recovery policy fixtures", () => {
  runInfraTests();
});

test("VOC-113 caller wiring exposes recovery and read-only verifiers", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  const governancePolicy = readFileSync(governancePolicyPath, "utf8");
  const repositoryGovernance = readFileSync(repositoryGovernancePath, "utf8");
  const mergeGate = readFileSync(mergeWorkflowPath, "utf8");
  const release = readFileSync(releaseWorkflowPath, "utf8");

  assert.match(pipeline, /recover-promotion-pr-checks/);
  assert.match(pipeline, /promotion_pr_number:/);
  assert.match(
    pipeline,
    /promotion_pr_number: \$\{\{ inputs\.promotion_pr_number \}\}/,
  );
  const ciBlock =
    pipeline.split("\n  ci:", 2)[1]?.split("\n  plan-review:", 1)[0] ?? "";
  assert.match(ciBlock, /inputs\.action == 'recover-promotion-pr-checks'/);
  assert.doesNotMatch(pipeline, /\n  recover-promotion-pr-checks:/);
  const dispatchInputBlock =
    pipeline
      .split("  workflow_dispatch:", 2)[1]
      ?.split("\n# Explicit floor", 1)[0] ?? "";
  assert.ok(
    [...dispatchInputBlock.matchAll(/^      [a-z0-9_]+:$/gm)].length <= 25,
    "GitHub accepts at most 25 workflow_dispatch inputs",
  );
  assert.doesNotMatch(pipeline, /verify_reuse_proof_head_sha:/);
  assert.match(pipeline, /expected_proof_head_sha: \$\{\{ github\.sha \}\}/);
  assert.match(pipeline, /verify-promotion-check-recovery/);
  assert.match(pipeline, /verify-post-promotion-workflow/);
  assert.match(pipeline, /resolve-integration-recovery-target:/);
  assert.match(pipeline, /git\/ref\/heads\/develop/);
  assert.match(pipeline, /recover-integration-push:/);
  assert.match(pipeline, /recovery_mode: integration_push/);
  assert.match(
    pipeline,
    /target_sha: \$\{\{ needs\.resolve-integration-recovery-target\.outputs\.target_sha \}\}/,
  );
  assert.match(
    pipeline,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/verify-promotion-check-recovery\.yml@main/,
  );
  assert.match(
    pipeline,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/verify-post-promotion-workflow\.yml@main/,
  );
  assert.match(governancePolicy, /recovery_pr_number/);
  assert.match(repositoryGovernance, /recovery_pr_number/);
  assert.match(repositoryGovernance, /pr-validation/);
  assert.match(repositoryGovernance, /pr-ancestry/);
  assert.match(repositoryGovernance, /Select strict capture provenance mode/);
  assert.match(
    readFileSync(governancePolicyPath, "utf8"),
    /inputs\.recovery_pr_number \|\| github\.run_id/,
  );
  assert.match(
    mergeGate,
    /Recover missing integration push workflows for merged SHA/,
  );
  assert.match(mergeGate, /actions-check-recovery-runner\.py/);
  assert.match(release, /Recover missing exact-head promotion checks/);
  assert.doesNotMatch(
    mergeGate.split("Recover missing integration", 1)[1] ?? "",
    /github\.token merge/i,
  );
});

test("VOC-113 fixture mirror keeps completion CLI and checkout pin coherent", () => {
  const help = spawnSync(
    "python3",
    [completionRunnerPath, "publish", "--help"],
    {
      cwd: fixtureInfraRoot,
      encoding: "utf8",
      env: {
        ...process.env,
        PYTHONPATH: path.join(fixtureInfraRoot, "config"),
      },
    },
  );
  assert.equal(help.status, 0, help.stderr || help.stdout);
  assert.match(help.stdout, /--reviewed-base-sha/);
  const fixtureCi = readFileSync(fixtureCiPath, "utf8");
  assert.doesNotMatch(fixtureCi, /actions\/checkout@11d5960a/);
  assert.match(fixtureCi, /actions\/checkout@3d3c42e5/);
  assert.match(fixtureCi, /persist-credentials: false/);
});

test("VOC-113-TEST-08/09 contracts bind read-only verifier job names", () => {
  const verifyPromotion = readFileSync(verifyPromotionWorkflowPath, "utf8");
  const verifyPostPromotion = readFileSync(
    verifyPostPromotionWorkflowPath,
    "utf8",
  );
  const contractT01 = readFileSync(contractT01Path, "utf8");
  const contractT02 = readFileSync(contractT02Path, "utf8");
  assert.match(verifyPromotion, /name: verify/);
  assert.match(verifyPostPromotion, /name: verify/);
  assert.match(contractT01, /verify-promotion-check-recovery \/ verify/);
  assert.match(contractT02, /verify-post-promotion-workflow \/ verify/);
  assert.match(contractT01, /exact_pr_head/);
  assert.match(contractT02, /exact_pr_head/);
  assert.doesNotMatch(contractT01, /^dispatch:/m);
  assert.doesNotMatch(contractT02, /^dispatch:/m);
  assert.match(contractT01, /promotion_pr_number=947/);
  assert.match(contractT02, /promotion_pr_number=947/);
});

test("VOC-113-TEST-12: verifier jobs are read-only", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  const verifyPromotionBlock =
    pipeline.split("verify-promotion-check-recovery:", 2)[1] ?? "";
  const verifyPostPromotionBlock =
    pipeline.split("verify-post-promotion-workflow:", 2)[1] ?? "";
  assert.match(verifyPromotionBlock, /actions: read/);
  assert.match(verifyPostPromotionBlock, /actions: read/);
  assert.doesNotMatch(verifyPromotionBlock, /secrets: inherit/);
  assert.doesNotMatch(verifyPostPromotionBlock, /secrets: inherit/);
});
