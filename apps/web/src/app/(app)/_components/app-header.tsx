"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { createApiClient } from "@/lib/api";
import { getAuthErrorMessage } from "@/lib/auth-feedback";
import { CSRF_COOKIE_NAME, deleteCookie } from "@/lib/cookies";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { clearOAuthContinuation } from "@/lib/oauth-continuation";
import { handleSessionExpired, isSessionExpiredError } from "@/lib/session";
import { BrandMark } from "@/ui/brand-mark";

import {
  clearSentenceFeedbackDrafts,
  hasSentenceFeedbackDrafts,
} from "./sentence-feedback-drafts";
import { isPrimaryNavItemActive } from "./bottom-nav-state";

const DESKTOP_NAV_ITEMS = [
  ["/home", "Home"],
  ["/discover", "Journey"],
  ["/progress", "Progress"],
] as const;

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
    if (
      hasSentenceFeedbackDrafts() &&
      !window.confirm(
        "You have an unsent practice draft in this tab. Signing out will discard it. Do you want to sign out?",
      )
    ) {
      return;
    }

    setStatus({ type: "loading", message: "" });
    const client = createApiClient();
    try {
      const csrfToken = await getOrRefreshCSRFToken();
      if (!csrfToken) {
        setStatus({
          type: "error",
          message: "We couldn't prepare a secure sign-out. Please try again.",
        });
        return;
      }
      await client.logout({
        headers: { "X-CSRF-Token": csrfToken },
      });
      clearSentenceFeedbackDrafts();
      clearOAuthContinuation();
      deleteCookie(CSRF_COOKIE_NAME);
      window.location.assign("/login?signedOut=1");
    } catch (error) {
      // A recovery GET /me that confirms the session is no longer accepted
      // follows the normal re-authentication path. Drafts remain until a
      // server-side logout actually succeeds.
      if (isSessionExpiredError(error)) {
        handleSessionExpired();
        return;
      }
      setStatus({
        type: "error",
        message: getAuthErrorMessage(error, "logout"),
      });
    }
  }

  return (
    <>
      <header className="sticky top-0 z-10 border-b border-neutral-200/90 bg-white/95 pt-[env(safe-area-inset-top)] backdrop-blur">
        <div className="mx-auto flex min-h-16 w-full max-w-[76rem] items-center gap-[var(--spacing-sm)] px-[var(--spacing-md)] sm:px-[var(--spacing-xl)]">
          <Link
            href="/home"
            className="inline-flex min-h-11 items-center gap-[var(--spacing-sm)] rounded-md pr-[var(--spacing-sm)]"
          >
            <BrandMark />
          </Link>
          <nav
            aria-label="Primary"
            className="ml-auto hidden items-center gap-1 lg:flex"
          >
            {DESKTOP_NAV_ITEMS.map(([href, label]) => {
              const active = isPrimaryNavItemActive(pathname, href);
              return (
                <Link
                  key={href}
                  href={href}
                  aria-current={active ? "page" : undefined}
                  className={`inline-flex min-h-11 items-center rounded-md px-3 text-sm font-semibold transition-colors ${active ? "bg-primary-50 text-primary-800" : "text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900"}`}
                >
                  {label}
                </Link>
              );
            })}
          </nav>
          <div className="ml-auto flex items-center gap-[var(--spacing-xs)] lg:ml-4">
            <Link
              href="/settings"
              aria-label="Settings"
              aria-current={
                pathname === "/settings" || pathname.startsWith("/settings/")
                  ? "page"
                  : undefined
              }
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
