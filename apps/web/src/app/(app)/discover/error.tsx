"use client";

import { ActionLink, Button } from "@/ui/controls";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";

interface DiscoverErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function DiscoverError(props: DiscoverErrorProps) {
  const { reset, error } = props;
  const { retry, isRetrying } = useRouteErrorRetry(reset, error);
  return (
    <div className="p-[var(--spacing-lg)]">
      <h1 className="text-xl font-semibold text-neutral-900">
        Something went wrong
      </h1>
      <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
        We couldn&apos;t load Discover. Please try again.
      </p>
      <div className="mt-[var(--spacing-lg)] flex flex-wrap gap-[var(--spacing-md)]">
        <Button
          variant="primary"
          onClick={retry}
          disabled={isRetrying}
          aria-busy={isRetrying}
          className="min-w-[var(--spacing-2xl)] text-base text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)]"
        >
          {isRetrying ? "Trying again..." : "Try again"}
        </Button>
        <ActionLink
          variant="secondary"
          href="/home"
          className="min-w-[var(--spacing-2xl)] text-base transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)]"
        >
          Back to Home
        </ActionLink>
      </div>
    </div>
  );
}
