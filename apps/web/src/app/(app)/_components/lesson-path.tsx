import Link from "next/link";
import type { LessonSummary, Situation } from "@vocanova/api-client";
import { getSituationOutcome } from "./situation-presentation";

export function LessonPath({
  lessons,
  situations = [],
  compact = false,
}: {
  lessons: LessonSummary[];
  situations?: Situation[];
  compact?: boolean;
}) {
  const groups = new Map<
    string,
    {
      title: string;
      description?: string;
      browsable: boolean;
      lessons: LessonSummary[];
    }
  >();
  for (const situation of situations) {
    groups.set(situation.slug, {
      title: situation.title,
      description:
        getSituationOutcome(situation.slug) ?? situation.shortDescription,
      browsable: true,
      lessons: [],
    });
  }
  for (const lesson of lessons) {
    const group = groups.get(lesson.situationSlug) ?? {
      title: lesson.situationTitle,
      description: getSituationOutcome(lesson.situationSlug),
      browsable: false,
      lessons: [],
    };
    group.lessons.push(lesson);
    groups.set(lesson.situationSlug, group);
  }
  if (!groups.size) return null;
  const ordered = [...lessons].sort(
    (a, b) =>
      Number(b.status === "in_progress") - Number(a.status === "in_progress"),
  );
  if (compact) {
    const next =
      ordered.find((lesson) => lesson.status !== "completed") ?? ordered[0];
    return next ? (
      <ol>
        <LessonRow lesson={next} />
      </ol>
    ) : null;
  }
  return (
    <section aria-label="Guided lessons">
      <div className="mb-5 flex flex-wrap items-end justify-between gap-3">
        <h2 className="text-2xl font-bold tracking-tight text-neutral-900">
          Explore by situation
        </h2>
        {lessons.length > 0 && (
          <p className="text-sm text-neutral-600">
            {lessons.filter((lesson) => lesson.status === "completed").length}{" "}
            of {lessons.length} lessons completed. Choose any situation.
          </p>
        )}
      </div>
      <div className="grid items-start gap-3 md:grid-cols-2 lg:grid-cols-3">
        {[...groups.entries()].map(([slug, group]) => {
          const complete = group.lessons.filter(
            (lesson) => lesson.status === "completed",
          ).length;
          const continuing = group.lessons.some(
            (lesson) => lesson.status === "in_progress",
          );
          return (
            <details
              key={slug}
              open={continuing}
              className={`rounded-xl border bg-white ${continuing ? "border-primary-300" : "border-neutral-200"}`}
            >
              <summary className="min-h-11 cursor-pointer rounded-xl p-4 marker:text-primary-700 hover:bg-primary-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
                <span className="font-semibold text-neutral-900">
                  {group.title}
                </span>
                {group.description && (
                  <span className="mt-2 block text-sm leading-relaxed font-normal text-neutral-600">
                    {group.description}
                  </span>
                )}
                <span className="mt-3 block text-sm font-medium text-primary-700">
                  {continuing
                    ? "Lesson in progress"
                    : group.lessons.length
                      ? `${complete} of ${group.lessons.length} completed`
                      : "Explore words and examples"}
                </span>
              </summary>
              <div className="border-t border-neutral-200 px-4 pb-3">
                <ol className="divide-y divide-neutral-200">
                  {group.lessons.map((lesson) => (
                    <LessonRow key={lesson.key} lesson={lesson} />
                  ))}
                </ol>
                {group.browsable && (
                  <Link
                    href={`/discover/${encodeURIComponent(slug)}`}
                    className="mt-2 inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
                  >
                    Explore words
                    <span className="sr-only"> in {group.title}</span>
                  </Link>
                )}
              </div>
            </details>
          );
        })}
      </div>
    </section>
  );
}

function LessonRow({ lesson }: { lesson: LessonSummary }) {
  return (
    <li className="py-3">
      <h3 className="font-semibold leading-relaxed text-neutral-900">
        {lesson.title}
      </h3>
      <p className="mt-1 text-sm text-neutral-600">
        {lesson.status === "completed"
          ? "Completed"
          : lesson.status === "in_progress"
            ? `${lesson.completedSteps} of ${lesson.stepCount} steps`
            : `${lesson.wordCount} words`}
      </p>
      {lesson.status === "in_progress" && (
        <progress
          className="mt-2 h-2 w-full accent-primary-700"
          aria-label={`${lesson.title} progress`}
          value={lesson.completedSteps}
          max={lesson.stepCount}
        />
      )}
      <Link
        href={`/learn/${encodeURIComponent(lesson.key)}`}
        className="mt-1 inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
      >
        {lesson.status === "in_progress"
          ? "Continue lesson"
          : lesson.status === "completed"
            ? "Revisit words"
            : "Start lesson"}
        <span className="sr-only">: {lesson.title}</span>
      </Link>
    </li>
  );
}
