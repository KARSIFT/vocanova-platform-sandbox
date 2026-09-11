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
