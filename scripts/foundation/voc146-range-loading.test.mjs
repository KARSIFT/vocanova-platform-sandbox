// VOC-146-T00 — fail-closed --base/--head range loading in governance scripts.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  mkdtempSync,
  mkdirSync,
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

const implementationPrBase = "6f1e72206d04dbf7327a9194661ae1a4a806572e";
const nonexistentBase = "376e00dd769afb0fe850052b3a5cb48f729e73ad";
const nonexistentHead = "376e00dd769afb0fe850052b3a5cb48f729e73ad";

const validateGovernance = path.join(
  repositoryRoot,
  "scripts/governance/validate-governance.sh",
);
const validateMonitoringImpact = path.join(
  repositoryRoot,
  "scripts/governance/validate-monitoring-impact.sh",
);
const classifyChangeRisk = path.join(
  repositoryRoot,
  "scripts/governance/classify-change-risk.sh",
);
const loadChangedFiles = path.join(
  repositoryRoot,
  "scripts/governance/load-changed-files.sh",
);

function runScript(script, args = [], options = {}) {
  const { env: optionEnv = {}, ...rest } = options;
  const env = { ...process.env, ...optionEnv };
  delete env.GITHUB_EVENT_PATH;
  if (!("GITHUB_EVENT_NAME" in optionEnv)) {
    delete env.GITHUB_EVENT_NAME;
  }
  return spawnSync("bash", [script, ...args], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env,
    ...rest,
  });
}

function runShell(command, options = {}) {
  const isolatedEnv = { ...process.env };
  delete isolatedEnv.GITHUB_EVENT_PATH;
  delete isolatedEnv.GITHUB_EVENT_NAME;
  return spawnSync("bash", ["-lc", command], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: isolatedEnv,
    ...options,
  });
}

function assertNoSuccessLine(output) {
  assert.doesNotMatch(
    output,
    /Governance structure validation passed\./,
    output,
  );
}

test("VOC-146-TEST-00: nonexistent --base fails governance validation (issue #1127 class)", () => {
  const result = runScript(validateGovernance, [
    "--base",
    nonexistentBase,
    "--head",
    implementationPrBase,
  ]);
  assert.notEqual(result.status, 0, result.stderr || result.stdout);
  assertNoSuccessLine(`${result.stdout}\n${result.stderr}`);

  const nested = runScript(validateMonitoringImpact, [
    "--base",
    nonexistentBase,
    "--head",
    implementationPrBase,
  ]);
  assert.notEqual(nested.status, 0, nested.stderr || nested.stdout);
});

test("VOC-146-TEST-01: nonexistent --head fails governance validation", () => {
  const result = runScript(validateGovernance, [
    "--base",
    implementationPrBase,
    "--head",
    nonexistentHead,
  ]);
  assert.notEqual(result.status, 0, result.stderr || result.stdout);
  assertNoSuccessLine(`${result.stdout}\n${result.stderr}`);
});

test("VOC-146-TEST-02: unrelated histories with no merge base fail closed", () => {
  const repoDir = mkdtempSync(path.join(tmpdir(), "voc146-unrelated-"));

  const init = runShell(`git init -q "${repoDir}"`);
  assert.equal(init.status, 0, init.stderr || init.stdout);

  writeFileSync(path.join(repoDir, "a.txt"), "a\n");
  const first = runShell(
    `cd "${repoDir}" && git add a.txt && git -c user.email=t@e.com -c user.name=t commit -q -m a`,
  );
  assert.equal(first.status, 0, first.stderr || first.stdout);
  const baseSha = runShell(
    `cd "${repoDir}" && git rev-parse HEAD`,
  ).stdout.trim();

  writeFileSync(path.join(repoDir, "b.txt"), "b\n");
  const orphan = runShell(
    `cd "${repoDir}" && git checkout --orphan orphan -q && git rm -rf --cached . >/dev/null 2>&1; git add b.txt && git -c user.email=t@e.com -c user.name=t commit -q -m b`,
  );
  assert.equal(orphan.status, 0, orphan.stderr || orphan.stdout);
  const headSha = runShell(
    `cd "${repoDir}" && git rev-parse HEAD`,
  ).stdout.trim();

  const nested = runScript(
    validateMonitoringImpact,
    ["--base", baseSha, "--head", headSha],
    { cwd: repoDir },
  );
  assert.notEqual(nested.status, 0, nested.stderr || nested.stdout);

  const classifier = runScript(
    classifyChangeRisk,
    ["--base", baseSha, "--head", headSha],
    { cwd: repoDir },
  );
  assert.notEqual(classifier.status, 0, classifier.stderr || classifier.stdout);
  assert.doesNotMatch(
    `${classifier.stdout}\n${classifier.stderr}`,
    /No changed files to classify\./,
  );

  rmSync(repoDir, { recursive: true, force: true });
});

test("VOC-146-TEST-03: range loading does not use mapfile process substitution", () => {
  const monitoring = readFileSync(validateMonitoringImpact, "utf8");
  const classifier = readFileSync(classifyChangeRisk, "utf8");
  const helper = readFileSync(loadChangedFiles, "utf8");

  assert.doesNotMatch(monitoring, /mapfile -t files < <\(git diff/);
  assert.doesNotMatch(classifier, /mapfile -t files < <\(git diff/);
  assert.match(
    helper,
    /git diff --no-renames --name-only --diff-filter=ACDMRTUXB/,
  );
  assert.match(helper, /resolve_governance_commit/);
});

test("VOC-146-TEST-04: partial --base or --head fails closed", () => {
  const partialBase = runScript(validateMonitoringImpact, [
    "--base",
    implementationPrBase,
  ]);
  assert.notEqual(
    partialBase.status,
    0,
    partialBase.stderr || partialBase.stdout,
  );
  assert.match(
    partialBase.stderr,
    /requires both --base and --head/,
    partialBase.stderr,
  );

  const partialHead = runScript(validateMonitoringImpact, [
    "--head",
    implementationPrBase,
  ]);
  assert.notEqual(
    partialHead.status,
    0,
    partialHead.stderr || partialHead.stdout,
  );

  const classifierPartial = runScript(classifyChangeRisk, [
    "--base",
    implementationPrBase,
  ]);
  assert.notEqual(
    classifierPartial.status,
    0,
    classifierPartial.stderr || classifierPartial.stdout,
  );
});

test("VOC-146-TEST-05: valid range and --files-from still succeed", () => {
  const headSha = runShell("git rev-parse HEAD").stdout.trim();

  const validRange = runScript(validateMonitoringImpact, [
    "--base",
    implementationPrBase,
    "--head",
    headSha,
    "--declarations-only",
  ]);
  assert.equal(validRange.status, 0, validRange.stderr || validRange.stdout);

  const filesFromDir = mkdtempSync(path.join(tmpdir(), "voc146-files-from-"));
  const filesFrom = path.join(filesFromDir, "changed.txt");
  writeFileSync(filesFrom, "docs/development.md\n");
  const withFilesFrom = runScript(
    validateMonitoringImpact,
    ["--files-from", filesFrom, "--declarations-only"],
    {
      env: {
        ...process.env,
        GITHUB_EVENT_NAME: "pull_request",
      },
    },
  );
  assert.equal(
    withFilesFrom.status,
    0,
    withFilesFrom.stderr || withFilesFrom.stdout,
  );
  rmSync(filesFromDir, { recursive: true, force: true });

  const declarationsOnly = runScript(
    validateMonitoringImpact,
    ["--declarations-only"],
    {
      env: {
        GITHUB_EVENT_NAME: "pull_request",
      },
    },
  );
  assert.equal(
    declarationsOnly.status,
    0,
    declarationsOnly.stderr || declarationsOnly.stdout,
  );
});

test("VOC-146-TEST-06: classify-change-risk.sh uses the same fail-closed range contract", () => {
  const invalidBase = runScript(classifyChangeRisk, [
    "--base",
    nonexistentBase,
    "--head",
    implementationPrBase,
  ]);
  assert.notEqual(
    invalidBase.status,
    0,
    invalidBase.stderr || invalidBase.stdout,
  );
  assert.doesNotMatch(
    `${invalidBase.stdout}\n${invalidBase.stderr}`,
    /No changed files to classify\./,
  );

  const headSha = runShell("git rev-parse HEAD").stdout.trim();
  const validEmptyRange = runScript(classifyChangeRisk, [
    "--base",
    implementationPrBase,
    "--head",
    headSha,
  ]);
  assert.equal(
    validEmptyRange.status,
    0,
    validEmptyRange.stderr || validEmptyRange.stdout,
  );
});

test("VOC-146-TEST-07: AGENTS.md documents unresolved commits and invalid diff ranges as fail-closed", () => {
  const agents = readFileSync(path.join(repositoryRoot, "AGENTS.md"), "utf8");
  assert.match(agents, /unresolved commit/);
  assert.match(agents, /invalid diff range|three-dot/);
  assert.match(agents, /Partial/);
  assert.match(agents, /fails closed instead of/);
});
