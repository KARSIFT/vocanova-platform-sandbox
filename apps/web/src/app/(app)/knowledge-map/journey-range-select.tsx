"use client";

import type { ReactNode } from "react";

export function JourneyRangeSelect({
  band,
  children,
}: {
  band: number;
  children: ReactNode;
}) {
  return (
    <select
      id="range-choice"
      name="band"
      aria-describedby="range-choice-navigation"
      value={band}
      onChange={(event) => event.currentTarget.form?.requestSubmit()}
    >
      {children}
    </select>
  );
}
