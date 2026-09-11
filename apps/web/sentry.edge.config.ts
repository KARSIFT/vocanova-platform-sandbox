import * as Sentry from "@sentry/nextjs";
import {
  sanitizeTelemetryEvent,
  sanitizeTelemetryBreadcrumb,
} from "./src/lib/telemetry-privacy";

const sentryDsn = process.env.NEXT_PUBLIC_SENTRY_DSN;

if (sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    sendDefaultPii: false,
    beforeSend: sanitizeTelemetryEvent,
    beforeBreadcrumb: sanitizeTelemetryBreadcrumb,
    environment:
      process.env.SENTRY_ENVIRONMENT ??
      process.env.NEXT_PUBLIC_SENTRY_ENVIRONMENT ??
      process.env.ENVIRONMENT ??
      process.env.NODE_ENV,
    release:
      process.env.SENTRY_RELEASE ?? process.env.NEXT_PUBLIC_SENTRY_RELEASE,
    debug: false,
    spotlight: false,
  });
}
