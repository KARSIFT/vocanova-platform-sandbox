import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

import { getApiBaseURL } from "./lib/env";

export const config = {
  matcher: [
    "/home",
    "/discover",
    "/discover/:path*",
    "/progress",
    "/reviews",
    "/reviews/:path*",
  ],
};

export async function middleware(request: NextRequest): Promise<NextResponse> {
  const returnTo = new URLSearchParams({
    returnTo: `${request.nextUrl.pathname}${request.nextUrl.search}`,
  }).toString();
  const signInUrl = new URL(`/signin?${returnTo}`, request.url);

  const apiBaseURL = getApiBaseURL();
  const cookieHeader = request.headers.get("cookie") ?? "";

  try {
    const response = await fetch(`${apiBaseURL}/api/v1/me`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        Cookie: cookieHeader,
      },
      credentials: "include",
    });

    if (response.status === 401) {
      return NextResponse.redirect(signInUrl);
    }

    if (!response.ok) {
      return NextResponse.redirect(signInUrl);
    }
  } catch {
    return NextResponse.redirect(signInUrl);
  }

  return NextResponse.next();
}
