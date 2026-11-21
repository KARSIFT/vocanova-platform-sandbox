"use client";

import { ActionLink, Button } from "@/ui/controls";
import { PageContainer, PageHeader } from "@/ui/surface";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";

interface SavedWordsErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function SavedWordsError({
  reset,
  error,
}: SavedWordsErrorProps) {
  const { retry, isRetrying } = useRouteErrorRetry(reset, error);
  return (
    <PageContainer width="standard">
      <PageHeader
        title="Something went wrong"
        lede="We couldn't load your saved vocabulary. Please try again."
      />
      <div className="flex flex-wrap gap-3">
        <Button
          variant="secondary"
          type="button"
          onClick={retry}
          disabled={isRetrying}
          aria-busy={isRetrying}
        >
          {isRetrying ? "Trying again..." : "Try again"}
        </Button>
        <ActionLink
          variant="secondary"
          href="/discover"
          className="min-w-[var(--spacing-2xl)] text-base"
        >
          Go to Journey
        </ActionLink>
      </div>
    </PageContainer>
  );
}
