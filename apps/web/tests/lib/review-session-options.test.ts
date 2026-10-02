import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { describe, it } from "node:test";

import type { DueWord } from "@vocanova/api-client";

import { buildMultipleChoiceOptions } from "../../src/app/(app)/reviews/_components/review-session-options";

interface Seed {
  canonical_words: {
    id: string;
    text: string;
    normalized_text: string;
    status: string;
  }[];
  word_meanings: {
    id: string;
    word_id: string;
    part_of_speech: string;
    short_definition: string;
    status: string;
  }[];
}

const seed = JSON.parse(
  readFileSync(
    new URL("../../../api/cmd/seed/voc026-p1.json", import.meta.url),
    "utf8",
  ),
) as Seed;

function seededCard(meaningId: string): DueWord {
  const meaning = seed.word_meanings.find((item) => item.id === meaningId);
  assert.ok(meaning, `Expected canonical meaning ${meaningId}`);
  assert.equal(meaning.status, "active");
  const word = seed.canonical_words.find((item) => item.id === meaning.word_id);
  assert.ok(word, `Expected canonical word for meaning ${meaningId}`);
  assert.equal(word.status, "active");
  return Object.freeze({
    userWordId: `saved-${meaning.id}`,
    meaningId: meaning.id,
    wordId: word.id,
    wordSlug: word.normalized_text.replaceAll(" ", "-"),
    wordText: word.text,
    partOfSpeech: meaning.part_of_speech,
    shortDefinition: meaning.short_definition,
    status: "due",
    reviewStep: 0,
  });
}

const multipleMeaningPairs = [
  {
    word: "reservation",
    meaningIds: [
      "52c8ed28-544d-5a80-a911-6ddd03953606",
      "edb20c6e-1023-5f10-af4b-0dc059a781b2",
    ],
  },
  {
    word: "deadline",
    meaningIds: [
      "b571a540-89bb-533b-be81-9cedf443a64a",
      "f582a137-c77d-58df-942e-884fe336cd64",
    ],
  },
  {
    word: "follow-up",
    meaningIds: [
      "4c3724d8-5abf-5159-8171-76ccdcdae7c0",
      "446b7b9e-8190-5987-9401-0378e95ff493",
    ],
  },
] as const;

const safeAlternatives = [
  "329c4ec9-6562-568e-9e71-9134da2f0d8d", // menu
  "5d21c02b-ef51-5545-bd8a-01f5814d3a67", // bill
  "2534b539-a80f-5001-b9a4-27a5e0679539", // appetizer
  "bd5eacb6-be32-5e7b-a8d7-d56db1908347", // tip
].map(seededCard);

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
  for (const pair of multipleMeaningPairs) {
    it(`does not mark another seeded sense of ${pair.word} as a wrong answer`, () => {
      const meanings = pair.meaningIds.map(seededCard);
      assert.equal(meanings[0]!.wordText, pair.word);
      assert.equal(meanings[1]!.wordId, meanings[0]!.wordId);
      assert.notEqual(meanings[0]!.meaningId, meanings[1]!.meaningId);
      const dueWords = Object.freeze([...meanings, ...safeAlternatives]);

      // Both senses can be the current card; safe alternatives later in the
      // queue must fill the four options instead of including the other sense.
      for (const currentIndex of [0, 1]) {
        const current = dueWords[currentIndex]!;
        const options = buildMultipleChoiceOptions(
          dueWords,
          currentIndex,
          "session-multiple-meanings",
        );
        assert.deepEqual(
          new Set(options.map((option) => option.meaningId)),
          new Set([
            current.meaningId,
            ...safeAlternatives.slice(0, 3).map((item) => item.meaningId),
          ]),
          `Only ${current.meaningId} can represent ${pair.word} in the options`,
        );
        assert.equal(options.length, 4);
        assert.deepEqual(
          buildMultipleChoiceOptions(
            JSON.parse(JSON.stringify(dueWords)) as DueWord[],
            currentIndex,
            "session-multiple-meanings",
          ),
          options,
          "Server and client order must agree for the filtered queue",
        );
      }
    });

    it(`leaves ${pair.word} eligible for self-check when safe alternatives are insufficient`, () => {
      const meanings = pair.meaningIds.map(seededCard);
      const dueWords = Object.freeze([
        ...meanings,
        ...safeAlternatives.slice(0, 2),
      ]);
      for (const currentIndex of [0, 1]) {
        const options = buildMultipleChoiceOptions(
          dueWords,
          currentIndex,
          "session-multiple-meanings",
        );
        // The review UI falls back to self-check below four options. Two
        // valid senses must not make a four-card queue look unambiguous.
        assert.equal(options.length, 3);
        assert.deepEqual(
          new Set(options.map((option) => option.meaningId)),
          new Set([
            dueWords[currentIndex]!.meaningId,
            ...safeAlternatives.slice(0, 2).map((item) => item.meaningId),
          ]),
        );
      }
    });
  }

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
