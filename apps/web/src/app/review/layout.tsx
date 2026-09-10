import type { ReactNode } from "react";

import AppShellLayout from "../(app)/layout";

export default function DocumentedReviewLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return <AppShellLayout>{children}</AppShellLayout>;
}
