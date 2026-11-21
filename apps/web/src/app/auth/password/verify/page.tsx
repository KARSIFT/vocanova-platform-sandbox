"use client";

import { ActionLink, Button } from "@/ui/controls";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useRef, useState } from "react";

import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { AuthShell } from "@/ui/auth-shell";

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
    <AuthShell title="Verify your email">
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
          className="justify-self-start"
        >
          {status === "verifying"
            ? "Verifying..."
            : status === "retry"
              ? "Retry verification"
              : "Verify email"}
        </Button>
      ) : null}
      {status === "done" ? (
        <ActionLink
          ref={signInLink}
          href="/login"
          className="justify-self-start"
        >
          Sign in
        </ActionLink>
      ) : null}
      {status === "error" ? (
        <ActionLink
          ref={signupLink}
          href="/signup"
          variant="secondary"
          className="justify-self-start"
        >
          Create account
        </ActionLink>
      ) : null}
    </AuthShell>
  );
}

export default function VerifyPasswordSignupPage() {
  return (
    <Suspense
      fallback={
        <AuthShell title="Verify your email">
          <p role="status" className="text-base text-neutral-700">
            Loading verification...
          </p>
        </AuthShell>
      }
    >
      <VerifyPasswordSignupContent />
    </Suspense>
  );
}
