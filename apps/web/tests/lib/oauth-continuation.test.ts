import assert from "node:assert/strict";
import { afterEach, describe, it } from "node:test";

import {
  consumeOAuthContinuation,
  rememberOAuthContinuation,
} from "../../src/lib/oauth-continuation";

const storage = new Map<string, string>();
const originalWindow = globalThis.window;
const originalDateNow = Date.now;

function installBrowserStorage() {
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: {
      sessionStorage: {
        getItem: (key: string) => storage.get(key) ?? null,
        removeItem: (key: string) => storage.delete(key),
        setItem: (key: string, value: string) => storage.set(key, value),
      },
    },
  });
}

afterEach(() => {
  storage.clear();
  Date.now = originalDateNow;
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: originalWindow,
  });
});

describe("OAuth continuation", () => {
  it("continues a safe route once after the allowed Home callback", () => {
    installBrowserStorage();
    rememberOAuthContinuation("/reviews?mode=due");
    assert.equal(consumeOAuthContinuation(), "/reviews?mode=due");
    assert.equal(consumeOAuthContinuation(), null);
  });

  it("does not continue to an external destination", () => {
    installBrowserStorage();
    rememberOAuthContinuation("https://evil.example");
    assert.equal(consumeOAuthContinuation(), "/home");
  });

  it("drops a continuation that has expired or was created in the future", () => {
    installBrowserStorage();
    Date.now = () => 1_000;
    rememberOAuthContinuation("/reviews");
    Date.now = () => 16 * 60 * 1_000 + 1_001;
    assert.equal(consumeOAuthContinuation(), null);

    Date.now = () => 2_000;
    rememberOAuthContinuation("/settings");
    Date.now = () => 1_999;
    assert.equal(consumeOAuthContinuation(), null);
  });

  it("does not persist an email-change URL for OAuth continuation", () => {
    installBrowserStorage();
    rememberOAuthContinuation("/auth/email-change?token=not-stored-as-a-destination");
    assert.equal(consumeOAuthContinuation(), "/home");
  });
});
