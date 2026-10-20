import Link from "next/link";
import { createHash } from "node:crypto";
import { Suspense } from "react";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { MeaningPicture } from "@/ui/meaning-picture";
import { StartDailyLessonButton } from "../reviews/_components/daily-lesson-start";
import { PracticeJourney } from "../practice/_components/practice-journey";
import {
  ApiResponseError,
  type PracticeSessionsResponse,
  type SavedMeaning,
  type MeaningImagesResponse,
} from "@vocanova/api-client";

async function HomePractice({
  read,
}: {
  read: Promise<PromiseSettledResult<{ data: PracticeSessionsResponse }>>;
}) {
  const result = await read;
  if (
    result.status === "rejected" &&
    result.reason instanceof ApiResponseError &&
    result.reason.status === 401
  )
    requireAuthRedirect(result.reason, "/home");
  return (
    <PracticeJourney
      journey={
        result.status === "fulfilled"
          ? (result.value.data.journey ?? null)
          : null
      }
      compact
    />
  );
}

// Match the API's meaningidentity.Revision encoding, including Go JSON's HTML
// and line-separator escapes. This binds artwork to the saved-word text already
// displayed on Home without adding a blocking canonical-detail request.
function spotlightMeaningRevision(word: SavedMeaning): string | undefined {
  const identity = [
    word.meaningId,
    word.wordText,
    word.partOfSpeech,
    word.shortDefinition,
  ];
  if (identity.some((value) => typeof value !== "string" || !value))
    return undefined;
  const encoded = JSON.stringify(identity).replace(
    /[<>&\u2028\u2029]/g,
    (character) =>
      `\\u${character.charCodeAt(0).toString(16).padStart(4, "0")}`,
  );
  return createHash("sha256").update(encoded).digest("hex");
}

async function HomeMeaningPicture({
  read,
  meaningId,
  meaningRevision,
}: {
  read: Promise<PromiseSettledResult<{ data: MeaningImagesResponse }>>;
  meaningId: string;
  meaningRevision: string;
}) {
  const result = await read;
  const image =
    result.status === "fulfilled"
      ? result.value.data.items.find((item) => item.meaningId === meaningId)
      : undefined;
  return (
    <MeaningPicture
      meaningId={meaningId}
      meaningRevision={meaningRevision}
      image={image}
    />
  );
}

export default async function HomePage() {
  const client = await createServerApiClient();
  // Start the secondary read concurrently, capturing rejection immediately.
  // Its latency or recoverable failure must not block the primary review card.
  const practiceRead = Promise.allSettled([client.listPracticeSessions()]).then(
    ([result]) => result,
  );
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
  const spotlightRevision = firstWord
    ? spotlightMeaningRevision(firstWord)
    : undefined;
  // Optional media never delays Review or the primary Home controls. The lookup
  // is bounded to one meaning, catches failure immediately and has a deadline.
  const spotlightRead =
    firstWord && spotlightRevision && !firstWord.image
      ? Promise.allSettled([
          client.getMeaningImages([firstWord.meaningId], {
            cache: "no-store",
            signal: AbortSignal.timeout(3000),
          }),
        ]).then(([result]) => result)
      : undefined;
  const dueReviewWords = dailyLessonResponse.data.dueCount;
  const currentLesson = dailyLessonResponse.data.session;
  const mission = dailyMissionResponse.data;
  const { reviewsCompleted, streak } = mission;
  const action = currentLesson
    ? {
        href: "/review",
        label: "Continue review",
        title: "Review words",
      }
    : dueReviewWords > 0
      ? {
          href: "/review",
          label: "Start review",
          title: "Review words",
        }
      : {
          href: "/discovery",
          label: "Discover words",
          title: "Review complete",
        };

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
          <section aria-label="Review words" className="daily-focus">
            <h2 id="todays-mission-heading">{action.title}</h2>
            <p className="mt-2 text-neutral-600">
              {currentLesson
                ? "Pick up where you left off."
                : dueReviewWords > 0
                  ? `${dueReviewWords} ${dueReviewWords === 1 ? "word is" : "words are"} ready to remember.`
                  : "No words are due. Discover something new whenever you like."}
            </p>
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
                  Reviewed today:{" "}
                  <strong className="text-neutral-900">
                    {reviewsCompleted}
                  </strong>
                </span>
              </div>
            </div>
          </section>
          <div className="mt-6">
            <Suspense
              fallback={
                <section
                  aria-label="Your learning path"
                  aria-busy="true"
                  className="min-h-[25rem] rounded-2xl border border-neutral-200 bg-white p-5 sm:min-h-[23rem]"
                >
                  <h2 className="text-2xl font-bold text-neutral-900">
                    Today’s practice
                  </h2>
                  <p role="status" className="mt-2 text-neutral-600">
                    Loading practice…
                  </p>
                </section>
              }
            >
              <HomePractice read={practiceRead} />
            </Suspense>
          </div>
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
              {spotlightRead && spotlightRevision ? (
                <Suspense
                  fallback={
                    <MeaningPicture
                      meaningId={firstWord.meaningId}
                      meaningRevision={spotlightRevision}
                    />
                  }
                >
                  <HomeMeaningPicture
                    read={spotlightRead}
                    meaningId={firstWord.meaningId}
                    meaningRevision={spotlightRevision}
                  />
                </Suspense>
              ) : (
                <MeaningPicture
                  meaningId={firstWord.meaningId}
                  image={firstWord.image}
                  meaningRevision={spotlightRevision}
                />
              )}
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
