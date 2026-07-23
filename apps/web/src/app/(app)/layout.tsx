import type { ReactNode } from "react";

import { BottomNav } from "./_components/bottom-nav";

export default function AppShellLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <>
      <main className="min-h-screen pb-16">{children}</main>
      <BottomNav />
    </>
  );
}
