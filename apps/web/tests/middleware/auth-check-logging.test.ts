import assert from "node:assert/strict";
import { afterEach, beforeEach, describe, it } from "node:test";

import { NextRequest } from "next/server";

import { middleware } from "../../src/middleware";

const PROTECTED_ROUTE = "/home";
const PROTECTED_ROUTE_URL = `http://localhost:3000${PROTECTED_ROUTE}`;
const SESSION_COOKIE_NAME = "vocanova_session";
const SESSION_COOKIE_VALUE = "s3cr3t-session-token-value-used-only-in-this-test";
const SESSION_COOKIE_HEADER = `${SESSION_COOKIE_NAME}=${SESSION_COOKIE_VALUE}`;
const AUTH_CHECK_FAILURE_EVENT = "middleware_auth_check_failure";
const ALLOWED_LOG_FIELDS = new Set(["event", "category", "routePath", "status"]);
const SERVICE_UNAVAILABLE = 503;

interface AuthCheckFailureLog {
  event: string;
  category: string;
  routePath: string;
  status?: number;
}

const originalFetch = globalThis.fetch;
const originalConsoleError = console.error;
const originalApiBaseURL = process.env.API_BASE_URL;

let loggedLines: string[] = [];

function captureConsoleError(): void {
  loggedLines = [];
  console.error = (...args: unknown[]): void => {
    loggedLines.push(args.map((arg) => String(arg)).join(" "));
  };
}

function stubFetch(implementation: () => Promise<Response>): void {
  globalThis.fetch = implementation as typeof globalThis.fetch;
}

function requestWithSessionCookie(): NextRequest {
  return new NextRequest(PROTECTED_ROUTE_URL, {
    headers: { cookie: SESSION_COOKIE_HEADER },
  });
}

function singleFailureLine(): string {
  const [line] = loggedLines;
  assert.equal(
    loggedLines.length,
    1,
    `expected exactly one log line, got ${JSON.stringify(loggedLines)}`,
  );
  assert.ok(line);
  return line;
}

function assertRedirectsToSignIn(response: Response): void {
  assert.equal(response.status, 307);
  const location = response.headers.get("location");
  assert.ok(location, "expected a redirect Location header");
  assert.equal(new URL(location).pathname, "/signin");
}

function assertNoCredentialMaterial(line: string): void {
  assert.ok(
    !line.includes(SESSION_COOKIE_VALUE),
    "log line leaked the session cookie value",
  );
  assert.ok(
    !line.toLowerCase().includes("cookie"),
    "log line leaked cookie material",
  );
  for (const field of Object.keys(JSON.parse(line) as AuthCheckFailureLog)) {
    assert.ok(
      ALLOWED_LOG_FIELDS.has(field),
      `unexpected field "${field}" in auth-check failure log`,
    );
  }
}

describe("middleware auth-check failure logging", () => {
  beforeEach(() => {
    process.env.API_BASE_URL = "http://api.internal:8080";
    captureConsoleError();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    console.error = originalConsoleError;
    if (originalApiBaseURL === undefined) {
      delete process.env.API_BASE_URL;
    } else {
      process.env.API_BASE_URL = originalApiBaseURL;
    }
  });

  it("logs the fetch_threw category when the /api/v1/me fetch rejects", async () => {
    stubFetch(() => Promise.reject(new TypeError("fetch failed")));

    const response = await middleware(requestWithSessionCookie());

    const line = singleFailureLine();
    const log = JSON.parse(line) as AuthCheckFailureLog;
    assert.equal(log.event, AUTH_CHECK_FAILURE_EVENT);
    assert.equal(log.category, "fetch_threw");
    assert.equal(log.routePath, PROTECTED_ROUTE);
    assert.equal(log.status, undefined);
    assertNoCredentialMaterial(line);
    assertRedirectsToSignIn(response);
  });

  it("logs the unauthorized_401 category with its status when the API returns 401", async () => {
    stubFetch(() => Promise.resolve(new Response(null, { status: 401 })));

    const response = await middleware(requestWithSessionCookie());

    const line = singleFailureLine();
    const log = JSON.parse(line) as AuthCheckFailureLog;
    assert.equal(log.event, AUTH_CHECK_FAILURE_EVENT);
    assert.equal(log.category, "unauthorized_401");
    assert.equal(log.routePath, PROTECTED_ROUTE);
    assert.equal(log.status, 401);
    assertNoCredentialMaterial(line);
    assertRedirectsToSignIn(response);
  });

  it("logs the non_ok_response category with its status for any other non-ok status", async () => {
    stubFetch(() =>
      Promise.resolve(new Response(null, { status: SERVICE_UNAVAILABLE })),
    );

    const response = await middleware(requestWithSessionCookie());

    const line = singleFailureLine();
    const log = JSON.parse(line) as AuthCheckFailureLog;
    assert.equal(log.event, AUTH_CHECK_FAILURE_EVENT);
    assert.equal(log.category, "non_ok_response");
    assert.equal(log.routePath, PROTECTED_ROUTE);
    assert.equal(log.status, SERVICE_UNAVAILABLE);
    assertNoCredentialMaterial(line);
    assertRedirectsToSignIn(response);
  });

  it("logs nothing when the auth check succeeds", async () => {
    stubFetch(() =>
      Promise.resolve(
        Response.json({ onboardingStatus: "completed" }, { status: 200 }),
      ),
    );

    await middleware(
      new NextRequest("http://localhost:3000/onboarding", {
        headers: { cookie: SESSION_COOKIE_HEADER },
      }),
    );

    assert.deepEqual(loggedLines, []);
  });
});
