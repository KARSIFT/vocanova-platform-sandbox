import { readFileSync } from "node:fs";

// Exercise the editorial content that the real seed command installs. The mock
// still owns transport and learner state; this does not simulate PostgreSQL.
const seed = JSON.parse(
  readFileSync(
    new URL("../../../api/cmd/seed/voc026-p1.json", import.meta.url),
    "utf8",
  ),
);
const situation = seed.journey_situations.find(
  (item) => item.slug === "daily-conversation",
);
const links = seed.journey_words
  .filter((item) => item.journey_situation_id === situation.id)
  .sort(
    (a, b) =>
      Number(b.is_core) - Number(a.is_core) ||
      a.display_order - b.display_order ||
      b.relevance_score - a.relevance_score,
  );

export const dailyConversationWords = {};
const meanings = links.map((link) => {
  const meaning = seed.word_meanings.find(
    (item) => item.id === link.meaning_id,
  );
  const word = seed.canonical_words.find((item) => item.id === meaning.word_id);
  const slug = word.normalized_text.replaceAll(" ", "-");
  dailyConversationWords[slug] = {
    id: word.id,
    text: word.text,
    slug,
    wordType: word.word_type,
    difficultyLevel: word.difficulty_level,
    meanings: [
      {
        id: meaning.id,
        partOfSpeech: meaning.part_of_speech,
        shortDefinition: meaning.short_definition,
        learnerDefinition: meaning.learner_definition,
        saved: false,
        examples: seed.word_examples
          .filter((item) => item.meaning_id === meaning.id)
          .sort((a, b) => a.example_order - b.example_order)
          .map((item) => ({ id: item.id, exampleText: item.example_text })),
        usageNotes: seed.usage_notes
          .filter((item) => item.meaning_id === meaning.id)
          .sort((a, b) => a.note_order - b.note_order)
          .map((item) => ({
            id: item.id,
            noteType: item.note_type,
            noteText: item.note_text,
          })),
      },
    ],
  };
  return {
    meaningId: meaning.id,
    wordId: word.id,
    wordSlug: slug,
    wordText: word.text,
    partOfSpeech: meaning.part_of_speech,
    shortDefinition: meaning.short_definition,
    saved: false,
  };
});

export const dailyConversationFixture = {
  situation: {
    id: situation.id,
    slug: situation.slug,
    title: situation.title,
    shortDescription: situation.short_description,
    levelBand: situation.level_band,
    category: situation.category,
    displayOrder: situation.display_order,
  },
  meanings,
};
