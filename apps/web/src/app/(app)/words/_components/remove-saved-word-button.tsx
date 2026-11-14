"use client";

import { StatusMessage } from "@/ui/status-message";
import { Button } from "@/ui/controls";
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
      <StatusMessage
        as="p"
        role="status"
        className="text-sm font-medium text-neutral-700"
      >
        Removed from saved words
      </StatusMessage>
    );
  }

  return (
    <div className="flex flex-col items-end gap-[var(--spacing-xs)]">
      <Button
        variant="secondary"
        type="button"
        onClick={removeSavedWord}
        disabled={status === "loading" || status === "removed"}
        aria-busy={status === "loading"}
        className="min-w-[var(--spacing-2xl)] text-base text-neutral-800 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)]"
      >
        {status === "loading" ? "Removing..." : `Remove ${wordText}`}
      </Button>
      {errorMessage ? (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          className="text-sm text-red-700"
        >
          {errorMessage}
        </StatusMessage>
      ) : null}
    </div>
  );
}
