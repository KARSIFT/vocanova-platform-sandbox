import { StatusMessage } from "@/ui/status-message";
import { PageContainer } from "@/ui/surface";

export default function SettingsAccountLoading() {
  return (
    <PageContainer width="reading">
      <section
        aria-busy="true"
        aria-labelledby="account-loading-status"
        className="motion-safe:animate-pulse"
      >
        <StatusMessage
          as="p"
          tone="info"
          id="account-loading-status"
          role="status"
          className="sr-only"
        >
          Loading account settings
        </StatusMessage>
        <div className="h-11 w-1/3 max-w-[9rem] rounded bg-neutral-100" />
        <div className="mt-2 mb-3 h-8 w-1/2 max-w-[10rem] rounded bg-neutral-200" />
        <div className="h-4 w-3/4 rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-lg)] space-y-[var(--spacing-lg)]">
          <div className="rounded-[1.25rem] border border-neutral-200 bg-white p-4 sm:p-6">
            <div className="h-5 w-1/4 rounded bg-neutral-200" />
            <div className="mt-[var(--spacing-sm)] h-4 w-2/3 rounded bg-neutral-200" />
            <div className="mt-[var(--spacing-md)] h-[var(--spacing-2xl)] w-1/3 rounded bg-neutral-200" />
          </div>
          <div className="rounded-[1.25rem] border border-neutral-200 bg-white p-4 sm:p-6">
            <div className="h-5 w-1/4 rounded bg-neutral-200" />
            <div className="mt-[var(--spacing-sm)] h-4 w-3/4 rounded bg-neutral-200" />
            <div className="mt-[var(--spacing-md)] h-[var(--spacing-2xl)] w-1/2 rounded bg-neutral-200" />
          </div>
        </div>
      </section>
    </PageContainer>
  );
}
