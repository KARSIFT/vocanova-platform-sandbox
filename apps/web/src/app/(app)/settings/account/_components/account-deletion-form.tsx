"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { createApiClient } from "@/lib/api";
import {
  CSRF_COOKIE_NAME,
  deleteCookie,
  getCookieValue,
  SESSION_COOKIE_NAME,
} from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

type DeletionPhase =
  | { type: "idle" }
  | { type: "confirming" }
  | { type: "error"; message: string };

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
  const [typedPhrase, setTypedPhrase] = useState("");
  const [isDeleting, setIsDeleting] = useState(false);

  async function handleDelete() {
    if (typedPhrase.trim() !== CONFIRMATION_PHRASE) {
      setPhase({
        type: "error",
        message: `Please type "${CONFIRMATION_PHRASE}" exactly to confirm.`,
      });
      return;
    }

    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setPhase({
        type: "error",
        message:
          "Your session is missing a security token. Please refresh the page and try again.",
      });
      return;
    }

    setIsDeleting(true);
    const client = createApiClient();
    try {
      const { data } = await client.createAccountDeletionRequest(
        generateIdempotencyKey(),
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      // The server already revoked every active session; the local
      // cookie is just a presentation concern. Clear it so subsequent
      // requests (e.g. via a still-pending client request) cannot
      // re-use the deactivated session, then follow up with a
      // logout call to flush the requester-side session.
      deleteCookie(CSRF_COOKIE_NAME);
      deleteCookie(SESSION_COOKIE_NAME);
      try {
        await client.logout({ headers: { "X-CSRF-Token": csrfToken } });
      } catch {
        // Best-effort: the server has already revoked the session
        // even if the logout call fails on the client.
      }
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
      setPhase({ type: "error", message });
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <div className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]">
      {phase.type === "idle" ? (
        <button
          type="button"
          onClick={() => setPhase({ type: "confirming" })}
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-red-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-red-800 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-red-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-700"
        >
          I want to delete my account
        </button>
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
          <input
            id="delete-confirmation"
            name="confirmation"
            type="text"
            autoComplete="off"
            value={typedPhrase}
            onChange={(event) => setTypedPhrase(event.target.value)}
            disabled={isDeleting}
            className="mt-[var(--spacing-xs)] block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-red-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-red-600"
            aria-describedby="delete-confirmation-helper"
          />
          <p
            id="delete-confirmation-helper"
            className="mt-[var(--spacing-xs)] text-sm text-neutral-700"
          >
            Type the phrase exactly, with the same lowercase letters and spaces.
          </p>
          <div className="mt-[var(--spacing-md)] flex flex-wrap gap-[var(--spacing-sm)]">
            <button
              type="button"
              onClick={handleDelete}
              disabled={
                isDeleting || typedPhrase.trim() !== CONFIRMATION_PHRASE
              }
              aria-busy={isDeleting}
              className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-red-700 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-red-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isDeleting
                ? "Deactivating..."
                : "Permanently deactivate my account"}
            </button>
            <button
              type="button"
              onClick={() => {
                setPhase({ type: "idle" });
                setTypedPhrase("");
              }}
              disabled={isDeleting}
              className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : null}

      {phase.type === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="rounded-md border border-red-300 bg-red-50 p-[var(--spacing-sm)] text-base text-red-800"
        >
          {phase.message}
        </p>
      ) : null}
    </div>
  );
}
