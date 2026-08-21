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
  assert.match(
    liveEvidenceDoc,
    /`remediate\.yml` does \*\*not\*\* dispatch/,
  );
  assert.match(
    liveEvidenceDoc,
    /retain today's bounded remediation retry/,
  );
  assert.match(
    infraReadme,
    /suppress `implement\.yml`/,
  );
});

test("VOC-106-TEST-11: caller wiring exposes read-only verify-remediate-operator-ownership action", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  const verifier = readFileSync(infraVerifierPath, "utf8");
  const contract = readFileSync(contractPath, "utf8");
  const remediate = readFileSync(infraRemediatePath, "utf8");
  assert.match(
    pipeline,
    /verify-remediate-operator-ownership/,
    "pipeline must expose the proof action",
  );
  assert.match(
    pipeline,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/verify-remediate-operator-ownership\.yml@main/,
  );
  const verifyBlock =
    pipeline.split("verify-remediate-operator-ownership:", 2)[1] ?? "";
  assert.ok(verifyBlock.length > 0, "verify job must exist");
  assert.match(verifyBlock, /actions: read/);
  assert.doesNotMatch(verifyBlock, /secrets: inherit/);
  assert.match(verifier, /jobs:\s+verify:/);
  assert.match(contract, /- verify-remediate-operator-ownership \/ verify/);
  assert.match(remediate, /remediate-ownership-classifier\.py/);
  assert.match(remediate, /remediate-escalate-operator\.py/);
  assert.match(remediate, /remediate-fail-closed\.py/);
  assert.match(remediate, /path: pr-head/);
});
