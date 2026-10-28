export default function HomeLoading() {
  return (
    <div
      aria-busy="true"
      aria-label="Loading home"
      className="mx-auto grid w-full max-w-[72rem] gap-5 px-[var(--spacing-md)] py-4 motion-safe:animate-pulse sm:px-[var(--spacing-xl)] lg:py-8"
    >
      <div aria-hidden="true">
        <div className="h-9 w-32 rounded-lg bg-neutral-200" />
      </div>
      <div
        aria-hidden="true"
        className="h-56 rounded-[1.5rem] bg-primary-100"
      />
      <div
        aria-hidden="true"
        className="h-48 rounded-[1.25rem] border border-neutral-200 bg-white"
      />
      <div
        aria-hidden="true"
        className="h-40 rounded-2xl border border-neutral-200 bg-white"
      />
    </div>
  );
}
