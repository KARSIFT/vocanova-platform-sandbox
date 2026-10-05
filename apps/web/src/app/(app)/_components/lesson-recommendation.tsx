import Link from "next/link";
import type { LessonRecommendationResponse } from "@vocanova/api-client";

const action =
  "mt-4 inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const link =
  "inline-flex min-h-11 items-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

export function RecommendedLesson({
  data,
  className,
  compact = false,
  browseHref = "/discover",
  browseLabel = "All lessons",
}: {
  data: LessonRecommendationResponse | null;
  className?: string;
  compact?: boolean;
  browseHref?: string;
  browseLabel?: string;
}) {
  const recommendation = data?.recommendation;
  const lesson = recommendation?.lesson;
  const unavailable = !data || data.status === "content_unavailable";
  return (
    <section
      aria-labelledby="recommended-lesson-heading"
      className={`${compact ? "border-b border-neutral-200 pb-5" : "rounded-2xl border border-neutral-200 bg-white p-5 sm:p-6"} ${className ?? ""}`}
    >
      <h2
        id="recommended-lesson-heading"
        className="text-lg font-bold tracking-tight text-neutral-900"
      >
        {recommendation?.reason === "resume"
          ? "Pick up your lesson"
          : "Your next lesson"}
      </h2>
      {lesson && recommendation ? (
        <>
          <p className="mt-3 text-sm font-semibold text-primary-700">
            {lesson.situationTitle}
          </p>
          <h3 className="mt-1 text-xl font-bold text-neutral-900">
            {lesson.title}
          </h3>
          <p className="mt-2 text-neutral-700">{lesson.description}</p>
          <p className="mt-3 text-sm text-neutral-600">
            {recommendation.reason === "resume"
              ? `${lesson.completedSteps} of ${lesson.stepCount} steps complete.`
              : `${recommendation.usefulTargetCount} ${recommendation.usefulTargetCount === 1 ? "meaning" : "meanings"} to practise${recommendation.matchesFocus ? " in your chosen focus" : ""}.`}
          </p>
          <Link
            href={`/learn/${encodeURIComponent(lesson.key)}`}
            className={
              compact
                ? `${link} mt-3 rounded-xl border border-primary-200 px-5 py-2 hover:bg-primary-50`
                : action
            }
          >
            {recommendation.reason === "resume"
              ? "Continue lesson"
              : "Start lesson"}
            <span className="sr-only">: {lesson.title}</span>
          </Link>
        </>
      ) : (
        <p
          role={unavailable ? "status" : undefined}
          className="mt-3 text-neutral-700"
        >
          {unavailable
            ? "Your lesson recommendation is unavailable right now. You can still browse lessons or review your words."
            : data.status === "no_unfinished_lessons"
              ? "You have completed the current guided lessons. Revisit their words or try a focused practice."
              : "You have marked the remaining lesson meanings as known or reached their mastered review stage. Revisit any lesson whenever you like."}
        </p>
      )}
      <div className="mt-3 flex flex-wrap gap-4">
        <Link href={browseHref} className={link}>
          {browseLabel}
        </Link>
        {!lesson && (
          <Link href="/practice" className={link}>
            Choose a practice
          </Link>
        )}
      </div>
    </section>
  );
}
