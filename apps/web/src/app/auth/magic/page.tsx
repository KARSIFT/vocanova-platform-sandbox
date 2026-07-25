"use client";

import { Suspense } from "react";

import { MagicLinkPageContent } from "./_components/magic-link-page-content";

export default function MagicLinkPage() {
  return (
    <Suspense
      fallback={
        <main className="grid min-h-screen place-items-center p-6">
          <div className="w-full max-w-md rounded-xl border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm">
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
