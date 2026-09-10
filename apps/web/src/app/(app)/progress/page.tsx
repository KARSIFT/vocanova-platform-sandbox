import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";

const SAVED_VOCABULARY_DISPLAY_LIMIT = 10;

function formatWeekdayLabel(localDate: string): string {
  const date = new Date(`${localDate}T00:00:00Z`);
  return new Intl.DateTimeFormat(undefined, {
    weekday: "short",
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

  const historyWithLabels = completionHistory.map((day) => ({
    ...day,
    label: formatWeekdayLabel(day.localDate),
  }));

  return (
    <PageContainer>
      <Eyebrow>Your practice record</Eyebrow>
      <h1 className="mt-[var(--spacing-xs)] text-3xl font-bold tracking-tight text-neutral-900">
        Progress
      </h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Every practice session builds your confidence.
      </p>

      <div className="mt-[var(--spacing-lg)] grid grid-cols-2 gap-[var(--spacing-md)]">
        <Surface aria-labelledby="confidence-points-heading" tone="primary">
          <h2
            id="confidence-points-heading"
            className="text-sm font-medium text-primary-900"
          >
            Confidence Points
          </h2>
          <p className="mt-[var(--spacing-xs)] text-3xl font-semibold text-primary-900">
            {confidencePointsTotal.toLocaleString()}
          </p>
        </Surface>

        <Surface aria-labelledby="streak-heading">
          <h2
            id="streak-heading"
            className="text-sm font-medium text-neutral-700"
          >
            Your streaks
          </h2>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-800">
            {currentStreakDays}-day streak
          </p>
          <p className="mt-[var(--spacing-xs)] text-sm text-neutral-600">
            Best: {longestStreakDays} days
          </p>
        </Surface>
      </div>

      <Surface
        aria-labelledby="completion-history-heading"
        className="mt-[var(--spacing-md)]"
      >
        <h2
          id="completion-history-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          This week
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          Keep showing up — each day counts.
        </p>
        {historyWithLabels.length > 0 ? (
          <ul className="mt-[var(--spacing-md)] grid grid-cols-7 gap-[var(--spacing-xs)]">
            {historyWithLabels.map((day) => (
              <li
                key={day.localDate}
                className={`min-w-0 rounded-md py-[var(--spacing-sm)] text-center ${day.completed ? "bg-primary-100 text-primary-900" : "bg-neutral-100 text-neutral-700"}`}
              >
                <p className="text-xs font-semibold">{day.label}</p>
                <p className="mt-[var(--spacing-xs)] text-xs">
                  {day.completed ? "Done" : "Rest"}
                </p>
              </li>
            ))}
          </ul>
        ) : (
          <div className="mt-[var(--spacing-md)]">
            <p className="text-base text-neutral-700">
              No mission history yet. Complete your first daily mission to start
              building your streak.
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
        className="mt-[var(--spacing-md)]"
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
    </PageContainer>
  );
}
