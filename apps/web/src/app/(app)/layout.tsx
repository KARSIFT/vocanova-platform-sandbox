import type { ReactNode } from "react";

import { AppHeader } from "./_components/app-header";
import { BottomNav } from "./_components/bottom-nav";

export default function AppShellLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <>
      <AppHeader />
      <main className="min-h-screen pb-16 pt-14">{children}</main>
      <BottomNav />
    </>
  );
}
