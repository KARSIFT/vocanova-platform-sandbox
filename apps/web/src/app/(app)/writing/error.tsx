"use client";
import Link from "@/ui/link";
import { useRouteErrorRetry } from "@/ui/use-route-error-retry";
import { PageContainer, Surface } from "@/ui/surface";
import {
  primaryAction,
  textLink,
} from "../practice/_components/practice-styles";
export default function WritingError({ reset }: { reset: () => void }) {
  const { retry, isRetrying } = useRouteErrorRetry(reset);
  return (
    <PageContainer>
      <Surface>
        <h1 className="text-2xl font-bold text-neutral-900">
          Writing practice could not load
        </h1>
        <p role="alert" className="my-4 text-neutral-700">
          Try again to load your topics and current saved meanings.
        </p>
        <button
          type="button"
          className={primaryAction}
          onClick={retry}
          disabled={isRetrying}
          aria-busy={isRetrying}
        >
          {isRetrying ? "Trying again..." : "Try again"}
        </button>
        <Link href="/practice" className={`${textLink} ml-4`}>
          Back to Practice
        </Link>
      </Surface>
    </PageContainer>
  );
}
