"use client";

import { StatusMessage } from "@/ui/status-message";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState, useTransition } from "react";
import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError, isSessionExpiredError } from "@/lib/session";

interface MeaningSaveButtonProps {
  meaningId: string;
  wordSlug: string;
  source: "journey" | "search";
  initialSaved: boolean;
  wordText: string;
  shortDefinition: string;
}

interface PendingSave {
  key: string;
  body: { meaningId: string; source: "journey" | "search" };
}

const SESSION_NOT_READY = "Session is not ready. Please refresh the page.";

export function MeaningSaveButton({
  meaningId,
  wordSlug,
  source,
  initialSaved,
  wordText,
  shortDefinition,
}: MeaningSaveButtonProps) {
  const router = useRouter();
  const [saved, setSaved] = useState(initialSaved);
  const [busy, setBusy] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [checkNeeded, setCheckNeeded] = useState(false);
  const [refreshing, startRefresh] = useTransition();
  // A save with an unknown outcome; retried with the same key and body.
  const pending = useRef<PendingSave | null>(null);
  const inFlight = useRef(false);
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(() => {
    if (!refreshing && !inFlight.current && !pending.current)
      setSaved(initialSaved);
  }, [initialSaved, refreshing]);

  function errorText(cause: unknown, fallback: string) {
    return isSessionExpiredError(cause)
      ? handleApiError(cause, fallback)
      : fallback;
  }

  async function readSaved(id: string) {
    const { data } = await createApiClient().getCanonicalWord(wordSlug, {
      cache: "no-store",
    });
    const meaning = data.word.meanings.find((item) => item.id === id);
    if (
      data.word.slug !== wordSlug ||
      !meaning ||
      (meaning.saved && !meaning.userWordId)
    ) {
      throw new Error("Saved state unknown");
    }
    return meaning.saved;
  }

  // Read-only: never sends a new save.
  async function check() {
    setCheckNeeded(true);
    try {
      const isSaved = await readSaved(
        pending.current?.body.meaningId ?? meaningId,
      );
      if (!mounted.current) return;
      pending.current = null;
      setCheckNeeded(false);
      setSaved(isSaved);
      if (isSaved) startRefresh(() => router.refresh());
      else setErrorMessage("This word was not saved. You can save it again.");
    } catch (cause) {
      if (mounted.current) {
        setErrorMessage(
          errorText(cause, "We could not check this word. Please try again."),
        );
      }
    }
  }

  async function save() {
    pending.current ??= {
      key: generateIdempotencyKey(),
      body: { meaningId, source },
    };
    const { key, body } = pending.current;
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) return setErrorMessage(SESSION_NOT_READY);
      const { data } = await createApiClient().saveUserWord(body, key, {
        headers: { "X-CSRF-Token": token },
      });
      if (!mounted.current) return;
      if (
        data?.meaningId !== body.meaningId ||
        data.saved !== true ||
        !data.userWordId
      ) {
        return check();
      }
      pending.current = null;
      setSaved(true);
      // Sentence practice is server-rendered from saved state; refresh to show it.
      startRefresh(() => router.refresh());
    } catch (cause) {
      if (!mounted.current) return;
      if (
        cause instanceof ApiResponseError &&
        (cause.status === 404 || cause.status === 409)
      ) {
        return check();
      }
      setErrorMessage(
        errorText(cause, "We could not save this word. Please try again."),
      );
    }
  }

  async function remove() {
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) return setErrorMessage(SESSION_NOT_READY);
      await createApiClient().unsaveUserWord(meaningId, {
        headers: { "X-CSRF-Token": token },
      });
      if (!mounted.current) return;
      setSaved(false);
      startRefresh(() => router.refresh());
    } catch (cause) {
      if (mounted.current) {
        setErrorMessage(
          errorText(cause, "Unable to update saved state. Please try again."),
        );
      }
    }
  }

  function handleClick() {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    setErrorMessage("");
    const action = checkNeeded ? check() : saved ? remove() : save();
    void action.finally(() => {
      inFlight.current = false;
      if (mounted.current) setBusy(false);
    });
  }

  const ariaLabel = checkNeeded
    ? `Check saved status for ${wordText}`
    : saved
      ? `Remove ${wordText} from saved words`
      : `Save ${wordText}: ${shortDefinition}`;
  const label = checkNeeded
    ? "Check saved status"
    : saved
      ? "Saved"
      : "Save to review";

  return (
    <div className="flex flex-col items-end gap-[var(--spacing-xs)]">
      <button
        type="button"
        onClick={handleClick}
        aria-disabled={busy || undefined}
        aria-pressed={saved}
        aria-label={ariaLabel}
        aria-busy={busy}
        className={`inline-flex min-h-11 min-w-11 items-center gap-1.5 rounded-full border px-4 text-sm font-semibold transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 aria-disabled:cursor-not-allowed aria-disabled:opacity-50 ${saved ? "border-primary-200 bg-primary-50 text-primary-800 hover:bg-primary-100" : "border-primary-300 bg-white text-primary-700 hover:bg-primary-50"}`}
      >
        <svg
          aria-hidden="true"
          viewBox="0 0 24 24"
          className="size-4 fill-none stroke-current stroke-2"
        >
          {saved ? (
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="m5 12.5 4.5 4.5L19 7.5"
            />
          ) : (
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M12 5v14M5 12h14"
            />
          )}
        </svg>
        {busy ? (checkNeeded ? "Checking…" : "Saving…") : label}
      </button>
      {errorMessage ? (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          aria-live="polite"
          className="text-sm text-red-700"
        >
          {errorMessage}
        </StatusMessage>
      ) : null}
    </div>
  );
}

function generateIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}
