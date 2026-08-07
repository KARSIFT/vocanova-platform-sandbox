import assert from "node:assert/strict";
import { afterEach, it } from "node:test";

import { createServerApiClient } from "../../src/lib/api-server";

const COOKIE_HEADER_GLOBAL_KEY = "__voc_test_cookie_header";
const MULTI_COOKIE_HEADER =
  "vocanova_session=s3cr3t-session-token-value-used-only-in-this-test; csrf_token=fake-csrf-token-value-used-only-in-this-test";
const originalApiBaseURL = process.env.API_BASE_URL;
const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  if (originalApiBaseURL === undefined) {
    delete process.env.API_BASE_URL;
  } else {
    process.env.API_BASE_URL = originalApiBaseURL;
  }
  delete (globalThis as Record<string, unknown>)[COOKIE_HEADER_GLOBAL_KEY];
});

it("forwards the raw incoming Cookie header byte-for-byte", async () => {
  process.env.API_BASE_URL = "http://api.internal:8080";
  (globalThis as Record<string, unknown>)[COOKIE_HEADER_GLOBAL_KEY] =
    MULTI_COOKIE_HEADER;

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

  assert.equal(forwardedCookieHeader, MULTI_COOKIE_HEADER);
});
