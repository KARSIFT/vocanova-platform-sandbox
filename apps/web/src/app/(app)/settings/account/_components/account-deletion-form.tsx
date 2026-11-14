"use client";

import { StatusMessage } from "@/ui/status-message";

import { Button, TextInput } from "@/ui/controls";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, deleteCookie } from "@/lib/cookies";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { clearOAuthContinuation } from "@/lib/oauth-continuation";
import { handleApiError } from "@/lib/session";

import { clearSentenceFeedbackDrafts } from "../../../_components/sentence-feedback-drafts";
import { clearPreferencesDrafts } from "@/lib/preferences-draft";

type DeletionPhase = { type: "idle" } | { type: "confirming" };

const CONFIRMATION_PHRASE = "delete my account";

function generateIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export function AccountDeletionForm() {
  const router = useRouter();
  const [phase, setPhase] = useState<DeletionPhase>({ type: "idle" });
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [typedPhrase, setTypedPhrase] = useState("");
  const [isDeleting, setIsDeleting] = useState(false);
  const deleteTriggerRef = useRef<HTMLButtonElement>(null);
  const confirmationInputRef = useRef<HTMLInputElement>(null);
  const shouldRestoreTriggerFocus = useRef(false);

  useEffect(() => {
    if (phase.type === "confirming") {
      confirmationInputRef.current?.focus();
      return;
    }
    if (shouldRestoreTriggerFocus.current) {
      shouldRestoreTriggerFocus.current = false;
      deleteTriggerRef.current?.focus();
    }
  }, [phase]);

  async function handleDelete() {
    if (typedPhrase.trim() !== CONFIRMATION_PHRASE) {
      setErrorMessage(
        `Please type "${CONFIRMATION_PHRASE}" exactly to confirm.`,
      );
      return;
    }

    setIsDeleting(true);
    const client = createApiClient();
    try {
      const csrfToken = await getOrRefreshCSRFToken();
      if (!csrfToken) {
        setErrorMessage(
          "We couldn't prepare this request securely. Please try again.",
        );
        return;
      }
      const { data } = await client.createAccountDeletionRequest(
        generateIdempotencyKey(),
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      // The server revoked all active sessions and cleared its HttpOnly
      // session cookie in this response. The browser-visible CSRF cookie can
      // be cleared here before leaving the authenticated shell.
      clearSentenceFeedbackDrafts();
      clearPreferencesDrafts();
      clearOAuthContinuation();
      deleteCookie(CSRF_COOKIE_NAME);
      // The acknowledgement must be outside the authenticated shell: this
      // request revokes the session, making its header and navigation invalid.
      router.replace(
        `/account-deactivated?purgeAfter=${encodeURIComponent(data.purgeAfter)}`,
      );
    } catch (error) {
      // T06: a 401 mid-account-deletion means the session expired
      // before the deactivation was issued. We never want to claim
      // an account is deactivated when the server rejected the
      // request; handleApiError routes the learner to re-auth and
      // the typed confirmation is preserved in component state so
      // the learner can re-confirm after re-authentication.
      const message = handleApiError(
        error,
        "We couldn't deactivate your account. Please try again.",
      );
      setErrorMessage(message);
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <div className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]">
      {phase.type === "idle" ? (
        <Button
          ref={deleteTriggerRef}
          type="button"
          onClick={() => {
            setErrorMessage(null);
            setPhase({ type: "confirming" });
          }}
          variant="danger"
          className="min-w-[var(--spacing-2xl)]"
        >
          I want to delete my account
        </Button>
      ) : null}

      {phase.type === "confirming" ? (
        <div className="rounded-md border border-red-300 bg-white p-[var(--spacing-md)] text-base text-neutral-900 shadow-sm">
          <h3 className="text-lg font-semibold text-red-900">
            Are you absolutely sure?
          </h3>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-800">
            This will deactivate your account immediately. Your saved words,
            reviews, and practice history will be permanently anonymized after
            30 days.
          </p>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-800">
            You can sign in again before then to reactivate, but after 30 days
            this is irreversible.
          </p>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-900">
            To continue, type{" "}
            <span className="font-mono font-semibold">
              {CONFIRMATION_PHRASE}
            </span>{" "}
            below.
          </p>
          <label className="sr-only" htmlFor="delete-confirmation">
            Type the confirmation phrase
          </label>
          <TextInput
            ref={confirmationInputRef}
            id="delete-confirmation"
            name="confirmation"
            type="text"
            autoComplete="off"
            value={typedPhrase}
            onChange={(event) => {
              setTypedPhrase(event.target.value);
              setErrorMessage(null);
            }}
            disabled={isDeleting}
            className="mt-[var(--spacing-xs)] w-full"
            aria-describedby={
              errorMessage
                ? "delete-confirmation-helper delete-confirmation-error"
                : "delete-confirmation-helper"
            }
          />
          <p
            id="delete-confirmation-helper"
            className="mt-[var(--spacing-xs)] text-sm text-neutral-700"
          >
            Type the phrase exactly, with the same lowercase letters and spaces.
          </p>
          <div className="mt-[var(--spacing-md)] flex flex-wrap gap-[var(--spacing-sm)]">
            <Button
              type="button"
              onClick={handleDelete}
              disabled={
                isDeleting || typedPhrase.trim() !== CONFIRMATION_PHRASE
              }
              aria-busy={isDeleting}
              variant="danger"
              className="min-w-[var(--spacing-2xl)]"
            >
              {isDeleting ? "Deactivating..." : "Deactivate my account"}
            </Button>
            <Button
              type="button"
              onClick={() => {
                shouldRestoreTriggerFocus.current = true;
                setPhase({ type: "idle" });
                setTypedPhrase("");
                setErrorMessage(null);
              }}
              disabled={isDeleting}
              variant="secondary"
              className="min-w-[var(--spacing-2xl)]"
            >
              Cancel
            </Button>
          </div>
        </div>
      ) : null}

      {errorMessage ? (
        <StatusMessage
          as="p"
          tone="error"
          id="delete-confirmation-error"
          role="alert"
          aria-live="assertive"
          className="rounded-md border border-red-300 bg-red-50 p-[var(--spacing-sm)] text-base"
        >
          {errorMessage}
        </StatusMessage>
      ) : null}
    </div>
  );
}
