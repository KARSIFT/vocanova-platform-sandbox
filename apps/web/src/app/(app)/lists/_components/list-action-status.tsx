"use client";
import { useEffect, useRef } from "react";
import { listButton } from "./list-styles";
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
        <p role="status" className="mt-3 text-neutral-700">
          Checking your list…
        </p>
      )}
      {action.message && (
        <p role="status" className="mt-3 text-primary-700">
          {action.message}
        </p>
      )}
      {action.error && (
        <div
          ref={alertRef}
          role="alert"
          tabIndex={-1}
          className="mt-4 scroll-my-24 rounded-xl border border-secondary-300 bg-secondary-50 p-4 text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          <p>{action.error}</p>
          <div className="mt-3 flex flex-wrap gap-3">
            {action.needsRetry && (
              <button
                type="button"
                disabled={action.busy}
                className={listButton}
                onClick={() => void action.run()}
              >
                Retry same change
              </button>
            )}
            <button
              type="button"
              disabled={action.busy}
              className={listButton}
              onClick={reload}
            >
              Load current status
            </button>
          </div>
        </div>
      )}
    </>
  );
}
