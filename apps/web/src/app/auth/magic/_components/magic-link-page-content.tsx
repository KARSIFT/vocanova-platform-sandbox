"use client";

import { StatusMessage } from "@/ui/status-message";

import { ActionLink } from "@/ui/controls";

import Link from "@/ui/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { createApiClient } from "@/lib/api";
import { getAuthErrorMessage } from "@/lib/auth-feedback";
import { normalizeReturnTo } from "@/lib/return-to";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

export function MagicLinkPageContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const email = searchParams.get("email") ?? "";
  const returnTo = normalizeReturnTo(searchParams.get("returnTo"));
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
        window.location.href = returnTo;
      })
      .catch((error: unknown) => {
        setStatus({
          type: "error",
          message: getAuthErrorMessage(error, "magic-consume"),
        });
      });
  }, [token, email, returnTo]);

  return (
    <AuthShell>
      <Surface className="w-full space-y-[var(--spacing-md)]">
        <h1 className="text-2xl font-semibold text-neutral-900">
          Sign in link
        </h1>
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          aria-live="polite"
          className={`text-base ${status.type === "error" ? "text-red-700" : "text-neutral-700"}`}
        >
          {status.message}
        </StatusMessage>
        {status.type === "error" ? (
          <div className="flex flex-wrap gap-[var(--spacing-sm)]">
            <ActionLink
              href={`/login?${new URLSearchParams({ returnTo }).toString()}`}
            >
              Back to sign in
            </ActionLink>
            <Link
              href="/"
              className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Back to home
            </Link>
          </div>
        ) : null}
      </Surface>
    </AuthShell>
  );
}
