// VOC-112-T03 — Graphify pilot opt-in configuration, lock identity, and fail-closed runner.
//
// Runs via `node --test scripts/foundation/voc112-graphify.test.mjs`.

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import {
  cpSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import test from "node:test";
import { validateAgentSkillsTree } from "./voc112-agent-skills.test.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const graphifyHome = path.join(repositoryRoot, "scripts/graphify");
const skillPath = path.join(
  repositoryRoot,
  ".agents/skills/graphify-pilot/SKILL.md",
);
const graphifyIgnorePath = path.join(repositoryRoot, ".graphifyignore");
const identityPath = path.join(graphifyHome, "runtime-identity.yaml");
const requirementsLockPath = path.join(graphifyHome, "requirements.lock");

const REQUIRED_IGNORE_PATTERNS = [
  ".env",
  "node_modules/",
  "graphify-out/",
  "scripts/graphify/.venv/",
  "dist/",
  ".next/",
];

const FORBIDDEN_RUNNER_MARKERS = [
  /graphify\s+install\b/i,
  /hook\s+install/i,
  /uv\s+tool\s+install/i,
  /pipx\s+install/i,
  /@latest\b/,
  /npm\s+install\s+-g/i,
];

function sha256Hex(content) {
  return createHash("sha256").update(content).digest("hex");
}

function readText(filePath) {
  return readFileSync(filePath, "utf8");
}

function parseFrontmatter(source) {
  const match = source.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n?([\s\S]*)$/);
  assert.ok(match, "graphify-pilot SKILL.md must include YAML frontmatter");
  const frontmatter = {};
  for (const line of match[1].split(/\r?\n/)) {
    const separator = line.indexOf(":");
    if (separator < 0) {
      continue;
    }
    const key = line.slice(0, separator).trim();
    const value = line.slice(separator + 1).trim();
    frontmatter[key] = value;
  }
  return { frontmatter, body: match[2] };
}

function parseIdentityYaml(source) {
  const record = {};
  for (const line of source.split(/\r?\n/)) {
    if (!line.trim() || line.trim().startsWith("#")) {
      continue;
    }
    const separator = line.indexOf(":");
    if (separator < 0) {
      continue;
    }
    const key = line.slice(0, separator).trim();
    const value = line
      .slice(separator + 1)
      .trim()
      .replace(/^["']|["']$/g, "");
    record[key] = value;
  }
  return record;
}

function runCheck(graphifyHomeOverride) {
  return spawnSync("bash", [path.join(graphifyHome, "check")], {
    cwd: repositoryRoot,
    env: {
      ...process.env,
      GRAPHIFY_HOME: graphifyHomeOverride,
    },
    encoding: "utf8",
  });
}

function copyGraphifyFixtureWithoutVenv() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "graphify-fixture-"));
  cpSync(graphifyHome, fixtureRoot, { recursive: true });
  rmSync(path.join(fixtureRoot, ".venv"), { recursive: true, force: true });
  return fixtureRoot;
}

test("VOC-112-TEST-10: graphify pilot is opt-in, code-only, and hash-locked", () => {
  const skillSource = readText(skillPath);
  const { frontmatter, body } = parseFrontmatter(skillSource);

  assert.equal(frontmatter.name, "graphify-pilot");
  assert.equal(frontmatter["disable-model-invocation"], "true");

  const ignoreSource = readText(graphifyIgnorePath);
  for (const pattern of REQUIRED_IGNORE_PATTERNS) {
    assert.match(
      ignoreSource,
      new RegExp(pattern.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")),
      `missing .graphifyignore pattern: ${pattern}`,
    );
  }

  const identity = parseIdentityYaml(readText(identityPath));
  assert.equal(identity.pypi_package, "graphifyy");
  assert.equal(identity.pypi_version, "0.9.48");
  assert.equal(
    identity.upstream_commit,
    "b2cd36267456c166788c95be6e68574064a92a42",
  );
  assert.match(identity.upstream_repo, /Graphify-Labs\/graphify/);

  const requirementsIn = readText(path.join(graphifyHome, "requirements.in"));
  assert.equal(
    sha256Hex(requirementsIn),
    identity.requirements_in_sha256,
    "requirements.in digest must match runtime identity",
  );
  assert.equal(
    sha256Hex(readText(requirementsLockPath)),
    identity.requirements_lock_sha256,
    "requirements.lock digest must match runtime identity",
  );
  assert.match(readText(requirementsLockPath), /graphifyy==0\.9\.48/);

  const runScript = readText(path.join(graphifyHome, "run.sh"));
  assert.match(runScript, /--code-only/);
  assert.match(runScript, /GRAPHIFY_QUERY_LOG_DISABLE=1/);
  assert.match(
    runScript,
    /unset OPENAI_API_KEY/,
    "runner must strip provider credentials",
  );

  const checkScript = readText(path.join(graphifyHome, "check"));
  assert.doesNotMatch(
    checkScript,
    /pip\s+install|uv\s+tool|pipx\s+install|curl\s+|wget\s+/i,
    "check must not download packages",
  );

  for (const marker of FORBIDDEN_RUNNER_MARKERS) {
    assert.doesNotMatch(
      runScript,
      marker,
      `run.sh must not contain: ${marker}`,
    );
    assert.doesNotMatch(
      checkScript,
      marker,
      `check must not contain: ${marker}`,
    );
    assert.doesNotMatch(
      readText(path.join(graphifyHome, "setup.sh")),
      /graphify\s+install|hook\s+install/i,
      `setup.sh must not invoke graphify install/hooks`,
    );
  }

  assert.match(body, /hint/i);
  assert.match(body, /verify.*current source/i);
  assert.match(body, /Automatic invocation is disabled/i);
  assert.match(
    body,
    /graphify install/i,
    "must warn against upstream install flows",
  );

  const gitignore = readText(path.join(repositoryRoot, ".gitignore"));
  assert.match(gitignore, /graphify-out\//);
  assert.match(gitignore, /scripts\/graphify\/\.venv\//);

  const treeErrors = validateAgentSkillsTree();
  assert.deepEqual(treeErrors, [], treeErrors.join("\n"));
});

test("VOC-112-TEST-11: check fails closed without locked runtime", () => {
  const fixtureRoot = copyGraphifyFixtureWithoutVenv();
  try {
    const result = runCheck(fixtureRoot);
    assert.notEqual(result.status, 0, "check must fail when .venv is absent");
    assert.match(
      result.stderr,
      /locked runtime missing|run: bash scripts\/graphify\/setup\.sh/i,
    );
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
});

test("VOC-112-TEST-11 supplemental: check passes with valid locked runtime", () => {
  const venvPython = path.join(graphifyHome, ".venv/bin/python");
  if (!statSync(venvPython, { throwIfNoEntry: false })) {
    return;
  }

  const result = runCheck(graphifyHome);
  assert.equal(
    result.status,
    0,
    `expected check to pass with locked runtime: ${result.stderr}`,
  );
  assert.match(result.stdout, /locked runtime identity OK/);
});

test("VOC-112-TEST-10 supplemental: identity mismatch fails closed", () => {
  const fixtureRoot = copyGraphifyFixtureWithoutVenv();
  try {
    const tamperedLock = path.join(fixtureRoot, "requirements.lock");
    writeFileSync(tamperedLock, "# tampered\n", "utf8");
    const result = runCheck(fixtureRoot);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /digest mismatch/i);
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
});
