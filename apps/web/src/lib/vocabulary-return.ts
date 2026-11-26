const APP_ORIGIN = "https://vocanova.invalid";
const VOCABULARY_PATH = "/vocabulary";
/** Generous for q (≤100 chars, percent-encoded), filters and a cursor. */
const MAX_RETURN_LENGTH = 4096;

/**
 * Accepts only a relative Vocabulary search URL (`/vocabulary` or
 * `/vocabulary?…`) carried from search results or the map. Anything else —
 * another path, an external host, a fragment, or an oversized value — is
 * rejected so the word detail falls back to the plain search link.
 * Returns null for the default, unfiltered search as well.
 */
export function vocabularySearchReturnHref(value: unknown): string | null {
  if (typeof value !== "string" || !value || value.length > MAX_RETURN_LENGTH)
    return null;
  if (
    !(value === VOCABULARY_PATH || value.startsWith(`${VOCABULARY_PATH}?`)) ||
    value.includes("#") ||
    value.includes("\\")
  )
    return null;
  try {
    const destination = new URL(value, APP_ORIGIN);
    if (
      destination.origin !== APP_ORIGIN ||
      destination.pathname !== VOCABULARY_PATH ||
      destination.hash
    )
      return null;
    if (!destination.search) return null;
    return `${destination.pathname}${destination.search}`;
  } catch {
    return null;
  }
}

/** Builds a word-detail href, carrying a nondefault search return state. */
export function vocabularyWordHref(
  wordSlug: string,
  anchor: string,
  searchReturn?: string | null,
): string {
  const returnTo = vocabularySearchReturnHref(searchReturn);
  const query = returnTo
    ? `?${new URLSearchParams({ fromSearch: returnTo })}`
    : "";
  return `/vocabulary/${encodeURIComponent(wordSlug)}${query}#${anchor}`;
}
