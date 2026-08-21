// VOC-102-T00 — deterministic ownership gate, workflow boundary, and docs checks.

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const pipelinePath = path.join(repositoryRoot, ".github/workflows/pipeline.yml");
const liveEvidenceDocPath = path.join(
  repositoryRoot,
  "docs/operations/live-evidence.md",
);
const infraAutoAdvancePath = path.join(
  repositoryRoot,
  "karsift-ai-infra/.github/workflows/auto-advance.yml",
);
const infraReadmePath = path.join(repositoryRoot, "karsift-ai-infra/README.md");

function runInfraTests() {
  const result = spawnSync(
    "python3",
    [
      "-m",
      "unittest",
      "discover",
      "-s",
      "tests",
      "-p",
      "test_auto_advance_ownership.py",
      "-v",
    ],
    {
      cwd: path.join(repositoryRoot, "karsift-ai-infra"),
      encoding: "utf8",
    },
  );
  assert.equal(
    result.status,
    0,
    result.stderr || result.stdout || "unittest discover failed",
  );
}

test("VOC-102-TEST-00 through TEST-07: infra ownership classifier fixtures", () => {
  runInfraTests();
});

test("VOC-102-TEST-10: docs describe skip vs dispatch when touched", () => {
  const liveEvidenceDoc = readFileSync(liveEvidenceDocPath, "utf8");
  const infraReadme = readFileSync(infraReadmePath, "utf8");
  assert.match(
    liveEvidenceDoc,
    /auto-advance\.yml` skips general\s+implementer dispatch/is,
  );
  assert.match(
    infraReadme,
    /operator-owned or live-actions-only.*auto-advance instead prepares/is,
  );
  assert.doesNotMatch(
    infraReadme,
    /auto-advance\.yml.*always dispatches implement for every next task/is,
  );
});

test("VOC-102-TEST-11 through TEST-13: carrier, permissions, verifier", () => {
  const source = readFileSync(
    path.join(
      repositoryRoot,
      "karsift-ai-infra/tests/test_auto_advance_ownership.py",
    ),
    "utf8",
  );
  for (const method of [
    "test_voc102_test_11_evidence_path_is_strict_and_idempotent_helpers",
    "test_voc102_test_12_permission_boundary",
    "test_voc102_test_13_verifier_fail_closed_and_exact_head",
  ]) {
    assert.match(source, new RegExp(`def ${method}\\(`));
  }
});

test("VOC-102 caller wiring exposes read-only verify-auto-advance-live-evidence action", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  assert.match(
    pipeline,
    /verify-auto-advance-live-evidence/,
    "pipeline must expose the proof action",
  );
  assert.match(
    pipeline,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/verify-auto-advance-live-evidence\.yml@main/,
  );
  const verifyBlock = pipeline.split("verify-auto-advance-live-evidence:", 2)[1] ?? "";
  assert.ok(verifyBlock.length > 0, "verify-auto-advance-live-evidence job must exist");
  assert.match(verifyBlock, /actions: read/);
  assert.doesNotMatch(verifyBlock, /secrets: inherit/);
});

test("VOC-102 auto-advance consumes ownership gate outputs", () => {
  assert.ok(existsSync(infraAutoAdvancePath), "infra auto-advance workflow must exist");
  const workflow = readFileSync(infraAutoAdvancePath, "utf8");
  assert.match(workflow, /auto-advance-classifier\.py/);
  assert.match(workflow, /prepare-live-evidence:/);
  assert.match(workflow, /decision == 'prepare-live-evidence'/);
  assert.match(workflow, /decision == 'fail-closed'/);
  assert.match(workflow, /auto-advance-carrier-publisher\.py/);
  assert.match(
    workflow,
    /prepare-live-evidence:[\s\S]*auto-advance-carrier-publisher\.py[\s\S]*fail-closed:/,
  );
  assert.doesNotMatch(
    workflow.match(
      /prepare-live-evidence:[\s\S]*?fail-closed:/,
    )[0],
    /implement\.yml/,
  );
});
