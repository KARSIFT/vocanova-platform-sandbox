import { StatusMessage } from "@/ui/status-message";
export default function LessonLoading() {
  return (
    <section
      aria-busy="true"
      aria-labelledby="lesson-loading-status"
      className="mx-auto w-full max-w-[44rem] p-[var(--spacing-lg)]"
    >
      <StatusMessage
        as="p"
        id="lesson-loading-status"
        role="status"
        className="text-neutral-700"
      >
        Loading lesson…
      </StatusMessage>
      <div
        aria-hidden="true"
        className="mt-6 space-y-4 motion-safe:animate-pulse"
      >
        <div className="h-7 w-2/3 rounded bg-neutral-200" />
        <div className="h-4 w-full rounded bg-neutral-200" />
        <div className="h-64 rounded-xl border border-neutral-200 bg-neutral-50" />
      </div>
    </section>
  );
}
