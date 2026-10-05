import type { Metadata } from "next";

import { normalizeReturnTo } from "@/lib/return-to";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

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
  const destination = normalizeReturnTo(returnTo);
  const retryPath = new URL(destination, "https://vocanova.invalid").pathname;
  const safeDestination =
    retryPath === "/connection-error" ? "/home" : destination;

  return (
    <AuthShell>
      <Surface className="space-y-5">
        <h1 className="text-2xl font-semibold text-neutral-900">
          We couldn&apos;t connect
        </h1>
        <p className="text-base leading-7 text-neutral-700">
          Vocanova is temporarily unavailable. Try again in a moment to
          continue.
        </p>
        {/* A full navigation rechecks the session rather than reusing a cached
            failed route transition. No private learner data is rendered here. */}
        <a
          href={safeDestination}
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Try again
        </a>
      </Surface>
    </AuthShell>
  );
}
