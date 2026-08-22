// VOC-097-T04 / VOC-097-TEST-15 — stranded #779 / #785 migration locks.

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

const t04EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t04-evidence.md",
);

const stranded = [
  {
    taskId: "VOC-093-T01",
    issue: "779",
    pr: "789",
    packageDir:
      "specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure",
    branch: "main",
    workflowFile: "scheduled-synthetics.yml",
    jobName: "synthetic.production.authenticated-route-content-sweep",
    lineage: "exact_sha",
    exactSha: "a508647ea5a345f06c975b086c76f8cd40b1624d",
    requiresDispatch: true,
  },
  {
    taskId: "VOC-094-T01",
    issue: "785",
    pr: "791",
    packageDir:
      "specs/changes/VOC-094-operational-failure-deploy-staging-cancelled",
    branch: "develop",
    workflowFile: "deploy-staging.yml",
    jobName: "deploy to staging",
    lineage: "integration_contains_pr_head",
    requiresDispatch: false,
  },
];

function read(relativePath) {
  return readFileSync(path.join(repositoryRoot, relativePath), "utf8");
}

function validateContractYaml(relativePath, taskId) {
  const modulePath = path.join(
    repositoryRoot,
    "tooling/governance/fixtures/karsift-ai-infra/config/live_evidence_reconcile.py",
  );
  const filePath = path.join(repositoryRoot, relativePath);
  const script = `
import importlib.util
from pathlib import Path
spec = importlib.util.spec_from_file_location(
    "live_evidence_reconcile",
    ${JSON.stringify(modulePath)},
)
module = importlib.util.module_from_spec(spec)
import sys
sys.modules[spec.name] = module
spec.loader.exec_module(module)
text = Path(${JSON.stringify(filePath)}).read_text(encoding="utf-8")
parsed = module.parse_contract_yaml(text)
module.validate_contract(parsed, ${JSON.stringify(taskId)})
print("ok")
`;
  const result = spawnSync("python3", ["-c", script], {
    encoding: "utf8",
  });
  assert.equal(
    result.status,
    0,
    result.stderr || result.stdout || "contract validation failed",
  );
}

test("VOC-097-TEST-15: t04-evidence documents governed migration for #779 and #785", () => {
  const evidence = readFileSync(t04EvidencePath, "utf8");
  assert.match(evidence, /evidence_id: VOC-097-EV-04/);
  assert.match(evidence, /task_id: VOC-097-T04/);
  assert.match(evidence, /#779/);
  assert.match(evidence, /#785/);
  assert.match(evidence, /#789/);
  assert.match(evidence, /#791/);
  assert.match(evidence, /voc-093-t01-live-verify/);
  assert.match(evidence, /safe migration/i);
  assert.match(evidence, /live_reconcile_claimed: false/);
  assert.doesNotMatch(evidence, /ghp_[A-Za-z0-9]+/);
});

for (const item of stranded) {
  test(`VOC-097-TEST-15: ${item.taskId} contract and waiting evidence exist`, () => {
    const contractPath = `${item.packageDir}/.karsift/live-evidence/${item.taskId}.yaml`;
    const evidencePath = `${item.packageDir}/t01-evidence.md`;
    const tasksPath = `${item.packageDir}/tasks.md`;

    assert.ok(existsSync(path.join(repositoryRoot, contractPath)));
    assert.ok(existsSync(path.join(repositoryRoot, evidencePath)));

    const contract = read(contractPath);
    assert.match(contract, new RegExp(`task_id: ${item.taskId}`));
    assert.match(contract, /ownership: operator/);
    assert.match(contract, new RegExp(`workflow_file: ${item.workflowFile}`));
    assert.match(contract, new RegExp(`- ${item.jobName}`));
    assert.match(contract, new RegExp(`mode: ${item.lineage}`));
    assert.match(contract, new RegExp(`branch: ${item.branch}`));

    if (item.exactSha) {
      assert.match(contract, new RegExp(`sha: ${item.exactSha}`));
    }

    if (item.requiresDispatch) {
      assert.match(contract, /^dispatch:/m);
    } else {
      assert.doesNotMatch(contract, /^dispatch:/m);
    }

    validateContractYaml(contractPath, item.taskId);

    const taskEvidence = read(evidencePath);
    assert.match(taskEvidence, /waiting-for-operator-live-evidence/);
    assert.match(taskEvidence, new RegExp(item.taskId));
    assert.match(taskEvidence, new RegExp(contractPath.replace(/\./g, "\\.")));

    const tasks = read(tasksPath);
    assert.match(tasks, new RegExp(`${item.taskId}\\.yaml`));
    assert.match(tasks, /live-evidence/);
  });
}

test("VOC-097-TEST-15: VOC-093 pollution must not ship in caller pipeline", () => {
  const pipeline = read(".github/workflows/pipeline.yml");
  assert.doesNotMatch(pipeline, /voc-093-t01-live-verify/);
});

test("VOC-097-TEST-15: production dispatch contracts target the protected production branch", () => {
  const contract = read(
    "specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure/.karsift/live-evidence/VOC-093-T01.yaml",
  );
  assert.match(contract, /^branch: main$/m);
  assert.doesNotMatch(contract, /^branch: agent\//m);
});
