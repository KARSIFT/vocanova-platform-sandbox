// Guards against a specific regression class this repo already hit once:
// apps/web/tests/e2e/mock-api-server.mjs (the E2E test double for the real
// API) had a genuine CORS bug diagnosed and fixed in its own history (see
// its comment above the Access-Control-Allow-Origin block), but that fix
// was never propagated to the real apps/api server - so the entire E2E
// suite exercised realistic cross-origin behavior against a server that
// only exists in tests, while the real deployed API had none, for the
// whole lifetime of the project (found live 2026-08-02, VOC-037-EV-03).
//
// This is not a behavioral-equivalence test (the mock's CORS is
// intentionally more permissive - it echoes any Origin, since a test
// double has no real security boundary to enforce; the real API's is a
// strict founder-configured allowlist, see apps/api/app/api/cors.go).
// It only checks that CORS *support exists and stays wired* on both
// sides, so a future accidental removal on the real side is caught here
// instead of silently reintroducing the exact bug this test was written
// after.

import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "node:test";
import assert from "node:assert/strict";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

function read(relativePath) {
  return readFileSync(path.join(repositoryRoot, relativePath), "utf8");
}

test("VOC-037-EV-03 real API: CORS middleware exists and is wired into the live server", () => {
  const corsSource = read("apps/api/app/api/cors.go");
  assert.match(
    corsSource,
    /Access-Control-Allow-Origin/,
    "apps/api/app/api/cors.go must set Access-Control-Allow-Origin",
  );
  assert.match(
    corsSource,
    /func corsMiddleware\(/,
    "apps/api/app/api/cors.go must define corsMiddleware",
  );

  const productionSource = read("apps/api/app/api/production.go");
  assert.match(
    productionSource,
    /mux\.Use\(corsMiddleware\(/,
    "apps/api/app/api/production.go's NewProductionAPI must wire corsMiddleware onto the live mux - " +
      "if this line is missing, every credentialed cross-origin browser request silently fails again " +
      "(the exact regression this test exists to catch)",
  );
});

test("VOC-037-EV-03 E2E mock: CORS support still present (keeps the test double realistic)", () => {
  const mockSource = read("apps/web/tests/e2e/mock-api-server.mjs");
  assert.match(
    mockSource,
    /Access-Control-Allow-Origin/,
    "apps/web/tests/e2e/mock-api-server.mjs must set Access-Control-Allow-Origin, or E2E tests stop " +
      "exercising realistic cross-origin behavior and this whole regression class goes uncaught again",
  );
});
