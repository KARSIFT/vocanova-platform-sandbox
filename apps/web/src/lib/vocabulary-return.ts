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

/** Review stages accepted by the Saved vocabulary filter. */
export const SAVED_WORD_STAGES = [
  "new",
  "learning",
  "reviewing",
  "mastered",
  "ignored",
  "archived",
] as const;

const SAVED_WORDS_PATH = "/words";
const SAVED_WORDS_KEYS = ["q", "stage", "due", "after"] as const;
const MAX_FILTER_QUERY_CHARS = 100;
const LIST_RETURN_PATTERN = /^\/lists\/([^/?#\\]+)(?:\?([^#\\]*))?$/;
const LIST_ID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
// The URL parser silently strips tabs/newlines, so reject raw controls first.
// eslint-disable-next-line no-control-regex
const CONTROL_CHARACTERS = /[\u0000-\u001f\u007f]/;

/** Shared raw checks matching vocabularySearchReturnHref's constraints. */
function parseReturnPath(value: unknown, path: string): URL | null {
  if (typeof value !== "string" || !value || value.length > MAX_RETURN_LENGTH)
    return null;
  if (
    !(value === path || value.startsWith(`${path}?`)) ||
    value.includes("#") ||
    value.includes("\\") ||
    CONTROL_CHARACTERS.test(value)
  )
    return null;
  try {
    const destination = new URL(value, APP_ORIGIN);
    if (
      destination.origin !== APP_ORIGIN ||
      destination.pathname !== path ||
      destination.hash
    )
      return null;
    return destination;
  } catch {
    return null;
  }
}

function hasDuplicateKeys(params: URLSearchParams): boolean {
  const keys = Array.from(params.keys());
  return new Set(keys).size !== keys.length;
}

/**
 * Accepts only a relative Saved vocabulary URL (`/words?…`) with known
 * filters (q, stage, due=true) and an optional cursor. Returns a canonical
 * href, or null for the default collection and anything malformed.
 */
export function savedWordsReturnHref(value: unknown): string | null {
  const destination = parseReturnPath(value, SAVED_WORDS_PATH);
  if (!destination) return null;
  const params = destination.searchParams;
  if (hasDuplicateKeys(params)) return null;
  if (
    Array.from(params.keys()).some(
      (key) => !(SAVED_WORDS_KEYS as readonly string[]).includes(key),
    )
  )
    return null;
  const q = params.get("q");
  const stage = params.get("stage");
  const due = params.get("due");
  const after = params.get("after");
  const canonical = new URLSearchParams();
  if (q !== null) {
    const trimmed = q.trim();
    if (
      !trimmed ||
      Array.from(trimmed).length > MAX_FILTER_QUERY_CHARS ||
      trimmed.includes("\0")
    )
      return null;
    canonical.set("q", trimmed);
  }
  if (stage !== null) {
    if (!(SAVED_WORD_STAGES as readonly string[]).includes(stage)) return null;
    canonical.set("stage", stage);
  }
  if (due !== null) {
    if (due !== "true") return null;
    canonical.set("due", "true");
  }
  if (after !== null) {
    if (!after) return null;
    canonical.set("after", after);
  }
  return canonical.size ? `${SAVED_WORDS_PATH}?${canonical}` : null;
}

/** Saved collection href with the same filters but restarted pagination. */
export function savedWordsRestartHref(value: unknown): string {
  const returnTo = savedWordsReturnHref(value);
  if (!returnTo) return SAVED_WORDS_PATH;
  const destination = new URL(returnTo, APP_ORIGIN);
  destination.searchParams.delete("after");
  return `${destination.pathname}${destination.search}`;
}

/** Builds a saved-detail href, carrying a nondefault Saved return state. */
export function savedWordHref(
  userWordId: string,
  savedReturn?: string | null,
): string {
  const returnTo = savedWordsReturnHref(savedReturn);
  const query = returnTo
    ? `?${new URLSearchParams({ fromSaved: returnTo })}`
    : "";
  return `/words/${encodeURIComponent(userWordId)}${query}`;
}

/** A bounded list-member filter from the URL, or "" when absent/unsafe. */
export function listSearchQuery(value: unknown): string {
  if (typeof value !== "string") return "";
  const q = value.trim();
  return q &&
    Array.from(q).length <= MAX_FILTER_QUERY_CHARS &&
    !CONTROL_CHARACTERS.test(q)
    ? q
    : "";
}

/** Builds `/lists/<id>` with an optional filter; null for a non-UUID id. */
export function listPageHref(listId: string, q?: string): string | null {
  if (!LIST_ID_PATTERN.test(listId)) return null;
  const query = listSearchQuery(q);
  return `/lists/${listId}${query ? `?${new URLSearchParams({ q: query })}` : ""}`;
}

/**
 * Accepts only a relative personal-list URL (`/lists/<uuid>` with an
 * optional `q`) and returns its canonical form, or null otherwise.
 */
export function listReturnHref(value: unknown): string | null {
  if (typeof value !== "string" || value.length > MAX_RETURN_LENGTH) return null;
  const listId = LIST_RETURN_PATTERN.exec(value)?.[1];
  if (!listId || !LIST_ID_PATTERN.test(listId)) return null;
  const destination = parseReturnPath(value, `/lists/${listId}`);
  if (!destination) return null;
  const params = destination.searchParams;
  if (hasDuplicateKeys(params)) return null;
  if (Array.from(params.keys()).some((key) => key !== "q")) return null;
  const q = params.get("q");
  const query = listSearchQuery(q);
  if (q !== null && (!query || query !== q)) return null;
  return listPageHref(listId, query);
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
