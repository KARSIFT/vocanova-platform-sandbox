import * as Sentry from "@sentry/nextjs";

const sentryDsn = process.env.NEXT_PUBLIC_SENTRY_DSN;

if (sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    environment:
      process.env.NEXT_PUBLIC_SENTRY_ENVIRONMENT ?? process.env.NODE_ENV,
    release: process.env.NEXT_PUBLIC_SENTRY_RELEASE,
    // `debug` and `spotlight` are the two options that can surface a Sentry
    // developer UI/console output in the browser. Both are pinned off rather
    // than left to their defaults so a production or staging build can never
    // render Sentry's dev overlay to an end user (VOC-051-TEST-01).
    debug: false,
    spotlight: false,
  });
}

export const onRouterTransitionStart = Sentry.captureRouterTransitionStart;
