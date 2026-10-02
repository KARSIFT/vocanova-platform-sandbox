import Link from "next/link";
import {
  ApiResponseError,
  type VocabularySearchResponse,
} from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { MeaningSaveButton } from "../discover/[situation]/[word]/_components/meaning-save-button";
import { VocabularyMap } from "../_components/vocabulary-map";

const categories = [
  ["", "All situations"],
  ["daily_life", "Daily life"],
  ["travel", "Travel"],
  ["work", "Work"],
  ["study", "Study"],
  ["social", "Social"],
] as const;
const levels = ["", "a1", "a2", "b1", "b2", "c1", "unknown"] as const;
const inputStyle =
  "mt-2 min-h-12 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

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
  let invalidCursor = false;
  let invalidSearch = false;
  try {
    if (!invalidQuery)
      data = (
        await (
          await createServerApiClient()
        ).searchVocabulary({ q, category, level, knowledge, after, limit: 20 })
      ).data;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 400) {
      invalidCursor = Boolean(after);
      invalidSearch = !after;
    } else if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, "/vocabulary");
  }
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
    <PageContainer className="max-w-[64rem]">
      <Link
        href="/discover"
        className="inline-flex min-h-11 items-center font-semibold text-primary-700"
      >
        Back to Journey
      </Link>
      <div className="mt-3 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-neutral-900">
            {mapView ? "Your vocabulary map" : "Find your next word"}
          </h1>
          <p className="mt-2 text-neutral-700">
            Explore practical words, understand their meanings, and keep the
            ones you want to learn.
          </p>
        </div>
        <Link
          href="/words"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700"
        >
          Saved vocabulary
        </Link>
      </div>
      <nav
        aria-label="Vocabulary views"
        className="mt-4 flex flex-wrap gap-4 text-primary-700"
      >
        <Link
          href="/vocabulary"
          aria-current={!mapView ? "page" : undefined}
          className="inline-flex min-h-12 items-center font-semibold underline-offset-4 aria-[current=page]:underline"
        >
          Word search
        </Link>
        <Link
          href="/vocabulary?view=map"
          aria-current={mapView ? "page" : undefined}
          className="inline-flex min-h-12 items-center font-semibold underline-offset-4 aria-[current=page]:underline"
        >
          Knowledge map
        </Link>
        <Link
          href="/vocabulary/check"
          className="inline-flex min-h-12 items-center font-semibold underline-offset-4"
        >
          Find your starting words
        </Link>
      </nav>
      <form
        action="/vocabulary"
        role="search"
        className="my-6 rounded-2xl border border-neutral-200 bg-white p-4 sm:p-5"
      >
        {mapView && <input type="hidden" name="view" value="map" />}
        <label
          className="block font-semibold text-neutral-900"
          htmlFor="vocabulary-query"
        >
          Search words and meanings
        </label>
        <input
          id="vocabulary-query"
          name="q"
          type="search"
          defaultValue={q}
          maxLength={100}
          placeholder="Try a word or a meaning"
          className={inputStyle}
        />
        <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-[1fr_1fr_1fr_auto] sm:items-end">
          <label className="block text-sm font-semibold text-neutral-700">
            Situation
            <select
              name="category"
              defaultValue={category}
              className={inputStyle}
            >
              {categories.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </label>
          <label className="block text-sm font-semibold text-neutral-700">
            Level
            <select name="level" defaultValue={level} className={inputStyle}>
              {levels.map((value) => (
                <option key={value} value={value}>
                  {value === ""
                    ? "All levels"
                    : value === "unknown"
                      ? "Not specified"
                      : value.toUpperCase()}
                </option>
              ))}
            </select>
          </label>
          <label className="block text-sm font-semibold text-neutral-700">
            My knowledge
            <select
              name="knowledge"
              defaultValue={knowledge}
              className={inputStyle}
            >
              <option value="">All words</option>
              <option value="known">Already known</option>
              <option value="saved">Saved to learn</option>
              <option value="unexplored">Not explored yet</option>
            </select>
          </label>
          <button
            type="submit"
            className="min-h-12 rounded-xl bg-primary-700 px-6 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Search
          </button>
        </div>
      </form>
      {invalidQuery || invalidSearch ? (
        <p
          role="alert"
          className="rounded-xl bg-secondary-50 p-5 text-neutral-900"
        >
          {invalidQuery
            ? "Use a search of 100 characters or fewer, without special control characters."
            : "We could not use that search. Check your words and filters, then try again."}
        </p>
      ) : invalidCursor ? (
        <div
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
        </div>
      ) : !data ? (
        <div
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
        </div>
      ) : (
        data && (
          <>
            <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
              <p className="text-neutral-700">
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
            {data.items.length && mapView ? (
              <VocabularyMap items={data.items} />
            ) : data.items.length ? (
              <ul className="grid gap-4 md:grid-cols-2">
                {data.items.map((item) => (
                  <li
                    key={item.meaningId}
                    className="flex flex-col rounded-2xl border border-neutral-200 bg-white p-5"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <Link
                        href={`/vocabulary/${encodeURIComponent(item.wordSlug)}`}
                        className="inline-flex min-h-11 items-center text-xl font-bold text-primary-700"
                      >
                        {item.wordText}
                      </Link>
                      <span className="pt-3 text-sm text-neutral-600">
                        {item.difficultyLevel &&
                        item.difficultyLevel !== "unknown"
                          ? item.difficultyLevel.toUpperCase()
                          : ""}
                      </span>
                    </div>
                    <p className="text-sm text-neutral-600">
                      {item.partOfSpeech}
                    </p>
                    <p className="mt-3 grow text-neutral-900">
                      {item.shortDefinition}
                    </p>
                    {item.selfReportedKnown && (
                      <p className="mt-3 text-sm font-semibold text-primary-700">
                        Already known · your assessment
                      </p>
                    )}
                    <div className="mt-4 flex items-center justify-between gap-3">
                      <span className="text-sm text-neutral-600">
                        {item.due
                          ? "Due for review"
                          : item.saved
                            ? "In your vocabulary"
                            : ""}
                      </span>
                      <MeaningSaveButton
                        meaningId={item.meaningId}
                        source="search"
                        initialSaved={item.saved}
                        wordText={item.wordText}
                        shortDefinition={item.shortDefinition}
                      />
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <div className="rounded-2xl border border-neutral-200 bg-white p-6">
                <h2 className="text-xl font-semibold text-neutral-900">
                  No matching words yet
                </h2>
                <p className="mt-2 text-neutral-700">
                  Try a shorter search or another situation. We are growing the
                  vocabulary collection.
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
                <Link
                  href={resultURL(data.nextCursor)}
                  className="ml-auto inline-flex min-h-12 items-center rounded-xl border border-neutral-300 bg-white px-5 font-semibold text-primary-700"
                >
                  Next results
                </Link>
              )}
            </nav>
          </>
        )
      )}
    </PageContainer>
  );
}
