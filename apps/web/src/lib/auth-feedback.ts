import { ApiResponseError } from "@vocanova/api-client";

export type AuthErrorContext =
  "magic-request" | "magic-consume" | "oauth-start" | "logout";

/**
 * Keeps API implementation details out of learner-facing authentication
 * screens. The status codes are stable contract signals; response bodies are
 * intentionally not repeated because they can be terse operational strings.
 */
export function getAuthErrorMessage(
  error: unknown,
  context: AuthErrorContext,
): string {
  const status = error instanceof ApiResponseError ? error.status : undefined;

  if (context === "magic-request") {
    if (status === 429) {
      return "Too many sign-in links were requested. Please wait a few minutes, then try again.";
    }
    if (status === 503) {
      return "Email sign-in is unavailable right now. Please try again later.";
    }
    return "We couldn't send a sign-in link. Check the address and try again.";
  }

  if (context === "magic-consume") {
    if (status === 401) {
      return "This sign-in link is invalid, expired, or has already been used. Request a new link.";
    }
    if (status === 429) {
      return "Too many sign-in attempts were made. Please wait a few minutes, then request a new link.";
    }
    if (status === 503) {
      return "Email sign-in is unavailable right now. Please try again later.";
    }
    return "We couldn't verify this sign-in link. Request a new one and try again.";
  }

  if (context === "oauth-start") {
    if (status === 429) {
      return "Too many Google sign-in attempts were made. Please wait a few minutes, then try again.";
    }
    if (status === 404 || status === 503) {
      return "Google sign-in is unavailable right now. You can use email instead.";
    }
    return "We couldn't start Google sign-in. Please try again or use email instead.";
  }

  if (status === 403) {
    return "This page needs to be refreshed before you can sign out securely. Please refresh and try again.";
  }
  if (status === 429) {
    return "Too many sign-out attempts were made. Please wait a few minutes, then try again.";
  }
  return "We couldn't sign you out. Please try again.";
}

export function getOAuthCallbackMessage(value?: string): string | null {
  switch (value) {
    case "cancelled":
      return "Google sign-in was cancelled. You can try again or use email instead.";
    case "expired":
      return "This Google sign-in attempt expired or could not be verified. Please try again.";
    case "unavailable":
      return "Google sign-in is unavailable right now. You can use email instead.";
    case "failed":
      return "Google could not complete sign-in. Please try again or use email instead.";
    default:
      return null;
  }
}
