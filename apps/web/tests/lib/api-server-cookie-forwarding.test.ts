import assert from "node:assert/strict";
import { afterEach, it } from "node:test";

import { createServerApiClient } from "../../src/lib/api-server";

const COOKIE_HEADER_GLOBAL_KEY = "__voc_test_cookie_header";
const BUILD_ENVIRONMENT_KEY = "NEXT_PUBLIC_VOCANOVA_BUILD_ENVIRONMENT";
const LEGACY_SESSION = "vocanova_session=legacy-session-test-value";
const LEGACY_CSRF = "vocanova_csrf=legacy-csrf-test-value";
const STAGING_SESSION = "vocanova_staging_session=staging-session-test-value";
const STAGING_CSRF = "vocanova_staging_csrf=staging-csrf-test-value";
const PRODUCTION_SESSION =
  "vocanova_production_session=production-session-test-value";
const PRODUCTION_CSRF = "vocanova_production_csrf=production-csrf-test-value";
const THEME = "vocanova_theme=dark";
const UNRELATED = "csrf_token=fake-csrf-token-value-used-only-in-this-test";
// The shared parent cookie domain delivers every tier's cookies together.
const MIXED_COOKIE_HEADER = [
  PRODUCTION_CSRF,
  UNRELATED,
  STAGING_SESSION,
  LEGACY_SESSION,
  THEME,
  PRODUCTION_SESSION,
  LEGACY_CSRF,
  STAGING_CSRF,
].join("; ");
const originalApiBaseURL = process.env.API_BASE_URL;
const originalBuildEnvironment = process.env[BUILD_ENVIRONMENT_KEY];
const originalFetch = globalThis.fetch;

function setEnv(key: string, value: string | undefined): void {
  if (value === undefined) {
    delete process.env[key];
  } else {
    process.env[key] = value;
  }
}

afterEach(() => {
  globalThis.fetch = originalFetch;
  setEnv("API_BASE_URL", originalApiBaseURL);
  setEnv(BUILD_ENVIRONMENT_KEY, originalBuildEnvironment);
  delete (globalThis as Record<string, unknown>)[COOKIE_HEADER_GLOBAL_KEY];
});

async function forwardedCookieFor(
  buildEnvironment: string | undefined,
  incomingCookieHeader: string,
): Promise<string | null> {
  setEnv("API_BASE_URL", "http://api.internal:8080");
  setEnv(BUILD_ENVIRONMENT_KEY, buildEnvironment);
  (globalThis as Record<string, unknown>)[COOKIE_HEADER_GLOBAL_KEY] =
    incomingCookieHeader;

  let forwardedCookieHeader: string | null = null;
  globalThis.fetch = ((url: string, init?: RequestInit): Promise<Response> => {
    assert.equal(url, "http://api.internal:8080/api/v1/me");
    forwardedCookieHeader = new Headers(init?.headers).get("Cookie");
    return Promise.resolve(
      Response.json({ onboardingStatus: "completed" }, { status: 200 }),
    );
  }) as typeof globalThis.fetch;

  const client = await createServerApiClient();
  await client.getCurrentUser();
  return forwardedCookieHeader;
}

it("forwards only the production tier's own cookies in production", async () => {
  assert.equal(
    await forwardedCookieFor("production", MIXED_COOKIE_HEADER),
    [PRODUCTION_CSRF, THEME, PRODUCTION_SESSION].join("; "),
  );
});

it("forwards only the staging tier's own cookies in staging", async () => {
  assert.equal(
    await forwardedCookieFor("staging", MIXED_COOKIE_HEADER),
    [STAGING_SESSION, THEME, STAGING_CSRF].join("; "),
  );
});

it("preserves local fixture context when the build environment is unknown", async () => {
  assert.equal(
    await forwardedCookieFor(undefined, MIXED_COOKIE_HEADER),
    MIXED_COOKIE_HEADER,
  );
});

it("sends no Cookie header when only foreign or legacy cookies arrive", async () => {
  const foreignOnly = [
    PRODUCTION_SESSION,
    PRODUCTION_CSRF,
    LEGACY_SESSION,
    LEGACY_CSRF,
    UNRELATED,
  ].join("; ");
  assert.equal(await forwardedCookieFor("staging", foreignOnly), null);
});
