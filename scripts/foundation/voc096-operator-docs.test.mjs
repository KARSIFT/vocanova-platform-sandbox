// VOC-096-T02 — production operator procedure and evidence coverage
// (TEST-12, TEST-13).

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
  "docs/operations/production-controlled-signup.md",
);
const stagingDocPath = path.join(
  repositoryRoot,
  "docs/operations/staging-controlled-signup.md",
);
const operationsReadmePath = path.join(
  repositoryRoot,
  "docs/operations/README.md",
);
const t02EvidencePath = path.join(
  repositoryRoot,
  "specs/changes/VOC-096-persist-controlled-production-google-signup-cohort/t02-evidence.md",
);

test("VOC-096-TEST-12: operator procedure documents secret update and verification", () => {
  assert.ok(
    existsSync(operatorDocPath),
    "production-controlled-signup.md must exist",
  );
  const doc = readFileSync(operatorDocPath, "utf8");

  assert.match(doc, /PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST/);
  assert.match(doc, /deploy-production\.yml/);
  assert.match(doc, /Add or remove a cohort member/i);
  assert.match(doc, /Pick up the change/i);
  assert.match(doc, /controlled signup ready: true/);
  assert.match(doc, /Prove cohort preservation across automatic deploys/i);
  assert.match(doc, /two consecutive/i);
  assert.match(doc, /push/i);
  assert.match(doc, /staging-controlled-signup\.md/);
  assert.match(doc, /verify-production-oauth-start\.sh/);
  assert.match(doc, /synthetic\.production\.oauth-expected-state/);
  assert.match(doc, /Periodic real-provider audit/i);
  assert.match(doc, /at least\s+quarterly/i);
  assert.doesNotMatch(doc, /STAGING_NEW_USER_SIGNUP_ALLOWLIST/);
  assert.doesNotMatch(doc, /[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i);
});

test("VOC-096-TEST-12: operations index links the production controlled-signup guide", () => {
  const readme = readFileSync(operationsReadmePath, "utf8");
  assert.match(readme, /production-controlled-signup\.md/);
});

test("VOC-096-TEST-13: deploy-and-verify evidence file exists with required sections", () => {
  assert.ok(existsSync(t02EvidencePath), "t02-evidence.md must exist");
  const evidence = readFileSync(t02EvidencePath, "utf8");

  assert.match(evidence, /evidence_id:\s*VOC-096-EV-02/);
  assert.match(evidence, /VOC-096-AC-09/);
  assert.match(evidence, /VOC-096-AC-10/);
  assert.match(evidence, /controlled_signup_ready/);
  assert.match(evidence, /synthetic\.production\.oauth-expected-state/);
  assert.match(evidence, /verify-production-oauth-start\.sh/);
  assert.match(evidence, /deploy-production/);
  assert.match(evidence, /smoke-test-production/);

  const addresses = evidence.match(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/gi);
  assert.deepEqual(addresses ?? [], []);
});

test("VOC-096-TEST-12: production doc cross-links staging without duplicating VOC-092 harness detail", () => {
  assert.ok(
    existsSync(stagingDocPath),
    "staging-controlled-signup.md must exist",
  );
  const productionDoc = readFileSync(operatorDocPath, "utf8");
  const stagingDoc = readFileSync(stagingDocPath, "utf8");

  assert.match(
    productionDoc,
    /Repository-managed OAuth callback E2E harness \(VOC-092\)/,
  );
  assert.match(
    productionDoc,
    /staging-controlled-signup\.md#repository-managed-oauth-callback-e2e-harness-voc-092/,
  );
  assert.match(stagingDoc, /run-controlled-signup-oauth-e2e\.sh/);
  assert.doesNotMatch(productionDoc, /run-controlled-signup-oauth-e2e\.sh/);
});
