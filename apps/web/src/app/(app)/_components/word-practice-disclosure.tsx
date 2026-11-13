"use client";

import { useEffect, useRef, type ReactNode } from "react";

function matchesDisclosureTarget(hash: string, anchor: string, alias?: string) {
  return hash === `#${anchor}` || (alias !== undefined && hash === `#${alias}`);
}

function persistDisclosureTarget(
  anchor: string,
  open: boolean,
  alias?: string,
) {
  const url = new URL(window.location.href);
  if (open) {
    if (matchesDisclosureTarget(url.hash, anchor, alias)) return;
    url.hash = anchor;
  } else {
    // Only remove this disclosure's fragment; leave other targets alone.
    if (!matchesDisclosureTarget(url.hash, anchor, alias)) return;
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
  alias,
}: {
  children: ReactNode;
  anchor?: string;
  alias?: string;
}) {
  const disclosure = useRef<HTMLDetailsElement>(null);
  useEffect(() => {
    const revealTarget = () => {
      if (
        !matchesDisclosureTarget(window.location.hash, anchor, alias) ||
        !disclosure.current
      )
        return;
      disclosure.current.open = true;
      // Legacy draft links remain reachable through any containing disclosure,
      // including older layouts with writing inside the personal-tools group.
      let ancestor = disclosure.current.parentElement?.closest("details");
      while (ancestor) {
        ancestor.open = true;
        ancestor = ancestor.parentElement?.closest("details") ?? null;
      }
      disclosure.current.scrollIntoView({ block: "start" });
    };
    revealTarget();
    // Native details can be activated before hydration attaches onClick.
    // Reconcile that already-open state before a reload or sign-in return.
    if (disclosure.current?.open) persistDisclosureTarget(anchor, true, alias);
    window.addEventListener("hashchange", revealTarget);
    return () => window.removeEventListener("hashchange", revealTarget);
  }, [anchor, alias]);

  // A summary target gives native fragment navigation a keyboard starting
  // point inside writing in WebKit too. Keep distinct legacy wrapper anchors.
  return (
    <details
      ref={disclosure}
      id={alias === anchor ? undefined : anchor}
      className="mt-4 scroll-mt-24 border-t border-neutral-200 pt-2"
    >
      <summary
        id={alias}
        className="min-h-12 scroll-mt-24 cursor-pointer content-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        onClick={(event) => {
          if (!event.defaultPrevented && disclosure.current) {
            // Keyboard activation also dispatches click. Record the native
            // action immediately; toggle itself is queued by the browser.
            persistDisclosureTarget(anchor, !disclosure.current.open, alias);
          }
        }}
      >
        Optional writing
      </summary>
      {children}
    </details>
  );
}
