import type { DailyReviewGoal } from "@vocanova/api-client";

/** Confirmed effort and a session's status are independent of recall accuracy. */
export function dailyReviewView({
  goal,
  dueCount,
  session,
  legacyReviewed = 0,
  legacyTarget = 0,
}: {
  goal?: DailyReviewGoal;
  dueCount: number;
  session?: { status: "in_progress" | "completed" } | null;
  legacyReviewed?: number;
  legacyTarget?: number;
}) {
  const active = session?.status === "in_progress";
  const complete = goal?.complete ?? false;
  return {
    active,
    complete,
    reviewed: goal?.reviewed ?? legacyReviewed,
    target: goal?.target ?? legacyTarget,
    allAvailableFinished: !active && dueCount === 0,
    canReviewExtra: complete && dueCount > 0,
    discoveryIsNext: complete || (!active && dueCount === 0),
  };
}

function calendarDate(instant: Date, timeZone: string) {
  const parts = new Intl.DateTimeFormat("en", {
    timeZone,
    year: "numeric",
    month: "numeric",
    day: "numeric",
  }).formatToParts(instant);
  const part = (name: string) =>
    Number(parts.find((value) => value.type === name)?.value);
  return Date.UTC(part("year"), part("month") - 1, part("day"));
}

/** Calendar-day distance avoids 23/25-hour DST days; exact dates remain available. */
export function relativeReviewDate(
  instant: string,
  timeZone?: string,
  now = new Date(),
): { label: string; exact: string } | undefined {
  const date = new Date(instant);
  if (
    !timeZone ||
    !Number.isFinite(date.getTime()) ||
    !Number.isFinite(now.getTime())
  )
    return undefined;
  try {
    const days = Math.round(
      (calendarDate(date, timeZone) - calendarDate(now, timeZone)) / 86_400_000,
    );
    const label =
      days <= 0
        ? "Today"
        : days === 1
          ? "Tomorrow"
          : days >= 28
            ? `In ${Math.round(days / 30)} ${Math.round(days / 30) === 1 ? "month" : "months"}`
            : days >= 7 && days % 7 === 0
              ? `In ${days / 7} ${days === 7 ? "week" : "weeks"}`
              : `In ${days} days`;
    return {
      label,
      exact: new Intl.DateTimeFormat("en", {
        timeZone,
        year: "numeric",
        month: "long",
        day: "numeric",
        hour: "numeric",
        minute: "2-digit",
        timeZoneName: "short",
      }).format(date),
    };
  } catch {
    return undefined;
  }
}

/** Missing legacy zone/date cannot establish that a learner-local day changed. */
export function reviewGoalDayChanged(
  goal: Pick<DailyReviewGoal, "localDate" | "timeZone">,
  now = new Date(),
): boolean {
  try {
    if (!goal.timeZone || !/^\d{4}-\d{2}-\d{2}$/u.test(goal.localDate))
      return false;
    const [year, month, day] = goal.localDate.split("-").map(Number);
    return (
      calendarDate(now, goal.timeZone) !== Date.UTC(year!, month! - 1, day!)
    );
  } catch {
    return false;
  }
}
