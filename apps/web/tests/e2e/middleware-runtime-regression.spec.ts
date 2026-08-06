// VOC-039-T01 regression test for issue #297.
//
// The bug: `src/middleware.ts` runs on the Edge runtime by default, and
// the Edge runtime cannot see runtime-only server environment variables
// (`API_BASE_URL` is supplied to the container at start time, not
// inlined into the Edge bundle at build time). `getApiBaseURL()`
// therefore falls back to `http://localhost:8080` inside the Edge
// sandbox, the `/api/v1/me` auth check never reaches the deployed API,
// and every authenticated learner is redirected back to `/signin`.
//
// This test executes the real, unmodified middleware source inside the
// real Edge sandbox Next.js ships, and inside Node, against a stub API
// on an ephemeral port that is only discoverable through the runtime
// environment. It therefore fails before VOC-039-T00's
// `export const runtime = "nodejs"` and passes after it, for the same
// reason production failed - not because of a source-text check.
//
// The assertions run in the Playwright worker's Node process, so no
// browser or app server is involved; the test is registered here
// because `apps/web/tests/e2e` is the suite this repo's CI already runs
// (`pnpm --filter @vocanova/web test:e2e`).

import { expect, test } from "@playwright/test";

import {
  CURRENT_USER_PATH,
  DEFAULT_MIDDLEWARE_RUNTIME,
  SESSION_COOKIE_NAME,
  readDeclaredMiddlewareRuntime,
  runMiddlewareOnDeclaredRuntime,
  runMiddlewareOnEdgeRuntime,
  startStubApiServer,
  type StubApiServer,
} from "./middleware-runtime-harness";

const DESKTOP_PROJECT_NAME = "home-desktop-1280";
const PROTECTED_ROUTE_URL = "https://app.vocanova.test/home";
const SIGN_IN_PATH = "/signin";
const VALID_SESSION_COOKIE = `${SESSION_COOKIE_NAME}=valid-session-for-regression-test`;

test.describe("Middleware runtime regression (VOC-039-T01)", () => {
  let stubApi: StubApiServer;

  test.beforeEach(async ({}, testInfo) => {
    test.skip(
      testInfo.project.name !== DESKTOP_PROJECT_NAME,
      "Runtime behavior is viewport-independent; run once on the representative desktop project.",
    );
    stubApi = await startStubApiServer();
  });

  test.afterEach(async () => {
    await stubApi?.close();
  });

  // Characterization half: proves the harness really does exercise the
  // Edge runtime constraint this issue was about. If a future Next.js
  // version starts exposing runtime environment variables to Edge
  // middleware, this test fails and the regression gate below stops
  // being meaningful - which is exactly when someone should re-examine
  // it, rather than trusting a gate that silently no longer bites.
  test("Edge runtime cannot reach the deployed API, reproducing issue #297", async () => {
    await runMiddlewareOnEdgeRuntime({
      url: PROTECTED_ROUTE_URL,
      cookieHeader: VALID_SESSION_COOKIE,
      apiBaseURL: stubApi.baseURL,
    });

    // The assertion is about which origin the auth check reached, not
    // about the redirect it returned: under Edge the call goes to
    // `getApiBaseURL()`'s `http://localhost:8080` fallback, which on a
    // developer machine and in this repo's own e2e run happens to be
    // the mock API server - which is exactly why the bug stayed
    // invisible outside production.
    expect(
      stubApi.receivedPaths,
      "Edge middleware is expected NOT to reach the runtime-configured API origin; " +
        "if it did, the Edge/Node gap issue #297 identified no longer exists and this suite needs revisiting.",
    ).toEqual([]);
  });

  // Regression gate: this is the assertion that fails against a
  // pre-VOC-039-T00 `middleware.ts` (which runs on Edge, never reaches
  // the API, and redirects an authenticated learner to /signin) and
  // passes once the module declares the Node.js runtime.
  test("declared runtime lets the auth check reach the API and admits an authenticated learner", async () => {
    const declaredRuntime = await readDeclaredMiddlewareRuntime();

    const outcome = await runMiddlewareOnDeclaredRuntime({
      url: PROTECTED_ROUTE_URL,
      cookieHeader: VALID_SESSION_COOKIE,
      apiBaseURL: stubApi.baseURL,
    });

    expect(
      stubApi.receivedPaths,
      `middleware.ts declares the "${declaredRuntime}" runtime (default is "${DEFAULT_MIDDLEWARE_RUNTIME}"). ` +
        `Under that runtime the auth check never reached ${CURRENT_USER_PATH} at the deployed API, ` +
        "which is exactly the production failure in issue #297.",
    ).toEqual([CURRENT_USER_PATH]);

    expect(
      outcome,
      "An authenticated learner with a completed onboarding status must be admitted to the protected route, not bounced to " +
        SIGN_IN_PATH,
    ).toEqual({ kind: "next" });
  });
});
