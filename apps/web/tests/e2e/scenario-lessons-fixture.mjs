// The five guided lesson / story pairs use the real starter meaning identities.
import stories from "./stories-catalog-fixture.json" with { type: "json" };
import seed from "../../../api/cmd/seed/voc026-p1.json" with { type: "json" };

const pairs = [
  ["work-meeting", "Leave a meeting with a plan", "work-meeting", "before-the-team-meeting"],
  ["airport", "Find your flight", "airport", "a-new-gate"],
  ["restaurant-other-plans", "Plan a meal out or at home", "restaurant", "dinner-to-go"],
  ["shopping-returns", "Return a purchase", "shopping", "the-lamp-that-went-dark"],
  ["delivery-parcel", "Receive a parcel", "deliveries", "a-parcel-next-door"],
];

export const scenarioLessons = new Map(pairs.map(([key, title, situationSlug, storyKey]) => {
  const story = stories.find(({Story}) => Story.key === storyKey).Story;
  const situation = seed.journey_situations.find((item) => item.slug === situationSlug);
  const words = story.vocabulary.map(({meaningId, wordText, wordSlug}) => {
    const meaning = seed.word_meanings.find((item) => item.id === meaningId);
    const example = seed.word_examples.find((item) => item.meaning_id === meaningId && item.example_order === 1);
    return { meaningId, wordText, wordSlug, definition: meaning.short_definition, partOfSpeech: meaning.part_of_speech, example: example.example_text };
  });
  return [key, {
    storyKey, words,
    summary: { key, title, version: "1", situationSlug, situationTitle: situation.title, description: story.description, wordCount: words.length, stepCount: words.length * 3 },
  }];
}));

export function showScenarioLessons(cookies, state) {
  if (cookies.e2e_lessons === "scenarios") return true;
  const storyKeys = new Set(pairs.map(([, , , storyKey]) => storyKey));
  return Array.from(state.storySessions?.values() ?? []).some((session) => storyKeys.has(session.snapshot.Story.key));
}
