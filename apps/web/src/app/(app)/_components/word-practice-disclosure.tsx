"use client";

import { useEffect, useRef, type ReactNode } from "react";

function persistDisclosureTarget(anchor: string, open: boolean) {
  const url = new URL(window.location.href);
  if (open) {
    if (url.hash === `#${anchor}`) return;
    url.hash = anchor;
  } else {
    // Only remove this disclosure's fragment; leave other targets alone.
    if (url.hash !== `#${anchor}`) return;
    url.hash = "";
  }
  // Preserve Next's history state without adding Back-button stops or scrolling.
  window.history.replaceState(
    window.history.state,
    "",
    `${url.pathname}${url.search}${url.hash}`,
  );
}

// Keep the editor mounted when closed: local drafts, pending requests and exact
// retry identities belong to the editor, not the presentation disclosure.
export function WordPracticeDisclosure({
  children,
  anchor = "sentence-practice",
}: {
  children: ReactNode;
  anchor?: string;
}) {
  const disclosure = useRef<HTMLDetailsElement>(null);
  useEffect(() => {
    const revealTarget = () => {
      if (window.location.hash !== `#${anchor}` || !disclosure.current) return;
      disclosure.current.open = true;
      // Legacy draft links remain reachable even when optional writing lives
      // inside the collapsed personal-tools group.
      let ancestor = disclosure.current.parentElement?.closest("details");
      while (ancestor) {
        ancestor.open = true;
        ancestor = ancestor.parentElement?.closest("details") ?? null;
      }
      disclosure.current.scrollIntoView({ block: "start" });
    };
    revealTarget();
    // Native details can be activated before hydration attaches onToggle.
    // Reconcile that already-open state before a reload or sign-in return.
    if (disclosure.current?.open) persistDisclosureTarget(anchor, true);
    window.addEventListener("hashchange", revealTarget);
    return () => window.removeEventListener("hashchange", revealTarget);
  }, [anchor]);

  return (
    <details
      ref={disclosure}
      id={anchor}
      className="mt-4 scroll-mt-24 border-t border-neutral-200 pt-2"
      onToggle={(event) => {
        persistDisclosureTarget(anchor, event.currentTarget.open);
      }}
    >
      <summary
        className="min-h-12 cursor-pointer content-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        onClick={(event) => {
          if (!event.defaultPrevented && disclosure.current) {
            // Keyboard activation also dispatches click. Record the native
            // action immediately; toggle itself is queued by the browser.
            persistDisclosureTarget(anchor, !disclosure.current.open);
          }
        }}
      >
        Optional writing
      </summary>
      {children}
    </details>
  );
}
