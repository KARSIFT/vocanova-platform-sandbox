import { StatusMessage } from "@/ui/status-message";
import { PageContainer } from "@/ui/surface";

export default function StoriesLoading() {
  return (
    <PageContainer>
      <StatusMessage as="p" role="status" className="text-neutral-700">
        Loading your story…
      </StatusMessage>
      <div
        aria-hidden="true"
        className="mt-5 space-y-3 motion-safe:animate-pulse"
      >
        <div className="h-7 w-2/3 rounded bg-neutral-200" />
        <div className="h-24 rounded-xl border border-neutral-200 bg-white" />
      </div>
    </PageContainer>
  );
}
