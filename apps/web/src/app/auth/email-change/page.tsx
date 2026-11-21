"use client";

import { Suspense } from "react";
import { AuthShell } from "@/ui/auth-shell";

import { EmailChangePageContent } from "./_components/email-change-page-content";

export default function EmailChangePage() {
  return (
    <Suspense
      fallback={
        <AuthShell title="Confirm your new email" linkBrand={false}>
          <p role="status" className="text-base text-neutral-700">
            Verifying your confirmation link...
          </p>
        </AuthShell>
      }
    >
      <EmailChangePageContent />
    </Suspense>
  );
}
