import type { Metadata } from "next";
import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { PracticeEntry } from "./_components/practice-entry";

export const metadata: Metadata = {
  title: "Practice — Vocanova",
  description: "Review, remember and use your English vocabulary.",
};

const linkStyle =
  "inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

// Each activity can remain useful when another service is unavailable. An
// expired session still redirects instead of looking like an empty vocabulary.
async function loadActivity<T>(request: Promise<T>): Promise<T | null> {
  try {
    return await request;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401) {
      requireAuthRedirect(error, "/practice");
    }
    return null;
  }
}

export default async function PracticePage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  const requestedList =
    query.list !== undefined
      ? typeof query.list === "string" && query.list
        ? query.list
        : "invalid-list-selection"
      : "";
  const client = await createServerApiClient();
  const [
    dueResponse,
    lessonResponse,
    savedResponse,
    missionResponse,
    practiceResponse,
    listsResponse,
  ] = await Promise.all([
    loadActivity(client.listDueWords({ limit: 1 })),
    loadActivity(client.listLessons()),
    loadActivity(client.listSavedWords({ limit: 3 })),
    loadActivity(client.getDailyMission()),
    loadActivity(client.listPracticeSessions()),
    loadActivity(client.listWordLists()),
  ]);

  const dueCount = dueResponse?.data.totalCount;
  const mission = missionResponse?.data;
  const reviewTargetComplete = mission
    ? mission.reviewsCompleted >= mission.reviewTarget
    : false;
  const canStartReview =
    dueCount !== undefined && dueCount > 0 && mission && !reviewTargetComplete;
  const lessons = lessonResponse?.data.items;
  // Prefer vocabulary already taught, never a merely opened lesson or an
  // unstudied recommendation. Explicit URL selections keep their meaning.
  const studiedLesson =
    lessons?.find((lesson) => lesson.status === "completed") ??
    lessons?.find(
      (lesson) =>
        lesson.status === "in_progress" &&
        lesson.completedSteps >= lesson.wordCount,
    );
  const selectedLessonKey =
    query.lesson === undefined
      ? (studiedLesson?.key ?? "")
      : (lessons?.find((lesson) => lesson.key === query.lesson)?.key ?? "");
  const nextLesson =
    lessons?.find((lesson) => lesson.status === "in_progress") ??
    lessons?.find((lesson) => lesson.status === "not_started") ??
    lessons?.[0];
  const savedWords = savedResponse?.data.items;
  const firstWord = savedWords?.[0];

  return (
    <PageContainer className="max-w-[56rem]">
      <Link href="/discover" className={linkStyle}>
        Back to Journey
      </Link>
      <header className="mb-6 mt-3">
        <h1 className="text-3xl font-bold tracking-tight text-neutral-900">
          Practice
        </h1>
      </header>

      <PracticeEntry
        key={`${selectedLessonKey}:${requestedList}`}
        initialSessions={practiceResponse?.data ?? null}
        lessons={lessons ?? []}
        initialLessonKey={selectedLessonKey}
        initialLists={listsResponse?.data.items ?? null}
        initialListId={requestedList}
      />

      <nav
        aria-label="Practice activities"
        className="flex flex-wrap gap-x-6 border-y border-neutral-200 py-3"
      >
        <Link href="/writing" className={linkStyle}>
          Topic writing
        </Link>
        <Link href="/stories" className={linkStyle}>
          Short stories
        </Link>
      </nav>

      <section
        aria-labelledby="practice-review-heading"
        className="flex flex-wrap items-center justify-between gap-3 border-b border-neutral-200 py-5"
      >
        <div className="min-w-0 flex-1 basis-48">
          <h2
            id="practice-review-heading"
            className="text-lg font-bold text-neutral-900"
          >
            Scheduled review
          </h2>
          {dueCount === undefined ? (
            <p role="status" className="mt-1 text-sm text-neutral-600">
              We could not load your review count. Open your reviews to try
              again.
            </p>
          ) : (
            <p className="mt-1 text-sm text-neutral-600">
              {dueCount === 0
                ? "No words are due right now."
                : `${dueCount} ${dueCount === 1 ? "word is" : "words are"} due for review.`}
            </p>
          )}
          {reviewTargetComplete && (
            <p className="mt-1 text-sm text-neutral-600">
              Today’s review target is complete. Your next review session is
              tomorrow.
            </p>
          )}
        </div>
        <Link href="/review" className={linkStyle}>
          {canStartReview ? "Start review" : "View reviews"}
        </Link>
      </section>

      <section
        aria-labelledby="practice-lesson-heading"
        className="flex flex-wrap items-center justify-between gap-3 border-b border-neutral-200 py-5"
      >
        <div className="min-w-0 flex-1 basis-48">
          <h2
            id="practice-lesson-heading"
            className="text-lg font-bold text-neutral-900"
          >
            Guided lessons
          </h2>
          {nextLesson ? (
            <>
              <p className="mt-1 text-sm text-neutral-600">
                {nextLesson.title}
              </p>
              {nextLesson.status === "in_progress" && (
                <p className="mt-1 text-sm text-neutral-600">
                  {nextLesson.completedSteps} of {nextLesson.stepCount} steps
                  saved
                </p>
              )}
            </>
          ) : (
            <p
              role={lessons ? undefined : "status"}
              className="mt-1 text-sm text-neutral-600"
            >
              {lessons
                ? "No guided lessons are available yet."
                : "We could not load your lessons. You can still practise with your vocabulary."}
            </p>
          )}
        </div>
        <Link
          href={
            nextLesson
              ? `/learn/${encodeURIComponent(nextLesson.key)}`
              : "/discover"
          }
          className={linkStyle}
        >
          {nextLesson
            ? nextLesson.status === "in_progress"
              ? "Continue lesson"
              : nextLesson.status === "completed"
                ? "Revisit lesson"
                : "Start lesson"
            : "Explore situations"}
        </Link>
      </section>

      <section aria-labelledby="practice-writing-heading" className="py-5">
        <h2
          id="practice-writing-heading"
          className="scroll-mt-24 text-lg font-bold text-neutral-900"
        >
          Write a sentence
        </h2>
        {savedWords?.length ? (
          <div className="mt-1 flex flex-wrap gap-x-5">
            {savedWords.map((word) => (
              <Link
                key={word.userWordId}
                href={`/words/${encodeURIComponent(word.userWordId)}#sentence-practice`}
                className={linkStyle}
              >
                Write with {word.wordText}
              </Link>
            ))}
            <Link href="/words" className={linkStyle}>
              Choose another saved word
            </Link>
          </div>
        ) : (
          <>
            <p
              role={savedWords ? undefined : "status"}
              className="mt-1 text-sm text-neutral-600"
            >
              {savedWords
                ? "Save a word first, then use it in your own sentence."
                : "We could not load your saved words. Open your vocabulary to try again."}
            </p>
            <Link
              href={savedWords ? "/vocabulary" : "/words"}
              className={linkStyle}
            >
              {savedWords ? "Find a word to save" : "Open saved vocabulary"}
            </Link>
          </>
        )}
      </section>
      <div className="flex flex-wrap gap-x-6 border-t border-neutral-200 pt-3">
        <Link href="/progress/sentences" className={linkStyle}>
          Open sentence history
        </Link>
        <Link
          href={
            firstWord
              ? `/vocabulary/${encodeURIComponent(firstWord.wordSlug)}`
              : "/vocabulary"
          }
          className={linkStyle}
        >
          {firstWord
            ? `Listen to ${firstWord.wordText}`
            : "Find a word to hear"}
        </Link>
      </div>
    </PageContainer>
  );
}
