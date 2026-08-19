// VOC-095-T00 — bounded Playwright Chromium install contract tests.
//
// Runs via `node --test scripts/foundation/voc095-playwright-install.test.mjs`.

import assert from "node:assert/strict";
import {
  chmodSync,
  existsSync,
  mkdtempSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const installScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/install-playwright-chromium.sh",
);
const evidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-095-harden-playwright-setup-against-hosted-runner-apt/t00-evidence.md",
);

const DENYLIST_PATTERNS = [
  /continue-on-error/i,
  /PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD/i,
  /PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS/i,
  /--ignore-deps/i,
  /install --with-deps/i,
  /\|\|\s*true/,
];

function readInstallScript() {
  assert.ok(existsSync(installScriptPath));
  return readFileSync(installScriptPath, "utf8");
}

function extractConstant(source, name) {
  const match = source.match(
    new RegExp(`readonly\\s+${name}\\s*=\\s*(\\d+)`, "m"),
  );
  assert.ok(match, `expected readonly constant ${name}`);
  return Number.parseInt(match[1], 10);
}

function runInstallScriptWithFixture({
  depsMode = "all-fail",
  createBrowserBinary = true,
  packageManagerBusy = false,
} = {}) {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "voc095-playwright-"));
  const binDir = path.join(fixtureRoot, "bin");
  mkdirSync(binDir, { recursive: true });

  const homeDir = path.join(fixtureRoot, "home");
  mkdirSync(homeDir, { recursive: true });
  const browserDir = path.join(
    homeDir,
    ".cache/ms-playwright/chromium-fixture/chrome-linux",
  );
  const mirrorListPath = path.join(fixtureRoot, "apt-mirrors.txt");
  writeFileSync(mirrorListPath, "http://hosted-runner.invalid/ubuntu/\n");

  writeFileSync(
    path.join(binDir, "timeout"),
    `#!/usr/bin/env bash
set -euo pipefail
while [[ "\${1:-}" == --* ]]; do shift; done
shift
exec "$@"
`,
  );

  writeFileSync(
    path.join(binDir, "sleep"),
    `#!/usr/bin/env bash
exit 0
`,
  );

  writeFileSync(
    path.join(binDir, "sudo"),
    `#!/usr/bin/env bash
set -euo pipefail
if [[ "\${1:-}" == "-n" ]]; then shift; fi
exec "$@"
`,
  );

  writeFileSync(
    path.join(binDir, "fuser"),
    `#!/usr/bin/env bash
if [[ "\${PACKAGE_MANAGER_BUSY}" == "1" ]]; then exit 0; fi
exit 1
`,
  );

  writeFileSync(
    path.join(binDir, "pnpm"),
    `#!/usr/bin/env bash
set -euo pipefail
cmd="$*"
counter_file="$FIXTURE_ROOT/deps-attempts.count"
if [[ "$cmd" == *"playwright install chromium"* ]]; then
  if [[ "$CREATE_BROWSER" == "1" ]]; then
    mkdir -p "$BROWSER_DIR"
    printf '' > "$BROWSER_DIR/chrome"
    chmod +x "$BROWSER_DIR/chrome"
  fi
  exit 0
fi
if [[ "$cmd" == *"playwright install-deps chromium"* ]]; then
  attempts=0
  if [[ -f "$counter_file" ]]; then
    attempts="$(cat "$counter_file")"
  fi
  attempts=$((attempts + 1))
  printf '%s' "$attempts" > "$counter_file"
  case "$DEPS_MODE" in
    immediate-success) exit 0 ;;
    primary-timeout-fallback-success)
      grep -Fxq "https://archive.ubuntu.com/ubuntu/" "$MIRROR_LIST_PATH" && exit 0
      exit 124
      ;;
    all-fail) exit 17 ;;
    *) exit 65 ;;
  esac
fi
echo "unexpected pnpm invocation: $cmd" >&2
exit 64
`,
  );

  chmodSync(path.join(binDir, "timeout"), 0o755);
  chmodSync(path.join(binDir, "sleep"), 0o755);
  chmodSync(path.join(binDir, "sudo"), 0o755);
  chmodSync(path.join(binDir, "fuser"), 0o755);
  chmodSync(path.join(binDir, "pnpm"), 0o755);

  const result = spawnSync("bash", [installScriptPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: {
      ...process.env,
      HOME: homeDir,
      PATH: `${binDir}:${process.env.PATH}`,
      FIXTURE_ROOT: fixtureRoot,
      CREATE_BROWSER: createBrowserBinary ? "1" : "0",
      BROWSER_DIR: browserDir,
      DEPS_MODE: depsMode,
      MIRROR_LIST_PATH: mirrorListPath,
      PLAYWRIGHT_APT_MIRROR_LIST_PATH: mirrorListPath,
      GITHUB_ACTIONS: "true",
      RUNNER_OS: "Linux",
      RUNNER_ARCH: "X64",
      PACKAGE_MANAGER_BUSY: packageManagerBusy ? "1" : "0",
    },
  });
  return { ...result, fixtureRoot, mirrorListPath };
}

test("VOC-095-TEST-00: root-cause evidence references run 32299315180 and apt stall phase", () => {
  const evidence = readFileSync(evidencePath, "utf8");

  assert.match(evidence, /32299315180/);
  assert.match(evidence, /fdde12daeccf521a5a8be171294eba79a138717b/);
  assert.match(evidence, /deploy-staging/i);
  assert.match(
    evidence,
    /playwright install|Install Playwright Chromium/i,
    "evidence must identify Playwright install as the timeout phase",
  );
  assert.match(
    evidence,
    /deploy convergence|Deploy to staging host|health/i,
    "evidence must record deploy convergence before Playwright timeout",
  );
  assert.doesNotMatch(evidence, /Bearer\s+[A-Za-z0-9._-]+/);
  assert.doesNotMatch(evidence, /vocanova_session=/);
});

test("VOC-095-TEST-01: install script declares timeout and retry constants", () => {
  const source = readInstallScript();

  assert.match(source, /^set -euo pipefail/m);
  assert.match(source, /pnpm --filter @vocanova\/web exec playwright/);

  const maxAttempts = extractConstant(source, "PLAYWRIGHT_DEPS_MAX_ATTEMPTS");
  const attemptTimeoutSeconds = extractConstant(
    source,
    "PLAYWRIGHT_DEPS_ATTEMPT_TIMEOUT_SECONDS",
  );
  const retrySleepSeconds = extractConstant(
    source,
    "PLAYWRIGHT_DEPS_RETRY_SLEEP_SECONDS",
  );

  assert.equal(maxAttempts, 3);
  assert.ok(
    attemptTimeoutSeconds >= 60 && attemptTimeoutSeconds <= 300,
    "per-attempt timeout must stay within documented AC-01 bounds",
  );
  assert.ok(
    retrySleepSeconds >= 5 && retrySleepSeconds <= 120,
    "retry sleep must stay within documented AC-01 bounds",
  );
  assert.match(
    source,
    /timeout\s+--signal=TERM[\s\S]*--kill-after="\$\{PLAYWRIGHT_DEPS_KILL_GRACE_SECONDS\}"/,
  );
  assert.match(source, /wait_for_package_manager_quiescence/);
  assert.match(source, /sudo -n fuser/);
});

test("VOC-095-TEST-02: install script verifies Chromium binary after install", () => {
  const source = readInstallScript();

  assert.match(
    source,
    /chromium-\*\/chrome-linux\/chrome/,
    "script must verify the Playwright Chromium binary path",
  );
  assert.match(source, /verify_chromium_binary/);
  assert.match(source, /verify_chromium_binary\s*$/m);
});

test("VOC-095-TEST-03: install script has no skip or bypass patterns", () => {
  const source = readInstallScript();

  for (const pattern of DENYLIST_PATTERNS) {
    assert.doesNotMatch(
      source,
      pattern,
      `install script must not contain bypass pattern ${pattern}`,
    );
  }

  assert.match(source, /install chromium/);
  assert.match(source, /install-deps chromium/);
});

test("VOC-095-TEST-03b: install script fails closed when deps retries exhaust", () => {
  const result = runInstallScriptWithFixture({ depsMode: "all-fail" });

  assert.notEqual(result.status, 0);
  assert.match(
    result.stderr,
    /exhausted 3 attempts|failed \(exit/i,
    "stderr must report bounded deps failure without secrets",
  );
});

test("VOC-095-TEST-03c: install script succeeds when deps succeed and binary exists", () => {
  const result = runInstallScriptWithFixture({ depsMode: "immediate-success" });

  assert.equal(result.status, 0, result.stderr || result.stdout);
  assert.match(result.stdout, /succeeded on attempt 1\/3/);
});

test("VOC-095-TEST-03d: install script fails closed when Chromium binary is missing", () => {
  const result = runInstallScriptWithFixture({
    depsMode: "immediate-success",
    createBrowserBinary: false,
  });

  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /Chromium binary not found/);
});

test("VOC-095-TEST-03e: primary timeout activates HTTPS fallback and succeeds", () => {
  const result = runInstallScriptWithFixture({
    depsMode: "primary-timeout-fallback-success",
  });

  assert.equal(result.status, 0, result.stderr || result.stdout);
  assert.match(
    result.stdout,
    /activated verified HTTPS Ubuntu archive fallback/,
  );
  assert.match(result.stdout, /succeeded on attempt 2\/3/);
  assert.equal(
    readFileSync(result.mirrorListPath, "utf8"),
    "http://hosted-runner.invalid/ubuntu/\n",
    "runner mirror list must be restored after the bounded fallback",
  );
});

test("VOC-095-TEST-03f: primary and HTTPS fallback failure remains non-zero", () => {
  const result = runInstallScriptWithFixture({ depsMode: "all-fail" });

  assert.notEqual(result.status, 0);
  assert.match(
    result.stdout,
    /activated verified HTTPS Ubuntu archive fallback/,
  );
  assert.match(result.stderr, /exhausted 3 attempts/);
});

test("VOC-095-TEST-03g: busy package-manager locks block a poisoned retry", () => {
  const result = runInstallScriptWithFixture({
    depsMode: "primary-timeout-fallback-success",
    packageManagerBusy: true,
  });

  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /package-manager locks remained busy/);
  assert.equal(
    readFileSync(path.join(result.fixtureRoot, "deps-attempts.count"), "utf8"),
    "1",
    "a timed-out attempt with live locks must not start another apt process",
  );
});

test("VOC-095-TEST-01b: install script exists at the repository-managed path", () => {
  assert.equal(
    path.basename(installScriptPath),
    "install-playwright-chromium.sh",
  );
  assert.match(readInstallScript(), /install-playwright-chromium\.sh/);
});
