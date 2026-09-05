import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { normalizeReturnTo } from "../../src/lib/return-to";

describe("normalizeReturnTo", () => {
  it("preserves an app-relative destination with its query", () => {
    assert.equal(normalizeReturnTo("/reviews?mode=due"), "/reviews?mode=due");
  });

  it("falls back to home for external and protocol-relative destinations", () => {
    assert.equal(normalizeReturnTo("https://evil.example/reviews"), "/home");
    assert.equal(normalizeReturnTo("//evil.example/reviews"), "/home");
    assert.equal(normalizeReturnTo("/\\evil.example/reviews"), "/home");
  });

  it("falls back to home when the link has no destination", () => {
    assert.equal(normalizeReturnTo(null), "/home");
  });

  it("rejects paths that normalize into protocol-relative destinations", () => {
    for (const value of [
      "/safe/..//evil.example",
      "/%2e//evil.example",
      "https://vocanova.invalid//evil.example",
    ]) {
      assert.equal(normalizeReturnTo(value), "/home", value);
    }
  });
});
