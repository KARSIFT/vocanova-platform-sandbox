import type { DailyMission, DailyReviewGoal } from "@vocanova/api-client";

export type ReviewedTodayProgress = {
  reviewed: number;
  target: number;
  localDate: string;
};

/** Home shows raw effort; mission credit remains capped at the frozen target. */
export function reviewedTodayEvidence(
  mission: Pick<DailyMission, "localDate" | "reviewTarget" | "reviewsCompleted">,
  goal?: Pick<DailyReviewGoal, "localDate" | "target" | "reviewed">,
): ReviewedTodayProgress & { homeLabel: string } {
  if (!goal) throw new Error("Current daily review goal evidence is required");
  if (!/^\d{4}-\d{2}-\d{2}$/u.test(goal.localDate) || mission.localDate !== goal.localDate)
    throw new Error("Mission and review goal local date changed during the read");
  if (!Number.isSafeInteger(goal.target) || goal.target <= 0 || mission.reviewTarget !== goal.target)
    throw new Error("Mission and review goal target must agree");
  if (!Number.isSafeInteger(goal.reviewed) || goal.reviewed < 0)
    throw new Error("Raw reviewed effort must be a non-negative integer");
  if (mission.reviewsCompleted !== Math.min(goal.reviewed, goal.target))
    throw new Error("Capped mission review credit must match raw goal effort");
  return {
    reviewed: goal.reviewed,
    target: goal.target,
    localDate: goal.localDate,
    homeLabel: `Reviewed today: ${goal.reviewed} of ${goal.target}`,
  };
}

/** Extra reviews must add fresh effort even if the mission was already full. */
export function reviewedTodayIncrease(
  before: ReviewedTodayProgress,
  after: ReviewedTodayProgress,
  freshCards: number,
): { minimumExpected: number; reached: boolean } {
  if (before.localDate !== after.localDate)
    throw new Error("Review goal local date changed during the staging journey");
  if (before.target !== after.target)
    throw new Error("Frozen review goal target changed during the staging journey");
  if (!Number.isSafeInteger(freshCards) || freshCards < 1)
    throw new Error("At least one fresh correctly credited card is required");
  const minimumExpected = before.reviewed + freshCards;
  return { minimumExpected, reached: after.reviewed >= minimumExpected };
}
