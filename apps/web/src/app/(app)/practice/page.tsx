import type { Metadata } from "next";
import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";
import { PracticeEntry } from "./_components/practice-entry";

export const metadata: Metadata = {
  title: "Practice — Vocanova",
  description: "Review, remember and use your English vocabulary.",
};

const actionStyle =
  "mt-4 inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
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
  const client = await createServerApiClient();
  const [
    dueResponse,
    lessonResponse,
    savedResponse,
    missionResponse,
    practiceResponse,
  ] = await Promise.all([
    loadActivity(client.listDueWords({ limit: 1 })),
    loadActivity(client.listLessons()),
    loadActivity(client.listSavedWords({ limit: 3 })),
    loadActivity(client.getDailyMission()),
    loadActivity(client.listPracticeSessions()),
  ]);

  const dueCount = dueResponse?.data.totalCount;
  const mission = missionResponse?.data;
  const reviewTargetComplete = mission
    ? mission.reviewsCompleted >= mission.reviewTarget
    : false;
  const canStartReview =
    dueCount !== undefined && dueCount > 0 && mission && !reviewTargetComplete;
  const lessons = lessonResponse?.data.items;
  const selectedLessonKey =
    lessons?.find((lesson) => lesson.key === query.lesson)?.key ?? "";
  const nextLesson =
    lessons?.find((lesson) => lesson.status === "in_progress") ??
    lessons?.find((lesson) => lesson.status === "not_started") ??
    lessons?.[0];
  const savedWords = savedResponse?.data.items;
  const firstWord = savedWords?.[0];

  return (
    <PageContainer className="max-w-[64rem]">
      <Link href="/discover" className={linkStyle}>
        Back to Journey
      </Link>
      <header className="mb-6 mt-3">
        <h1 className="text-3xl font-bold tracking-tight text-neutral-900">
          Practice your way
        </h1>
        <p className="mt-2 max-w-[40rem] text-lg text-neutral-700">
          Remember a word, use it in a sentence, or hear how it sounds. Choose
          what you want to work on today.
        </p>
      </header>

      <PracticeEntry
        key={selectedLessonKey}
        initialSessions={practiceResponse?.data ?? null}
        lessons={lessons ?? []}
        initialLessonKey={selectedLessonKey}
      />

      <div className="grid gap-4 md:grid-cols-2">
        <Surface aria-labelledby="practice-review-heading" tone="primary">
          <Eyebrow>Keep words fresh</Eyebrow>
          <h2
            id="practice-review-heading"
            className="mt-2 text-xl font-bold text-neutral-900"
          >
            Scheduled review
          </h2>
          {dueCount === undefined ? (
            <p role="status" className="mt-3 text-neutral-700">
              We could not load your review count. Open your reviews to try
              again.
            </p>
          ) : (
            <p className="mt-3 font-semibold text-neutral-900">
              {dueCount === 0
                ? "No words are due right now."
                : `${dueCount} ${dueCount === 1 ? "word is" : "words are"} due for review.`}
            </p>
          )}
          <p className="mt-2 text-neutral-700">
            {reviewTargetComplete
              ? "Today’s review target is complete. Your next review session is tomorrow."
              : canStartReview
                ? "A short session helps you remember the words you saved."
                : "Reviews follow your saved words and daily target. You can also choose another activity below."}
          </p>
          <Link href="/review" className={actionStyle}>
            {canStartReview ? "Start review" : "View reviews"}
          </Link>
        </Surface>

        <Surface aria-labelledby="practice-lesson-heading">
          <Eyebrow>Remember and use</Eyebrow>
          <h2
            id="practice-lesson-heading"
            className="mt-2 text-xl font-bold text-neutral-900"
          >
            Guided lessons
          </h2>
          {nextLesson ? (
            <>
              <p className="mt-3 font-semibold text-neutral-900">
                {nextLesson.title}
              </p>
              <p className="mt-2 text-neutral-700">{nextLesson.description}</p>
              <p className="mt-2 text-sm text-neutral-600">
                {nextLesson.status === "in_progress"
                  ? `${nextLesson.completedSteps} of ${nextLesson.stepCount} steps saved. Pick up where you left off.`
                  : nextLesson.status === "completed"
                    ? "Lesson complete. Revisit its words and examples."
                    : `${nextLesson.wordCount} words, with meaning and context questions.`}
              </p>
              <Link
                href={`/learn/${encodeURIComponent(nextLesson.key)}`}
                className={actionStyle}
              >
                {nextLesson.status === "in_progress"
                  ? "Continue lesson"
                  : nextLesson.status === "completed"
                    ? "Revisit lesson"
                    : "Start lesson"}
              </Link>
            </>
          ) : (
            <>
              <p
                role={lessons ? undefined : "status"}
                className="mt-3 text-neutral-700"
              >
                {lessons
                  ? "There are no guided lessons available right now. Explore words in a real-life situation."
                  : "We could not load your lessons. You can still practise with your vocabulary."}
              </p>
              <Link href="/discover" className={actionStyle}>
                Explore situations
              </Link>
            </>
          )}
        </Surface>

        <Surface aria-labelledby="practice-writing-heading">
          <Eyebrow>Make it your own</Eyebrow>
          <h2
            id="practice-writing-heading"
            className="mt-2 text-xl font-bold text-neutral-900"
          >
            Write a sentence
          </h2>
          {savedWords?.length ? (
            <>
              <p className="mt-3 text-neutral-700">
                Choose a saved meaning, write your own sentence, and get focused
                feedback.
              </p>
              <ul className="mt-3 divide-y divide-neutral-200">
                {savedWords.map((word) => (
                  <li key={word.userWordId} className="py-2">
                    <Link
                      href={`/words/${encodeURIComponent(word.userWordId)}`}
                      className={linkStyle}
                    >
                      Write with {word.wordText}
                    </Link>
                    <p className="text-sm text-neutral-700">
                      {word.shortDefinition}
                    </p>
                  </li>
                ))}
              </ul>
              <Link href="/words" className={`${linkStyle} mt-3`}>
                Choose another saved word
              </Link>
            </>
          ) : (
            <>
              <p
                role={savedWords ? undefined : "status"}
                className="mt-3 text-neutral-700"
              >
                {savedWords
                  ? "Save a word first, then use its meaning in a sentence of your own."
                  : "We could not load your saved words. Open your vocabulary to try again."}
              </p>
              <Link
                href={savedWords ? "/vocabulary" : "/words"}
                className={actionStyle}
              >
                {savedWords ? "Find a word to save" : "Open saved vocabulary"}
              </Link>
            </>
          )}
        </Surface>

        <Surface aria-labelledby="practice-listening-heading" tone="secondary">
          <Eyebrow>Hear it, then say it</Eyebrow>
          <h2
            id="practice-listening-heading"
            className="mt-2 text-xl font-bold text-neutral-900"
          >
            Listen and repeat
          </h2>
          <p className="mt-3 text-neutral-700">
            Open a word and use device pronunciation to hear the word and its
            examples. Try the slower speed, then say it aloud yourself.
          </p>
          <p className="mt-2 text-sm text-neutral-600">
            This is listening practice. Your voice is not recorded or scored.
          </p>
          <Link
            href={
              firstWord
                ? `/vocabulary/${encodeURIComponent(firstWord.wordSlug)}`
                : "/vocabulary"
            }
            className={actionStyle}
          >
            {firstWord
              ? `Listen to ${firstWord.wordText}`
              : "Find a word to hear"}
          </Link>
        </Surface>
      </div>

      <Surface aria-labelledby="practice-history-heading" className="mt-4">
        <h2
          id="practice-history-heading"
          className="text-xl font-bold text-neutral-900"
        >
          Learn from your sentences
        </h2>
        <p className="mt-2 text-neutral-700">
          Return to sentences you have written and read their feedback. Notice
          what worked and what you want to try next.
        </p>
        <Link href="/progress/sentences" className={actionStyle}>
          Open sentence history
        </Link>
      </Surface>
    </PageContainer>
  );
}
