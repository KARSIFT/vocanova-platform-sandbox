/**
 * A canonical word can belong to more than one Journey situation, but a
 * detail URL is valid only when its word appears in the situation segment.
 * Keep this check independent of React so the route boundary is testable.
 */
export function isWordInSituation(
  meanings: ReadonlyArray<{ wordSlug: string }>,
  wordSlug: string,
): boolean {
  return meanings.some((meaning) => meaning.wordSlug === wordSlug);
}
