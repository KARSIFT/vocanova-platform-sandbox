// Every merged change deploys to staging, including docs-only releases.
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const workflow = readFileSync(
  new URL("../../.github/workflows/deploy-staging.yml", import.meta.url),
  "utf8",
);
const production = readFileSync(
  new URL("../../.github/workflows/deploy-production.yml", import.meta.url),
  "utf8",
);

test("every main push deploys to staging without a changed-path filter", () => {
  const push = workflow.slice(
    workflow.indexOf("  push:"),
    workflow.indexOf("  workflow_dispatch:"),
  );
  assert.match(push, /branches:\s*\n\s*- main\s*\n/);
  assert.doesNotMatch(push, /paths(?:-ignore)?:|branches-ignore:|tags:/);
  assert.equal((push.match(/^\s*- /gm) ?? []).length, 1);
});

test("manual recovery, queued deployments and real staging journeys remain available", () => {
  assert.match(workflow, /workflow_dispatch:/);
  assert.match(workflow, /skip_ssh_deploy:/);
  assert.match(workflow, /recovery_target_sha:/);
  assert.match(workflow, /group: staging-deploy/);
  assert.match(workflow, /cancel-in-progress: false/);
  assert.match(workflow, /queue: max/);
  assert.match(workflow, /STAGING_SSH_HOST/);
  assert.match(workflow, /Run the staging core-loop journey/);
  assert.match(workflow, /Run the staging maturity journey/);
});

test("both environments derive and verify the same commit-based version from full history", () => {
  for (const source of [workflow, production]) {
    assert.match(source, /fetch-depth: 0/);
    assert.match(
      source,
      /bash scripts\/foundation\/release-metadata\.sh >> "\$GITHUB_OUTPUT"/,
    );
    assert.match(
      source,
      /VOCANOVA_BUILD_VERSION=\$\{\{ steps\.release\.outputs\.version \}\}/,
    );
    assert.match(
      source,
      /NEXT_PUBLIC_VOCANOVA_BUILD_VERSION=\$\{\{ steps\.release\.outputs\.version \}\}/,
    );
    assert.match(
      source,
      /RELEASE_VERSION: \$\{\{ steps\.release\.outputs\.version \}\}/,
    );
  }
});

test("production still requires deliberate manual dispatch", () => {
  const triggers = production.slice(
    production.indexOf("\non:"),
    production.indexOf("\npermissions:"),
  );
  assert.match(triggers, /workflow_dispatch:/);
  assert.doesNotMatch(
    triggers,
    /^  (push|pull_request|workflow_run|release):/m,
  );
});
