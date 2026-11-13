"use client";

import { Suspense } from "react";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

import { MagicLinkPageContent } from "./_components/magic-link-page-content";

export default function MagicLinkPage() {
  return (
    <Suspense
      fallback={
        <AuthShell>
          <Surface className="w-full space-y-[var(--spacing-md)]">
            <h1 className="text-2xl font-semibold text-neutral-900">
              Sign in link
            </h1>
            <p className="text-base text-neutral-700">
              Verifying your sign-in link...
            </p>
          </Surface>
        </AuthShell>
      }
    >
      <MagicLinkPageContent />
    </Suspense>
  );
}
