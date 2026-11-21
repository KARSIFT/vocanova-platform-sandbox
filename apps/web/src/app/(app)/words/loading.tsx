import { StatusMessage } from "@/ui/status-message";
import { PageContainer } from "@/ui/surface";

export default function SavedWordsLoading() {
  return (
    <PageContainer width="standard">
      <section aria-busy="true" aria-labelledby="saved-words-loading-status">
        <StatusMessage
          as="p"
          id="saved-words-loading-status"
          role="status"
          className="sr-only"
        >
          Loading saved vocabulary
        </StatusMessage>
        <div aria-hidden="true" className="motion-safe:animate-pulse">
          <div className="h-8 w-2/3 max-w-[15rem] rounded bg-neutral-200" />
          <div className="mt-3 h-4 w-3/4 max-w-[22rem] rounded bg-neutral-200" />
          <div className="vn-list mt-8 space-y-3 p-4 sm:p-5">
            <div className="h-4 w-1/3 rounded bg-neutral-200" />
            <div className="h-12 w-full rounded-xl bg-neutral-100" />
            <div className="h-12 w-2/3 rounded-xl bg-neutral-100" />
          </div>
          <ul className="vn-list mt-6">
            {["w-1/4", "w-1/3", "w-1/5"].map((width) => (
              <li key={width} className="space-y-3 p-4">
                <div className={`h-5 rounded bg-neutral-200 ${width}`} />
                <div className="h-4 w-3/4 rounded bg-neutral-100" />
              </li>
            ))}
          </ul>
        </div>
      </section>
    </PageContainer>
  );
}
