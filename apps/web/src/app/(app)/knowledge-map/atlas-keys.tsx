"use client";

import { useEffect, useRef, type KeyboardEvent, type ReactNode } from "react";

const MOVES: Record<string, number> = {
  ArrowLeft: -1,
  ArrowRight: 1,
  ArrowUp: -10,
  ArrowDown: 10,
};

/**
 * Progressive keyboard enhancement for the 100-square word map: one Tab stop
 * (the selected range) and arrow keys to move between squares. Without
 * JavaScript every square stays an ordinary link in the Tab order.
 */
export function AtlasKeys({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  const root = useRef<HTMLDivElement>(null);
  const links = () =>
    Array.from(root.current?.querySelectorAll<HTMLAnchorElement>("a") ?? []);

  useEffect(() => {
    const all = links();
    const current =
      all.find((link) => link.getAttribute("aria-current") === "true") ??
      all[0];
    for (const link of all) link.tabIndex = link === current ? 0 : -1;
  });

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const all = links();
    const index = all.indexOf(document.activeElement as HTMLAnchorElement);
    if (index < 0) return;
    let next: number | undefined;
    if (event.key in MOVES) next = index + MOVES[event.key]!;
    else if (event.key === "Home") next = index - (index % 10);
    else if (event.key === "End") next = index - (index % 10) + 9;
    if (next === undefined || next < 0 || next >= all.length) return;
    event.preventDefault();
    all[index]!.tabIndex = -1;
    all[next]!.tabIndex = 0;
    all[next]!.focus();
  }

  return (
    <div ref={root} className={className} onKeyDown={onKeyDown}>
      <p className="sr-only">
        Use the arrow keys to move between squares, then Enter to open a range.
      </p>
      {children}
    </div>
  );
}
