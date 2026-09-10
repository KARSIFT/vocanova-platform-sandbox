"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { Surface } from "@/ui/surface";

type ConfirmationState =
  | { type: "loading" }
  | { type: "success"; email: string }
  | { type: "signin" }
  | { type: "error"; message: string };

export function EmailChangePageContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const started = useRef(false);
  const [state, setState] = useState<ConfirmationState>({ type: "loading" });

  useEffect(() => {
    if (started.current) return;
    started.current = true;

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

    async function confirmEmailChange() {
      try {
        await client.getCurrentUser();
      } catch (error) {
        if (error instanceof ApiResponseError && error.status === 401) {
          setState({ type: "signin" });
          return;
        }
        setState({ type: "error", message: invalidLinkMessage });
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
          }
          setState({ type: "error", message: invalidLinkMessage });
          return;
        }
        setState({
          type: "error",
          message:
            error instanceof ApiResponseError
              ? error.message
              : invalidLinkMessage,
        });
      }
    }

    void confirmEmailChange();
  }, [token]);

  const returnTo = `/auth/email-change?${new URLSearchParams({ token }).toString()}`;

  return (
    <main className="grid min-h-screen place-items-center bg-neutral-100 p-6">
      <Surface className="w-full max-w-[28rem] space-y-[var(--spacing-md)]">
        <p className="text-sm font-bold tracking-wide text-primary-700">
          VocaNova
        </p>
        <h1 className="text-2xl font-semibold text-neutral-900">
          Confirm your new email
        </h1>

        {state.type === "loading" ? (
          <p
            role="status"
            aria-live="polite"
            className="text-base text-neutral-700"
          >
            Verifying your confirmation link...
          </p>
        ) : null}

        {state.type === "success" ? (
          <>
            <p
              role="status"
              aria-live="polite"
              className="text-base text-green-800"
            >
              Your sign-in email is now {state.email}.
            </p>
            <Link
              href="/settings/account"
              className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Return to account settings
            </Link>
          </>
        ) : null}

        {state.type === "signin" ? (
          <>
            <p className="text-base text-neutral-700">
              Sign in to the account that requested this change, then open the
              confirmation link again.
            </p>
            <Link
              href={`/login?${new URLSearchParams({ returnTo, magicOnly: "1" }).toString()}`}
              className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Sign in
            </Link>
          </>
        ) : null}

        {state.type === "error" ? (
          <>
            <p
              role="alert"
              aria-live="assertive"
              className="text-base text-red-700"
            >
              {state.message}
            </p>
            <Link
              href="/settings/account"
              className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Go to account settings
            </Link>
          </>
        ) : null}
      </Surface>
    </main>
  );
}
