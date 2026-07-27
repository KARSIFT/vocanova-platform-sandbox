export default function ProgressLoading() {
  return (
    <div
      aria-busy="true"
      aria-label="Loading progress"
      className="animate-pulse p-[var(--spacing-lg)]"
    >
      <div className="h-7 w-1/3 rounded bg-neutral-200" />
      <div className="mt-[var(--spacing-xs)] h-4 w-2/3 rounded bg-neutral-200" />

      <section className="mt-[var(--spacing-lg)] rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-md)] shadow-sm">
        <div className="h-4 w-1/3 rounded bg-primary-200" />
        <div className="mt-[var(--spacing-xs)] h-8 w-1/4 rounded bg-primary-200" />
      </section>

      <section className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm">
        <div className="h-5 w-1/3 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-sm)] h-4 w-1/4 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-xs)] h-4 w-1/3 rounded bg-neutral-200" />
      </section>

      <section className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm">
        <div className="h-5 w-1/3 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-xs)] h-4 w-1/4 rounded bg-neutral-200" />
      </section>

      <section className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm">
        <div className="h-5 w-1/3 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-xs)] h-4 w-1/2 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-md)] grid grid-cols-7 gap-[var(--spacing-xs)]">
          {Array.from({ length: 7 }).map((_, index) => (
            <div key={index} className="h-16 rounded-md bg-neutral-200" />
          ))}
        </div>
      </section>
    </div>
  );
}
