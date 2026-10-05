export default function PracticeLoading() {
  return (
    <section
      aria-busy="true"
      aria-labelledby="practice-loading-status"
      className="mx-auto w-full max-w-[64rem] p-[var(--spacing-lg)]"
    >
      <p
        id="practice-loading-status"
        role="status"
        className="text-neutral-700"
      >
        Loading practice…
      </p>
      <div
        aria-hidden="true"
        className="mt-6 space-y-4 motion-safe:animate-pulse"
      >
        <div className="h-7 w-2/3 rounded bg-neutral-200" />
        <div className="h-4 w-full rounded bg-neutral-200" />
        <div className="h-28 rounded-xl border border-neutral-200 bg-neutral-50" />
      </div>
    </section>
  );
}
