import type { Metadata } from "next";
import Link from "next/link";

import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
import { getOAuthCallbackMessage } from "@/lib/auth-feedback";
import { normalizeReturnTo } from "@/lib/return-to";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

import { MagicLinkForm, OAuthButton } from "./_components/auth-forms";
import { PasswordLoginForm } from "../auth/password/_components/password-forms";

export const metadata: Metadata = {
  title: "Sign in — Vocanova",
  description: "Sign in to Vocanova.",
};

interface SignInPageProps {
  searchParams: Promise<{
    magicOnly?: string;
    oauth?: string;
    reason?: string;
    returnTo?: string;
    signedOut?: string;
  }>;
}

export default async function SignInPage({ searchParams }: SignInPageProps) {
  const { magicOnly, oauth, reason, returnTo, signedOut } = await searchParams;
  const safeReturnTo = normalizeReturnTo(returnTo);
  const capabilities = await getSignInAuthCapabilities();
  const { magicLinkEnabled, oauthEnabled, passwordEnabled } = capabilities;
  const magicOnlyUnavailable = magicOnly === "1" && !magicLinkEnabled;
  const showPassword = passwordEnabled && magicOnly !== "1";
  const showOAuth = oauthEnabled && (magicOnly !== "1" || magicOnlyUnavailable);
  const oauthMessage = getOAuthCallbackMessage(oauth, capabilities);

  return (
    <AuthShell>
      <Surface className="w-full space-y-[var(--spacing-lg)] border-neutral-200 bg-white shadow-[0_12px_28px_rgb(15_23_42_/_0.07)]">
        <div className="space-y-[var(--spacing-sm)]">
          <h1 className="text-xl font-semibold text-neutral-900">
            Sign in to Vocanova
          </h1>
          <p className="text-base text-neutral-700">
            {magicOnly === "1" && !magicOnlyUnavailable
              ? "Enter your email to continue securely."
              : showPassword
                ? showOAuth || magicLinkEnabled
                  ? "Sign in with your email and password, or choose another secure method."
                  : "Sign in with your email and password."
                : showOAuth
                  ? "Continue with Google to sign in."
                  : magicLinkEnabled
                    ? "Enter your email to continue securely."
                    : "Sign in to continue your vocabulary practice."}
          </p>
        </div>

        {signedOut === "1" ? (
          <p
            role="status"
            aria-live="polite"
            className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-sm)] text-base text-primary-900"
          >
            You&apos;re signed out.
          </p>
        ) : null}

        {signedOut === "expired" ? (
          <p
            role="status"
            aria-live="polite"
            className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-sm)] text-base text-primary-900"
          >
            Your previous session had already expired. You can sign in again.
          </p>
        ) : null}

        {reason === "session-expired" ? (
          <p
            role="status"
            aria-live="polite"
            className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-sm)] text-base text-primary-900"
          >
            Your session expired. Sign in again to continue.
          </p>
        ) : null}

        {oauthMessage ? (
          <p
            role="alert"
            aria-live="assertive"
            className="rounded-md border border-red-200 bg-red-50 p-[var(--spacing-sm)] text-base text-red-800"
          >
            {oauthMessage}
          </p>
        ) : null}

        {magicOnlyUnavailable ? (
          <p
            role="status"
            aria-live="polite"
            className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-sm)] text-base text-primary-900"
          >
            Email sign-in links are unavailable right now.{" "}
            {showOAuth ? "You can continue with Google below. " : null}
            {passwordEnabled ? (
              <>
                For email and password,{" "}
                <Link
                  href={`/login?${new URLSearchParams({ returnTo: safeReturnTo }).toString()}`}
                  className="font-semibold underline"
                >
                  use the standard sign-in page
                </Link>
                .
              </>
            ) : !showOAuth ? (
              "Please try again later."
            ) : null}
          </p>
        ) : null}

        {showPassword ? <PasswordLoginForm returnTo={safeReturnTo} /> : null}

        {showPassword && (showOAuth || magicLinkEnabled) ? (
          <div className="relative flex items-center gap-[var(--spacing-sm)]">
            <div className="h-px flex-1 bg-neutral-200" />
            <span className="text-sm text-neutral-500">
              or use another method
            </span>
            <div className="h-px flex-1 bg-neutral-200" />
          </div>
        ) : null}

        {showOAuth ? (
          <>
            <OAuthButton returnTo={safeReturnTo} capabilities={capabilities} />

            {!showPassword && magicLinkEnabled ? (
              <div className="relative flex items-center gap-[var(--spacing-sm)]">
                <div className="h-px flex-1 bg-neutral-200" />
                <span className="text-sm text-neutral-500">or</span>
                <div className="h-px flex-1 bg-neutral-200" />
              </div>
            ) : null}
          </>
        ) : null}

        {magicLinkEnabled ? (
          <MagicLinkForm
            returnTo={safeReturnTo}
            emailLabel={passwordEnabled ? "Email for sign-in link" : undefined}
          />
        ) : null}

        {!passwordEnabled && !magicLinkEnabled && !showOAuth ? (
          <p
            role="alert"
            aria-live="assertive"
            className="rounded-md border border-red-200 bg-red-50 p-[var(--spacing-sm)] text-base text-red-800"
          >
            Sign-in is temporarily unavailable. Please try again later.
          </p>
        ) : null}
        <Link
          href="/help"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 underline underline-offset-4"
        >
          Help and support
        </Link>
      </Surface>
    </AuthShell>
  );
}
