"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

interface MeaningSaveButtonProps {
  meaningId: string;
  source: "journey";
  initialSaved: boolean;
  wordText: string;
  shortDefinition: string;
}

export function MeaningSaveButton({
  meaningId,
  source,
  initialSaved,
  wordText,
  shortDefinition,
}: MeaningSaveButtonProps) {
  const router = useRouter();
  const [saved, setSaved] = useState(initialSaved);
  const [status, setStatus] = useState<"idle" | "loading" | "error">("idle");
  const [errorMessage, setErrorMessage] = useState("");

  async function toggleSave() {
    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setStatus("error");
      setErrorMessage("Session is not ready. Please refresh the page.");
      return;
    }

    setStatus("loading");
    setErrorMessage("");

    const client = createApiClient();
    try {
      if (saved) {
        await client.unsaveUserWord(meaningId, {
          headers: { "X-CSRF-Token": csrfToken },
        });
        setSaved(false);
      } else {
        const idempotencyKey = generateIdempotencyKey();
        await client.saveUserWord({ meaningId, source }, idempotencyKey, {
          headers: { "X-CSRF-Token": csrfToken },
        });
        setSaved(true);
      }
      setStatus("idle");
      // Word Detail's SentenceFeedback entry point (PRD §2 / #1181) is
      // gated on the server-rendered `meaning.saved` + `meaning.userWordId`
      // fields from the parent Server Component, not on this button's own
      // local `saved` state. Without a refresh, saving a word flips this
      // button to "Saved" but the sentence-practice widget below it never
      // appears until the learner manually reloads the page - refresh the
      // route so the server refetches and the widget shows up immediately.
      router.refresh();
    } catch (error) {
      setStatus("error");
      // handleApiError detects 401 and routes the learner to
      // re-authentication (T06: session-expiry mid-flow). For any
      // other failure it returns a stable, caller-supplied message.
      setErrorMessage(
        handleApiError(
          error,
          "Unable to update saved state. Please try again.",
        ),
      );
    }
  }

  const label = saved ? "Saved" : "Save";
  const ariaLabel = saved
    ? `Remove ${wordText} from saved words`
    : `Save ${wordText}: ${shortDefinition}`;

  return (
    <div className="flex flex-col items-end gap-[var(--spacing-xs)]">
      <button
        type="button"
        onClick={toggleSave}
        disabled={status === "loading"}
        aria-pressed={saved}
        aria-label={ariaLabel}
        aria-busy={status === "loading"}
        className="min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] rounded-md border border-neutral-200 bg-neutral-50 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-700 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {status === "loading" ? "Saving..." : label}
      </button>
      {errorMessage ? (
        <p role="alert" aria-live="polite" className="text-sm text-red-700">
          {errorMessage}
        </p>
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
