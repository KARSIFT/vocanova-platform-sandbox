"use client";

import { StatusMessage } from "@/ui/status-message";
import { Button } from "@/ui/controls";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";
import { PageContainer, Surface } from "@/ui/surface";
import {
  primaryAction,
  textLink,
} from "../practice/_components/practice-styles";
export default function StoryError({
  reset,
  error,
}: {
  reset: () => void;
  error: Error & { digest?: string };
}) {
  const { retry, isRetrying } = useRouteErrorRetry(reset, error);
  return (
    <PageContainer className="max-w-[44rem]">
      <Surface>
        <h1 className="text-2xl font-bold text-neutral-900">
          Your story could not load
        </h1>
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          className="my-4 text-neutral-700"
        >
          Try again to load your saved progress.
        </StatusMessage>
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
        {/* eslint-disable-next-line @next/next/no-html-link-for-pages -- Recovery needs a fresh document, not the failed cached RSC payload. */}
        <a href="/stories" className={`${textLink} ml-4`}>
          Back to Stories
        </a>
      </Surface>
    </PageContainer>
  );
}
