import assert from "node:assert/strict";
import { describe, it } from "node:test";

import type { DueWord } from "@vocanova/api-client";

import { buildMultipleChoiceOptions } from "../../src/app/(app)/reviews/_components/review-session-options";

function card(index: number): DueWord {
  return {
    userWordId: `saved-${index}`,
    meaningId: `meaning-${index}`,
    wordId: `word-${index}`,
    wordSlug: `word-${index}`,
    wordText: `Word ${index}`,
    partOfSpeech: "noun",
    shortDefinition: `Definition ${index}`,
    status: "due",
    reviewStep: 0,
  };
}

describe("review multiple-choice options", () => {
  it("varies the same card's answer position across separate session seeds", () => {
    const dueWords = Array.from({ length: 4 }, (_, index) => card(index));
    const positions = new Set<number>();
    for (let index = 0; index < 64; index++) {
      const options = buildMultipleChoiceOptions(
        dueWords,
        0,
        `session-${index}`,
      );
      positions.add(
        options.findIndex(
          (option) => option.meaningId === dueWords[0]!.meaningId,
        ),
      );
    }
    assert.equal(positions.size, 4);
  });

  it("keeps server and client order identical despite different random sources", (t) => {
    const dueWords = Array.from({ length: 4 }, (_, index) => card(index));
    const random = t.mock.method(Math, "random", () => 0);
    const serverOptions = buildMultipleChoiceOptions(dueWords, 0, "session-a");
    random.mock.mockImplementation(() => 0.999999);
    const clientOptions = buildMultipleChoiceOptions(
      JSON.parse(JSON.stringify(dueWords)) as DueWord[],
      0,
      "session-a",
    );
    assert.deepEqual(clientOptions, serverOptions);
  });

  it("includes the answer and first three distractors without changing the queue", () => {
    const dueWords = Array.from({ length: 6 }, (_, index) =>
      Object.freeze(card(index)),
    );
    Object.freeze(dueWords);
    const options = buildMultipleChoiceOptions(dueWords, 4, "session-a");
    assert.deepEqual(
      [...options].sort((a, b) => a.meaningId.localeCompare(b.meaningId)),
      [0, 1, 2, 4].map((index) => ({
        meaningId: `meaning-${index}`,
        label: `noun — Definition ${index}`,
      })),
    );
    assert.deepEqual(
      dueWords.map((item) => item.meaningId),
      [0, 1, 2, 3, 4, 5].map((index) => `meaning-${index}`),
    );
  });

  it("varies the answer position across saved cards and meanings", () => {
    for (const identity of ["userWordId", "meaningId"] as const) {
      const positions = new Set<number>();
      for (let index = 0; index < 64; index++) {
        const current = { ...card(0), [identity]: `identity-${index}` };
        const options = buildMultipleChoiceOptions(
          [current, card(1), card(2), card(3)],
          0,
          "session-a",
        );
        positions.add(
          options.findIndex((option) => option.meaningId === current.meaningId),
        );
      }
      assert.equal(
        positions.size,
        4,
        `${identity} should vary the answer across all positions`,
      );
    }
  });

  it("keeps order stable when unrelated scheduling metadata changes", () => {
    const dueWords = Array.from({ length: 4 }, (_, index) => card(index));
    const options = buildMultipleChoiceOptions(dueWords, 0, "session-a");
    assert.deepEqual(
      buildMultipleChoiceOptions(
        dueWords.map((item) => ({
          ...item,
          reviewStep: 3,
          status: "learning",
        })),
        0,
        "session-a",
      ),
      options,
    );
  });

  it("handles missing cards and small queues for the self-check fallback", () => {
    assert.deepEqual(buildMultipleChoiceOptions([], 0, "session-a"), []);
    assert.deepEqual(buildMultipleChoiceOptions([card(0)], 1, "session-a"), []);
    for (let count = 1; count < 4; count++) {
      assert.equal(
        buildMultipleChoiceOptions(
          Array.from({ length: count }, (_, index) => card(index)),
          0,
          "session-a",
        ).length,
        count,
      );
    }
  });
});
