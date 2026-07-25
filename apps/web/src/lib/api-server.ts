import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { ApiResponseError, VocanovaClient } from "@vocanova/api-client";

import { getApiBaseURL } from "./env";

export async function createServerApiClient(): Promise<VocanovaClient> {
  const cookieStore = await cookies();
  const cookieHeader = cookieStore.toString();

  return new VocanovaClient({
    baseURL: getApiBaseURL(),
    fetch: (input: RequestInfo | URL, init?: RequestInit) => {
      const headers = new Headers(init?.headers);
      if (cookieHeader) {
        headers.set("Cookie", cookieHeader);
      }
      return fetch(input, { ...init, headers });
    },
  });
}

export function requireAuthRedirect(error: unknown, returnTo: string): never {
  if (error instanceof ApiResponseError && error.status === 401) {
    const searchParams = new URLSearchParams();
    searchParams.set("returnTo", returnTo);
    redirect(`/signin?${searchParams.toString()}`);
  }
  throw error;
}
