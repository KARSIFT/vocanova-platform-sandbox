import assert from "node:assert/strict";
import { describe, it } from "node:test";

import type { KnowledgeMapItem, WordDetail } from "@vocanova/api-client";
import {
  eligibleReviewSeeds,
  MAX_REVIEW_SEED_CANDIDATES,
  reviewedMeaningAnswer,
  reviewedGapAnswer,
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
      choiceText: "To store sound or information.",
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
      reviewedMeaningAnswer({ ...step, wordText: "different" }, word),
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

type ReviewStep = NonNullable<import("@vocanova/api-client").DailyLesson["currentStep"]>;

function catalogWord(text: string): WordDetail {
  const id = `meaning-${text}`;
  const image = {
    meaningId: id,
    meaningRevision: `revision-${text}`,
    imageUrl: `https://images.example/${text}.webp`,
    alt: text,
    width: 960,
    height: 640,
    sha256: "a".repeat(64),
  };
  return {
    id: `word-${text}`,
    text,
    slug: text.replaceAll(" ", "-"),
    wordType: "word",
    meanings: [{
      id,
      meaningRevision: image.meaningRevision,
      image,
      partOfSpeech: "noun",
      shortDefinition: `Definition of ${text}`,
      saved: false,
      selfReportedKnown: false,
      examples: [{ id: `example-${text}`, exampleText: `Our flight leaves from ${text} twelve.` }],
      usageNotes: [],
    }],
  };
}

function mixedFixture(kind: "picture_choice" | "gap_choice") {
  const words = ["gate", "syllabus", "luggage", "appetizer"].map(catalogWord);
  for (const word of words.slice(1)) word.meanings[0]!.examples = [];
  const step: ReviewStep = {
    id: "opaque-question",
    kind,
    wordText: kind === "gap_choice" ? "" : "gate",
    wordSlug: kind === "gap_choice" ? "" : "gate",
    context: kind === "gap_choice" ? "Our flight leaves from _____ twelve." : undefined,
    partOfSpeech: "noun",
    choices: words.map((word, index) => ({
      id: `opaque-option-${index}`,
      text: kind === "gap_choice" ? word.text : word.meanings[0]!.shortDefinition,
      image: kind === "picture_choice" ? word.meanings[0]!.image : undefined,
    })),
  };
  return { words, step };
}

describe("mixed staging review answers from public canonical content", () => {
  it("resolves pictures without a target pronunciation identity", () => {
    const { words, step } = mixedFixture("picture_choice");
    assert.deepEqual(reviewedMeaningAnswer(step, words[0]!), {
      meaningId: "meaning-gate",
      definition: "Definition of gate",
      choiceId: "opaque-option-0",
      choiceText: "Definition of gate",
    });
  });

  it("rejects stale picture definition, revision, image identity and ambiguous senses", () => {
    const { words, step } = mixedFixture("picture_choice");
    for (const patch of [
      { text: "Edited definition" },
      { image: { ...step.choices[0]!.image!, meaningRevision: "stale" } },
      { image: { ...step.choices[0]!.image!, meaningId: "different" } },
      { image: { ...step.choices[0]!.image!, sha256: "b".repeat(64) } },
    ]) {
      const changed = structuredClone(step);
      Object.assign(changed.choices[0]!, patch);
      assert.equal(reviewedMeaningAnswer(changed, words[0]!), null);
    }
    const ambiguous = structuredClone(words[0]!);
    ambiguous.meanings.push(structuredClone(ambiguous.meanings[0]!));
    assert.equal(reviewedMeaningAnswer(step, ambiguous), null);
    const staleCanonical = structuredClone(words[0]!);
    staleCanonical.meanings[0]!.meaningRevision = "edited";
    assert.equal(reviewedMeaningAnswer(step, staleCanonical), null);
  });

  it("rejects duplicate picture option IDs and duplicate image meaning identities", () => {
    const { words, step } = mixedFixture("picture_choice");
    const duplicateId = structuredClone(step);
    duplicateId.choices[1]!.id = duplicateId.choices[0]!.id;
    assert.equal(reviewedMeaningAnswer(duplicateId, words[0]!), null);
    const duplicateImage = structuredClone(step);
    duplicateImage.choices[1]!.image = duplicateImage.choices[0]!.image;
    assert.equal(reviewedMeaningAnswer(duplicateImage, words[0]!), null);
  });

  it("resolves gaps through exactly four visible candidate-word reads, separating choice from definition", async () => {
    const { words, step } = mixedFixture("gap_choice");
    const reads: string[] = [];
    const answer = await reviewedGapAnswer(step, async (slug) => {
      reads.push(slug);
      return words.find((word) => word.slug === slug) ?? null;
    });
    assert.deepEqual(reads, ["gate", "syllabus", "luggage", "appetizer"]);
    assert.deepEqual(answer, {
      meaningId: "meaning-gate",
      definition: "Definition of gate",
      choiceId: "opaque-option-0",
      choiceText: "gate",
    });
  });

  it("derives multiword canonical slugs from visible option text", async () => {
    const { words, step } = mixedFixture("gap_choice");
    words[0] = catalogWord("key card");
    words[0]!.meanings[0]!.examples[0]!.exampleText = "My key card stopped working.";
    step.choices[0]!.text = "key card";
    step.context = "My _____ stopped working.";
    const reads: string[] = [];
    const answer = await reviewedGapAnswer(step, async (slug) => {
      reads.push(slug);
      return words.find((word) => word.slug === slug) ?? null;
    });
    assert.equal(reads[0], "key-card");
    assert.equal(answer?.choiceText, "key card");
  });

  it("rejects edited examples, repeated targets, wrong POS and unrelated canonical responses", async () => {
    for (const change of ["edited", "repeated", "pos", "response"] as const) {
      const { words, step } = mixedFixture("gap_choice");
      const meaning = words[0]!.meanings[0]!;
      if (change === "edited") meaning.examples[0]!.exampleText = "A gate stands here.";
      if (change === "repeated") meaning.examples[0]!.exampleText = "Our flight leaves from gate gate twelve.";
      if (change === "pos") meaning.partOfSpeech = "verb";
      if (change === "response") words[0]!.text = "other";
      assert.equal(await reviewedGapAnswer(step, async (slug) =>
        words.find((word) => word.slug === slug) ?? null), null, change);
    }
  });

  it("rejects multiple defensible candidate examples and multiple matching senses", async () => {
    const { words, step } = mixedFixture("gap_choice");
    words[1]!.meanings[0]!.examples = [{ id: "ambiguous", exampleText: "Our flight leaves from syllabus twelve." }];
    assert.equal(await reviewedGapAnswer(step, async (slug) =>
      words.find((word) => word.slug === slug) ?? null), null);
    words[1]!.meanings[0]!.examples = [];
    words[0]!.meanings.push({ ...structuredClone(words[0]!.meanings[0]!), id: "other-gate-sense" });
    assert.equal(await reviewedGapAnswer(step, async (slug) =>
      words.find((word) => word.slug === slug) ?? null), null);
  });

  it("refuses to infer uniqueness when any public candidate is missing", async () => {
    const { words, step } = mixedFixture("gap_choice");
    let reads = 0;
    assert.equal(await reviewedGapAnswer(step, async (slug) => {
      reads++;
      return slug === "luggage" ? null : words.find((word) => word.slug === slug) ?? null;
    }), null);
    assert.equal(reads, 4);
  });

  it("rejects malformed, duplicate and oversized gaps before any candidate reads", async () => {
    const { step } = mixedFixture("gap_choice");
    const malformed = [
      { ...step, choices: [...step.choices, { id: "fifth", text: "extra" }] },
      { ...step, choices: step.choices.slice(0, 3) },
      { ...step, context: "No blank here" },
      { ...step, context: "_____ and _____" },
    ];
    const duplicateId = structuredClone(step);
    duplicateId.choices[1]!.id = duplicateId.choices[0]!.id;
    const duplicateText = structuredClone(step);
    duplicateText.choices[1]!.text = "Gate";
    malformed.push(duplicateId, duplicateText);
    let reads = 0;
    for (const candidate of malformed) {
      assert.equal(await reviewedGapAnswer(candidate, async () => { reads++; return null; }), null);
    }
    assert.equal(reads, 0);
  });

  it("never resolves a gap through the legacy pronunciation identity", () => {
    const { words, step } = mixedFixture("gap_choice");
    assert.equal(reviewedMeaningAnswer({ ...step, wordText: "gate", wordSlug: "gate", pronunciationMeaningId: "meaning-gate" }, words[0]!), null);
  });
});
