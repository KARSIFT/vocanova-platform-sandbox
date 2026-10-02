import Link from "next/link";
import type { AchievementList } from "@vocanova/api-client";

const nextSteps = {
  lessons: { href: "/discover", label: "Choose a lesson" },
  practice: { href: "/practice", label: "Choose a practice" },
  reviews: { href: "/review", label: "View scheduled reviews" },
  writing: { href: "/words", label: "Write with a saved word" },
} as const;

export function Achievements({
  achievements,
}: {
  achievements: AchievementList;
}) {
  const earnedCount = achievements.items.filter((item) => item.earned).length;
  return (
    <section aria-labelledby="achievements-heading" className="mt-8">
      <div className="flex flex-wrap items-end justify-between gap-2">
        <h2
          id="achievements-heading"
          className="text-2xl font-bold text-neutral-900"
        >
          Your milestones
        </h2>
        <p className="text-sm font-semibold text-primary-700">
          {earnedCount} of {achievements.items.length} earned
        </p>
      </div>
      <p className="mt-2 text-neutral-700">
        Small steps worth celebrating. Each milestone shows exactly what you
        did.
      </p>
      <ul className="mt-4 grid gap-4 sm:grid-cols-2">
        {achievements.items.map((item) => {
          const next = nextSteps[item.category];
          return (
            <li
              key={item.id}
              className={`rounded-2xl border p-5 ${item.earned ? "border-primary-300 bg-primary-50" : "border-neutral-200 bg-white"}`}
            >
              <div className="flex items-center gap-3">
                <span
                  aria-hidden="true"
                  className={`inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-full ${item.earned ? "bg-primary-700 text-white" : "bg-neutral-100 text-neutral-600"}`}
                >
                  <svg
                    viewBox="0 0 24 24"
                    className="h-6 w-6 fill-none stroke-current stroke-2"
                  >
                    {item.earned ? (
                      <path
                        d="m5 12 4 4L19 6"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    ) : (
                      <circle cx="12" cy="12" r="7" />
                    )}
                  </svg>
                </span>
                <div>
                  <h3 className="text-lg font-bold text-neutral-900">
                    {item.label}
                  </h3>
                  <p className="text-sm font-semibold text-primary-700">
                    {item.earned ? "Earned" : "In progress"}
                    {item.criterion === "unaided_recall"
                      ? " · Recall without help"
                      : " · Participation"}
                  </p>
                </div>
              </div>
              <p className="mt-3 text-sm text-neutral-700">
                {item.description}
              </p>
              {item.earned && item.earnedAt ? (
                <p className="mt-3 text-sm text-neutral-600">
                  Earned{" "}
                  <time dateTime={item.earnedAt}>
                    {new Intl.DateTimeFormat("en", {
                      year: "numeric",
                      month: "short",
                      day: "numeric",
                      timeZone: "UTC",
                    }).format(new Date(item.earnedAt))}
                  </time>
                </p>
              ) : (
                <>
                  <p className="mt-3 text-sm text-neutral-700">
                    {item.current} of {item.target}
                  </p>
                  <progress
                    aria-label={`${item.label} progress`}
                    value={item.current}
                    max={item.target}
                    className="mt-2 h-2 w-full accent-primary-700"
                  />
                  <Link
                    href={next.href}
                    className="mt-2 inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
                  >
                    {next.label}
                    <span className="sr-only">: {item.label}</span>
                  </Link>
                </>
              )}
            </li>
          );
        })}
      </ul>
    </section>
  );
}
