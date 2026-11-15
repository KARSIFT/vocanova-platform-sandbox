import assert from "node:assert/strict";
import { afterEach, beforeEach, describe, it } from "node:test";
import { NextRequest } from "next/server";
import { config, middleware } from "../../src/middleware";

const originalFetch = globalThis.fetch;
const originalError = console.error;
const originalURL = process.env.API_BASE_URL;
const request = (session = true) => new NextRequest("https://production.example.test/", {
  headers: session ? { cookie: "vocanova_session=synthetic-root-session" } : {},
});

describe("authenticated public root entry", () => {
  beforeEach(() => {
    process.env.API_BASE_URL = "http://api.internal:8180";
    console.error = () => {};
  });
  afterEach(() => {
    globalThis.fetch = originalFetch;
    console.error = originalError;
    if (originalURL === undefined) delete process.env.API_BASE_URL;
    else process.env.API_BASE_URL = originalURL;
  });

  it("includes the root in identity middleware matching", () => {
    assert.ok(config.matcher.includes("/"));
  });
  it("keeps anonymous marketing available without an identity request", async () => {
    let calls = 0;
    globalThis.fetch = async () => { calls++; throw new Error("unexpected identity read"); };
    const response = await middleware(request(false));
    assert.equal(response.status, 200);
    assert.equal(response.headers.get("x-middleware-next"), "1");
    assert.equal(calls, 0);
  });
  it("redirects a verified learner to Home with private no-store caching", async () => {
    globalThis.fetch = async () => Response.json({ onboardingStatus: "completed" });
    const response = await middleware(request());
    assert.equal(response.headers.get("location"), "https://production.example.test/home");
    assert.equal(response.headers.get("cache-control"), "private, no-store");
  });
  it("routes unfinished onboarding to setup", async () => {
    globalThis.fetch = async () => Response.json({ onboardingStatus: "in_progress" });
    const response = await middleware(request());
    assert.equal(response.headers.get("location"), "https://production.example.test/onboarding");
  });
  it("keeps root public for an expired cookie without a redirect loop", async () => {
    globalThis.fetch = async () => new Response(null, { status: 401 });
    const response = await middleware(request());
    assert.equal(response.status, 200);
    assert.equal(response.headers.get("location"), null);
    assert.match(response.headers.get("cache-control") ?? "", /no-store/);
  });
  it("shows recoverable failure when identity cannot be checked", async () => {
    globalThis.fetch = async () => new Response(null, { status: 503 });
    const response = await middleware(request());
    assert.equal(response.status, 503);
    assert.equal(response.headers.get("location"), null);
    assert.equal(response.headers.get("cache-control"), "no-store");
    assert.ok((await response.text()).includes('href="/home"'));
  });
});
