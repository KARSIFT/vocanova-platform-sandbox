// VOC-106-T00 — deterministic remediation ownership gate, workflow boundary, and docs checks.

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
const pipelineVerifyPath = path.join(
  repositoryRoot,
  ".github/workflows/pipeline-verify.yml",
);
const liveEvidenceDocPath = path.join(
  repositoryRoot,
  "docs/operations/live-evidence.md",
);
const fixtureInfraRoot = path.join(
  repositoryRoot,
  "tooling/governance/fixtures/karsift-ai-infra",
);
const infraReadmePath = path.join(fixtureInfraRoot, "README.md");
const infraRemediatePath = path.join(
  fixtureInfraRoot,
  ".github/workflows/remediate.yml",
);
const infraVerifierPath = path.join(
  fixtureInfraRoot,
  ".github/workflows/verify-remediate-operator-ownership.yml",
);
const contractPath = path.join(
  repositoryRoot,
  "specs/changes/VOC-106-operator-owned-live-evidence-failures-can/.karsift/live-evidence/VOC-106-T01.yaml",
);

function runInfraTests(extraArgs = []) {
  const result = spawnSync(
    "python3",
    [
      "-m",
      "unittest",
      "discover",
      "-s",
      "tooling/governance/tests",
      "-p",
      "test_remediate_ownership.py",
      ...extraArgs,
      "-v",
    ],
    {
      cwd: repositoryRoot,
      encoding: "utf8",
    },
  );
  assert.equal(
    result.status,
    0,
    result.stderr || result.stdout || "unittest discover failed",
  );
}

test("VOC-106-TEST-00 through TEST-07: remediation ownership classifier fixtures", () => {
  runInfraTests();
});

test("VOC-106-TEST-10: docs describe ownership-gated remediation", () => {
  const liveEvidenceDoc = readFileSync(liveEvidenceDocPath, "utf8");
  const infraReadme = readFileSync(infraReadmePath, "utf8");
  assert.match(liveEvidenceDoc, /`remediate\.yml` does \*\*not\*\* dispatch/);
  assert.match(liveEvidenceDoc, /retain today's bounded remediation retry/);
  assert.match(infraReadme, /suppress `implement\.yml`/);
});

test("VOC-106-TEST-11: caller wiring exposes read-only verify-remediate-operator-ownership action", () => {
  const pipelineVerify = readFileSync(pipelineVerifyPath, "utf8");
  const verifier = readFileSync(infraVerifierPath, "utf8");
  const contract = readFileSync(contractPath, "utf8");
  const remediate = readFileSync(infraRemediatePath, "utf8");
  assert.match(
    pipelineVerify,
    /verify-remediate-operator-ownership/,
    "pipeline must expose the proof action",
  );
  assert.match(
    pipelineVerify,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/verify-remediate-operator-ownership\.yml@main/,
  );
  const verifyBlock =
    pipelineVerify.split("verify-remediate-operator-ownership:", 2)[1] ?? "";
  assert.ok(verifyBlock.length > 0, "verify job must exist");
  assert.match(verifyBlock, /actions: read/);
  assert.doesNotMatch(verifyBlock, /secrets: inherit/);
  assert.match(
    verifyBlock,
    /source_run_id: \$\{\{ inputs\.live_evidence_run_id \}\}/,
  );
  assert.match(
    verifyBlock,
    /pr_number: \$\{\{ inputs\.live_evidence_pr_number \}\}/,
  );
  const dispatchBlock =
    pipelineVerify
      .split("  workflow_dispatch:", 2)[1]
      ?.split("\npermissions:", 1)[0] ?? "";
  const dispatchInputs = dispatchBlock.match(/^      [a-z][a-z0-9_]+:/gm) ?? [];
  assert.ok(
    dispatchInputs.length <= 25,
    "workflow_dispatch must stay within GitHub's 25-input limit",
  );
  assert.match(verifier, /jobs:\s+verify:/);
  assert.match(contract, /- verify-remediate-operator-ownership \/ verify/);
  assert.match(remediate, /remediation-ownership\.py/);
  assert.match(remediate, /--repository-root caller/);
  assert.match(remediate, /--ownership-state "\$ownership_state"/);
  assert.match(remediate, /echo "operator_escalation=true"/);
  assert.match(remediate, /path: caller/);
});
