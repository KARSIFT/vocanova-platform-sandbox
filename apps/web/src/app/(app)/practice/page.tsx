import { StatusMessage } from "@/ui/status-message";
import { ActionLink } from "@/ui/controls";
import type { Metadata } from "next";
import Link from "@/ui/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader, PageSection, Surface } from "@/ui/surface";
import { PracticeEntry } from "./_components/practice-entry";
import { LearnIcon } from "../_components/learn-icons";

export const metadata: Metadata = {
  title: "Practice — Vocanova",
  description: "A learning path for listening, sentences and your own words.",
};

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
    <PageContainer>
      <PageHeader
        title="Today’s practice"
        back={{ href: "/home", label: "Back to Today" }}
      />

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

      <Surface
        aria-labelledby="practice-review-heading"
        tone="flat"
        className="mt-6 flex flex-wrap items-center justify-between gap-3"
      >
        <div className="min-w-0 flex-1 basis-48">
          <h2 id="practice-review-heading" className="section-title">
            Review words
          </h2>
          {dueCount === undefined ? (
            <StatusMessage
              as="p"
              role="status"
              className="mt-1 text-sm text-neutral-600"
            >
              We could not load your review count. Open your reviews to try
              again.
            </StatusMessage>
          ) : (
            <p className="mt-1 text-sm text-neutral-600">
              {dueCount === 0
                ? "No words are due right now."
                : `${dueCount} ${dueCount === 1 ? "word is" : "words are"} due for review.`}
            </p>
          )}
        </div>
        <ActionLink variant="secondary" href="/review">
          {canStartReview ? "Start review" : "Open reviews"}
        </ActionLink>
      </Surface>

      <PageSection
        title="More ways to practise"
        titleId="practice-activities-heading"
        className="mt-8"
      >
        <nav aria-label="Practice activities">
          <ul className="vn-list">
            {[
              {
                href: "/discovery",
                label: "Discover words",
                icon: "discover" as const,
              },
              {
                href: "/stories",
                label: "Short stories",
                icon: "stories" as const,
              },
              {
                href: "/courses",
                label: "Situation lessons",
                icon: "situations" as const,
              },
              {
                href: "/progress/sentences",
                label: "Sentence history",
                icon: "history" as const,
              },
              {
                href: firstWord
                  ? `/vocabulary/${encodeURIComponent(firstWord.wordSlug)}`
                  : "/vocabulary",
                label: firstWord
                  ? `Listen to ${firstWord.wordText}`
                  : "Find a word to hear",
                icon: "listen" as const,
              },
            ].map((item) => (
              <li key={item.label}>
                <Link href={item.href} className="learn-nav-link">
                  <LearnIcon name={item.icon} />
                  <span className="min-w-0 flex-1">{item.label}</span>
                  <LearnIcon name="arrow" />
                </Link>
              </li>
            ))}
          </ul>
        </nav>
      </PageSection>
    </PageContainer>
  );
}
