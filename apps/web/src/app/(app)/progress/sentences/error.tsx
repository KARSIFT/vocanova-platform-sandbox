"use client";

import { ActionLink, Button } from "@/ui/controls";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";

import { PageContainer, PageHeader } from "@/ui/surface";

export default function SentenceHistoryError({
  reset,
  error,
}: Readonly<{ error: Error & { digest?: string }; reset: () => void }>) {
  const { retry: handleRetry, isRetrying } = useRouteErrorRetry(reset, error);

  return (
    <PageContainer>
      <PageHeader
        title="We couldn't load your sentence history"
        lede="We couldn't retrieve your saved feedback. Try again or return to Progress."
      />
      <div className="mt-[var(--spacing-md)] flex flex-wrap gap-[var(--spacing-sm)]">
        <Button
          variant="primary"
          type="button"
          onClick={handleRetry}
          disabled={isRetrying}
          aria-busy={isRetrying}
          className="text-white"
        >
          {isRetrying ? "Trying again..." : "Try again"}
        </Button>
        <ActionLink variant="secondary" href="/progress">
          Back to Progress
        </ActionLink>
      </div>
    </PageContainer>
  );
}
