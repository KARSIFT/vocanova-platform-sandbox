import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

import { getApiBaseURL } from "./lib/env";

export const config = {
  matcher: [
    "/onboarding",
    "/home",
    "/discover",
    "/discover/:path*",
    "/progress",
    "/reviews",
    "/reviews/:path*",
    "/settings",
    "/settings/:path*",
  ],
};

interface CurrentUserResponse {
  onboardingStatus?: "not_started" | "in_progress" | "completed";
}

export async function middleware(request: NextRequest): Promise<NextResponse> {
  const returnTo = new URLSearchParams({
    returnTo: `${request.nextUrl.pathname}${request.nextUrl.search}`,
  }).toString();
  const signInUrl = new URL(`/signin?${returnTo}`, request.url);

  const apiBaseURL = getApiBaseURL();
  const cookieHeader = request.headers.get("cookie") ?? "";

  let meResponse: Response;
  try {
    meResponse = await fetch(`${apiBaseURL}/api/v1/me`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        Cookie: cookieHeader,
      },
      credentials: "include",
    });
  } catch {
    return NextResponse.redirect(signInUrl);
  }

  if (meResponse.status === 401) {
    return NextResponse.redirect(signInUrl);
  }
  if (!meResponse.ok) {
    return NextResponse.redirect(signInUrl);
  }

  // Parse the additive onboardingStatus field introduced by
  // VOC-031-T01. Treat a missing/malformed field as "not_started" so
  // the conservative gate wins (any authenticated learner who
  // hasn't been grandfathered past onboarding is funneled through it).
  let onboardingStatus: "not_started" | "in_progress" | "completed" =
    "not_started";
  try {
    const me = (await meResponse.json()) as CurrentUserResponse;
    if (
      me.onboardingStatus === "not_started" ||
      me.onboardingStatus === "in_progress" ||
      me.onboardingStatus === "completed"
    ) {
      onboardingStatus = me.onboardingStatus;
    }
  } catch {
    // fall through with the conservative "not_started" default
  }

  const isOnboardingRoute = request.nextUrl.pathname === "/onboarding";
  if (onboardingStatus !== "completed" && !isOnboardingRoute) {
    const onboardingUrl = new URL("/onboarding", request.url);
    return NextResponse.redirect(onboardingUrl);
  }
  if (onboardingStatus === "completed" && isOnboardingRoute) {
    const homeUrl = new URL("/home", request.url);
    return NextResponse.redirect(homeUrl);
  }

  return NextResponse.next();
}
