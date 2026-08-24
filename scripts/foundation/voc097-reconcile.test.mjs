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

  const workflowRun = pipeline
    .split("  workflow_run:", 2)[1]
    ?.split("  workflow_dispatch:", 1)[0];
  assert.ok(
    workflowRun,
    "terminal prerequisite workflows must wake authoritative release evaluation",
  );
  assert.doesNotMatch(
    workflowRun,
    /pipeline/,
    "the pipeline must not observe itself and create a recursive wake loop",
  );
  assert.match(
    pipeline,
    /types: \[opened, synchronize, reopened, ready_for_review, closed\]/,
    "an unchanged reviewed draft SHA must be rechecked when it becomes ready",
  );
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
  const operatorJob = pipeline.split("  live-evidence-reconcile:", 2)[1];
  assert.ok(operatorJob, "operator reconciliation job must exist");
  assert.doesNotMatch(
    operatorJob,
    /^\s{4}secrets: inherit$/m,
    "operator job must not inherit unrelated caller secrets",
  );
  assert.match(
    operatorJob,
    /^\s{4}secrets:\n\s{6}KARSIFT_BOT_APP_ID: \$\{\{ secrets\.KARSIFT_BOT_APP_ID \}\}\n\s{6}KARSIFT_BOT_PRIVATE_KEY: \$\{\{ secrets\.KARSIFT_BOT_PRIVATE_KEY \}\}/m,
  );
  assert.match(
    operatorJob,
    /^\s{4}permissions:\n\s{6}actions: write\n\s{6}checks: read\n\s{6}contents: read\n\s{6}issues: read\n\s{6}pull-requests: read/m,
    "only the operator-owned caller job may dispatch the contract-allowlisted workflow",
  );
  const integrationRecoveryJob = pipeline
    .split("  recover-integration-push:", 2)[1]
    ?.split("\n  implement:", 1)[0];
  assert.ok(
    integrationRecoveryJob,
    "scheduled exact-SHA integration recovery job must exist",
  );
  assert.match(
    integrationRecoveryJob,
    /^\s{4}permissions:\n\s{6}actions: write\n\s{6}checks: read\n\s{6}contents: read\n\s{6}pull-requests: read\n\s{6}statuses: read/m,
    "integration recovery may write Actions only through its dedicated job",
  );
  assert.equal(
    (pipeline.match(/^\s{6}actions: write$/gm) ?? []).length,
    4,
    "Actions write must remain isolated to live-evidence, integration recovery, merge-gate, and release jobs",
  );
  assert.equal(
    (
      pipeline.match(
        /expected_base_sha: \$\{\{ github\.event\.pull_request\.base\.sha \}\}/g,
      ) ?? []
    ).length,
    4,
    "plan review, task review, remediation, merge gate, and ready-for-review reuse must bind the event base",
  );
  const reuseBlock =
    pipeline
      .split("\n  ready-for-review-reuse:", 2)[1]
      ?.split("\n  ci:", 1)[0] ?? "";
  assert.match(reuseBlock, /github\.event\.pull_request\.base\.sha \|\| ''/);
  assert.doesNotMatch(
    reuseBlock,
    /github\.event\.pull_request\.base\.sha \|\| github\.sha/,
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

  const implementerJob = pipeline
    .split("  implement:", 2)[1]
    ?.split("\n  plan:", 1)[0];
  assert.ok(implementerJob, "implementer job must exist");
  assert.doesNotMatch(
    implementerJob,
    /^\s{6}actions: write$/m,
    "implementer must not inherit operator Actions-write authority",
  );
});

test("VOC-097-T02 evidence file records mechanism without secrets", () => {
  assert.ok(existsSync(evidencePath), "t02-evidence.md must exist");
  const evidence = readFileSync(evidencePath, "utf8");
  assert.match(evidence, /evidence_id:\s*VOC-097-EV-02/);
  assert.match(
    evidence,
    /caller_recorded_implementation_pair_reviewed:\s*true/,
  );
  assert.doesNotMatch(evidence, /caller_exact_sha_reviewed:/);
  assert.match(evidence, /caller_reviewed_base_sha:\s*[0-9a-f]{40}/);
  assert.match(evidence, /caller_reviewed_implementation_sha:\s*[0-9a-f]{40}/);
  assert.match(evidence, /live-evidence-reconcile\.yml/);
  assert.match(evidence, /shared infrastructure PR/i);
  assert.match(evidence, /hourly metadata reconciliation/i);
  assert.match(evidence, /permission_compatibility_recovery_claimed:\s*true/);
  assert.match(
    evidence,
    /permission_compatibility_validated_sha:\s*[0-9a-f]{40}/,
  );
  assert.match(
    evidence,
    /permission_compatibility_pipeline_run:\s*[1-9][0-9]*/,
  );
  assert.match(
    evidence,
    /permission_compatibility_independent_review_pass:\s*true/,
  );
  assert.doesNotMatch(evidence, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});
