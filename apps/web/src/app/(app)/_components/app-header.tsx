"use client";

import { useState } from "react";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, deleteCookie, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

export function AppHeader() {
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
      window.location.href = "/signin";
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
    <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b border-neutral-200 bg-white px-[var(--spacing-md)]">
      <span className="text-lg font-semibold text-neutral-900">Vocanova</span>
      <div className="flex items-center gap-[var(--spacing-sm)]">
        {status.message ? (
          <p role="alert" aria-live="polite" className="text-sm text-red-700">
            {status.message}
          </p>
        ) : null}
        <button
          type="button"
          onClick={handleLogout}
          disabled={status.type === "loading"}
          aria-busy={status.type === "loading"}
          className="min-h-[var(--spacing-xl)] rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-sm font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {status.type === "loading" ? "Signing out..." : "Log out"}
        </button>
      </div>
    </header>
  );
}
