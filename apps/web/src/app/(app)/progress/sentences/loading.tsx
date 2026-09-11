import { PageContainer, Surface } from "@/ui/surface";

export default function SentenceHistoryLoading() {
  return (
    <PageContainer>
      <span role="status" className="sr-only">
        Loading sentence history
      </span>
      <Surface aria-busy="true">
        <div className="h-5 w-28 animate-pulse rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-md)] h-8 w-56 animate-pulse rounded bg-neutral-200" />
        <div className="mt-[var(--spacing-lg)] h-32 animate-pulse rounded bg-neutral-100" />
      </Surface>
    </PageContainer>
  );
}
