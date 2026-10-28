import Link from "@/ui/link";
import { createHash } from "node:crypto";
import { Suspense } from "react";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { MeaningPicture } from "@/ui/meaning-picture";
import { StartDailyLessonButton } from "../reviews/_components/daily-lesson-start";
import { PracticeJourney } from "../practice/_components/practice-journey";
import {
  ApiResponseError,
  type DailyMission,
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

const MISSION_STATUS: Record<DailyMission["status"], string> = {
  open: "In progress",
  completed: "Completed",
  protected: "Protected by a grace day",
  missed: "Missed",
};

/** Only the parts today's mission actually contains, with its own numbers. */
function missionGoals(mission: DailyMission) {
  const goals = [
    {
      label: "Reviews",
      done: mission.reviewsCompleted,
      target: mission.reviewTarget,
      optional: false,
    },
  ];
  if (mission.newWordTarget)
    goals.push({
      label: "New words",
      done: mission.newWordsCompleted ?? 0,
      target: mission.newWordTarget,
      optional: false,
    });
  if (mission.sentencePracticeTarget)
    goals.push({
      label: "Sentence practice",
      done: mission.sentencePracticesCompleted ?? 0,
      target: mission.sentencePracticeTarget,
      optional: true,
    });
  return goals;
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
  // Saved words arrive newest first, so this is the learner's latest word.
  const latestWord = savedWordsResponse.data.items[0];
  const latestRevision = latestWord
    ? spotlightMeaningRevision(latestWord)
    : undefined;
  // Optional media never delays Review or the primary Home controls. The lookup
  // is bounded to one meaning, catches failure immediately and has a deadline.
  const pictureRead =
    latestWord && latestRevision && !latestWord.image
      ? Promise.allSettled([
          client.getMeaningImages([latestWord.meaningId], {
            cache: "no-store",
            signal: AbortSignal.timeout(3000),
          }),
        ]).then(([result]) => result)
      : undefined;
  const dueReviewWords = dailyLessonResponse.data.dueCount;
  const currentLesson = dailyLessonResponse.data.session;
  const mission = dailyMissionResponse.data;
  const { reviewsCompleted, streak } = mission;
  const discoveryIsNext = !currentLesson && dueReviewWords === 0;
  const goals = missionGoals(mission);
  const wordHref = latestWord
    ? `/words/${encodeURIComponent(latestWord.userWordId)}`
    : undefined;

  return (
    <PageContainer className="today max-w-[72rem]">
      <header className="today-head">
        <h1>Today</h1>
        {streak.currentStreakCount > 0 ? (
          <Link
            href="/progress"
            className="streak-note"
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
        ) : null}
      </header>

      <section aria-label="Review words" className="next-step">
        <p className="next-step-label">
          {currentLesson
            ? "Review in progress"
            : dueReviewWords > 0
              ? "Your next step"
              : "All caught up"}
        </p>
        <h2 id="todays-mission-heading">
          {discoveryIsNext ? "Review complete" : "Review words"}
        </h2>
        <p className="next-step-copy">
          {currentLesson
            ? "Pick up where you left off."
            : dueReviewWords > 0
              ? `${dueReviewWords} ${dueReviewWords === 1 ? "word is" : "words are"} ready to remember.`
              : "No words are due. Discover something new whenever you like."}
        </p>
        <div className="next-step-actions">
          {currentLesson ? (
            <Link
              href="/review"
              prefetch="auto"
              className="vn-btn vn-btn-primary"
            >
              Continue review
            </Link>
          ) : dueReviewWords > 0 ? (
            <StartDailyLessonButton
              className="vn-btn vn-btn-primary"
              label="Start review"
            />
          ) : (
            <Link href="/discovery" className="vn-btn vn-btn-primary">
              Discover words
            </Link>
          )}
          <p className="next-step-meta">
            Reviewed today: <strong>{reviewsCompleted}</strong>
          </p>
        </div>
      </section>

      <section aria-labelledby="today-goal-heading" className="today-goal">
        <div className="today-goal-head">
          <h2 id="today-goal-heading" className="section-title">
            Today’s goal
          </h2>
          <p className="goal-status" data-status={mission.status}>
            {MISSION_STATUS[mission.status]}
          </p>
        </div>
        <dl className="goal-rows">
          {goals.map((goal) => (
            <div key={goal.label} className="goal-row">
              <dt>
                {goal.label}
                {goal.optional ? (
                  <span className="ml-1.5 text-sm font-normal text-neutral-600">
                    (optional)
                  </span>
                ) : null}
              </dt>
              <dd>
                {goal.done} of {goal.target}
              </dd>
              <progress
                aria-hidden="true"
                value={Math.min(goal.done, goal.target)}
                max={goal.target}
              />
            </div>
          ))}
        </dl>
        {mission.status === "open" &&
        !currentLesson &&
        dueReviewWords === 0 &&
        reviewsCompleted < mission.reviewTarget ? (
          <p className="pb-2 text-sm text-neutral-600">
            No words are due right now. Save new words in Discover, then check
            back here for review.
          </p>
        ) : null}
      </section>

      <div className="keep-going">
        <h2 className="section-title">Keep going</h2>
        <Suspense
          fallback={
            <section
              aria-label="Your learning path"
              aria-busy="true"
              className="min-h-[9.5rem] rounded-2xl border border-neutral-200 bg-white p-4"
            >
              <h3 className="font-semibold text-neutral-900">
                Today’s practice
              </h3>
              <p role="status" className="mt-1 text-neutral-600">
                Loading practice…
              </p>
            </section>
          }
        >
          <HomePractice read={practiceRead} />
        </Suspense>
        {discoveryIsNext ? null : (
          <div className="keep-going-row">
            <div>
              <h3>New words</h3>
              <p>Hear a word, then choose Learn or Already know.</p>
            </div>
            <Link href="/discovery" className="vn-btn vn-btn-secondary">
              Discover
            </Link>
          </div>
        )}
        <nav aria-label="Explore your English" className="flex flex-wrap gap-2">
          {EXPLORE_LINKS.map(([href, label, prefetch]) => (
            <Link
              key={href}
              href={href}
              prefetch={prefetch}
              className="vn-chip-link"
            >
              {label}
            </Link>
          ))}
        </nav>
      </div>

      <section aria-labelledby="latest-word-heading" className="latest-word">
        {latestWord && wordHref ? (
          <>
            {pictureRead && latestRevision ? (
              <Suspense
                fallback={
                  <MeaningPicture
                    meaningId={latestWord.meaningId}
                    meaningRevision={latestRevision}
                  />
                }
              >
                <HomeMeaningPicture
                  read={pictureRead}
                  meaningId={latestWord.meaningId}
                  meaningRevision={latestRevision}
                />
              </Suspense>
            ) : (
              <MeaningPicture
                meaningId={latestWord.meaningId}
                image={latestWord.image}
                meaningRevision={latestRevision}
              />
            )}
            <div className="latest-word-body">
              <div className="latest-word-head">
                <h2
                  id="latest-word-heading"
                  className="text-sm font-semibold text-neutral-600"
                >
                  Your latest word
                </h2>
                <Link
                  href="/words"
                  prefetch="auto"
                  className="vn-text-link text-sm"
                >
                  All your words
                </Link>
              </div>
              <p className="latest-word-title">
                <Link href={wordHref} prefetch="auto" className="headword">
                  {latestWord.wordText}
                </Link>
                <span className="pos">{latestWord.partOfSpeech}</span>
              </p>
              <p className="text-neutral-700">{latestWord.shortDefinition}</p>
              <div className="latest-word-actions">
                <Link
                  href={`${wordHref}#sentence-practice`}
                  prefetch="auto"
                  className="vn-text-link"
                >
                  Write a sentence with it
                </Link>
              </div>
            </div>
          </>
        ) : (
          <div className="latest-word-empty">
            <h2
              id="latest-word-heading"
              className="text-sm font-semibold text-neutral-600"
            >
              Your latest word
            </h2>
            <p className="headword text-3xl">Nothing saved yet</p>
            <p className="text-neutral-600">
              Choose Learn on a new word and it will wait for you here.
            </p>
            <Link href="/vocabulary" className="vn-text-link">
              Search for a word
            </Link>
          </div>
        )}
      </section>
    </PageContainer>
  );
}

// Practice and stories have instant loading states, which only show at once
// when prefetched; the others prefetch on intent.
const EXPLORE_LINKS = [
  ["/practice", "Practice", "auto"],
  ["/stories", "Short stories", "auto"],
  ["/courses", "Situations", false],
  ["/lists", "Personal lists", false],
  ["/plan", "Your plan", false],
] as const;
