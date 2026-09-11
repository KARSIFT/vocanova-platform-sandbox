import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";
import { SentenceFeedback } from "../_components/sentence-feedback";

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
          detail: `${dueReviewWords} ${dueReviewWords === 1 ? "word is" : "words are"} ready when you are.`,
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
      <div className="mb-[var(--spacing-md)] flex items-center justify-between gap-[var(--spacing-md)]">
        <div>
          <Eyebrow>Today’s learning space</Eyebrow>
        </div>
        <div className="shrink-0 rounded-xl bg-secondary-50 px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-right">
          <p className="text-xs font-semibold text-secondary-800">STREAK</p>
          <p className="text-lg font-bold leading-tight text-secondary-900">
            {currentStreakDays} day{currentStreakDays === 1 ? "" : "s"}
          </p>
        </div>
      </div>
      <div className="lg:grid lg:grid-cols-[minmax(0,1.45fr)_minmax(18rem,0.75fr)] lg:gap-[var(--spacing-lg)]">
        <section
          aria-labelledby="todays-mission-heading"
          className="rounded-[0.9rem] bg-primary-800 p-[var(--spacing-lg)] text-white shadow-[0_10px_24px_rgb(30_58_138_/_0.18)] sm:p-[var(--spacing-xl)]"
        >
          <div className="flex items-start justify-between gap-[var(--spacing-md)]">
            <div>
              <h1
                id="todays-mission-heading"
                className="text-sm font-semibold tracking-[0.08em] text-primary-200 uppercase"
              >
                Today&apos;s Mission
              </h1>
              <h2 className="mt-[var(--spacing-xs)] text-xl font-bold tracking-tight sm:text-2xl">
                {missionComplete
                  ? "Mission complete"
                  : "Build your review habit"}
              </h2>
            </div>
            <span
              className={`rounded-full px-[var(--spacing-sm)] py-1 text-xs font-bold ${missionComplete ? "bg-white text-primary-800" : "bg-primary-700 text-primary-100"}`}
            >
              {missionComplete ? "Complete" : "In progress"}
            </span>
          </div>
          <p className="mt-[var(--spacing-md)] text-base text-primary-100">
            {missionComplete
              ? `You reviewed ${reviewedWordsToday} of ${missionTargetWords} words today.`
              : `${reviewedWordsToday} of ${missionTargetWords} reviews complete`}
          </p>
          <div
            role="progressbar"
            aria-label="Today’s mission progress"
            aria-valuemin={0}
            aria-valuemax={missionTargetWords}
            aria-valuenow={Math.min(reviewedWordsToday, missionTargetWords)}
            className="mt-[var(--spacing-md)] h-2 w-full overflow-hidden rounded-full bg-primary-900/50"
          >
            <div
              className="h-full rounded-full bg-primary-200 transition-[width] duration-[var(--duration-slow)]"
              style={{ width: `${missionProgressPercent}%` }}
            />
          </div>
          <Link
            href={primaryAction.href}
            className="mt-[var(--spacing-md)] inline-flex min-h-11 items-center justify-center rounded-xl bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-primary-800 shadow-sm transition-colors hover:bg-primary-50"
          >
            {primaryAction.label}
            <span aria-hidden="true" className="ml-[var(--spacing-sm)]">
              →
            </span>
          </Link>
          <p className="mt-[var(--spacing-sm)] text-sm text-primary-100">
            {primaryAction.detail}
          </p>
        </section>

        <Surface
          aria-labelledby="saved-words-heading"
          className="mt-[var(--spacing-md)] lg:mt-0"
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
    </PageContainer>
  );
}
