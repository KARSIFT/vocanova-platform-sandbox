"use client";

import { useRouter } from "next/navigation";
import { useTransition } from "react";

/** Clear the failed server payload as well as the client error boundary. */
export function useRouteErrorRetry(reset: () => void) {
  const router = useRouter();
  const [isRetrying, startTransition] = useTransition();
  function retry() {
    startTransition(() => {
      router.refresh();
      reset();
    });
  }
  return { retry, isRetrying };
}
