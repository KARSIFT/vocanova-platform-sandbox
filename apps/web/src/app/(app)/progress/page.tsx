import Link from "@/ui/link";
import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, Surface } from "@/ui/surface";
import { getCompletionDayView } from "./completion-day-view";
import { Achievements } from "../_components/achievements";

// A count the server did not report is shown as unknown, never as zero.
function formatCount(value: unknown): string {
  return typeof value === "number" ? value.toLocaleString() : "—";
}

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
            What you have learned, the days you practised and the milestones you
            reached, as recorded.
          </p>
        </div>
        <Link href="/words" className="vn-chip-link">
          Saved words
        </Link>
      </div>

      {knowledge ? (
        <Surface className="mt-6" aria-labelledby="vocabulary-map-heading">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 id="vocabulary-map-heading" className="section-title">
              Your saved words
            </h2>
            <Link href="/discover" className="vn-text-link">
              Open your word map
            </Link>
          </div>
          {/* Two separate records: words saved for review, and words the
              learner said they already know. Neither is a proficiency score. */}
          <dl className="progress-totals">
            <div>
              <dt>Saved to learn</dt>
              <dd>{formatCount(knowledge.saved)}</dd>
            </div>
            <div>
              <dt>Marked already known</dt>
              <dd>{formatCount(knowledge.selfReportedKnown)}</dd>
            </div>
          </dl>
          <StageBar stages={stages} />
          <ol aria-label="Review stages" className="stage-flow">
            {stages.map(({ label, count, swatch }) => (
              <li key={label}>
                <span className="stage-flow-label">
                  <span
                    aria-hidden="true"
                    className={`size-2.5 rounded-full ${swatch}`}
                  />
                  {label}
                </span>
                <span className="stage-flow-count">{count}</span>
              </li>
            ))}
          </ol>
          <div className="mt-4 flex flex-wrap items-center justify-between gap-2 border-t border-neutral-200 pt-3">
            <p className="text-sm text-neutral-600">
              {knowledge.saved} saved · {knowledge.due} ready for review
            </p>
            {knowledge.due > 0 && (
              <Link href="/review" className="vn-text-link">
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
              className="vn-text-link mt-2"
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
          Your saved words are unavailable right now. Your other progress is
          below.
        </p>
      )}

      <Surface className="mt-4" aria-labelledby="completion-history-heading">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 id="completion-history-heading" className="section-title">
            Recent activity
          </h2>
          <p className="inline-flex min-h-9 flex-wrap items-center gap-x-3 gap-y-1 text-sm text-neutral-700">
            <span className="inline-flex items-center gap-1.5 font-semibold text-neutral-900">
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                className="size-4 fill-none stroke-current stroke-[1.8] text-primary-700"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M13 3c1 5-4 5-3 9 1-1 2-2 2-3 4 2 6 5 5 8a6 6 0 0 1-11-1c-1-5 4-7 7-13Z"
                />
              </svg>
              {currentStreakDays}-day streak
            </span>
            <span>Best: {longestStreakDays} days</span>
          </p>
        </div>
        {historyWithLabels.length > 0 ? (
          <ul className="activity-days">
            {historyWithLabels.map((day) => (
              <li
                key={day.localDate}
                data-status={day.view.label}
                className={`activity-day ${day.view.className}`}
              >
                <time dateTime={day.localDate}>{day.label}</time>
                <p>{day.view.label}</p>
              </li>
            ))}
          </ul>
        ) : (
          <div className="mt-3">
            <p className="text-neutral-700">
              Complete a daily mission to start your history.
            </p>
            <Link href="/home" className="vn-text-link mt-2">
              Go to today&apos;s mission
            </Link>
          </div>
        )}
      </Surface>

      {achievementsResponse ? (
        <Achievements achievements={achievementsResponse.data} />
      ) : (
        <p role="status" className="mt-6 text-neutral-700">
          Your milestones could not load. Your other progress is still
          available.
        </p>
      )}

      <div className="mt-4 divide-y divide-neutral-200 rounded-[1.25rem] border border-neutral-200 bg-white px-[var(--spacing-md)] sm:px-[var(--spacing-lg)]">
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
                      className="vn-text-link"
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
            <Link href="/courses" className="vn-text-link mb-2">
              Browse courses
            </Link>
          </details>
        ) : (
          <p role="status" className="py-4 text-neutral-700">
            Your lesson progress could not load.{" "}
            <Link href="/courses" className="vn-text-link">
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
                <Link href="/writing" className="vn-text-link mt-2">
                  Write your first sentence
                </Link>
              ))}
            <Link href="/progress/sentences" className="vn-text-link mt-2 mb-2">
              View sentence history
            </Link>
          </details>
          {!sentencesResponse && (
            <p role="status" className="pb-3 text-neutral-700">
              Your recent writing could not load.{" "}
              <Link href="/progress/sentences" className="vn-text-link">
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
      className="mt-4 flex h-2 w-full overflow-hidden rounded-full bg-neutral-200"
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
