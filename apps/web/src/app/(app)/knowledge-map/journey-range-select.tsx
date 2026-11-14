"use client";

import { Select } from "@/ui/controls";
import type { ReactNode } from "react";

export function JourneyRangeSelect({
  band,
  children,
}: {
  band: number;
  children: ReactNode;
}) {
  return (
    <Select
      id="range-choice"
      name="band"
      aria-describedby="range-choice-navigation"
      value={band}
      onChange={(event) => event.currentTarget.form?.requestSubmit()}
    >
      {children}
    </Select>
  );
}
