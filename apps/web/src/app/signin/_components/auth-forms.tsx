"use client";

import { useEffect, useRef, useState } from "react";

import { createApiClient } from "@/lib/api";
import { getAuthErrorMessage } from "@/lib/auth-feedback";
import { getAppOrigin } from "@/lib/env";
import {
  clearOAuthContinuation,
  rememberOAuthContinuation,
} from "@/lib/oauth-continuation";

interface MagicLinkFormProps {
  returnTo: string;
}

const RESEND_COOLDOWN_SECONDS = 60;

function oauthReturnTo(returnTo: string): string {
  // OAuth's server allowlist intentionally permits only the post-auth entry
  // points deployed with the provider configuration. A tab-local continuation
  // resumes the original safe destination after the Home callback.
  return returnTo === "/onboarding" || returnTo === "/home"
    ? returnTo
    : "/home";
}

export function MagicLinkForm({ returnTo }: MagicLinkFormProps) {
  const [email, setEmail] = useState("");
  const [phase, setPhase] = useState<"idle" | "sending" | "sent">("idle");
  const [hasSent, setHasSent] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [resendSeconds, setResendSeconds] = useState(0);
  const requesting = useRef(false);

  useEffect(() => {
    if (resendSeconds === 0) {
      return;
    }
    const interval = window.setInterval(() => {
      setResendSeconds((seconds) => Math.max(0, seconds - 1));
    }, 1000);
    return () => window.clearInterval(interval);
  }, [resendSeconds]);

  async function requestLink() {
    if (requesting.current) {
      return;
    }
    requesting.current = true;
    const wasSent = phase === "sent";
    setPhase("sending");
    setErrorMessage(null);

    const client = createApiClient();
    try {
      await client.requestMagicLink({ email, returnTo });
      setHasSent(true);
      setPhase("sent");
      setResendSeconds(RESEND_COOLDOWN_SECONDS);
    } catch (error) {
      setPhase(wasSent ? "sent" : "idle");
      setErrorMessage(getAuthErrorMessage(error, "magic-request"));
    } finally {
      requesting.current = false;
    }
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await requestLink();
  }

  if (hasSent) {
    const isSending = phase === "sending";
    return (
      <section
        aria-labelledby="check-email-heading"
        className="space-y-[var(--spacing-md)]"
      >
        <div className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-md)] text-primary-900">
          <h2 id="check-email-heading" className="text-lg font-semibold">
            Check your email
          </h2>
          <p
            role="status"
            aria-live="polite"
            className="mt-[var(--spacing-xs)] text-base"
          >
            If an account can use this address, we&apos;ll send a link there.
            <span className="mt-[var(--spacing-xs)] block break-all font-medium">
              {email}
            </span>
            The link signs you in wherever you open it.
          </p>
        </div>
        <div className="flex flex-wrap gap-[var(--spacing-sm)]">
          <button
            type="button"
            onClick={() => void requestLink()}
            disabled={isSending || resendSeconds > 0}
            aria-busy={isSending}
            className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isSending
              ? "Sending..."
              : resendSeconds > 0
                ? `Resend available in ${resendSeconds}s`
                : "Resend sign-in link"}
          </button>
          <button
            type="button"
            onClick={() => {
              setPhase("idle");
              setHasSent(false);
              setErrorMessage(null);
              setResendSeconds(0);
              clearOAuthContinuation();
            }}
            disabled={isSending}
            className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Use a different email
          </button>
        </div>
        {errorMessage ? (
          <p
            role="alert"
            aria-live="assertive"
            className="text-base text-red-700"
          >
            {errorMessage}
          </p>
        ) : null}
      </section>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-[var(--spacing-md)]">
      <div>
        <label
          htmlFor="email"
          className="block text-base font-medium text-neutral-900"
        >
          Email address
        </label>
        <input
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="mt-[var(--spacing-xs)] block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
        />
      </div>
      {errorMessage ? (
        <p
          role="alert"
          aria-live="assertive"
          className="text-base text-red-700"
        >
          {errorMessage}
        </p>
      ) : null}
      <button
        type="submit"
        disabled={phase === "sending"}
        aria-busy={phase === "sending"}
        className="min-h-[var(--spacing-2xl)] w-full rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {phase === "sending" ? "Sending..." : "Send sign-in link"}
      </button>
    </form>
  );
}

interface OAuthButtonProps {
  returnTo: string;
}

export function OAuthButton({ returnTo }: OAuthButtonProps) {
  const [status, setStatus] = useState<{
    type: "idle" | "loading" | "error";
    message: string;
  }>({
    type: "idle",
    message: "",
  });

  async function handleClick() {
    setStatus({ type: "loading", message: "Redirecting to Google..." });
    const client = createApiClient();
    try {
      rememberOAuthContinuation(returnTo);
      const redirectUri = `${getAppOrigin()}${oauthReturnTo(returnTo)}`;
      const { data } = await client.startOAuth({ redirectUri });
      window.location.href = data.url;
    } catch (error) {
      clearOAuthContinuation();
      setStatus({
        type: "error",
        message: getAuthErrorMessage(error, "oauth-start"),
      });
    }
  }

  return (
    <div className="space-y-[var(--spacing-sm)]">
      <button
        type="button"
        onClick={handleClick}
        disabled={status.type === "loading"}
        aria-busy={status.type === "loading"}
        className="min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] w-full rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {status.type === "loading" ? "Redirecting..." : "Continue with Google"}
      </button>
      {status.message ? (
        <p role="alert" aria-live="polite" className="text-base text-red-700">
          {status.message}
        </p>
      ) : null}
    </div>
  );
}
