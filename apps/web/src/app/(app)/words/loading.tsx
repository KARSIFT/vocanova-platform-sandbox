export default function SavedWordsLoading() {
  return (
    <section
      aria-busy="true"
      aria-labelledby="saved-words-loading-status"
      className="animate-pulse p-[var(--spacing-lg)]"
    >
      <p id="saved-words-loading-status" role="status" className="sr-only">
        Loading saved vocabulary
      </p>
      <div className="h-7 w-1/2 rounded bg-neutral-200" />
      <div className="mt-[var(--spacing-xs)] h-4 w-3/4 rounded bg-neutral-200" />
      <div className="mt-[var(--spacing-lg)] grid grid-cols-1 gap-[var(--spacing-md)] md:grid-cols-2">
        {Array.from({ length: 4 }).map((_, index) => (
          <div
            key={index}
            className="h-40 rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
          />
        ))}
      </div>
    </section>
  );
}
