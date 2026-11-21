import { StatusMessage } from "@/ui/status-message";
import Link from "@/ui/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader } from "@/ui/surface";
import { LessonPath } from "../_components/lesson-path";
import { RecommendedLesson } from "../_components/lesson-recommendation";

async function optional<T>(request: Promise<T>): Promise<T | null> {
  try {
    return await request;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, "/courses");
    return null;
  }
}

export default async function CoursesPage() {
  const client = await createServerApiClient();
  let response: Awaited<ReturnType<typeof client.listJourneySituations>>;
  let lessonResponse: Awaited<ReturnType<typeof client.listLessons>> | null;
  let recommendation: Awaited<
    ReturnType<typeof client.getLessonRecommendation>
  > | null;
  try {
    [response, lessonResponse, recommendation] = await Promise.all([
      client.listJourneySituations(),
      optional(client.listLessons()),
      optional(client.getLessonRecommendation()),
    ]);
  } catch (error) {
    requireAuthRedirect(error, "/courses");
  }
  return (
    <PageContainer width="wide">
      <PageHeader
        title="Situation courses"
        lede="Short guided lessons for the places you use English."
        actions={
          <>
            <Link href="/vocabulary" className="vn-text-link">
              Find a word
            </Link>
            <Link href="/words" className="vn-text-link">
              Saved words
            </Link>
          </>
        }
      />
      <RecommendedLesson
        data={recommendation?.data ?? null}
        browseHref="#journey-lessons"
        browseLabel={lessonResponse ? "All lessons" : "Explore situations"}
      />
      <nav aria-label="Ways to practise" className="my-5 flex flex-wrap gap-2">
        <Link href="/practice" className="vn-chip-link">
          Choose a practice
        </Link>
        <Link href="/stories" className="vn-chip-link">
          Short stories
        </Link>
      </nav>
      {!lessonResponse && (
        <StatusMessage as="p" role="status" className="mb-4 text-neutral-700">
          Guided lessons are unavailable right now. You can still explore the
          situations below.
        </StatusMessage>
      )}
      <div id="journey-lessons" className="scroll-mt-24" tabIndex={-1}>
        <LessonPath
          lessons={lessonResponse?.data.items ?? []}
          situations={response.data.items}
        />
        {!response.data.items.length && !lessonResponse?.data.items.length && (
          <div className="vn-empty">
            <h2 className="text-xl font-semibold text-neutral-900">
              New situations are on the way
            </h2>
            <p className="mt-2 text-neutral-700">
              Review your saved words while we add more situations.
            </p>
            <Link
              href="/review"
              className="mt-3 inline-flex min-h-11 items-center font-semibold text-primary-700"
            >
              Go to Reviews
            </Link>
          </div>
        )}
      </div>
      <div className="mt-6 flex flex-wrap gap-x-6 border-t border-neutral-200 pt-3">
        <Link href="/lists" className="vn-text-link">
          Personal lists
        </Link>
        <Link href="/plan" className="vn-text-link">
          Your learning plan
        </Link>
      </div>
    </PageContainer>
  );
}
