"use client";

import { Button } from "@/ui/controls";

import Link from "@/ui/link";
import { PageContainer, PageHeader } from "@/ui/surface";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";

interface SettingsAccountErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function SettingsAccountError(props: SettingsAccountErrorProps) {
  const { reset, error } = props;
  const { retry, isRetrying } = useRouteErrorRetry(reset, error);
  return (
    <PageContainer width="reading">
      <PageHeader
        title="We couldn't load your account"
        lede="Please try again. Your account is still safe."
      />
      <div className="flex flex-wrap gap-3">
        <Button
          type="button"
          onClick={retry}
          disabled={isRetrying}
          aria-busy={isRetrying}
          className="min-w-[var(--spacing-2xl)]"
        >
          {isRetrying ? "Trying again..." : "Try again"}
        </Button>
        <Link
          href="/settings"
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Back to Settings
        </Link>
      </div>
    </PageContainer>
  );
}
