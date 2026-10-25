// Authored catalog fixture for a word with multiple real meanings.
// The normal mock covers only the smaller starter subset; this handler exposes
// actual canonical meanings without inventing related content.
import { createHash } from "node:crypto";
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
  if (cookies.e2e_teaching !== "true" || req.method !== "GET") return false;
  if (url.pathname === "/api/v1/meaning-images") {
    const ids = (url.searchParams.get("meaningIds") ?? "").split(",");
    if (ids.length > 100) {
      jsonResponse(res, 400, { error: "too_many_meanings" });
      return true;
    }
    const requested = new Set(ids);
    const items = ["ready", "stale"].includes(cookies.e2e_meaning_image)
      ? seed.word_meanings.filter((meaning) => requested.has(meaning.id) && meaning.status === "active").map((meaning) => {
          const word = seed.canonical_words.find((item) => item.id === meaning.word_id);
          const meaningRevision = cookies.e2e_meaning_image === "stale" ? "f".repeat(64) : createHash("sha256").update(JSON.stringify([meaning.id, word.text, meaning.part_of_speech, meaning.short_definition])).digest("hex");
          const sha256 = "a".repeat(64);
          return { meaningId: meaning.id, meaningRevision, sha256, imageUrl: `https://images.vocanova.site/meanings/${meaning.id}/${meaningRevision}/${sha256}.webp`, alt: "A learner reviewing a document at a bright desk.", width: 960, height: 640 };
        })
      : [];
    const respond = () => {
      if (!res.destroyed) jsonResponse(res, cookies.e2e_meaning_images_status === "503" ? 503 : 200, { items });
    };
    const delay = Math.max(0, Math.min(Number(cookies.e2e_meaning_images_delay_ms) || 0, 10000));
    if (delay) setTimeout(respond, delay); else respond();
    return true;
  }
  if (!url.pathname.startsWith("/api/v1/canonical-words/")) return false;
  const slug = decodeURIComponent(url.pathname.split("/").at(-1));
  const word = seed.canonical_words.find(
    (item) => item.normalized_text.replaceAll(" ", "-") === slug,
  );
  if (!word) return false;
  const meanings = seed.word_meanings
    .filter((item) => item.word_id === word.id && item.status === "active")
    .map((meaning) => ({
      id: meaning.id,
      meaningRevision: cookies.e2e_meaning_image === "stale" ? "f".repeat(64) : createHash("sha256").update(JSON.stringify([
        meaning.id, word.text, meaning.part_of_speech, meaning.short_definition,
      ])).digest("hex"),
      ...(cookies.e2e_meaning_image === "ready" ? {
        image: (() => {
          const meaningRevision = createHash("sha256").update(JSON.stringify([
            meaning.id, word.text, meaning.part_of_speech, meaning.short_definition,
          ])).digest("hex");
          const sha256 = "a".repeat(64);
          return {
            meaningId: meaning.id, meaningRevision, sha256,
            imageUrl: `https://images.vocanova.site/meanings/${meaning.id}/${meaningRevision}/${sha256}.webp`,
            alt: "A learner reviewing a document at a bright desk.", width: 960, height: 640,
          };
        })(),
      } : {}),
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
