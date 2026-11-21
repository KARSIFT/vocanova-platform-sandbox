import type { Metadata } from "next";

import { connectionRetryDestination } from "@/lib/connection-error";
import { AuthShell } from "@/ui/auth-shell";
import { actionClasses } from "@/ui/controls";

export const metadata: Metadata = {
  title: "Connection interrupted — Vocanova",
  robots: { index: false, follow: false },
};

export default async function ConnectionErrorPage({
  searchParams,
}: {
  searchParams: Promise<{ returnTo?: string }>;
}) {
  const { returnTo } = await searchParams;
  const safeDestination = connectionRetryDestination(returnTo);

  return (
    <AuthShell
      title="We couldn't connect"
      description="Vocanova is temporarily unavailable. Try again in a moment to continue."
    >
      {/* A full navigation rechecks the session rather than reusing a cached
            failed route transition. No private learner data is rendered here. */}
      <a
        href={safeDestination}
        className={actionClasses("primary", "justify-self-start")}
      >
        Try again
      </a>
    </AuthShell>
  );
}
