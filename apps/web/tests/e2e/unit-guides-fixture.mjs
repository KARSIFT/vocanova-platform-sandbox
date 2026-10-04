// Opt-in real editorial catalog transport fixture; it grants no progress.
import { readFileSync } from "node:fs";
const seed = JSON.parse(
  readFileSync(
    new URL("../../../api/cmd/seed/voc026-p1.json", import.meta.url),
    "utf8",
  ),
);
export async function handleUnitGuideFeatures({
  req,
  res,
  url,
  cookies,
  jsonResponse,
}) {
  if (cookies.e2e_unit_guides !== "true" || req.method !== "GET") return false;
  const send = (code, body) => {
    jsonResponse(res, code, body);
    return true;
  };
  const canonicalWord = (word) => ({
    id: word.id,
    text: word.text,
    slug: word.normalized_text.replaceAll(" ", "-"),
    wordType: word.word_type,
    difficultyLevel: word.difficulty_level,
    meanings: seed.word_meanings
      .filter((m) => m.word_id === word.id && m.status === "active")
      .map((m) => ({
        id: m.id,
        partOfSpeech: m.part_of_speech,
        shortDefinition: m.short_definition,
        learnerDefinition: m.learner_definition,
        saved: false,
        selfReportedKnown: false,
        examples: seed.word_examples
          .filter((e) => e.meaning_id === m.id)
          .map((e) => ({ id: e.id, exampleText: e.example_text })),
        usageNotes: seed.usage_notes
          .filter((n) => n.meaning_id === m.id)
          .map((n) => ({
            id: n.id,
            noteType: n.note_type,
            noteText: n.note_text,
          })),
      })),
  });
  const match = url.pathname.match(/^\/api\/v1\/journey-situations\/([^/]+)$/u);
  if (match) {
    const slug = decodeURIComponent(match[1]);
    const s = seed.journey_situations.find(
      (s) => s.slug === slug && s.status === "active",
    );
    if (!s) return send(404, { detail: "Situation not found" });
    return send(200, {
      situation: {
        id: s.id,
        slug: s.slug,
        title: s.title,
        shortDescription: s.short_description,
        levelBand: s.level_band,
        category: s.category,
        displayOrder: s.display_order,
      },
      meanings: seed.journey_words
        .filter((w) => w.journey_situation_id === s.id)
        .map((link) => {
          const m = seed.word_meanings.find((m) => m.id === link.meaning_id);
          const w = seed.canonical_words.find((w) => w.id === m.word_id);
          return {
            meaningId: m.id,
            wordId: w.id,
            wordSlug: w.normalized_text.replaceAll(" ", "-"),
            wordText: w.text,
            partOfSpeech: m.part_of_speech,
            shortDefinition: m.short_definition,
            saved: false,
            selfReportedKnown: false,
          };
        }),
    });
  }
  const word = url.pathname.match(/^\/api\/v1\/canonical-words\/([^/]+)$/u);
  if (word) {
    const w = seed.canonical_words.find(
      (w) =>
        w.normalized_text.replaceAll(" ", "-") ===
          decodeURIComponent(word[1]) && w.status === "active",
    );
    return w
      ? send(200, { word: canonicalWord(w) })
      : send(404, { detail: "Word not found" });
  }
  return false;
}
