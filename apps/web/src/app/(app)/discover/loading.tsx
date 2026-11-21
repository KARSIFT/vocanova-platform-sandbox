import { StatusMessage } from "@/ui/status-message";
import { PageContainer } from "@/ui/surface";

export default function DiscoverLoading() {
  return (
    <PageContainer width="wide">
      <section aria-busy="true" aria-labelledby="journey-loading-status">
        <StatusMessage
          as="p"
          id="journey-loading-status"
          role="status"
          className="sr-only"
        >
          Loading Journey
        </StatusMessage>
        <div aria-hidden="true" className="motion-safe:animate-pulse">
          <div className="h-8 w-1/2 max-w-32 rounded bg-neutral-200" />
          <div className="mt-3 h-4 w-2/3 max-w-72 rounded bg-neutral-200" />
          <div className="learning-surface mt-6 grid grid-cols-4 gap-2 p-4 sm:grid-cols-8 sm:p-6">
            {Array.from({ length: 8 }, (_, index) => (
              <div
                key={index}
                className="aspect-square max-h-24 rounded-lg bg-neutral-100"
              />
            ))}
          </div>
        </div>
      </section>
    </PageContainer>
  );
}
