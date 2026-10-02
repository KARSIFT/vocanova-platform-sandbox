import Link from "next/link";
import {
  ApiResponseError,
  type ListSavedWordsResponse,
} from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer } from "@/ui/surface";

import { RemoveSavedWordButton } from "./_components/remove-saved-word-button";
import { ReloadSavedWordsButton } from "./_components/reload-saved-words-button";
import { getSavedWordsView } from "./_components/saved-word-view";

const PAGE_SIZE = 20;
const stages = [
  "new",
  "learning",
  "reviewing",
  "mastered",
  "ignored",
  "archived",
] as const;
const stageLabels: Record<string, string> = {
  new: "New",
  learning: "Learning",
  reviewing: "Reviewing",
  mastered: "Mastered",
  ignored: "Ignored",
  archived: "Archived",
};
const inputStyle =
  "mt-2 min-h-12 w-full min-w-0 rounded-xl border border-neutral-300 bg-white px-3 py-2 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const actionStyle =
  "inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const linkStyle =
  "inline-flex min-h-12 items-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

interface SavedWordsPageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export default async function SavedWordsPage({
  searchParams,
}: SavedWordsPageProps) {
  const query = await searchParams;
  const single = (key: string) =>
    typeof query[key] === "string" ? (query[key] as string) : "";
  const q = single("q").trim();
  const stage = stages.find((value) => value === single("stage"));
  const due = single("due") === "true";
  const after = single("after");
  const invalidQuery = Array.from(q).length > 100 || q.includes("\0");
  const invalidFilters =
    (Boolean(single("stage")) && !stage) ||
    !["", "true", "false"].includes(single("due")) ||
    ["q", "stage", "due", "after"].some((key) => Array.isArray(query[key]));
  const filtered = Boolean(q || stage || due);
  function resultsURL(cursor?: string) {
    const params = new URLSearchParams();
    if (q) params.set("q", q);
    if (stage) params.set("stage", stage);
    if (due) params.set("due", "true");
    if (cursor) params.set("after", cursor);
    return `/words${params.size ? `?${params}` : ""}`;
  }
  const client = await createServerApiClient();
  let data: ListSavedWordsResponse | null = null;
  let invalidCursor = false;
  let rejectedFilters = false;
  try {
    if (!invalidQuery && !invalidFilters)
      data = (
        await client.listSavedWords({ q, stage, due, after, limit: PAGE_SIZE })
      ).data;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, resultsURL(after));
    if (error instanceof ApiResponseError && error.status === 400) {
      invalidCursor = Boolean(after);
      rejectedFilters = !after;
    }
    if (error instanceof ApiResponseError && error.status === 422)
      rejectedFilters = true;
  }

  const items = data?.items ?? [];
  const nextCursor = data?.nextCursor;
  const view = getSavedWordsView(items.length, Boolean(after));
  const emptyCollection =
    data?.totalCount === 0 && !filtered && view === "empty";

  return (
    <PageContainer>
      <div className="flex flex-wrap items-start justify-between gap-[var(--spacing-md)]">
        <div>
          <Eyebrow>Words you chose</Eyebrow>
          <h1 className="mt-[var(--spacing-xs)] text-3xl font-bold tracking-tight text-neutral-900">
            Saved vocabulary
          </h1>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            Revisit words, practice them in a sentence, or remove what you no
            longer need.
          </p>
        </div>
        <Link
          href="/discover"
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Discover new words
        </Link>
      </div>

      {!emptyCollection && (
        <form
          key={resultsURL()}
          action="/words"
          role="search"
          aria-label="Saved vocabulary filters"
          className="my-6 rounded-2xl border border-neutral-200 bg-white p-4 sm:p-5"
        >
          <label
            htmlFor="saved-word-query"
            className="block font-semibold text-neutral-900"
          >
            Search saved words and meanings
          </label>
          <input
            id="saved-word-query"
            name="q"
            type="search"
            defaultValue={q}
            maxLength={100}
            className={inputStyle}
            aria-describedby="saved-word-search-help"
          />
          <p
            id="saved-word-search-help"
            className="mt-2 text-sm text-neutral-600"
          >
            Find a word or a part of its meaning. Use up to 100 characters.
          </p>
          <div className="mt-4 grid min-w-0 gap-4 sm:grid-cols-[1fr_1fr_auto] sm:items-end">
            <label className="min-w-0 font-semibold text-neutral-900">
              Review stage
              <select
                name="stage"
                defaultValue={stage ?? ""}
                className={inputStyle}
              >
                <option value="">All stages</option>
                {stages.map((value) => (
                  <option key={value} value={value}>
                    {stageLabels[value]}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex min-h-12 cursor-pointer items-center gap-3 rounded-xl px-1 text-base text-neutral-900">
              <input
                name="due"
                type="checkbox"
                value="true"
                defaultChecked={due}
                className="h-5 w-5 shrink-0 accent-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              />
              Due for review now
            </label>
            <button type="submit" className={actionStyle}>
              Apply filters
            </button>
          </div>
          <div className="mt-2 flex flex-wrap items-center justify-between gap-x-4">
            <p className="text-sm text-neutral-600">
              Stages describe your review progress, not your English level.
            </p>
            {(filtered || invalidFilters || invalidQuery) && (
              <Link href="/words" className={linkStyle}>
                Clear filters
              </Link>
            )}
          </div>
        </form>
      )}

      {invalidQuery || invalidFilters || rejectedFilters ? (
        <section aria-labelledby="saved-filters-invalid" className="py-6">
          <h2
            id="saved-filters-invalid"
            className="text-xl font-semibold text-neutral-900"
          >
            Check your filters
          </h2>
          <p role="alert" className="mt-2 text-neutral-700">
            {invalidQuery
              ? "Keep your search to 100 characters or fewer and remove any unsupported characters."
              : "Choose a listed review stage and try your search again."}
          </p>
        </section>
      ) : invalidCursor ? (
        <section aria-labelledby="saved-page-changed" className="py-6">
          <h2
            id="saved-page-changed"
            className="text-xl font-semibold text-neutral-900"
          >
            Start from the first page
          </h2>
          <p className="mt-2 text-neutral-700">
            This page link no longer matches your filters. Keep your filters and
            load the latest results.
          </p>
          <Link href={resultsURL()} className={`${linkStyle} mt-3`}>
            Restart these results
          </Link>
        </section>
      ) : !data ? (
        <section aria-labelledby="saved-unavailable" className="py-6">
          <h2
            id="saved-unavailable"
            className="text-xl font-semibold text-neutral-900"
          >
            Saved vocabulary is unavailable
          </h2>
          <p role="status" className="mt-2 text-neutral-700">
            We could not load your words. Try again with the same filters.
          </p>
          <ReloadSavedWordsButton />
        </section>
      ) : view === "empty" && filtered ? (
        <section aria-labelledby="saved-no-matches" className="py-6">
          <h2
            id="saved-no-matches"
            className="text-xl font-semibold text-neutral-900"
          >
            No saved words match these filters
          </h2>
          <p className="mt-2 text-neutral-700">
            Try a different word or stage, or turn off “Due for review now”.
          </p>
          <Link href="/words" className={`${linkStyle} mt-3`}>
            Show all saved words
          </Link>
        </section>
      ) : view === "empty" ? (
        <section className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            Your vocabulary starts here
          </h2>
          {/* Explicit width avoids the named spacing-token collision. */}
          <p className="mt-[var(--spacing-sm)] max-w-[32rem] text-base text-neutral-700">
            Save a useful word from Journey and it will appear here, ready for
            review and sentence practice.
          </p>
          <Link
            href="/discover"
            className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Explore Journey
          </Link>
        </section>
      ) : view === "exhausted" ? (
        <section className="py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            You&apos;ve reached the end
          </h2>
          <Link
            href={resultsURL()}
            className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Back to newest words
          </Link>
        </section>
      ) : (
        <>
          <p role="status" className="text-sm text-neutral-700">
            Showing {items.length} of {data.totalCount}{" "}
            {filtered ? "matching" : "saved"}{" "}
            {data.totalCount === 1 ? "meaning" : "meanings"}.
          </p>
          <ul
            aria-label="Saved vocabulary results"
            className="mt-[var(--spacing-lg)] grid grid-cols-1 gap-[var(--spacing-md)] md:grid-cols-2"
          >
            {items.map((savedWord) => {
              const statusLabel =
                stageLabels[savedWord.reviewState] ?? "Stage unavailable";
              return (
                <li
                  key={savedWord.userWordId}
                  className="flex flex-col justify-between gap-[var(--spacing-md)] rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm"
                >
                  <div>
                    <div className="flex flex-wrap items-baseline gap-[var(--spacing-xs)]">
                      <h2 className="text-lg font-semibold text-neutral-900">
                        <Link
                          href={`/words/${savedWord.userWordId}`}
                          className="rounded-sm hover:text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
                        >
                          {savedWord.wordText}
                        </Link>
                      </h2>
                      <span className="text-sm text-neutral-600">
                        {savedWord.partOfSpeech}
                      </span>
                    </div>
                    <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                      {savedWord.shortDefinition}
                    </p>
                    {statusLabel ? (
                      <p className="mt-[var(--spacing-sm)] text-sm font-semibold text-primary-700">
                        {statusLabel}
                      </p>
                    ) : null}
                    <p className="mt-1 text-sm text-neutral-700">
                      {savedWord.due ? "Due now" : "Not due now"}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center justify-between gap-[var(--spacing-sm)]">
                    <Link
                      href={`/words/${savedWord.userWordId}`}
                      aria-label={`Open ${savedWord.wordText} details and sentence practice`}
                      className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
                    >
                      Open word
                    </Link>
                    <RemoveSavedWordButton
                      meaningId={savedWord.meaningId}
                      wordText={savedWord.wordText}
                    />
                  </div>
                </li>
              );
            })}
          </ul>

          <nav
            aria-label="Saved vocabulary pages"
            className="mt-[var(--spacing-lg)] flex flex-wrap justify-between gap-[var(--spacing-md)]"
          >
            {after ? (
              <Link
                href={resultsURL()}
                className="inline-flex min-h-[var(--spacing-2xl)] items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                Back to newest
              </Link>
            ) : (
              <span />
            )}
            {nextCursor ? (
              <Link
                href={resultsURL(nextCursor)}
                className="inline-flex min-h-[var(--spacing-2xl)] items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                More saved words
              </Link>
            ) : null}
          </nav>
        </>
      )}
    </PageContainer>
  );
}
