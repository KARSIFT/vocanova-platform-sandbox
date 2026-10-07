"use client";

import { useEffect, useRef, type ReactNode } from "react";

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
      disclosure.current.scrollIntoView({ block: "start" });
    };
    revealTarget();
    window.addEventListener("hashchange", revealTarget);
    return () => window.removeEventListener("hashchange", revealTarget);
  }, [anchor]);

  return (
    <details
      ref={disclosure}
      id={anchor}
      className="mt-4 scroll-mt-24 border-t border-neutral-200 pt-2"
      onToggle={(event) => {
        const url = new URL(window.location.href);
        if (event.currentTarget.open) {
          if (url.hash === `#${anchor}`) return;
          url.hash = anchor;
        } else {
          // Only remove this disclosure's fragment; leave other targets alone.
          if (url.hash !== `#${anchor}`) return;
          url.hash = "";
        }
        // Manual opening must survive a sign-in round trip just like an anchor
        // link. Replace the current entry (and preserve Next's history state)
        // so opening/closing does not add extra Back-button stops or scroll.
        window.history.replaceState(
          window.history.state,
          "",
          `${url.pathname}${url.search}${url.hash}`,
        );
      }}
    >
      <summary className="min-h-12 cursor-pointer content-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
        Practise in a sentence
      </summary>
      {children}
    </details>
  );
}
