// DOC-09 §6 and the API validator both define the limit in Unicode code
// points. JavaScript's string length and a textarea's native maxLength use
// UTF-16 code units, so use Array.from to keep the UI boundary aligned.
export const MAX_SENTENCE_CHARACTERS = 300;

export function countSentenceCharacters(value: string): number {
  return Array.from(value).length;
}

// Keep the existing value when an edit would exceed the limit. Truncating the
// proposed value could otherwise discard the trailing part of a valid sentence
// when a learner inserts or pastes text in its middle.
export function acceptSentenceEdit(
  previousValue: string,
  nextValue: string,
  limit = MAX_SENTENCE_CHARACTERS,
): string {
  return countSentenceCharacters(nextValue) <= limit
    ? nextValue
    : previousValue;
}
