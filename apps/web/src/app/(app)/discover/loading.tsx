export default function DiscoverLoading() {
  return (
    <div
      aria-busy="true"
      aria-label="Loading Journey"
      className="animate-pulse p-[var(--spacing-lg)]"
    >
      <div className="h-7 w-1/3 rounded bg-neutral-200" />
      <div className="mt-[var(--spacing-xs)] h-4 w-2/3 rounded bg-neutral-200" />

      <div className="mt-[var(--spacing-lg)] space-y-[var(--spacing-md)]">
        {Array.from({ length: 3 }).map((_, index) => (
          <div
            key={index}
            className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
          >
            <div className="h-5 w-1/3 rounded bg-neutral-200" />
            <div className="mt-[var(--spacing-xs)] h-4 w-2/3 rounded bg-neutral-200" />
          </div>
        ))}
      </div>
    </div>
  );
}
