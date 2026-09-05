import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  acceptSentenceEdit,
  countSentenceCharacters,
  MAX_SENTENCE_CHARACTERS,
} from "../../src/app/(app)/_components/sentence-feedback-input";

describe("sentence feedback character limit", () => {
  it("counts Unicode code points rather than UTF-16 code units", () => {
    const sentence = "I work 📝 every day.";

    assert.equal(countSentenceCharacters(sentence), 19);
    assert.equal(sentence.length, 20);
  });

  it("allows the API's full 300-code-point limit", () => {
    const sentence = "I work " + "📝".repeat(286) + " today.";

    assert.equal(countSentenceCharacters(sentence), MAX_SENTENCE_CHARACTERS);
    assert.equal(acceptSentenceEdit("", sentence), sentence);
  });

  it("preserves a full sentence when a middle insertion would exceed the limit", () => {
    const sentence = "I work " + "a".repeat(286) + " today.";
    const edited = sentence.slice(0, 20) + "📝" + sentence.slice(20);

    assert.equal(countSentenceCharacters(sentence), MAX_SENTENCE_CHARACTERS);
    assert.equal(acceptSentenceEdit(sentence, edited), sentence);
  });

  it("preserves a full sentence when a pasted value would exceed the limit", () => {
    const sentence = "I work " + "a".repeat(286) + " today.";
    const pasted = sentence + "📝";

    assert.equal(acceptSentenceEdit(sentence, pasted), sentence);
  });

  it("accepts normal edits below the limit", () => {
    const sentence = "I work today.";
    const edited = "I worked today.";

    assert.equal(acceptSentenceEdit(sentence, edited), edited);
  });
});
