import assert from "node:assert/strict";
import { afterEach, beforeEach, describe, it } from "node:test";
import { handleSessionExpired } from "../../src/lib/session";
import { normalizeReturnTo } from "../../src/lib/return-to";
import {
  getSessionExit,
  replaceSessionLocation,
  SESSION_LOCATION_CHANGED,
} from "../../src/app/(app)/_components/session-route";

const previousWindow = Object.getOwnPropertyDescriptor(globalThis, "window");
const previousDocument = Object.getOwnPropertyDescriptor(
  globalThis,
  "document",
);
const location = {
  pathname: "/words/one",
  search: "?meaning=first",
  hash: "#sentence-practice",
  href: "",
};
const browserDocument = { cookie: "vocanova_csrf=test" };

describe("session expiry destination", () => {
  beforeEach(() => {
    location.href = "";
    browserDocument.cookie = "vocanova_csrf=test";
    Object.defineProperty(globalThis, "window", {
      configurable: true,
      value: { location },
    });
    Object.defineProperty(globalThis, "document", {
      configurable: true,
      value: browserDocument,
    });
  });
  afterEach(() => {
    if (previousWindow)
      Object.defineProperty(globalThis, "window", previousWindow);
    else Reflect.deleteProperty(globalThis, "window");
    if (previousDocument)
      Object.defineProperty(globalThis, "document", previousDocument);
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

describe("focused session exit destinations", () => {
  it("returns discovery to its exact map range and anchor", () => {
    for (const band of [1, 41, 100])
      assert.deepEqual(
        getSessionExit("/discovery", `?band=${band}&word=4001`),
        {
          href: `/knowledge-map?band=${band}#range-words`,
          label: "Back to word map",
        },
      );
    assert.equal(
      getSessionExit("/discovery", new URLSearchParams("band=42"))?.href,
      "/knowledge-map?band=42#range-words",
    );
  });
  it("uses the first range for missing, duplicate or invalid bands", () => {
    for (const query of [
      "",
      "band=",
      "band=0",
      "band=-1",
      "band=101",
      "band=1.5",
      "band=01",
      "band=41&band=42",
      "band=https://other.invalid",
      "band=%3Cscript%3E",
    ])
      assert.equal(
        getSessionExit("/discovery", query)?.href,
        "/knowledge-map?band=1#range-words",
        query,
      );
  });
  it("preserves pathname-only session detection and other exits", () => {
    assert.ok(getSessionExit("/discovery"));
    assert.equal(getSessionExit("/home", "band=41"), undefined);
    assert.deepEqual(getSessionExit("/practice/session/one", "band=41"), {
      href: "/practice",
      label: "Back to Practice",
    });
    assert.deepEqual(getSessionExit("/review/session", "band=41"), {
      href: "/home",
      label: "Back to Home",
    });
  });
});

it("canonical session location preserves opaque history state and notifies readers", () => {
  const original = Object.getOwnPropertyDescriptor(globalThis, "window");
  const opaque = { key: "keep", nested: { revision: 4 } };
  const writes: { value: unknown; href: string }[] = [];
  const events: string[] = [];
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: {
      history: {
        state: opaque,
        replaceState: (value: unknown, _unused: string, href: string) =>
          writes.push({ value, href }),
      },
      dispatchEvent: (event: Event) => {
        events.push(event.type);
        return true;
      },
    },
  });
  try {
    replaceSessionLocation("/discovery?band=41&word=4001");
    assert.equal(writes.length, 1);
    assert.strictEqual(writes[0]?.value, opaque);
    assert.equal(writes[0]?.href, "/discovery?band=41&word=4001");
    assert.deepEqual(events, [SESSION_LOCATION_CHANGED]);
  } finally {
    if (original) Object.defineProperty(globalThis, "window", original);
    else Reflect.deleteProperty(globalThis, "window");
  }
});
