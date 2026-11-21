import { StatusMessage } from "@/ui/status-message";
import { PageContainer, Surface } from "@/ui/surface";

export default function ProgressLoading() {
  return (
    <PageContainer width="standard">
      <section aria-busy="true" aria-labelledby="progress-loading-status">
        <StatusMessage
          as="p"
          id="progress-loading-status"
          role="status"
          className="sr-only"
        >
          Loading progress
        </StatusMessage>
        <div aria-hidden="true" className="motion-safe:animate-pulse">
          <div className="h-8 w-1/3 max-w-[12rem] rounded bg-neutral-200" />
          <div className="mt-3 h-4 w-2/3 max-w-[18rem] rounded bg-neutral-200" />
          <Surface className="mt-6">
            <div className="h-5 w-1/2 max-w-[12rem] rounded bg-neutral-200" />
            <div className="mt-6 grid grid-cols-2 gap-4">
              {["w-1/2", "w-2/3"].map((width) => (
                <div key={width} className="space-y-2">
                  <div className={`h-4 rounded bg-neutral-200 ${width}`} />
                  <div className="h-8 w-1/3 rounded bg-neutral-200" />
                </div>
              ))}
            </div>
            <div className="mt-6 h-2 w-full rounded-full bg-neutral-200" />
            <div className="mt-4 h-12 w-full rounded-lg bg-neutral-100" />
          </Surface>
          <div className="mt-8 h-5 w-2/5 max-w-[12rem] rounded bg-neutral-200" />
          <div className="activity-days">
            {Array.from({ length: 4 }, (_, index) => (
              <div key={index} className="min-h-12 rounded-lg bg-neutral-100" />
            ))}
          </div>
        </div>
      </section>
    </PageContainer>
  );
}
