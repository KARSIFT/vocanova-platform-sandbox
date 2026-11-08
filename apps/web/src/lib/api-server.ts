import { headers } from "next/headers";
import { redirect } from "next/navigation";

import { ApiResponseError, VocanovaClient } from "@vocanova/api-client";

import { filterForwardedCookieHeader } from "./cookies";
import { getApiBaseURL } from "./env";
import { fetchWithServerReadDeadline } from "./server-read-fetch";

export async function createServerApiClient(): Promise<VocanovaClient> {
  const incomingHeaders = await headers();
  // Only this tier's own credentials reach the API; see cookies.ts.
  const cookieHeader = filterForwardedCookieHeader(
    incomingHeaders.get("cookie") ?? "",
  );

  return new VocanovaClient({
    baseURL: getApiBaseURL(),
    fetch: (input: RequestInfo | URL, init?: RequestInit) => {
      const headers = new Headers(init?.headers);
      if (cookieHeader) {
        headers.set("Cookie", cookieHeader);
      }
      return fetchWithServerReadDeadline(input, { ...init, headers });
    },
  });
}

export function requireAuthRedirect(error: unknown, returnTo: string): never {
  if (error instanceof ApiResponseError && error.status === 401) {
    // A protected read can fail after the initial identity check passed.
    // Mark the recovery route so an existing cookie cannot bounce back to it.
    const searchParams = new URLSearchParams({
      returnTo,
      reason: "session-expired",
    });
    redirect(`/login?${searchParams.toString()}`);
  }
  throw error;
}
