import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { MeaningPicture } from "@/ui/meaning-picture";
import { RecommendedLesson } from "../_components/lesson-recommendation";

export default async function HomePage() {
  const client = await createServerApiClient();
  let savedWordsResponse: Awaited<ReturnType<typeof client.listSavedWords>>;
  let dueResponse: Awaited<ReturnType<typeof client.listDueWords>>;
  let dailyMissionResponse: Awaited<ReturnType<typeof client.getDailyMission>>;
  let recommendation: Awaited<
    ReturnType<typeof client.getLessonRecommendation>
  > | null;
  try {
    [savedWordsResponse, dueResponse, dailyMissionResponse, recommendation] =
      await Promise.all([
        client.listSavedWords({ limit: 1 }),
        client.listDueWords({ limit: 1 }),
        client.getDailyMission(),
        client.getLessonRecommendation().catch((error: unknown) => {
          if (error instanceof ApiResponseError && error.status === 401)
            requireAuthRedirect(error, "/home");
          return null;
        }),
      ]);
  } catch (error) {
    requireAuthRedirect(error, "/home");
  }
  const firstWord = savedWordsResponse.data.items[0];
  const dueReviewWords = dueResponse.data.totalCount;
  const mission = dailyMissionResponse.data;
  const {
    reviewTarget,
    reviewsCompleted,
    newWordTarget,
    newWordsCompleted,
    sentencePracticeTarget,
    sentencePracticesCompleted,
    streak,
  } = mission;
  const complete = mission.status === "completed";
  const remainingReviews = reviewsCompleted < reviewTarget;
  const remainingNewWords =
    typeof newWordTarget === "number" &&
    typeof newWordsCompleted === "number" &&
    newWordsCompleted < newWordTarget;
  const remainingSentences =
    typeof sentencePracticeTarget === "number" &&
    typeof sentencePracticesCompleted === "number" &&
    sentencePracticesCompleted < sentencePracticeTarget;
  const writingHref = firstWord
    ? `/words/${encodeURIComponent(firstWord.userWordId)}#sentence-practice`
    : "/writing";
  const action = complete
    ? {
        href: "/discover",
        label: "Explore a new situation",
        title: "A little progress. Every day.",
      }
    : dueReviewWords > 0 && remainingReviews
      ? {
          href: "/review",
          label: "Start review",
          title: "Keep your words fresh.",
        }
      : !firstWord
        ? {
            href: "/discover",
            label: "Start your Journey",
            title: "Find your first useful word.",
          }
        : remainingNewWords || remainingReviews
          ? {
              href: "/discover",
              label: "Explore a new situation",
              title: "What will you learn today?",
            }
          : remainingSentences
            ? {
                href: writingHref,
                label: "Practice a sentence",
                title: "Make a word your own.",
              }
            : {
                href: "/progress",
                label: "View your progress",
                title: "See how far you’ve come.",
              };
  const progress =
    reviewTarget > 0
      ? Math.min(100, Math.round((reviewsCompleted / reviewTarget) * 100))
      : 0;

  return (
    <PageContainer className="learning-overview max-w-[72rem]">
      <header className="overview-title">
        <h1>Today</h1>
        <Link
          href="/progress"
          className="streak-label"
          aria-label={`${streak.currentStreakCount}-day streak. View progress`}
        >
          <svg
            aria-hidden="true"
            viewBox="0 0 24 24"
            className="size-5 fill-none stroke-current stroke-[1.7]"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M13 3c1 5-4 5-3 9 1-1 2-2 2-3 4 2 6 5 5 8a6 6 0 0 1-11-1c-1-5 4-7 7-13Z"
            />
          </svg>
          {streak.currentStreakCount}-day streak
        </Link>
      </header>
      <div className="study-desk">
        <div className="study-next">
          <section
            aria-label={complete ? "Mission complete" : "Today’s practice"}
            className="daily-focus"
          >
            <div className="daily-focus-top">
              <span className="text-sm font-semibold text-primary-700">
                {complete ? "Mission complete" : "Today’s practice"}
              </span>
              {complete && (
                <span aria-hidden="true" className="text-primary-700">
                  ✓
                </span>
              )}
            </div>
            <h2 id="todays-mission-heading">{action.title}</h2>
            <Link href={action.href} className="learning-primary">
              {action.label}
            </Link>
            <div className="daily-meter">
              <div className="flex flex-wrap justify-between gap-x-4 gap-y-1 text-sm text-neutral-600">
                <span>
                  Reviews today:{" "}
                  <strong className="text-neutral-900">
                    {reviewsCompleted} of {reviewTarget}
                  </strong>
                </span>
                <span>
                  Currently due: {dueReviewWords}{" "}
                  {dueReviewWords === 1 ? "word" : "words"}.
                </span>
              </div>
              <div
                role="progressbar"
                aria-label="Reviews today"
                aria-valuemin={0}
                aria-valuemax={reviewTarget}
                aria-valuenow={Math.min(reviewsCompleted, reviewTarget)}
                className="mt-3 h-1.5 overflow-hidden rounded-full bg-neutral-200"
              >
                <div
                  className="h-full rounded-full bg-primary-700"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
          </section>
          <RecommendedLesson data={recommendation?.data ?? null} compact />
          <nav aria-label="Explore your English" className="study-shortcuts">
            <Link href="/practice">
              <span className="shortcut-symbol" aria-hidden="true">
                ↺
              </span>
              <span>Practice</span>
            </Link>
            <Link href="/stories">
              <span className="shortcut-symbol" aria-hidden="true">
                Aa
              </span>
              <span>Short stories</span>
            </Link>
            <Link href="/plan">
              <svg
                className="shortcut-symbol fill-none stroke-current stroke-[1.6]"
                aria-hidden="true"
                viewBox="0 0 24 24"
              >
                <path d="M6 5h12v15H6zM9 3v4m6-4v4M9 11h6m-6 4h4" />
              </svg>
              <span>Your plan</span>
            </Link>
          </nav>
        </div>
        <section
          aria-labelledby="saved-words-heading"
          className="word-spotlight"
        >
          <div className="flex items-center justify-between gap-3">
            <h2
              id="saved-words-heading"
              className="text-sm font-semibold text-neutral-600"
            >
              Your vocabulary
            </h2>
            <Link href="/words" className="learning-text-link text-sm">
              See all
            </Link>
          </div>
          {firstWord ? (
            <>
              <MeaningPicture meaningId={firstWord.meaningId} priority />
              <Link
                href={`/words/${encodeURIComponent(firstWord.userWordId)}`}
                className="spotlight-word"
              >
                {firstWord.wordText}
              </Link>
              <p className="mt-1 text-neutral-600">
                {firstWord.shortDefinition}
              </p>
              <div
                id="sentence-practice"
                className="mt-4 flex flex-wrap gap-x-5 border-t border-neutral-200 pt-2"
              >
                <Link href={writingHref} className="learning-text-link">
                  Write with {firstWord.wordText}
                </Link>
                <Link href="/lists" className="learning-text-link">
                  Personal lists
                </Link>
              </div>
            </>
          ) : (
            <div className="py-8">
              <p className="text-2xl font-semibold tracking-tight text-neutral-900">
                A word worth keeping.
              </p>
              <p className="mt-2 text-neutral-600">
                Save a word to start your collection.
              </p>
              <Link href="/vocabulary" className="learning-text-link mt-3">
                Find a word
              </Link>
            </div>
          )}
          <Link href="/writing" className="learning-text-link text-sm">
            Choose a writing topic
          </Link>
        </section>
      </div>
    </PageContainer>
  );
}
