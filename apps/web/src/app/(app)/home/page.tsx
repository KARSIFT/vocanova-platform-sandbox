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
          <span>
            <strong>{streak.currentStreakCount}</strong>-day streak
          </span>
        </Link>
      </header>
      <div className="study-desk">
        <section aria-label="Review words" className="daily-focus">
          <p className="mission-kicker">
            {currentLesson
              ? "Review in progress"
              : dueReviewWords > 0
                ? "Today’s review"
                : "All caught up"}
          </p>
          <h2 id="todays-mission-heading">{action.title}</h2>
          <p className="mission-copy">
            {currentLesson
              ? "Pick up where you left off."
              : dueReviewWords > 0
                ? `${dueReviewWords} ${dueReviewWords === 1 ? "word is" : "words are"} ready to remember.`
                : "No words are due. Discover something new whenever you like."}
          </p>
          <div className="mission-actions">
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
            <p className="daily-meter">
              Reviewed today: <strong>{reviewsCompleted}</strong>
            </p>
          </div>
        </section>
        <section
          aria-labelledby="saved-words-heading"
          className="word-spotlight"
        >
          <div className="word-spotlight-head">
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
            <div className="word-spotlight-body">
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
              <div id="sentence-practice" className="spotlight-links">
                <Link href="/discovery" className="vn-chip-link">
                  Discover another word
                </Link>
                <Link href="/lists" className="vn-chip-link">
                  Personal lists
                </Link>
              </div>
            </div>
          ) : (
            <div className="spotlight-empty">
              <div className="spotlight-empty-art" aria-hidden="true">
                Aa
              </div>
              <p className="text-2xl font-bold tracking-tight text-neutral-900">
                A word worth keeping.
              </p>
              <p className="text-neutral-600">
                Save a word to start your collection.
              </p>
              <Link href="/vocabulary" className="vn-chip-link mt-2">
                Find a word
              </Link>
            </div>
          )}
        </section>
        <div className="home-practice">
          <Suspense
            fallback={
              <section
                aria-label="Your learning path"
                aria-busy="true"
                className="min-h-[25rem] rounded-3xl border border-neutral-200 bg-white p-5 shadow-[var(--vn-shadow-card)] sm:min-h-[23rem]"
              >
                <h2 className="text-xl font-bold text-neutral-900">
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
        <div className="home-explore">
          <h2 id="explore-heading" className="home-section-title">
            More ways to learn
          </h2>
          <nav aria-label="Explore your English">
            <ul className="study-shortcuts">
              {EXPLORE_LINKS.map(({ href, label, caption, symbol, tone }) => (
                <li key={href}>
                  <span
                    className={`shortcut-symbol${tone === "lavender" ? " shortcut-symbol-lavender" : ""}`}
                    aria-hidden="true"
                  >
                    {symbol}
                  </span>
                  <span>
                    <Link
                      href={href}
                      aria-describedby={`explore-${href.slice(1)}`}
                    >
                      {label}
                    </Link>
                    <span
                      id={`explore-${href.slice(1)}`}
                      className="shortcut-caption"
                    >
                      {caption}
                    </span>
                  </span>
                </li>
              ))}
            </ul>
          </nav>
        </div>
      </div>
    </PageContainer>
  );
}

function ExploreIcon({ d }: { d: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className="fill-none stroke-current stroke-[1.8]"
      aria-hidden="true"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d={d} />
    </svg>
  );
}

const EXPLORE_LINKS = [
  {
    href: "/practice",
    label: "Practice",
    caption: "Listening, recall and mistakes",
    symbol: <ExploreIcon d="M4.75 12a7.25 7.25 0 1 0 2.2-5.2M4.75 4.75v4h4" />,
    tone: "blue",
  },
  {
    href: "/knowledge-map",
    label: "Word map",
    caption: "Browse 100-word groups",
    symbol: (
      <ExploreIcon d="M4.75 4.75h5.5v5.5h-5.5zM13.75 4.75h5.5v5.5h-5.5zM4.75 13.75h5.5v5.5h-5.5zM13.75 13.75h5.5v5.5h-5.5z" />
    ),
    tone: "blue",
  },
  {
    href: "/stories",
    label: "Short stories",
    caption: "Words inside a conversation",
    symbol: "Aa",
    tone: "lavender",
  },
  {
    href: "/plan",
    label: "Your plan",
    caption: "Pace and direction",
    symbol: <ExploreIcon d="M6 5h12v15H6zM9 3v4m6-4v4M9 11h6m-6 4h4" />,
    tone: "blue",
  },
  {
    href: "/courses",
    label: "Situations",
    caption: "Guided lessons by place",
    symbol: (
      <ExploreIcon d="M12 20.25s-6.25-5.4-6.25-10.5a6.25 6.25 0 1 1 12.5 0c0 5.1-6.25 10.5-6.25 10.5Zm0-8.25a2.25 2.25 0 1 0 0-4.5 2.25 2.25 0 0 0 0 4.5Z" />
    ),
    tone: "lavender",
  },
] as const;
