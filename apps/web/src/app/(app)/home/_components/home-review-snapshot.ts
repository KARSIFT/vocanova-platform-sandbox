import type { DailyReviewGoal } from "@vocanova/api-client";

/** Compare only the confirmed fields that determine Home's review controls. */
export function homeReviewSnapshot({
  goal,
  dueCount,
  session,
}: {
  goal?: DailyReviewGoal;
  dueCount: number;
  session?: { id: string; status: "in_progress" | "completed" } | null;
}): string {
  return JSON.stringify([
    goal
      ? [
          goal.localDate,
          goal.timeZone,
          goal.target,
          goal.reviewed,
          goal.remaining,
          goal.complete,
          goal.nextReviewAt ?? null,
        ]
      : null,
    dueCount,
    session ? [session.id, session.status] : null,
  ]);
}
