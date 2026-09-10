"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, deleteCookie, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

export function AppHeader() {
  const pathname = usePathname();
  const [status, setStatus] = useState<{
    type: "idle" | "loading" | "error";
    message: string;
  }>({
    type: "idle",
    message: "",
  });

  async function handleLogout() {
    setStatus({ type: "loading", message: "" });
    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setStatus({
        type: "error",
        message: "Unable to log out. Please try again.",
      });
      return;
    }

    const client = createApiClient();
    try {
      await client.logout({
        headers: { "X-CSRF-Token": csrfToken },
      });
      deleteCookie(CSRF_COOKIE_NAME);
      window.location.href = "/login";
    } catch (error) {
      // T06: a 401 on logout is the documented "session already
      // expired" case — clear the local session cookie anyway and
      // route the learner to sign in, matching the same
      // session-expiry mid-flow handler used by the core loop.
      setStatus({
        type: "error",
        message: handleApiError(error, "Unable to log out. Please try again."),
      });
    }
  }

  return (
    <>
      <header className="sticky top-0 z-10 border-b border-neutral-200/90 bg-white/95 pt-[env(safe-area-inset-top)] shadow-[0_1px_0_rgb(255_255_255_/_0.8)] backdrop-blur">
        <div className="mx-auto flex h-16 w-full max-w-[48rem] items-center justify-between px-[var(--spacing-md)] sm:px-[var(--spacing-xl)]">
          <Link
            href="/home"
            className="group inline-flex min-h-11 items-center gap-[var(--spacing-sm)] rounded-md pr-[var(--spacing-sm)]"
          >
            <span
              aria-hidden="true"
              className="grid h-9 w-9 place-items-center rounded-xl bg-primary-700 text-base font-bold text-white shadow-sm transition-transform duration-[var(--duration-fast)] group-hover:-translate-y-0.5"
            >
              V
            </span>
            <span>
              <span className="block text-lg font-bold tracking-tight text-neutral-900">
                VocaNova
              </span>
              <span className="block text-xs font-medium text-neutral-500">
                practical English
              </span>
            </span>
          </Link>
          <div className="flex items-center gap-[var(--spacing-xs)]">
            <Link
              href="/settings"
              aria-label="Settings"
              aria-current={pathname === "/settings" ? "page" : undefined}
              className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-xl px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-neutral-700 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-50 hover:text-primary-800"
            >
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                className="h-5 w-5 fill-none stroke-current stroke-[1.8]"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M12 15.25A3.25 3.25 0 1 0 12 8.75a3.25 3.25 0 0 0 0 6.5Z"
                />
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="m19.4 15 .05.05 1.1 1.1-2.05 3.55-1.5-.5a7.6 7.6 0 0 1-1.7 1l-.3 1.55h-4.1l-.3-1.55a7.6 7.6 0 0 1-1.7-1l-1.5.5-2.05-3.55 1.1-1.1.05-.05a7.1 7.1 0 0 1 0-2l-.05-.05-1.1-1.1L7.35 8.3l1.5.5a7.6 7.6 0 0 1 1.7-1l.3-1.55h4.1l.3 1.55a7.6 7.6 0 0 1 1.7 1l1.5-.5 2.05 3.55-1.1 1.1-.05.05a7.1 7.1 0 0 1 0 2Z"
                />
              </svg>
            </Link>
            <button
              type="button"
              onClick={handleLogout}
              disabled={status.type === "loading"}
              aria-busy={status.type === "loading"}
              className="min-h-11 rounded-xl px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-sm font-semibold text-neutral-600 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-100 hover:text-neutral-900 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {status.type === "loading" ? "Signing out..." : "Log out"}
            </button>
          </div>
        </div>
      </header>
      {status.message ? (
        <p
          role="alert"
          aria-live="polite"
          className="border-b border-red-200 bg-red-50 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-center text-sm text-red-700"
        >
          {status.message}
        </p>
      ) : null}
    </>
  );
}
