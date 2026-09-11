"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useTransition } from "react";

import { PageContainer, Surface } from "@/ui/surface";

export default function SentenceHistoryError({
  reset,
}: Readonly<{ error: Error & { digest?: string }; reset: () => void }>) {
  const router = useRouter();
  const [isRetrying, startTransition] = useTransition();

  function handleRetry() {
    startTransition(() => {
      router.refresh();
      reset();
    });
  }

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
          <button
            type="button"
            onClick={handleRetry}
            disabled={isRetrying}
            aria-busy={isRetrying}
            className="inline-flex min-h-11 items-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] font-semibold text-white hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            {isRetrying ? "Trying again..." : "Try again"}
          </button>
          <Link
            href="/progress"
            className="inline-flex min-h-11 items-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Back to Progress
          </Link>
        </div>
      </Surface>
    </PageContainer>
  );
}
