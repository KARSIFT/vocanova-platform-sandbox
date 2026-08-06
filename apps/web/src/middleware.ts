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

const AUTH_CHECK_FAILURE_EVENT = "middleware_auth_check_failure";

function logAuthCheckFailure({
  category,
  routePath,
  status,
}: {
  category: "fetch_threw" | "unauthorized_401" | "non_ok_response";
  routePath: string;
  status?: number;
}): void {
  console.error(
    JSON.stringify({
      event: AUTH_CHECK_FAILURE_EVENT,
      category,
      routePath,
      ...(status === undefined ? {} : { status }),
    }),
  );
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
    logAuthCheckFailure({
      category: "fetch_threw",
      routePath: request.nextUrl.pathname,
    });
    return NextResponse.redirect(signInUrl);
  }

  if (meResponse.status === 401) {
    logAuthCheckFailure({
      category: "unauthorized_401",
      routePath: request.nextUrl.pathname,
      status: meResponse.status,
    });
    return NextResponse.redirect(signInUrl);
  }
  if (!meResponse.ok) {
    logAuthCheckFailure({
      category: "non_ok_response",
      routePath: request.nextUrl.pathname,
      status: meResponse.status,
    });
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
