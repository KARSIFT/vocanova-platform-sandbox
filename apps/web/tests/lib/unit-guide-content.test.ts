import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { describe, it } from "node:test";
import {
  unitGuides,
  storySituationSlug,
} from "../../src/app/(app)/discover/[situation]/_components/unit-guide-content";

interface Seed {
  journey_situations: { id: string; slug: string; status: string }[];
  journey_words: { journey_situation_id: string; meaning_id: string }[];
  word_meanings: { id: string; word_id: string; status: string }[];
  canonical_words: { id: string; text: string; status: string }[];
}
const seed = JSON.parse(
  readFileSync(
    new URL("../../../api/cmd/seed/voc026-p1.json", import.meta.url),
    "utf8",
  ),
) as Seed;
describe("original unit guides", () => {
  it("covers all seventeen real situations with meaning-aligned phrase anchors", () => {
    const situations = seed.journey_situations.filter(
      (s) => s.status === "active",
    );
    assert.equal(situations.length, 17);
    assert.deepEqual(
      Object.keys(unitGuides).sort(),
      situations.map((s) => s.slug).sort(),
    );
    const allPhrases = new Set<string>();
    for (const situation of situations) {
      const guide = unitGuides[situation.slug];
      assert.ok(guide);
      assert.ok(guide.goal.length > 20);
      assert.ok(guide.tip.length > 20);
      assert.equal(guide.phrases.length, 3);
      const words = seed.journey_words
        .filter((w) => w.journey_situation_id === situation.id)
        .map((w) => {
          const meaning = seed.word_meanings.find(
            (m) => m.id === w.meaning_id && m.status === "active",
          );
          return seed.canonical_words.find(
            (c) => c.id === meaning?.word_id && c.status === "active",
          )?.text;
        });
      for (const phrase of guide.phrases) {
        assert.ok(
          words.includes(phrase.word),
          `${situation.slug}: ${phrase.word} is a canonical situation meaning`,
        );
        assert.ok(phrase.text.toLowerCase().includes(phrase.word));
        assert.ok(phrase.purpose.length > 15);
        assert.ok(!allPhrases.has(phrase.text));
        allPhrases.add(phrase.text);
      }
    }
    assert.equal(allPhrases.size, 51);
  });
  it("maps story categories only to existing unit guides", () => {
    for (const category of [
      "restaurant",
      "transport",
      "work",
      "shopping",
      "friends",
      "hotel",
    ])
      assert.ok(unitGuides[storySituationSlug(category)!]);
    assert.equal(storySituationSlug("unsupported"), undefined);
  });
});
