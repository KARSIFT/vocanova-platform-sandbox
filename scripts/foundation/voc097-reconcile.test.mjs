// VOC-097-T02 — live-evidence reconciler caller wiring and policy locks.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
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
const evidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t02-evidence.md",
);
const infraWorkflowPath = path.join(
  repositoryRoot,
  "karsift-ai-infra/.github/workflows/live-evidence-reconcile.yml",
);

test("VOC-097-T02 caller wires observe, timeout, and dispatch reconcile paths", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");

  assert.match(pipeline, /workflow_run:\s*\n\s*types: \[completed\]/);
  assert.match(pipeline, /schedule:\s*\n\s*- cron: "0 \* \* \* \*"/);
  assert.match(pipeline, /reconcile-live-evidence/);
  assert.match(
    pipeline,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/live-evidence-reconcile\.yml@main/,
  );
  assert.match(
    pipeline,
    /mode: \$\{\{ github\.event_name == 'schedule' && 'timeout'/,
  );
  assert.match(
    pipeline,
    /workflow_run_id: \$\{\{ github\.event\.workflow_run\.id \|\| '' \}\}/,
  );
});

test("VOC-097-TEST-05: reconciler permissions stay off the implementer", () => {
  assert.ok(
    existsSync(infraWorkflowPath),
    "local infra checkout must contain reconcile workflow",
  );
  const workflow = readFileSync(infraWorkflowPath, "utf8");
  const implementFixture = readFileSync(
    path.join(
      repositoryRoot,
      "karsift-ai-infra/.github/workflows/implement.yml",
    ),
    "utf8",
  );
  const permissionBlock = implementFixture.match(
    /    permissions:\n([\s\S]*?)    steps:\n/,
  );
  assert.ok(permissionBlock, "implement.yml must declare job permissions");
  assert.doesNotMatch(permissionBlock[1], /actions:/);
});

test("VOC-097-TEST-10: reconcile workflow avoids log and artifact APIs", () => {
  const workflow = readFileSync(infraWorkflowPath, "utf8");
  const runner = readFileSync(
    path.join(
      repositoryRoot,
      "karsift-ai-infra/config/live-evidence-reconcile-runner.py",
    ),
    "utf8",
  );
  const combined = workflow + runner;
  assert.doesNotMatch(combined, /\/actions\/jobs\/\$.*\/logs/);
  assert.doesNotMatch(combined, /download-artifact/);
  assert.doesNotMatch(combined, /\/actions\/artifacts/);
});

test("VOC-097-TEST-11: wake path chains review and merge-gate on exact head", () => {
  const workflow = readFileSync(infraWorkflowPath, "utf8");
  assert.match(
    workflow,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/review\.yml@main/,
  );
  assert.match(
    workflow,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/merge-gate\.yml@main/,
  );
  assert.match(
    workflow,
    /expected_head_sha: \$\{\{ needs\.reconcile\.outputs\.head_sha \}\}/,
  );
  assert.match(
    workflow,
    /pr_number: \$\{\{ needs\.reconcile\.outputs\.pr_number \}\}/,
  );
});

test("VOC-097-T02 evidence file records mechanism without secrets", () => {
  assert.ok(existsSync(evidencePath), "t02-evidence.md must exist");
  const evidence = readFileSync(evidencePath, "utf8");
  assert.match(evidence, /evidence_id:\s*VOC-097-EV-02/);
  assert.match(evidence, /live-evidence-reconcile\.yml/);
  assert.match(evidence, /allowlisted keys/i);
  assert.doesNotMatch(evidence, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});
