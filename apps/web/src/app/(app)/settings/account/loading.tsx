export default function SettingsAccountLoading() {
  return (
    <div
      aria-busy="true"
      aria-label="Loading account settings"
      className="animate-pulse p-[var(--spacing-lg)]"
    >
      <div className="mb-[var(--spacing-md)] flex items-center justify-between">
        <div className="h-7 w-1/4 rounded bg-neutral-200" />
        <div className="h-5 w-1/6 rounded bg-neutral-200" />
      </div>
      <div className="h-4 w-3/4 rounded bg-neutral-200" />
      <div className="mt-[var(--spacing-lg)] space-y-[var(--spacing-lg)]">
        <div className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)]">
          <div className="h-5 w-1/4 rounded bg-neutral-200" />
          <div className="mt-[var(--spacing-sm)] h-4 w-2/3 rounded bg-neutral-200" />
          <div className="mt-[var(--spacing-md)] h-[var(--spacing-2xl)] w-1/3 rounded bg-neutral-200" />
        </div>
        <div className="rounded-md border border-red-200 bg-red-50 p-[var(--spacing-md)]">
          <div className="h-5 w-1/4 rounded bg-neutral-200" />
          <div className="mt-[var(--spacing-sm)] h-4 w-3/4 rounded bg-neutral-200" />
          <div className="mt-[var(--spacing-md)] h-[var(--spacing-2xl)] w-1/2 rounded bg-neutral-200" />
        </div>
      </div>
    </div>
  );
}
