import type {
  DailyLesson,
  KnowledgeMapItem,
  WordDetail,
} from "@vocanova/api-client";

export type ReviewSeedAction = "save";
export const MAX_REVIEW_SEED_CANDIDATES = 5;
export const REVIEW_SEED_BANDS = [99, 98, 97] as const;

/**
 * Returns the real saved-word transitions the persistent staging account must
 * perform before opening Review. Saved history and today's review audit are
 * preserved. Saving/restoring a word never proves that it is eligible today.
 */
export function reviewSeedActions(initiallySaved: boolean): ReviewSeedAction[] {
  return initiallySaved ? [] : ["save"];
}

/** Fresh candidates only. The caller must confirm actual due state after saving. */
export function eligibleReviewSeeds(
  items: readonly KnowledgeMapItem[],
  limit = MAX_REVIEW_SEED_CANDIDATES,
): KnowledgeMapItem[] {
  return items
    .filter(
      (item) =>
        item.saved === false &&
        item.selfReportedKnown === false &&
        item.due === false &&
        item.meaningId &&
        item.wordId &&
        item.wordSlug &&
        item.wordText &&
        item.partOfSpeech &&
        item.shortDefinition.trim(),
    )
    .slice(0, Math.max(0, Math.min(MAX_REVIEW_SEED_CANDIDATES, limit)));
}

type ReviewStep = NonNullable<DailyLesson["currentStep"]>;
type ReviewedAnswer = {
  meaningId: string;
  definition: string;
  choiceId: string;
  choiceText: string;
};

function distinctChoices(step: ReviewStep): boolean {
  return (
    step.choices.length > 0 &&
    step.choices.every((choice) => choice.id && choice.text.trim()) &&
    new Set(step.choices.map((choice) => choice.id)).size === step.choices.length &&
    new Set(step.choices.map((choice) => choice.text.trim().toLowerCase())).size ===
      step.choices.length
  );
}

/** Resolve only unique public canonical evidence, never opaque IDs or an answer key. */
export function reviewedMeaningAnswer(
  step: ReviewStep,
  word: WordDetail,
): ReviewedAnswer | null {
  if (
    !distinctChoices(step) ||
    step.kind === "gap_choice" ||
    step.wordText !== word.text ||
    step.wordSlug !== word.slug
  )
    return null;

  if (step.kind === "picture_choice") {
    if (
      step.choices.length !== 4 ||
      step.choices.some((choice) => !choice.image) ||
      new Set(step.choices.map((choice) => choice.image!.meaningId)).size !== 4
    )
      return null;
    const matches = word.meanings.flatMap((meaning) => {
      if (
        meaning.partOfSpeech !== step.partOfSpeech ||
        !meaning.shortDefinition.trim() ||
        !meaning.meaningRevision ||
        !meaning.image ||
        meaning.image.meaningId !== meaning.id ||
        meaning.image.meaningRevision !== meaning.meaningRevision
      )
        return [];
      return step.choices
        .filter(
          (choice) =>
            choice.text === meaning.shortDefinition &&
            choice.image!.meaningId === meaning.id &&
            choice.image!.meaningRevision === meaning.meaningRevision &&
            choice.image!.sha256 === meaning.image!.sha256,
        )
        .map((choice) => ({
          meaningId: meaning.id,
          definition: meaning.shortDefinition,
          choiceId: choice.id,
          choiceText: choice.text,
        }));
    });
    return matches.length === 1 ? matches[0]! : null;
  }

  // Legacy definition questions still require the exact published target identity.
  if (!step.pronunciationMeaningId) return null;
  const meanings = word.meanings.filter(
    (candidate) =>
      candidate.id === step.pronunciationMeaningId &&
      candidate.partOfSpeech === step.partOfSpeech,
  );
  const meaning = meanings.length === 1 ? meanings[0] : undefined;
  if (!meaning || !meaning.shortDefinition.trim()) return null;
  const choices = step.choices.filter(
    (choice) => choice.text === meaning.shortDefinition,
  );
  if (choices.length !== 1) return null;
  return {
    meaningId: meaning.id,
    definition: meaning.shortDefinition,
    choiceId: choices[0]!.id,
    choiceText: choices[0]!.text,
  };
}

/** At most four public candidate reads; edited/ambiguous authored examples are skipped. */
export async function reviewedGapAnswer(
  step: ReviewStep,
  readWord: (slug: string) => Promise<WordDetail | null>,
): Promise<ReviewedAnswer | null> {
  if (
    step.kind !== "gap_choice" ||
    !distinctChoices(step) ||
    step.choices.length !== 4 ||
    !step.context ||
    step.context.split("_____").length !== 2 ||
    step.choices.some((choice) => !/^[a-z]+(?:[ '-][a-z]+)*$/i.test(choice.text))
  )
    return null;
  // Mirrors the public canonical slug convention, using visible candidate text only.
  const slugs = step.choices.map((choice) =>
    choice.text.toLowerCase().replaceAll(" ", "-"),
  );
  if (new Set(slugs).size !== 4) return null;
  const words = await Promise.all(slugs.map((slug) => readWord(slug)));
  // Missing content cannot establish that a second defensible answer is absent.
  if (words.some((word) => !word)) return null;
  const matches: ReviewedAnswer[] = [];
  for (const [index, choice] of step.choices.entries()) {
    const word = words[index]!;
    if (word.text !== choice.text || word.slug !== slugs[index]) return null;
    for (const meaning of word.meanings) {
      if (
        meaning.partOfSpeech === step.partOfSpeech &&
        meaning.shortDefinition.trim() &&
        meaning.examples.some(
          (example) =>
            example.exampleText.split(word.text).length === 2 &&
            example.exampleText.replace(word.text, "_____") === step.context,
        )
      )
        matches.push({
          meaningId: meaning.id,
          definition: meaning.shortDefinition,
          choiceId: choice.id,
          choiceText: choice.text,
        });
    }
  }
  return matches.length === 1 ? matches[0]! : null;
}
