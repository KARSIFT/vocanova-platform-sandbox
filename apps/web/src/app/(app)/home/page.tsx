import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { MeaningPicture } from "@/ui/meaning-picture";
import { SentenceFeedback } from "../_components/sentence-feedback";
import { RecommendedLesson } from "../_components/lesson-recommendation";

export default async function HomePage() {
  const client = await createServerApiClient();
  let savedWordsResponse: Awaited<ReturnType<typeof client.listSavedWords>>;
  let dueResponse: Awaited<ReturnType<typeof client.listDueWords>>;
  let dailyMissionResponse: Awaited<ReturnType<typeof client.getDailyMission>>;
  let currentUserResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  let recommendation: Awaited<
    ReturnType<typeof client.getLessonRecommendation>
  > | null;
  try {
    [
      savedWordsResponse,
      dueResponse,
      dailyMissionResponse,
      currentUserResponse,
      recommendation,
    ] = await Promise.all([
      client.listSavedWords({ limit: 3 }),
      client.listDueWords({ limit: 1 }),
      client.getDailyMission(),
      client.getCurrentUser(),
      client.getLessonRecommendation().catch((error: unknown) => {
        if (error instanceof ApiResponseError && error.status === 401)
          requireAuthRedirect(error, "/home");
        return null;
      }),
    ]);
  } catch (error) {
    requireAuthRedirect(error, "/home");
  }
  const { items: savedWords } = savedWordsResponse.data;
  const firstWord = savedWords[0];
  const dueReviewWords = dueResponse.data.totalCount;
  const mission = dailyMissionResponse.data;
  const {
    reviewTarget: missionTargetWords,
    reviewsCompleted: reviewedWordsToday,
    newWordTarget,
    newWordsCompleted,
    sentencePracticeTarget,
    sentencePracticesCompleted,
    streak,
  } = mission;
  const missionComplete = mission.status === "completed";
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
  const primaryAction = missionComplete
    ? {
        href: "/discover",
        label: "Explore a new situation",
        title: "You made time for English today.",
        detail:
          "Your daily goal is complete. Explore something new whenever you like.",
      }
    : dueReviewWords > 0 && hasRemainingReviewTarget
      ? {
          href: "/review",
          label: "Start review",
          title: "Keep your words fresh.",
          detail: "A short session with the words due now.",
        }
      : savedWords.length === 0
        ? {
            href: "/discover",
            label: "Start your Journey",
            title: "Start with English you can use.",
            detail:
              "Choose a real-life situation and save a word you want to remember.",
          }
        : hasRemainingNewWordTarget || hasRemainingReviewTarget
          ? {
              href: "/discover",
              label: "Explore a new situation",
              title: "Find your next useful word.",
              detail: hasRemainingReviewTarget
                ? "Nothing is due right now. Add a useful word to keep building today’s practice."
                : "Choose a useful word to continue today’s mission.",
            }
          : remainingSentencePractices !== null &&
              remainingSentencePractices > 0
            ? {
                href: "#sentence-practice",
                label: "Practice a sentence",
                title: "Make a word your own.",
                detail: "Use a word you saved in a sentence about your life.",
              }
            : {
                href: "/progress",
                label: "View your progress",
                title: "See what you have practised.",
                detail:
                  "Your saved words, lessons and writing are waiting for you.",
              };
  const missionProgressPercent =
    missionTargetWords > 0
      ? Math.min(
          100,
          Math.round((reviewedWordsToday / missionTargetWords) * 100),
        )
      : 0;
  const link =
    "inline-flex min-h-11 items-center rounded-lg font-semibold text-primary-700 hover:text-primary-800";

  return (
    <PageContainer className="max-w-[72rem]">
      <header className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-neutral-900">
            Today
          </h1>
          <p className="mt-1 text-neutral-600">
            A little English. Something you can use.
          </p>
        </div>
        <p className="inline-flex items-center gap-2 rounded-full bg-white px-4 py-2 text-sm font-semibold text-neutral-700">
          <span
            aria-hidden="true"
            className="size-2 rounded-full bg-primary-600"
          />
          {streak.currentStreakCount}-day streak
        </p>
      </header>
      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1.55fr)_minmax(18rem,0.85fr)]">
        <div className="min-w-0 space-y-6">
          <section
            aria-label={
              missionComplete ? "Mission complete" : "Today’s practice"
            }
            className="overflow-hidden rounded-2xl bg-primary-800 p-5 text-white sm:p-7"
          >
            <p className="text-sm font-semibold text-primary-100">
              {missionComplete ? "Mission complete" : "Today’s practice"}
            </p>
            <h2
              id="todays-mission-heading"
              className="mt-2 max-w-[30rem] text-[clamp(1.65rem,3vw,2.25rem)] font-bold leading-tight tracking-tight"
            >
              {primaryAction.title}
            </h2>
            <p className="mt-3 max-w-[34rem] text-primary-100">
              {primaryAction.detail}
            </p>
            <Link
              href={primaryAction.href}
              className="mt-5 inline-flex min-h-12 items-center justify-center rounded-xl bg-white px-6 py-3 font-bold text-primary-800 hover:bg-primary-50"
            >
              {primaryAction.label}
            </Link>
            <div className="mt-5 flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-primary-700 pt-4 text-sm text-primary-100">
              <span>
                Reviews today: {reviewedWordsToday} of {missionTargetWords}
              </span>
              <span>
                Currently due: {dueReviewWords}{" "}
                {dueReviewWords === 1 ? "word" : "words"}.
              </span>
              <div
                role="progressbar"
                aria-label="Reviews today"
                aria-valuemin={0}
                aria-valuemax={missionTargetWords}
                aria-valuenow={Math.min(reviewedWordsToday, missionTargetWords)}
                className="h-1.5 w-full overflow-hidden rounded-full bg-primary-900/50"
              >
                <div
                  className="h-full rounded-full bg-primary-200"
                  style={{ width: `${missionProgressPercent}%` }}
                />
              </div>
            </div>
          </section>
          <RecommendedLesson data={recommendation?.data ?? null} compact />
          <section
            id="sentence-practice"
            aria-labelledby="use-a-word-heading"
            className="scroll-mt-24 border-t border-neutral-200 pt-6"
          >
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2
                id="use-a-word-heading"
                className="text-2xl font-bold tracking-tight text-neutral-900"
              >
                Use a word
              </h2>
              <Link href="/writing" className={link}>
                Choose a writing topic
              </Link>
            </div>
            {firstWord ? (
              <div className="mt-3">
                <p className="text-neutral-700">
                  Write something true for you with{" "}
                  <strong className="text-neutral-900">
                    {firstWord.wordText}
                  </strong>
                  .
                </p>
                <SentenceFeedback
                  targetWord={firstWord.wordText}
                  attemptId={firstWord.userWordId}
                  source="word_detail"
                  userId={currentUserResponse.data.id}
                  shortDefinition={firstWord.shortDefinition}
                />
              </div>
            ) : (
              <div className="mt-3">
                <p className="text-neutral-700">
                  Save a useful word, then try it in your own sentence.
                </p>
                <Link href="/vocabulary" className={link}>
                  Find a word to write with
                </Link>
              </div>
            )}
          </section>
        </div>
        <aside
          aria-labelledby="saved-words-heading"
          className="min-w-0 rounded-2xl bg-white p-5"
        >
          <div className="flex items-center justify-between gap-3">
            <h2
              id="saved-words-heading"
              className="text-xl font-bold tracking-tight text-neutral-900"
            >
              Your vocabulary
            </h2>
            <Link href="/words" className={`${link} text-sm`}>
              See all
            </Link>
          </div>
          {firstWord ? (
            <>
              <MeaningPicture meaningId={firstWord.meaningId} />
              <ul className="mt-3 divide-y divide-neutral-200">
                {savedWords.map((word) => (
                  <li key={word.userWordId} className="py-3">
                    <Link
                      href={`/words/${encodeURIComponent(word.userWordId)}`}
                      className={`${link} text-lg`}
                    >
                      {word.wordText}
                    </Link>
                    <p className="text-sm text-neutral-600">
                      {word.shortDefinition}
                    </p>
                  </li>
                ))}
              </ul>
            </>
          ) : (
            <p className="mt-4 text-neutral-600">
              Your saved words will live here. Begin with a situation you know.
            </p>
          )}
          <div className="mt-3 border-t border-neutral-200 pt-3">
            <Link href="/lists" className={link}>
              Personal lists
            </Link>
          </div>
        </aside>
      </div>
      <nav
        aria-label="Explore your English"
        className="mt-8 grid gap-3 border-t border-neutral-200 pt-6 sm:grid-cols-3"
      >
        {(
          [
            [
              "/practice",
              "Practise what you learned",
              "Remember words or listen again.",
            ],
            [
              "/stories",
              "Read a short story",
              "Follow English in everyday life.",
            ],
            ["/plan", "Your learning plan", "Choose what matters to you."],
          ] as const
        ).map(([href, label, detail]) => (
          <Link
            key={href}
            href={href}
            className="rounded-xl px-4 py-3 hover:bg-white"
          >
            <span className="block font-bold text-primary-700">{label}</span>
            <span className="mt-1 block text-sm text-neutral-600">
              {detail}
            </span>
          </Link>
        ))}
      </nav>
    </PageContainer>
  );
}
