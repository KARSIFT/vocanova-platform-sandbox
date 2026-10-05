import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, Surface } from "@/ui/surface";
import { getCompletionDayView } from "./completion-day-view";
import { KnowledgeOverview } from "../_components/knowledge-overview";
import { Achievements } from "../_components/achievements";

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
  let knowledgeResponse: Awaited<
    ReturnType<typeof client.getKnowledgeSummary>
  > | null;
  let lessonResponse: Awaited<ReturnType<typeof client.listLessons>> | null;
  let achievementsResponse: Awaited<
    ReturnType<typeof client.listAchievements>
  > | null;
  let sentencesResponse: Awaited<
    ReturnType<typeof client.listLearnerSentences>
  > | null;
  async function optional<T>(request: Promise<T>): Promise<T | null> {
    try {
      return await request;
    } catch (error) {
      if (error instanceof ApiResponseError && error.status === 401)
        requireAuthRedirect(error, "/progress");
      return null;
    }
  }
  try {
    [
      savedWordsResponse,
      progressResponse,
      knowledgeResponse,
      lessonResponse,
      achievementsResponse,
      sentencesResponse,
    ] = await Promise.all([
      client.listSavedWords({ limit: SAVED_VOCABULARY_DISPLAY_LIMIT }),
      client.getProgress(),
      optional(client.getKnowledgeSummary()),
      optional(client.listLessons()),
      optional(client.listAchievements()),
      optional(
        client.listLearnerSentences({ limit: 2 }, { cache: "no-store" }),
      ),
    ]);
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
    .map((day) => ({
      ...day,
      label: formatActivityDate(day.localDate),
      view: getCompletionDayView(day),
    }));

  return (
    <PageContainer className="max-w-[64rem]">
      <h1 className="mt-[var(--spacing-xs)] text-3xl font-bold tracking-tight text-neutral-900">
        Progress
      </h1>
      <p className="mt-2 text-neutral-600">
        Your words, lessons and writing. See what you have practised and keep
        going.
      </p>

      {knowledgeResponse ? (
        <KnowledgeOverview summary={knowledgeResponse.data} />
      ) : (
        <p role="status" className="mt-6 text-neutral-700">
          Your vocabulary map is unavailable right now. Your other progress is
          below.
        </p>
      )}
      {lessonResponse && (
        <Surface className="mt-4" aria-labelledby="lesson-progress-heading">
          <h2
            id="lesson-progress-heading"
            className="text-xl font-bold text-neutral-900"
          >
            Your learning path
          </h2>
          <p className="mt-2 text-neutral-700">
            {
              lessonResponse.data.items.filter(
                (item) => item.status === "completed",
              ).length
            }{" "}
            of {lessonResponse.data.items.length} lessons completed
          </p>
          {lessonResponse.data.items
            .filter((item) => item.status === "in_progress")
            .map((item) => (
              <Link
                key={item.key}
                href={`/learn/${encodeURIComponent(item.key)}`}
                className="mt-3 flex min-h-12 flex-wrap items-center justify-between gap-2 rounded-xl bg-primary-50 px-4 py-3 font-semibold text-primary-800 hover:bg-primary-100"
              >
                <span>
                  {item.title}
                  <span className="mt-1 block text-sm font-normal">
                    {item.completedSteps} of {item.stepCount} steps saved
                  </span>
                </span>
                <span className="text-sm">Continue lesson</span>
              </Link>
            ))}
          {lessonResponse.data.items.some(
            (item) => item.status === "completed",
          ) && (
            <details className="mt-3">
              <summary className="min-h-11 cursor-pointer content-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
                Lesson history
              </summary>
              <ul className="mt-4 divide-y divide-neutral-200">
                {lessonResponse.data.items
                  .filter((item) => item.status === "completed")
                  .map((item) => (
                    <li key={item.key} className="py-3">
                      <Link
                        href={`/learn/${encodeURIComponent(item.key)}`}
                        className="inline-flex min-h-11 items-center font-semibold text-primary-700"
                      >
                        {item.title}
                      </Link>
                      <p className="text-sm text-neutral-600">Completed</p>
                    </li>
                  ))}
              </ul>
            </details>
          )}
          <Link
            href="/discover"
            className="mt-3 inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            Continue your Journey
          </Link>
        </Surface>
      )}

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
              Your sentences and feedback, saved together.
            </p>
          </div>
          <Link
            href="/progress/sentences"
            className="inline-flex min-h-11 items-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-white hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            View sentence history
          </Link>
        </div>
        {sentencesResponse ? (
          sentencesResponse.data.items.length > 0 ? (
            <ul aria-label="Recent sentences" className="mt-5 space-y-3">
              {sentencesResponse.data.items.map((sentence) => (
                <li key={sentence.id} className="rounded-xl bg-white p-4">
                  <p className="break-words text-base text-neutral-900 [overflow-wrap:anywhere]">
                    {sentence.originalSentence}
                  </p>
                  <p className="mt-2 text-sm text-neutral-600">
                    {sentence.processingStatus === "pending"
                      ? "Feedback is being prepared"
                      : sentence.processingStatus === "completed"
                        ? "Feedback saved in your sentence history"
                        : "Sentence saved — feedback unavailable"}
                  </p>
                </li>
              ))}
            </ul>
          ) : (
            <Link
              href="/writing"
              className="mt-4 inline-flex min-h-11 items-center font-semibold text-primary-700"
            >
              Write your first sentence
            </Link>
          )
        ) : (
          <p role="status" className="mt-4 text-neutral-700">
            Your recent writing could not load. Open sentence history to try
            again.
          </p>
        )}
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

      {achievementsResponse ? (
        <Achievements achievements={achievementsResponse.data} />
      ) : (
        <p role="status" className="mt-6 text-neutral-700">
          Your milestones could not load. Your other progress is still
          available.
        </p>
      )}

      <div className="mt-[var(--spacing-md)] lg:grid lg:grid-cols-2 lg:items-start lg:gap-[var(--spacing-md)]">
        <Surface aria-labelledby="completion-history-heading">
          <h2
            id="completion-history-heading"
            className="text-lg font-semibold text-neutral-900"
          >
            Recent activity
          </h2>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            Your recorded mission days.
          </p>
          {historyWithLabels.length > 0 ? (
            <ul className="mt-[var(--spacing-md)] space-y-[var(--spacing-xs)]">
              {historyWithLabels.map((day) => (
                <li
                  key={day.localDate}
                  className={`flex min-w-0 items-center justify-between gap-[var(--spacing-sm)] rounded-md border-l-4 px-[var(--spacing-md)] py-[var(--spacing-sm)] ${day.view.className}`}
                >
                  <time
                    dateTime={day.localDate}
                    className="shrink-0 text-sm font-semibold"
                  >
                    {day.label}
                  </time>
                  <p className="min-w-0 text-right text-sm">{day.view.label}</p>
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
