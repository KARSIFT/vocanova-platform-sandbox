import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { LessonPath } from "../_components/lesson-path";
import { RecommendedLesson } from "../_components/lesson-recommendation";

async function optional<T>(request: Promise<T>): Promise<T | null> {
  try {
    return await request;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, "/discover");
    return null;
  }
}

export default async function DiscoverPage() {
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
    requireAuthRedirect(error, "/discover");
  }
  return (
    <PageContainer className="max-w-[72rem]">
      <header className="mb-6 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-neutral-900">
            Journey
          </h1>
        </div>
        <div className="flex flex-wrap gap-x-5">
          <Link
            href="/vocabulary"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
          >
            Find a word
          </Link>
          <Link
            href="/words"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
          >
            Saved words
          </Link>
        </div>
      </header>
      <RecommendedLesson
        data={recommendation?.data ?? null}
        browseHref="#journey-lessons"
        browseLabel={lessonResponse ? "All lessons" : "Explore situations"}
      />
      <nav
        aria-label="Ways to practise"
        className="my-4 flex flex-wrap gap-x-5 border-b border-neutral-200 pb-3"
      >
        {(
          [
            [
              "/practice",
              "Choose a practice",
              "Remember words and revisit mistakes.",
            ],
            [
              "/writing",
              "Topic writing",
              "Put your words into a message of your own.",
            ],
            [
              "/stories",
              "Short stories",
              "Follow a conversation and check its meaning.",
            ],
          ] as const
        ).map(([href, label]) => (
          <Link
            key={href}
            href={href}
            className="inline-flex min-h-11 items-center rounded-lg text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            <span className="block font-semibold">{label}</span>
          </Link>
        ))}
      </nav>
      {!lessonResponse && (
        <p role="status" className="mb-4 text-neutral-700">
          Guided lessons are unavailable right now. You can still explore the
          situations below.
        </p>
      )}
      <div id="journey-lessons" className="scroll-mt-24" tabIndex={-1}>
        <LessonPath
          lessons={lessonResponse?.data.items ?? []}
          situations={response.data.items}
        />
        {!response.data.items.length && !lessonResponse?.data.items.length && (
          <div className="py-8">
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
        <Link
          href="/lists"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
        >
          Personal lists
        </Link>
        <Link
          href="/plan"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
        >
          Your learning plan
        </Link>
      </div>
    </PageContainer>
  );
}
