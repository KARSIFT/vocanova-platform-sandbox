"use client";

import { StatusMessage } from "@/ui/status-message";

import { Button } from "@/ui/controls";
import { useEffect, useRef } from "react";
import type { useListAction } from "./use-list-action";

export function ListActionStatus({
  action,
  reload,
}: {
  action: ReturnType<typeof useListAction>;
  reload: () => void;
}) {
  const alertRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (!action.error || !alertRef.current) return;
    alertRef.current.scrollIntoView({ block: "center" });
    alertRef.current.focus({ preventScroll: true });
  }, [action.error]);
  return (
    <>
      {action.busy && (
        <StatusMessage as="p" tone="info" role="status" className="mt-3">
          Checking your list…
        </StatusMessage>
      )}
      {action.message && (
        <StatusMessage as="p" tone="info" role="status" className="mt-3">
          {action.message}
        </StatusMessage>
      )}
      {action.error && (
        <StatusMessage
          tone="error"
          ref={alertRef}
          role="alert"
          tabIndex={-1}
          className="mt-4 scroll-my-24 rounded-xl border border-secondary-300 bg-secondary-50 p-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          <p>{action.error}</p>
          <div className="mt-3 flex flex-wrap gap-3">
            {action.needsRetry && (
              <Button
                type="button"
                disabled={action.busy}
                variant="secondary"
                onClick={() => void action.run()}
              >
                Retry same change
              </Button>
            )}
            <Button
              type="button"
              disabled={action.busy}
              variant="secondary"
              onClick={reload}
            >
              Load current status
            </Button>
          </div>
        </StatusMessage>
      )}
    </>
  );
}
