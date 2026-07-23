import Link from "next/link";

const MOCK_HOME_STATE = {
  missionTargetWords: 10,
  reviewedWordsToday: 3,
  currentStreakDays: 5,
  dueReviewWords: 8,
} as const;

export default function HomePage() {
  const { missionTargetWords, reviewedWordsToday, currentStreakDays, dueReviewWords } =
    MOCK_HOME_STATE;

  const missionProgressPercent = Math.min(
    100,
    Math.round((reviewedWordsToday / missionTargetWords) * 100),
  );

  return (
    <div className="p-[var(--spacing-lg)]">
      {/* Placeholder local state for VOC-019 static UI; replace with real API wiring in follow-up package. */}
      <section
        aria-labelledby="todays-mission-heading"
        className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h1 id="todays-mission-heading" className="text-xl font-semibold text-neutral-900">
          Today&apos;s Mission
        </h1>
        <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
          Review target: {missionTargetWords} words
        </p>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          {reviewedWordsToday} of {missionTargetWords} words reviewed today ({missionProgressPercent}
          %)
        </p>

        <div
          aria-hidden="true"
          className="mt-[var(--spacing-md)] h-[var(--spacing-sm)] w-full rounded-full bg-neutral-200"
        >
          <div
            className="h-full rounded-full bg-primary-600"
            style={{ width: `${missionProgressPercent}%` }}
          />
        </div>
      </section>

      <p className="mt-[var(--spacing-lg)] text-base text-neutral-800">
        {currentStreakDays}-day streak
      </p>

      <p className="mt-[var(--spacing-sm)] text-base text-neutral-800">{dueReviewWords} words due today</p>

      <Link
        href="/discover"
        className="mt-[var(--spacing-lg)] inline-flex min-h-11 min-w-11 items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-out hover:bg-primary-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
      >
        Go to Journey
      </Link>
    </div>
  );
}
