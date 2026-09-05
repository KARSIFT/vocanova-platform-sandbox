import type { Metadata } from "next";

import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
import { normalizeReturnTo } from "@/lib/return-to";

import { MagicLinkForm, OAuthButton } from "./_components/auth-forms";

export const metadata: Metadata = {
  title: "Sign in — Vocanova",
  description: "Sign in to Vocanova.",
};

interface SignInPageProps {
  searchParams: Promise<{ returnTo?: string }>;
}

export default async function SignInPage({ searchParams }: SignInPageProps) {
  const { returnTo } = await searchParams;
  const safeReturnTo = normalizeReturnTo(returnTo);
  const { oauthEnabled } = await getSignInAuthCapabilities();

  return (
    <main className="grid min-h-screen place-items-center p-6">
      {/* max-w-[28rem] (not max-w-md): this repo's tokens.generated.css only
          defines a --spacing-* scale, so Tailwind resolves the named
          max-w-md utility to --spacing-md (16px) instead of the intended
          28rem, collapsing this card to a single-character column. See
          the matching note on /onboarding's page.tsx. */}
      <div className="w-full max-w-[28rem] space-y-[var(--spacing-lg)] rounded-xl border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm">
        <div className="space-y-[var(--spacing-xs)]">
          <h1 className="text-2xl font-semibold text-neutral-900">
            Sign in to Vocanova
          </h1>
          <p className="text-base text-neutral-700">
            Choose a sign-in method to continue.
          </p>
        </div>

        {oauthEnabled ? (
          <>
            <OAuthButton returnTo={safeReturnTo} />

            <div className="relative flex items-center gap-[var(--spacing-sm)]">
              <div className="h-px flex-1 bg-neutral-200" />
              <span className="text-sm text-neutral-500">or</span>
              <div className="h-px flex-1 bg-neutral-200" />
            </div>
          </>
        ) : null}

        <MagicLinkForm returnTo={safeReturnTo} />
      </div>
    </main>
  );
}
