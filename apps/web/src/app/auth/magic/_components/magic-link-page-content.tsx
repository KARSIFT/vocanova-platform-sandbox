"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";

export function MagicLinkPageContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const email = searchParams.get("email") ?? "";
  const [status, setStatus] = useState<{
    type: "loading" | "error";
    message: string;
  }>({
    type: "loading",
    message: "Verifying your sign-in link...",
  });

  useEffect(() => {
    if (!token || !email) {
      setStatus({
        type: "error",
        message: "This sign-in link is incomplete. Please request a new one.",
      });
      return;
    }

    const client = createApiClient();
    client
      .consumeMagicLink({ token, email })
      .then(() => {
        window.location.href = "/home";
      })
      .catch((error: unknown) => {
        const message =
          error instanceof ApiResponseError
            ? error.message
            : "This sign-in link is invalid or has expired. Please request a new one.";
        setStatus({ type: "error", message });
      });
  }, [token, email]);

  return (
    <main className="grid min-h-screen place-items-center p-6">
      {/* max-w-[28rem] (not max-w-md): see the token-collision note on
          /onboarding's page.tsx - tokens.generated.css's --spacing-md
          (16px) shadows the intended 28rem max-w-md container size. */}
      <div className="w-full max-w-[28rem] space-y-[var(--spacing-md)] rounded-xl border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm">
        <h1 className="text-2xl font-semibold text-neutral-900">
          Sign in link
        </h1>
        <p
          role="alert"
          aria-live="polite"
          className={`text-base ${status.type === "error" ? "text-red-700" : "text-neutral-700"}`}
        >
          {status.message}
        </p>
        {status.type === "error" ? (
          <Link
            href="/signin"
            className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Back to sign in
          </Link>
        ) : null}
      </div>
    </main>
  );
}
