import assert from "node:assert/strict";
import { describe, it } from "node:test";

import type { KnowledgeMapItem, WordDetail } from "@vocanova/api-client";
import {
  eligibleReviewSeeds,
  MAX_REVIEW_SEED_CANDIDATES,
  reviewedMeaningAnswer,
  reviewSeedActions,
} from "../staging-e2e/review-seed";

describe("persistent staging review seed", () => {
  it("saves an initially unsaved word once", () => {
    assert.deepEqual(reviewSeedActions(false), ["save"]);
  });

  it("preserves an already-saved word and never claims restoration resets today's eligibility", () => {
    assert.deepEqual(reviewSeedActions(true), []);
  });

  it("considers only bounded fresh unsaved meanings without altering known or saved history", () => {
    const item = (
      rank: number,
      patch: Partial<KnowledgeMapItem> = {},
    ): KnowledgeMapItem => ({
      rank,
      meaningId: `meaning-${rank}`,
      wordId: `word-${rank}`,
      wordSlug: `word-${rank}`,
      wordText: `Word ${rank}`,
      partOfSpeech: "noun",
      shortDefinition: `Definition ${rank}`,
      selfReportedKnown: false,
      saved: false,
      due: false,
      revision: "revision",
      ...patch,
    });
    const candidates = [
      item(1, { saved: true }),
      item(2, { selfReportedKnown: true }),
      item(3, { due: true }),
      item(4, { shortDefinition: " " }),
      ...Array.from({ length: 8 }, (_, index) => item(index + 5)),
    ];
    const before = structuredClone(candidates);
    assert.deepEqual(
      eligibleReviewSeeds(candidates).map((candidate) => candidate.rank),
      [5, 6, 7, 8, 9],
    );
    assert.equal(
      eligibleReviewSeeds(candidates).length,
      MAX_REVIEW_SEED_CANDIDATES,
    );
    assert.equal(eligibleReviewSeeds(candidates, 2).length, 2);
    assert.deepEqual(candidates, before);
  });

  it("chooses the exact meaning's definition before grading, including same-spelling different parts of speech", () => {
    const word: WordDetail = {
      id: "record-word",
      slug: "record",
      text: "record",
      wordType: "word",
      meanings: [
        {
          id: "noun",
          partOfSpeech: "noun",
          shortDefinition: "Stored information.",
          saved: false,
          selfReportedKnown: false,
          examples: [],
          usageNotes: [],
        },
        {
          id: "verb",
          partOfSpeech: "verb",
          shortDefinition: "To store sound or information.",
          saved: false,
          selfReportedKnown: false,
          examples: [],
          usageNotes: [],
        },
      ],
    };
    const step = {
      id: "question",
      kind: "meaning_choice" as const,
      wordText: "record",
      wordSlug: "record",
      partOfSpeech: "verb",
      pronunciationMeaningId: "verb",
      choices: [
        { id: "wrong", text: "Stored information." },
        { id: "correct", text: "To store sound or information." },
      ],
    };
    assert.deepEqual(reviewedMeaningAnswer(step, word), {
      meaningId: "verb",
      definition: "To store sound or information.",
      choiceId: "correct",
    });
    assert.equal(
      reviewedMeaningAnswer(
        { ...step, pronunciationMeaningId: undefined },
        word,
      ),
      null,
    );
    assert.equal(
      reviewedMeaningAnswer({ ...step, pronunciationMeaningId: "noun" }, word),
      null,
    );
    assert.equal(
      reviewedMeaningAnswer({ ...step, wordSlug: "different" }, word),
      null,
    );
    assert.equal(
      reviewedMeaningAnswer(
        {
          ...step,
          choices: [{ id: "other", text: "A stale frozen definition." }],
        },
        word,
      ),
      null,
    );
    assert.equal(
      reviewedMeaningAnswer(
        {
          ...step,
          choices: [
            ...step.choices,
            { id: "duplicate", text: "To store sound or information." },
          ],
        },
        word,
      ),
      null,
    );
  });
});
