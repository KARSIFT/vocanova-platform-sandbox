"use client";

import Link from "@/ui/link";
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
        <button
          type="button"
          onClick={retry}
          disabled={isRetrying}
          aria-busy={isRetrying}
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          {isRetrying ? "Trying again..." : "Try again"}
        </button>
        <Link
          href="/discover"
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Go to Journey
        </Link>
      </div>
    </div>
  );
}
