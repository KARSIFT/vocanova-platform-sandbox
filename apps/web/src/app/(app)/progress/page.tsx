import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";

const SAVED_VOCABULARY_DISPLAY_LIMIT = 10;

function formatActivityDate(localDate: string): string {
  const date = new Date(`${localDate}T00:00:00Z`);
  return new Intl.DateTimeFormat(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(date);
}

export default async function ProgressPage() {
  const client = await createServerApiClient();
  let savedWordsResponse: Awaited<ReturnType<typeof client.listSavedWords>>;
  let progressResponse: Awaited<ReturnType<typeof client.getProgress>>;
  try {
    savedWordsResponse = await client.listSavedWords({
      limit: SAVED_VOCABULARY_DISPLAY_LIMIT,
    });
    progressResponse = await client.getProgress();
  } catch (error) {
    requireAuthRedirect(error, "/progress");
  }

  const { items: savedWords } = savedWordsResponse.data;

  const {
    confidencePointsBalance: confidencePointsTotal,
    streak,
    completionHistory,
  } = progressResponse.data;
  const {
    currentStreakCount: currentStreakDays,
    longestStreakCount: longestStreakDays,
  } = streak;

  const historyWithLabels = [...completionHistory]
    .sort((first, second) => first.localDate.localeCompare(second.localDate))
    .map((day) => ({ ...day, label: formatActivityDate(day.localDate) }));

  return (
    <PageContainer className="max-w-[64rem]">
      <Eyebrow>Your practice record</Eyebrow>
      <h1 className="mt-[var(--spacing-xs)] text-3xl font-bold tracking-tight text-neutral-900">
        Progress
      </h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        See the practice you&apos;ve completed and choose a useful next step.
      </p>

      <Surface
        aria-labelledby="sentence-history-heading"
        className="mt-[var(--spacing-lg)] bg-secondary-50"
      >
        <div className="flex flex-wrap items-center justify-between gap-[var(--spacing-md)]">
          <div>
            <h2
              id="sentence-history-heading"
              className="text-lg font-semibold text-neutral-900"
            >
              Sentence practice
            </h2>
            <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
              Revisit your writing and the feedback that can guide your next
              sentence.
            </p>
          </div>
          <Link
            href="/progress/sentences"
            className="inline-flex min-h-11 items-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-white hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            View sentence history
          </Link>
        </div>
      </Surface>

      <Surface
        className="mt-[var(--spacing-md)]"
        tone="primary"
        aria-label="Learning summary"
      >
        <div className="grid grid-cols-2 gap-[var(--spacing-md)]">
          <div>
            <h2
              id="confidence-points-heading"
              className="text-sm font-medium text-primary-900"
            >
              Confidence Points
            </h2>
            <p className="mt-[var(--spacing-xs)] text-3xl font-semibold tabular-nums text-primary-900">
              {confidencePointsTotal.toLocaleString()}
            </p>
          </div>
          <div className="border-l border-primary-200 pl-[var(--spacing-md)]">
            <h2
              id="streak-heading"
              className="text-sm font-medium text-primary-900"
            >
              Your streaks
            </h2>
            <p className="mt-[var(--spacing-sm)] text-base text-primary-900">
              {currentStreakDays}-day streak
            </p>
            <p className="mt-[var(--spacing-xs)] text-sm text-primary-900">
              Best: {longestStreakDays} days
            </p>
          </div>
        </div>
        <p className="mt-[var(--spacing-md)] text-sm text-primary-900">
          Earn points by saving words, reviewing, and practising sentences.
        </p>
      </Surface>

      <div className="mt-[var(--spacing-md)] lg:grid lg:grid-cols-2 lg:items-start lg:gap-[var(--spacing-md)]">
        <Surface aria-labelledby="completion-history-heading">
          <h2
            id="completion-history-heading"
            className="text-lg font-semibold text-neutral-900"
          >
            Recent activity
          </h2>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            Recorded mission days, shown in date order.
          </p>
          {historyWithLabels.length > 0 ? (
            <ul className="mt-[var(--spacing-md)] space-y-[var(--spacing-xs)]">
              {historyWithLabels.map((day) => (
                <li
                  key={day.localDate}
                  className={`flex min-w-0 items-center justify-between gap-[var(--spacing-sm)] rounded-md border-l-4 px-[var(--spacing-md)] py-[var(--spacing-sm)] ${day.completed ? "border-primary-700 bg-primary-50 text-primary-900" : "border-neutral-400 bg-neutral-100 text-neutral-700"}`}
                >
                  <time
                    dateTime={day.localDate}
                    className="shrink-0 text-sm font-semibold"
                  >
                    {day.label}
                  </time>
                  <p className="min-w-0 text-right text-sm">
                    {day.completed ? "Completed or protected" : "Not complete"}
                  </p>
                </li>
              ))}
            </ul>
          ) : (
            <div className="mt-[var(--spacing-md)]">
              <p className="text-base text-neutral-700">
                No mission history yet. Complete your first daily mission to
                start building your streak.
              </p>
              <Link
                href="/home"
                className="mt-[var(--spacing-sm)] inline-flex min-h-11 items-center font-semibold text-primary-700"
              >
                Go to today&apos;s mission
              </Link>
            </div>
          )}
        </Surface>

        <Surface
          aria-labelledby="saved-vocabulary-heading"
          className="mt-[var(--spacing-md)] lg:mt-0"
        >
          <h2
            id="saved-vocabulary-heading"
            className="text-lg font-semibold text-neutral-900"
          >
            Recent saved vocabulary
          </h2>
          {savedWords.length > 0 ? (
            <>
              <ul className="mt-[var(--spacing-md)] grid grid-cols-2 gap-[var(--spacing-xs)] sm:grid-cols-3">
                {savedWords.map((savedWord) => (
                  <li
                    key={savedWord.userWordId}
                    className="min-w-0 rounded-lg bg-neutral-50"
                  >
                    <Link
                      href={`/words/${savedWord.userWordId}`}
                      className="flex min-h-11 items-center px-[var(--spacing-sm)] py-[var(--spacing-xs)] wrap-anywhere font-semibold text-primary-800 hover:text-primary-600"
                    >
                      {savedWord.wordText}
                    </Link>
                  </li>
                ))}
              </ul>
              <Link
                href="/words"
                className="mt-[var(--spacing-md)] inline-flex min-h-[var(--spacing-2xl)] items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                View all saved vocabulary
              </Link>
            </>
          ) : (
            <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
              No saved words yet. Save words from a journey to track your
              vocabulary here.
            </p>
          )}
        </Surface>
      </div>
    </PageContainer>
  );
}
