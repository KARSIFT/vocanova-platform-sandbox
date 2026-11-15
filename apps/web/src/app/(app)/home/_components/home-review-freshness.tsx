"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import type { DailyReviewGoal } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { Button } from "@/ui/controls";
import { reviewGoalDayChanged } from "../../reviews/_components/daily-review-view";
import { homeReviewSnapshot } from "./home-review-snapshot";

/** Refresh SSR controls only after a confirmed read; this leaf never owns credit. */
export function HomeReviewFreshness({
  goal,
  dueCount,
  session,
}: {
  goal?: DailyReviewGoal;
  dueCount: number;
  session: { id: string; status: "in_progress" | "completed" } | null;
}) {
  const router = useRouter();
  const [error, setError] = useState(false);
  const [busy, setBusy] = useState(false);
  const lastRequestAt = useRef(0);
  const refreshRequested = useRef<string | undefined>(undefined);
  const retry = useRef<() => void>(() => {});
  const snapshot = homeReviewSnapshot({ goal, dueCount, session });
  const localDate = goal?.localDate;
  const timeZone = goal?.timeZone;

  useEffect(() => {
    if (!localDate || !timeZone) return;
    let active = true;
    let inFlight = false;
    let failed = false;
    const controller = new AbortController();
    setError(false);
    async function check(reason: "timer" | "foreground" | "retry") {
      if (!active || inFlight || document.visibilityState === "hidden") return;
      if (
        reason === "timer" &&
        (failed ||
          !reviewGoalDayChanged({ localDate: localDate!, timeZone: timeZone! }))
      )
        return;
      // Focus and visibility events commonly arrive together. Coalesce them,
      // and leave failed background checks stopped until the learner returns.
      if (reason !== "retry" && Date.now() - lastRequestAt.current < 15_000)
        return;
      lastRequestAt.current = Date.now();
      inFlight = true;
      setBusy(true);
      try {
        const { data } = await createApiClient().getCurrentDailyLesson({
          cache: "no-store",
          signal: AbortSignal.any([
            controller.signal,
            AbortSignal.timeout(5000),
          ]),
        });
        if (!active) return;
        if (!data.goal) throw new Error("Daily progress unavailable");
        failed = false;
        setError(false);
        const confirmed = homeReviewSnapshot(data);
        if (confirmed !== snapshot && confirmed !== refreshRequested.current) {
          refreshRequested.current = confirmed;
          router.refresh();
        }
      } catch {
        if (active) {
          failed = true;
          setError(true);
        }
      } finally {
        inFlight = false;
        if (active) setBusy(false);
      }
    }
    const foreground = () => void check("foreground");
    retry.current = () => void check("retry");
    const timer = window.setInterval(() => void check("timer"), 30_000);
    document.addEventListener("visibilitychange", foreground);
    window.addEventListener("focus", foreground);
    void check("timer");
    return () => {
      active = false;
      controller.abort();
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", foreground);
      window.removeEventListener("focus", foreground);
      retry.current = () => {};
    };
  }, [localDate, timeZone, snapshot, router]);

  if (!error) return null;
  return (
    <div className="mt-3 text-sm text-neutral-700">
      <p role="status">
        Progress may be out of date. Refresh to check today’s reviews.
      </p>
      <Button
        type="button"
        variant="secondary"
        className="mt-2"
        disabled={busy}
        onClick={() => retry.current()}
      >
        Refresh today’s progress
      </Button>
    </div>
  );
}
