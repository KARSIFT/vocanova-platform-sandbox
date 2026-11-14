"use client";

import { StatusMessage } from "@/ui/status-message";

import { ActionLink, Button } from "@/ui/controls";

import Link from "@/ui/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

type ConfirmationState =
  | { type: "loading" }
  | { type: "success"; email: string }
  | { type: "signin" }
  | { type: "error"; message: string; recovery?: "retry" | "refresh" };

export function EmailChangePageContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const inFlight = useRef(false);
  const retryButton = useRef<HTMLButtonElement>(null);
  const accountLink = useRef<HTMLAnchorElement>(null);
  const signInLink = useRef<HTMLAnchorElement>(null);
  const [state, setState] = useState<ConfirmationState>({ type: "loading" });

  const confirmEmailChange = useCallback(async () => {
    if (inFlight.current) return;
    if (!token) {
      setState({
        type: "error",
        message:
          "This confirmation link is incomplete. Request a new one from account settings.",
      });
      return;
    }

    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setState({ type: "signin" });
      return;
    }

    const client = createApiClient();
    const headers = { "X-CSRF-Token": csrfToken };
    const invalidLinkMessage =
      "This confirmation link is invalid or has expired. Request a new one from account settings.";

    function showRecoverableError(error: unknown) {
      const status =
        error instanceof ApiResponseError ? error.status : undefined;
      setState({
        type: "error",
        recovery: status === 403 ? "refresh" : "retry",
        message:
          status === 403
            ? "This page needs to be refreshed before we can confirm your email securely."
            : status === 429
              ? "Too many confirmation attempts were made. Please wait a few minutes, then try again."
              : "We couldn't confirm your email right now. Please try again.",
      });
    }

    inFlight.current = true;
    setState({ type: "loading" });
    try {
      try {
        await client.getCurrentUser();
      } catch (error) {
        if (error instanceof ApiResponseError && error.status === 401) {
          setState({ type: "signin" });
          return;
        }
        showRecoverableError(error);
        return;
      }

      try {
        const { data } = await client.consumeEmailChangeLink(
          { token },
          { headers },
        );
        setState({ type: "success", email: data.email });
      } catch (error) {
        if (error instanceof ApiResponseError && error.status === 401) {
          // The API deliberately uses 401 for both an absent session and an
          // invalid/expired token. Re-check the session before deciding
          // whether sign-in can help; otherwise an unusable token would loop
          // through sign-in forever.
          try {
            await client.getCurrentUser();
          } catch (sessionError) {
            if (
              sessionError instanceof ApiResponseError &&
              sessionError.status === 401
            ) {
              setState({ type: "signin" });
              return;
            }
            // Without a confirmed session, a failed recheck cannot establish
            // that the proof itself is invalid. Keep it for an explicit retry.
            showRecoverableError(sessionError);
            return;
          }
          setState({ type: "error", message: invalidLinkMessage });
          return;
        }
        if (error instanceof ApiResponseError && error.status === 409) {
          setState({
            type: "error",
            message:
              "This email address could not be used. Request a new confirmation link from account settings.",
          });
          return;
        }
        showRecoverableError(error);
      }
    } finally {
      inFlight.current = false;
    }
  }, [token]);

  useEffect(() => {
    void confirmEmailChange();
  }, [confirmEmailChange]);

  useEffect(() => {
    if (state.type === "loading") return;
    if (state.type === "signin") signInLink.current?.focus();
    else if (state.type === "error" && state.recovery)
      retryButton.current?.focus();
    else accountLink.current?.focus();
  }, [state]);

  const returnTo = `/auth/email-change?${new URLSearchParams({ token }).toString()}`;

  return (
    <AuthShell>
      <Surface className="w-full space-y-[var(--spacing-md)]">
        <h1 className="text-2xl font-semibold text-neutral-900">
          Confirm your new email
        </h1>

        {state.type === "loading" ? (
          <StatusMessage
            as="p"
            tone="info"
            role="status"
            aria-live="polite"
            className="text-base"
          >
            Verifying your confirmation link...
          </StatusMessage>
        ) : null}

        {state.type === "success" ? (
          <>
            <StatusMessage
              as="p"
              tone="success"
              role="status"
              aria-live="polite"
              className="text-base [overflow-wrap:anywhere]"
            >
              Your sign-in email is now {state.email}.
            </StatusMessage>
            <ActionLink ref={accountLink} href="/settings/account">
              Return to account settings
            </ActionLink>
          </>
        ) : null}

        {state.type === "signin" ? (
          <>
            <p className="text-base text-neutral-700">
              Sign in to the account that requested this change, then open the
              confirmation link again.
            </p>
            <ActionLink
              ref={signInLink}
              href={`/login?${new URLSearchParams({ returnTo, magicOnly: "1" }).toString()}`}
            >
              Sign in
            </ActionLink>
            <Link
              href="/"
              className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Back to home
            </Link>
          </>
        ) : null}

        {state.type === "error" ? (
          <>
            <StatusMessage
              as="p"
              tone="error"
              role="alert"
              aria-live="assertive"
              className="text-base"
            >
              {state.message}
            </StatusMessage>
            {state.recovery ? (
              <Button
                ref={retryButton}
                type="button"
                onClick={() => {
                  if (state.recovery === "refresh") window.location.reload();
                  else void confirmEmailChange();
                }}
              >
                {state.recovery === "refresh"
                  ? "Refresh page"
                  : "Retry confirmation"}
              </Button>
            ) : null}
            <ActionLink
              ref={accountLink}
              href="/settings/account"
              variant="secondary"
            >
              Go to account settings
            </ActionLink>
            <Link
              href="/"
              className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Back to home
            </Link>
          </>
        ) : null}
      </Surface>
    </AuthShell>
  );
}
