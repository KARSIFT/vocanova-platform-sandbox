import type { ReactNode } from "react";

import { NavigationProgress } from "@/ui/navigation-feedback";

import { AppHeader } from "./_components/app-header";
import { BottomNav } from "./_components/bottom-nav";
import { CSRFRecovery } from "./_components/csrf-recovery";
import { OAuthContinuation } from "./_components/oauth-continuation";

export default function AppShellLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <>
      <a
        href="#main-content"
        className="fixed left-[var(--spacing-md)] -top-full z-30 inline-flex min-h-11 items-center rounded-md bg-primary-700 px-[var(--spacing-md)] py-[var(--spacing-xs)] text-sm font-medium text-neutral-50 focus:top-[var(--spacing-md)] focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-700"
      >
        Skip to main content
      </a>
      <NavigationProgress />
      <OAuthContinuation />
      <CSRFRecovery />
      <AppHeader />
      <main
        id="main-content"
        tabIndex={-1}
        className="min-h-[calc(100dvh_-_3.5rem_-_env(safe-area-inset-top))] pb-[calc(5rem+env(safe-area-inset-bottom))] lg:min-h-[calc(100dvh_-_4.5rem)] lg:pb-0"
      >
        {children}
      </main>
      <BottomNav />
    </>
  );
}
