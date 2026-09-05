import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { isWordInSituation } from "../../src/app/(app)/discover/[situation]/[word]/_components/word-route";

describe("discover word-detail route membership (issue #1235)", () => {
  const cafeMeanings = [
    { wordSlug: "pour" },
    { wordSlug: "bill" },
  ];

  it("accepts a word listed by the requested Journey situation", () => {
    assert.equal(isWordInSituation(cafeMeanings, "pour"), true);
  });

  it("rejects a canonical word that is absent from the requested situation", () => {
    assert.equal(isWordInSituation(cafeMeanings, "boarding-pass"), false);
  });
});
