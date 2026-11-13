"use client";

import { Suspense } from "react";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

import { EmailChangePageContent } from "./_components/email-change-page-content";

export default function EmailChangePage() {
  return (
    <Suspense
      fallback={
        <AuthShell>
          <Surface className="w-full space-y-[var(--spacing-md)]">
            <h1 className="text-2xl font-semibold text-neutral-900">
              Confirm your new email
            </h1>
            <p className="text-base text-neutral-700">
              Verifying your confirmation link...
            </p>
          </Surface>
        </AuthShell>
      }
    >
      <EmailChangePageContent />
    </Suspense>
  );
}
