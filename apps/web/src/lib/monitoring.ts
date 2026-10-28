import type { BrowserOptions } from "@sentry/nextjs";

import {
  sanitizeTelemetryBreadcrumb,
  sanitizeTelemetryEvent,
} from "./telemetry-privacy";

/** The part of the browser SDK the app uses. */
export interface MonitoringSdk {
  init(options: BrowserOptions): unknown;
  captureException(error: unknown): unknown;
  captureRouterTransitionStart(href: string, navigationType: string): void;
}

type Events = Pick<EventTarget, "addEventListener" | "removeEventListener">;

export const MAX_EARLY_ERRORS = 20;
export const MAX_LOAD_ATTEMPTS = 3;

type Timer = ReturnType<typeof setTimeout>;

/**
 * Loads the error monitor on demand. Until it is ready, global errors are kept
 * in a small buffer; one successful start drains that buffer exactly once and
 * hands global handling to the SDK. A failed download or initialisation never
 * rejects. Recovery is bounded: attempts are at least `retryIntervalMs` apart
 * and capped. A reason to try again (an error or report while errors wait,
 * coming back online, or such a signal during a failing attempt) schedules one
 * retry for the earliest allowed time instead of being dropped. While the
 * browser reports being offline, it waits for the online event instead of
 * spending attempts.
 */
export function createMonitor({
  load,
  options,
  events,
  retryIntervalMs = 10_000,
  now = Date.now,
  setTimer = (callback, delay) => setTimeout(callback, delay),
  clearTimer = (timer) => clearTimeout(timer),
  isOnline = () =>
    typeof navigator === "undefined" || navigator.onLine !== false,
}: {
  load: () => Promise<MonitoringSdk>;
  options: BrowserOptions;
  events?: Events;
  retryIntervalMs?: number;
  now?: () => number;
  setTimer?: (callback: () => void, delay: number) => Timer;
  clearTimer?: (timer: Timer) => void;
  isOnline?: () => boolean;
}) {
  let sdk: MonitoringSdk | undefined;
  let loading: Promise<MonitoringSdk | null> | undefined;
  let attempts = 0;
  let lastAttempt = Number.NEGATIVE_INFINITY;
  let listening = false;
  let retryTimer: Timer | undefined;
  // A recovery signal arrived while an attempt was still running.
  let retryAfterAttempt = false;
  const early: unknown[] = [];

  function keep(error: unknown) {
    if (early.length < MAX_EARLY_ERRORS) early.push(error);
  }
  const onError = (event: Event) => {
    keep((event as ErrorEvent).error ?? (event as ErrorEvent).message);
    void ready();
  };
  const onRejection = (event: Event) => {
    keep((event as PromiseRejectionEvent).reason);
    void ready();
  };
  const onOnline = () => void ready();

  function listen(on: boolean) {
    if (!events || listening === on) return;
    listening = on;
    const method = on ? "addEventListener" : "removeEventListener";
    events[method]("error", onError);
    events[method]("unhandledrejection", onRejection);
    events[method]("online", onOnline);
  }

  /** One pending retry at most, never past the attempt cap. */
  function scheduleRetry() {
    if (sdk || retryTimer !== undefined || attempts >= MAX_LOAD_ATTEMPTS)
      return;
    const wait = Math.max(0, lastAttempt + retryIntervalMs - now());
    retryTimer = setTimer(() => {
      retryTimer = undefined;
      void ready();
    }, wait);
  }

  async function attempt(): Promise<MonitoringSdk | null> {
    let started: MonitoringSdk;
    try {
      const loaded = await load();
      loaded.init(options);
      started = loaded;
    } catch {
      // Chunk download or initialisation failed; keep buffering and allow a
      // later bounded retry.
      return null;
    }
    sdk = started;
    if (retryTimer !== undefined) clearTimer(retryTimer);
    retryTimer = undefined;
    // The SDK installs its own global handlers during init.
    listen(false);
    for (const error of early.splice(0)) {
      try {
        started.captureException(error);
      } catch {
        // Reporting must never break the page.
      }
    }
    return started;
  }

  /** Resolves to the started SDK, or null; it never rejects. */
  function ready(): Promise<MonitoringSdk | null> {
    if (sdk) return Promise.resolve(sdk);
    if (loading) {
      retryAfterAttempt = true;
      return loading;
    }
    if (attempts >= MAX_LOAD_ATTEMPTS) return Promise.resolve(null);
    if (attempts > 0 && now() - lastAttempt < retryIntervalMs) {
      scheduleRetry();
      return Promise.resolve(null);
    }
    if (retryTimer !== undefined) clearTimer(retryTimer);
    retryTimer = undefined;
    retryAfterAttempt = false;
    attempts += 1;
    lastAttempt = now();
    loading = attempt().finally(() => {
      loading = undefined;
      if (sdk) return;
      // Errors still waiting, or a signal received meanwhile, are reasons to
      // try again later; offline browsers wait for the online event instead.
      if ((early.length > 0 || retryAfterAttempt) && isOnline())
        scheduleRetry();
      retryAfterAttempt = false;
    });
    return loading;
  }

  return {
    /** Starts collecting global errors until the SDK takes over. */
    listen: () => {
      if (!sdk) listen(true);
    },
    ready,
    capture(error: unknown) {
      if (sdk) {
        try {
          sdk.captureException(error);
        } catch {
          // Reporting must never break the page.
        }
        return;
      }
      keep(error);
      void ready();
    },
    routerTransitionStart(href: string, navigationType: string) {
      try {
        sdk?.captureRouterTransitionStart(href, navigationType);
      } catch {
        // Navigation must not depend on monitoring.
      }
    },
    pending: () => early.length,
  };
}

const sentryDsn = process.env.NEXT_PUBLIC_SENTRY_DSN;

// One monitor per browser page; server renders and DSN-less builds have none.
const monitor =
  sentryDsn && typeof window !== "undefined"
    ? createMonitor({
        load: () => import("./monitoring-sdk"),
        options: {
          dsn: sentryDsn,
          sendDefaultPii: false,
          beforeSend: sanitizeTelemetryEvent,
          beforeBreadcrumb: sanitizeTelemetryBreadcrumb,
          environment:
            process.env.NEXT_PUBLIC_SENTRY_ENVIRONMENT ?? process.env.NODE_ENV,
          release: process.env.NEXT_PUBLIC_SENTRY_RELEASE,
          // `debug` and `spotlight` are the two options that can surface a
          // Sentry developer UI/console output in the browser. Both are pinned
          // off rather than left to their defaults so a production or staging
          // build can never render Sentry's dev overlay to an end user
          // (VOC-051-TEST-01).
          debug: false,
          spotlight: false,
        },
        events: window,
      })
    : undefined;

/**
 * Begins buffering global errors now and starts the monitor once the page has
 * loaded and the browser is idle, so it never competes with the first render.
 */
export function startMonitoring(): void {
  if (!monitor) return;
  monitor.listen();
  const start = () => {
    void monitor.ready();
  };
  const whenIdle = () => {
    if ("requestIdleCallback" in window)
      window.requestIdleCallback(start, { timeout: 3000 });
    else setTimeout(start, 1);
  };
  if (document.readyState === "complete") whenIdle();
  else window.addEventListener("load", whenIdle, { once: true });
}

/** Reports an error, starting the monitor first when it is not ready yet. */
export function captureError(error: unknown): void {
  monitor?.capture(error);
}

export function routerTransitionStart(
  href: string,
  navigationType: string,
): void {
  monitor?.routerTransitionStart(href, navigationType);
}
