import assert from "node:assert/strict";
import test from "node:test";

import { wordDifficultyLabel } from "../../src/app/(app)/_components/word-difficulty-label";

test("word details omit unavailable difficulty metadata instead of an UNKNOWN badge", () => {
  for (const value of [
    undefined,
    null,
    "",
    "unknown",
    "UNKNOWN",
    " unknown ",
    " ",
  ]) {
    assert.equal(
      wordDifficultyLabel(value),
      null,
      `unavailable level: ${value}`,
    );
  }
});

test("word details preserve real levels and existing authored level bands", () => {
  for (const level of ["a1", "a2", "b1", "b2", "c1", "c2"]) {
    assert.equal(wordDifficultyLabel(level), level.toUpperCase());
  }
  assert.equal(wordDifficultyLabel("B1"), "B1");
  assert.equal(wordDifficultyLabel("a2_b1"), "A2–B1");
  assert.equal(wordDifficultyLabel("mixed"), "Mixed levels");
});
