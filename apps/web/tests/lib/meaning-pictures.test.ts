import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { describe, it } from "node:test";
import { fileURLToPath } from "node:url";
import { getReadyMeaningPicture } from "../../src/lib/meaning-pictures";
import { createRuntimeCatalogue } from "../../scripts/generate-meaning-pictures.mjs";
import runtimeCatalogue from "../../src/lib/meaning-pictures.runtime.json" with { type: "json" };

interface AuthoredMeaningPicture {
  meaningId: string;
  word: string;
  definition: string;
  scene: string;
  alt: string;
  src: string;
  status: "ready" | "pending";
}

// Authoring content is read only in tests and the generator, never by client UI.
const MEANING_PICTURES = JSON.parse(
  readFileSync(
    new URL("../../src/lib/meaning-pictures.json", import.meta.url),
    "utf8",
  ),
) as Record<string, AuthoredMeaningPicture>;

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
    assert.equal(meanings.length, 92);
    assert.deepEqual(
      Object.keys(MEANING_PICTURES).sort(),
      meanings.map((meaning) => meaning.id).sort(),
    );
    const sources = new Set<string>();
    const scenes = new Set<string>();
    for (const meaning of meanings) {
      const picture = MEANING_PICTURES[meaning.id]!;
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
    for (const id of [
      "reservation",
      "unknown-meaning",
      "toString",
      "constructor",
      "__proto__",
    ]) {
      assert.equal(getReadyMeaningPicture(id), undefined);
    }
  });

  it("omits unavailable and failed media without inventing a replacement", () => {
    assert.equal(getReadyMeaningPicture("unknown-meaning"), undefined);
    for (const picture of Object.values(MEANING_PICTURES)) {
      assert.equal(
        getReadyMeaningPicture(picture.meaningId, picture.src),
        undefined,
      );
      assert.deepEqual(
        getReadyMeaningPicture(picture.meaningId),
        picture.status === "ready"
          ? { alt: picture.alt, src: picture.src }
          : undefined,
      );
    }
  });

  it("ships a synchronized catalogue containing only ready IDs and alt text", () => {
    const expected = createRuntimeCatalogue(MEANING_PICTURES);
    assert.equal(Object.keys(expected).length, 92);
    assert.deepEqual(runtimeCatalogue, expected);
    assert.ok(
      Object.values(runtimeCatalogue).every((alt) => typeof alt === "string"),
    );
    const result = spawnSync(
      process.execPath,
      [
        fileURLToPath(
          new URL(
            "../../scripts/generate-meaning-pictures.mjs",
            import.meta.url,
          ),
        ),
        "--check",
      ],
      { encoding: "utf8" },
    );
    assert.equal(result.status, 0, result.stderr);
  });

  it("excludes pending artwork and rejects mismatched identities or empty descriptions", () => {
    const picture = Object.values(MEANING_PICTURES)[0]!;
    assert.deepEqual(
      createRuntimeCatalogue({
        [picture.meaningId]: { ...picture, status: "pending" },
      }),
      {},
    );
    for (const invalid of [
      { ...picture, meaningId: "different" },
      { ...picture, src: "/images/another-meaning.webp" },
      { ...picture, status: "unreviewed" },
      { ...picture, alt: " " },
    ]) {
      assert.throws(() =>
        createRuntimeCatalogue({ [picture.meaningId]: invalid }),
      );
    }
    assert.throws(() => createRuntimeCatalogue({ unknown: picture }));
  });
});
