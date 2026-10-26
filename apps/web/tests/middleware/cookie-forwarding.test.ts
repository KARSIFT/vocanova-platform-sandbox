import assert from "node:assert/strict";
import { afterEach, it } from "node:test";

import { NextRequest } from "next/server";

import { middleware } from "../../src/middleware";

const BUILD_ENVIRONMENT_KEY = "NEXT_PUBLIC_VOCANOVA_BUILD_ENVIRONMENT";
// The shared parent cookie domain delivers every tier's cookies together.
const INCOMING_COOKIE_HEADER = [
  "vocanova_production_session=production-session-test-value",
  "vocanova_session=legacy-session-test-value",
  "vocanova_staging_session=staging-session-test-value",
  "vocanova_theme=light",
  "vocanova_production_csrf=production-csrf-test-value",
  "vocanova_staging_csrf=staging-csrf-test-value",
].join("; ");

const originalFetch = globalThis.fetch;
const originalApiBaseURL = process.env.API_BASE_URL;
const originalBuildEnvironment = process.env[BUILD_ENVIRONMENT_KEY];

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
});

async function forwardedCookieFor(
  buildEnvironment: string | undefined,
): Promise<string | null> {
  setEnv("API_BASE_URL", "http://api.internal:8080");
  setEnv(BUILD_ENVIRONMENT_KEY, buildEnvironment);
  let forwardedCookieHeader: string | null = null;
  globalThis.fetch = ((_url: string, init?: RequestInit): Promise<Response> => {
    forwardedCookieHeader = new Headers(init?.headers).get("Cookie");
    return Promise.resolve(Response.json({ onboardingStatus: "completed" }));
  }) as typeof globalThis.fetch;

  const response = await middleware(
    new NextRequest("http://localhost:3000/home", {
      headers: { cookie: INCOMING_COOKIE_HEADER },
    }),
  );
  assert.equal(response.status, 200);
  return forwardedCookieHeader;
}

it("forwards only staging credentials and the theme to /me in staging", async () => {
  assert.equal(
    await forwardedCookieFor("staging"),
    "vocanova_staging_session=staging-session-test-value; vocanova_theme=light; vocanova_staging_csrf=staging-csrf-test-value",
  );
});

it("forwards only production credentials and the theme to /me in production", async () => {
  assert.equal(
    await forwardedCookieFor("production"),
    "vocanova_production_session=production-session-test-value; vocanova_theme=light; vocanova_production_csrf=production-csrf-test-value",
  );
});

it("preserves local fixture context when the build environment is unknown", async () => {
  assert.equal(
    await forwardedCookieFor(undefined),
    INCOMING_COOKIE_HEADER,
  );
});
