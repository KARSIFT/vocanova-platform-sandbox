import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  countSentenceCharacters,
  limitSentenceCharacters,
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
    assert.equal(limitSentenceCharacters(sentence), sentence);
  });

  it("truncates only after the 300th code point", () => {
    const sentence = "I work " + "📝".repeat(293) + " today.";

    const limited = limitSentenceCharacters(sentence);
    assert.equal(countSentenceCharacters(limited), MAX_SENTENCE_CHARACTERS);
    assert.equal(limited, Array.from(sentence).slice(0, 300).join(""));
  });
});
