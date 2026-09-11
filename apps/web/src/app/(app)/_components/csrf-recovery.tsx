"use client";

import { useEffect } from "react";

import { getOrRefreshCSRFToken } from "@/lib/csrf";

/** Restores a missing CSRF cookie after the app shell reaches a browser. */
export function CSRFRecovery() {
  useEffect(() => {
    void getOrRefreshCSRFToken().catch(() => {
      // Authenticated mutations handle a failed recovery with their existing
      // session/error paths. Avoid a background request changing page state.
    });
  }, []);

  return null;
}
