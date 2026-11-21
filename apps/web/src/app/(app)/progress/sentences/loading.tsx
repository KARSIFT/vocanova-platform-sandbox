import { PageContainer } from "@/ui/surface";

export default function SentenceHistoryLoading() {
  return (
    <PageContainer>
      <span role="status" className="sr-only">
        Loading sentence history
      </span>
      <div
        aria-hidden="true"
        aria-busy="true"
        className="motion-safe:animate-pulse"
      >
        <div className="h-11 w-1/3 max-w-[10rem] rounded bg-neutral-100" />
        <div className="mt-2 h-4 w-1/3 max-w-[10rem] rounded bg-neutral-200" />
        <div className="mt-2 h-8 w-2/3 max-w-[17rem] rounded bg-neutral-200" />
        <div className="vn-list mt-6 space-y-3 p-4 sm:p-6">
          <div className="h-5 w-1/3 rounded bg-neutral-200" />
          <div className="h-20 rounded-xl bg-neutral-100" />
        </div>
      </div>
    </PageContainer>
  );
}
