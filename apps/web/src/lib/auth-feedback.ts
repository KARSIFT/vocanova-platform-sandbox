import { ApiResponseError } from "@vocanova/api-client";

import type { SignInAuthCapabilities } from "./auth-capabilities";

export type AuthErrorContext =
  | "magic-request"
  | "magic-consume"
  | "oauth-start"
  | "logout"
  | "password-login"
  | "password-signup"
  | "password-reset-request"
  | "password-reset";

function withAvailableEmailMethod(
  message: string,
  capabilities?: SignInAuthCapabilities,
): string {
  if (capabilities?.passwordEnabled) {
    return `${message} You can sign in with your email and password.`;
  }
  if (capabilities?.magicLinkEnabled) {
    return `${message} You can request an email sign-in link.`;
  }
  return message;
}

/**
 * Keeps API implementation details out of learner-facing authentication
 * screens. The status codes are stable contract signals; response bodies are
 * intentionally not repeated because they can be terse operational strings.
 */
export function getAuthErrorMessage(
  error: unknown,
  context: AuthErrorContext,
  capabilities?: SignInAuthCapabilities,
): string {
  const status = error instanceof ApiResponseError ? error.status : undefined;

  if (context === "password-login") {
    if (status === 429) {
      return "Too many password sign-in attempts were made. Please wait a few minutes, then try again.";
    }
    if (status === 503) {
      return "Password sign-in is temporarily unavailable. Please try again later.";
    }
    return "We couldn't sign you in with that email and password. Check them and try again.";
  }

  if (context === "password-signup") {
    if (status === 429) {
      return "Too many account requests were made. Please wait a few minutes, then try again.";
    }
    if (status === 503) {
      return "Account creation is temporarily unavailable. Please try again later.";
    }
    return "We couldn't create your account. Check the details and try again.";
  }

  if (context === "password-reset-request") {
    if (status === 429) {
      return "Too many password reset requests were made. Please wait a few minutes, then try again.";
    }
    if (status === 503) {
      return "Password email is temporarily unavailable. Please try again later.";
    }
    return "We couldn't send that email. Check the address and try again.";
  }

  if (context === "password-reset") {
    if (status === 401) {
      return "This password link is invalid, expired, or has already been used. Request a new one.";
    }
    if (status === 429) {
      return "Too many password reset attempts were made. Please wait a few minutes, then try again.";
    }
    if (status === 503) {
      return "Password reset is temporarily unavailable. Please try again later.";
    }
    return "We couldn't save your new password. Check the password and try again.";
  }

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
      return withAvailableEmailMethod(
        "Google sign-in is unavailable right now. Please try again later.",
        capabilities,
      );
    }
    return withAvailableEmailMethod(
      "We couldn't start Google sign-in. Please try again.",
      capabilities,
    );
  }

  if (status === 403) {
    return "This page needs to be refreshed before you can sign out securely. Please refresh and try again.";
  }
  if (status === 429) {
    return "Too many sign-out attempts were made. Please wait a few minutes, then try again.";
  }
  return "We couldn't sign you out. Please try again.";
}

export function getOAuthCallbackMessage(
  value?: string,
  capabilities?: SignInAuthCapabilities,
): string | null {
  switch (value) {
    case "cancelled":
      return withAvailableEmailMethod(
        "Google sign-in was cancelled. You can try again.",
        capabilities,
      );
    case "expired":
      return "This Google sign-in attempt expired or could not be verified. Please try again.";
    case "unavailable":
      return withAvailableEmailMethod(
        "Google sign-in is unavailable right now. Please try again later.",
        capabilities,
      );
    case "failed":
      return withAvailableEmailMethod(
        "Google could not complete sign-in. Please try again.",
        capabilities,
      );
    default:
      return null;
  }
}
