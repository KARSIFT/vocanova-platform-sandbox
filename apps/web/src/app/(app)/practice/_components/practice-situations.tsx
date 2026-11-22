import type { LessonSummary } from "@vocanova/api-client";
import Link from "@/ui/link";
import { PageSection } from "@/ui/surface";

const topics = [
  {
    title: "Work",
    situations: ["work-meeting", "everyday-work", "job-interview"],
  },
  {
    title: "Travel",
    situations: ["airport", "public-transport", "hotel-check-in"],
  },
  { title: "Dining", situations: ["restaurant"] },
  { title: "Shopping", situations: ["shopping", "everyday-payments"] },
  {
    title: "Everyday problems",
    situations: [
      "deliveries",
      "home-and-renting",
      "everyday-services",
      "daily-conversation",
    ],
  },
];

export function PracticeSituations({ lessons }: { lessons: LessonSummary[] }) {
  const cards = topics.flatMap(({ title, situations }) => {
    const matches = situations.flatMap((slug) =>
      lessons.filter((lesson) => lesson.situationSlug === slug),
    );
    const next =
      matches.find((lesson) => lesson.status === "in_progress") ??
      matches.find((lesson) => lesson.status === "not_started") ??
      matches[0];
    return next ? [{ title, lesson: next }] : [];
  });
  if (!cards.length) return null;
  return (
    <PageSection
      title="Practise a situation"
      titleId="practice-situations-heading"
      className="mb-6"
    >
      <ul className="grid gap-2 sm:grid-cols-2">
        {cards.map(({ title, lesson }) => (
          <li key={title}>
            <Link
              href={`/learn/${encodeURIComponent(lesson.key)}`}
              className="block h-full min-h-12 rounded-xl border border-neutral-200 bg-white p-4 hover:border-primary-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              <span className="text-sm font-semibold text-neutral-600">
                {title}
              </span>
              <span className="mt-1 block font-semibold text-neutral-900">
                {lesson.title}
              </span>
              <span className="mt-1 block text-sm text-neutral-600">
                {lesson.status === "in_progress"
                  ? `Resume · ${lesson.completedSteps} of ${lesson.stepCount} steps completed`
                  : lesson.status === "completed"
                    ? "View completed lesson"
                    : `${lesson.wordCount} ${lesson.wordCount === 1 ? "word" : "words"} · ${lesson.stepCount} steps`}
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </PageSection>
  );
}
