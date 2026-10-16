import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { test } from "node:test";

const script = fileURLToPath(new URL("./release-metadata.sh", import.meta.url));
const env = {
  ...process.env,
  GIT_CONFIG_NOSYSTEM: "1",
  GIT_CONFIG_GLOBAL: "/dev/null",
};

function run(cwd, command, args) {
  const result = spawnSync(command, args, {
    cwd,
    env,
    encoding: "utf8",
    timeout: 10000,
  });
  assert.ifError(result.error);
  return result;
}

function git(cwd, ...args) {
  const result = run(cwd, "git", args);
  assert.equal(result.status, 0, result.stderr);
  return result.stdout.trim();
}

function fixture(t, version = "0.4.1") {
  const root = mkdtempSync(join(tmpdir(), "vocanova-release-test-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  git(root, "init", "-b", "main");
  git(root, "config", "user.name", "Release test");
  git(root, "config", "user.email", "release-test@example.invalid");
  git(root, "config", "commit.gpgsign", "false");
  let count = 0;
  const commit = (filename = "README.md", content = `Change ${++count}\n`) => {
    writeFileSync(join(root, filename), content);
    git(root, "add", "--", filename);
    git(root, "commit", "-m", "Test release change");
    return git(root, "rev-parse", "HEAD");
  };
  commit("VERSION", `${version}\n`);
  return { root, commit };
}

function metadata(root) {
  const result = run(root, "bash", [script]);
  assert.equal(result.status, 0, result.stderr);
  const entries = result.stdout.trim().split("\n");
  assert.equal(entries.length, 3, "Only public release metadata is emitted");
  const values = Object.fromEntries(entries.map((line) => line.split("=")));
  assert.equal(values.commit, git(root, "rev-parse", "HEAD"));
  assert.match(values.built_at, /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/);
  assert.ok(Number.isFinite(Date.parse(values.built_at)));
  return values;
}

function rejects(root) {
  const result = run(root, "bash", [script]);
  assert.notEqual(result.status, 0, "Unsafe release state must fail closed");
  assert.ok(result.stderr.trim(), "Failure must explain the rejected state");
  assert.equal(
    result.stdout,
    "",
    "Do not emit usable metadata for a rejected build",
  );
}

test("baseline, docs-only changes and repeat builds have deterministic versions", (t) => {
  const { root, commit } = fixture(t);
  assert.equal(metadata(root).version, "0.4.1");
  commit();
  assert.equal(metadata(root).version, "0.4.2");
  assert.equal(metadata(root).version, "0.4.2");
  commit("another-document.md", "Another documentation change\n");
  assert.equal(metadata(root).version, "0.4.3");
});

test("a normal merge counts once regardless of topic branch commit count", (t) => {
  const { root, commit } = fixture(t);
  git(root, "checkout", "-b", "topic");
  commit();
  commit();
  commit();
  git(root, "checkout", "main");
  git(root, "merge", "--no-ff", "topic", "-m", "Merge topic");
  assert.equal(metadata(root).version, "0.4.2");
  commit("main-document.md", "Next main change\n");
  assert.equal(metadata(root).version, "0.4.3");
});

test("a VERSION change merged from a topic branch starts a new baseline", (t) => {
  const { root, commit } = fixture(t);
  commit();
  git(root, "checkout", "-b", "minor-release");
  commit("VERSION", "0.5.0\n");
  commit();
  git(root, "checkout", "main");
  commit("main-document.md", "Concurrent main change\n");
  git(root, "merge", "--no-ff", "minor-release", "-m", "New minor release");
  assert.equal(metadata(root).version, "0.5.0");
  commit();
  assert.equal(metadata(root).version, "0.5.1");
});

test("checking out an older commit retains its original release identity", (t) => {
  const { root, commit } = fixture(t);
  const previous = commit();
  assert.equal(metadata(root).version, "0.4.2");
  commit("VERSION", "1.0.0\n");
  commit();
  assert.equal(metadata(root).version, "1.0.1");
  git(root, "checkout", "--detach", previous);
  assert.equal(metadata(root).version, "0.4.2");
});

test("shallow history cannot silently produce a reused release version", (t) => {
  const { root, commit } = fixture(t);
  commit();
  const shallow = mkdtempSync(join(tmpdir(), "vocanova-release-shallow-"));
  t.after(() => rmSync(shallow, { recursive: true, force: true }));
  git(root, "clone", "--depth=1", pathToFileURL(root).href, shallow);
  assert.equal(git(shallow, "rev-parse", "--is-shallow-repository"), "true");
  rejects(shallow);
});

for (const staged of [false, true]) {
  test(`rejects ${staged ? "staged" : "unstaged"} VERSION changes`, (t) => {
    const { root } = fixture(t);
    writeFileSync(join(root, "VERSION"), "9.0.0\n");
    if (staged) git(root, "add", "VERSION");
    rejects(root);
  });
}

test("an untracked VERSION cannot identify the checked-out release", (t) => {
  const { root, commit } = fixture(t);
  git(root, "rm", "VERSION");
  commit();
  writeFileSync(join(root, "VERSION"), "0.4.1\n");
  rejects(root);
});

for (const invalid of [
  "01.4.1",
  "0.04.1",
  "0.4.01",
  "0.4",
  "v0.4.1",
  "0.4.1-rc.1",
  "0.4.1+build.1",
  "0.4.1 extra",
  "",
]) {
  test(`rejects unsupported release version ${JSON.stringify(invalid)}`, (t) => {
    const { root } = fixture(t, invalid);
    rejects(root);
  });
}
