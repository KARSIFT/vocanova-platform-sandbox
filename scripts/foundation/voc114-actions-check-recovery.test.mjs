// VOC-114-T00 — recovery metadata-read token contract and fail-closed diagnostics.

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  chmodSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const fixtureInfraRoot = path.join(
  repositoryRoot,
  "tooling/governance/fixtures/karsift-ai-infra",
);
const devopsOperationsPath = path.join(
  repositoryRoot,
  "docs/operations/11-devops-and-ci-cd.md",
);
const repositoryGovernancePath = path.join(
  repositoryRoot,
  ".github/workflows/repository-governance.yml",
);
const governancePolicyPath = path.join(
  repositoryRoot,
  ".github/workflows/governance-policy.yml",
);

function extractProvenanceStep(workflow) {
  return workflow
    .split("- name: Select strict capture provenance mode", 2)[1]
    .split("\n      - name:", 1)[0];
}

function runFixtureDiffSelector(gitDiffExit, promotionRefs = {}) {
  const binDir = mkdtempSync(path.join(tmpdir(), "voc114-git-"));
  const gitWrapper = path.join(binDir, "git");
  writeFileSync(
    gitWrapper,
    `#!/usr/bin/env bash
if [ "$1" = "diff" ]; then
  exit ${gitDiffExit}
fi
exit 0
`,
  );
  chmodSync(gitWrapper, 0o755);

  const script = `
set -euo pipefail
mode=squash-safe-push
if git diff --quiet "base" "head" -- \\
  scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json \\
  scripts/foundation/fixtures/voc112-skill-discovery-evidence.json; then
  fixture_diff_status=0
else
  fixture_diff_status=$?
fi
case "$fixture_diff_status" in
  1)
    mode=pr-ancestry
    ;;
  0)
    if [ "$PR_BASE_REF" = "main" ] && \\
      [ "$PR_HEAD_REF" = "develop" ] && \\
      [ "$PR_HEAD_REPOSITORY" = "$CURRENT_REPOSITORY" ]; then
      mode=squash-safe-push
    else
      mode=pr-validation
    fi
    ;;
  *)
    echo "capture fixture comparison failed" >&2
    exit 1
    ;;
esac
printf '%s' "$mode"
`;

  const result = spawnSync("bash", ["-c", script], {
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${binDir}:${process.env.PATH}`,
      PR_BASE_REF: promotionRefs.baseRef ?? "main",
      PR_HEAD_REF: promotionRefs.headRef ?? "develop",
      PR_HEAD_REPOSITORY:
        promotionRefs.headRepository ?? "KARSIFT/vocanova-platform",
      CURRENT_REPOSITORY:
        promotionRefs.currentRepository ?? "KARSIFT/vocanova-platform",
    },
  });
  rmSync(binDir, { recursive: true, force: true });
  return result;
}

function runInfraTests(pattern) {
  const result = spawnSync(
    "python3",
    ["-m", "unittest", "discover", "-s", "tests", "-p", pattern],
    {
      cwd: fixtureInfraRoot,
      encoding: "utf8",
      env: {
        ...process.env,
        PYTHONPATH: path.join(fixtureInfraRoot, "config"),
      },
    },
  );
  assert.equal(
    result.status,
    0,
    result.stderr || result.stdout || `fixture tests failed for ${pattern}`,
  );
}

test("VOC-114-TEST-01 through TEST-05: fixture recovery metadata policy tests", () => {
  runInfraTests("test_*voc114*");
});

test("VOC-114 caller docs describe the recovery job-token contract", () => {
  const devopsOperations = readFileSync(devopsOperationsPath, "utf8");
  assert.match(devopsOperations, /`GITHUB_TOKEN`/);
  assert.match(devopsOperations, /`checks: read`/);
  assert.match(devopsOperations, /`statuses: read`/);
  assert.match(devopsOperations, /`statuses: write`/);
  assert.match(devopsOperations, /`actions: write`/);
  assert.match(
    devopsOperations,
    /App token[\s\S]*limited to PR, issue, and content mutations/,
  );
  assert.match(devopsOperations, /check_runs_read_failed/);
  assert.match(devopsOperations, /workflow_runs_read_failed/);
  assert.match(devopsOperations, /commit_metadata_read_failed/);
});

test("VOC-114 fixture mirror separates recovery and mutation tokens", () => {
  const mergeGate = readFileSync(
    path.join(fixtureInfraRoot, ".github/workflows/merge-gate.yml"),
    "utf8",
  );
  const release = readFileSync(
    path.join(fixtureInfraRoot, ".github/workflows/release.yml"),
    "utf8",
  );
  const reusable = readFileSync(
    path.join(fixtureInfraRoot, ".github/workflows/recover-actions-checks.yml"),
    "utf8",
  );
  const template = readFileSync(
    path.join(
      fixtureInfraRoot,
      "templates/project-repo/.github/workflows/pipeline.yml",
    ),
    "utf8",
  );
  const templateMergeGate = template
    .split("\n  merge-gate:\n", 2)[1]
    .split("\n  release:\n", 1)[0];
  const templateRelease = template
    .split("\n  release:\n", 2)[1]
    .split("\n  auto-advance:\n", 1)[0];
  for (const workflow of [mergeGate, reusable]) {
    assert.match(workflow, /checks: read/);
    assert.match(workflow, /statuses: read/);
    assert.match(workflow, /actions: write/);
  }
  assert.match(release, /checks: read/);
  assert.match(release, /statuses: write/);
  assert.match(release, /promotion-status-attestation-runner\.py/);
  assert.match(release, /actions: write/);
  assert.match(templateMergeGate, /statuses: read/);
  assert.doesNotMatch(templateMergeGate, /statuses: write/);
  assert.match(templateRelease, /statuses: write/);
  assert.doesNotMatch(templateRelease, /statuses: read/);
  assert.doesNotMatch(
    reusable,
    /Mint App installation token for recovery dispatch/,
  );
  assert.match(reusable, /GH_TOKEN: \$\{\{ github\.token \}\}/);
  assert.doesNotMatch(reusable, /steps\.app-token\.outputs\.token/);
  const runner = readFileSync(
    path.join(fixtureInfraRoot, "config/actions-check-recovery-runner.py"),
    "utf8",
  );
  assert.doesNotMatch(runner, /github_metadata_read_failed/);
});

test("VOC-114 recovery workflows resolve PR metadata before checkout", () => {
  for (const workflowPath of [repositoryGovernancePath, governancePolicyPath]) {
    const workflow = readFileSync(workflowPath, "utf8");
    const resolveStep = workflow
      .split("- name: Resolve ", 2)[1]
      .split("\n      - name:", 1)[0];
    assert.match(resolveStep, /GH_REPO: \$\{\{ github\.repository \}\}/);
    assert.match(resolveStep, /gh pr view/);
  }
});

test("VOC-114 canonical promotion uses squash-safe provenance on its associated check", () => {
  const provenance = extractProvenanceStep(
    readFileSync(repositoryGovernancePath, "utf8"),
  );
  assert.match(provenance, /PR_BASE_REF:.*pull_request\.base\.ref/);
  assert.match(provenance, /PR_HEAD_REF:.*pull_request\.head\.ref/);
  assert.match(
    provenance,
    /PR_HEAD_REPOSITORY:.*pull_request\.head\.repo\.full_name/,
  );
  assert.match(provenance, /CURRENT_REPOSITORY:.*github\.repository/);
  assert.match(
    provenance,
    /git diff --quiet[\s\S]*voc112-navigation-benchmark-traces\.json/,
  );
  assert.match(
    provenance,
    /git diff --quiet[\s\S]*voc112-skill-discovery-evidence\.json/,
  );
  assert.match(provenance, /fixture_diff_status=\$\?/);
  assert.match(provenance, /if git diff --quiet/);
  assert.match(provenance, /case "\$fixture_diff_status" in/);
  assert.match(provenance, /1\)[\s\S]*mode=pr-ancestry/);
  assert.match(
    provenance,
    /0\)[\s\S]*PR_BASE_REF.*main[\s\S]*PR_HEAD_REF.*develop[\s\S]*PR_HEAD_REPOSITORY.*CURRENT_REPOSITORY[\s\S]*mode=squash-safe-push/,
  );
  assert.match(provenance, /mode=squash-safe-push[\s\S]*mode=pr-validation/);
  assert.match(
    provenance,
    /\*\)[\s\S]*capture fixture comparison failed[\s\S]*exit 1/,
  );
  assert.doesNotMatch(provenance, /if ! git diff --quiet/);
});

test("VOC-112-EHR workflow selector: fixture diff precedes promotion exception", () => {
  const provenance = extractProvenanceStep(
    readFileSync(repositoryGovernancePath, "utf8"),
  );
  for (const fixturePath of [
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
  ]) {
    assert.match(provenance, new RegExp(fixturePath.replaceAll("/", "\\/")));
  }
  const fixtureDiffIndex = provenance.indexOf("git diff --quiet");
  const promotionIndex = provenance.indexOf('[ "$PR_BASE_REF" = "main" ]');
  const prValidationIndex = provenance.indexOf("mode=pr-validation");
  assert.ok(fixtureDiffIndex >= 0);
  assert.ok(promotionIndex > fixtureDiffIndex);
  assert.ok(prValidationIndex > promotionIndex);
  assert.match(
    provenance,
    /git diff --quiet[\s\S]*voc112-navigation-benchmark-traces\.json[\s\S]*voc112-skill-discovery-evidence\.json[\s\S]*mode=pr-ancestry/,
  );
});

test("VOC-112-EHR dual-fixture selector: exit 0 unchanged, 1 changed, >1 fail-closed", () => {
  const provenance = extractProvenanceStep(
    readFileSync(repositoryGovernancePath, "utf8"),
  );
  assert.match(provenance, /fixture_diff_status=\$\?/);
  assert.match(provenance, /capture fixture comparison failed/);
  assert.doesNotMatch(provenance, /if ! git diff --quiet/);

  const unchangedPromotion = runFixtureDiffSelector(0);
  assert.equal(unchangedPromotion.status, 0, unchangedPromotion.stderr);
  assert.equal(unchangedPromotion.stdout, "squash-safe-push");

  const unchangedOrdinary = runFixtureDiffSelector(0, {
    headRef: "feature/example",
  });
  assert.equal(unchangedOrdinary.status, 0, unchangedOrdinary.stderr);
  assert.equal(unchangedOrdinary.stdout, "pr-validation");

  const changed = runFixtureDiffSelector(1);
  assert.equal(changed.status, 0, changed.stderr);
  assert.equal(changed.stdout, "pr-ancestry");

  const comparisonFailure = runFixtureDiffSelector(2);
  assert.notEqual(comparisonFailure.status, 0);
  assert.match(
    comparisonFailure.stderr,
    /capture fixture comparison failed/,
    comparisonFailure.stdout,
  );
});
