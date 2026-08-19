// VOC-087-T02 — rotation recovery, shell harness execution, and evidence checks.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { redactSecrets } from "../../infra/monitoring/kuma-sync/redact.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const workflowPath = path.join(
  repositoryRoot,
  ".github/workflows/sync-monitoring.yml",
);
const rotateScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/kuma-rotate-credentials.sh",
);
const recoveryGatePath = path.join(
  repositoryRoot,
  "infra/scripts/kuma-rotation-recovery-gate.sh",
);
const voc086T02EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-086-manage-monitoring-inventory/t02-evidence.md",
);
const voc087T02EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-087-make-the-first-repository-managed-kuma-sync-adopt/t02-evidence.md",
);

function extractStepBlock(workflowSource, stepName) {
  const marker = `- name: ${stepName}`;
  const stepStartIndex = workflowSource.indexOf(marker);
  assert.notEqual(
    stepStartIndex,
    -1,
    `sync-monitoring workflow is missing step: ${stepName}`,
  );

  const nextStepIndex = workflowSource.indexOf(
    "\n      - name:",
    stepStartIndex + 1,
  );
  return workflowSource.slice(
    stepStartIndex,
    nextStepIndex === -1 ? workflowSource.length : nextStepIndex,
  );
}

function runSelftest(relativePath) {
  const scriptPath = path.join(repositoryRoot, relativePath);
  const result = spawnSync("bash", [scriptPath], {
    cwd: repositoryRoot,
    encoding: "utf8",
  });
  assert.equal(
    result.status,
    0,
    `${relativePath} must pass (exit ${result.status}):\n${result.stdout}\n${result.stderr}`,
  );
}

function runRecoveryGate(command, env) {
  const result = spawnSync("bash", [recoveryGatePath, command], {
    cwd: repositoryRoot,
    encoding: "utf8",
    env: { ...process.env, ...env },
  });
  assert.equal(
    result.status,
    0,
    `recovery gate ${command} failed (exit ${result.status}):\n${result.stdout}\n${result.stderr}`,
  );
  return result.stdout.trim();
}

test("VOC-087-TEST-09: shell harnesses execute from the deterministic test entry point", () => {
  runSelftest("infra/scripts/kuma-rotate-credentials.selftest.sh");
  runSelftest("infra/scripts/sync-kuma-inventory.selftest.sh");
});

test("VOC-087-TEST-10: reset-success / proof-transfer-failure remains recoverable and fail-closed", () => {
  const workflow = readFileSync(workflowPath, "utf8");

  const rotateStep = extractStepBlock(
    workflow,
    "Rotate Kuma credentials on host (explicit opt-in)",
  );
  assert.match(rotateStep, /id: rotate_host/);
  assert.doesNotMatch(
    rotateStep,
    /reset-password\.js/,
    "host rotate step invokes the wrapper script, not a second reset-password.js path",
  );

  const fetchStep = extractStepBlock(
    workflow,
    "Fetch Kuma reset proof and username metadata from host",
  );
  assert.match(fetchStep, /id: fetch_proof/);
  assert.match(fetchStep, /host_reachable=/);
  assert.match(fetchStep, /proof_fetched=/);

  const storeStep = extractStepBlock(
    workflow,
    "Store rotated Kuma password in monitoring environment",
  );
  assert.match(storeStep, /id: store_password/);
  assert.match(storeStep, /kuma-rotation-recovery-gate\.sh/);
  assert.match(
    storeStep,
    /Proof transfer unavailable; storing from runner-local file after confirmed host rotation/,
  );
  assert.match(
    storeStep,
    /Cannot confirm whether this attempt reset Kuma; retaining password copies/,
  );

  const recoverStep = extractStepBlock(
    workflow,
    "Recover retained Kuma password from host (no reset)",
  );
  assert.match(recoverStep, /id: recover_store/);
  assert.match(recoverStep, /inputs\.recover_store_only/);
  assert.doesNotMatch(
    recoverStep,
    /reset-password/,
    "recover-store must never invoke reset-password.js",
  );
  assert.match(recoverStep, /kuma-new-password\.secret/);
  assert.match(recoverStep, /kuma-reset-applied\.env/);
  assert.match(recoverStep, /kuma-rotate-metadata\.env/);
  assert.match(recoverStep, /KUMA_ROTATION_ATTEMPT_ID/);
  assert.match(recoverStep, /metadata_attempt_id.*attempt_id/s);
  assert.doesNotMatch(
    recoverStep,
    /if scp/,
    "proof and username metadata are mandatory recovery inputs",
  );

  const refuseStep = extractStepBlock(
    workflow,
    "Refuse combined rotate and recover-store",
  );
  assert.match(refuseStep, /recover never resets/);

  const preflightStep = extractStepBlock(
    workflow,
    "Refuse a new rotation while recovery material exists",
  );
  assert.match(preflightStep, /id: rotation_preflight/);
  assert.match(preflightStep, /kuma-new-password\.secret/);
  assert.match(preflightStep, /kuma-reset-applied\.env/);
  assert.match(preflightStep, /kuma-rotate-metadata\.env/);

  const scrubGateStep = extractStepBlock(
    workflow,
    "Decide rotation credential scrub",
  );
  assert.match(scrubGateStep, /kuma-rotation-recovery-gate\.sh/);
  assert.match(scrubGateStep, /scrub-decision/);
  assert.match(
    scrubGateStep,
    /STORE_PASSWORD_STORED/,
    "scrub must key off password_stored output, not store step success (soft skip is not a store)",
  );
  assert.match(
    scrubGateStep,
    /STORE_USERNAME_STORED/,
    "scrub must require the username secret store as well as the password store",
  );
  assert.match(scrubGateStep, /RECOVER_USERNAME_STORED/);
  assert.match(scrubGateStep, /ROTATION_PREFLIGHT_OUTCOME/);
  assert.doesNotMatch(
    scrubGateStep,
    /steps\.store_password\.outcome/,
    "scrub must not treat store_password.outcome success as proof the secret was stored",
  );

  const hostCleanupStep = extractStepBlock(
    workflow,
    "Remove host rotation credential material",
  );
  assert.match(
    hostCleanupStep,
    /steps\.scrub_gate\.outputs\.decision == 'SCRUB'/,
    "host scrub must run only when the recovery gate says SCRUB",
  );

  const runnerCleanupStep = extractStepBlock(
    workflow,
    "Remove runner rotation credential material",
  );
  assert.match(
    runnerCleanupStep,
    /steps\.scrub_gate\.outputs\.decision == 'SCRUB'/,
    "runner scrub must run only when the recovery gate says SCRUB",
  );

  const retainStep = extractStepBlock(
    workflow,
    "Fail closed when rotation credentials must be retained",
  );
  assert.match(retainStep, /exit 1/);
  assert.match(retainStep, /no second reset-password\.js/);

  const rotateCount = (workflow.match(/Rotate Kuma credentials on host/g) ?? [])
    .length;
  assert.equal(
    rotateCount,
    1,
    "workflow must invoke host rotation exactly once (no second reset-password.js path)",
  );

  // Procedure 1–3: model host reset success + failed proof fetch, including
  // post-reset rotate_host failure (the attempt-1 High finding).
  assert.equal(
    runRecoveryGate("store-decision", {
      ROTATE_HOST_OUTCOME: "success",
      PROOF_MATCHES: "false",
      HOST_REACHABLE: "false",
      PROOF_FETCHED: "false",
      PASSWORD_STORED: "false",
      RECOVER_STORE_ONLY: "false",
    }),
    "STORE",
    "rotate_host success with failed fetch must store from the runner-local file",
  );
  assert.equal(
    runRecoveryGate("scrub-decision", {
      ROTATE_HOST_OUTCOME: "skipped",
      ROTATION_PREFLIGHT_OUTCOME: "failure",
      PROOF_MATCHES: "false",
      HOST_REACHABLE: "false",
      PROOF_FETCHED: "false",
      PASSWORD_STORED: "false",
      USERNAME_STORED: "false",
      RECOVER_STORE_ONLY: "false",
    }),
    "RETAIN",
    "a preflight-blocked rotation must not scrub the retained bundle it found",
  );
  assert.equal(
    runRecoveryGate("scrub-decision", {
      ROTATE_HOST_OUTCOME: "success",
      PROOF_MATCHES: "false",
      HOST_REACHABLE: "false",
      PROOF_FETCHED: "false",
      PASSWORD_STORED: "false",
      RECOVER_STORE_ONLY: "false",
    }),
    "RETAIN",
    "last password copy must not be scrubbed before store succeeds",
  );
  assert.equal(
    runRecoveryGate("store-decision", {
      ROTATE_HOST_OUTCOME: "failure",
      PROOF_MATCHES: "false",
      HOST_REACHABLE: "false",
      PROOF_FETCHED: "false",
      PASSWORD_STORED: "false",
      RECOVER_STORE_ONLY: "false",
    }),
    "RETAIN",
    "post-reset rotate_host failure plus failed fetch must not store an unproven password",
  );
  assert.equal(
    runRecoveryGate("scrub-decision", {
      ROTATE_HOST_OUTCOME: "failure",
      PROOF_MATCHES: "false",
      HOST_REACHABLE: "false",
      PROOF_FETCHED: "false",
      PASSWORD_STORED: "false",
      RECOVER_STORE_ONLY: "false",
    }),
    "RETAIN",
    "post-reset rotate_host failure plus failed fetch must not scrub the last copy",
  );
  assert.equal(
    runRecoveryGate("scrub-decision", {
      ROTATE_HOST_OUTCOME: "failure",
      PROOF_MATCHES: "false",
      HOST_REACHABLE: "false",
      PROOF_FETCHED: "false",
      PASSWORD_STORED: "true",
      USERNAME_STORED: "false",
      RECOVER_STORE_ONLY: "false",
    }),
    "RETAIN",
    "password-only secret storage must retain the host bundle",
  );
  assert.equal(
    runRecoveryGate("scrub-decision", {
      ROTATE_HOST_OUTCOME: "failure",
      PROOF_MATCHES: "false",
      HOST_REACHABLE: "false",
      PROOF_FETCHED: "false",
      PASSWORD_STORED: "true",
      USERNAME_STORED: "true",
      RECOVER_STORE_ONLY: "false",
    }),
    "SCRUB",
    "scrub is allowed only after both credential stores complete",
  );
});

test("VOC-087-TEST-11: sync/rotation tooling bans SQLite deployment paths", () => {
  const paths = [
    ".github/workflows/sync-monitoring.yml",
    "infra/scripts/kuma-rotate-credentials.sh",
    "infra/scripts/kuma-rotation-recovery-gate.sh",
    "infra/scripts/sync-kuma-inventory.sh",
  ];
  const forbidden = [/kuma\.db/i, /\bsqlite\b/i, /\/app\/data/i];

  for (const relativePath of paths) {
    const source = readFileSync(
      path.join(repositoryRoot, relativePath),
      "utf8",
    );
    for (const pattern of forbidden) {
      assert.ok(
        !pattern.test(source),
        `${relativePath} must not reference SQLite deployment paths (${pattern})`,
      );
    }
  }
});

test("VOC-087-TEST-12: credential redaction masks secrets in fixture logs", () => {
  const rotateScript = readFileSync(rotateScriptPath, "utf8");
  const fixture =
    "rotation failed password=super-secret KUMA_PASSWORD=abc123 Bearer eyJhbGciOiJIUzI1NiJ9";
  const redacted = redactSecrets(fixture);

  assert.ok(!redacted.includes("super-secret"));
  assert.ok(!redacted.includes("abc123"));
  assert.match(redacted, /\[REDACTED\]/);
  assert.match(rotateScript, /redact_sensitive/);
  assert.doesNotMatch(
    rotateScript,
    /echo\s+["']?\$password/,
    "rotate script must not echo the password variable",
  );

  const evidence = readFileSync(voc087T02EvidencePath, "utf8");
  assert.doesNotMatch(
    evidence,
    /password=[^\s]+|KUMA_PASSWORD=\S+/i,
    "T02 evidence must not embed credential material",
  );
});

test("VOC-087-TEST-13: VOC-086 T02 evidence describes proof-gated store and recovery boundary", () => {
  const evidence = readFileSync(voc086T02EvidencePath, "utf8");

  assert.doesNotMatch(
    evidence,
    /stored from the runner-local file \*before\* that fetch/,
    "T02 evidence must not claim password is stored before proof fetch",
  );
  assert.match(
    evidence,
    /proof-gated|proof transfer|reset-applied proof/i,
    "T02 evidence must describe proof-gated store",
  );
  assert.match(
    evidence,
    /recovery|proof-transfer-failure|VOC-087-D04/i,
    "T02 evidence must document the recovery boundary",
  );
  assert.match(
    evidence,
    /selftest|harness/i,
    "T02 evidence must state harnesses execute in CI",
  );
  assert.ok(
    existsSync(voc087T02EvidencePath),
    "VOC-087 T02 evidence must exist",
  );
});

test("VOC-087-TEST-14: package tasks do not claim live inventory apply", () => {
  const evidence = readFileSync(voc087T02EvidencePath, "utf8");

  assert.match(evidence, /live_inventory_apply_claimed:\s*false/);
  assert.match(
    evidence,
    /deferred|not claimed|blocked/i,
    "T02 evidence must state live apply remains deferred",
  );
  assert.doesNotMatch(
    evidence,
    /sync_inventory:\s*true.*live|dispatch.*live Kuma/i,
    "T02 evidence must not dispatch live inventory apply",
  );
});
