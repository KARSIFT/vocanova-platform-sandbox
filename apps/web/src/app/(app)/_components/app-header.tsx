"use client";

import { Suspense, useState, useSyncExternalStore } from "react";
import Link from "next/link";
import Form from "next/form";
import { usePathname, useSearchParams } from "next/navigation";

import { createApiClient } from "@/lib/api";
import { getAuthErrorMessage } from "@/lib/auth-feedback";
import { CSRF_COOKIE_NAME, deleteCookie } from "@/lib/cookies";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { clearOAuthContinuation } from "@/lib/oauth-continuation";
import { handleSessionExpired, isSessionExpiredError } from "@/lib/session";
import { BrandMark, BrandSymbol } from "@/ui/brand-mark";

import {
  clearSentenceFeedbackDrafts,
  hasSentenceFeedbackDrafts,
} from "./sentence-feedback-drafts";
import { isPrimaryNavItemActive } from "./bottom-nav-state";
import { HomeIcon, JourneyIcon, ProgressIcon } from "./nav-icons";
import { getSessionExit, SESSION_LOCATION_CHANGED } from "./session-route";

const DESKTOP_NAV_ITEMS = [
  ["/home", "Home", HomeIcon],
  ["/discover", "Journey", JourneyIcon],
  ["/progress", "Progress", ProgressIcon],
] as const;

const sessionExitClass =
  "inline-flex min-h-11 items-center gap-2 rounded-full px-3 text-sm font-semibold text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700";

function subscribeSessionLocation(onChange: () => void) {
  window.addEventListener(SESSION_LOCATION_CHANGED, onChange);
  window.addEventListener("popstate", onChange);
  return () => {
    window.removeEventListener(SESSION_LOCATION_CHANGED, onChange);
    window.removeEventListener("popstate", onChange);
  };
}

function browserSearchSnapshot() {
  return window.location.search;
}

/** This leaf observes owned history writes even before router effects install. */
function DiscoveryExitLink() {
  const search = useSearchParams();
  const serverSearch = search.size ? `?${search.toString()}` : "";
  const currentSearch = useSyncExternalStore(
    subscribeSessionLocation,
    browserSearchSnapshot,
    () => serverSearch,
  );
  const exit = getSessionExit("/discovery", currentSearch)!;
  return (
    <Link href={exit.href} prefetch={false} className={sessionExitClass}>
      <span aria-hidden="true">×</span>
      {exit.label}
    </Link>
  );
}

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

  const sessionExit = getSessionExit(pathname);
  if (sessionExit) {
    return (
      <header className="app-header">
        <div className="mx-auto flex min-h-14 max-w-[52rem] items-center justify-between gap-4 px-4">
          <span className="inline-flex items-center gap-2 text-sm font-bold text-neutral-900">
            <BrandSymbol size="sm" />
            VocaNova
          </span>
          {pathname === "/discovery" ? (
            <Suspense
              fallback={
                <Link
                  href={sessionExit.href}
                  prefetch={false}
                  className={sessionExitClass}
                >
                  <span aria-hidden="true">×</span>
                  {sessionExit.label}
                </Link>
              }
            >
              <DiscoveryExitLink />
            </Suspense>
          ) : (
            <Link href={sessionExit.href} className={sessionExitClass}>
              <span aria-hidden="true">×</span>
              {sessionExit.label}
            </Link>
          )}
        </div>
      </header>
    );
  }

  return (
    <>
      <header className="app-header">
        <div
          className={`app-header-inner ${pathname === "/vocabulary" ? "app-header-search-page" : ""}`}
        >
          <Link
            href="/home"
            className="inline-flex min-h-11 items-center gap-[var(--spacing-sm)] rounded-lg pr-[var(--spacing-sm)]"
          >
            <BrandMark />
          </Link>
          {pathname !== "/vocabulary" && (
            <Form
              action="/vocabulary"
              role="search"
              aria-label="Word search"
              className="header-search"
              key={pathname}
            >
              <label htmlFor="header-word-query" className="sr-only">
                Search any English word
              </label>
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                className="size-5 shrink-0 fill-none stroke-current stroke-[1.8]"
              >
                <circle cx="10.5" cy="10.5" r="6.5" />
                <path strokeLinecap="round" d="m16 16 4.5 4.5" />
              </svg>
              <input
                id="header-word-query"
                name="q"
                type="search"
                placeholder="Search any word…"
                maxLength={100}
                autoComplete="off"
                spellCheck={false}
              />
              <button type="submit" aria-label="Search words">
                <svg
                  aria-hidden="true"
                  viewBox="0 0 24 24"
                  className="size-5 fill-none stroke-current stroke-[1.8]"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M5 12h14m-5-5 5 5-5 5"
                  />
                </svg>
              </button>
            </Form>
          )}
          <nav
            aria-label="Primary"
            className="header-primary hidden items-center gap-0.5 lg:flex"
          >
            {DESKTOP_NAV_ITEMS.map(([href, label, Icon]) => {
              const active = isPrimaryNavItemActive(pathname, href);
              return (
                <Link
                  key={href}
                  href={href}
                  aria-current={active ? "page" : undefined}
                >
                  <Icon />
                  {label}
                </Link>
              );
            })}
          </nav>
          <div className="header-account flex items-center gap-0.5">
            <Link
              href="/settings"
              aria-label="Settings"
              aria-current={
                pathname === "/settings" || pathname.startsWith("/settings/")
                  ? "page"
                  : undefined
              }
              className="header-icon-button"
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
              className="header-icon-button"
            >
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                className="size-[1.1rem] fill-none stroke-current stroke-[1.8]"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M14 4.75h3.25a2 2 0 0 1 2 2v10.5a2 2 0 0 1-2 2H14M10 16l4-4-4-4m4 4H4.75"
                />
              </svg>
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
