import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { ApiResponseError, VocanovaClient } from "./index.js";

describe("VocanovaClient", () => {
  it("sends GET /api/v1/me with Accept header", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/me");
      assert.equal(init.method, "GET");
      assert.equal(new Headers(init.headers).get("Accept"), "application/json");
      return Promise.resolve(
        new Response(JSON.stringify({ email: "user@example.com" }), {
          headers: { "Content-Type": "application/json" },
          status: 200,
        }),
      );
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { data } = await client.getCurrentUser();
    assert.equal(data.email, "user@example.com");
  });

  it("sends POST /api/v1/auth/magic-links with JSON body", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/auth/magic-links");
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("Content-Type"),
        "application/json",
      );
      assert.equal(new Headers(init.headers).get("Accept"), "application/json");
      assert.equal(init.body, JSON.stringify({ email: "user@example.com" }));
      return Promise.resolve(new Response(null, { status: 204 }));
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.requestMagicLink({
      email: "user@example.com",
    });
    assert.equal(response.status, 204);
  });

  it("sends CSRF header on logout", async () => {
    const fetch = (url: string, init: RequestInit): Promise<Response> => {
      assert.equal(url, "https://api.example.com/api/v1/auth/logout");
      assert.equal(init.method, "POST");
      assert.equal(
        new Headers(init.headers).get("X-CSRF-Token"),
        "csrf-token-value",
      );
      return Promise.resolve(new Response(null, { status: 204 }));
    };

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    const { response } = await client.logout({
      headers: { "X-CSRF-Token": "csrf-token-value" },
    });
    assert.equal(response.status, 204);
  });

  it("throws ApiResponseError for problem+json errors", async () => {
    const fetch = (): Promise<Response> =>
      Promise.resolve(
        new Response(JSON.stringify({ detail: "authentication required" }), {
          headers: { "Content-Type": "application/problem+json" },
          status: 401,
        }),
      );

    const client = new VocanovaClient({
      baseURL: "https://api.example.com",
      fetch: fetch as typeof globalThis.fetch,
    });
    await assert.rejects(client.getCurrentUser(), (error: unknown) => {
      if (!(error instanceof ApiResponseError)) {
        return false;
      }
      assert.equal(error.status, 401);
      assert.equal(error.message, "authentication required");
      return true;
    });
  });
});
