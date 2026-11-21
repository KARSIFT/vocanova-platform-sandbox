"use client";

import { Suspense } from "react";
import { AuthShell } from "@/ui/auth-shell";

import { MagicLinkPageContent } from "./_components/magic-link-page-content";

export default function MagicLinkPage() {
  return (
    <Suspense
      fallback={
        <AuthShell title="Sign in link" linkBrand={false}>
          <p role="status" className="text-base text-neutral-700">
            Verifying your sign-in link...
          </p>
        </AuthShell>
      }
    >
      <MagicLinkPageContent />
    </Suspense>
  );
}
