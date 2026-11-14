"use client";

import { ActionLink, Button } from "@/ui/controls";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";

import { PageContainer, Surface } from "@/ui/surface";

export default function SentenceHistoryError({
  reset,
  error,
}: Readonly<{ error: Error & { digest?: string }; reset: () => void }>) {
  const { retry: handleRetry, isRetrying } = useRouteErrorRetry(reset, error);

  return (
    <PageContainer>
      <Surface aria-labelledby="sentence-history-error-heading">
        <h1
          id="sentence-history-error-heading"
          className="text-xl font-semibold text-neutral-900"
        >
          We couldn&apos;t load your sentence history
        </h1>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          We couldn&apos;t retrieve your saved feedback. Try again or return to
          Progress.
        </p>
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
      </Surface>
    </PageContainer>
  );
}
