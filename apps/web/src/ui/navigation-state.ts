// Per-tab interface state only: which in-app navigation is waiting for the
// server. It is set from browser events, so the server render always sees idle.
let pendingSince: number | null = null;
const listeners = new Set<() => void>();
let clearTimer: ReturnType<typeof setTimeout> | undefined;

function emit() {
  for (const listener of listeners) listener();
}

export function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function snapshot() {
  return pendingSince;
}

export function serverSnapshot() {
  return null;
}

// Compared the way the router reports search parameters, so `%20` and `+`
// spellings of the same query do not look like a navigation.
function sameAddress(target: URL): boolean {
  return (
    target.pathname === window.location.pathname &&
    new URLSearchParams(target.search).toString() ===
      new URLSearchParams(window.location.search).toString()
  );
}

/**
 * Marks an in-app navigation as started. A link to the page already shown
 * (including hash-only links) starts nothing, and supersedes any navigation
 * that was still waiting: the router abandons it for the newer request.
 */
export function markNavigationStart(href: string | null | undefined): void {
  if (!href || typeof window === "undefined") return;
  const target = new URL(href, window.location.href);
  if (target.origin !== window.location.origin) return;
  if (sameAddress(target)) {
    markNavigationDone();
    return;
  }
  pendingSince = Date.now();
  clearTimeout(clearTimer);
  // A navigation that never commits (offline, full-page fallback) must not
  // leave the indicator behind.
  clearTimer = setTimeout(markNavigationDone, 15_000);
  emit();
}

/** Whether an in-app navigation is waiting (for tests and diagnostics). */
export function isNavigationPending(): boolean {
  return pendingSince !== null;
}

export function markNavigationDone(): void {
  if (pendingSince === null) return;
  pendingSince = null;
  clearTimeout(clearTimer);
  emit();
}
