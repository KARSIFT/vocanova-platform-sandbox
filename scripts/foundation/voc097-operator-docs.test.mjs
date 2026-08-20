// VOC-097-T00 — live-evidence author/operator documentation coverage
// (TEST-00, TEST-01).

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const operatorDocPath = path.join(
  repositoryRoot,
  "docs/operations/live-evidence.md",
);
const operationsReadmePath = path.join(
  repositoryRoot,
  "docs/operations/README.md",
);
const templateReadmePath = path.join(
  repositoryRoot,
  "specs/templates/change-package/README.md",
);
const templateTasksPath = path.join(
  repositoryRoot,
  "specs/templates/change-package/tasks.md",
);
const t00EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t00-evidence.md",
);

test("VOC-097-TEST-00: operator docs describe live-evidence declaration", () => {
  assert.ok(existsSync(operatorDocPath), "live-evidence.md must exist");
  const doc = readFileSync(operatorDocPath, "utf8");

  assert.match(doc, /ownership/i);
  assert.match(doc, /operator|live-actions/);
  assert.match(doc, /\.karsift\/live-evidence\/<task_id>\.yaml/);
  assert.match(doc, /workflow_file/);
  assert.match(doc, /job_names/);
  assert.match(doc, /sha_lineage/);
  assert.match(doc, /fail closed|fails closed/i);
  assert.match(
    doc,
    /logs|artifacts|secrets|OAuth|session|cookie|token|user identifiers/i,
  );
  assert.match(doc, /not.*implementation defect|not an implementation defect/i);
  assert.match(doc, /operational-failure-monitoring\.yml/);
  assert.doesNotMatch(doc, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});

test("VOC-097-TEST-00: operations index links the live-evidence guide", () => {
  const readme = readFileSync(operationsReadmePath, "utf8");
  assert.match(readme, /live-evidence\.md/);
});

test("VOC-097-TEST-00: t00 evidence records DEP-03 contract path resolution", () => {
  assert.ok(existsSync(t00EvidencePath), "t00-evidence.md must exist");
  const evidence = readFileSync(t00EvidencePath, "utf8");

  assert.match(evidence, /evidence_id:\s*VOC-097-EV-00/);
  assert.match(evidence, /VOC-097-AC-00/);
  assert.match(evidence, /VOC-097-DEP-03/);
  assert.match(evidence, /\.karsift\/live-evidence\/<task_id>\.yaml/);
  assert.match(evidence, /gate_status:\s*repository-complete-automation-deferred/);
});

test("VOC-097-TEST-01: change-package template mentions live-evidence ownership", () => {
  const templateReadme = readFileSync(templateReadmePath, "utf8");
  const templateTasks = readFileSync(templateTasksPath, "utf8");

  assert.match(templateReadme, /live-evidence|live evidence/i);
  assert.match(templateReadme, /\.karsift\/live-evidence/);
  assert.match(templateReadme, /operator-owned|ownership:\s*operator/i);
  assert.match(templateReadme, /docs\/operations\/live-evidence\.md/);
  assert.doesNotMatch(
    templateReadme,
    /live-evidence.*TBD|TBD.*live-evidence/i,
  );

  assert.match(templateTasks, /live-evidence|live evidence/i);
  assert.match(templateTasks, /\.karsift\/live-evidence/);
  assert.doesNotMatch(
    templateTasks,
    /live-evidence.*TBD|TBD.*live-evidence/i,
  );
});
