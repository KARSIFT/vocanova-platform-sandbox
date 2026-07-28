"use client";

import { Suspense } from "react";

import { MagicLinkPageContent } from "./_components/magic-link-page-content";

export default function MagicLinkPage() {
  return (
    <Suspense
      fallback={
        <main className="grid min-h-screen place-items-center p-6">
          {/* max-w-[28rem] (not max-w-md): see the token-collision note on
              /onboarding's page.tsx - tokens.generated.css's --spacing-md
              (16px) shadows the intended 28rem max-w-md container size. */}
          <div className="w-full max-w-[28rem] rounded-xl border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm">
            <h1 className="text-2xl font-semibold text-neutral-900">
              Sign in link
            </h1>
            <p className="mt-[var(--spacing-md)] text-base text-neutral-700">
              Verifying your sign-in link...
            </p>
          </div>
        </main>
      }
    >
      <MagicLinkPageContent />
    </Suspense>
  );
}
