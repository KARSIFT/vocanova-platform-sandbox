"use client";
import { listButton } from "./list-styles";
import type { useListAction } from "./use-list-action";

export function ListActionStatus({
  action,
  reload,
}: {
  action: ReturnType<typeof useListAction>;
  reload: () => void;
}) {
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
          role="alert"
          className="mt-4 rounded-xl border border-secondary-300 bg-secondary-50 p-4 text-neutral-900"
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
