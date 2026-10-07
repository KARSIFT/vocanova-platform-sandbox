import { headers } from "next/headers";
import { redirect } from "next/navigation";

import { ApiResponseError, VocanovaClient } from "@vocanova/api-client";

import { getApiBaseURL } from "./env";
import { fetchWithServerReadDeadline } from "./server-read-fetch";

export async function createServerApiClient(): Promise<VocanovaClient> {
  const incomingHeaders = await headers();
  const cookieHeader = incomingHeaders.get("cookie") ?? "";

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
    const searchParams = new URLSearchParams();
    searchParams.set("returnTo", returnTo);
    redirect(`/login?${searchParams.toString()}`);
  }
  throw error;
}
