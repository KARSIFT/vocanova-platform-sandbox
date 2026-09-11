"use client";

import { useEffect } from "react";

import { consumeOAuthContinuation } from "@/lib/oauth-continuation";

/** Continues a completed OAuth sign-in from its allowed Home landing route. */
export function OAuthContinuation() {
  useEffect(() => {
    const returnTo = consumeOAuthContinuation();
    if (returnTo && returnTo !== "/home") {
      window.location.replace(returnTo);
    }
  }, []);

  return null;
}
