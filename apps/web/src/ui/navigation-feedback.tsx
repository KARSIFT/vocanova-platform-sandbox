"use client";

import { usePathname, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState, useSyncExternalStore } from "react";

import {
  markNavigationDone,
  serverSnapshot,
  snapshot,
  subscribe,
} from "./navigation-state";

function CommitWatcher() {
  const pathname = usePathname();
  const search = useSearchParams().toString();
  useEffect(() => {
    markNavigationDone();
  }, [pathname, search]);
  useEffect(() => {
    // Back and Forward replace whatever link navigation was waiting.
    window.addEventListener("popstate", markNavigationDone);
    return () => window.removeEventListener("popstate", markNavigationDone);
  }, []);
  return null;
}

/**
 * A thin top bar that acknowledges a tap immediately while the next page is
 * being prepared. It is indeterminate on purpose: it reports that the request
 * is in flight, not invented progress. Slower waits are announced politely.
 */
export function NavigationProgress() {
  const since = useSyncExternalStore(subscribe, snapshot, serverSnapshot);
  const [announce, setAnnounce] = useState(false);
  useEffect(() => {
    if (since === null) {
      setAnnounce(false);
      return;
    }
    const timer = setTimeout(() => setAnnounce(true), 1000);
    return () => clearTimeout(timer);
  }, [since]);
  return (
    <>
      <Suspense fallback={null}>
        <CommitWatcher />
      </Suspense>
      <div
        className="navigation-progress"
        data-active={since === null ? undefined : "true"}
        aria-hidden="true"
      />
      {/* A plain polite live region: page-level status roles stay unique. */}
      <p aria-live="polite" className="sr-only">
        {announce ? "Loading the next page…" : ""}
      </p>
    </>
  );
}
