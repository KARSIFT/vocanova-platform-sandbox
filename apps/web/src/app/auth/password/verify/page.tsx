"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useRef, useState } from "react";

import { createApiClient } from "@/lib/api";
import { Surface } from "@/ui/surface";

function VerifyPasswordSignupContent() {
  const token = useSearchParams().get("token") ?? "";
  const [status, setStatus] = useState<
    "ready" | "verifying" | "done" | "error"
  >(token ? "ready" : "error");
  const verificationInFlight = useRef(false);

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
      setStatus("done");
    } catch {
      verificationInFlight.current = false;
      setStatus("error");
    }
  }

  const message =
    status === "ready"
      ? "Confirm this email verification to activate your password."
      : status === "verifying"
        ? "Verifying your email..."
        : status === "done"
          ? "Your email is verified and your password is ready to use."
          : "This verification link is invalid, expired, or has already been used.";
  return (
    <main className="grid min-h-screen place-items-center bg-neutral-100 p-6">
      <Surface className="w-full max-w-[28rem] space-y-[var(--spacing-md)]">
        <p className="text-sm font-bold tracking-wide text-primary-700">
          VOCANOVA
        </p>
        <h1 className="text-2xl font-semibold text-neutral-900">
          Verify your email
        </h1>
        <p
          role={status === "error" ? "alert" : "status"}
          aria-live="polite"
          className={`text-base ${status === "error" ? "text-red-700" : "text-neutral-700"}`}
        >
          {message}
        </p>
        {status === "ready" || status === "verifying" ? (
          <button
            type="button"
            onClick={() => void verifyEmail()}
            disabled={status === "verifying"}
            aria-busy={status === "verifying"}
            className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {status === "verifying" ? "Verifying..." : "Verify email"}
          </button>
        ) : null}
        {status === "done" ? (
          <Link
            href="/login"
            className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Sign in
          </Link>
        ) : null}
        {status === "error" ? (
          <Link
            href="/signup"
            className="font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Create account
          </Link>
        ) : null}
      </Surface>
    </main>
  );
}

export default function VerifyPasswordSignupPage() {
  return (
    <Suspense
      fallback={
        <main className="grid min-h-screen place-items-center p-6">
          <p className="text-base text-neutral-700">Loading verification...</p>
        </main>
      }
    >
      <VerifyPasswordSignupContent />
    </Suspense>
  );
}
