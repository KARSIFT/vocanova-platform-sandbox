import type { Breadcrumb, ErrorEvent } from "@sentry/nextjs";

function withoutQuery(value: string): string {
  return value.split(/[?#]/, 1)[0] ?? "";
}

// Verification links are bearer credentials. Monitoring must not retain their
// queries, nor any request body/cookies that could contain a password or session.
export function sanitizeTelemetryEvent(event: ErrorEvent): ErrorEvent {
  if (event.request) {
    const request = { ...event.request };
    delete request.data;
    delete request.cookies;
    delete request.headers;
    delete request.query_string;
    event.request = {
      ...request,
      ...(request.url ? { url: withoutQuery(request.url) } : {}),
    };
  }
  return event;
}

export function sanitizeTelemetryBreadcrumb(
  breadcrumb: Breadcrumb,
): Breadcrumb {
  if (!breadcrumb.data) return breadcrumb;
  const data = { ...breadcrumb.data };
  for (const key of ["url", "from", "to"]) {
    if (typeof data[key] === "string") data[key] = withoutQuery(data[key]);
  }
  return { ...breadcrumb, data };
}
