// VOC-085-T02 — production route-sweep coverage and fail-closed auth handling.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const smokeScriptPath = path.join(
  repositoryRoot,
  "infra/scripts/smoke-test-production.sh",
);
const deployProductionPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-production.yml",
);
const voc079InvariantsPath = path.join(
  repositoryRoot,
  "scripts/foundation/voc079-single-edge-invariants.test.mjs",
);

const smokeScript = readFileSync(smokeScriptPath, "utf8");
const deployProduction = readFileSync(deployProductionPath, "utf8");

const EXPECTED_PUBLIC_ROUTES = ["/", "/signin", "/auth/magic"];
const EXPECTED_AUTHENTICATED_ROUTES = [
  "/onboarding",
  "/home",
  "/discover",
  "/reviews",
  "/progress",
  "/settings",
  "/settings/account",
];

function extractBashArray(script, arrayName) {
  const start = script.indexOf(`${arrayName}=(`);
  assert.ok(start >= 0, `missing ${arrayName} declaration in smoke script`);
  const end = script.indexOf(")", start);
  assert.ok(end > start, `unterminated ${arrayName} declaration`);
  const block = script.slice(start, end + 1);
  const matches = [...block.matchAll(/"([^"]+)"/g)].map((match) => match[1]);
  return matches;
}

test("VOC-085-TEST-06: smoke script enumerates the ten fixed production web routes", () => {
  const publicRoutes = extractBashArray(
    smokeScript,
    "PRODUCTION_PUBLIC_WEB_ROUTES",
  );
  const authenticatedRoutes = extractBashArray(
    smokeScript,
    "PRODUCTION_AUTHENTICATED_WEB_ROUTES",
  );

  assert.deepEqual(
    publicRoutes,
    EXPECTED_PUBLIC_ROUTES,
    "public route inventory must match AC-05 exactly",
  );
  assert.deepEqual(
    authenticatedRoutes,
    EXPECTED_AUTHENTICATED_ROUTES,
    "authenticated route inventory must match AC-05 exactly",
  );
  assert.equal(
    publicRoutes.length + authenticatedRoutes.length,
    10,
    "route sweep must cover exactly ten fixed routes",
  );

  assert.match(
    smokeScript,
    /discover\/\$situation_slug/,
    "dynamic discover situation route must be derived from API content",
  );
  assert.match(
    smokeScript,
    /discover\/\$situation_slug\/\$word_slug/,
    "dynamic discover word route must be derived from API content",
  );
  assert.doesNotMatch(
    smokeScript,
    /discover\/airport/,
    "route sweep must not hard-code a single staging-only slug",
  );
});

test("VOC-085-TEST-07: route sweep fails closed on missing session cookie and uses auth for protected routes", () => {
  assert.match(
    smokeScript,
    /authenticated route sweep requires SMOKE_TEST_SESSION_COOKIE \(fail closed/,
    "missing cookie must fail the route sweep instead of skipping",
  );
  assert.match(
    smokeScript,
    /assert_web_route_reachable[\s\S]*require_auth/,
    "protected routes must distinguish sign-in redirects from in-app redirects",
  );
  assert.match(
    smokeScript,
    /redirected to sign-in/,
    "sign-in redirects on protected routes must fail the suite",
  );
  assert.doesNotMatch(
    smokeScript,
    /skip "authenticated route sweep/,
    "route sweep must not silently skip when cookie is missing on deploy path",
  );
});

test("VOC-085-TEST-10: production deploy still invokes the strengthened smoke suite on canonical :443", () => {
  assert.match(
    deployProduction,
    /bash infra\/scripts\/smoke-test-production\.sh/,
    "deploy-production must invoke the production smoke suite",
  );
  assert.match(
    deployProduction,
    /SMOKE_TEST_SESSION_COOKIE: \$\{\{ steps\.mint_production_session\.outputs\.session_cookie \}\}/,
    "deploy must pass the workflow-minted synthetic session into the smoke suite",
  );

  const smokeStepStart = deployProduction.indexOf(
    "name: Run production smoke-test suite",
  );
  assert.ok(smokeStepStart >= 0, "missing Run production smoke-test suite step");
  const smokeStepBlock = deployProduction.slice(smokeStepStart, smokeStepStart + 1200);
  assert.doesNotMatch(
    smokeStepBlock,
    /:8443/,
    "smoke step must target canonical HTTPS :443 hostnames",
  );
  assert.match(
    readFileSync(voc079InvariantsPath, "utf8"),
    /VOC-079-TEST-02/,
    "shared-edge / no 8081-8443 invariants remain covered by foundation tests",
  );
});
