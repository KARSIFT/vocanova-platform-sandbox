import * as Sentry from "@sentry/nextjs";

export async function register(): Promise<void> {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    await import("../sentry.server.config");
  }

  if (process.env.NEXT_RUNTIME === "edge") {
    await import("../sentry.edge.config");
  }
}

// Next.js calls this hook for every server-side request error (Route
// Handlers, Server Actions, RSC rendering). Without it those errors never
// reach Sentry: `register()` only initialises the SDK, which by itself covers
// uncaught/global handlers, not errors Next.js catches and turns into a 500.
export const onRequestError = Sentry.captureRequestError;
