import assert from "node:assert/strict";
import { afterEach, beforeEach, test } from "node:test";

import {
  isNavigationPending,
  markNavigationDone,
  markNavigationStart,
} from "../../src/ui/navigation-state";

const globals = globalThis as unknown as { window?: { location: URL } };

function at(href: string) {
  globals.window = { location: new URL(href, "https://vocanova.invalid") };
}

beforeEach(() => markNavigationDone());
afterEach(() => {
  markNavigationDone();
  delete globals.window;
});

test("a link to another page starts the indicator", () => {
  at("/home");
  markNavigationStart("/progress");
  assert.equal(isNavigationPending(), true);
});

test("the same page, a hash or an equivalent query spelling starts nothing", () => {
  at("/vocabulary?q=keep+in+touch");
  for (const href of [
    "/vocabulary?q=keep+in+touch",
    "/vocabulary?q=keep%20in%20touch",
    "/vocabulary?q=keep+in+touch#results",
    "https://elsewhere.invalid/progress",
    undefined,
  ]) {
    markNavigationStart(href);
    assert.equal(isNavigationPending(), false, String(href));
  }
});

test("choosing the current page supersedes a waiting navigation", () => {
  at("/home");
  markNavigationStart("/courses");
  assert.equal(isNavigationPending(), true);
  markNavigationStart("/home");
  assert.equal(isNavigationPending(), false);
});

test("a different query on the same path is a navigation", () => {
  at("/discover?band=2");
  markNavigationStart("/discover?band=3#range-words");
  assert.equal(isNavigationPending(), true);
});
