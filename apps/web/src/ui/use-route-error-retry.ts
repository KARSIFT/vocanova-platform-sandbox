"use client";

import { useRouter } from "next/navigation";
import { useRef, useState, useTransition } from "react";

import { isSameAppChunkLoadError } from "./route-chunk-error";

/** Clear the failed server payload as well as the client error boundary. */
export function useRouteErrorRetry(reset: () => void, error: unknown) {
  const router = useRouter();
  const [isRetrying, startTransition] = useTransition();
  const reloadStarted = useRef(false);
  const [isReloading, setIsReloading] = useState(false);
  function retry() {
    if (reloadStarted.current) return;
    // Only an explicit Retry may discard the failed runtime/chunk cache.
    if (isSameAppChunkLoadError(error, window.location.origin)) {
      reloadStarted.current = true;
      setIsReloading(true);
      window.location.reload();
      return;
    }
    startTransition(() => {
      router.refresh();
      reset();
    });
  }
  return { retry, isRetrying: isRetrying || isReloading };
}
