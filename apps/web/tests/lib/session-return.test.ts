import assert from "node:assert/strict";
import { afterEach, beforeEach, describe, it } from "node:test";
import { handleSessionExpired } from "../../src/lib/session";
import { normalizeReturnTo } from "../../src/lib/return-to";

const previousWindow = Object.getOwnPropertyDescriptor(globalThis, "window");
const previousDocument = Object.getOwnPropertyDescriptor(globalThis, "document");
const location = { pathname: "/words/one", search: "?meaning=first", hash: "#sentence-practice", href: "" };
const browserDocument = { cookie: "vocanova_csrf=test" };

describe("session expiry destination", () => {
  beforeEach(() => {
    location.href = "";
    browserDocument.cookie = "vocanova_csrf=test";
    Object.defineProperty(globalThis, "window", { configurable: true, value: { location } });
    Object.defineProperty(globalThis, "document", { configurable: true, value: browserDocument });
  });
  afterEach(() => {
    if (previousWindow) Object.defineProperty(globalThis, "window", previousWindow);
    else Reflect.deleteProperty(globalThis, "window");
    if (previousDocument) Object.defineProperty(globalThis, "document", previousDocument);
    else Reflect.deleteProperty(globalThis, "document");
  });
  it("returns to the selected writing section after sign-in", () => {
    handleSessionExpired();
    const redirect = new URL(location.href, "https://vocanova.invalid");
    assert.equal(redirect.pathname, "/login");
    assert.equal(redirect.searchParams.get("reason"), "session-expired");
    const destination = redirect.searchParams.get("returnTo");
    assert.equal(destination, "/words/one?meaning=first#sentence-practice");
    assert.equal(normalizeReturnTo(destination), destination);
    assert.match(browserDocument.cookie, /vocanova_csrf=; Max-Age=0/);
  });
  it("keeps an explicit caller destination authoritative", () => {
    handleSessionExpired("/review");
    const redirect = new URL(location.href, "https://vocanova.invalid");
    assert.equal(redirect.searchParams.get("returnTo"), "/review");
  });
});
