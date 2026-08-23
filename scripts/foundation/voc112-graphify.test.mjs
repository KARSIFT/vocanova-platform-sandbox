// VOC-112-T03 — Graphify pilot opt-in configuration, lock identity, and fail-closed runner.
//
// Runs via `node --test scripts/foundation/voc112-graphify.test.mjs`.

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import {
  cpSync,
  chmodSync,
  mkdirSync,
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
const runtimeManifestPath = path.join(
  repositoryRoot,
  ".agents/skills/graphify-pilot/RUNTIME-MANIFEST.yaml",
);

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
  /pip\s+install\s+--upgrade/i,
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

function runCheck(fixtureHome) {
  return spawnSync("bash", [path.join(fixtureHome, "check")], {
    cwd: path.resolve(fixtureHome, "../.."),
    env: process.env,
    encoding: "utf8",
  });
}

function copyGraphifyFixtureWithoutVenv() {
  const fixtureRepository = mkdtempSync(
    path.join(tmpdir(), "graphify-fixture-"),
  );
  const fixtureHome = path.join(fixtureRepository, "scripts/graphify");
  mkdirSync(path.dirname(fixtureHome), { recursive: true });
  cpSync(graphifyHome, fixtureHome, { recursive: true });
  rmSync(path.join(fixtureHome, ".venv"), { recursive: true, force: true });
  writeFileSync(path.join(fixtureRepository, ".graphifyignore"), ".env\n");
  return { fixtureRepository, fixtureHome };
}

function lockedDistributions() {
  return [...readText(requirementsLockPath).matchAll(/^([\w.-]+==[^\s\\]+)/gm)]
    .map((match) => match[1].toLowerCase().replace(/[_.]+/g, "-"))
    .sort();
}

function parseRuntimeManifest(source) {
  return [
    ...source.matchAll(/  - path: ([^\n]+)\n    sha256: ([a-f0-9]{64})/g),
  ].map((match) => ({ path: match[1], sha256: match[2] }));
}

function addValidRuntimeFixture(fixtureHome) {
  const binDirectory = path.join(fixtureHome, ".venv/bin");
  mkdirSync(binDirectory, { recursive: true });

  const pythonPath = path.join(binDirectory, "python");
  const inventory = lockedDistributions()
    .map((distribution) => `printf '%s\\n' '${distribution}'`)
    .join("\n");
  writeFileSync(
    pythonPath,
    `#!/usr/bin/env bash
set -euo pipefail
if [[ "$*" == "-m pip list --format=freeze" ]]; then
${inventory}
  exit 0
fi
if [[ "$*" == "-m pip check" ]]; then
  exit 0
fi
exit 2
`,
  );
  chmodSync(pythonPath, 0o755);

  const cliPath = path.join(binDirectory, "graphify");
  writeFileSync(
    cliPath,
    `#!/usr/bin/env bash
set -euo pipefail
if [[ "\${1:-}" == "--version" ]]; then
  echo "graphify 0.9.48"
  exit 0
fi
{
  printf 'ARG=<%s>\\n' "$@"
  env | LC_ALL=C sort
} > "$HOME/invocation.txt"
`,
  );
  chmodSync(cliPath, 0o755);
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
  assert.equal(identity.python_min, "3.12");
  assert.equal(identity.locked_distribution_count, "30");
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
  const lockSource = readText(requirementsLockPath);
  assert.match(lockSource, /graphifyy==0\.9\.48/);
  assert.equal(lockedDistributions().length, 30);
  const pinStarts = [...lockSource.matchAll(/^[\w.-]+==/gm)].map(
    (match) => match.index,
  );
  for (const [position, start] of pinStarts.entries()) {
    const packageBlock = lockSource.slice(
      start,
      pinStarts[position + 1] ?? lockSource.length,
    );
    assert.match(packageBlock, /--hash=sha256:[a-f0-9]{64}/);
  }

  const runScript = readText(path.join(graphifyHome, "run.sh"));
  assert.match(runScript, /--code-only/);
  assert.match(runScript, /GRAPHIFY_QUERY_LOG_DISABLE=1/);
  assert.match(runScript, /exec env -i/);
  assert.match(runScript, /target must remain inside the repository/);
  assert.match(
    runScript,
    /HOME="\$RUNTIME_HOME"/,
    "runner must isolate the user profile",
  );

  const checkScript = readText(path.join(graphifyHome, "check"));
  assert.doesNotMatch(
    checkScript,
    /pip\s+install|uv\s+tool|pipx\s+install|curl\s+|wget\s+/i,
    "check must not download packages",
  );
  assert.match(checkScript, /pip list --format=freeze/);
  assert.match(checkScript, /pip check/);

  const setupScript = readText(path.join(graphifyHome, "setup.sh"));
  assert.match(setupScript, /--require-hashes/);
  assert.match(setupScript, /--only-binary=:all:/);
  assert.doesNotMatch(setupScript, /pip install --upgrade/);

  for (const source of [runScript, checkScript, setupScript]) {
    assert.doesNotMatch(
      source,
      /GRAPHIFY_HOME="\$\{GRAPHIFY_HOME/,
      "runtime location must not be environment-overridable",
    );
  }

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

  const provenance = readText(
    path.join(repositoryRoot, ".agents/skills/graphify-pilot/PROVENANCE.yaml"),
  );
  assert.match(provenance, /LICENSE-MIT/);
  assert.ok(
    statSync(
      path.join(repositoryRoot, ".agents/skills/graphify-pilot/LICENSE-MIT"),
      { throwIfNoEntry: false },
    ),
    "NOTICE references LICENSE-MIT; file must be committed",
  );

  const gitignore = readText(path.join(repositoryRoot, ".gitignore"));
  assert.match(gitignore, /graphify-out\//);
  assert.match(gitignore, /scripts\/graphify\/\.venv\//);
  assert.match(gitignore, /scripts\/graphify\/\.runtime-home\//);

  const treeErrors = validateAgentSkillsTree();
  assert.deepEqual(treeErrors, [], treeErrors.join("\n"));
});

test("VOC-112-TEST-10 supplemental: provenance-covered runtime manifest is exact", () => {
  const entries = parseRuntimeManifest(readText(runtimeManifestPath));
  assert.deepEqual(
    entries.map((entry) => entry.path),
    [
      "scripts/graphify/setup.sh",
      "scripts/graphify/check",
      "scripts/graphify/run.sh",
      "scripts/graphify/runtime-identity.yaml",
      "scripts/graphify/requirements.in",
      "scripts/graphify/requirements.lock",
      ".graphifyignore",
    ],
  );

  for (const entry of entries) {
    const resolved = path.resolve(repositoryRoot, entry.path);
    assert.ok(
      resolved.startsWith(`${repositoryRoot}${path.sep}`),
      `runtime manifest path escapes repository: ${entry.path}`,
    );
    assert.equal(
      sha256Hex(readFileSync(resolved)),
      entry.sha256,
      `runtime manifest digest mismatch: ${entry.path}`,
    );
  }
});

test("VOC-112-TEST-11: check fails closed without locked runtime", () => {
  const { fixtureRepository, fixtureHome } = copyGraphifyFixtureWithoutVenv();
  try {
    const result = runCheck(fixtureHome);
    assert.notEqual(result.status, 0, "check must fail when .venv is absent");
    assert.match(
      result.stderr,
      /locked runtime missing|run: bash scripts\/graphify\/setup\.sh/i,
    );
  } finally {
    rmSync(fixtureRepository, { recursive: true, force: true });
  }
});

test("VOC-112-TEST-11 supplemental: check passes with valid locked runtime", () => {
  const { fixtureRepository, fixtureHome } = copyGraphifyFixtureWithoutVenv();
  try {
    addValidRuntimeFixture(fixtureHome);
    const result = runCheck(fixtureHome);
    assert.equal(
      result.status,
      0,
      `expected check to pass with locked runtime: ${result.stderr}`,
    );
    assert.match(result.stdout, /locked runtime identity OK/);
    assert.match(result.stdout, /30 distributions/);
  } finally {
    rmSync(fixtureRepository, { recursive: true, force: true });
  }
});

test("VOC-112-TEST-10 supplemental: identity mismatch fails closed", () => {
  const { fixtureRepository, fixtureHome } = copyGraphifyFixtureWithoutVenv();
  try {
    const tamperedLock = path.join(fixtureHome, "requirements.lock");
    writeFileSync(tamperedLock, "# tampered\n", "utf8");
    const result = runCheck(fixtureHome);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /digest mismatch/i);
  } finally {
    rmSync(fixtureRepository, { recursive: true, force: true });
  }
});

test("VOC-112-TEST-11 supplemental: runner is hermetic and repository-bounded", () => {
  const { fixtureRepository, fixtureHome } = copyGraphifyFixtureWithoutVenv();
  try {
    addValidRuntimeFixture(fixtureHome);
    const target = path.join(fixtureRepository, "apps/api");
    mkdirSync(target, { recursive: true });

    const result = spawnSync(
      "bash",
      [path.join(fixtureHome, "run.sh"), target],
      {
        cwd: fixtureRepository,
        env: {
          ...process.env,
          OPENAI_API_KEY: "must-not-reach-child",
          AWS_ACCESS_KEY_ID: "must-not-reach-child",
          VOC112_SECRET_SENTINEL: "must-not-reach-child",
        },
        encoding: "utf8",
      },
    );
    assert.equal(result.status, 0, result.stderr);

    const invocation = readText(
      path.join(fixtureHome, ".runtime-home/invocation.txt"),
    );
    assert.match(invocation, /ARG=<extract>/);
    assert.match(invocation, /ARG=<--code-only>/);
    assert.match(
      invocation,
      new RegExp(`ARG=<${target.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}>`),
    );
    assert.match(invocation, /GRAPHIFY_QUERY_LOG_DISABLE=1/);
    assert.match(invocation, /GRAPHIFY_QUERY_LOG=\/dev\/null/);
    assert.doesNotMatch(
      invocation,
      /OPENAI_API_KEY|AWS_ACCESS_KEY_ID|VOC112_SECRET_SENTINEL/,
    );

    const escaped = spawnSync(
      "bash",
      [path.join(fixtureHome, "run.sh"), tmpdir()],
      { cwd: fixtureRepository, env: process.env, encoding: "utf8" },
    );
    assert.notEqual(escaped.status, 0);
    assert.match(escaped.stderr, /target must remain inside the repository/);
  } finally {
    rmSync(fixtureRepository, { recursive: true, force: true });
  }
});
