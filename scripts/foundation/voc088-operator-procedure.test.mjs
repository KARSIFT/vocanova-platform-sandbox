// VOC-088-T03 — operator procedure documentation coverage (TEST-12).

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
const operationsReadmePath = path.join(
  repositoryRoot,
  "docs/operations/README.md",
);
const t03EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/t03-evidence.md",
);

test("VOC-088-TEST-12: operator procedure documents secret update and verification", () => {
  assert.ok(
    existsSync(operatorDocPath),
    "staging-controlled-signup.md must exist",
  );
  const doc = readFileSync(operatorDocPath, "utf8");

  assert.match(doc, /STAGING_NEW_USER_SIGNUP_ALLOWLIST/);
  assert.match(doc, /deploy-staging\.yml/);
  assert.match(doc, /Add or remove a cohort member/i);
  assert.match(doc, /Pick up the change/i);
  assert.match(doc, /controlled signup ready: true/);
  assert.match(doc, /Prove cohort preservation across automatic deploys/i);
  assert.match(doc, /two consecutive/i);
  assert.match(doc, /push/i);
  assert.match(doc, /gh run cancel RUN_ID/);
  assert.match(doc, /plan-from-issue/);
  assert.doesNotMatch(doc, /temporarily\s+pointing|deliberate.*failure/i);
  assert.doesNotMatch(doc, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});

test("VOC-088-TEST-12: operations index links the staging controlled-signup guide", () => {
  const readme = readFileSync(operationsReadmePath, "utf8");
  assert.match(readme, /staging-controlled-signup\.md/);
});

test("VOC-088-TEST-13: deploy-and-verify evidence file exists with required sections", () => {
  assert.ok(existsSync(t03EvidencePath), "t03-evidence.md must exist");
  const evidence = readFileSync(t03EvidencePath, "utf8");

  assert.match(evidence, /evidence_id:\s*VOC-088-EV-03/);
  assert.match(evidence, /VOC-088-AC-08/);
  assert.match(evidence, /VOC-088-AC-09/);
  assert.match(evidence, /controlled_signup_ready/);
  assert.match(evidence, /synthetic\.staging\.oauth-expected-state/);
  assert.match(evidence, /operational-failure/);

  assert.match(
    evidence,
    /live_signin_claimed:\s*true/,
    "AC-09 requires live sign-in evidence (allowlisted success + unlisted 503)",
  );
  assert.match(
    evidence,
    /live_failure_fixture_claimed:\s*true/,
    "AC-09 requires live controlled failure-fixture evidence (App-created issue + dedup)",
  );
  assert.match(
    evidence,
    /gate_status:\s*complete/,
    "all five AC-09 checklist items must be recorded before gate passes",
  );

  const addresses = evidence.match(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/gi);
  assert.deepEqual(addresses ?? [], []);
});
