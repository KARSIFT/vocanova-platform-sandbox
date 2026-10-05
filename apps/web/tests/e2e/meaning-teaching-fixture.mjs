// Authored catalog fixture for a word with multiple real meanings.
// The normal mock covers only the smaller starter subset; this handler exposes
// actual canonical meanings without inventing related content.
import { readFileSync } from "node:fs";
const seed = JSON.parse(
  readFileSync(
    new URL("../../../api/cmd/seed/voc026-p1.json", import.meta.url),
    "utf8",
  ),
);
export function handleMeaningTeaching({
  req,
  res,
  url,
  cookies,
  jsonResponse,
}) {
  if (
    cookies.e2e_teaching !== "true" ||
    req.method !== "GET" ||
    !url.pathname.startsWith("/api/v1/canonical-words/")
  )
    return false;
  const slug = decodeURIComponent(url.pathname.split("/").at(-1));
  const word = seed.canonical_words.find(
    (item) => item.normalized_text.replaceAll(" ", "-") === slug,
  );
  if (!word) return false;
  const meanings = seed.word_meanings
    .filter((item) => item.word_id === word.id && item.status === "active")
    .map((meaning) => ({
      id: meaning.id,
      partOfSpeech: meaning.part_of_speech,
      shortDefinition: meaning.short_definition,
      learnerDefinition: meaning.learner_definition,
      saved: false,
      examples: seed.word_examples
        .filter(
          (item) => item.meaning_id === meaning.id && item.status === "active",
        )
        .sort((a, b) => a.example_order - b.example_order)
        .map((item) => ({ id: item.id, exampleText: item.example_text })),
      usageNotes: seed.usage_notes
        .filter(
          (item) => item.meaning_id === meaning.id && item.status === "active",
        )
        .sort((a, b) => a.note_order - b.note_order)
        .map((item) => ({
          id: item.id,
          noteType: item.note_type,
          noteText: item.note_text,
        })),
    }));
  jsonResponse(res, 200, {
    word: {
      id: word.id,
      text: word.text,
      slug,
      wordType: word.word_type,
      difficultyLevel: word.difficulty_level,
      meanings,
    },
  });
  return true;
}
