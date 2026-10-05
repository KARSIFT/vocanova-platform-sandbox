import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, Surface } from "@/ui/surface";
import { SentenceFeedback } from "../_components/sentence-feedback";
import { RecommendedLesson } from "../_components/lesson-recommendation";

export default async function HomePage() {
  const client = await createServerApiClient();
  let savedWordsResponse: Awaited<ReturnType<typeof client.listSavedWords>>;
  let dueResponse: Awaited<ReturnType<typeof client.listDueWords>>;
  let dailyMissionResponse: Awaited<ReturnType<typeof client.getDailyMission>>;
  let currentUserResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    [
      savedWordsResponse,
      dueResponse,
      dailyMissionResponse,
      currentUserResponse,
    ] = await Promise.all([
      client.listSavedWords({ limit: 3 }),
      client.listDueWords({ limit: 1 }),
      client.getDailyMission(),
      client.getCurrentUser(),
    ]);
  } catch (error) {
    requireAuthRedirect(error, "/home");
  }

  const recommendation = await client
    .getLessonRecommendation()
    .catch((error: unknown) => {
      if (error instanceof ApiResponseError && error.status === 401)
        requireAuthRedirect(error, "/home");
      return null;
    });

  const { items: savedWords } = savedWordsResponse.data;
  const dueReviewWords = dueResponse.data.totalCount;

  const {
    reviewTarget: missionTargetWords,
    reviewsCompleted: reviewedWordsToday,
    newWordTarget,
    newWordsCompleted,
    sentencePracticeTarget,
    sentencePracticesCompleted,
    streak,
  } = dailyMissionResponse.data;
  const currentStreakDays = streak.currentStreakCount;

  const missionProgressPercent =
    missionTargetWords > 0
      ? Math.min(
          100,
          Math.round((reviewedWordsToday / missionTargetWords) * 100),
        )
      : 0;
  const missionComplete = dailyMissionResponse.data.status === "completed";
  const hasDueReviews = dueReviewWords > 0;
  const hasRemainingReviewTarget = reviewedWordsToday < missionTargetWords;
  const hasRemainingNewWordTarget =
    typeof newWordTarget === "number" &&
    typeof newWordsCompleted === "number" &&
    newWordsCompleted < newWordTarget;
  const remainingSentencePractices =
    typeof sentencePracticeTarget === "number" &&
    typeof sentencePracticesCompleted === "number"
      ? Math.max(0, sentencePracticeTarget - sentencePracticesCompleted)
      : null;
  const hasRemainingSentenceTarget =
    remainingSentencePractices !== null && remainingSentencePractices > 0;
  const primaryAction = missionComplete
    ? {
        href: "/discover",
        label: "Explore a new situation",
        detail:
          "Your daily review goal is complete. Keep the momentum with useful vocabulary.",
      }
    : hasDueReviews && hasRemainingReviewTarget
      ? {
          href: "/review",
          label: "Start review",
          detail: "A short session with the words due now.",
        }
      : savedWords.length === 0
        ? {
            href: "/discover",
            label: "Start your Journey",
            detail:
              "Save a word from a real-life situation to begin your review habit.",
          }
        : hasRemainingNewWordTarget || hasRemainingReviewTarget
          ? {
              href: "/discover",
              label: "Explore a new situation",
              detail: hasRemainingReviewTarget
                ? "Nothing is due right now. Add a useful word to keep building today’s practice."
                : "Choose a useful word from a real-life situation to continue today’s mission.",
            }
          : hasRemainingSentenceTarget && savedWords.length > 0
            ? {
                href: "#sentence-practice",
                label: "Practice a sentence",
                detail: `${remainingSentencePractices} sentence ${remainingSentencePractices === 1 ? "practice is" : "practices are"} left in today’s mission.`,
              }
            : {
                href: "/progress",
                label: "View your progress",
                detail:
                  "Your next mission step will be ready when there’s something new to practice.",
              };

  return (
    <PageContainer className="max-w-[72rem]">
      <div className="mb-5 flex items-center justify-between gap-4">
        <h1 className="text-3xl font-bold tracking-tight text-neutral-900">
          Today
        </h1>
        <p className="rounded-xl bg-secondary-50 px-3 py-2 text-sm font-semibold text-secondary-900">
          {currentStreakDays}-day streak
        </p>
      </div>
      <div className="lg:grid lg:items-start lg:grid-cols-[minmax(0,1.45fr)_minmax(18rem,0.75fr)] lg:gap-[var(--spacing-lg)]">
        <section
          aria-labelledby="todays-mission-heading"
          className="rounded-[0.9rem] bg-primary-800 p-[var(--spacing-md)] text-white shadow-[0_10px_24px_rgb(30_58_138_/_0.18)] sm:p-[var(--spacing-lg)]"
        >
          <div className="flex items-start justify-between gap-[var(--spacing-md)]">
            <div>
              <h2
                id="todays-mission-heading"
                className="text-xl font-bold tracking-tight"
              >
                {missionComplete ? "Mission complete" : "Today’s practice"}
              </h2>
            </div>
            <span
              className={`rounded-full px-[var(--spacing-sm)] py-1 text-xs font-bold ${missionComplete ? "bg-white text-primary-800" : "bg-primary-700 text-primary-100"}`}
            >
              {missionComplete ? "Complete" : "In progress"}
            </span>
          </div>
          <p className="mt-[var(--spacing-sm)] text-base text-primary-100">
            Reviews today: {reviewedWordsToday} of {missionTargetWords}
          </p>
          <p className="mt-[var(--spacing-xs)] text-sm text-primary-100">
            Currently due: {dueReviewWords}{" "}
            {dueReviewWords === 1 ? "word" : "words"}.
          </p>
          <div
            role="progressbar"
            aria-label="Reviews today"
            aria-valuemin={0}
            aria-valuemax={missionTargetWords}
            aria-valuenow={Math.min(reviewedWordsToday, missionTargetWords)}
            className="mt-[var(--spacing-sm)] h-2 w-full overflow-hidden rounded-full bg-primary-900/50"
          >
            <div
              className="h-full rounded-full bg-primary-200"
              style={{ width: `${missionProgressPercent}%` }}
            />
          </div>
          <Link
            href={primaryAction.href}
            className="mt-[var(--spacing-sm)] inline-flex min-h-11 items-center justify-center rounded-xl bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-800 shadow-sm transition-colors hover:bg-primary-50"
          >
            {primaryAction.label}
          </Link>
          <p className="mt-[var(--spacing-xs)] text-sm text-primary-100">
            {primaryAction.detail}
          </p>
        </section>
        <RecommendedLesson
          data={recommendation?.data ?? null}
          className="mt-4 lg:col-start-1"
        />
        <Surface
          aria-labelledby="saved-words-heading"
          className="mt-[var(--spacing-md)] lg:col-start-2 lg:row-start-1 lg:row-span-2 lg:mt-0"
        >
          <div className="flex items-baseline justify-between gap-[var(--spacing-md)]">
            <h2
              id="saved-words-heading"
              className="text-lg font-bold tracking-tight text-neutral-900"
            >
              Your vocabulary
            </h2>
            {savedWords.length > 0 ? (
              <Link
                href="/words"
                className="inline-flex min-h-11 items-center text-sm font-semibold text-primary-700 hover:text-primary-800"
              >
                See all
              </Link>
            ) : null}
          </div>
          {savedWords.length > 0 ? (
            <>
              <ul className="mt-[var(--spacing-md)] divide-y divide-neutral-100">
                {savedWords.map((savedWord) => (
                  <li
                    key={savedWord.userWordId}
                    className="py-[var(--spacing-sm)] first:pt-0 last:pb-0"
                  >
                    <p className="font-semibold text-neutral-900">
                      <Link
                        href={`/words/${savedWord.userWordId}`}
                        className="inline-flex min-h-11 items-center rounded-sm hover:text-primary-700"
                      >
                        {savedWord.wordText}
                      </Link>
                      <span className="ml-[var(--spacing-xs)] text-sm font-normal text-neutral-500">
                        {savedWord.partOfSpeech}
                      </span>
                    </p>
                    <p className="mt-0.5 text-sm text-neutral-600">
                      {savedWord.shortDefinition}
                    </p>
                  </li>
                ))}
              </ul>
              <details
                id="sentence-practice"
                className="mt-[var(--spacing-md)] rounded-xl border border-secondary-100 bg-secondary-50 px-[var(--spacing-md)]"
              >
                <summary className="min-h-11 cursor-pointer content-center text-sm font-semibold text-secondary-900 marker:text-secondary-700">
                  Practice “{savedWords[0]!.wordText}” in a sentence
                </summary>
                <SentenceFeedback
                  targetWord={savedWords[0]!.wordText}
                  attemptId={savedWords[0]!.userWordId}
                  source="word_detail"
                  userId={currentUserResponse.data.id}
                  shortDefinition={savedWords[0]!.shortDefinition}
                />
              </details>
            </>
          ) : (
            <div className="mt-[var(--spacing-md)] rounded-xl bg-neutral-50 p-[var(--spacing-md)]">
              <p className="font-semibold text-neutral-900">
                Start with a situation you know.
              </p>
              <p className="mt-1 text-sm text-neutral-600">
                Choose Airport, Work, or everyday conversation, then save the
                words you want to use.
              </p>
            </div>
          )}
        </Surface>
      </div>
      <div className="mt-5 flex flex-wrap gap-x-6 gap-y-1">
        <Link
          href="/vocabulary"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
        >
          Find a word
        </Link>
        <Link
          href="/plan"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
        >
          Your learning plan
        </Link>
      </div>
    </PageContainer>
  );
}
