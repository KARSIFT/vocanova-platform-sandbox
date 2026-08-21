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
test("VOC-097-T02 caller wires bounded polling and explicit observe/dispatch paths", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");

  assert.doesNotMatch(pipeline, /^\s{2}workflow_run:/m);
  assert.match(pipeline, /schedule:\s*\n\s*- cron: "0 \* \* \* \*"/);
  assert.match(pipeline, /reconcile-live-evidence/);
  assert.match(pipeline, /live_evidence_mode:/);
  assert.match(pipeline, /options: \[reconcile, observe, dispatch\]/);
  assert.match(pipeline, /live_evidence_run_id:/);
  assert.match(pipeline, /live_evidence_pr_number:/);
  assert.match(
    pipeline,
    /uses: KARSIFT\/karsift-ai-infra\/\.github\/workflows\/live-evidence-reconcile\.yml@main/,
  );
  assert.match(
    pipeline,
    /mode: \$\{\{ github\.event_name == 'schedule' && 'reconcile' \|\| inputs\.live_evidence_mode \}\}/,
  );
  assert.match(
    pipeline,
    /workflow_run_id: \$\{\{ inputs\.live_evidence_run_id \}\}/,
  );
  assert.match(
    pipeline,
    /pr_number: \$\{\{ inputs\.live_evidence_pr_number \}\}/,
  );
  assert.equal(
    (
      pipeline.match(
        /expected_base_sha: \$\{\{ github\.event\.pull_request\.base\.sha \}\}/g,
      ) ?? []
    ).length,
    4,
    "plan review, task review, remediation, and merge gate must bind the event base",
  );
});

test("VOC-097-TEST-05: reconciler permissions stay off the implementer", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  const permissionBlock = pipeline.match(
    /^permissions:\n([\s\S]*?)^concurrency:/m,
  );
  assert.ok(
    permissionBlock,
    "pipeline must declare an explicit permission floor",
  );
  assert.match(permissionBlock[1], /^  actions: read$/m);
  assert.doesNotMatch(permissionBlock[1], /^  actions: write$/m);
});

test("VOC-097-T02 evidence file records mechanism without secrets", () => {
  assert.ok(existsSync(evidencePath), "t02-evidence.md must exist");
  const evidence = readFileSync(evidencePath, "utf8");
  assert.match(evidence, /evidence_id:\s*VOC-097-EV-02/);
  assert.match(evidence, /live-evidence-reconcile\.yml/);
  assert.match(evidence, /shared infrastructure PR/i);
  assert.match(evidence, /hourly metadata reconciliation/i);
  assert.doesNotMatch(evidence, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});
