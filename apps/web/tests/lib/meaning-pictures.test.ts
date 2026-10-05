import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { describe, it } from "node:test";
import {
  MEANING_PICTURES,
  getMeaningPicture,
  getReadyMeaningPicture,
} from "../../src/lib/meaning-pictures";

const seed = JSON.parse(
  readFileSync(
    new URL("../../../api/cmd/seed/voc026-p1.json", import.meta.url),
    "utf8",
  ),
) as {
  canonical_words: { id: string; text: string }[];
  word_meanings: {
    id: string;
    word_id: string;
    short_definition: string;
    status: string;
  }[];
};

describe("meaning-specific picture content", () => {
  it("authors one distinct scene for every active canonical meaning", () => {
    const meanings = seed.word_meanings.filter(
      (meaning) => meaning.status === "active",
    );
    assert.deepEqual(
      Object.keys(MEANING_PICTURES).sort(),
      meanings.map((meaning) => meaning.id).sort(),
    );
    const sources = new Set<string>();
    const scenes = new Set<string>();
    for (const meaning of meanings) {
      const picture = getMeaningPicture(meaning.id)!;
      assert.equal(picture.meaningId, meaning.id);
      assert.equal(
        picture.word,
        seed.canonical_words.find((word) => word.id === meaning.word_id)?.text,
      );
      assert.equal(picture.definition, meaning.short_definition);
      assert.equal(picture.src, `/images/meanings/${meaning.id}.webp`);
      assert.ok(
        picture.scene.length > 70,
        `${picture.word}: usable scene brief`,
      );
      assert.ok(
        picture.alt.length > 15 && picture.alt.length < 160,
        `${picture.word}: concise visual description`,
      );
      assert.ok(!sources.has(picture.src), `${picture.word}: its own asset`);
      assert.ok(
        !scenes.has(picture.scene),
        `${picture.word}: its own meaning-specific scene`,
      );
      assert.equal(
        picture.status,
        "ready",
        `${picture.word}: canonical meaning has reviewed artwork`,
      );
      sources.add(picture.src);
      scenes.add(picture.scene);
      if (picture.status === "ready") {
        const asset = new URL(`../../public${picture.src}`, import.meta.url);
        assert.ok(existsSync(asset), `${picture.word}: ready asset exists`);
        const bytes = readFileSync(asset);
        assert.equal(bytes.subarray(0, 4).toString(), "RIFF");
        assert.equal(bytes.subarray(8, 12).toString(), "WEBP");
      }
    }
  });

  it("keeps repeated words separated by meaning, never by keyword guessing", () => {
    for (const word of ["reservation", "follow-up", "deadline"]) {
      const pictures = Object.values(MEANING_PICTURES).filter(
        (picture) => picture.word === word,
      );
      assert.equal(pictures.length, 2);
      assert.ok(pictures[0] && pictures[1]);
      assert.notEqual(pictures[0].src, pictures[1].src);
      assert.notEqual(pictures[0].scene, pictures[1].scene);
      assert.notEqual(pictures[0].alt, pictures[1].alt);
    }
    assert.equal(getMeaningPicture("reservation"), undefined);
    assert.equal(getMeaningPicture("unknown-meaning"), undefined);
    assert.equal(getMeaningPicture("toString"), undefined);
  });

  it("omits unavailable and failed media without inventing a replacement", () => {
    assert.equal(getReadyMeaningPicture("unknown-meaning"), undefined);
    for (const picture of Object.values(MEANING_PICTURES)) {
      assert.equal(
        getReadyMeaningPicture(picture.meaningId, picture.src),
        undefined,
      );
      assert.equal(
        getReadyMeaningPicture(picture.meaningId),
        picture.status === "ready" ? picture : undefined,
      );
    }
  });
});
