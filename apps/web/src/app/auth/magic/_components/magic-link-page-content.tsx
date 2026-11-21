"use client";

import { StatusMessage } from "@/ui/status-message";

import { ActionLink } from "@/ui/controls";

import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { createApiClient } from "@/lib/api";
import { getAuthErrorMessage } from "@/lib/auth-feedback";
import { normalizeReturnTo } from "@/lib/return-to";
import { AuthShell } from "@/ui/auth-shell";

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
    <AuthShell title="Sign in link" linkBrand={false}>
      <StatusMessage
        as="p"
        tone={status.type === "error" ? "error" : "info"}
        role={status.type === "error" ? "alert" : "status"}
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
          <ActionLink href="/" variant="text">
            Back to home
          </ActionLink>
        </div>
      ) : null}
    </AuthShell>
  );
}
