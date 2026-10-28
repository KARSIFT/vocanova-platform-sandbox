import type { Metadata } from "next";
import Link from "@/ui/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { PracticeEntry } from "./_components/practice-entry";

export const metadata: Metadata = {
  title: "Practice — Vocanova",
  description: "A learning path for listening, sentences and your own words.",
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
    practiceResponse,
    listsResponse,
  ] = await Promise.all([
    loadActivity(client.listDueWords({ limit: 1 })),
    loadActivity(client.listLessons()),
    loadActivity(client.listSavedWords({ limit: 3 })),
    loadActivity(client.listPracticeSessions()),
    loadActivity(client.listWordLists()),
  ]);

  const dueCount = dueResponse?.data.totalCount;
  const canStartReview = dueCount !== undefined && dueCount > 0;
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
  const savedWords = savedResponse?.data.items;
  const firstWord = savedWords?.[0];

  return (
    <PageContainer className="max-w-[56rem]">
      <Link href="/home" className={linkStyle}>
        Back to Today
      </Link>
      <header className="mb-4 mt-2">
        <h1 className="text-2xl font-bold tracking-tight text-neutral-900">
          Today’s practice
        </h1>
      </header>

      <PracticeEntry
        key={`${selectedLessonKey}:${requestedList}`}
        initialSessions={practiceResponse?.data ?? null}
        lessons={lessons ?? []}
        lessonsUnavailable={!lessonResponse}
        initialLessonKey={selectedLessonKey}
        initialLists={listsResponse?.data.items ?? null}
        initialListId={requestedList}
        explicitSelection={
          query.list !== undefined || query.lesson !== undefined
        }
      />

      <section
        aria-labelledby="practice-review-heading"
        className="mt-6 flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-white p-4"
      >
        <div className="min-w-0 flex-1 basis-48">
          <h2 id="practice-review-heading" className="section-title">
            Review words
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
        </div>
        <Link href="/review" className="vn-btn vn-btn-secondary min-h-11">
          {canStartReview ? "Start review" : "Open reviews"}
        </Link>
      </section>

      {/* One list of the other ways to practise, instead of scattered rows. */}
      <nav aria-label="Practice activities" className="mt-6">
        <h2 className="section-title">More ways to practise</h2>
        <ul className="mt-2 divide-y divide-neutral-200 rounded-2xl border border-neutral-200 bg-white">
          {[
            {
              href: "/discovery",
              label: "Discover words",
              text: "Hear a new word and choose Learn or Already know.",
            },
            {
              href: "/stories",
              label: "Short stories",
              text: "Read your words inside a short conversation.",
            },
            {
              href: "/courses",
              label: "Explore situation lessons",
              text: "Guided lessons for a place, such as a cafe or an airport.",
            },
            {
              href: "/progress/sentences",
              label: "Open sentence history",
              text: "Your sentences and their feedback.",
            },
            {
              href: firstWord
                ? `/vocabulary/${encodeURIComponent(firstWord.wordSlug)}`
                : "/vocabulary",
              label: firstWord
                ? `Listen to ${firstWord.wordText}`
                : "Find a word to hear",
              text: firstWord
                ? "Hear your latest word again."
                : "Search any word and hear it.",
            },
          ].map((item) => (
            <li key={item.label} className="px-4 py-2">
              <Link href={item.href} className={linkStyle}>
                {item.label}
              </Link>
              <p className="pb-1 text-sm text-neutral-600">{item.text}</p>
            </li>
          ))}
        </ul>
      </nav>
    </PageContainer>
  );
}
