import Link from "next/link";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { MeaningPicture } from "@/ui/meaning-picture";
import { StartDailyLessonButton } from "../reviews/_components/daily-lesson-start";

export default async function HomePage() {
  const client = await createServerApiClient();
  let savedWordsResponse: Awaited<ReturnType<typeof client.listSavedWords>>;
  let dailyLessonResponse: Awaited<
    ReturnType<typeof client.getCurrentDailyLesson>
  >;
  let dailyMissionResponse: Awaited<ReturnType<typeof client.getDailyMission>>;
  try {
    [savedWordsResponse, dailyLessonResponse, dailyMissionResponse] =
      await Promise.all([
        client.listSavedWords({ limit: 1 }),
        client.getCurrentDailyLesson(),
        client.getDailyMission(),
      ]);
  } catch (error) {
    requireAuthRedirect(error, "/home");
  }
  const firstWord = savedWordsResponse.data.items[0];
  const dueReviewWords = dailyLessonResponse.data.dueCount;
  const currentLesson = dailyLessonResponse.data.session;
  const mission = dailyMissionResponse.data;
  const { reviewTarget, reviewsCompleted, streak } = mission;
  const complete = mission.status === "completed";
  const action = currentLesson
    ? {
        href: "/review",
        label: "Continue lesson",
        title: "Pick up where you left off.",
      }
    : dueReviewWords > 0
      ? {
          href: "/review",
          label: "Start today’s lesson",
          title: "Remember today’s words.",
        }
      : {
          href: "/discovery",
          label: "Discover words",
          title: "Meet your next useful word.",
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
            {!currentLesson && dueReviewWords > 0 ? (
              <StartDailyLessonButton
                className="learning-primary"
                label={action.label}
              />
            ) : (
              <Link href={action.href} className="learning-primary">
                {action.label}
              </Link>
            )}
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
          <Link
            href="/knowledge-map"
            className="mt-6 flex min-h-16 items-center justify-between gap-4 rounded-xl border border-neutral-200 bg-white p-4 text-lg font-semibold text-neutral-900 hover:border-primary-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700"
          >
            <span>Your word map</span>
            <span className="text-sm font-medium text-primary-700">
              Explore 100-word groups
            </span>
          </Link>
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
                <Link href="/discovery" className="learning-text-link">
                  Discover another word
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
        </section>
      </div>
    </PageContainer>
  );
}
