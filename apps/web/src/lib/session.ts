"use client";

import { ApiResponseError } from "@vocanova/api-client";

import { CSRF_COOKIE_NAME, deleteCookie, SESSION_COOKIE_NAME } from "./cookies";

/**
 * isSessionExpiredError reports whether the supplied error indicates the
 * learner's session has expired or is otherwise no longer accepted by the
 * server.
 *
 * Every client component that makes an authenticated request MUST run this
 * check on its catch path so the T06 "session-expiry mid-flow" cross-cutting
 * guarantee holds: when a 401 comes back mid-review-session,
 * mid-sentence-submission, mid-onboarding, mid-settings-write, or
 * mid-account-deletion, the learner is consistently routed to re-auth
 * instead of seeing a generic "try again" message that hides the real
 * problem.
 */
export function isSessionExpiredError(error: unknown): boolean {
  return error instanceof ApiResponseError && error.status === 401;
}

/**
 * handleSessionExpired is the single client-side entry point for the
 * session-expiry mid-flow handler. It clears the local session and CSRF
 * cookies (so a stale request cannot be replayed against the same identity)
 * and routes the learner to /signin with the current page as returnTo, so
 * after re-authentication the learner lands back where they were.
 *
 * Components must NOT catch a 401 silently and continue: the cross-cutting
 * property T06 guarantees is that the learner is never left looking at a
 * screen that claims an action succeeded when the server rejected it, and
 * is never shown a generic error in place of a clear "please sign in again"
 * affordance.
 */
export function handleSessionExpired(currentPath?: string): void {
  if (typeof window === "undefined") {
    return;
  }
  deleteCookie(SESSION_COOKIE_NAME);
  deleteCookie(CSRF_COOKIE_NAME);
  const returnTo =
    currentPath ?? `${window.location.pathname}${window.location.search}`;
  const params = new URLSearchParams({ returnTo });
  window.location.href = `/signin?${params.toString()}`;
}

/**
 * handleApiError is the T06 cross-cutting catch-path helper for every
 * authenticated client request. It detects a 401 and routes the learner to
 * re-authentication; for every other error it returns the error to the
 * caller so the existing per-screen error display can render a stable
 * message.
 *
 * Usage:
 *
 *   try {
 *     await client.someCall(...);
 *   } catch (error) {
 *     handleApiError(error, "Unable to save your changes. Please try again.");
 *   }
 */
export function handleApiError(
  error: unknown,
  fallbackMessage: string,
): string {
  if (isSessionExpiredError(error)) {
    handleSessionExpired();
    // The redirect is in flight; the return value is best-effort for the
    // brief moment before the navigation completes.
    return "Your session expired. Redirecting you to sign in...";
  }
  if (error instanceof ApiResponseError) {
    return error.message;
  }
  return fallbackMessage;
}
