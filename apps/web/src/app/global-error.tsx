"use client";

import NextError from "next/error";
import { useEffect } from "react";

import { captureError } from "@/lib/monitoring";

type GlobalErrorProps = {
  error: Error & { digest?: string };
};

export default function GlobalError({ error }: GlobalErrorProps) {
  useEffect(() => {
    captureError(error);
  }, [error]);

  return (
    <html lang="en">
      <body>
        <NextError statusCode={500} />
      </body>
    </html>
  );
}
