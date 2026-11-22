import type { DailyLessonAction, DailyReviewGoal } from "@vocanova/api-client";

/** The server's reason for an "I'm not sure" answer with help. */
export const ASSISTED_REASON = "assisted";

type CheckedFeedback = { correct: boolean; skipped: boolean; reason?: string };

export function isAssistedFeedback(feedback: CheckedFeedback): boolean {
  return !feedback.skipped && feedback.reason === ASSISTED_REASON;
}

/** Assisted help, deliberate skips, misses and correct answers stay distinct. */
export function dailyFeedbackHeading(feedback: CheckedFeedback): string {
  if (feedback.skipped)
    return feedback.reason === "already_known"
      ? "Already marked as known"
      : feedback.reason === "already_reviewed"
        ? "Already reviewed elsewhere"
        : feedback.reason === "removed"
          ? "This word has changed"
          : "Skipped for now";
  if (isAssistedFeedback(feedback)) return "Here’s the meaning";
  return feedback.correct ? "Correct" : "Not quite";
}

/**
 * "I'm not sure" reuses the existing answer contract: help requested and no
 * choice. The server schedules it like a miss and grants no credit or effort.
 */
export function unsureAction(
  stepId: string,
  expectedRevision: number,
  clientActionId: string,
): DailyLessonAction {
  return {
    stepId,
    expectedRevision,
    clientActionId,
    action: "answer",
    wasHintUsed: true,
  };
}

export const MISTAKES_PRACTICE_HREF =
  "/practice?mode=mistakes#practice-mode-mistakes";

/** Link straight to Mistakes only when the server confirms one is available. */
export function completionPracticeLink(availableMistakes?: number): {
  href: string;
  label: string;
} {
  return availableMistakes && availableMistakes > 0
    ? { href: MISTAKES_PRACTICE_HREF, label: "Practice your mistakes" }
    : { href: "/practice", label: "Today’s practice" };
}

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

/** Match the API's preserved elapsed-day contract, then label in learner time. */
export function draftReviewDate(
  interval: `${number}d`,
  timeZone?: string,
  now = new Date(),
): { label: string; exact: string } | undefined {
  const days = Number(interval.slice(0, -1));
  if (
    ![1, 3, 7, 14, 30, 60, 120].includes(days) ||
    !Number.isFinite(now.getTime())
  )
    return undefined;
  return relativeReviewDate(
    new Date(now.getTime() + days * 86_400_000).toISOString(),
    timeZone,
    now,
  );
}
