import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { describe, it } from "node:test";

import { getConversationPractice } from "../../src/app/(app)/discover/[situation]/_components/conversation-context-content";

interface Seed {
  journey_situations: { id: string; slug: string; status: string }[];
  journey_words: { journey_situation_id: string; meaning_id: string }[];
  word_meanings: { id: string; word_id: string; status: string }[];
  canonical_words: {
    id: string;
    text: string;
    normalized_text: string;
    status: string;
  }[];
}

const seed = JSON.parse(
  readFileSync(
    new URL("../../../api/cmd/seed/voc026-p1.json", import.meta.url),
    "utf8",
  ),
) as Seed;
const situation = seed.journey_situations.find(
  (item) => item.slug === "daily-conversation" && item.status === "active",
)!;
const meanings = seed.journey_words
  .filter((item) => item.journey_situation_id === situation.id)
  .map((item) => {
    const meaning = seed.word_meanings.find(
      (entry) => entry.id === item.meaning_id && entry.status === "active",
    )!;
    const word = seed.canonical_words.find(
      (entry) => entry.id === meaning.word_id && entry.status === "active",
    )!;
    return Object.freeze({
      meaningId: meaning.id,
      wordText: word.text,
      // The selected English references use spaces only; verify the same
      // canonical normalized-text slug contract used by the content API.
      wordSlug: word.normalized_text.replaceAll(" ", "-"),
    });
  });
Object.freeze(meanings);

describe("Daily Conversation context content", () => {
  it("resolves three reviewed contrasts from active canonical situation meanings", () => {
    const cases = getConversationPractice(situation.id, meanings);
    assert.equal(cases.length, 3);
    assert.equal(new Set(cases.map((item) => item.id)).size, 3);
    assert.deepEqual(
      cases.map(
        (item) =>
          item.choices.find(
            (choice) => choice.meaningId === item.correctMeaningId,
          )!.wordText,
      ),
      ["invite", "sounds good", "reschedule"],
    );
    assert.deepEqual(
      cases.map((item) => item.choices.map((choice) => choice.wordText)),
      [
        ["invite", "join"],
        ["keep in touch", "sounds good"],
        ["reschedule", "cancel"],
      ],
    );
    for (const item of cases) {
      assert.ok(item.context.trim());
      assert.ok(item.prompt.trim());
      for (const choice of item.choices) {
        const meaning = meanings.find(
          (entry) => entry.meaningId === choice.meaningId,
        )!;
        assert.equal(choice.wordText, meaning.wordText);
        assert.equal(choice.wordSlug, meaning.wordSlug);
        assert.ok(choice.explanation.trim());
      }
    }
  });

  it("omits the optional activity on other or empty situations", () => {
    assert.deepEqual(getConversationPractice("other-situation", meanings), []);
    assert.deepEqual(getConversationPractice(situation.id, []), []);
  });

  it("requires each authored reference to remain available exactly once", () => {
    const cases = getConversationPractice(situation.id, meanings);
    for (const reference of cases.flatMap((item) => item.choices)) {
      assert.deepEqual(
        getConversationPractice(
          situation.id,
          meanings.filter((item) => item.meaningId !== reference.meaningId),
        ),
        [],
      );
      assert.deepEqual(
        getConversationPractice(situation.id, [...meanings, reference]),
        [],
      );
    }
  });

  it("does not offer a reference whose canonical text or route changed", () => {
    const target = meanings.find((item) => item.wordText === "invite")!;
    for (const replacement of [
      { ...target, wordText: "invitation" },
      { ...target, wordSlug: "different-sense" },
    ]) {
      assert.deepEqual(
        getConversationPractice(
          situation.id,
          meanings.map((item) =>
            item.meaningId === target.meaningId ? replacement : item,
          ),
        ),
        [],
      );
    }
  });
});
