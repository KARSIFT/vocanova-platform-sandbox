import type { Metadata } from "next";

import { MagicLinkForm, OAuthButton } from "./_components/auth-forms";

export const metadata: Metadata = {
  title: "Sign in — Vocanova",
  description: "Sign in to Vocanova with email or Google.",
};

interface SignInPageProps {
  searchParams: Promise<{ returnTo?: string }>;
}

export default async function SignInPage({ searchParams }: SignInPageProps) {
  const { returnTo } = await searchParams;
  const safeReturnTo = normalizeReturnTo(returnTo);

  return (
    <main className="grid min-h-screen place-items-center p-6">
      <div className="w-full max-w-md space-y-[var(--spacing-lg)] rounded-xl border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm">
        <div className="space-y-[var(--spacing-xs)]">
          <h1 className="text-2xl font-semibold text-neutral-900">
            Sign in to Vocanova
          </h1>
          <p className="text-base text-neutral-700">
            Choose a sign-in method to continue.
          </p>
        </div>

        <OAuthButton returnTo={safeReturnTo} />

        <div className="relative flex items-center gap-[var(--spacing-sm)]">
          <div className="h-px flex-1 bg-neutral-200" />
          <span className="text-sm text-neutral-500">or</span>
          <div className="h-px flex-1 bg-neutral-200" />
        </div>

        <MagicLinkForm returnTo={safeReturnTo} />
      </div>
    </main>
  );
}

function normalizeReturnTo(value?: string): string {
  if (!value || typeof value !== "string") {
    return "/home";
  }
  const trimmed = value.trim();
  if (!trimmed.startsWith("/")) {
    return "/home";
  }
  // Reject URLs that include a scheme or host to prevent open redirects.
  if (trimmed.includes("://") || trimmed.includes("//")) {
    return "/home";
  }
  return trimmed;
}
