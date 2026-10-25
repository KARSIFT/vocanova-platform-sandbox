import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { describe, it } from "node:test";
import { fileURLToPath } from "node:url";
import { getExternalMeaningPicture, getReadyMeaningPicture } from "../../src/lib/meaning-pictures";
import { requestMeaningPicture } from "../../src/lib/meaning-picture-requests";
import { createRuntimeCatalogue } from "../../scripts/generate-meaning-pictures.mjs";
import runtimeCatalogue from "../../src/lib/meaning-pictures.runtime.json" with { type: "json" };

interface AuthoredMeaningPicture {
  meaningId: string;
  meaningRevision: string;
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
      const ready = getReadyMeaningPicture(picture.meaningId);
      assert.ok(ready);
      assert.equal(ready.alt, picture.alt);
      assert.equal(getReadyMeaningPicture(picture.meaningId, undefined, picture.meaningRevision)?.src, ready.src);
      assert.equal(getReadyMeaningPicture(picture.meaningId, undefined, "f".repeat(64)), undefined);
      const source = new URL(ready.src, "https://vocanova.invalid");
      assert.equal(source.pathname, picture.src);
      assert.match(source.searchParams.get("v")!, /^[a-f0-9]{16}$/);
      assert.equal(
        source.searchParams.get("v"),
        runtimeCatalogue[picture.meaningId as keyof typeof runtimeCatalogue]
          .version,
      );
      assert.equal(
        getReadyMeaningPicture(picture.meaningId, ready.src),
        undefined,
      );
    }
  });

  it("ships a synchronized catalogue with only ready IDs, alt text and byte revisions", () => {
    const expected = createRuntimeCatalogue(MEANING_PICTURES);
    assert.equal(Object.keys(expected).length, 92);
    assert.deepEqual(runtimeCatalogue, expected);
    assert.ok(
      Object.values(runtimeCatalogue).every(
        (picture) =>
          typeof picture.alt === "string" &&
          /^[a-f0-9]{16}$/.test(picture.version) &&
          Object.keys(picture).sort().join(",") === "alt,meaningRevision,version",
      ),
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

  it("gives changed artwork bytes a new cache identity while preserving the meaning and alt", () => {
    const picture = Object.values(MEANING_PICTURES)[0]!;
    const manifest = { [picture.meaningId]: picture };
    const first = createRuntimeCatalogue(manifest, () =>
      Buffer.from("first artwork payload"),
    );
    const repeated = createRuntimeCatalogue(manifest, () =>
      Buffer.from("first artwork payload"),
    );
    const revised = createRuntimeCatalogue(manifest, () =>
      Buffer.from("revised artwork payload"),
    );
    assert.deepEqual(
      first,
      repeated,
      "Unchanged bytes retain their cache identity",
    );
    const original = first[picture.meaningId];
    const replacement = revised[picture.meaningId];
    assert.ok(original && replacement);
    assert.notEqual(original.version, replacement.version);
    assert.equal(original.alt, replacement.alt);
    assert.deepEqual(Object.keys(first), Object.keys(revised));
    assert.throws(
      () =>
        createRuntimeCatalogue(manifest, () => {
          throw new Error("Missing ready asset");
        }),
      /Missing ready asset/,
    );
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
      { ...picture, meaningRevision: "invalid" },
    ]) {
      assert.throws(() =>
        createRuntimeCatalogue({ [picture.meaningId]: invalid }),
      );
    }
    assert.throws(() => createRuntimeCatalogue({ unknown: picture }));
  });
});


describe("external meaning picture delivery", () => {
  const meaningId = "00000000-0000-4000-8000-000000000001";
  const meaningRevision = "b".repeat(64);
  const sha256 = "c".repeat(64);
  const image = {
    meaningId, meaningRevision, sha256,
    imageUrl: `https://images.vocanova.site/meanings/${meaningId}/${meaningRevision}/${sha256}.webp`,
    alt: "A cyclist moves across a sunny bridge.", width: 900, height: 600,
  };
  it("accepts only exact meaning/revision/content-bound approved CDN URLs", () => {
    const result = getExternalMeaningPicture(meaningId, image, meaningRevision);
    assert.equal(result?.src, image.imageUrl);
    assert.equal(result?.external, true);
    assert.equal(result?.alt, image.alt);
    assert.equal(getExternalMeaningPicture(meaningId, image, "d".repeat(64)), undefined);
    assert.equal(getExternalMeaningPicture("00000000-0000-4000-8000-000000000002", image), undefined);
    assert.equal(getExternalMeaningPicture(meaningId, image, meaningRevision, [image.imageUrl]), undefined);
  });
  it("rejects arbitrary origins, altered object paths and malformed metadata", () => {
    for (const bad of [
      { ...image, imageUrl: image.imageUrl.replace("images.vocanova.site", "images.vocanova.site.evil.test") },
      { ...image, imageUrl: image.imageUrl.replace("https:", "http:") },
      { ...image, imageUrl: `${image.imageUrl}?token=untrusted` },
      { ...image, imageUrl: image.imageUrl.replace(sha256, "d".repeat(64)) },
      { ...image, meaningRevision: "old" },
      { ...image, sha256: "unknown" },
      { ...image, alt: " " },
      { ...image, width: 0 },
      { ...image, width: 961 },
      { ...image, height: 641 },
      { ...image, height: 2.5 },
    ]) assert.equal(getExternalMeaningPicture(meaningId, bad), undefined);
    assert.equal(getExternalMeaningPicture(meaningId), undefined);
    for (const malformedId of ["-".repeat(36), "a".repeat(36)]) {
      const malformed = { ...image, meaningId: malformedId, imageUrl: `https://images.vocanova.site/meanings/${malformedId}/${meaningRevision}/${sha256}.webp` };
      assert.equal(getExternalMeaningPicture(malformedId, malformed), undefined);
    }
  });
});


it("coalesces concurrent teaching lookups into bounded batches and shares repeat reads", async () => {
  const originalFetch = globalThis.fetch;
  const calls: string[][] = [];
  globalThis.fetch = async (input) => {
    const ids = new URL(String(input)).searchParams.get("meaningIds")!.split(",");
    calls.push(ids);
    return Response.json({ items: ids.map((meaningId) => ({ meaningId })) });
  };
  try {
    const ids = Array.from({ length: 205 }, (_, index) => `batch-test-${index}`);
    const promises = ids.map((id) => requestMeaningPicture(id));
    assert.equal(requestMeaningPicture(ids[0]!), promises[0]);
    const results = await Promise.all(promises);
    assert.equal(calls.length, 3);
    assert.ok(calls.every((ids) => ids.length <= 100));
    assert.deepEqual(calls.flat(), ids);
    assert.deepEqual(results.map((result) => result?.meaningId), ids);
    await requestMeaningPicture(ids[0]!);
    assert.equal(calls.length, 3);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

it("aborts a stalled teaching-image lookup and settles its waiting learners", {
  timeout: 6000,
}, async () => {
  const originalFetch = globalThis.fetch;
  let aborted = false;
  const keepAlive = setTimeout(() => {}, 10_000);
  globalThis.fetch = async (_input, init) => {
    const signal = init?.signal;
    if (!signal) throw new Error("Teaching image request has no deadline");
    return new Promise<Response>((_resolve, reject) => {
      signal.addEventListener(
        "abort",
        () => {
          aborted = true;
          reject(signal.reason);
        },
        { once: true },
      );
    });
  };
  try {
    const first = requestMeaningPicture("stalled-teaching-lookup");
    assert.equal(requestMeaningPicture("stalled-teaching-lookup"), first);
    assert.equal(await first, undefined);
    assert.equal(
      aborted,
      true,
      "the network request must be aborted, not left running",
    );
    assert.equal(await requestMeaningPicture("stalled-teaching-lookup"), undefined);
  } finally {
    clearTimeout(keepAlive);
    globalThis.fetch = originalFetch;
  }
});
