"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

interface RemoveSavedWordButtonProps {
  meaningId: string;
  wordText: string;
  redirectTo?: string;
}

export function RemoveSavedWordButton({
  meaningId,
  wordText,
  redirectTo,
}: RemoveSavedWordButtonProps) {
  const router = useRouter();
  const [status, setStatus] = useState<
    "idle" | "loading" | "removed" | "error"
  >("idle");
  const [errorMessage, setErrorMessage] = useState("");

  async function removeSavedWord() {
    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setStatus("error");
      setErrorMessage("Session is not ready. Please refresh the page.");
      return;
    }

    setStatus("loading");
    setErrorMessage("");
    try {
      await createApiClient().unsaveUserWord(meaningId, {
        headers: { "X-CSRF-Token": csrfToken },
      });
      setStatus("removed");
      if (redirectTo) {
        router.push(redirectTo);
      } else {
        router.refresh();
      }
    } catch (error) {
      setStatus("error");
      setErrorMessage(
        handleApiError(
          error,
          `Unable to remove ${wordText}. Please try again.`,
        ),
      );
    }
  }

  if (status === "removed" && !redirectTo) {
    return (
      <p role="status" className="text-sm font-medium text-neutral-700">
        Removed from saved words
      </p>
    );
  }

  return (
    <div className="flex flex-col items-end gap-[var(--spacing-xs)]">
      <button
        type="button"
        onClick={removeSavedWord}
        disabled={status === "loading" || status === "removed"}
        aria-busy={status === "loading"}
        className="min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-800 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {status === "loading" ? "Removing..." : `Remove ${wordText}`}
      </button>
      {errorMessage ? (
        <p role="alert" aria-live="polite" className="text-sm text-red-700">
          {errorMessage}
        </p>
      ) : null}
    </div>
  );
}
