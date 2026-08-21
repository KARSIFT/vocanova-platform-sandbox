// VOC-097-T01 — caller exact-SHA and superseded-run safeguards.

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
  "specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t01-evidence.md",
);

test("VOC-097 caller wiring binds lifecycle decisions to one PR head", () => {
  const pipeline = readFileSync(pipelinePath, "utf8");
  const exactBinding =
    "expected_head_sha: ${{ github.event.pull_request.head.sha }}";

  assert.match(
    pipeline,
    /group: \$\{\{ github\.workflow \}\}-\$\{\{ github\.event\.pull_request\.number \|\| github\.run_id \}\}/,
  );
  assert.match(
    pipeline,
    /cancel-in-progress: \$\{\{ github\.event_name == 'pull_request' && github\.event\.action != 'closed' \}\}/,
  );
  assert.equal(
    pipeline.split(exactBinding).length - 1,
    5,
    "plan review, task review, remediation, merge gate, and ready-for-review reuse must receive the triggering PR SHA",
  );
});

test("VOC-097-TEST-05: evidence records the least-privilege shared policy", () => {
  assert.ok(existsSync(evidencePath), "t01-evidence.md must exist");
  const evidence = readFileSync(evidencePath, "utf8");

  assert.match(evidence, /evidence_id:\s*VOC-097-EV-01/);
  assert.match(evidence, /gate_status:\s*complete/);
  assert.match(evidence, /post_merge_source_run_claimed:\s*false/);
  assert.match(evidence, /VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE/);
  assert.match(evidence, /should_retry=false/);
  assert.match(evidence, /implementer.*no.*actions.*permission/is);
  assert.doesNotMatch(evidence, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});
