import assert from "node:assert/strict";
import { afterEach, it } from "node:test";

import { getOrRefreshCSRFToken } from "../../src/lib/csrf";

const originalDocument = globalThis.document;
const originalWindow = globalThis.window;
const originalFetch = globalThis.fetch;
let cookies = "";

function installBrowser(): void {
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: { location: { origin: "http://localhost:3000" } },
  });
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: {
      get cookie() {
        return cookies;
      },
      set cookie(value: string) {
        cookies = value;
      },
    },
  });
}

afterEach(() => {
  cookies = "";
  globalThis.fetch = originalFetch;
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: originalDocument,
  });
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: originalWindow,
  });
});

it("preserves an existing CSRF cookie without requesting /me", async () => {
  installBrowser();
  cookies = "vocanova_csrf=already-present";
  let requests = 0;
  globalThis.fetch = (() => {
    requests += 1;
    return Promise.resolve(Response.json({}));
  }) as typeof globalThis.fetch;

  assert.equal(await getOrRefreshCSRFToken(), "already-present");
  assert.equal(requests, 0);
});

it("shares one /me recovery request between concurrent callers", async () => {
  installBrowser();
  let requests = 0;
  let release: (() => void) | undefined;
  globalThis.fetch = (() => {
    requests += 1;
    return new Promise<Response>((resolve) => {
      release = () => {
        cookies = "vocanova_csrf=restored";
        resolve(Response.json({ onboardingStatus: "completed" }));
      };
    });
  }) as typeof globalThis.fetch;

  const first = getOrRefreshCSRFToken();
  const second = getOrRefreshCSRFToken();
  assert.equal(requests, 1);
  release?.();
  assert.deepEqual(await Promise.all([first, second]), ["restored", "restored"]);
});

it("keeps recovery shared until the browser exposes the refreshed cookie", async (t) => {
  installBrowser();
  t.mock.timers.enable({ apis: ["setTimeout"] });
  let requests = 0;
  globalThis.fetch = (() => {
    requests += 1;
    // Linux WebKit can finish /me while document.cookie still contains its
    // earlier missing-cookie snapshot. The browser exposes the new cookie on
    // a subsequent task, independently of the response body being readable.
    setTimeout(() => {
      cookies = "vocanova_csrf=restored";
    }, 0);
    return Promise.resolve(Response.json({ onboardingStatus: "completed" }));
  }) as typeof globalThis.fetch;

  const first = getOrRefreshCSRFToken();
  await new Promise<void>((resolve) => setImmediate(resolve));
  const second = getOrRefreshCSRFToken();
  t.mock.timers.tick(0);

  assert.deepEqual(await Promise.all([first, second]), ["restored", "restored"]);
  assert.equal(requests, 1);
});

it("returns no token when a successful recovery never supplies a cookie", async () => {
  installBrowser();
  let requests = 0;
  globalThis.fetch = (() => {
    requests += 1;
    return Promise.resolve(Response.json({ onboardingStatus: "completed" }));
  }) as typeof globalThis.fetch;

  assert.equal(await getOrRefreshCSRFToken(), null);
  assert.equal(requests, 1);
});

it("allows a later recovery after /me fails", async () => {
  installBrowser();
  let requests = 0;
  globalThis.fetch = (() => {
    requests += 1;
    if (requests === 1) return Promise.reject(new Error("Network unavailable"));
    cookies = "vocanova_csrf=restored";
    return Promise.resolve(Response.json({ onboardingStatus: "completed" }));
  }) as typeof globalThis.fetch;

  await assert.rejects(getOrRefreshCSRFToken(), /Network unavailable/);
  assert.equal(await getOrRefreshCSRFToken(), "restored");
  assert.equal(requests, 2);
});

it("treats a malformed CSRF cookie as missing and recovers it via /me", async () => {
  installBrowser();
  cookies = "vocanova_theme=dark; vocanova_csrf=%E0%A4%A";
  let requests = 0;
  globalThis.fetch = (() => {
    requests += 1;
    cookies = "vocanova_theme=dark; vocanova_csrf=restored";
    return Promise.resolve(Response.json({ onboardingStatus: "completed" }));
  }) as typeof globalThis.fetch;

  assert.equal(await getOrRefreshCSRFToken(), "restored");
  assert.equal(requests, 1);
});
