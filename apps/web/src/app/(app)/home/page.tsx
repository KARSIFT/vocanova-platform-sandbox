import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";

// VOC-019 P4-pending mock fields: mission target, reviewed-today count, and
// streak have no P1/P2 equivalent and stay mocked pending P4.
// VOC-027-D05: the due-review count is wired to the real P2 due-queue and is
// no longer part of MOCK_HOME_STATE.
const MOCK_HOME_STATE = {
  missionTargetWords: 10,
  reviewedWordsToday: 3,
  currentStreakDays: 5,
} as const;

export default async function HomePage() {
  const client = await createServerApiClient();
  let savedWordsResponse: Awaited<ReturnType<typeof client.listSavedWords>>;
  let dueResponse: Awaited<ReturnType<typeof client.listDueWords>>;
  try {
    savedWordsResponse = await client.listSavedWords({ limit: 10 });
    dueResponse = await client.listDueWords({ limit: 1 });
  } catch (error) {
    requireAuthRedirect(error, "/home");
  }

  const { items: savedWords } = savedWordsResponse.data;
  const dueReviewWords = dueResponse.data.totalCount;

  const { missionTargetWords, reviewedWordsToday, currentStreakDays } =
    MOCK_HOME_STATE;

  const missionProgressPercent = Math.min(
    100,
    Math.round((reviewedWordsToday / missionTargetWords) * 100),
  );

  return (
    <div className="p-[var(--spacing-lg)]">
      <section
        aria-labelledby="todays-mission-heading"
        className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h1
          id="todays-mission-heading"
          className="text-xl font-semibold text-neutral-900"
        >
          Today&apos;s Mission
        </h1>
        <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
          Review target: {missionTargetWords} words
        </p>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          {reviewedWordsToday} of {missionTargetWords} words reviewed today (
          {missionProgressPercent}
          %)
        </p>

        <div
          aria-hidden="true"
          className="mt-[var(--spacing-md)] h-[var(--spacing-sm)] w-full rounded-full bg-neutral-200"
        >
          <div
            className="h-full rounded-full bg-primary-600"
            style={{ width: `${missionProgressPercent}%` }}
          />
        </div>
      </section>

      <section
        aria-labelledby="saved-words-heading"
        className="mt-[var(--spacing-lg)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h2
          id="saved-words-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          Saved words
        </h2>
        {savedWords.length > 0 ? (
          <ul className="mt-[var(--spacing-sm)] space-y-[var(--spacing-xs)]">
            {savedWords.map((savedWord) => (
              <li
                key={savedWord.userWordId}
                className="rounded-md p-[var(--spacing-sm)]"
              >
                <p className="font-medium text-neutral-900">
                  {savedWord.wordText}
                  <span className="ml-[var(--spacing-xs)] text-sm font-normal text-neutral-600">
                    {savedWord.partOfSpeech}
                  </span>
                </p>
                <p className="text-base text-neutral-700">
                  {savedWord.shortDefinition}
                </p>
              </li>
            ))}
          </ul>
        ) : (
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            You haven&apos;t saved any words yet. Explore a journey to start
            building your vocabulary.
          </p>
        )}
      </section>

      <p className="mt-[var(--spacing-lg)] text-base text-neutral-800">
        {currentStreakDays}-day streak
      </p>

      <p className="mt-[var(--spacing-sm)] text-base text-neutral-800">
        {dueReviewWords} words due today
      </p>

      <div className="mt-[var(--spacing-lg)] flex flex-wrap gap-[var(--spacing-md)]">
        <Link
          href="/discover"
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Go to Journey
        </Link>

        <Link
          href="/reviews"
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Start review
        </Link>
      </div>
    </div>
  );
}
