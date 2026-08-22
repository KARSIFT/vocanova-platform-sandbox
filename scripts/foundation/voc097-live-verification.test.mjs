// VOC-097-T05 — live verification evidence structure and claim gates.
//
// Hosted Actions proof is recorded only in t05-evidence.md when the
// corresponding live_*_claimed flags are true. These tests never treat
// fixture coverage as a substitute for the hosted waiting/reconcile runs.

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const t05EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t05-evidence.md",
);
const liveEvidenceDocPath = path.join(
  repositoryRoot,
  "docs/operations/live-evidence.md",
);
const pipelinePath = path.join(
  repositoryRoot,
  ".github/workflows/pipeline.yml",
);
const observerWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/operational-failure-monitoring.yml",
);
const errorMonitoringWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/error-monitoring.yml",
);
const voc102ResultPath = path.join(
  repositoryRoot,
  "specs/changes/VOC-102-auto-advance-dispatches-implementer-for-operator/.karsift/live-evidence/VOC-102-T01.result.json",
);

function frontMatterFlag(source, name) {
  const match = source.match(new RegExp(`^${name}:\\s*(\\S+)`, "m"));
  return match?.[1] ?? "";
}

test("VOC-097-T05 evidence file exists with required frontmatter", () => {
  assert.ok(existsSync(t05EvidencePath), "t05-evidence.md must exist");
  const evidence = readFileSync(t05EvidencePath, "utf8");

  assert.equal(frontMatterFlag(evidence, "evidence_id"), "VOC-097-EV-05");
  assert.equal(frontMatterFlag(evidence, "task_id"), "VOC-097-T05");
  assert.match(evidence, /gate_status:/);
  assert.match(evidence, /live_fixture_claimed:/);
  assert.match(evidence, /live_reconcile_claimed:/);
  assert.match(evidence, /observer_health_claimed:/);
  assert.match(evidence, /rollback_owner:/);
  assert.match(evidence, /docs\/operations\/live-evidence\.md/);
});

test("VOC-097-TEST-11 live claim is evidence-gated (reconcile wake + fresh review)", () => {
  const evidence = readFileSync(t05EvidencePath, "utf8");
  const claimed = frontMatterFlag(evidence, "live_reconcile_claimed");

  assert.match(evidence, /live-evidence-reconcile/);
  assert.match(evidence, /fresh exact-SHA/i);
  assert.ok(
    existsSync(voc102ResultPath),
    "qualified result fixture must exist",
  );

  if (claimed === "true") {
    assert.match(
      evidence,
      /actions\/runs\/\d+/,
      "live reconcile proof requires a recorded Actions run URL",
    );
    assert.match(
      evidence,
      /state:\s*qualified|qualified/,
      "live reconcile proof must reference qualification",
    );
    const reconcileRun = frontMatterFlag(evidence, "reconcile_proof_run");
    assert.match(reconcileRun, /^\d+$/, "reconcile_proof_run must be numeric");
    assert.match(evidence, new RegExp(reconcileRun));
  }
});

test("VOC-097-TEST-03 live claim is evidence-gated (waiting skips remediation)", () => {
  const evidence = readFileSync(t05EvidencePath, "utf8");
  const claimed = frontMatterFlag(evidence, "live_fixture_claimed");

  assert.match(evidence, /WAITING FOR OPERATOR LIVE EVIDENCE/);
  assert.match(evidence, /remediate \/ retry/);

  if (claimed === "true") {
    const waitingRun = frontMatterFlag(evidence, "waiting_proof_run");
    assert.match(waitingRun, /^\d+$/, "waiting_proof_run must be numeric");
    assert.match(evidence, new RegExp(waitingRun));
    assert.match(
      evidence,
      /skipped|should_retry=false/i,
      "waiting proof must record remediation suppression",
    );
  }
});

test("VOC-097-TEST-16 live claim is evidence-gated (observer/Sentry separation)", () => {
  const evidence = readFileSync(t05EvidencePath, "utf8");
  const claimed = frontMatterFlag(evidence, "observer_health_claimed");

  assert.ok(existsSync(observerWorkflowPath));
  assert.ok(existsSync(errorMonitoringWorkflowPath));
  const observer = readFileSync(observerWorkflowPath, "utf8");
  const pipeline = readFileSync(pipelinePath, "utf8");

  assert.doesNotMatch(
    pipeline,
    /operational-failure-monitoring/,
    "caller pipeline must not wire live-evidence into the failure observer",
  );
  assert.doesNotMatch(observer, /karsift-live-evidence/);

  if (claimed === "true") {
    assert.match(
      evidence,
      /operational-failure-monitoring/,
      "observer separation proof must name the observer workflow",
    );
    assert.match(
      evidence,
      /error-monitoring/,
      "observer separation proof must confirm Sentry path health",
    );
    assert.match(
      evidence,
      /actions\/runs\/\d+/,
      "observer health proof requires scrubbed run references",
    );
  }
});

test("VOC-097-T05 evidence forbids sensitive material", () => {
  const evidence = readFileSync(t05EvidencePath, "utf8");
  const liveDoc = readFileSync(liveEvidenceDocPath, "utf8");

  assert.match(evidence, /No logs|no logs|secrets|OAuth/i);
  assert.doesNotMatch(evidence, /Bearer\s+[A-Za-z0-9._-]+/);
  assert.doesNotMatch(evidence, /ghp_[A-Za-z0-9]+/);
  assert.match(liveDoc, /Forbidden/);
});
