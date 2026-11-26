import { StatusMessage } from "@/ui/status-message";
import { ActionLink, Button, Select, TextInput } from "@/ui/controls";
import { getPrimaryDefinition } from "@/lib/word-definition";
import { vocabularyWordHref } from "@/lib/vocabulary-return";
import Link from "@/ui/link";
import Form from "next/form";
import {
  ApiResponseError,
  type VocabularySearchResponse,
  type DictionaryEntry,
} from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader } from "@/ui/surface";
import { MeaningSaveButton } from "../discover/[situation]/[word]/_components/meaning-save-button";
import { VocabularyMap } from "../_components/vocabulary-map";
import { DictionaryCard } from "../_components/dictionary-card";
import { meaningAnchor } from "../_components/meaning-comparison";

const categories = [
  ["", "All situations"],
  ["daily_life", "Daily life"],
  ["travel", "Travel"],
  ["work", "Work"],
  ["study", "Study"],
  ["social", "Social"],
] as const;
const levels = ["", "a1", "a2", "b1", "b2", "c1", "unknown"] as const;
const inputStyle = "mt-2";

export default async function VocabularyPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  const single = (key: string) =>
    typeof query[key] === "string" ? (query[key] as string) : "";
  const q = single("q").trim();
  const invalidQuery = Array.from(q).length > 100 || q.includes("\0");
  const category = categories.some(([value]) => value === single("category"))
    ? single("category")
    : "";
  const level = levels.some((value) => value === single("level"))
    ? single("level")
    : "";
  const after = single("after");
  const mapView = single("view") === "map";
  const knowledge =
    (["known", "saved", "unexplored"] as const).find(
      (value) => value === single("knowledge"),
    ) ?? "";
  let data: VocabularySearchResponse | null = null;
  let dictionary: DictionaryEntry | null = null;
  let dictionaryError: "unavailable" | "limited" | null = null;
  let invalidCursor = false;
  let invalidSearch = false;
  const dictionaryEligible =
    !invalidQuery &&
    !mapView &&
    !category &&
    !level &&
    !knowledge &&
    q.length <= 48 &&
    /^[a-zA-Z]+(?:['-][a-zA-Z]+)*$/.test(q);
  if (!invalidQuery) {
    const client = await createServerApiClient();
    const results = await Promise.allSettled([
      client.searchVocabulary({
        q,
        category,
        level,
        knowledge,
        after,
        limit: 20,
      }),
      dictionaryEligible ? client.lookupDictionary(q) : Promise.resolve(null),
    ]);
    // Check both reads before rendering: an expired session must never look like an empty search.
    for (const result of results) {
      if (
        result.status === "rejected" &&
        result.reason instanceof ApiResponseError &&
        result.reason.status === 401
      )
        requireAuthRedirect(result.reason, resultURL(after));
    }
    const [catalog, lookup] = results;
    if (catalog.status === "fulfilled") data = catalog.value.data;
    else if (
      catalog.reason instanceof ApiResponseError &&
      catalog.reason.status === 400
    ) {
      invalidCursor = Boolean(after);
      invalidSearch = !after;
    }
    if (lookup.status === "fulfilled") dictionary = lookup.value?.data ?? null;
    else if (!(
      lookup.reason instanceof ApiResponseError && lookup.reason.status === 404
    ))
      dictionaryError =
        lookup.reason instanceof ApiResponseError &&
        lookup.reason.status === 429
          ? "limited"
          : "unavailable";
  }
  const exactLessonWord = data?.items.some(
    (item) => item.wordText.toLowerCase() === q.toLowerCase(),
  );
  const showDictionary =
    dictionary && !exactLessonWord && !invalidSearch && !invalidCursor;
  function resultURL(cursor?: string) {
    const next = new URLSearchParams();
    if (q) next.set("q", q);
    if (category) next.set("category", category);
    if (level) next.set("level", level);
    if (knowledge) next.set("knowledge", knowledge);
    if (mapView) next.set("view", "map");
    if (cursor) next.set("after", cursor);
    return `/vocabulary${next.size ? `?${next}` : ""}`;
  }
  return (
    <PageContainer width="standard">
      <PageHeader title={mapView ? "Your vocabulary map" : "Word search"} />
      <Form
        action="/vocabulary"
        prefetch={false}
        key={resultURL(after)}
        role="search"
        className="vn-list p-4 sm:p-5"
      >
        {mapView && <input type="hidden" name="view" value="map" />}
        <label className="sr-only" htmlFor="vocabulary-query">
          Search words and meanings
        </label>
        <div className="flex flex-wrap items-center gap-3">
          <TextInput
            id="vocabulary-query"
            name="q"
            type="search"
            defaultValue={q}
            maxLength={100}
            placeholder="Search any English word…"
            className="flex-[1_1_12rem]"
            autoComplete="off"
            spellCheck={false}
          />
          <Button variant="primary" type="submit" className="shrink-0">
            Search
          </Button>
        </div>
        <details
          className="mt-2"
          open={Boolean(category || level || knowledge)}
        >
          <summary className="vn-summary text-sm">Filter lesson words</summary>
          <div className="grid gap-3 sm:grid-cols-3 sm:items-end">
            <label className="block min-w-0 text-sm font-semibold text-neutral-700">
              Situation
              <Select
                name="category"
                defaultValue={category}
                className={inputStyle}
              >
                {categories.map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </Select>
            </label>
            <label className="block min-w-0 text-sm font-semibold text-neutral-700">
              Level
              <Select name="level" defaultValue={level} className={inputStyle}>
                {levels.map((value) => (
                  <option key={value} value={value}>
                    {value === ""
                      ? "All levels"
                      : value === "unknown"
                        ? "Not specified"
                        : value.toUpperCase()}
                  </option>
                ))}
              </Select>
            </label>
            <label className="block min-w-0 text-sm font-semibold text-neutral-700">
              My knowledge
              <Select
                name="knowledge"
                defaultValue={knowledge}
                className={inputStyle}
              >
                <option value="">All words</option>
                <option value="known">Already known</option>
                <option value="saved">Saved to learn</option>
                <option value="unexplored">Not explored yet</option>
              </Select>
            </label>
          </div>
        </details>
      </Form>
      <nav
        aria-label="Vocabulary views"
        className="mt-2 mb-6 flex flex-wrap gap-x-5 gap-y-1 border-b border-neutral-200 pb-2"
      >
        <Link
          href="/vocabulary"
          aria-current={!mapView ? "page" : undefined}
          className="vn-text-link underline-offset-4 aria-[current=page]:underline"
        >
          Word search
        </Link>
        <Link
          href="/vocabulary?view=map"
          aria-current={mapView ? "page" : undefined}
          className="vn-text-link underline-offset-4 aria-[current=page]:underline"
        >
          Knowledge map
        </Link>
        <Link href="/vocabulary/check" className="vn-text-link">
          Find your starting words
        </Link>
        <Link href="/words" className="vn-text-link">
          Saved vocabulary
        </Link>
      </nav>
      {showDictionary && dictionary && <DictionaryCard entry={dictionary} />}
      {dictionaryError && !exactLessonWord && (
        <StatusMessage
          role="status"
          className="mb-5 rounded-xl bg-secondary-50 p-5 text-neutral-900"
        >
          <p>
            {dictionaryError === "limited"
              ? "You’ve searched several words quickly. Try again in a minute."
              : "We could not load the dictionary right now."}
          </p>
          <Link
            href={resultURL()}
            className="inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Try dictionary again
          </Link>
        </StatusMessage>
      )}
      {invalidQuery || invalidSearch ? (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          className="rounded-xl bg-secondary-50 p-5 text-neutral-900"
        >
          {invalidQuery
            ? "Use a search of 100 characters or fewer, without special control characters."
            : "We could not use that search. Check your words and filters, then try again."}
        </StatusMessage>
      ) : invalidCursor ? (
        <StatusMessage
          role="status"
          className="rounded-xl bg-secondary-50 p-5 text-neutral-900"
        >
          <p>This results page is no longer available.</p>
          <Link
            href={resultURL()}
            className="mt-3 inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Load results again
          </Link>
        </StatusMessage>
      ) : !data ? (
        <StatusMessage
          role="status"
          className="rounded-xl bg-secondary-50 p-5 text-neutral-900"
        >
          <p>We could not load vocabulary right now.</p>
          <Link
            href={resultURL()}
            className="mt-3 inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Try searching again
          </Link>
        </StatusMessage>
      ) : (
        data && (
          <>
            {(!showDictionary || data.totalCount > 0) && (
              <div className="mb-3 flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                <p className="text-sm text-neutral-700 [overflow-wrap:anywhere]">
                  {data.totalCount}{" "}
                  {data.totalCount === 1 ? "meaning" : "meanings"}
                  {q ? ` matching “${q}”` : " to explore"}
                </p>
                {(q || category || level || knowledge) && (
                  <Link
                    href={mapView ? "/vocabulary?view=map" : "/vocabulary"}
                    className="inline-flex min-h-11 items-center font-semibold text-primary-700"
                  >
                    Clear filters
                  </Link>
                )}
              </div>
            )}
            {data.items.length && mapView ? (
              <VocabularyMap
                items={data.items}
                searchReturn={resultURL(after)}
              />
            ) : data.items.length ? (
              <ul className="vn-list">
                {data.items.map((item) => (
                  <li key={item.meaningId} className="vn-list-row">
                    <div className="min-w-0 flex-[1_1_20rem] [overflow-wrap:anywhere]">
                      <div className="flex flex-wrap items-baseline gap-x-3">
                        <h2 className="headword text-xl">
                          <Link
                            href={vocabularyWordHref(
                              item.wordSlug,
                              meaningAnchor(item.meaningId),
                              resultURL(after),
                            )}
                            className="vn-text-link"
                          >
                            {item.wordText}
                          </Link>
                        </h2>
                        <span className="pos text-sm">{item.partOfSpeech}</span>
                        {item.difficultyLevel &&
                          item.difficultyLevel !== "unknown" && (
                            <span className="text-sm text-neutral-600">
                              {item.difficultyLevel.toUpperCase()}
                            </span>
                          )}
                      </div>
                      <p className="text-neutral-900">
                        {getPrimaryDefinition(
                          item.shortDefinition,
                          item.learnerDefinition,
                        )}
                      </p>
                      {item.selfReportedKnown && (
                        <p className="mt-2 text-sm text-neutral-600">
                          Already known · your assessment
                        </p>
                      )}
                      {(item.due || item.saved) && (
                        <p className="mt-1 text-sm text-neutral-600">
                          {item.due ? "Due for review" : "In your vocabulary"}
                        </p>
                      )}
                    </div>
                    <div className="min-w-0">
                      <MeaningSaveButton
                        meaningId={item.meaningId}
                        wordSlug={item.wordSlug}
                        source="search"
                        initialSaved={item.saved}
                        wordText={item.wordText}
                        shortDefinition={item.shortDefinition}
                      />
                    </div>
                  </li>
                ))}
              </ul>
            ) : showDictionary || dictionaryError ? null : (
              <div className="vn-empty">
                <h2 className="text-xl font-semibold text-neutral-900">
                  No matching words yet
                </h2>
                <p className="mt-2 text-neutral-700">
                  Try another spelling or a shorter search.
                </p>
                <Link
                  href="/vocabulary"
                  className="mt-3 inline-flex min-h-11 items-center font-semibold text-primary-700"
                >
                  Browse all vocabulary
                </Link>
              </div>
            )}
            <nav
              aria-label="Search results pages"
              className="mt-6 flex flex-wrap justify-between gap-3"
            >
              {after && (
                <Link
                  href={resultURL()}
                  className="inline-flex min-h-12 items-center font-semibold text-primary-700"
                >
                  First results
                </Link>
              )}
              {data.hasMore && data.nextCursor && (
                <ActionLink
                  variant="secondary"
                  href={resultURL(data.nextCursor)}
                  className="ml-auto"
                >
                  Next results
                </ActionLink>
              )}
            </nav>
          </>
        )
      )}
    </PageContainer>
  );
}
