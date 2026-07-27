export default function HomeLoading() {
  return (
    <div
      aria-busy="true"
      aria-label="Loading home"
      className="animate-pulse p-[var(--spacing-lg)]"
    >
      <section className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm">
        <div className="h-6 w-1/3 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-sm)] h-4 w-1/2 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-xs)] h-4 w-2/3 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-md)] h-[var(--spacing-sm)] w-full rounded-full bg-neutral-200" />
      </section>

      <section className="mt-[var(--spacing-lg)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm">
        <div className="h-5 w-1/3 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-sm)] h-4 w-full rounded bg-neutral-200" />
      </section>

      <div className="mt-[var(--spacing-lg)] h-4 w-1/4 rounded bg-neutral-200" />
      <div className="mt-[var(--spacing-sm)] h-4 w-1/4 rounded bg-neutral-200" />
      <div className="mt-[var(--spacing-lg)] flex gap-[var(--spacing-md)]">
        <div className="h-[var(--spacing-2xl)] w-1/3 rounded bg-neutral-200" />
        <div className="h-[var(--spacing-2xl)] w-1/3 rounded bg-neutral-200" />
      </div>
    </div>
  );
}
