"use client";

import { Button } from "@/ui/controls";

import Link from "@/ui/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useRef, useState } from "react";

import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

function VerifyPasswordSignupContent() {
  const token = useSearchParams().get("token") ?? "";
  const [status, setStatus] = useState<
    "ready" | "verifying" | "done" | "retry" | "error"
  >(token ? "ready" : "error");
  const [retryMessage, setRetryMessage] = useState("");
  const verificationInFlight = useRef(false);
  const moveOutcomeFocus = useRef(false);
  const retryButton = useRef<HTMLButtonElement>(null);
  const signInLink = useRef<HTMLAnchorElement>(null);
  const signupLink = useRef<HTMLAnchorElement>(null);

  useEffect(() => {
    if (!moveOutcomeFocus.current || status === "verifying") return;
    moveOutcomeFocus.current = false;
    if (status === "retry") retryButton.current?.focus();
    else if (status === "done") signInLink.current?.focus();
    else if (status === "error") signupLink.current?.focus();
  }, [status]);

  async function verifyEmail() {
    if (!token || verificationInFlight.current) {
      return;
    }
    verificationInFlight.current = true;
    setStatus("verifying");
    try {
      await createApiClient().verifyPasswordSignup({ token });
      // Keep the success state in memory, but remove this one-time proof from
      // the address bar so it cannot be copied, restored, or sent as a
      // referrer by a later navigation.
      window.history.replaceState(
        window.history.state,
        "",
        "/auth/password/verify",
      );
      moveOutcomeFocus.current = true;
      setStatus("done");
    } catch (error) {
      verificationInFlight.current = false;
      moveOutcomeFocus.current = true;
      // The API uses 401 for invalid/expired proof. Connection errors,
      // rate limits and service failures do not establish that it is invalid.
      if (error instanceof ApiResponseError && error.status === 401) {
        setStatus("error");
      } else {
        setRetryMessage(
          error instanceof ApiResponseError && error.status === 429
            ? "Too many verification attempts were made. Please wait a few minutes, then try again."
            : "We couldn't confirm your email right now. Please try again.",
        );
        setStatus("retry");
      }
    }
  }

  const message =
    status === "ready"
      ? "Confirm this email verification to activate your password."
      : status === "verifying"
        ? "Verifying your email..."
        : status === "done"
          ? "Your email is verified and your password is ready to use."
          : status === "retry"
            ? retryMessage
            : "This verification link is invalid, expired, or has already been used.";
  return (
    <AuthShell>
      <Surface className="w-full space-y-[var(--spacing-md)]">
        <h1 className="text-2xl font-semibold text-neutral-900">
          Verify your email
        </h1>
        <p
          role={status === "error" || status === "retry" ? "alert" : "status"}
          aria-live="polite"
          className={`text-base ${status === "error" || status === "retry" ? "text-red-700" : "text-neutral-700"}`}
        >
          {message}
        </p>
        {status === "ready" || status === "verifying" || status === "retry" ? (
          <Button
            ref={retryButton}
            type="button"
            onClick={() => void verifyEmail()}
            disabled={status === "verifying"}
            aria-busy={status === "verifying"}
          >
            {status === "verifying"
              ? "Verifying..."
              : status === "retry"
                ? "Retry verification"
                : "Verify email"}
          </Button>
        ) : null}
        {status === "done" ? (
          <Link
            ref={signInLink}
            href="/login"
            className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Sign in
          </Link>
        ) : null}
        {status === "error" ? (
          <Link
            ref={signupLink}
            href="/signup"
            className="font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Create account
          </Link>
        ) : null}
      </Surface>
    </AuthShell>
  );
}

export default function VerifyPasswordSignupPage() {
  return (
    <Suspense
      fallback={
        <AuthShell>
          <p className="text-base text-neutral-700">Loading verification...</p>
        </AuthShell>
      }
    >
      <VerifyPasswordSignupContent />
    </Suspense>
  );
}
