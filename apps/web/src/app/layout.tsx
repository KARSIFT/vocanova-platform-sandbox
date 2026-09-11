import type { Metadata } from "next";
import type { ReactNode } from "react";

import "./globals.css";

import { ThemeBootstrap, ThemeProvider } from "./_components/theme-preference";

export const metadata: Metadata = {
  title: "Vocanova",
  description:
    "Vocanova is an AI-powered platform for practical English learning.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <ThemeBootstrap />
      </head>
      <body>
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
