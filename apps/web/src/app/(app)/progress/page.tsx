// Placeholder local state for VOC-020 static UI; replace with real API wiring in a follow-up package.
const MOCK_PROGRESS_STATE = {
  confidencePointsTotal: 1240,
  currentStreakDays: 12,
  longestStreakDays: 30,
  completionHistory: [
    { label: "Mon", completed: true },
    { label: "Tue", completed: true },
    { label: "Wed", completed: false },
    { label: "Thu", completed: true },
    { label: "Fri", completed: true },
    { label: "Sat", completed: false },
    { label: "Sun", completed: true },
  ],
} as const;

export default function ProgressPage() {
  const {
    confidencePointsTotal,
    currentStreakDays,
    longestStreakDays,
    completionHistory,
  } = MOCK_PROGRESS_STATE;

  return (
    <div className="p-[var(--spacing-lg)]">
      <h1 className="text-2xl font-semibold text-neutral-900">Progress</h1>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Every practice session builds your confidence.
      </p>

      <section
        aria-labelledby="confidence-points-heading"
        className="mt-[var(--spacing-lg)] rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-md)] shadow-sm"
      >
        <p
          id="confidence-points-heading"
          className="text-sm font-medium text-primary-900"
        >
          Confidence Points
        </p>
        <p className="mt-[var(--spacing-xs)] text-3xl font-semibold text-primary-900">
          {confidencePointsTotal.toLocaleString()}
        </p>
      </section>

      <section
        aria-labelledby="streak-heading"
        className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h2
          id="streak-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          Your streaks
        </h2>
        <p className="mt-[var(--spacing-sm)] text-base text-neutral-800">
          {currentStreakDays}-day streak
        </p>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          Longest streak: {longestStreakDays} days
        </p>
      </section>

      <section
        aria-labelledby="completion-history-heading"
        className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h2
          id="completion-history-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          This week
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          Keep showing up — each day counts.
        </p>
        <ul className="mt-[var(--spacing-md)] grid grid-cols-7 gap-[var(--spacing-xs)]">
          {completionHistory.map((day) => (
            <li
              key={day.label}
              className={`rounded-md p-[var(--spacing-xs)] text-center ${
                day.completed
                  ? "bg-primary-100 text-primary-900"
                  : "bg-neutral-200 text-neutral-800"
              }`}
            >
              <p className="text-sm font-semibold">{day.label}</p>
              <p className="mt-[var(--spacing-xs)] text-xs font-medium">
                {day.completed ? "✓ Done" : "— Rest"}
              </p>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
