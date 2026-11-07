"use client";

import { useEffect } from "react";

export function JourneyRangeAnchor() {
  useEffect(() => {
    const { pathname, search, hash } = window.location;
    // Next can append the fragment to a cached Journey URL that already has it.
    if (!/^(?:#range-words){2,}$/.test(hash)) return;
    window.history.replaceState(
      window.history.state,
      "",
      `${pathname}${search}#range-words`,
    );
    document.getElementById("range-words")?.scrollIntoView();
  }, []);

  return null;
}
