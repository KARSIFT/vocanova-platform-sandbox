import { routerTransitionStart, startMonitoring } from "./lib/monitoring";

// The error monitor is about a quarter of the shared JavaScript, so it loads
// after the page has rendered instead of in the critical bundle. Errors raised
// before then are kept and reported once it is ready (see lib/monitoring.ts).
startMonitoring();

export function onRouterTransitionStart(
  href: string,
  navigationType: string,
): void {
  routerTransitionStart(href, navigationType);
}
