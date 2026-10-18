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

/** Resolve a visible question from its exact meaning, never from a guessed choice. */
export function reviewedMeaningAnswer(
  step: NonNullable<DailyLesson["currentStep"]>,
  word: WordDetail,
) {
  if (
    !step.pronunciationMeaningId ||
    step.wordText !== word.text ||
    step.wordSlug !== word.slug
  )
    return null;
  const meaning = word.meanings.find(
    (candidate) =>
      candidate.id === step.pronunciationMeaningId &&
      candidate.partOfSpeech === step.partOfSpeech,
  );
  if (!meaning || !meaning.shortDefinition.trim()) return null;
  const choices = step.choices.filter(
    (choice) => choice.text === meaning.shortDefinition,
  );
  if (choices.length !== 1) return null;
  return {
    meaningId: meaning.id,
    definition: meaning.shortDefinition,
    choiceId: choices[0]!.id,
  };
}
