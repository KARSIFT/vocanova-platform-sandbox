export default function ReviewsLoading() {
  return (
    <section
      aria-busy="true"
      aria-labelledby="reviews-loading-status"
      className="animate-pulse p-[var(--spacing-lg)]"
    >
      <p id="reviews-loading-status" role="status" className="sr-only">
        Loading reviews
      </p>
      <div className="flex items-center justify-between">
        <div className="h-7 w-1/4 rounded bg-neutral-200" />
        <div className="h-4 w-1/4 rounded bg-neutral-200" />
      </div>

      <section className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm">
        <div className="mx-auto h-5 w-1/4 rounded bg-neutral-200" />
        <div className="mx-auto mt-[var(--spacing-sm)] h-9 w-1/2 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-lg)] space-y-[var(--spacing-sm)]">
          {Array.from({ length: 4 }).map((_, index) => (
            <div key={index} className="h-12 rounded bg-neutral-200" />
          ))}
        </div>
      </section>
    </section>
  );
}
