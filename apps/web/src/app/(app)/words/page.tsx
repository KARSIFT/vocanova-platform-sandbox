import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";

import { RemoveSavedWordButton } from "./_components/remove-saved-word-button";
import {
  formatSavedWordStatus,
  getSavedWordsView,
} from "./_components/saved-word-view";

const PAGE_SIZE = 20;

interface SavedWordsPageProps {
  searchParams: Promise<{ after?: string }>;
}

export default async function SavedWordsPage({
  searchParams,
}: SavedWordsPageProps) {
  const { after } = await searchParams;
  const client = await createServerApiClient();
  let response: Awaited<ReturnType<typeof client.listSavedWords>>;
  try {
    response = await client.listSavedWords({ after, limit: PAGE_SIZE });
  } catch (error) {
    requireAuthRedirect(
      error,
      after ? `/words?after=${encodeURIComponent(after)}` : "/words",
    );
  }

  const { items, nextCursor } = response.data;
  const view = getSavedWordsView(items.length, Boolean(after));

  return (
    <div className="p-[var(--spacing-lg)]">
      <div className="flex flex-wrap items-start justify-between gap-[var(--spacing-md)]">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-900">
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

      {view === "empty" ? (
        <section className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            Your vocabulary starts here
          </h2>
          <p className="mt-[var(--spacing-sm)] max-w-lg text-base text-neutral-700">
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
            href="/words"
            className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Back to newest words
          </Link>
        </section>
      ) : (
        <>
          <ul className="mt-[var(--spacing-lg)] grid grid-cols-1 gap-[var(--spacing-md)] md:grid-cols-2">
            {items.map((savedWord) => {
              const statusLabel = formatSavedWordStatus(savedWord.status);
              return (
                <li
                  key={savedWord.userWordId}
                  className="flex flex-col justify-between gap-[var(--spacing-md)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
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
                href="/words"
                className="inline-flex min-h-[var(--spacing-2xl)] items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                Back to newest
              </Link>
            ) : (
              <span />
            )}
            {nextCursor ? (
              <Link
                href={`/words?after=${encodeURIComponent(nextCursor)}`}
                className="inline-flex min-h-[var(--spacing-2xl)] items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                More saved words
              </Link>
            ) : null}
          </nav>
        </>
      )}
    </div>
  );
}
