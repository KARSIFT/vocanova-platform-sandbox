import type { ReactNode } from "react";

import { AppHeader } from "./_components/app-header";
import { BottomNav } from "./_components/bottom-nav";

export default function AppShellLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <>
      <a
        href="#main-content"
        className="fixed left-[var(--spacing-md)] -top-full z-20 inline-flex min-h-11 items-center rounded-md bg-primary-700 px-[var(--spacing-md)] py-[var(--spacing-xs)] text-sm font-medium text-neutral-50 focus:top-[var(--spacing-md)] focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-700"
      >
        Skip to main content
      </a>
      <AppHeader />
      <main
        id="main-content"
        tabIndex={-1}
        className="min-h-screen pb-16 pt-14"
      >
        {children}
      </main>
      <BottomNav />
    </>
  );
}
