"use client";

import NextLink from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, type ComponentProps } from "react";

import { markNavigationStart } from "./navigation-state";

type LinkProps = ComponentProps<typeof NextLink> & {
  /** Turn off intent prefetch too, for dense groups such as map cells. */
  prefetchOnIntent?: boolean;
};

// A pointer resting on a link, or keyboard focus staying on it, signals
// intent; passing over links while moving or tabbing does not.
const HOVER_INTENT_MS = 120;
const FOCUS_INTENT_MS = 300;

function internalHref(href: LinkProps["href"]): string | undefined {
  return typeof href === "string" &&
    href.startsWith("/") &&
    !href.startsWith("//")
    ? href
    : undefined;
}

/**
 * The app's link. Every protected route's prefetch runs the server identity
 * check, so prefetching each visible link cost one API read per link per page
 * view. Links therefore prefetch on intent (a resting mouse pointer or keyboard
 * focus; not touch, where a scroll starts on whatever link is under the
 * finger). Pass `prefetch="auto"` where viewport prefetch is worth it: primary
 * navigation and prominent links into routes with a loading state. Without a
 * prefetched loading state, a tap shows it only after the identity check and
 * React then holds the content back briefly (about 150ms slower in lab tests).
 * Navigation itself, its URL and modifier-click behaviour are unchanged, and
 * the shared progress bar acknowledges the tap immediately.
 */
export default function Link({
  prefetch = false,
  prefetchOnIntent: intent = true,
  onNavigate,
  onPointerEnter,
  onPointerLeave,
  onFocus,
  onBlur,
  ...props
}: LinkProps) {
  const router = useRouter();
  const href = internalHref(props.href);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const cancelIntent = () => clearTimeout(timer.current);
  const intendAfter = (delay: number) => {
    if (!intent || prefetch !== false || !href) return;
    cancelIntent();
    timer.current = setTimeout(() => router.prefetch(href), delay);
  };
  useEffect(() => cancelIntent, []);
  return (
    <NextLink
      {...props}
      prefetch={prefetch}
      onNavigate={(event) => {
        let cancelled = false;
        onNavigate?.({
          preventDefault() {
            cancelled = true;
            event.preventDefault();
          },
        });
        if (!cancelled) markNavigationStart(href);
      }}
      onPointerEnter={(event) => {
        onPointerEnter?.(event);
        if (event.pointerType === "mouse") intendAfter(HOVER_INTENT_MS);
      }}
      onPointerLeave={(event) => {
        onPointerLeave?.(event);
        cancelIntent();
      }}
      onFocus={(event) => {
        onFocus?.(event);
        intendAfter(FOCUS_INTENT_MS);
      }}
      onBlur={(event) => {
        onBlur?.(event);
        cancelIntent();
      }}
    />
  );
}
