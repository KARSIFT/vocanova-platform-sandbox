"use client";
import { StatusMessage } from "@/ui/status-message";
import { Button } from "@/ui/controls";
import Link from "@/ui/link";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";
import { PageContainer, PageHeader } from "@/ui/surface";
import {
  primaryAction,
  textLink,
} from "../practice/_components/practice-styles";
export default function WritingError({
  reset,
  error,
}: {
  reset: () => void;
  error: Error & { digest?: string };
}) {
  const { retry, isRetrying } = useRouteErrorRetry(reset, error);
  return (
    <PageContainer>
      <PageHeader title="Writing practice could not load" />
      <div className="vn-empty py-0">
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          className="mb-4 text-neutral-700"
        >
          Try again to load your topics and current saved meanings.
        </StatusMessage>
        <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
          <Button
            variant="primary"
            type="button"
            className={primaryAction}
            onClick={retry}
            disabled={isRetrying}
            aria-busy={isRetrying}
          >
            {isRetrying ? "Trying again..." : "Try again"}
          </Button>
          <Link href="/practice" className={textLink}>
            Back to Practice
          </Link>
        </div>
      </div>
    </PageContainer>
  );
}
