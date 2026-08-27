// VOC-111-T00 — fail-closed push path selection for deploy-staging.
//
// Runs via `node --test scripts/foundation/voc111-deploy-staging-paths.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const workflowPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);
const devopsDocPath = path.join(
  repositoryRoot,
  "docs/operations/11-devops-and-ci-cd.md",
);
const evidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-111-skip-full-staging-deploys-for-documentation-and/t00-evidence.md",
);

const EXPECTED_PUSH_PATHS = [
  "*",
  "apps/**",
  "packages/**",
  "infra/**",
  "tests/staging-e2e/**",
  ".github/workflows/deploy-staging.yml",
  "scripts/foundation/voc111-deploy-staging-paths.test.mjs",
];

function readWorkflow() {
  return readFileSync(workflowPath, "utf8");
}

function extractPushPathsBlock(source) {
  const pushStart = source.indexOf("  push:");
  assert.ok(pushStart >= 0, "deploy-staging.yml must define on.push");
  const dispatchStart = source.indexOf("  workflow_dispatch:", pushStart);
  assert.ok(
    dispatchStart > pushStart,
    "deploy-staging.yml must define workflow_dispatch",
  );
  return source.slice(pushStart, dispatchStart);
}

function extractPushPaths(source) {
  const pushBlock = extractPushPathsBlock(source);
  const pathsStart = pushBlock.indexOf("    paths:");
  assert.ok(
    pathsStart >= 0,
    "deploy-staging.yml push trigger must define paths",
  );

  return [...pushBlock.slice(pathsStart).matchAll(/^      - (.+)$/gm)].map(
    ([, value]) => {
      const trimmed = value.trim();
      const quote = trimmed.at(0);
      if ((quote === '"' || quote === "'") && trimmed.at(-1) === quote) {
        return trimmed.slice(1, -1);
      }
      return trimmed;
    },
  );
}

function pathSelectsPushDeploy(changedPath) {
  return extractPushPaths(readWorkflow()).some((glob) => {
    if (glob === "*") {
      return !changedPath.includes("/");
    }
    if (glob.endsWith("/**")) {
      return changedPath.startsWith(glob.slice(0, -2));
    }
    assert.ok(
      !glob.includes("*"),
      `test matcher must explicitly support workflow glob ${glob}`,
    );
    return changedPath === glob;
  });
}

test("VOC-111-TEST-00: evidence references issue #920 runs and missing path filter", () => {
  const evidence = readFileSync(evidencePath, "utf8");

  assert.match(evidence, /32568473144/);
  assert.match(evidence, /32568622178/);
  assert.match(evidence, /32572863842/);
  assert.match(evidence, /86df6779/);
  assert.match(evidence, /60822aa5/);
  assert.match(evidence, /#917/);
  assert.match(evidence, /no push path filter/i);
  assert.doesNotMatch(evidence, /Bearer\s+[A-Za-z0-9._-]+/);
});

test("VOC-111-TEST-01/02: docs, specs, and evidence-only paths do not select push deploy", () => {
  for (const changedPath of [
    "docs/operations/11-devops-and-ci-cd.md",
    "docs/governance/README.md",
    "specs/changes/VOC-111-skip-full-staging-deploys-for-documentation-and/t00-evidence.md",
    "specs/changes/VOC-110-operational-failure-deploy-staging-failure/specification.md",
    ".karsift/lessons.md",
    ".github/workflows/pipeline.yml",
    "scripts/governance/validate-governance.sh",
  ]) {
    assert.equal(
      pathSelectsPushDeploy(changedPath),
      false,
      `${changedPath} must not select push-triggered deploy-staging`,
    );
  }
});

test("VOC-111-TEST-03: application and shared-package changes remain selected", () => {
  for (const changedPath of [
    "apps/api/cmd/server/main.go",
    "apps/web/src/app/page.tsx",
    "apps/web/package.json",
    "packages/api-client/src/index.ts",
    "packages/design-tokens/package.json",
    "packages/eslint-config/index.js",
  ]) {
    assert.equal(
      pathSelectsPushDeploy(changedPath),
      true,
      `${changedPath} must select push-triggered deploy-staging`,
    );
  }
});

test("VOC-111-TEST-04: infra/deploy assets and staging e2e changes remain selected", () => {
  for (const changedPath of [
    "infra/docker-compose.yml",
    "infra/nginx/staging.conf",
    "infra/scripts/validate-staging-signup-allowlist.sh",
    "apps/web/tests/staging-e2e/core-loop.staging.spec.ts",
    "tests/staging-e2e/core-loop.staging.spec.ts",
    "tests/staging-e2e/playwright.config.ts",
  ]) {
    assert.equal(
      pathSelectsPushDeploy(changedPath),
      true,
      `${changedPath} must select push-triggered deploy-staging`,
    );
  }
});

test("VOC-111-TEST-05: every repository-root file remains selected", () => {
  for (const changedPath of [
    "package.json",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
    ".dockerignore",
    "eslint.config.js",
    "future-root-build-input.toml",
  ]) {
    assert.equal(
      pathSelectsPushDeploy(changedPath),
      true,
      `${changedPath} must select push-triggered deploy-staging`,
    );
  }

  for (const changedPath of [
    "docs/README.md",
    "specs/changes/VOC-111/example.md",
  ]) {
    assert.equal(
      pathSelectsPushDeploy(changedPath),
      false,
      `${changedPath} must not match the root-only selector`,
    );
  }
});

test("VOC-111-TEST-06: deploy workflow and selector test edits remain selected", () => {
  for (const changedPath of [
    ".github/workflows/deploy-staging.yml",
    "scripts/foundation/voc111-deploy-staging-paths.test.mjs",
  ]) {
    assert.equal(
      pathSelectsPushDeploy(changedPath),
      true,
      `${changedPath} must select push-triggered deploy-staging`,
    );
  }
});

test("VOC-111-TEST-08: workflow_dispatch and selected-push deploy semantics preserved", () => {
  const workflow = readWorkflow();
  const concurrencyStart = workflow.indexOf("concurrency:");
  const jobsStart = workflow.indexOf("jobs:", concurrencyStart);
  const concurrencyBlock = workflow.slice(concurrencyStart, jobsStart);

  assert.match(workflow, /workflow_dispatch:/);
  assert.match(workflow, /skip_ssh_deploy:/);
  assert.match(concurrencyBlock, /group: staging-deploy/);
  assert.match(concurrencyBlock, /queue: max/);
  assert.match(workflow, /docker\/build-push-action/);
  assert.match(workflow, /STAGING_SSH_HOST/);
  assert.match(workflow, /--config playwright\.staging\.config\.ts/);

  const dispatchBlock = workflow.slice(
    workflow.indexOf("workflow_dispatch:"),
    workflow.indexOf("\npermissions:"),
  );
  assert.doesNotMatch(dispatchBlock, /paths:/);
});

test("VOC-111-TEST-09: stale near-no-op documentation removed", () => {
  const workflow = readWorkflow();
  const devopsDoc = readFileSync(devopsDocPath, "utf8");

  assert.doesNotMatch(workflow, /near-no-op/i);
  assert.doesNotMatch(workflow, /every layer is cached/i);
  assert.match(workflow, /do \*\*not\*\* schedule this workflow on push/i);
  assert.match(devopsDoc, /runtime\/deploy allowlist|allowlisted runtime/i);
  assert.doesNotMatch(devopsDoc, /near-no-op/i);
});

test("VOC-129-TEST-06: empty diff and specs-only paths do not select push deploy", () => {
  const changedPaths = [];
  assert.equal(
    changedPaths.some(pathSelectsPushDeploy),
    false,
    "empty changed-path set must not select deploy",
  );
  for (const changedPath of [
    "specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the/specification.md",
    "specs/changes/VOC-127-converge-develop-to-the-exact-promotion-merge-sha/tasks.md",
  ]) {
    assert.equal(
      pathSelectsPushDeploy(changedPath),
      false,
      `${changedPath} must not select push-triggered deploy-staging`,
    );
  }
});

test("VOC-111 push allowlist: workflow paths block matches VOC-111-D03", () => {
  assert.deepEqual(
    extractPushPaths(readWorkflow()),
    EXPECTED_PUSH_PATHS,
    "push.paths must be the exact closed VOC-111-D03 allowlist",
  );
});
