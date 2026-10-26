import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

import { filterForwardedCookieHeader } from "./lib/cookies";
import { getApiBaseURL } from "./lib/env";
import { connectionErrorDocument } from "./lib/connection-error";
import { THEME_COOKIE_NAME } from "./lib/theme-preference";

export const config = {
  matcher: [
    "/onboarding",
    "/home",
    "/plan",
    "/discover",
    "/discover/:path*",
    "/vocabulary",
    "/vocabulary/:path*",
    "/knowledge-map",
    "/discovery",
    "/courses",
    "/learn/:path*",
    "/writing",
    "/lists",
    "/lists/:path*",
    "/stories",
    "/stories/:path*",
    "/practice",
    "/practice/:path*",
    "/words",
    "/words/:path*",
    "/progress",
    "/progress/:path*",
    "/review",
    "/review/:path*",
    "/reviews",
    "/reviews/:path*",
    "/settings",
    "/settings/:path*",
  ],
};

export const runtime = "nodejs";

interface CurrentUserResponse {
  onboardingStatus?: "not_started" | "in_progress" | "completed";
}

const AUTH_CHECK_FAILURE_EVENT = "middleware_auth_check_failure";

// `apps/web` has no shared server-side logger yet, and this package's scope
// excludes introducing a logging dependency. A single-line JSON payload on
// stderr is what the deployed `web` container's json-file log driver captures
// (see infra/docker-compose.yml's `*default-logging`), so these lines are
// greppable via `docker compose logs web` without SSH-time instrumentation.
// The payload deliberately carries no cookie or credential material.

function logAuthCheckFailure({
  category,
  routePath,
  status,
}: {
  category:
    "fetch_threw" | "body_read_failed" | "unauthorized_401" | "non_ok_response";
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

function connectionErrorResponse(request: NextRequest): NextResponse {
  const html = connectionErrorDocument(
    `${request.nextUrl.pathname}${request.nextUrl.search}`,
    request.cookies.get(THEME_COOKIE_NAME)?.value,
  );
  return new NextResponse(html, {
    status: 503,
    headers: {
      "Cache-Control": "no-store",
      "Content-Type": "text/html; charset=utf-8",
      "Content-Security-Policy":
        "default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'",
      "X-Content-Type-Options": "nosniff",
      "X-Robots-Tag": "noindex, nofollow",
    },
  });
}

export async function middleware(request: NextRequest): Promise<NextResponse> {
  const returnTo = new URLSearchParams({
    returnTo: `${request.nextUrl.pathname}${request.nextUrl.search}`,
  }).toString();
  const signInUrl = new URL(`/login?${returnTo}`, request.url);

  const apiBaseURL = getApiBaseURL();
  // Only this tier's own credentials reach the API; see lib/cookies.ts.
  const cookieHeader = filterForwardedCookieHeader(
    request.headers.get("cookie") ?? "",
  );

  let meResponse: Response;
  try {
    meResponse = await fetch(`${apiBaseURL}/api/v1/me`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        Cookie: cookieHeader,
      },
      credentials: "include",
      // A stalled identity read must fail closed with a recoverable screen,
      // rather than leaving navigation waiting indefinitely.
      signal: AbortSignal.timeout(10_000),
    });
  } catch {
    logAuthCheckFailure({
      category: "fetch_threw",
      routePath: request.nextUrl.pathname,
    });
    return connectionErrorResponse(request);
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
    return connectionErrorResponse(request);
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
    logAuthCheckFailure({
      category: "body_read_failed",
      routePath: request.nextUrl.pathname,
      status: meResponse.status,
    });
    return connectionErrorResponse(request);
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
