"use client";

import { useState } from "react";

import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";

interface EmailChangeFormProps {
  currentEmail: string;
}

type EmailPhase =
  | { type: "idle" }
  | { type: "requesting" }
  | { type: "pending"; newEmail: string }
  | { type: "completed"; newEmail: string; previousEmail: string }
  | { type: "error"; message: string };

export function EmailChangeForm({ currentEmail }: EmailChangeFormProps) {
  const [newEmail, setNewEmail] = useState("");
  const [token, setToken] = useState("");
  const [isConsuming, setIsConsuming] = useState(false);
  const [phase, setPhase] = useState<EmailPhase>({ type: "idle" });

  async function handleRequest(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = newEmail.trim();
    if (!trimmed) {
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

    setPhase({ type: "requesting" });
    const client = createApiClient();
    try {
      await client.requestEmailChangeLink(
        { newEmail: trimmed },
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      setPhase({ type: "pending", newEmail: trimmed });
    } catch (error) {
      const message =
        error instanceof ApiResponseError
          ? error.message
          : "We couldn't send the confirmation email. Please try again.";
      setPhase({ type: "error", message });
    }
  }

  async function handleConsume(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmedToken = token.trim();
    if (!trimmedToken) {
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

    setIsConsuming(true);
    const client = createApiClient();
    try {
      const { data } = await client.consumeEmailChangeLink(
        { token: trimmedToken },
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      setPhase({
        type: "completed",
        newEmail: data.email,
        previousEmail: data.previousEmail,
      });
      setToken("");
      setNewEmail("");
    } catch (error) {
      const message =
        error instanceof ApiResponseError
          ? error.message
          : "We couldn't confirm that link. Please try again.";
      setPhase({ type: "error", message });
    } finally {
      setIsConsuming(false);
    }
  }

  function handleStartOver() {
    setPhase({ type: "idle" });
    setNewEmail("");
    setToken("");
  }

  return (
    <div className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]">
      {phase.type === "completed" ? (
        <div
          role="status"
          aria-live="polite"
          className="rounded-md border border-green-300 bg-green-50 p-[var(--spacing-md)] text-base text-green-900"
        >
          <p className="font-semibold">Your sign-in email was updated.</p>
          <p className="mt-[var(--spacing-xs)]">
            New address: <span className="font-medium">{phase.newEmail}</span>
            {phase.previousEmail ? (
              <>
                {" "}
                (previously{" "}
                <span className="font-medium">{phase.previousEmail}</span>)
              </>
            ) : null}
          </p>
          <button
            type="button"
            onClick={handleStartOver}
            className="mt-[var(--spacing-sm)] text-sm font-medium text-primary-700 underline hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Change to another address
          </button>
        </div>
      ) : null}

      {phase.type === "pending" ? (
        <div className="space-y-[var(--spacing-md)]">
          <p
            role="status"
            aria-live="polite"
            className="rounded-md border border-primary-300 bg-primary-50 p-[var(--spacing-md)] text-base text-primary-900"
          >
            We sent a confirmation link to{" "}
            <span className="font-medium">{phase.newEmail}</span>. The link
            expires in 15 minutes. Paste the link&apos;s token below to finish
            the change.
          </p>
          <form
            onSubmit={handleConsume}
            aria-label="Confirm new sign-in email"
            className="space-y-[var(--spacing-sm)]"
          >
            <label
              htmlFor="email-change-token"
              className="block text-base font-medium text-neutral-900"
            >
              Confirmation token
            </label>
            <input
              id="email-change-token"
              name="token"
              type="text"
              required
              autoComplete="off"
              value={token}
              onChange={(event) => setToken(event.target.value)}
              disabled={isConsuming}
              className="block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
            />
            <div className="flex flex-wrap gap-[var(--spacing-sm)]">
              <button
                type="submit"
                disabled={isConsuming || token.trim().length === 0}
                aria-busy={isConsuming}
                className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {isConsuming ? "Confirming..." : "Confirm change"}
              </button>
              <button
                type="button"
                onClick={handleStartOver}
                disabled={isConsuming}
                className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      ) : null}

      {phase.type === "idle" ||
      phase.type === "requesting" ||
      phase.type === "error" ? (
        <form
          onSubmit={handleRequest}
          aria-label="Request new sign-in email"
          className="space-y-[var(--spacing-sm)]"
        >
          <label
            htmlFor="new-email"
            className="block text-base font-medium text-neutral-900"
          >
            New sign-in email
          </label>
          <input
            id="new-email"
            name="newEmail"
            type="email"
            autoComplete="email"
            required
            value={newEmail}
            onChange={(event) => {
              setNewEmail(event.target.value);
              if (phase.type === "error") {
                setPhase({ type: "idle" });
              }
            }}
            disabled={phase.type === "requesting"}
            placeholder="you@example.com"
            aria-describedby="new-email-helper"
            className="block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
          />
          <p id="new-email-helper" className="text-sm text-neutral-700">
            We&apos;ll send a confirmation link to this address. The response is
            the same whether this address is already registered or not, to
            protect learner privacy.
          </p>
          <button
            type="submit"
            disabled={
              phase.type === "requesting" || newEmail.trim().length === 0
            }
            aria-busy={phase.type === "requesting"}
            className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {phase.type === "requesting"
              ? "Sending confirmation link..."
              : "Send confirmation link"}
          </button>
        </form>
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

      {currentEmail ? (
        <p className="text-sm text-neutral-700">
          Your current sign-in address is{" "}
          <span className="font-medium text-neutral-900">{currentEmail}</span>.
        </p>
      ) : null}
    </div>
  );
}
