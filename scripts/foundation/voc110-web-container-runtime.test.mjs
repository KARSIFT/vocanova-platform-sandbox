// VOC-110-T00 — production web Docker boot/HTTP merge gate contract tests.
//
// Runs via `node --test scripts/foundation/voc110-web-container-runtime.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
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
const webPackagePath = path.join(repositoryRoot, "apps/web/package.json");
const evidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-110-operational-failure-deploy-staging-failure/t00-evidence.md",
);
const devWorkflowPath = path.join(
  repositoryRoot,
  "docs/operations/10-development-workflow.md",
);

const RELEVANT_PATH_PATTERN =
  /git diff --name-only[\s\S]*grep -Eq[\s\S]*\^\(package\\\.json\|pnpm-lock\\\.yaml\|pnpm-workspace\\\.yaml\|apps\/web\/\|packages\/\)/;

const DENYLIST_PATTERNS = [
  /continue-on-error:\s*true/,
  /secrets\./,
  /docker push/,
  /build-push-action/,
];

function readPipeline() {
  return readFileSync(pipelinePath, "utf8");
}

function extractJobBlock(source, jobKey) {
  const marker = `  ${jobKey}:`;
  const start = source.indexOf(marker);
  assert.ok(start >= 0, `pipeline.yml must define job ${jobKey}`);
  const remainder = source.slice(start + marker.length);
  const nextJobMatch = remainder.match(/\n  [a-z][a-z0-9-]*:/);
  const end =
    nextJobMatch && typeof nextJobMatch.index === "number"
      ? start + marker.length + nextJobMatch.index
      : source.length;
  return source.slice(start, end);
}

test("VOC-110-TEST-00: evidence records run 32566405628 and failing web health poll", () => {
  const evidence = readFileSync(evidencePath, "utf8");

  assert.match(evidence, /32566405628/);
  assert.match(evidence, /f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46/);
  assert.match(evidence, /Poll staging\.vocanova\.site\//);
  assert.match(evidence, /502/);
  assert.match(evidence, /@swc\/helpers/);
  assert.match(evidence, /16\.3\.1/);
  assert.match(evidence, /#859/);
  assert.doesNotMatch(evidence, /Bearer\s+[A-Za-z0-9._-]+/);
  assert.doesNotMatch(evidence, /vocanova_session=/);
});

test("VOC-110-TEST-01: paired Next.js packages are stable 16.3.2", () => {
  const webPackage = JSON.parse(readFileSync(webPackagePath, "utf8"));

  assert.equal(webPackage.dependencies.next, "16.3.2");
  assert.equal(
    webPackage.devDependencies["@next/eslint-plugin-next"],
    "16.3.2",
  );
});

test("VOC-110-TEST-03: pipeline defines path-aware web-container-runtime job", () => {
  const pipeline = readPipeline();
  const jobBlock = extractJobBlock(pipeline, "web-container-runtime");

  assert.match(jobBlock, /name: web container runtime/);
  assert.match(jobBlock, RELEVANT_PATH_PATTERN);
  assert.match(
    jobBlock,
    /Skip — no web runtime surface changed/,
    "irrelevant diffs must take the cheap no-op path",
  );
  assert.match(jobBlock, /docker build -f apps\/web\/Dockerfile/);
  assert.match(jobBlock, /docker run -d/);
  assert.match(
    jobBlock,
    /curl --silent --output \/dev\/null --write-out '%\{http_code\}'/,
  );
  assert.match(jobBlock, /docker inspect --format '\{\{\.State\.Running\}\}'/);
  assert.match(jobBlock, /trap cleanup EXIT/);
  assert.match(jobBlock, /docker rm -f/);
  assert.match(jobBlock, /docker rmi -f/);

  for (const pattern of DENYLIST_PATTERNS) {
    assert.doesNotMatch(
      jobBlock,
      pattern,
      `web-container-runtime must not contain bypass pattern ${pattern}`,
    );
  }
});

test("VOC-110-TEST-03: merge-gate waits on web-container-runtime", () => {
  const pipeline = readPipeline();
  const mergeGateBlock = extractJobBlock(pipeline, "merge-gate");

  assert.match(mergeGateBlock, /needs:[\s\S]*web-container-runtime/);
});

test("VOC-110-TEST-04: development workflow documents the shipped-artifact gate", () => {
  const doc = readFileSync(devWorkflowPath, "utf8");

  assert.match(doc, /web container runtime/i);
  assert.match(doc, /apps\/web\/Dockerfile/);
  assert.match(doc, /HTTP 2xx/);
});
