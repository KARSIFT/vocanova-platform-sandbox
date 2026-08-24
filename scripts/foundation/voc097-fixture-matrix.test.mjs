// VOC-097-T03 — deterministic fixture matrix for TEST-02 through TEST-14.
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

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
const evidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t03-evidence.md",
);
const fixturePinPath = path.join(
  repositoryRoot,
  "tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt",
);

const MATRIX = [
  {
    testId: "VOC-097-TEST-02",
    module: "test_voc097_live_evidence_lifecycle.py",
    method:
      "test_voc097_test_02_waiting_marker_is_machine_detectable_and_fail_dominant",
  },
  {
    testId: "VOC-097-TEST-03",
    module: "test_voc097_live_evidence_lifecycle.py",
    method: "test_voc097_test_03_waiting_does_not_set_remediation_retry",
  },
  {
    testId: "VOC-097-TEST-04",
    module: "test_voc097_live_evidence_lifecycle.py",
    method: "test_voc097_test_04_genuine_fail_and_ci_failure_still_retry",
  },
  {
    testId: "VOC-097-TEST-05",
    module: "test_voc097_live_evidence_lifecycle.py",
    method: "test_voc097_test_05_implementer_has_no_general_actions_permission",
  },
  {
    testId: "VOC-097-TEST-06",
    module: "test_voc097_live_evidence_reconcile.py",
    method: "test_voc097_test_06_wrong_workflow_identity_is_rejected",
  },
  {
    testId: "VOC-097-TEST-07",
    module: "test_voc097_live_evidence_reconcile.py",
    method: "test_voc097_test_07_wrong_or_missing_required_job_is_rejected",
  },
  {
    testId: "VOC-097-TEST-08",
    module: "test_voc097_live_evidence_reconcile.py",
    method:
      "test_voc097_test_08_wrong_event_branch_and_sha_lineage_are_rejected",
  },
  {
    testId: "VOC-097-TEST-09",
    module: "test_voc097_live_evidence_reconcile.py",
    method:
      "test_voc097_test_09_qualifying_output_contains_allowlisted_metadata_only",
  },
  {
    testId: "VOC-097-TEST-10",
    module: "test_voc097_live_evidence_reconcile.py",
    method: "test_voc097_test_10_workflow_never_calls_log_or_artifact_apis",
  },
  {
    testId: "VOC-097-TEST-11",
    module: "test_voc097_live_evidence_reconcile.py",
    method:
      "test_voc097_test_11_qualification_is_one_commit_then_fresh_pr_review",
  },
  {
    testId: "VOC-097-TEST-12",
    module: "test_voc097_live_evidence_reconcile.py",
    method: "test_voc097_test_12_stale_and_non_success_runs_are_rejected",
  },
  {
    testId: "VOC-097-TEST-13",
    module: "test_voc097_live_evidence_reconcile.py",
    method: "test_voc097_test_13_timeout_is_bounded_and_marker_is_single_use",
  },
  {
    testId: "VOC-097-TEST-14",
    module: "test_voc097_live_evidence_reconcile.py",
    method:
      "test_voc097_test_14_duplicate_result_short_circuits_reconciliation",
  },
];

function runPythonTests(pattern) {
  return spawnSync(
    "python3",
    [
      "-m",
      "unittest",
      "discover",
      "-s",
      "tooling/governance/tests",
      "-p",
      pattern,
      "-v",
    ],
    {
      cwd: repositoryRoot,
      encoding: "utf8",
    },
  );
}

test("VOC-097-T03 matrix declares TEST-02 through TEST-14 without gaps", () => {
  const ids = MATRIX.map((entry) => entry.testId);
  for (let index = 2; index <= 14; index += 1) {
    const expected = `VOC-097-TEST-${String(index).padStart(2, "0")}`;
    assert.ok(ids.includes(expected), `missing ${expected}`);
  }
  assert.equal(ids.length, 13);
});

test("VOC-097-T03 vendored infra pin is recorded for fixture replay", () => {
  assert.ok(existsSync(fixturePinPath), "PINNED_SHA.txt must exist");
  const pin = readFileSync(fixturePinPath, "utf8").trim();
  // VOC-117 advances the pinned fixture to the reviewed karsift-ai-infra merge
  // while preserving the full VOC-097 regression matrix; keep this exact rather
  // than accepting any SHA.
  assert.equal(pin, "27a44b298f1c234a94e02127eaeb55d66b28e30d");
});

test("VOC-097-T03 Python fixture matrix passes", () => {
  const result = runPythonTests("test_voc097_*.py");
  const combined = `${result.stdout}\n${result.stderr}`;
  assert.equal(result.status, 0, combined || "python unittest failed");
  const count = combined.match(/Ran (\d+) tests/)?.[1];
  assert.ok(
    count,
    "Python unittest summary must report its executed test count",
  );
  assert.ok(
    Number(count) >= MATRIX.length,
    `expected at least ${MATRIX.length} Python tests, received ${count}`,
  );
});

test("VOC-097-T03 matrix methods exist in the Python modules", () => {
  for (const entry of MATRIX) {
    const modulePath = path.join(
      repositoryRoot,
      "tooling/governance/tests",
      entry.module,
    );
    const source = readFileSync(modulePath, "utf8");
    assert.match(
      source,
      new RegExp(`def ${entry.method}\\(`),
      `${entry.testId} must map to ${entry.method}`,
    );
  }
});

test("VOC-097-T03 evidence file records matrix completion without secrets", () => {
  assert.ok(existsSync(evidencePath), "t03-evidence.md must exist");
  const evidence = readFileSync(evidencePath, "utf8");
  assert.match(evidence, /evidence_id:\s*VOC-097-EV-03/);
  assert.match(evidence, /VOC-097-AC-09/);
  for (let index = 2; index <= 14; index += 1) {
    assert.match(
      evidence,
      new RegExp(`VOC-097-TEST-${String(index).padStart(2, "0")}`),
    );
  }
  assert.match(evidence, /18 tests passed|Ran 18 tests/);
  assert.doesNotMatch(evidence, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});
