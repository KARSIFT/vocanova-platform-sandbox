import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { compareSentences } from "../../src/app/(app)/_components/sentence-comparison";

describe("exact sentence comparison", () => {
  it("preserves original and suggested text while highlighting word changes", () => {
    const result = compareSentences("I invite teh team.", "I invite the team.");
    assert.equal(
      result.original.map((part) => part.text).join(""),
      "I invite teh team.",
    );
    assert.equal(
      result.suggested.map((part) => part.text).join(""),
      "I invite the team.",
    );
    assert.deepEqual(
      result.original.filter((part) => part.changed),
      [{ text: "teh", changed: true }],
    );
    assert.deepEqual(
      result.suggested.filter((part) => part.changed),
      [{ text: "the", changed: true }],
    );
  });
  it("does not conceal capitalization, punctuation, or whitespace edits", () => {
    const result = compareSentences("i invite you!", "I invite you.");
    assert.equal(result.hasChanges, true);
    assert.deepEqual(
      result.original.filter((part) => part.changed).map((part) => part.text),
      ["i", "!"],
    );
    assert.deepEqual(
      result.suggested.filter((part) => part.changed).map((part) => part.text),
      ["I", "."],
    );
    const whitespace = compareSentences("I  invite you.", "I invite you.");
    assert.equal(whitespace.hasChanges, true);
    assert.equal(
      whitespace.original.map((part) => part.text).join(""),
      "I  invite you.",
    );
  });
  it("does not invent a correction when the sentences are identical", () => {
    const result = compareSentences("I invite you.", "I invite you.");
    assert.equal(result.hasChanges, false);
    assert.ok(result.original.every((part) => !part.changed));
    assert.ok(result.suggested.every((part) => !part.changed));
  });
  it("handles inserted or deleted text and Unicode without losing characters", () => {
    for (const [original, suggested] of [
      ["", "Hello!"],
      ["Hello!", ""],
      ["I invite Zoë 📝.", "I invited Zoë 📝."],
    ]) {
      const result = compareSentences(original!, suggested!);
      assert.equal(result.original.map((part) => part.text).join(""), original);
      assert.equal(
        result.suggested.map((part) => part.text).join(""),
        suggested,
      );
    }
  });
  it("bounds work for unusually long provider output and still preserves all text", () => {
    const original = "a ".repeat(1000);
    const suggested = "a ".repeat(1000) + "end";
    const result = compareSentences(original, suggested);
    assert.deepEqual(result.original, [{ text: original, changed: true }]);
    assert.deepEqual(result.suggested, [{ text: suggested, changed: true }]);
  });
});
