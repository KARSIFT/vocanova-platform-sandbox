// VOC-104-T00 — ready_for_review reuse wiring, docs, and deterministic policy tests.

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
const docPath = path.join(
  repositoryRoot,
  "docs/operations/15-ai-native-product-and-engineering-operating-model.md",
);
const fixtureInfraRoot = path.join(
  repositoryRoot,
  "tooling/governance/fixtures/karsift-ai-infra",
);
const infraReadmePath = path.join(fixtureInfraRoot, "README.md");
const reuseWorkflowPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/ready-for-review-reuse.yml",
);
const verifyWorkflowPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/verify-ready-for-review-reuse.yml",
);
const ciWorkflowPath = path.join(fixtureInfraRoot, ".github/workflows/ci.yml");
const mergeWorkflowPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/merge-gate.yml",
);
const pinPath = path.join(fixtureInfraRoot, "PINNED_SHA.txt");
const fixtureTestsRoot = path.join(fixtureInfraRoot, "tests");
const contractPath = path.join(
  repositoryRoot,
  "specs/changes/VOC-104-ready-for-review-reruns-unchanged-exact-sha-ci/.karsift/live-evidence/VOC-104-T01.yaml",
);
function runInfraTests(extraArgs = []) {
  const result = spawnSync(
    "python3",
    [
      "-m",
      "unittest",
      "discover",
      "-s",
      fixtureTestsRoot,
      "-p",
      "test_*.py",
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
    result.stderr || result.stdout || "ready_for_review reuse tests failed",
  );
}

test("VOC-104-TEST-00 through TEST-07 and TEST-10/11A: infra reuse policy fixtures", () => {
  runInfraTests();
});

test("VOC-104-TEST-11: docs and caller wiring distinguish reuse from full path", () => {
  const doc = readFileSync(docPath, "utf8");
  const infraReadme = readFileSync(infraReadmePath, "utf8");
  const pipeline = readFileSync(pipelinePath, "utf8");
  assert.match(doc, /reuse precondition holds/);
  assert.match(doc, /normal full CI and review path/);
  assert.match(infraReadme, /reuse-evidence/);
  assert.match(infraReadme, /fail-closed-to-full-path/);
  assert.match(infraReadme, /authenticated shared-infra SHA/);
  assert.match(infraReadme, /pre-merge record/);
  assert.match(pipeline, /ready-for-review-reuse/);
  assert.match(pipeline, /reuse_evidence:/);
  assert.match(
    pipeline,
    /needs\.ready-for-review-reuse\.outputs\.outcome == 'reuse-evidence'/,
  );
  assert.doesNotMatch(
    pipeline.split("  ci:", 2)[1].split("  extract-package-path:", 1)[0],
    /needs\.ready-for-review-reuse\.outputs\.outcome != 'reuse-evidence'/,
  );
  assert.match(pipeline, /reuse_outcome:/);
  assert.match(pipeline, /reuse_prior_run_id:/);
});

test("VOC-104 fixture is pinned to the independently reviewed shared merge", () => {
  assert.equal(
    readFileSync(pinPath, "utf8").trim(),
    "12e5cd65159b5315b7e618facb251e0324dcfbb5",
  );
});

test("VOC-104 required CI context uses a marker instead of duplicate validation", () => {
  const workflow = readFileSync(ciWorkflowPath, "utf8");
  assert.match(workflow, /reuse_evidence:/);
  assert.match(workflow, /Record exact-SHA CI evidence reuse/);
  assert.match(workflow, /if: \$\{\{ !inputs\.reuse_evidence \}\}/);
});

test("VOC-104-TEST-01: ready_for_review remains subscribed", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  assert.match(
    pipeline,
    /types: \[opened, synchronize, reopened, ready_for_review, closed\]/,
  );
});

test("VOC-104-TEST-12: verifier is read-only and contract-bound", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  const verifier = readFileSync(verifyWorkflowPath, "utf8");
  const reuse = readFileSync(reuseWorkflowPath, "utf8");
  const mergeGate = readFileSync(mergeWorkflowPath, "utf8");
  const contract = readFileSync(contractPath, "utf8");
  assert.match(pipeline, /verify-ready-for-review-reuse/);
  assert.match(
    pipeline,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/verify-ready-for-review-reuse\.yml@main/,
  );
  const verifyBlock =
    pipeline.split("verify-ready-for-review-reuse:", 2)[1] ?? "";
  assert.match(verifyBlock, /actions: read/);
  assert.doesNotMatch(verifyBlock, /secrets: inherit/);
  assert.match(verifier, /jobs:\s+verify:/);
  assert.match(verifier, /name: verify/);
  assert.match(verifier, /source_pr_number:/);
  assert.match(verifier, /expected_source_head_sha:/);
  assert.match(verifier, /expected_source_base_sha:/);
  assert.match(verifier, /expected_proof_head_sha:/);
  assert.match(reuse, /ready-for-review-reuse-runner\.py/);
  assert.match(reuse, /name: decide \(\$\{\{ inputs\.event_action \}\}\)/);
  assert.match(mergeGate, /Publish immutable reuse transition attestation/);
  assert.match(mergeGate, /policy_sha:/);
  assert.doesNotMatch(pipeline, /verify_reuse_proof_head_sha:/);
  assert.match(pipeline, /expected_proof_head_sha: \$\{\{ github\.sha \}\}/);
  assert.match(contract, /verify-ready-for-review-reuse \/ verify/);
  assert.match(contract, /exact_pr_head/);
  assert.match(contract, /workflow_dispatch/);
});
