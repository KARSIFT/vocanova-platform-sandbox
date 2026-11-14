"use client";

import { ActionLink, Button } from "@/ui/controls";
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
    <div className="p-[var(--spacing-lg)]">
      <h1 className="text-xl font-semibold text-neutral-900">
        Something went wrong
      </h1>
      <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
        We couldn&apos;t load your saved vocabulary. Please try again.
      </p>
      <div className="mt-[var(--spacing-lg)] flex flex-wrap gap-[var(--spacing-md)]">
        <Button
          variant="secondary"
          type="button"
          onClick={retry}
          disabled={isRetrying}
          aria-busy={isRetrying}
          className="min-w-[var(--spacing-2xl)] text-base text-neutral-50"
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
    </div>
  );
}
