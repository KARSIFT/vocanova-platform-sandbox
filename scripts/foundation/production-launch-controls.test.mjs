import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync, statSync } from "node:fs";
import path from "node:path";

// VOC-037-T03 / VOC-037-TEST-03: static guards for the production
// launch controls DOC-11 §3 requires.
//
// The kill switches and the rollback path can only be *exercised*
// against the live production tier (see
// infra/scripts/rehearse-production-killswitch-rollback.sh and
// VOC-037-EV-03). What can be checked on every pull request is that
// the deploy path still makes them operable at all: that the four
// switch values come from the run's own inputs rather than from
// hardcoded literals, that a rollback run redeploys a named artifact
// instead of rebuilding one, and that it never reverses a migration
// on the way.
//
// These are the regressions that would silently remove a launch
// control while every other check stayed green.

const workflowPath = path.resolve(".github/workflows/deploy-production.yml");
const rehearsalPath = path.resolve(
  "infra/scripts/rehearse-production-killswitch-rollback.sh",
);
const selftestPath = path.resolve(
  "infra/scripts/rehearse-production-killswitch-rollback.selftest.sh",
);

const workflow = readFileSync(workflowPath, "utf8");
const rehearsal = readFileSync(rehearsalPath, "utf8");
const selftest = readFileSync(selftestPath, "utf8");

const KILL_SWITCHES = [
  ["AI_FEATURES_ENABLED", "PRODUCTION_AI_FEATURES_ENABLED", "ai"],
  [
    "EMAIL_MAGIC_LINK_ENABLED",
    "PRODUCTION_EMAIL_MAGIC_LINK_ENABLED",
    "magic",
  ],
  ["GOOGLE_OAUTH_ENABLED", "PRODUCTION_GOOGLE_OAUTH_ENABLED", "oauth"],
  [
    "NEW_USER_SIGNUP_ENABLED",
    "PRODUCTION_NEW_USER_SIGNUP_ENABLED",
    "signups",
  ],
];

test("production deploys behind the founder-reviewed environment", () => {
  assert.match(workflow, /^\s{4}environment: production$/m);
});

test("every kill switch is written from a workflow input", () => {
  for (const [switchName, envName] of KILL_SWITCHES) {
    assert.match(
      workflow,
      new RegExp(`echo "${switchName}=\\$\\{${envName}\\}"`),
      `${switchName} must be written from ${envName}`,
    );
    assert.doesNotMatch(
      workflow,
      new RegExp(`echo "${switchName}=(true|false)"`),
      `${switchName} must not be hardcoded; a hardcoded switch is not a switch`,
    );
  }
});

test("a rollback run redeploys a named artifact instead of building one", () => {
  const buildSteps = workflow.match(/- name: Build and push [^\n]+/g) ?? [];
  assert.equal(buildSteps.length, 2);

  for (const buildStep of buildSteps) {
    const guard = workflow.slice(
      workflow.indexOf(buildStep) + buildStep.length,
    );
    assert.match(
      guard.slice(0, 80),
      /if: inputs\.deploy_mode == 'build-and-deploy'/,
      `${buildStep} must not run in rollback mode`,
    );
  }

  // Only an immutable per-commit tag can identify one specific build,
  // which is what "redeploy the previous artifact by digest" requires.
  assert.match(workflow, /\^sha-\[0-9a-f\]\{7\}\$/);
  assert.match(workflow, /docker buildx imagetools inspect/);
});

test("a rollback run never reverses a production migration", () => {
  assert.match(
    workflow,
    /if \[ "\$PRODUCTION_DEPLOY_MODE" = "redeploy-existing-image" \]; then\n\s*echo "rollback mode: skipping migration apply/,
  );
});

test("the deploy bundle carries the rehearsal script to the host", () => {
  assert.match(
    workflow,
    /cp infra\/scripts\/rehearse-production-killswitch-rollback\.sh/,
  );
  assert.match(
    workflow,
    /chmod \+x [^\n]*rehearse-production-killswitch-rollback\.sh/,
  );
});

test("the rehearsal script and its harness are executable", () => {
  for (const scriptPath of [rehearsalPath, selftestPath]) {
    const mode = statSync(scriptPath).mode & 0o111;
    assert.notEqual(mode, 0, `${scriptPath} must be executable`);
  }
});

test("the rehearsal asserts each switch's documented disabled effect", () => {
  for (const [switchName] of KILL_SWITCHES) {
    assert.match(rehearsal, new RegExp(`api_env_set ${switchName} false`));
    assert.match(rehearsal, new RegExp(`api_env_set ${switchName} true`));
  }

  // 503 for the three auth paths; a 200 carrying the business error
  // code for AI feedback (apps/api/business/aifeedback/service.go).
  assert.match(rehearsal, /require_status 503/);
  assert.match(rehearsal, /AI_FEEDBACK_GENERATION_DISABLED/);
});

test("the harness proves the rehearsal fails when a control is broken", () => {
  for (const [, , stubName] of KILL_SWITCHES) {
    assert.match(
      selftest,
      new RegExp(`for broken_switch in [^\\n]*\\b${stubName}\\b`),
      `the harness must cover an ignored ${stubName} switch`,
    );
  }

  for (const brokenControl of [
    "FAKE_IGNORE_IMAGE_TAG=1",
    "FAKE_ROLLBACK_DROPS_ROW=1",
    "FAKE_SKIP_CLEANUP=1",
    "FAKE_NEVER_HEALTHY=1",
  ]) {
    assert.match(selftest, new RegExp(`expect_fail [^\\n]*${brokenControl}`));
  }
});
