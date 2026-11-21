"use client";

import { useEffect } from "react";

import { captureError } from "@/lib/monitoring";
import styles from "./global-error.module.css";

type GlobalErrorProps = {
  error: Error & { digest?: string };
};

export default function GlobalError({ error }: GlobalErrorProps) {
  useEffect(() => {
    captureError(error);
  }, [error]);

  return (
    <html lang="en">
      <body className={styles.canvas}>
        <main className={styles.content}>
          <p className={styles.brand}>VocaNova</p>
          <h1>Something went wrong</h1>
          <p>Return to VocaNova to continue learning.</p>
          {/* Force a new document because the root layout may have failed. */}
          {/* eslint-disable-next-line @next/next/no-html-link-for-pages */}
          <a href="/">Back to home</a>
        </main>
      </body>
    </html>
  );
}
