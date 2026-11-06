import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  mkdtempSync,
  mkdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";

const workflows = [
  {
    tier: "staging",
    step: "Sync AI-provider configuration from GitHub secrets",
    directory: "/opt/vocanova/infra",
    prefix: "",
  },
  {
    tier: "production",
    step: "Sync production AI-provider configuration from GitHub secrets",
    directory: "/opt/vocanova/production",
    prefix: "PRODUCTION_",
  },
];

function syncFixture(t, workflow) {
  const temporary = mkdtempSync(path.join(tmpdir(), "vocanova-provider-sync-"));
  t.after(() => rmSync(temporary, { recursive: true, force: true }));
  const directory = path.join(temporary, "deployment");
  mkdirSync(path.join(directory, "secrets"), { recursive: true });
  const file = path.join(directory, "secrets/api.env");
  const original = [
    "AI_PROVIDER=opencode",
    "AI_PROVIDER_BASE_URL=http://127.0.0.1:4096",
    "AI_PROVIDER_BASE_URL=https://stale-provider.invalid/v1",
    "AI_PROVIDER_API_KEY=old-synthetic-key",
    "PRONUNCIATION_PROVIDER=old-provider",
    "PRONUNCIATION_API_KEY=old-pronunciation-key",
    "UNRELATED_SETTING=preserve-this-value",
    "",
  ].join("\n");
  writeFileSync(file, original, { mode: 0o600 });
  const yaml = readFileSync(
    new URL(
      `../../.github/workflows/deploy-${workflow.tier}.yml`,
      import.meta.url,
    ),
    "utf8",
  );
  const step = yaml.indexOf(`      - name: ${workflow.step}\n`);
  assert.notEqual(step, -1);
  const scriptStart = yaml.indexOf("          script: |\n", step);
  assert.notEqual(scriptStart, -1);
  const lines = yaml
    .slice(scriptStart + "          script: |\n".length)
    .split("\n");
  const scriptLines = [];
  for (const line of lines) {
    if (line.trim() && !line.startsWith("            ")) break;
    scriptLines.push(line.slice(12));
  }
  // Execute the workflow's real shell block against an isolated temporary
  // deployment directory. No elevation, host configuration or credentials.
  let script = scriptLines.join("\n").replaceAll(workflow.directory, directory);
  assert.ok(!script.includes("/opt/vocanova/"));
  if (process.platform === "darwin")
    script = script.replaceAll("sed -i \\\n", 'sed -i "" \\\n');
  const env = {
    PATH: process.env.PATH,
    [`${workflow.prefix}CLOUDFLARE_API_TOKEN`]: "synthetic-cloudflare-key",
    [`${workflow.prefix}CLOUDFLARE_ACCOUNT_ID`]: "synthetic-account",
    PRONUNCIATION_API_KEY: "synthetic-pronunciation-key",
  };
  const run = (overrides = {}) =>
    spawnSync("bash", ["-c", 'sudo() { "$@"; }\n' + script], {
      encoding: "utf8",
      env: { ...env, ...overrides },
      timeout: 5_000,
    });
  return { file, original, run, env };
}

for (const workflow of workflows) {
  test(`${workflow.tier} provider sync replaces stale endpoints and remains idempotent`, (t) => {
    const fixture = syncFixture(t, workflow);
    for (let attempt = 0; attempt < 2; attempt += 1) {
      const result = fixture.run();
      assert.equal(result.status, 0, result.stderr);
      assert.equal(result.stdout, "");
      assert.equal(result.stderr, "");
      const lines = readFileSync(fixture.file, "utf8").trim().split("\n");
      assert.deepEqual(
        lines.filter((line) => line.startsWith("AI_PROVIDER_BASE_URL=")),
        ["AI_PROVIDER_BASE_URL=https://api.cloudflare.com/client/v4"],
      );
      assert.ok(lines.includes("AI_PROVIDER=cloudflare"));
      assert.ok(lines.includes("AI_PROVIDER_API_KEY=synthetic-cloudflare-key"));
      assert.ok(lines.includes("AI_PROVIDER_ACCOUNT_ID=synthetic-account"));
      assert.ok(lines.includes("PRONUNCIATION_PROVIDER=openai"));
      assert.ok(
        lines.includes("PRONUNCIATION_API_KEY=synthetic-pronunciation-key"),
      );
      assert.ok(lines.includes("UNRELATED_SETTING=preserve-this-value"));
      assert.equal(statSync(fixture.file).mode & 0o777, 0o600);
    }
  });

  for (const missing of [
    `${workflow.prefix}CLOUDFLARE_API_TOKEN`,
    `${workflow.prefix}CLOUDFLARE_ACCOUNT_ID`,
    "PRONUNCIATION_API_KEY",
  ]) {
    test(`${workflow.tier} provider sync refuses missing ${missing} before changing configuration`, (t) => {
      const fixture = syncFixture(t, workflow);
      const result = fixture.run({ [missing]: "" });
      assert.equal(result.status, 1);
      assert.equal(readFileSync(fixture.file, "utf8"), fixture.original);
      for (const value of Object.values(fixture.env).filter((value) =>
        value.startsWith("synthetic-"),
      )) {
        assert.ok(!`${result.stdout}${result.stderr}`.includes(value));
      }
    });
  }
}
