"use client";

import Link from "next/link";

interface SettingsErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function SettingsError(props: SettingsErrorProps) {
  const { reset } = props;
  return (
    <div className="p-[var(--spacing-lg)]">
      <h1 className="text-xl font-semibold text-neutral-900">
        We couldn&apos;t load your settings
      </h1>
      <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
        Please try again. Your settings are still safe.
      </p>
      <div className="mt-[var(--spacing-lg)] flex flex-wrap gap-[var(--spacing-md)]">
        <button
          type="button"
          onClick={reset}
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Try again
        </button>
        <Link
          href="/home"
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Back to Home
        </Link>
      </div>
    </div>
  );
}
