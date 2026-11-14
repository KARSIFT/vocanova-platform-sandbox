"use client";

import { StatusMessage } from "@/ui/status-message";

import { Button, TextInput } from "@/ui/controls";

import { useEffect, useRef, useState } from "react";

import { createApiClient } from "@/lib/api";
import type { SignInAuthCapabilities } from "@/lib/auth-capabilities";
import { getAuthErrorMessage, type OAuthIntent } from "@/lib/auth-feedback";
import { getAppOrigin } from "@/lib/env";
import {
  clearOAuthContinuation,
  rememberOAuthContinuation,
} from "@/lib/oauth-continuation";

interface MagicLinkFormProps {
  returnTo: string;
  emailLabel?: string;
  /** Use a quieter action when another sign-in method leads the page. */
  secondary?: boolean;
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

export function MagicLinkForm({
  returnTo,
  emailLabel = "Email address",
  secondary = false,
}: MagicLinkFormProps) {
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
        <div className="rounded-xl border border-primary-200 bg-primary-50 p-[var(--spacing-md)] text-primary-900">
          <h2 id="check-email-heading" className="text-lg font-semibold">
            Check your email
          </h2>
          <StatusMessage
            as="p"
            tone="info"
            role="status"
            aria-live="polite"
            className="mt-[var(--spacing-xs)] text-base"
          >
            If an account can use this address, we&apos;ll send a link there.
            <span className="mt-[var(--spacing-xs)] block break-all font-medium">
              {email}
            </span>
            The link signs you in wherever you open it.
          </StatusMessage>
        </div>
        <div className="flex flex-wrap gap-[var(--spacing-sm)]">
          <Button
            type="button"
            onClick={() => void requestLink()}
            disabled={isSending || resendSeconds > 0}
            aria-busy={isSending}
          >
            {isSending
              ? "Sending..."
              : resendSeconds > 0
                ? `Resend available in ${resendSeconds}s`
                : "Resend sign-in link"}
          </Button>
          <Button
            type="button"
            onClick={() => {
              setPhase("idle");
              setHasSent(false);
              setErrorMessage(null);
              setResendSeconds(0);
              clearOAuthContinuation();
            }}
            disabled={isSending}
            variant="secondary"
          >
            Use a different email
          </Button>
        </div>
        {errorMessage ? (
          <StatusMessage
            as="p"
            tone="error"
            role="alert"
            aria-live="assertive"
            className="text-base"
          >
            {errorMessage}
          </StatusMessage>
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
          {emailLabel}
        </label>
        <TextInput
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="mt-[var(--spacing-xs)]"
        />
      </div>
      {errorMessage ? (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          aria-live="assertive"
          className="text-base"
        >
          {errorMessage}
        </StatusMessage>
      ) : null}
      <Button
        type="submit"
        disabled={phase === "sending"}
        aria-busy={phase === "sending"}
        variant={secondary ? "secondary" : "primary"}
        className="w-full"
      >
        {phase === "sending" ? "Sending..." : "Send sign-in link"}
      </Button>
    </form>
  );
}

interface OAuthButtonProps {
  returnTo: string;
  capabilities: SignInAuthCapabilities;
  intent?: OAuthIntent;
}

export function OAuthButton({
  returnTo,
  capabilities,
  intent = "signin",
}: OAuthButtonProps) {
  const [rememberMe, setRememberMe] = useState(true);
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
      const { data } = await client.startOAuth({ redirectUri, rememberMe });
      window.location.href = data.url;
    } catch (error) {
      clearOAuthContinuation();
      setStatus({
        type: "error",
        message: getAuthErrorMessage(
          error,
          "oauth-start",
          capabilities,
          intent,
        ),
      });
    }
  }

  return (
    <div className="space-y-[var(--spacing-sm)]">
      <label className="flex min-h-[var(--spacing-2xl)] cursor-pointer items-center gap-[var(--spacing-sm)] text-base text-neutral-900">
        <input
          type="checkbox"
          checked={rememberMe}
          onChange={(event) => setRememberMe(event.target.checked)}
          disabled={status.type === "loading"}
          aria-describedby="google-session-duration"
          className="h-5 w-5 accent-primary-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        />
        Keep me signed in
      </label>
      <p id="google-session-duration" className="text-sm text-neutral-600">
        {rememberMe
          ? "Up to 30 days on this device."
          : "A shorter session, up to 24 hours."}
      </p>
      <Button
        type="button"
        onClick={handleClick}
        disabled={status.type === "loading"}
        aria-busy={status.type === "loading"}
        variant="secondary"
        className="w-full"
      >
        {status.type === "loading" ? "Redirecting..." : "Continue with Google"}
      </Button>
      {status.message ? (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          aria-live="polite"
          className="text-base"
        >
          {status.message}
        </StatusMessage>
      ) : null}
    </div>
  );
}
