import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { extractTopLevelJobBlock } from "../../infra/monitoring/scheduled-synthetics.mjs";

const root = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const read = (file) => readFileSync(path.join(root, file), "utf8");
const deploy = read(".github/workflows/deploy-staging.yml");
const scheduled = read(".github/workflows/scheduled-synthetics.yml");
const fixtureName = "prepare-synthetic-staging-journey";
const accountLock = "staging-synthetic-account";

// Read only the small scalar/block subset used here. Assertions compare
// values and command order, allowing whitespace, comments and quoted scalars.
function block(source, key) {
  const lines = source.split("\n");
  const marker = new RegExp(`^(\\s*)${key}:\\s*(?:\\|[-+]?)?\\s*$`);
  const start = lines.findIndex((line) => marker.test(line));
  assert.ok(start >= 0, `missing ${key} block`);
  const indent = lines[start].match(marker)[1].length;
  let end = start + 1;
  while (end < lines.length) {
    const line = lines[end];
    if (line.trim() && !line.trimStart().startsWith("#")) {
      if (line.match(/^\s*/)[0].length <= indent) break;
    }
    end++;
  }
  return lines.slice(start + 1, end).join("\n");
}

function scalars(source) {
  return Object.fromEntries(
    source
      .split("\n")
      .filter((line) => /^\s*[\w-]+\s*:/.test(line))
      .map((line) => {
        const [, key, raw] = line.match(/^\s*([\w-]+)\s*:\s*(.*?)\s*$/);
        const value = raw.replace(/\s+#.*$/, "").trim();
        return [
          key,
          value === "false"
            ? false
            : value === "true"
              ? true
              : value.replace(/^(["'])(.*)\1$/, "$2"),
        ];
      }),
  );
}

function step(source, name) {
  const entries = [...source.matchAll(/^\s*-\s*name:\s*(.+)$/gm)];
  const index = entries.findIndex(
    (entry) => entry[1].trim().replace(/^(["'])(.*)\1$/, "$2") === name,
  );
  assert.ok(index >= 0, `missing workflow step: ${name}`);
  return source.slice(entries[index].index, entries[index + 1]?.index);
}

function ordered(source, commands) {
  let previous = -1;
  for (const command of commands) {
    const index = source.search(command);
    assert.ok(index > previous, `missing or out-of-order command: ${command}`);
    previous = index;
  }
}

const prepareCommand = new RegExp(
  `^\\s*sh\\s+["']?/opt/vocanova/apps/api/scripts/${fixtureName}\\.sh["']?\\s*$`,
  "m",
);
const sourceEnvironment =
  /^\s*(?:\.|source)\s+["']?\/opt\/vocanova\/infra\/secrets\/api\.env["']?\s*$/m;
const legacySeedInvocation =
  /^\s*(?:sh|bash|\.|source)\s+["']?\/opt\/vocanova\/apps\/api\/scripts\/seed-synthetic-smoke-user\.sh\b/m;

test("staging fixture setup and browser journeys share a non-cancelling job lock", () => {
  for (const [source, jobId] of [
    [deploy, "deploy"],
    [scheduled, "staging-authenticated-core-journey"],
  ]) {
    const job = extractTopLevelJobBlock(source, jobId);
    assert.ok(job, `missing job ${jobId}`);
    const concurrency = scalars(block(job, "concurrency"));
    assert.equal(concurrency.group, accountLock);
    assert.equal(concurrency["cancel-in-progress"], false);
    assert.equal(concurrency.queue, "max");
    assert.match(job, prepareCommand);
    assert.doesNotMatch(
      job,
      legacySeedInvocation,
      "the legacy seed must not alter identity/history before guarded preparation",
    );
    ordered(job, [
      prepareCommand,
      /id:\s*mint_staging_session/,
      /playwright test/,
    ]);
  }

  // Keep the existing independent workflow queues; the new lock belongs only
  // to the jobs that can prepare or consume the staging account.
  assert.equal(scalars(block(deploy, "concurrency")).group, "staging-deploy");
  assert.equal(
    scalars(block(scheduled, "concurrency")).group,
    "scheduled-synthetics",
  );
});

test("staging ships both fixture files and prepares after migrations and canonical content", () => {
  const bundle = block(step(deploy, "Bundle deployable artifacts"), "run");
  for (const extension of ["sh", "sql"]) {
    assert.match(
      bundle,
      new RegExp(
        `^\\s*cp\\s+["']?apps/api/scripts/${fixtureName}\\.${extension}["']?\\s+["']?/tmp/deploy-bundle/apps/api/scripts/["']?\\s*$`,
        "m",
      ),
    );
  }

  const hostStep = step(deploy, "Deploy to staging host");
  const script = block(hostStep, "script");
  assert.match(script, /set\s+-euo\s+pipefail/);
  assert.doesNotMatch(hostStep, /continue-on-error:\s*true/);
  ordered(script, [
    sourceEnvironment,
    /sh\s+\/opt\/vocanova\/apps\/api\/scripts\/migrate\.sh/,
    /DATABASE_URL=.*\/opt\/vocanova\/apps\/api\/bin\/p1-content-seed/,
    prepareCommand,
    /^\s*docker\s+compose\s+up\s+-d\s+postgres\s+api\s+web\s*$/m,
  ]);
});

test("scheduled staging requires installed fixture preparation before session mint", () => {
  const job = extractTopLevelJobBlock(
    scheduled,
    "staging-authenticated-core-journey",
  );
  const setup = step(job, "Refresh reserved staging synthetic review state");
  const script = block(setup, "script");
  assert.match(script, /set\s+-euo\s+pipefail/);
  assert.doesNotMatch(setup, /continue-on-error:\s*true/);
  ordered(script, [sourceEnvironment, prepareCommand]);
  // A direct unconditional command, without a file-existence fallback or
  // ignored exit status, fails closed when the installed wrapper is missing.
  assert.doesNotMatch(script, /\bif\b|\|\|/);
});

test("fresh journey preparation stays outside production and the default seed", () => {
  for (const file of [
    ".github/workflows/deploy-production.yml",
    "apps/api/scripts/seed-synthetic-smoke-user.sh",
    "apps/api/scripts/seed-synthetic-smoke-user.sql",
  ]) {
    assert.ok(
      !read(file).includes(fixtureName),
      `${file} must not prepare a fresh journey`,
    );
  }
  for (const jobId of [
    "staging-oauth-expected-state",
    "production-oauth-expected-state",
    "production-journey-content",
    "production-authenticated-route-content-sweep",
  ]) {
    const job = extractTopLevelJobBlock(scheduled, jobId);
    assert.ok(job, `missing job ${jobId}`);
    assert.ok(
      !job.includes(fixtureName),
      `${jobId} must not prepare a fresh journey`,
    );
    assert.ok(
      !job.includes(accountLock),
      `${jobId} must not hold the staging account lock`,
    );
  }
});
