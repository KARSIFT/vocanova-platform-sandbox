// DOC-09 §6 and the API validator both define the limit in Unicode code
// points. JavaScript's string length and a textarea's native maxLength use
// UTF-16 code units, so use Array.from to keep the UI boundary aligned.
export const MAX_SENTENCE_CHARACTERS = 300;

export function countSentenceCharacters(value: string): number {
  return Array.from(value).length;
}

export function limitSentenceCharacters(
  value: string,
  limit = MAX_SENTENCE_CHARACTERS,
): string {
  return Array.from(value).slice(0, limit).join("");
}
