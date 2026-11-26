import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { normalizeReturnTo } from "../../src/lib/return-to";
import {
  vocabularySearchReturnHref,
  vocabularyWordHref,
} from "../../src/lib/vocabulary-return";

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

  it("does not return a learner to a public auth route after sign-in", () => {
    for (const value of ["/login", "/signin", "/magic-link", "/auth/magic"]) {
      assert.equal(normalizeReturnTo(value), "/home", value);
    }
  });
});

describe("Vocabulary search return", () => {
  it("preserves a filtered cursor with a query and exact meaning", () => {
    const search = `/vocabulary?${new URLSearchParams({
      q: "café + tea?",
      knowledge: "saved",
      view: "map",
      after: "cursor+/",
    })}`;
    const detail = new URL(
      vocabularyWordHref("take-out", "meaning-id", search),
      "https://vocanova.invalid",
    );
    assert.equal(vocabularySearchReturnHref(detail.searchParams.get("fromSearch")), search);
    assert.equal(detail.hash, "#meaning-id");
  });

  it("accepts only bounded destinations on the Vocabulary search page", () => {
    for (const value of [
      "/vocabulary",
      "/vocabulary/check?q=tea",
      "/vocabulary?q=tea#result",
      "//another.example/vocabulary?q=tea",
      "https://another.example/vocabulary?q=tea",
      ["/vocabulary?q=tea"],
      `/vocabulary?q=${"a".repeat(4096)}`,
    ]) {
      assert.equal(vocabularySearchReturnHref(value), null);
    }
  });
});
