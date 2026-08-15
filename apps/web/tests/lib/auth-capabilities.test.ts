import assert from "node:assert/strict";
import { afterEach, it } from "node:test";

import { getSignInAuthCapabilities } from "../../src/lib/auth-capabilities";

const originalApiBaseURL = process.env.API_BASE_URL;
const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  if (originalApiBaseURL === undefined) {
    delete process.env.API_BASE_URL;
  } else {
    process.env.API_BASE_URL = originalApiBaseURL;
  }
});

it("returns oauthEnabled true when kill_switches.oauth_enabled is true", async () => {
  process.env.API_BASE_URL = "http://api.internal:8080";

  globalThis.fetch = ((url: string): Promise<Response> => {
    assert.equal(url, "http://api.internal:8080/healthz");
    return Promise.resolve(
      Response.json(
        {
          status: "ok",
          kill_switches: { oauth_enabled: true },
        },
        { status: 200 },
      ),
    );
  }) as typeof globalThis.fetch;

  const capabilities = await getSignInAuthCapabilities();
  assert.equal(capabilities.oauthEnabled, true);
});

it("returns oauthEnabled false when kill_switches.oauth_enabled is false", async () => {
  process.env.API_BASE_URL = "http://api.internal:8080";

  globalThis.fetch = ((url: string): Promise<Response> => {
    assert.equal(url, "http://api.internal:8080/healthz");
    return Promise.resolve(
      Response.json(
        {
          status: "ok",
          kill_switches: { oauth_enabled: false },
        },
        { status: 200 },
      ),
    );
  }) as typeof globalThis.fetch;

  const capabilities = await getSignInAuthCapabilities();
  assert.equal(capabilities.oauthEnabled, false);
});

it("returns oauthEnabled false when kill_switches is absent", async () => {
  process.env.API_BASE_URL = "http://api.internal:8080";

  globalThis.fetch = ((): Promise<Response> => {
    return Promise.resolve(Response.json({ status: "ok" }, { status: 200 }));
  }) as typeof globalThis.fetch;

  const capabilities = await getSignInAuthCapabilities();
  assert.equal(capabilities.oauthEnabled, false);
});

it("returns oauthEnabled false when /healthz is unavailable", async () => {
  process.env.API_BASE_URL = "http://api.internal:8080";

  globalThis.fetch = ((): Promise<Response> => {
    return Promise.reject(new Error("network unreachable"));
  }) as typeof globalThis.fetch;

  const capabilities = await getSignInAuthCapabilities();
  assert.equal(capabilities.oauthEnabled, false);
});

it("reads oauth_enabled from a 503 /healthz body without throwing", async () => {
  process.env.API_BASE_URL = "http://api.internal:8080";

  globalThis.fetch = ((url: string): Promise<Response> => {
    assert.equal(url, "http://api.internal:8080/healthz");
    return Promise.resolve(
      Response.json(
        {
          status: "unhealthy",
          database: "unhealthy",
          kill_switches: { oauth_enabled: true },
        },
        { status: 503 },
      ),
    );
  }) as typeof globalThis.fetch;

  const capabilities = await getSignInAuthCapabilities();
  assert.equal(capabilities.oauthEnabled, true);
});
