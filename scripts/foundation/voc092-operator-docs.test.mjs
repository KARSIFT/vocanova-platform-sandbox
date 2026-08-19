// VOC-092-T02 — harness boundary documentation and VOC-088 evidence remediation
// coverage (TEST-13, TEST-14, TEST-16).

import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const operatorDocPath = path.join(
  repositoryRoot,
  "docs/operations/staging-controlled-signup.md",
);
const voc088T03EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/t03-evidence.md",
);
const localRunnerPath = path.join(
  repositoryRoot,
  "infra/scripts/run-controlled-signup-oauth-e2e.sh",
);
const callbackWorkflowPath = path.join(
  repositoryRoot,
  ".github/workflows/controlled-signup-oauth-e2e.yml",
);

test("VOC-092-TEST-13: operations doc states harness scope, non-goals, and human audit cadence", () => {
  assert.ok(
    existsSync(operatorDocPath),
    "staging-controlled-signup.md must exist",
  );
  const doc = readFileSync(operatorDocPath, "utf8");

  assert.match(doc, /Repository-managed OAuth callback E2E harness \(VOC-092\)/);
  assert.match(doc, /What the harness proves/i);
  assert.match(doc, /What the harness does not prove/i);
  assert.match(doc, /Google account login UI/i);
  assert.match(doc, /consent screen/i);
  assert.match(doc, /GoogleOAuthProvider/i);
  assert.match(doc, /disposable loopback PostgreSQL/i);
  assert.match(doc, /onboarding\/home/i);
  assert.match(doc, /HTTP 503/i);
  assert.match(doc, /synthetic\.staging\.oauth-expected-state/);
  assert.match(doc, /verify-staging-oauth-start\.sh/);

  assert.match(doc, /Periodic real-provider audit/i);
  assert.match(doc, /at least\s+quarterly/i);
  assert.match(doc, /auth callback/i);
  assert.match(doc, /allowlist policy/i);

  assert.match(doc, /run-controlled-signup-oauth-e2e\.sh/);
  assert.match(doc, /test:controlled-signup-oauth-e2e/);
  assert.match(doc, /controlled-signup-oauth-e2e\.yml/);

  assert.match(doc, /Human sign-in verification/i);
  assert.doesNotMatch(doc, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});

test("VOC-092-TEST-14: VOC-088 t03 evidence placeholder removed and allowlisted outcome expanded", () => {
  assert.ok(
    existsSync(voc088T03EvidencePath),
    "VOC-088 t03-evidence.md must exist",
  );
  const evidence = readFileSync(voc088T03EvidencePath, "utf8");

  assert.doesNotMatch(
    evidence,
    /bind-at-independent-review/,
    "reviewed_sha placeholder must not remain",
  );
  assert.doesNotMatch(
    evidence,
    /^reviewed_sha:/m,
    "reviewed_sha must not carry a self-referential placeholder",
  );

  assert.match(
    evidence,
    /Allowlisted first-time Google user \| pass\s+\| Reached staging onboarding\/home without HTTP 503/,
  );

  const addresses = evidence.match(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/gi);
  assert.deepEqual(addresses ?? [], []);
});

test("VOC-092-TEST-16: operator doc links reference committed harness commands", () => {
  assert.ok(
    existsSync(localRunnerPath),
    "local harness wrapper referenced by docs must exist",
  );
  assert.ok(
    existsSync(callbackWorkflowPath),
    "callback E2E workflow referenced by docs must exist",
  );

  const localRunner = readFileSync(localRunnerPath, "utf8");
  assert.match(localRunner, /ControlledSignupOAuth/);

  const packageJson = JSON.parse(
    readFileSync(path.join(repositoryRoot, "package.json"), "utf8"),
  );
  assert.match(
    packageJson.scripts["test:controlled-signup-oauth-e2e"],
    /ControlledSignupOAuth/,
  );
});
