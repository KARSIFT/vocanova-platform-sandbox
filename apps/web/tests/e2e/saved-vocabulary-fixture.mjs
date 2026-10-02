// UI fixtures only; full collection filtering and cursor authority are checked
// independently against the production Go/PostgreSQL repository.
import { createHash } from "node:crypto";

export function savedVocabularyFixture({ url, cookies, state, canonicalWords }) {
  if (cookies.e2e_saved_words === "unavailable")
    return { status: 503, body: { detail: "Fixture saved words unavailable" } };
  const q = (url.searchParams.get("q") ?? "").trim().toLowerCase();
  const stage = url.searchParams.get("stage") ?? "";
  const due = url.searchParams.get("due") === "true";
  const after = url.searchParams.get("after") ?? "";
  if (Array.from(q).length > 100 || q.includes("\0") ||
      !["", "new", "learning", "reviewing", "mastered", "ignored", "archived"].includes(stage))
    return { status: 400, body: { detail: "Invalid search" } };
  let items = Object.values(canonicalWords).flatMap((word) =>
    word.meanings.filter((meaning) => state.savedMeaningIds.has(meaning.id)).map((meaning) => ({
      userWordId: `uw-${meaning.id}`, meaningId: meaning.id, wordId: word.id,
      wordSlug: word.slug, wordText: word.text, partOfSpeech: meaning.partOfSpeech,
      shortDefinition: meaning.shortDefinition, status: "new",
      reviewState: state.reviewedMeaningIds.has(meaning.id) ? "learning" : "new",
      due: !state.reviewedMeaningIds.has(meaning.id), source: "journey", saved: true,
      addedAt: "2026-10-02T10:00:00Z",
    }))
  );
  if (cookies.e2e_saved_words === "filters") {
    items = Array.from({ length: 26 }, (_, index) => {
      const number = String(index + 1).padStart(2, "0");
      const reviewState = index < 24 ? "new" : index === 24 ? "reviewing" : "archived";
      return {
        userWordId: `fixture-saved-${number}`, meaningId: `fixture-meaning-${number}`,
        wordId: `fixture-word-${number}`, wordSlug: `fixture-saved-${number}`,
        wordText: index < 24 ? `Fixture saved word ${number}` : `Fixture ${reviewState} word`,
        partOfSpeech: "noun", shortDefinition: `Saved collection example ${number}`,
        status: reviewState, reviewState, due: index < 24, source: "journey", saved: true,
        addedAt: "2026-10-02T10:00:00Z",
      };
    });
  }
  const filtered = items.filter(item =>
    (!q || `${item.wordText} ${item.shortDefinition}`.toLowerCase().includes(q)) &&
    (!stage || item.reviewState === stage) && (!due || item.due));
  const fingerprint = createHash("sha256").update(JSON.stringify([
    cookies.vocanova_session, q, stage, due,
  ])).digest("hex");
  let offset = 0;
  if (after) {
    try {
      const cursor = JSON.parse(Buffer.from(after, "base64url").toString("utf8"));
      if (cursor.v !== 1 || cursor.f !== fingerprint || !Number.isSafeInteger(cursor.o) || cursor.o < 0)
        throw new Error("Invalid fixture cursor");
      offset = cursor.o;
    } catch {
      return { status: 400, body: { detail: "Invalid cursor" } };
    }
  }
  const limit = Math.max(1, Math.min(50, Number(url.searchParams.get("limit")) || 20));
  const hasMore = offset + limit < filtered.length;
  const nextCursor = hasMore ? Buffer.from(JSON.stringify({ v: 1, f: fingerprint, o: offset + limit })).toString("base64url") : undefined;
  return { status: 200, body: { items: filtered.slice(offset, offset + limit), totalCount: filtered.length, hasMore, nextCursor } };
}
