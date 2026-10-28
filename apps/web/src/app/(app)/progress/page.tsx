import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, Surface } from "@/ui/surface";
import { getCompletionDayView } from "./completion-day-view";
import { Achievements } from "../_components/achievements";

function formatActivityDate(localDate: string): string {
  const date = new Date(`${localDate}T00:00:00Z`);
  return new Intl.DateTimeFormat(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(date);
}

export default async function ProgressPage() {
  const client = await createServerApiClient();
  let progressResponse: Awaited<ReturnType<typeof client.getProgress>>;
  let knowledgeResponse: Awaited<
    ReturnType<typeof client.getKnowledgeSummary>
  > | null;
  let lessonResponse: Awaited<ReturnType<typeof client.listLessons>> | null;
  let achievementsResponse: Awaited<
    ReturnType<typeof client.listAchievements>
  > | null;
  let sentencesResponse: Awaited<
    ReturnType<typeof client.listLearnerSentences>
  > | null;
  async function optional<T>(request: Promise<T>): Promise<T | null> {
    try {
      return await request;
    } catch (error) {
      if (error instanceof ApiResponseError && error.status === 401)
        requireAuthRedirect(error, "/progress");
      return null;
    }
  }
  try {
    [
      progressResponse,
      knowledgeResponse,
      lessonResponse,
      achievementsResponse,
      sentencesResponse,
    ] = await Promise.all([
      client.getProgress(),
      optional(client.getKnowledgeSummary()),
      optional(client.listLessons()),
      optional(client.listAchievements()),
      optional(
        client.listLearnerSentences({ limit: 2 }, { cache: "no-store" }),
      ),
    ]);
  } catch (error) {
    requireAuthRedirect(error, "/progress");
  }

  const {
    confidencePointsBalance: confidencePointsTotal,
    streak,
    completionHistory,
  } = progressResponse.data;
  const {
    currentStreakCount: currentStreakDays,
    longestStreakCount: longestStreakDays,
  } = streak;

  const historyWithLabels = [...completionHistory]
    .sort((first, second) => first.localDate.localeCompare(second.localDate))
    .map((day) => ({
      ...day,
      label: formatActivityDate(day.localDate),
      view: getCompletionDayView(day),
    }));

  const summaryStyle =
    "min-h-12 cursor-pointer content-center font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
  const linkStyle =
    "inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800";
  const knowledge = knowledgeResponse?.data;
  const stages = knowledge
    ? ([
        { label: "New", count: knowledge.new, swatch: "bg-neutral-400" },
        {
          label: "Learning",
          count: knowledge.learning,
          swatch: "bg-secondary-400",
        },
        {
          label: "Reviewing",
          count: knowledge.reviewing,
          swatch: "bg-primary-400",
        },
        {
          label: "Mastered",
          count: knowledge.mastered,
          swatch: "bg-primary-700",
        },
      ] as const)
    : [];

  return (
    <PageContainer className="max-w-[64rem]">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="page-title">Progress</h1>
          <p className="page-lede">
            Your saved words, recent days and milestones, as recorded.
          </p>
        </div>
        <Link href="/words" className="vn-chip-link">
          Saved words
        </Link>
      </div>

      {knowledge ? (
        <Surface className="mt-6" aria-labelledby="vocabulary-map-heading">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2
              id="vocabulary-map-heading"
              className="text-xl font-bold text-neutral-900"
            >
              Your vocabulary map
            </h2>
            <Link href="/vocabulary?view=map" className={linkStyle}>
              Explore knowledge map
            </Link>
          </div>
          <StageBar stages={stages} />
          <dl className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-4 sm:gap-3">
            {stages.map(({ label, count, swatch }) => (
              <div
                key={label}
                className="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3"
              >
                <dt className="flex items-center gap-2 text-sm text-neutral-600">
                  <span
                    aria-hidden="true"
                    className={`size-2.5 rounded-full ${swatch}`}
                  />
                  {label}
                </dt>
                <dd className="mt-1 text-3xl font-semibold tabular-nums text-neutral-900">
                  {count}
                </dd>
              </div>
            ))}
          </dl>
          <div className="mt-5 flex flex-wrap items-center justify-between gap-2 border-t border-neutral-200 pt-3">
            <p className="text-sm text-neutral-600">
              {knowledge.saved} saved · {knowledge.due} ready for review
            </p>
            {knowledge.due > 0 && (
              <Link href="/review" className={linkStyle}>
                Start review
              </Link>
            )}
          </div>
          <details className="mt-1">
            <summary className={summaryStyle}>About these stages</summary>
            <p className="mt-2 text-sm text-neutral-600">
              New words become Learning, Reviewing and Mastered through your
              review history. These stages describe saved words, not your
              overall English level.
            </p>
            <Link
              href="/vocabulary?view=map&knowledge=known"
              className={`${linkStyle} mt-2`}
            >
              {knowledge.selfReportedKnown}{" "}
              {knowledge.selfReportedKnown === 1 ? "meaning" : "meanings"}{" "}
              marked already known
            </Link>
            <p className="text-sm text-neutral-600">
              Your own assessment, counted separately from review progress.
            </p>
            {(knowledge.ignored > 0 || knowledge.archived > 0) && (
              <p className="mt-2 text-sm text-neutral-600">
                Also saved: {knowledge.ignored} paused and {knowledge.archived}{" "}
                archived meanings.
              </p>
            )}
          </details>
        </Surface>
      ) : (
        <p role="status" className="mt-6 text-neutral-700">
          Your vocabulary map is unavailable right now. Your other progress is
          below.
        </p>
      )}

      <Surface className="mt-4" aria-labelledby="completion-history-heading">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2
            id="completion-history-heading"
            className="text-xl font-bold text-neutral-900"
          >
            Recent activity
          </h2>
          <p className="inline-flex min-h-9 items-center gap-1.5 rounded-full bg-primary-50 px-3 text-sm font-semibold text-primary-800">
            <svg
              aria-hidden="true"
              viewBox="0 0 24 24"
              className="size-4 fill-none stroke-current stroke-[1.8]"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M13 3c1 5-4 5-3 9 1-1 2-2 2-3 4 2 6 5 5 8a6 6 0 0 1-11-1c-1-5 4-7 7-13Z"
              />
            </svg>
            {currentStreakDays}-day streak
          </p>
        </div>
        {historyWithLabels.length > 0 ? (
          <ul className="mt-4 grid gap-2 sm:grid-cols-2">
            {historyWithLabels.map((day) => (
              <li
                key={day.localDate}
                className={`flex min-h-12 min-w-0 items-center justify-between gap-3 rounded-xl border-l-4 px-4 py-2.5 ${day.view.className}`}
              >
                <time
                  dateTime={day.localDate}
                  className="shrink-0 text-sm font-semibold"
                >
                  {day.label}
                </time>
                <p className="min-w-0 text-right text-sm">{day.view.label}</p>
              </li>
            ))}
          </ul>
        ) : (
          <div className="mt-3">
            <p className="text-neutral-700">
              Complete a daily mission to start your history.
            </p>
            <Link href="/home" className={`${linkStyle} mt-2`}>
              Go to today&apos;s mission
            </Link>
          </div>
        )}
      </Surface>

      <div className="mt-4 divide-y divide-neutral-200 rounded-[1.25rem] border border-neutral-200 bg-white px-[var(--spacing-md)] shadow-[var(--vn-shadow-card)] sm:px-[var(--spacing-lg)]">
        {lessonResponse ? (
          <details className="py-2">
            <summary className={summaryStyle}>
              Your learning path{" "}
              <span className="ml-2 text-sm font-normal text-neutral-600">
                {
                  lessonResponse.data.items.filter(
                    (item) => item.status === "completed",
                  ).length
                }{" "}
                of {lessonResponse.data.items.length} lessons
              </span>
            </summary>
            <ul className="divide-y divide-neutral-200 pb-2">
              {lessonResponse.data.items
                .filter(
                  (item) =>
                    item.status === "in_progress" ||
                    item.status === "completed",
                )
                .map((item) => (
                  <li key={item.key} className="py-3">
                    <Link
                      href={`/learn/${encodeURIComponent(item.key)}`}
                      className={linkStyle}
                    >
                      {item.title}
                    </Link>
                    <p className="text-sm text-neutral-600">
                      {item.status === "completed"
                        ? "Completed"
                        : `${item.completedSteps} of ${item.stepCount} steps saved`}
                    </p>
                  </li>
                ))}
            </ul>
            <Link href="/courses" className={`${linkStyle} mb-2`}>
              Browse courses
            </Link>
          </details>
        ) : (
          <p role="status" className="py-4 text-neutral-700">
            Your lesson progress could not load.{" "}
            <Link href="/courses" className={linkStyle}>
              Browse courses
            </Link>
          </p>
        )}

        <section aria-labelledby="sentence-history-heading" className="py-2">
          <details>
            <summary className={summaryStyle}>
              <span id="sentence-history-heading">Sentence practice</span>
            </summary>
            {sentencesResponse &&
              (sentencesResponse.data.items.length > 0 ? (
                <ul aria-label="Recent sentences" className="mt-2 space-y-3">
                  {sentencesResponse.data.items.map((sentence) => (
                    <li
                      key={sentence.id}
                      className="border-l-2 border-neutral-200 py-2 pl-4"
                    >
                      <p className="break-words text-neutral-900 [overflow-wrap:anywhere]">
                        {sentence.originalSentence}
                      </p>
                      <p className="mt-1 text-sm text-neutral-600">
                        {sentence.processingStatus === "pending"
                          ? "Feedback is being prepared"
                          : sentence.processingStatus === "completed"
                            ? "Feedback saved in your sentence history"
                            : "Sentence saved — feedback unavailable"}
                      </p>
                    </li>
                  ))}
                </ul>
              ) : (
                <Link href="/writing" className={`${linkStyle} mt-2`}>
                  Write your first sentence
                </Link>
              ))}
            <Link
              href="/progress/sentences"
              className={`${linkStyle} mt-2 mb-2`}
            >
              View sentence history
            </Link>
          </details>
          {!sentencesResponse && (
            <p role="status" className="pb-3 text-neutral-700">
              Your recent writing could not load.{" "}
              <Link href="/progress/sentences" className={linkStyle}>
                Open sentence history to try again.
              </Link>
            </p>
          )}
        </section>

        <section aria-label="Learning summary" className="py-2">
          <details>
            <summary className={summaryStyle}>
              Rewards{" "}
              <span className="ml-2 text-sm font-normal text-neutral-600">
                {confidencePointsTotal.toLocaleString()} points
              </span>
            </summary>
            <dl className="grid grid-cols-2 gap-4 py-3">
              <div>
                <dt className="text-sm text-neutral-600">Confidence Points</dt>
                <dd className="mt-1 text-2xl font-semibold text-neutral-900">
                  {confidencePointsTotal.toLocaleString()}
                </dd>
              </div>
              <div>
                <dt className="text-sm text-neutral-600">Your streaks</dt>
                <dd className="mt-1 text-neutral-900">
                  {currentStreakDays}-day streak{" "}
                  <span className="block text-sm text-neutral-600">
                    Best: {longestStreakDays} days
                  </span>
                </dd>
              </div>
            </dl>
            <p className="pb-3 text-sm text-neutral-600">
              Points reward saving, reviewing and writing. They are not a
              language proficiency score.
            </p>
          </details>
        </section>
      </div>
      {achievementsResponse ? (
        <Achievements achievements={achievementsResponse.data} />
      ) : (
        <p role="status" className="mt-6 text-neutral-700">
          Your milestones could not load. Your other progress is still
          available.
        </p>
      )}
    </PageContainer>
  );
}

/** Proportions of the counts shown beside it; decorative, so hidden from AT. */
function StageBar({
  stages,
}: {
  stages: ReadonlyArray<{ label: string; count: number; swatch: string }>;
}) {
  const total = stages.reduce((sum, stage) => sum + stage.count, 0);
  return (
    <div
      aria-hidden="true"
      className="mt-5 flex h-2.5 w-full overflow-hidden rounded-full bg-neutral-200"
    >
      {total > 0
        ? stages.map((stage) =>
            stage.count > 0 ? (
              <span
                key={stage.label}
                className={`h-full ${stage.swatch} [&+&]:border-l-2 [&+&]:border-white`}
                style={{ width: `${(stage.count / total) * 100}%` }}
              />
            ) : null,
          )
        : null}
    </div>
  );
}
