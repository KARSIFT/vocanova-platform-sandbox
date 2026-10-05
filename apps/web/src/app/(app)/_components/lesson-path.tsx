import Link from "next/link";
import type { LessonSummary } from "@vocanova/api-client";

export function LessonPath({
  lessons,
  compact = false,
}: {
  lessons: LessonSummary[];
  compact?: boolean;
}) {
  if (!lessons.length) return null;
  const ordered = [...lessons].sort(
    (a, b) =>
      Number(b.status === "in_progress") - Number(a.status === "in_progress"),
  );
  const next =
    ordered.find((item) => item.status !== "completed") ?? ordered[0];
  if (!next) return null;
  const visible = compact ? [next] : ordered;
  const grouped = !compact && visible.length > 8;
  const groups = new Map<string, { title: string; lessons: LessonSummary[] }>();
  for (const lesson of visible) {
    const group = groups.get(lesson.situationSlug) ?? {
      title: lesson.situationTitle,
      lessons: [],
    };
    group.lessons.push(lesson);
    groups.set(lesson.situationSlug, group);
  }
  return (
    <section
      aria-label={compact ? "Your next lesson" : "Guided lessons"}
      className="my-6"
    >
      <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-neutral-900">
            {compact ? "Your next lesson" : "Your lessons"}
          </h2>
          <p className="mt-1 text-neutral-700">
            Learn three words, then practise them.
          </p>
        </div>
        {compact && (
          <Link
            href="/discover"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700"
          >
            All lessons
          </Link>
        )}
      </div>
      {grouped ? (
        <div className="space-y-3">
          <p className="text-sm text-neutral-600">
            {lessons.filter((lesson) => lesson.status === "completed").length}{" "}
            of {lessons.length} lessons completed. Choose any situation.
          </p>
          {[...groups.entries()].map(([slug, group]) => (
            <details
              key={slug}
              open={slug === next.situationSlug}
              className="rounded-2xl border border-neutral-200 bg-white p-4"
            >
              <summary className="min-h-12 cursor-pointer py-3 font-bold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
                {group.title}
                <span className="ml-3 text-sm font-normal text-neutral-600">
                  {
                    group.lessons.filter(
                      (lesson) => lesson.status === "completed",
                    ).length
                  }{" "}
                  of {group.lessons.length} completed
                </span>
              </summary>
              <ol className="mt-3 grid gap-4 md:grid-cols-2">
                {group.lessons.map((lesson) => (
                  <LessonCard key={lesson.key} lesson={lesson} />
                ))}
              </ol>
            </details>
          ))}
        </div>
      ) : (
        <ol className={compact ? "" : "grid gap-4 md:grid-cols-2"}>
          {visible.map((lesson) => (
            <LessonCard key={lesson.key} lesson={lesson} />
          ))}
        </ol>
      )}
    </section>
  );
}

function LessonCard({ lesson }: { lesson: LessonSummary }) {
  return (
    <li
      className={`rounded-2xl border p-5 ${lesson.status === "in_progress" ? "border-primary-300 bg-primary-50" : "border-neutral-200 bg-white"}`}
    >
      <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
        <span className="font-semibold text-primary-700">
          {lesson.situationTitle}
        </span>
        <span className="text-neutral-600">
          {lesson.status === "completed"
            ? "Completed"
            : lesson.status === "in_progress"
              ? `${lesson.completedSteps} of ${lesson.stepCount} steps`
              : `${lesson.wordCount} words`}
        </span>
      </div>
      <h3 className="mt-3 text-xl font-bold text-neutral-900">
        {lesson.title}
      </h3>
      <p className="mt-2 text-neutral-700">{lesson.description}</p>
      {lesson.status === "in_progress" && (
        <progress
          className="mt-4 h-2 w-full accent-primary-700"
          aria-label={`${lesson.title} progress`}
          value={lesson.completedSteps}
          max={lesson.stepCount}
        />
      )}
      <Link
        href={`/learn/${encodeURIComponent(lesson.key)}`}
        className="mt-4 inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
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
