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
const rootPackagePath = path.join(repositoryRoot, "package.json");
const apiClientPackagePath = path.join(
  repositoryRoot,
  "packages/api-client/package.json",
);
const eslintConfigPackagePath = path.join(
  repositoryRoot,
  "packages/eslint-config/package.json",
);
const lockfilePath = path.join(repositoryRoot, "pnpm-lock.yaml");
const evidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-110-operational-failure-deploy-staging-failure/t00-evidence.md",
);
const devWorkflowPath = path.join(
  repositoryRoot,
  "docs/operations/10-development-workflow.md",
);

const WORKFLOW_PATH_PATTERN = String.raw`'^([^/]+$|apps/web/|packages/|\.github/workflows/pipeline\.yml|scripts/foundation/voc110-web-container-runtime\.test\.mjs)'`;
const WEB_RUNTIME_PATH_PATTERN =
  /^([^/]+$|apps\/web\/|packages\/|\.github\/workflows\/pipeline\.yml|scripts\/foundation\/voc110-web-container-runtime\.test\.mjs)/;

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

test("VOC-110-TEST-01: Next.js repair preserves every PR #859 update", () => {
  const rootPackage = JSON.parse(readFileSync(rootPackagePath, "utf8"));
  const webPackage = JSON.parse(readFileSync(webPackagePath, "utf8"));
  const apiClientPackage = JSON.parse(
    readFileSync(apiClientPackagePath, "utf8"),
  );
  const eslintConfigPackage = JSON.parse(
    readFileSync(eslintConfigPackagePath, "utf8"),
  );
  const lockfile = readFileSync(lockfilePath, "utf8");

  assert.equal(rootPackage.devDependencies.eslint, "10.8.1");
  assert.equal(rootPackage.devDependencies["socket.io-client"], "4.8.3");
  assert.equal(webPackage.dependencies["@sentry/nextjs"], "^10.70.0");
  assert.equal(webPackage.dependencies.next, "16.3.2");
  assert.equal(webPackage.devDependencies["@axe-core/playwright"], "4.13.0");
  assert.equal(
    webPackage.devDependencies["@next/eslint-plugin-next"],
    "16.3.2",
  );
  assert.equal(webPackage.devDependencies["@types/node"], "26.2.0");
  assert.equal(apiClientPackage.devDependencies["@types/node"], "26.2.0");
  assert.equal(eslintConfigPackage.dependencies.eslint, "10.8.1");
  assert.equal(eslintConfigPackage.dependencies["typescript-eslint"], "8.67.0");
  assert.match(lockfile, /next@16\.3\.2/);
  assert.match(lockfile, /@next\/eslint-plugin-next@16\.3\.2/);
  assert.doesNotMatch(lockfile, /next@16\.3\.1/);
  assert.doesNotMatch(lockfile, /@next\/eslint-plugin-next@16\.3\.1/);
});

test("VOC-110-TEST-02: production web image smoke is exact-SHA and fail-closed", () => {
  const pipeline = readPipeline();
  const jobBlock = extractJobBlock(pipeline, "web-container-runtime");

  assert.match(jobBlock, /name: web container runtime/);
  assert.match(
    jobBlock,
    /ref: \$\{\{ github\.event\.pull_request\.head\.sha \}\}/,
    "Docker runtime smoke must build the reviewed PR head SHA",
  );
  assert.match(jobBlock, /docker build -f apps\/web\/Dockerfile/);
  assert.match(jobBlock, /docker run -d/);
  assert.match(jobBlock, /--add-host=host\.docker\.internal:host-gateway/);
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

test("VOC-110-TEST-03: pipeline defines a fail-closed path selector", () => {
  const pipeline = readPipeline();
  const jobBlock = extractJobBlock(pipeline, "web-container-runtime");

  assert.ok(
    jobBlock.includes(WORKFLOW_PATH_PATTERN),
    "workflow and fixture must use the same fail-closed path selector",
  );
  assert.match(
    jobBlock,
    /Skip — no web runtime surface changed/,
    "irrelevant diffs must take the cheap no-op path",
  );

  for (const changedPath of [
    "package.json",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
    ".dockerignore",
    "eslint.config.js",
    "apps/web/Dockerfile",
    "apps/web/src/app/page.tsx",
    "packages/api-client/src/index.ts",
    "packages/design-tokens/package.json",
    ".github/workflows/pipeline.yml",
    "scripts/foundation/voc110-web-container-runtime.test.mjs",
  ]) {
    assert.match(
      changedPath,
      WEB_RUNTIME_PATH_PATTERN,
      `${changedPath} must select the Docker runtime smoke`,
    );
  }

  for (const changedPath of [
    "apps/api/go.mod",
    "docs/operations/README.md",
    "specs/changes/VOC-110/example.md",
    ".github/workflows/scheduled-synthetics.yml",
  ]) {
    assert.doesNotMatch(
      changedPath,
      WEB_RUNTIME_PATH_PATTERN,
      `${changedPath} must take the cheap no-op path`,
    );
  }
});

test("VOC-110-TEST-03: merge-gate waits on web-container-runtime", () => {
  const pipeline = readPipeline();
  const mergeGateBlock = extractJobBlock(pipeline, "merge-gate");

  assert.match(mergeGateBlock, /needs:[\s\S]*web-container-runtime/);
  assert.match(
    mergeGateBlock,
    /needs\.web-container-runtime\.result == 'success'/,
    "merge-gate must fail closed when the web container runtime job fails",
  );
});

test("VOC-110-TEST-04: development workflow documents the shipped-artifact gate", () => {
  const doc = readFileSync(devWorkflowPath, "utf8");

  assert.match(doc, /web container runtime/i);
  assert.match(doc, /apps\/web\/Dockerfile/);
  assert.match(doc, /HTTP 2xx/);
});

test("VOC-110-TEST-05: deploy-staging regression suites have passing evidence", () => {
  const evidenceLines = readFileSync(evidencePath, "utf8").split("\n");

  for (const suite of [
    "voc084-deploy-staging-oauth.test.mjs",
    "voc088-deploy-staging-allowlist.test.mjs",
    "voc095-playwright-install.test.mjs",
  ]) {
    const resultLine = evidenceLines.find(
      (line) => line.includes(suite) && /\|\s*pass/.test(line),
    );
    assert.ok(resultLine, `${suite} must have an explicit passing result`);
  }
});
