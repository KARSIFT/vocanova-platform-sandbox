"use client";

import { Suspense } from "react";

import { EmailChangePageContent } from "./_components/email-change-page-content";

export default function EmailChangePage() {
  return (
    <Suspense
      fallback={
        <main className="grid min-h-screen place-items-center p-6">
          <div className="w-full max-w-[28rem] rounded-xl border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm">
            <h1 className="text-2xl font-semibold text-neutral-900">
              Confirm your new email
            </h1>
            <p className="mt-[var(--spacing-md)] text-base text-neutral-700">
              Verifying your confirmation link...
            </p>
          </div>
        </main>
      }
    >
      <EmailChangePageContent />
    </Suspense>
  );
}
